package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/forge-slicer/forge/internal/analysis"
	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/engine"
	"github.com/forge-slicer/forge/internal/gcode"
	"github.com/forge-slicer/forge/internal/geometry"
	"github.com/forge-slicer/forge/internal/optimizer"
)

const Version = "0.3.1"

type Model struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Triangles int             `json:"triangles"`
	Bounds    geometry.Bounds `json:"bounds"`
	Volume    float64         `json:"volumeMm3"`
	Area      float64         `json:"areaMm2"`
	Warnings  []string        `json:"warnings"`
	mesh      *geometry.Mesh
}
type Metrics struct {
	Seconds           float64 `json:"seconds"`
	MassG             float64 `json:"massG"`
	FilamentMM        float64 `json:"filamentMm"`
	PathMM            float64 `json:"pathMm"`
	SupportFilamentMM float64 `json:"supportFilamentMm"`
	SliceSeconds      float64 `json:"sliceSeconds"`
}
type Report struct {
	optimizer.Report
	Readiness              *analysis.Report   `json:"readiness,omitempty"`
	SupportCandidates      []SupportCandidate `json:"supportCandidates,omitempty"`
	SupportSelectionReason string             `json:"supportSelectionReason,omitempty"`
	Actual                 *Metrics           `json:"actual,omitempty"`
	BaselineActual         *Metrics           `json:"baselineActual,omitempty"`
	TimeDifferencePercent  *float64           `json:"timeDifferencePercent,omitempty"`
	MassDifferencePercent  *float64           `json:"massDifferencePercent,omitempty"`
	Warnings               []string           `json:"warnings"`
}
type SupportCandidate struct {
	Style             string  `json:"style"`
	Seconds           float64 `json:"seconds"`
	MassG             float64 `json:"massG"`
	SupportFilamentMM float64 `json:"supportFilamentMm"`
	Selected          bool    `json:"selected"`
}
type Job struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	Progress     float64 `json:"progress"`
	Message      string  `json:"message"`
	Error        string  `json:"error,omitempty"`
	Report       *Report `json:"report,omitempty"`
	PreviewReady bool    `json:"previewReady"`
	cancel       context.CancelFunc
	preview      *gcode.Preview
	path, dir    string
}
type JobRequest struct {
	ModelID          string              `json:"modelId"`
	Printer          json.RawMessage     `json:"printer"`
	Filament         json.RawMessage     `json:"filament"`
	Mode             string              `json:"mode"`
	Transform        *geometry.Transform `json:"transform"`
	Budget           int                 `json:"budget"`
	Supports         *bool               `json:"supports"`
	SupportMode      string              `json:"supportMode"`
	SupportStyle     string              `json:"supportStyle"`
	SupportPlacement string              `json:"supportPlacement"`
	SupportAngle     float64             `json:"supportAngle"`
	MinWalls         int                 `json:"minWalls"`
	MinInfill        float64             `json:"minInfill"`
	LoadDirection    [3]float64          `json:"loadDirection"`
}
type Service struct {
	mu               sync.RWMutex
	models           map[string]Model
	jobs             map[string]*Job
	order            []string
	active           bool
	backend          *engine.Prusa
	printers         []config.Printer
	filaments        []config.Filament
	activePrinterID  string
	preferences      *config.Preferences
	dataDir, tempDir string
	warnings         []string
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	saveDialog       func(string) (string, error)
	analysisSlots    chan struct{}
}

