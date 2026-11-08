package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/geometry"
)

func request(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "localhost"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestHTTPRejectsCrossOriginAndMalformedProfile(t *testing.T) {
	s, e := New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	h := s.Handler(fstest.MapFS{"index.html": {Data: []byte("Forge")}})
	r := httptest.NewRequest("POST", "/api/engine", strings.NewReader(`{"path":"cmd.exe"}`))
	r.Host = "localhost"
	r.Header.Set("Origin", "https://malicious.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross origin mutation accepted")
	}
	w = request(t, h, "POST", "/api/profiles/validate", `{"kind":"printer","profile":{"schemaVersion":"1.0"}}`)
	var result struct {
		Valid  bool
		Errors []string
	}
	if e = json.Unmarshal(w.Body.Bytes(), &result); e != nil || result.Valid || len(result.Errors) == 0 {
		t.Fatalf("invalid profile validation result %s", w.Body.String())
	}
	w = request(t, h, "GET", "/api/jobs/unknown/gcode", "")
	if w.Code == 200 {
		t.Fatal("unvalidated export accepted")
	}
}
func fixtureCube(x, y, z float64) []geometry.Triangle {
	v := []geometry.Vec{{0, 0, 0}, {x, 0, 0}, {x, y, 0}, {0, y, 0}, {0, 0, z}, {x, 0, z}, {x, y, z}, {0, y, z}}
	faces := [][3]int{{0, 2, 1}, {0, 3, 2}, {4, 5, 6}, {4, 6, 7}, {0, 1, 5}, {0, 5, 4}, {1, 2, 6}, {1, 6, 5}, {2, 3, 7}, {2, 7, 6}, {3, 0, 4}, {3, 4, 7}}
	var ts []geometry.Triangle
	for _, f := range faces {
		ts = append(ts, geometry.Triangle{v[f[0]], v[f[1]], v[f[2]]})
	}
	return ts
}
func TestNativeEndToEnd(t *testing.T) {
	if os.Getenv("FORGE_INTEGRATION") != "1" {
		t.Skip("set FORGE_INTEGRATION=1 with a native PrusaSlicer installed")
	}
	for _, fixture := range []struct {
		name      string
		mode      string
		triangles []geometry.Triangle
	}{{"calibration-cube", "fast", fixtureCube(20, 20, 20)}, {"thin-wall", "quality", fixtureCube(20, .9, 10)}, {"boundary", "material", fixtureCube(210, 5, 2)}, {"bridge", "balanced", bridgeFixture()}, {"overhang-support", "strength", overhangFixture()}, {"easy-remove-support", "balanced", overhangFixture()}, {"stable-support", "balanced", overhangFixture()}} {
		t.Run(fixture.name, func(t *testing.T) {
			s, e := New(t.TempDir())
			if e != nil {
				t.Fatal(e)
			}
			defer s.Close()
			if s.backend == nil {
				t.Fatal("FORGE_INTEGRATION requested but PrusaSlicer missing")
			}
			h := s.Handler(fstest.MapFS{"index.html": {Data: []byte("Forge")}})
			m, e := geometry.Analyze(fixture.triangles)
			if e != nil {
				t.Fatal(e)
			}
			var stl bytes.Buffer
			if e = m.WriteSTL(&stl); e != nil {
				t.Fatal(e)
			}
			var upload bytes.Buffer
			mw := multipart.NewWriter(&upload)
			part, _ := mw.CreateFormFile("file", fixture.name+".stl")
			io.Copy(part, &stl)
			mw.Close()
			r := httptest.NewRequest("POST", "/api/models", &upload)
			r.Host = "localhost"
			r.Header.Set("Content-Type", mw.FormDataContentType())
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 201 {
				t.Fatal(w.Body.String())
			}
			var model Model
			if e = json.Unmarshal(w.Body.Bytes(), &model); e != nil {
				t.Fatal(e)
			}
			ps, fs := config.ReferenceFixture()
			p, f := ps[0], fs[0]
			p.Extruders[0].FlowVerified = false
			p.Extruders[0].MaxVolumetricFlow = 4
			delete(p.Provenance, "maxVolumetricFlow")
			profileBody, _ := json.Marshal(map[string]any{"kind": "printer", "profile": p})
			if response := request(t, h, "POST", "/api/profiles", string(profileBody)); response.Code != 200 {
				t.Fatal(response.Body.String())
			}
			supportMode := "auto"
			if fixture.name == "easy-remove-support" {
				supportMode = "easy_remove"
			}
			if fixture.name == "stable-support" {
				supportMode = "stable"
			}
			data, _ := json.Marshal(map[string]any{"modelId": model.ID, "printer": p, "filament": f, "mode": fixture.mode, "budget": 100, "supports": true, "supportMode": supportMode, "minWalls": 2})
			w = request(t, h, "POST", "/api/jobs", string(data))
			if w.Code != 202 {
				t.Fatal(w.Body.String())
			}
			var job Job
			json.Unmarshal(w.Body.Bytes(), &job)
			deadline := time.Now().Add(2 * time.Minute)
			for time.Now().Before(deadline) {
				w = request(t, h, "GET", "/api/jobs/"+job.ID, "")
				if e = json.Unmarshal(w.Body.Bytes(), &job); e != nil {
					t.Fatal(e)
				}
				if job.Status == "complete" {
					break
				}
				if job.Status == "failed" || job.Status == "cancelled" {
					t.Fatalf("native job %s: %s", job.Status, job.Error)
				}
				time.Sleep(100 * time.Millisecond)
			}
			if job.Status != "complete" {
				t.Fatal("native job timed out")
			}
			if job.Report == nil || job.Report.Actual == nil || job.Report.BaselineActual == nil {
				t.Fatalf("missing calculated baseline comparison: %+v", job.Report)
			}
			if job.Report.Readiness == nil {
				t.Fatal("missing geometry readiness analysis")
			}
			if fixture.name == "overhang-support" && (len(job.Report.SupportCandidates) != 2 || !job.Report.Selected.Supports || !job.Report.Readiness.SupportsRecommended) {
				t.Fatalf("overhang did not compare both support shapes: %+v", job.Report)
			}
			if fixture.name == "calibration-cube" && job.Report.Selected.Supports {
				t.Fatal("auto added unnecessary supports to cube")
			}
			if supportMode == "stable" && job.Report.Selected.SupportStyle != "grid" {
				t.Fatal("stable support not grid")
			}
			if supportMode == "easy_remove" && job.Report.Selected.SupportStyle != "organic" {
				t.Fatal("easy removal not organic")
			}
			w = request(t, h, "GET", "/api/jobs/"+job.ID+"/preview", "")
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"segments":[`) {
				t.Fatal("preview missing")
			}
			w = request(t, h, "GET", "/api/jobs/"+job.ID+"/gcode", "")
			if w.Code != 200 || w.Body.Len() < 1000 || !strings.Contains(w.Body.String(), "G1") {
				t.Fatal("genuine G-code export missing")
			}
			a, b := job.Report.Actual, job.Report.BaselineActual
			t.Logf("%s: selected %.0fs %.3fg, baseline %.0fs %.3fg, paths %.0fmm, native compute %.3fs", fixture.name, a.Seconds, a.MassG, b.Seconds, b.MassG, a.PathMM, a.SliceSeconds)
		})
	}
}
func shifted(ts []geometry.Triangle, x, y, z float64) []geometry.Triangle {
	for i := range ts {
		for j := range ts[i] {
			ts[i][j][0] += x
			ts[i][j][1] += y
			ts[i][j][2] += z
		}
	}
	return ts
}
func bridgeFixture() []geometry.Triangle {
	ts := fixtureCube(5, 15, 15)
	ts = append(ts, shifted(fixtureCube(5, 15, 15), 25, 0, 0)...)
	return append(ts, shifted(fixtureCube(30, 15, 3), 0, 0, 15)...)
}
func overhangFixture() []geometry.Triangle {
	return append(fixtureCube(6, 15, 20), shifted(fixtureCube(25, 15, 3), 0, 0, 20)...)
}
func TestCancelQueuedJob(t *testing.T) {
	s, e := New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	j := &Job{ID: "cancel-test", Status: "queued", cancel: cancel}
	s.jobs[j.ID] = j
	h := s.Handler(fstest.MapFS{})
	w := request(t, h, "DELETE", "/api/jobs/cancel-test", "")
	if w.Code != 200 || ctx.Err() != context.Canceled {
		t.Fatal("cancellation not propagated")
	}
}
func ExampleMetrics() {
	m := Metrics{MassG: 1.25, Seconds: 120}
	fmt.Printf("Calculated estimate: %.2f g, %.0f s", m.MassG, m.Seconds)
	// Output: Calculated estimate: 1.25 g, 120 s
}
func TestPrinterFirstSetupPersistenceAndMaterials(t *testing.T) {
	dir := t.TempDir()
	s, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	h := s.Handler(fstest.MapFS{})
	w := request(t, h, "GET", "/api/state", "")
	var state struct {
		Printers  []config.Printer  `json:"printers"`
		Filaments []config.Filament `json:"filaments"`
		Active    string            `json:"activePrinterId"`
	}
	if e = json.Unmarshal(w.Body.Bytes(), &state); e != nil {
		t.Fatal(e)
	}
	if len(state.Printers) != 0 || state.Active != "" {
		t.Fatal("first launch must not install a placeholder printer")
	}
	if len(state.Filaments) < 6 {
		t.Fatal("multiple material starters must be available")
	}
	if w = request(t, h, "GET", "/api/materials", ""); w.Code != 404 {
		t.Fatal("material compatibility cannot assume an unconfigured printer")
	}
	w = request(t, h, "GET", "/api/printer-template", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"width":null`) {
		t.Fatalf("template must preserve unknown hardware: %s", w.Body.String())
	}
	ps, _ := config.ReferenceFixture()
	p := ps[0]
	p.ID = "my-printer"
	p.Name = "My imported machine"
	p.Extruders[0].FlowVerified = false
	delete(p.Provenance, "maxVolumetricFlow")
	p.Thermal.MaxNozzle = 220
	p.Thermal.MaxBed = 70
	body, _ := json.Marshal(map[string]any{"kind": "printer", "profile": p})
	w = request(t, h, "POST", "/api/profiles", string(body))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request(t, h, "GET", "/api/materials?printerId=my-printer", "")
	var materials []MaterialChoice
	if e = json.Unmarshal(w.Body.Bytes(), &materials); e != nil || len(materials) < 6 {
		t.Fatalf("material catalogue missing: %s", w.Body.String())
	}
	compatible, incompatible := 0, 0
	for _, choice := range materials {
		if choice.Compatible {
			compatible++
		} else {
			incompatible++
			if len(choice.Reasons) == 0 {
				t.Fatal("unavailable recipe must explain machine constraint")
			}
		}
	}
	if compatible == 0 || incompatible == 0 {
		t.Fatal("catalogue must reflect actual thermal/cooling limits")
	}
	q := p
	q.ID = "second-printer"
	q.Name = "Second machine"
	q.BuildVolume.Width = 300
	body, _ = json.Marshal(map[string]any{"kind": "printer", "profile": q})
	w = request(t, h, "POST", "/api/profiles", string(body))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request(t, h, "POST", "/api/printers/select", `{"id":"my-printer"}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	s.Close()
	reopened, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer reopened.Close()
	w = request(t, reopened.Handler(fstest.MapFS{}), "GET", "/api/state", "")
	if e = json.Unmarshal(w.Body.Bytes(), &state); e != nil {
		t.Fatal(e)
	}
	if state.Active != "my-printer" || len(state.Printers) != 2 {
		t.Fatal("imported printer selection must survive restart")
	}
}
func TestCannotSliceWithUnimportedOrUnsavedPrinter(t *testing.T) {
	s, e := New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	h := s.Handler(fstest.MapFS{})
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	jobBody := func() string {
		b, _ := json.Marshal(map[string]any{"modelId": "unused", "printer": p, "filament": f, "mode": "balanced", "budget": 64, "supports": true, "minWalls": 2})
		return string(b)
	}
	w := request(t, h, "POST", "/api/jobs", jobBody())
	if w.Code != 409 || !strings.Contains(w.Body.String(), "import and save") {
		t.Fatal("unregistered printer bypassed setup")
	}
	body, _ := json.Marshal(map[string]any{"kind": "printer", "profile": p})
	w = request(t, h, "POST", "/api/profiles", string(body))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	p.BuildVolume.Width += 1
	w = request(t, h, "POST", "/api/jobs", jobBody())
	if w.Code != 409 || !strings.Contains(w.Body.String(), "save any changed") {
		t.Fatal("unsaved hardware changes bypassed imported profile")
	}
}
func TestPreferencesCanBeSavedAndRestored(t *testing.T) {
	dir := t.TempDir()
	s, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	w := request(t, s.Handler(fstest.MapFS{}), "POST", "/api/profiles", `{"kind":"optimization","profile":{"schemaVersion":"1.0","mode":"material","budget":64,"minWalls":3,"minInfill":0.15,"supports":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	s.Close()
	reopened, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer reopened.Close()
	if reopened.preferences == nil || reopened.preferences.Mode != "material" || reopened.preferences.MinWalls != 3 {
		t.Fatal("saved preferences were lost")
	}
}
