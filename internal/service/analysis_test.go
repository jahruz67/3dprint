package service

import (
	"encoding/json"
	"github.com/forge-slicer/forge/internal/analysis"
	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/geometry"
	"testing"
	"testing/fstest"
)

func TestReadinessRequiresMachineAndUsesTransform(t *testing.T) {
	s, e := New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	m, e := geometry.Analyze(fixtureCube(10, 10, 10))
	if e != nil {
		t.Fatal(e)
	}
	s.models["cube"] = Model{ID: "cube", mesh: m}
	h := s.Handler(fstest.MapFS{})
	if w := request(t, h, "POST", "/api/models/cube/analyze", `{}`); w.Code != 409 {
		t.Fatalf("analysis before setup: %d", w.Code)
	}
	ps, _ := config.ReferenceFixture()
	s.printers = ps
	s.activePrinterID = ps[0].ID
	w := request(t, h, "POST", "/api/models/cube/analyze", `{"supportAngle":45}`)
	var report analysis.Report
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &report) != nil || report.SupportsRecommended {
		t.Fatalf("cube readiness incorrect: %s", w.Body.String())
	}
	if w = request(t, h, "POST", "/api/models/cube/analyze", `{"supportAngle":90}`); w.Code != 400 {
		t.Fatal("invalid angle accepted")
	}
	w = request(t, h, "POST", "/api/models/cube/analyze", `{"transform":{"scale":[1,1,1],"rotation":[0,0,0],"position":[100,100,5]}}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if json.Unmarshal(w.Body.Bytes(), &report) != nil {
		t.Fatal("invalid report")
	}
	found := false
	for _, f := range report.Findings {
		if f.Action == "place_on_bed" {
			found = true
		}
	}
	if !found {
		t.Fatal("raised model has no place-on-bed guidance")
	}
}
