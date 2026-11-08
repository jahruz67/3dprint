package service

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/forge-slicer/forge/internal/analysis"
	"github.com/forge-slicer/forge/internal/geometry"
)

func (s *Service) analyze(ctx context.Context, m *geometry.Mesh, nozzle, angle float64) (analysis.Report, error) {
	select {
	case s.analysisSlots <- struct{}{}:
		defer func() { <-s.analysisSlots }()
	case <-ctx.Done():
		return analysis.Report{}, ctx.Err()
	}
	return analysis.Analyze(ctx, m, nozzle, angle)
}

func (s *Service) analyzeModel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PrinterID    string              `json:"printerId"`
		Transform    *geometry.Transform `json:"transform"`
		SupportAngle float64             `json:"supportAngle"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.SupportAngle == 0 {
		req.SupportAngle = 45
	}
	if math.IsNaN(req.SupportAngle) || math.IsInf(req.SupportAngle, 0) || req.SupportAngle < 20 || req.SupportAngle > 80 {
		errorOut(w, 400, errors.New("support angle must be 20–80 degrees"))
		return
	}
	s.mu.RLock()
	model, found := s.models[r.PathValue("id")]
	printerID := req.PrinterID
	if printerID == "" {
		printerID = s.activePrinterID
	}
	var nozzle, width, depth float64
	for _, p := range s.printers {
		if p.ID == printerID {
			nozzle = p.Extruders[0].NozzleDiameter
			width = p.BuildVolume.Width
			depth = p.BuildVolume.Depth
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		errorOut(w, 404, errors.New("model not found"))
		return
	}
	if nozzle == 0 {
		errorOut(w, 409, errors.New("import and select a printer profile before analyzing print readiness"))
		return
	}
	if req.Transform == nil {
		req.Transform = &geometry.Transform{Scale: geometry.Vec{1, 1, 1}, Position: geometry.Vec{width / 2, depth / 2, 0}}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	m, e := model.mesh.Transformed(*req.Transform)
	if e != nil {
		errorOut(w, 400, e)
		return
	}
	report, e := s.analyze(ctx, m, nozzle, req.SupportAngle)
	if e != nil {
		errorOut(w, 422, e)
		return
	}
	jsonOut(w, 200, report)
}
