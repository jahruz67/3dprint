// Package optimizer performs deterministic bounded multi-objective search.
// All estimates are analytical engineering approximations, not physical guarantees.
package optimizer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/forge-slicer/forge/internal/config"
)

type Geometry struct{ Width, Depth, Height, Volume, Area, OverhangArea, MinFeature float64 }
type Settings struct {
	LayerHeight             float64    `json:"layerHeight"`
	LineWidth               float64    `json:"lineWidth"`
	Walls                   int        `json:"walls"`
	Infill                  float64    `json:"infill"`
	InfillPattern           string     `json:"infillPattern"`
	Speed                   float64    `json:"speed"`
	TravelSpeed             float64    `json:"travelSpeed"`
	Acceleration            float64    `json:"acceleration"`
	Supports                bool       `json:"supports"`
	SupportStyle            string     `json:"supportStyle"`
	SupportMode             string     `json:"supportMode"`
	SupportContactZ         float64    `json:"supportContactZ"`
	SupportXYDistance       float64    `json:"supportXYDistance"`
	SupportSpacing          float64    `json:"supportSpacing"`
	SupportInterfaceSpacing float64    `json:"supportInterfaceSpacing"`
	SupportInterfaceLayers  int        `json:"supportInterfaceLayers"`
	SupportSheath           bool       `json:"supportSheath"`
	SupportBuildplateOnly   bool       `json:"supportBuildplateOnly"`
	SupportThreshold        float64    `json:"supportThreshold"`
	Rotation                [3]float64 `json:"rotation"`
}
type Estimate struct {
	Seconds       float64 `json:"seconds"`
	MassG         float64 `json:"massG"`
	VolumeMm3     float64 `json:"volumeMm3"`
	QualityError  float64 `json:"qualityError"`
	StrengthProxy float64 `json:"strengthProxy"`
}
type Report struct {
	Mode              string   `json:"mode"`
	Selected          Settings `json:"selected"`
	Baseline          Settings `json:"baseline"`
	Predicted         Estimate `json:"predicted"`
	BaselinePredicted Estimate `json:"baselinePredicted"`
	Evaluated         int      `json:"evaluated"`
	ParetoCount       int      `json:"paretoCount"`
	Constraints       []string `json:"constraints"`
	Assumptions       []string `json:"assumptions"`
	Changes           []string `json:"changes"`
}
type candidate struct {
	s     Settings
	e     Estimate
	score float64
}
type search struct {
	ctx      context.Context
	g        Geometry
	p        config.Printer
	f        config.Filament
	pref     config.Preferences
	baseline Estimate
	seen     map[Settings]bool
	all      []candidate
	best     candidate
	budget   int
}