func New(dataDir string) (*Service, error) {
	if dataDir == "" {
		base, e := os.UserConfigDir()
		if e != nil {
			return nil, e
		}
		dataDir = filepath.Join(base, "ForgeSlicer")
	}
	if e := os.MkdirAll(filepath.Join(dataDir, "profiles"), 0700); e != nil {
		return nil, e
	}
	temp, e := os.MkdirTemp("", "forge-slicer-")
	if e != nil {
		return nil, e
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{models: map[string]Model{}, jobs: map[string]*Job{}, backend: engine.Discover(), dataDir: dataDir, tempDir: temp, ctx: ctx, cancel: cancel, warnings: []string{}}
	s.analysisSlots = make(chan struct{}, 2)
	s.printers, s.filaments = config.Defaults()
	files, _ := filepath.Glob(filepath.Join(dataDir, "profiles", "*.json"))
	for _, file := range files {
		b, e := os.ReadFile(file)
		if e != nil {
			continue
		}
		if strings.HasPrefix(filepath.Base(file), "printer-") {
			p, e := config.ParsePrinter(b)
			if e != nil {
				s.warnings = append(s.warnings, filepath.Base(file)+": "+e.Error())
				continue
			}
			s.upsertPrinter(p)
		} else if strings.HasPrefix(filepath.Base(file), "filament-") {
			f, e := config.ParseFilament(b)
			if e != nil {
				s.warnings = append(s.warnings, filepath.Base(file)+": "+e.Error())
				continue
			}
			s.upsertFilament(f)
		}
	}
	if b, e := os.ReadFile(filepath.Join(dataDir, "engine-path.txt")); e == nil {
		if p, e := engine.New(strings.TrimSpace(string(b))); e == nil {
			s.backend = p
		}
	}
	if b, e := os.ReadFile(filepath.Join(dataDir, "active-printer.txt")); e == nil {
		candidate := strings.TrimSpace(string(b))
		for _, p := range s.printers {
			if p.ID == candidate {
				s.activePrinterID = candidate
				break
			}
		}
	}
	if s.activePrinterID == "" && len(s.printers) > 0 {
		s.activePrinterID = s.printers[0].ID
	}
	if b, e := os.ReadFile(filepath.Join(dataDir, "preferences.json")); e == nil {
		if pref, e := config.ParsePreferences(b); e == nil {
			s.preferences = &pref
		} else {
			s.warnings = append(s.warnings, "preferences.json: "+e.Error())
		}
	}
	return s, nil
}
func (s *Service) Close() { s.cancel(); s.wg.Wait(); _ = os.RemoveAll(s.tempDir) }
func id() string {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if e := json.NewEncoder(w).Encode(v); e != nil {
		slog.Error("response encoding", "error", e)
	}
}
func errorOut(w http.ResponseWriter, status int, e error) {
	jsonOut(w, status, map[string]string{"error": e.Error()})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		errorOut(w, 415, errors.New("Content-Type must be application/json"))
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		errorOut(w, 400, e)
		return false
	}
	if e := d.Decode(new(any)); e != io.EOF {
		errorOut(w, 400, errors.New("expected exactly one JSON document"))
		return false
	}
	return true
}
func (s *Service) Handler(assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.state)
	mux.HandleFunc("GET /api/materials", s.materials)
	mux.HandleFunc("GET /api/printer-template", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, config.PrinterTemplate()) })
	mux.HandleFunc("POST /api/printers/select", s.selectPrinter)
	mux.HandleFunc("POST /api/models", s.importModel)
	mux.HandleFunc("GET /api/models/{id}/mesh", s.mesh)
	mux.HandleFunc("POST /api/models/{id}/analyze", s.analyzeModel)
	mux.HandleFunc("POST /api/jobs", s.startJob)
	mux.HandleFunc("GET /api/jobs/{id}", s.job)
	mux.HandleFunc("DELETE /api/jobs/{id}", s.cancelJob)
	mux.HandleFunc("GET /api/jobs/{id}/preview", s.preview)
	mux.HandleFunc("GET /api/jobs/{id}/gcode", s.export)
	mux.HandleFunc("POST /api/profiles/validate", s.validateProfile)
	mux.HandleFunc("POST /api/profiles", s.saveProfile)
	mux.HandleFunc("POST /api/engine", s.setEngine)
	mux.HandleFunc("POST /api/jobs/{id}/save", s.saveNative)
	mux.Handle("/", http.FileServer(http.FS(assets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; worker-src 'self' blob:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		host := r.Host
		if h, _, ok := strings.Cut(host, ":"); ok {
			host = h
		}
		if host != "127.0.0.1" && host != "localhost" && host != "wails.localhost" && host != "wails" && host != "" {
			errorOut(w, 403, errors.New("unrecognized local host"))
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || !strings.EqualFold(u.Host, r.Host) {
					errorOut(w, 403, errors.New("cross-origin mutations are forbidden"))
					return
				}
			}
		}
		defer func() {
			if v := recover(); v != nil {
				slog.Error("request panic", "error", v)
				errorOut(w, 500, errors.New("internal error; see application log"))
			}
		}()
		mux.ServeHTTP(w, r)
	})
}
func (s *Service) state(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jsonOut(w, 200, map[string]any{"printers": s.printers, "filaments": s.filaments, "activePrinterId": s.activePrinterID, "preferences": s.preferences, "engine": s.backend.Status(), "version": Version, "warnings": s.warnings, "nativeDialogs": s.saveDialog != nil})
}

