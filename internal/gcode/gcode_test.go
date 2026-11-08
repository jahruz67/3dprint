package gcode

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/forge-slicer/forge/internal/config"
)

func program(p config.Printer, f config.Filament) string {
	return fmt.Sprintf("G21\nG90\nM82\nG28\nM190 S%g\nM109 S%g\nG92 E0\n;LAYER_CHANGE\n;Z:0.2\n;TYPE:External perimeter\nG1 X10 Y10 Z0.2 F1200\nG1 X20 E0.4\nM104 S0\nM140 S0\nM107\nM84\n", f.BedTemperature, f.NozzleTemperature)
}
func TestToolpathMetrics(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	v, e := Validate(context.Background(), strings.NewReader(program(p, f)), p, f)
	if e != nil {
		t.Fatal(e)
	}
	expected := .4 * math.Pi * math.Pow(p.Extruders[0].FilamentDiameter/2, 2) * f.Density / 1000
	if math.Abs(v.MassG-expected) > 1e-9 || v.FilamentMM != .4 || len(v.Layers) != 1 || len(v.Segments) != 18 || v.Segments[15] != 1 {
		t.Fatalf("bad metrics: %+v", v)
	}
}
func TestRejectsUnsafeStateAndLimits(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	base := program(p, f)
	cases := map[string]string{"bounds": strings.Replace(base, "X20 E0.4", "X999 E0.4", 1), "temperature": strings.Replace(base, fmt.Sprintf("M109 S%g", f.NozzleTemperature), "M109 S999", 1), "flow": strings.Replace(base, "E0.4", "E40", 1), "no wait": strings.Replace(base, "M109", "M104", 1), "no shutdown": strings.Replace(base, "M84", "", 1), "relative xyz": strings.Replace(base, "G90", "G91", 1), "missing extrusion mode": strings.Replace(base, "M82\n", "", 1), "unknown command": base + "M999\n", "nan": strings.Replace(base, "X20", "XNaN", 1), "unknown coordinate offset": base + "G92 X0\n"}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			if _, e := Validate(context.Background(), strings.NewReader(s), p, f); e == nil {
				t.Fatal("unsafe G-code accepted")
			}
		})
	}
}
func TestRetractionDoesNotInflateFilament(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	s := strings.Replace(program(p, f), "M104 S0", "G1 E0 F1200\nG1 E0.4 F1200\nM104 S0", 1)
	v, e := Validate(context.Background(), strings.NewReader(s), p, f)
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(v.FilamentMM-.4) > 1e-9 {
		t.Fatal("unretraction counted as deposition")
	}
}
func TestExcludedRegionCrossing(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	p.BuildVolume.ExcludedRegions = []config.Region{{X: 14, Y: 9, Width: 2, Depth: 2}}
	if _, e := Validate(context.Background(), strings.NewReader(program(p, f)), p, f); e == nil {
		t.Fatal("segment crossing excluded region accepted")
	}
}
func TestRelativeExtrusionAndCompactWords(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	s := strings.Replace(program(p, f), "M82", "M83", 1)
	s = strings.Replace(s, "G1 X20 E0.4", "G1X20E0.4", 1)
	if _, e := Validate(context.Background(), strings.NewReader(s), p, f); e != nil {
		t.Fatal(e)
	}
}
func TestFinalShutdownAndExtraAxes(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	p, f := ps[0], fs[0]
	base := program(p, f)
	for _, s := range []string{base + fmt.Sprintf("M104 S%g\n", f.NozzleTemperature), base + fmt.Sprintf("M140 S%g\n", f.BedTemperature), strings.Replace(base, "G1 X10", "M84\nG1 X10", 1), strings.Replace(base, "G1 X20 E0.4", "G1 X20 A90 E0.4", 1), strings.Replace(base, "M84", "M84 S30", 1)} {
		if _, e := Validate(context.Background(), strings.NewReader(s), p, f); e == nil {
			t.Fatal("invalid final shutdown or extra axis accepted")
		}
	}
}
