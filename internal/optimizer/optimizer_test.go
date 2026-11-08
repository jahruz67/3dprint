package optimizer

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/forge-slicer/forge/internal/config"
)

func cube() Geometry {
	return Geometry{Width: 20, Depth: 20, Height: 20, Volume: 8000, Area: 2400, OverhangArea: 0, MinFeature: 1}
}
func prefs(mode string) config.Preferences {
	return config.Preferences{SchemaVersion: "1.0", Mode: mode, Budget: 256, MinWalls: 2, MinInfill: .12, Supports: true}
}

func TestSearchDeterminismAndPhysicalLimits(t *testing.T) {
	p, f := config.ReferenceFixture()
	p[0].Extruders[0].FlowVerified = false
	p[0].Extruders[0].MaxVolumetricFlow = 4
	for _, mode := range []string{"fast", "balanced", "quality", "strength", "material"} {
		t.Run(mode, func(t *testing.T) {
			a, e := Optimize(context.Background(), cube(), p[0], f[0], prefs(mode))
			if e != nil {
				t.Fatal(e)
			}
			b, e := Optimize(context.Background(), cube(), p[0], f[0], prefs(mode))
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a, b) {
				t.Fatal("nondeterministic result")
			}
			s := a.Selected
			if a.Evaluated > 256 || a.Evaluated < 16 || a.ParetoCount < 1 {
				t.Fatal("invalid bounded search accounting")
			}
			if s.Speed*s.LineWidth*s.LayerHeight > 4+1e-9 {
				t.Fatal("flow limit exceeded")
			}
			if s.Speed > 120 || s.TravelSpeed > 180 || s.Acceleration > 1000 || s.Walls < 2 || s.Infill < .12 {
				t.Fatal("machine or functional constraint violated")
			}
			if Objective(mode, a.Predicted, a.BaselinePredicted) > Objective(mode, a.BaselinePredicted, a.BaselinePredicted)+1e-9 {
				t.Fatal("selected solution regressed its objective")
			}
		})
	}
}

func TestObjectivesProduceDifferentDecisions(t *testing.T) {
	p, f := config.ReferenceFixture()
	seen := map[Settings]bool{}
	var fast, quality, material, strength Report
	for _, mode := range []string{"fast", "balanced", "quality", "strength", "material"} {
		r, e := Optimize(context.Background(), cube(), p[0], f[0], prefs(mode))
		if e != nil {
			t.Fatal(e)
		}
		seen[r.Selected] = true
		switch mode {
		case "fast":
			fast = r
		case "quality":
			quality = r
		case "material":
			material = r
		case "strength":
			strength = r
		}
	}
	if len(seen) < 4 {
		t.Fatalf("only %d distinct decisions", len(seen))
	}
	if quality.Predicted.QualityError >= fast.Predicted.QualityError {
		t.Fatal("quality objective did not reduce quality error")
	}
	if material.Predicted.MassG >= strength.Predicted.MassG {
		t.Fatal("material objective failed to use less estimated material than strength")
	}
	if strength.Predicted.StrengthProxy <= material.Predicted.StrengthProxy {
		t.Fatal("strength proxy did not increase")
	}
}

func TestCancellationAndInfeasibleGeometry(t *testing.T) {
	p, f := config.ReferenceFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := Optimize(ctx, cube(), p[0], f[0], prefs("balanced")); e != context.Canceled {
		t.Fatalf("cancellation lost: %v", e)
	}
	g := cube()
	g.Width = 221
	if _, e := Optimize(context.Background(), g, p[0], f[0], prefs("fast")); e == nil {
		t.Fatal("oversized geometry accepted")
	}
	g = cube()
	g.Volume = math.NaN()
	if _, e := Optimize(context.Background(), g, p[0], f[0], prefs("balanced")); e == nil {
		t.Fatal("nonfinite geometry accepted")
	}
}

func TestAccelerationModelAndMass(t *testing.T) {
	p, f := config.ReferenceFixture()
	s := Settings{LayerHeight: .2, LineWidth: .44, Walls: 2, Infill: .2, InfillPattern: "grid", Speed: 50, TravelSpeed: 180, Acceleration: 1000}
	a := EstimatePrint(cube(), s, p[0], f[0], [3]float64{})
	s.Acceleration = 100
	b := EstimatePrint(cube(), s, p[0], f[0], [3]float64{})
	if b.Seconds <= a.Seconds {
		t.Fatal("lower acceleration must increase estimated time")
	}
	if math.Abs(a.MassG-a.VolumeMm3*f[0].Density/1000) > 1e-10 {
		t.Fatal("density conversion incorrect")
	}
}

func BenchmarkOptimizeCube(b *testing.B) {
	p, f := config.ReferenceFixture()
	for i := 0; i < b.N; i++ {
		if _, e := Optimize(context.Background(), cube(), p[0], f[0], prefs("balanced")); e != nil {
			b.Fatal(e)
		}
	}
}
