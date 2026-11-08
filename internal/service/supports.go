package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/engine"
	"github.com/forge-slicer/forge/internal/gcode"
	"github.com/forge-slicer/forge/internal/geometry"
	"github.com/forge-slicer/forge/internal/optimizer"
)

func supportCost(p *gcode.Preview, reference *gcode.Preview) float64 {
	return .75*p.SupportFilamentMM/math.Max(1, reference.SupportFilamentMM) + .25*p.EstimatedSeconds/math.Max(1, reference.EstimatedSeconds)
}

func (s *Service) sliceSupportCandidates(ctx context.Context, j *Job, m *geometry.Mesh, p config.Printer, f config.Filament, pref config.Preferences, settings optimizer.Settings, backend engine.Backend, report *Report) (engine.Result, *gcode.Preview, error) {
	variants := optimizer.SupportVariants(settings, pref)
	var chosen engine.Result
	var preview, reference *gcode.Preview
	best := math.Inf(1)
	selectedIndex := -1
	for i, v := range variants {
		if e := ctx.Err(); e != nil {
			return chosen, nil, e
		}
		s.update(j, "slicing", .15+.45*float64(i)/float64(len(variants)), fmt.Sprintf("Slicing support candidate %d/%d: %s", i+1, len(variants), v.SupportStyle))
		result, e := backend.Slice(ctx, engine.Request{Mesh: m, Printer: p, Filament: f, Settings: v, Directory: filepath.Join(j.dir, fmt.Sprintf("selected-%d", i))}, func(_ float64, msg string) { s.update(j, "slicing", .15+.45*float64(i)/float64(len(variants)), msg) })
		var candidate *gcode.Preview
		if e == nil {
			var file *os.File
			file, e = os.Open(result.Path)
			if e == nil {
				candidate, e = gcode.Validate(ctx, file, p, f)
				file.Close()
			}
		}
		if e != nil {
			if ctx.Err() != nil {
				return chosen, nil, ctx.Err()
			}
			report.Warnings = append(report.Warnings, fmt.Sprintf("Support candidate %s rejected: %v", v.SupportStyle, e))
			continue
		}
		if reference == nil {
			reference = candidate
		}
		index := len(report.SupportCandidates)
		if settings.Supports {
			report.SupportCandidates = append(report.SupportCandidates, SupportCandidate{Style: v.SupportStyle, Seconds: candidate.EstimatedSeconds, MassG: candidate.MassG, SupportFilamentMM: candidate.SupportFilamentMM})
		}
		score := supportCost(candidate, reference)
		if score < best-1e-10 {
			best = score
			chosen = result
			preview = candidate
			selectedIndex = index
			report.Selected = v
		}
	}
	if preview == nil {
		return chosen, nil, fmt.Errorf("no support candidate produced valid G-code; %v", report.Warnings)
	}
	if settings.Supports {
		report.SupportCandidates[selectedIndex].Selected = true
		if len(variants) > 1 {
			report.SupportSelectionReason = fmt.Sprintf("Attempted organic and snug with identical model, layers, walls, infill and speeds; %d of %d candidates passed validation. Selected the lowest 75%% support-filament + 25%% print-time score, normalized to the first valid candidate. Removal force and print success are not measured.", len(report.SupportCandidates), len(variants))
		} else {
			report.SupportSelectionReason = "Applied your selected support recipe and shape. Native slicing determines local support regions and omits recognized bridges; inspect the support toolpaths before printing."
		}
	} else {
		report.SupportSelectionReason = "Supports disabled by your choice or Auto geometry recommendation; no support-shape comparison was needed."
	}
	return chosen, preview, nil
}