type MaterialChoice struct {
	Profile    config.Filament `json:"profile"`
	Compatible bool            `json:"compatible"`
	Reasons    []string        `json:"reasons"`
}

func materialChoice(p config.Printer, f config.Filament) MaterialChoice {
	reasons := []string{}
	if f.NozzleTemperature < p.Thermal.MinNozzle || f.NozzleTemperature > p.Thermal.MaxNozzle {
		reasons = append(reasons, fmt.Sprintf("Recipe needs %.0f°C nozzle; this printer profile allows %.0f–%.0f°C.", f.NozzleTemperature, p.Thermal.MinNozzle, p.Thermal.MaxNozzle))
	}
	if f.BedTemperature < p.Thermal.MinBed || f.BedTemperature > p.Thermal.MaxBed {
		reasons = append(reasons, fmt.Sprintf("Recipe needs %.0f°C bed; this printer profile allows %.0f–%.0f°C.", f.BedTemperature, p.Thermal.MinBed, p.Thermal.MaxBed))
	}
	if !p.Thermal.PartCooling && f.FanPercent > 0 {
		reasons = append(reasons, "Recipe requests a part-cooling fan; this printer profile has none enabled.")
	}
	if len(reasons) == 0 {
		pref := config.Preferences{SchemaVersion: "1.0", Mode: "balanced", Budget: 64, MinWalls: 2, MinInfill: .1, Supports: true}
		if e := config.Compose(p, f, pref); e != nil {
			reasons = append(reasons, e.Error())
		}
	}
	return MaterialChoice{Profile: f, Compatible: len(reasons) == 0, Reasons: reasons}
}
func (s *Service) materials(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	printerID := r.URL.Query().Get("printerId")
	if printerID == "" {
		printerID = s.activePrinterID
	}
	var p *config.Printer
	for i := range s.printers {
		if s.printers[i].ID == printerID {
			p = &s.printers[i]
			break
		}
	}
	if p == nil {
		errorOut(w, 404, errors.New("import and select your printer profile before checking materials"))
		return
	}
	choices := make([]MaterialChoice, 0, len(s.filaments))
	for _, f := range s.filaments {
		choices = append(choices, materialChoice(*p, f))
	}
	jsonOut(w, 200, choices)
}
func (s *Service) selectPrinter(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if !decode(w, r, &req) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		errorOut(w, 409, errors.New("wait for slicing to finish before changing printers"))
		return
	}
	for _, p := range s.printers {
		if p.ID == req.ID {
			if e := s.activatePrinterLocked(p.ID); e != nil {
				errorOut(w, 500, e)
				return
			}
			jsonOut(w, 200, map[string]string{"activePrinterId": p.ID})
			return
		}
	}
	errorOut(w, 404, errors.New("printer profile has not been imported"))
}
func (s *Service) activatePrinterLocked(printerID string) error {
	if e := os.WriteFile(filepath.Join(s.dataDir, "active-printer.txt"), []byte(printerID), 0600); e != nil {
		return e
	}
	s.activePrinterID = printerID
	return nil
}
func (s *Service) importModel(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, geometry.MaxBytes+(1<<20))
	if e := r.ParseMultipartForm(8 << 20); e != nil {
		errorOut(w, 400, e)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	f, h, e := r.FormFile("file")
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, geometry.MaxBytes+1))
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	m, e := geometry.Parse(h.Filename, data)
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	model := Model{ID: id(), Name: filepath.Base(h.Filename), Triangles: len(m.Triangles), Bounds: m.Bounds, Volume: m.Volume, Area: m.Area, Warnings: m.Warnings, mesh: m}
	s.mu.Lock()
	if len(s.models) >= 6 {
		s.mu.Unlock()
		errorOut(w, 409, errors.New("six models are already loaded; restart to release them"))
		return
	}
	total := len(m.Triangles)
	for _, loaded := range s.models {
		total += loaded.Triangles
	}
	if total > 1000000 {
		s.mu.Unlock()
		errorOut(w, 409, errors.New("session exceeds one million mesh triangles; restart to release models"))
		return
	}
	s.models[model.ID] = model
	s.mu.Unlock()
	jsonOut(w, 201, model)
}
func (s *Service) mesh(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	m, ok := s.models[r.PathValue("id")]
	s.mu.RUnlock()
	if !ok {
		errorOut(w, 404, errors.New("model not found"))
		return
	}
	jsonOut(w, 200, map[string]any{"vertices": m.mesh.Vertices()})
}
func (s *Service) startJob(w http.ResponseWriter, r *http.Request) {
	var req JobRequest
	if !decode(w, r, &req) {
		return
	}
	p, e := config.ParsePrinter(req.Printer)
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	f, e := config.ParseFilament(req.Filament)
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	pref := config.Preferences{SchemaVersion: "1.0", Mode: req.Mode, Budget: req.Budget, MinWalls: req.MinWalls, MinInfill: req.MinInfill, Supports: true, LoadDirection: req.LoadDirection}
	pref.SupportMode, pref.SupportStyle, pref.SupportPlacement, pref.SupportAngle = req.SupportMode, req.SupportStyle, req.SupportPlacement, req.SupportAngle
	if pref.Budget == 0 {
		pref.Budget = 180
	}
	if pref.MinWalls == 0 {
		pref.MinWalls = 2
	}
	if req.Supports != nil {
		pref.Supports = *req.Supports
	}
	if e = config.Compose(p, f, pref); e != nil {
		errorOut(w, 400, e)
		return
	}
	s.mu.Lock()
	registered := false
	for _, stored := range s.printers {
		if stored.ID == p.ID {
			a, _ := json.Marshal(stored)
			b, _ := json.Marshal(p)
			registered = string(a) == string(b)
			break
		}
	}
	if !registered {
		s.mu.Unlock()
		errorOut(w, 409, errors.New("import and save this printer profile before slicing; save any changed machine values first"))
		return
	}
	if s.active {
		s.mu.Unlock()
		errorOut(w, 409, errors.New("one slicing job is already running; cancel or wait for completion"))
		return
	}
	m, ok := s.models[req.ModelID]
	if !ok {
		s.mu.Unlock()
		errorOut(w, 404, errors.New("model not found"))
		return
	}
	backend := s.backend
	if backend == nil {
		s.mu.Unlock()
		errorOut(w, 409, errors.New("configure the PrusaSlicer executable in Settings first"))
		return
	}
	if req.Transform == nil {
		req.Transform = &geometry.Transform{Scale: geometry.Vec{1, 1, 1}, Position: geometry.Vec{p.BuildVolume.Width / 2, p.BuildVolume.Depth / 2, 0}}
	}
	ctx, cancel := context.WithCancel(s.ctx)
	j := &Job{ID: id(), Status: "queued", Message: "Preparing geometry", Progress: 0, cancel: cancel}
	j.dir = filepath.Join(s.tempDir, j.ID)
	s.jobs[j.ID] = j
	s.order = append(s.order, j.ID)
	s.active = true
	// Retain at most three completed jobs and their preview buffers.
	for len(s.order) > 3 {
		old := s.order[0]
		s.order = s.order[1:]
		if o := s.jobs[old]; o != nil {
			delete(s.jobs, old)
			_ = os.RemoveAll(o.dir)
		}
	}
	snapshot := *j
	s.wg.Add(1)
	s.mu.Unlock()
	go func() {
		defer s.wg.Done()
		defer cancel()
		defer func() { s.mu.Lock(); s.active = false; s.mu.Unlock() }()
		s.run(ctx, j, m, *req.Transform, p, f, pref, backend)
	}()
	jsonOut(w, 202, snapshot)
}
func (s *Service) update(j *Job, status string, progress float64, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j.Status = status
	if progress > j.Progress {
		j.Progress = progress
	}
	j.Message = message
}
func (s *Service) fail(j *Job, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(e, context.Canceled) {
		j.Status = "cancelled"
		j.Message = "Slicing cancelled"
	} else {
		j.Status = "failed"
		j.Message = "Slicing failed"
		j.Error = e.Error()
		slog.Error("job failed", "job", j.ID, "error", e)
	}
}
func metrics(p *gcode.Preview, t time.Duration) *Metrics {
	return &Metrics{Seconds: p.EstimatedSeconds, MassG: p.MassG, FilamentMM: p.FilamentMM, PathMM: p.PathMM, SupportFilamentMM: p.SupportFilamentMM, SliceSeconds: t.Seconds()}
}
func (s *Service) run(ctx context.Context, j *Job, model Model, t geometry.Transform, p config.Printer, f config.Filament, pref config.Preferences, backend engine.Backend) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	s.update(j, "optimizing", .03, "Analyzing transformed mesh and machine constraints")
	m, e := model.mesh.Transformed(t)
	if e != nil {
		s.fail(j, e)
		return
	}
	for k, limit := range []float64{p.BuildVolume.Width, p.BuildVolume.Depth, p.BuildVolume.Height} {
		if m.Bounds.Min[k] < -1e-5 || m.Bounds.Max[k] > limit+1e-5 {
			s.fail(j, fmt.Errorf("transformed model exceeds build volume on axis %d", k))
			return
		}
	}
	for _, r := range p.BuildVolume.ExcludedRegions {
		if m.Bounds.Min[0] < r.X+r.Width && m.Bounds.Max[0] > r.X && m.Bounds.Min[1] < r.Y+r.Depth && m.Bounds.Max[1] > r.Y {
			s.fail(j, errors.New("model bounding box overlaps an excluded region"))
			return
		}
	}
	angle := pref.SupportAngle
	if angle == 0 {
		angle = 45
	}
	readiness, e := s.analyze(ctx, m, p.Extruders[0].NozzleDiameter, angle)
	if e != nil {
		s.fail(j, e)
		return
	}
	mode := optimizer.SupportMode(pref)
	enabled := mode != "off" && (mode != "auto" || readiness.SupportsRecommended)
	area := readiness.UnsupportedAreaMm2
	if enabled && area == 0 {
		area = m.OverhangArea
	}
	if !enabled {
		area = 0
	}
	g := optimizer.Geometry{Width: m.Bounds.Max[0] - m.Bounds.Min[0], Depth: m.Bounds.Max[1] - m.Bounds.Min[1], Height: m.Bounds.Max[2] - m.Bounds.Min[2], Volume: m.Volume, Area: m.Area, OverhangArea: area, MinFeature: m.MinFeature}
	opt, e := optimizer.Optimize(ctx, g, p, f, pref)
	if e != nil {
		s.fail(j, e)
		return
	}
	opt.Selected.Supports = enabled
	opt.Baseline.Supports = enabled
	opt.Baseline.SupportStyle = "grid"
	opt.Predicted = optimizer.EstimatePrint(g, opt.Selected, p, f, pref.LoadDirection)
	opt.BaselinePredicted = optimizer.EstimatePrint(g, opt.Baseline, p, f, pref.LoadDirection)
	report := &Report{Report: opt, Readiness: &readiness, Warnings: append([]string{}, m.Warnings...)}
	if readiness.SupportsRecommended && !enabled {
		report.Warnings = append(report.Warnings, "Geometry analysis recommends supports, but support mode is off. Review the unsupported regions before printing.")
	}
	if enabled && pref.SupportPlacement == "buildplate" {
		report.Warnings = append(report.Warnings, "Build-plate-only supports cannot reach every enclosed or obstructed overhang. Inspect support paths before printing.")
	}
	report.Warnings = append(report.Warnings, f.Notes...)
	report.Warnings = append(report.Warnings, "Baseline and selected results are sliced independently; comparisons use native G-code time estimates and calculated extrusion mass.")
	if strings.Contains(strings.ToLower(p.Name), "simulation") {
		report.Warnings = append(report.Warnings, "SIMULATION PROFILE: replace all fictional machine values with verified hardware data before printing.")
	}
	chosen, preview, e := s.sliceSupportCandidates(ctx, j, m, p, f, pref, opt.Selected, backend, report)
	if e != nil {
		s.fail(j, e)
		return
	}
	report.Actual = metrics(preview, chosen.Elapsed)
	s.update(j, "slicing", .75, "Slicing baseline for a calculated comparison")
	baseline, e := backend.Slice(ctx, engine.Request{Mesh: m, Printer: p, Filament: f, Settings: opt.Baseline, Directory: filepath.Join(j.dir, "baseline")}, func(_ float64, msg string) { s.update(j, "slicing", .8, msg) })
	if e == nil {
		var bp *gcode.Preview
		var bf *os.File
		bf, e = os.Open(baseline.Path)
		if e == nil {
			bp, e = gcode.Validate(ctx, bf, p, f)
			bf.Close()
		}
		if e == nil {
			report.BaselineActual = metrics(bp, baseline.Elapsed)
			td := (bp.EstimatedSeconds - preview.EstimatedSeconds) / bp.EstimatedSeconds * 100
			md := (bp.MassG - preview.MassG) / bp.MassG * 100
			report.TimeDifferencePercent = &td
			report.MassDifferencePercent = &md
		}
	}
	if ctx.Err() != nil {
		s.fail(j, ctx.Err())
		return
	}
	if e != nil {
		report.Warnings = append(report.Warnings, "Baseline comparison unavailable: "+e.Error())
	}
	report.Warnings = append(report.Warnings, preview.Warnings...)
	s.mu.Lock()
	j.Status = "complete"
	j.Progress = 1
	j.Message = "Validated G-code ready to preview and export"
	j.Report = report
	j.preview = preview
	j.path = chosen.Path
	j.PreviewReady = true
	s.mu.Unlock()
}
func (s *Service) job(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[r.PathValue("id")]
	if !ok {
		errorOut(w, 404, errors.New("job not found"))
		return
	}
	jsonOut(w, 200, j)
}
func (s *Service) cancelJob(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	j, ok := s.jobs[r.PathValue("id")]
	if ok {
		j.cancel()
	}
	s.mu.RUnlock()
	if !ok {
		errorOut(w, 404, errors.New("job not found"))
		return
	}
	jsonOut(w, 200, map[string]bool{"cancelled": true})
}
func (s *Service) preview(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	j, ok := s.jobs[r.PathValue("id")]
	if !ok || !j.PreviewReady {
		s.mu.RUnlock()
		errorOut(w, 409, errors.New("validated preview is not ready"))
		return
	}
	p := j.preview
	s.mu.RUnlock()
	jsonOut(w, 200, p)
}
func (s *Service) export(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	j, ok := s.jobs[r.PathValue("id")]
	if !ok || !j.PreviewReady || j.Status != "complete" {
		s.mu.RUnlock()
		errorOut(w, 409, errors.New("export requires successful G-code validation"))
		return
	}
	path := j.path
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="forge-`+j.ID+`.gcode"`)
	http.ServeFile(w, r, path)
}

type profileRequest struct {
	Kind    string          `json:"kind"`
	Profile json.RawMessage `json:"profile"`
}

func ValidateProfile(kind string, b []byte) error {
	switch kind {
	case "printer":
		_, e := config.ParsePrinter(b)
		return e
	case "filament":
		_, e := config.ParseFilament(b)
		return e
	case "optimization":
		_, e := config.ParsePreferences(b)
		return e
	default:
		return errors.New("unknown profile kind")
	}
}
func (s *Service) validateProfile(w http.ResponseWriter, r *http.Request) {
	var req profileRequest
	if !decode(w, r, &req) {
		return
	}
	errs := []string{}
	if e := ValidateProfile(req.Kind, req.Profile); e != nil {
		errs = append(errs, e.Error())
	}
	jsonOut(w, 200, map[string]any{"valid": len(errs) == 0, "errors": errs})
}
func (s *Service) saveProfile(w http.ResponseWriter, r *http.Request) {
	var req profileRequest
	if !decode(w, r, &req) {
		return
	}
	if e := ValidateProfile(req.Kind, req.Profile); e != nil {
		errorOut(w, 400, e)
		return
	}
	if req.Kind == "optimization" {
		pref, _ := config.ParsePreferences(req.Profile)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.active {
			errorOut(w, 409, errors.New("wait for slicing to finish before changing preferences"))
			return
		}
		if e := os.WriteFile(filepath.Join(s.dataDir, "preferences.json"), req.Profile, 0600); e != nil {
			errorOut(w, 500, e)
			return
		}
		s.preferences = &pref
		jsonOut(w, 200, pref)
		return
	}
	var profileID string
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		errorOut(w, 409, errors.New("wait for slicing to finish before changing profiles"))
		return
	}
	if req.Kind == "printer" {
		p, _ := config.ParsePrinter(req.Profile)
		profileID = p.ID
	} else {
		f, _ := config.ParseFilament(req.Profile)
		profileID = f.ID
	}
	// IDs pass schema regex, then are checked again before becoming filenames.
	if profileID == "" || strings.ContainsAny(profileID, "/\\: ") {
		errorOut(w, 400, errors.New("invalid profile ID for storage"))
		return
	}
	target := filepath.Join(s.dataDir, "profiles", req.Kind+"-"+profileID+".json")
	if e := os.WriteFile(target, req.Profile, 0600); e != nil {
		errorOut(w, 500, e)
		return
	}
	if req.Kind == "printer" {
		p, _ := config.ParsePrinter(req.Profile)
		s.upsertPrinter(p)
		if e := s.activatePrinterLocked(p.ID); e != nil {
			errorOut(w, 500, e)
			return
		}
	} else {
		f, _ := config.ParseFilament(req.Profile)
		s.upsertFilament(f)
	}
	var out any
	_ = json.Unmarshal(req.Profile, &out)
	jsonOut(w, 200, out)
}
func (s *Service) upsertPrinter(p config.Printer) {
	for i, v := range s.printers {
		if v.ID == p.ID {
			s.printers[i] = p
			return
		}
	}
	s.printers = append(s.printers, p)
}
func (s *Service) upsertFilament(f config.Filament) {
	for i, v := range s.filaments {
		if v.ID == f.ID {
			s.filaments[i] = f
			return
		}
	}
	s.filaments = append(s.filaments, f)
}
func (s *Service) setEngine(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if !decode(w, r, &req) {
		return
	}
	p, e := engine.New(req.Path)
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		errorOut(w, 409, errors.New("wait for the active slicing job before changing engine"))
		return
	}
	if e = os.WriteFile(filepath.Join(s.dataDir, "engine-path.txt"), []byte(p.Status().Path), 0600); e != nil {
		errorOut(w, 500, e)
		return
	}
	s.backend = p
	jsonOut(w, 200, p.Status())
}
func (s *Service) SetSaveDialog(fn func(string) (string, error)) { s.saveDialog = fn }
func (s *Service) saveNative(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	j, ok := s.jobs[r.PathValue("id")]
	dialog := s.saveDialog
	if !ok || !j.PreviewReady || j.Status != "complete" {
		s.mu.RUnlock()
		errorOut(w, 409, errors.New("validated output is not ready"))
		return
	}
	src := j.path
	name := "forge-" + j.ID + ".gcode"
	s.mu.RUnlock()
	if dialog == nil {
		errorOut(w, 409, errors.New("native file dialog is available only in desktop mode"))
		return
	}
	dest, e := dialog(name)
	if e != nil {
		errorOut(w, 500, e)
		return
	}
	if dest == "" {
		jsonOut(w, 200, map[string]bool{"cancelled": true})
		return
	}
	in, e := os.Open(src)
	if e != nil {
		errorOut(w, 500, e)
		return
	}
	defer in.Close()
	out, e := os.Create(dest)
	if e != nil {
		errorOut(w, 500, e)
		return
	}
	_, e = io.Copy(out, in)
	ce := out.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		errorOut(w, 500, e)
		return
	}
	jsonOut(w, 200, map[string]any{"saved": true, "path": dest})
}