func finite(x float64) bool         { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func clamp(x, a, b float64) float64 { return math.Max(a, math.Min(b, x)) }

// EstimatePrint computes a closed-form shell/infill/support volume approximation
// and acceleration-aware path-time estimate. It is deliberately separate from
// measured toolpath metrics returned by the slicing engine.
func EstimatePrint(g Geometry, s Settings, p config.Printer, f config.Filament, load [3]float64) Estimate {
	// Shell volume first-order surface offset, capped by actual solid volume.
	wall := float64(s.Walls) * s.LineWidth
	topBottom := 2 * g.Width * g.Depth * math.Max(.8, 3*s.LayerHeight)
	sideArea := math.Max(0, g.Area-2*g.Width*g.Depth)
	shell := math.Min(g.Volume, sideArea*wall+topBottom)
	core := math.Max(0, g.Volume-shell)
	supports := 0.0
	if s.Supports {
		supports = g.OverhangArea * g.Height * .18 * .12
	}
	v := shell + core*s.Infill + supports
	// Bead cross-section is a rectangle with semicircular ends.
	cross := s.LayerHeight*(s.LineWidth-s.LayerHeight) + math.Pi*s.LayerHeight*s.LayerHeight/4
	length := v / math.Max(.001, cross)
	characteristic := math.Max(2, math.Min(g.Width, g.Depth)*.55)
	// Exact trapezoidal/triangular rest-to-rest time for the representative segment.
	rampDistance := s.Speed * s.Speed / s.Acceleration
	segmentTime := 0.0
	if characteristic >= rampDistance {
		segmentTime = characteristic/s.Speed + s.Speed/s.Acceleration
	} else {
		segmentTime = 2 * math.Sqrt(characteristic/s.Acceleration)
	}
	printSeconds := length / characteristic * segmentTime
	layers := math.Ceil(g.Height / s.LayerHeight)
	travelLength := layers * (g.Width + g.Depth) * (.8 + 2*s.Infill)
	seconds := printSeconds + travelLength/s.TravelSpeed + layers*.6 + 90
	// Chord/staircase error proxy plus flow/speed variation. Units are mm proxy.
	detailPenalty := 0.0
	if g.MinFeature > 0 {
		detailPenalty = math.Max(0, s.LineWidth-g.MinFeature) * .7
	}
	q := s.LayerHeight*.55 + s.LineWidth*.08 + math.Pow(s.Speed/math.Max(1, p.Motion.MaxSpeed), 2)*.055 + detailPenalty
	// Relative shell/infill/load-direction proxy; no constitutive material model.
	loadNorm := math.Sqrt(load[0]*load[0] + load[1]*load[1] + load[2]*load[2])
	anisotropy := .78
	if loadNorm > 0 {
		anisotropy = 1 - .45*math.Abs(load[2])/loadNorm
	}
	bonding := clamp(1-s.LayerHeight/p.Extruders[0].NozzleDiameter*.22, .6, 1)
	strength := (1 - math.Exp(-float64(s.Walls)*.42)) * (.45 + .55*math.Sqrt(s.Infill)) * anisotropy * bonding
	return Estimate{Seconds: seconds, MassG: v * f.Density / 1000, VolumeMm3: v, QualityError: q, StrengthProxy: strength}
}

// Objective returns normalized, explicitly documented objective values.
// The tiny secondary terms deterministically break otherwise equivalent choices.
func Objective(mode string, e, b Estimate) float64 {
	t := e.Seconds / math.Max(1, b.Seconds)
	m := e.MassG / math.Max(.001, b.MassG)
	q := e.QualityError / math.Max(.001, b.QualityError)
	s := b.StrengthProxy / math.Max(.001, e.StrengthProxy)
	switch mode {
	case "fast":
		return t + .015*m + .01*q
	case "material":
		return m + .02*t + .01*q
	case "quality":
		return q + .025*t + .015*m
	case "strength":
		return s + .025*m + .015*t
	default:
		return .4*t + .25*m + .25*q + .1*s
	}
}

func Optimize(ctx context.Context, g Geometry, p config.Printer, f config.Filament, pref config.Preferences) (Report, error) {
	var report Report
	if err := config.Compose(p, f, pref); err != nil {
		return report, err
	}
	for _, v := range []float64{g.Width, g.Depth, g.Height, g.Volume, g.Area, g.OverhangArea, g.MinFeature} {
		if !finite(v) || v < 0 {
			return report, errors.New("geometry metrics must be finite and nonnegative")
		}
	}
	if g.Width <= 0 || g.Depth <= 0 || g.Height <= 0 || g.Volume <= 0 || g.Area <= 0 {
		return report, errors.New("optimization requires a nondegenerate closed model")
	}
	if g.Width > p.BuildVolume.Width || g.Depth > p.BuildVolume.Depth || g.Height > p.BuildVolume.Height {
		return report, errors.New("model exceeds printer build volume")
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}
	nozzle := p.Extruders[0].NozzleDiameter
	lo := math.Max(f.MinLayerHeight, nozzle*.2)
	hi := math.Min(f.MaxLayerHeight, nozzle*.8)
	base := Settings{LayerHeight: clamp(.5*nozzle, lo, hi), LineWidth: 1.1 * nozzle, Walls: maxInt(2, pref.MinWalls), Infill: math.Max(.2, pref.MinInfill), InfillPattern: "grid", Speed: math.Min(50, p.Motion.MaxSpeed), TravelSpeed: p.Motion.MaxTravelSpeed, Acceleration: p.Motion.MaxAcceleration, Supports: pref.Supports && g.OverhangArea > 0, SupportStyle: "grid"}
	base = SupportPolicy(base, pref, nozzle)
	base = limit(base, p, f)
	b := EstimatePrint(g, base, p, f, pref.LoadDirection)
	ss := search{ctx: ctx, g: g, p: p, f: f, pref: pref, baseline: b, seen: make(map[Settings]bool), budget: pref.Budget, best: candidate{score: math.Inf(1)}}
	ss.evaluate(base)
	// Low-discrepancy seeding visits the whole bounded region without enumerating
	// its Cartesian product. A fixed sequence makes runs reproducible.
	seedCount := minInt(64, maxInt(12, pref.Budget/3))
	for i := 1; i <= seedCount && len(ss.all) < ss.budget; i++ {
		s := base
		s.LayerHeight = lo + (hi-lo)*radical(i, 2)
		s.LineWidth = nozzle * (1 + radical(i, 3)*.35)
		s.Walls = pref.MinWalls + int(radical(i, 5)*float64(maxInt(1, 7-pref.MinWalls)))
		s.Infill = pref.MinInfill + (math.Max(.65, pref.MinInfill)-pref.MinInfill)*radical(i, 7)
		s.Speed = math.Min(p.Motion.MaxSpeed, 20+radical(i, 11)*math.Max(0, p.Motion.MaxSpeed-20))
		patterns := []string{"grid", "gyroid", "rectilinear"}
		s.InfillPattern = patterns[(i-1)%len(patterns)]
		ss.evaluate(limit(s, p, f))
	}
	// Coordinate pattern search around the best candidate. Mesh halves on stalled
	// rounds, so work grows with the budget rather than all variable combinations.
	steps := []float64{(hi - lo) / 3, nozzle * .12, 1, .12, p.Motion.MaxSpeed * .25}
	for round := 0; round < 18 && len(ss.all) < ss.budget; round++ {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		before := ss.best.score
		center := ss.best.s
		for dim := 0; dim < 5 && len(ss.all) < ss.budget; dim++ {
			for _, sign := range []float64{-1, 1} {
				s := center
				switch dim {
				case 0:
					s.LayerHeight = clamp(s.LayerHeight+sign*steps[dim], lo, hi)
				case 1:
					s.LineWidth = clamp(s.LineWidth+sign*steps[dim], nozzle, nozzle*1.35)
				case 2:
					s.Walls = maxInt(pref.MinWalls, minInt(12, s.Walls+int(sign)))
				case 3:
					s.Infill = clamp(s.Infill+sign*steps[dim], pref.MinInfill, 1)
				case 4:
					s.Speed = clamp(s.Speed+sign*steps[dim], math.Min(10, p.Motion.MaxSpeed), p.Motion.MaxSpeed)
				}
				ss.evaluate(limit(s, p, f))
			}
		}
		if ss.best.score >= before-1e-8 {
			for i := range steps {
				steps[i] *= .5
			}
			if steps[0] < .0005 && steps[3] < .002 {
				break
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}
	if len(ss.all) == 0 {
		return report, errors.New("no feasible candidate within the configured limits")
	}
	frontier := pareto(ss.all)
	report = Report{Mode: pref.Mode, Selected: ss.best.s, Baseline: base, Predicted: ss.best.e, BaselinePredicted: b, Evaluated: len(ss.all), ParetoCount: len(frontier), Constraints: []string{
		fmt.Sprintf("Layer height %.3f–%.3f mm from material and 0.2–0.8 nozzle ratios", lo, hi),
		fmt.Sprintf("Flow ≤ %.2f mm³/s (minimum of configured hot-end and material)", math.Min(p.Extruders[0].MaxVolumetricFlow, f.MaxVolumetricFlow)),
		fmt.Sprintf("Print speed ≤ %.1f mm/s; travel ≤ %.1f mm/s; acceleration ≤ %.0f mm/s²", p.Motion.MaxSpeed, p.Motion.MaxTravelSpeed, p.Motion.MaxAcceleration),
		fmt.Sprintf("Functional minimum: %d walls, %.1f%% infill", pref.MinWalls, pref.MinInfill*100),
	}, Assumptions: []string{
		"Analytical shell/infill/support and representative acceleration estimates; actual sliced paths provide separate metrics.",
		"Bounded deterministic candidate search; neither a global optimum nor physical print quality is established.",
		"Strength is an unvalidated dimensionless shell/infill/bonding proxy, without measured material properties or load validation.",
		"Orientation is the user-transformed model; search does not invent an orientation-dependent support estimate.",
		"Constant layer height; support regions, bridge exclusion and organic branches are generated by the native slicer. Contact gaps trade underside finish for removal ease.",
		"Top and bottom shells require at least three layers and 0.8 mm thickness on each top/bottom side in the estimate.",
	}, Changes: changes(base, ss.best.s)}
	if g.OverhangArea > 0 && !pref.Supports {
		report.Assumptions = append(report.Assumptions, "Overhangs were detected while supports are disabled; review unsupported areas before printing.")
	}
	if pref.Mode == "strength" && pref.LoadDirection == [3]float64{} {
		report.Assumptions = append(report.Assumptions, "Load direction unspecified: assumed mixed loading with 0.78 interlayer factor; no strength guarantee.")
	}
	return report, nil
}

func limit(s Settings, p config.Printer, f config.Filament) Settings {
	// Use rectangular cross section for a conservative upper bound on flow.
	flow := math.Min(p.Extruders[0].MaxVolumetricFlow, f.MaxVolumetricFlow)
	s.Speed = math.Min(s.Speed, flow/(s.LineWidth*s.LayerHeight))
	s.Speed = math.Min(s.Speed, p.Motion.MaxSpeed)
	s.TravelSpeed = math.Min(s.TravelSpeed, p.Motion.MaxTravelSpeed)
	s.Acceleration = math.Min(s.Acceleration, p.Motion.MaxAcceleration)
	return s
}
func (s *search) evaluate(v Settings) {
	v = SupportPolicy(v, s.pref, s.p.Extruders[0].NozzleDiameter)
	if len(s.all) >= s.budget || s.ctx.Err() != nil || s.seen[v] {
		return
	}
	s.seen[v] = true
	if v.Speed <= 0 || v.LayerHeight <= 0 || v.LineWidth <= 0 || v.Walls < s.pref.MinWalls || v.Infill < s.pref.MinInfill {
		return
	}
	e := EstimatePrint(s.g, v, s.p, s.f, s.pref.LoadDirection)
	c := candidate{s: v, e: e, score: Objective(s.pref.Mode, e, s.baseline)}
	s.all = append(s.all, c)
	if c.score < s.best.score-1e-12 {
		s.best = c
	}
}
func dominates(a, b Estimate) bool {
	return a.Seconds <= b.Seconds && a.MassG <= b.MassG && a.QualityError <= b.QualityError && a.StrengthProxy >= b.StrengthProxy && (a.Seconds < b.Seconds || a.MassG < b.MassG || a.QualityError < b.QualityError || a.StrengthProxy > b.StrengthProxy)
}
func pareto(all []candidate) []candidate {
	// Incremental dominance pruning keeps only the non-dominated frontier.
	front := make([]candidate, 0, len(all)/2)
	sort.SliceStable(all, func(i, j int) bool { return all[i].e.Seconds < all[j].e.Seconds })
	for _, c := range all {
		dominated := false
		for _, f := range front {
			if dominates(f.e, c.e) {
				dominated = true
				break
			}
		}
		if dominated {
			continue
		}
		kept := front[:0]
		for _, f := range front {
			if !dominates(c.e, f.e) {
				kept = append(kept, f)
			}
		}
		front = append(kept, c)
	}
	return front
}
func radical(n, base int) float64 {
	value, factor := 0.0, 1.0/float64(base)
	for n > 0 {
		value += float64(n%base) * factor
		n /= base
		factor /= float64(base)
	}
	return value
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func changes(a, b Settings) []string {
	result := []string{}
	if math.Abs(a.LayerHeight-b.LayerHeight) > .0001 {
		result = append(result, fmt.Sprintf("Layer height %.3f → %.3f mm", a.LayerHeight, b.LayerHeight))
	}
	if math.Abs(a.LineWidth-b.LineWidth) > .0001 {
		result = append(result, fmt.Sprintf("Line width %.3f → %.3f mm", a.LineWidth, b.LineWidth))
	}
	if a.Walls != b.Walls {
		result = append(result, fmt.Sprintf("Perimeters %d → %d", a.Walls, b.Walls))
	}
	if math.Abs(a.Infill-b.Infill) > .0001 {
		result = append(result, fmt.Sprintf("Infill %.1f%% → %.1f%%", a.Infill*100, b.Infill*100))
	}
	if math.Abs(a.Speed-b.Speed) > .001 {
		result = append(result, fmt.Sprintf("Print speed %.2f → %.2f mm/s", a.Speed, b.Speed))
	}
	if a.InfillPattern != b.InfillPattern {
		result = append(result, "Infill pattern "+a.InfillPattern+" → "+b.InfillPattern)
	}
	if len(result) == 0 {
		result = append(result, "The defined baseline remained best within this search budget.")
	}
	return result
}
