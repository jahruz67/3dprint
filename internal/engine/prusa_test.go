package engine

import (
	"strings"
	"testing"

	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/optimizer"
)

func TestAdapterExplicitLimitsAndNoScripts(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	s := optimizer.Settings{LayerHeight: .2, LineWidth: .45, Walls: 2, Infill: .15, InfillPattern: "gyroid", Speed: 1000, TravelSpeed: 1000, Acceleration: 500, Supports: true, SupportStyle: "snug"}
	ini, e := Translate(ps[0], fs[0], s)
	if e != nil {
		t.Fatal(e)
	}
	for _, expected := range []string{"post_process = \n", "binary_gcode = 0", "skirts = 0", "support_material = 1", "start_gcode = G21\\nG90\\nM82", "M104 S0", "M84", "machine_limits_usage = time_estimate_only"} {
		if !strings.Contains(ini, expected) {
			t.Fatalf("adapter missing %q", expected)
		}
	}
	if strings.Contains(ini, "perimeter_speed = 1000") {
		t.Fatal("speed not clamped to physical flow")
	}
}
func TestRejectShellAsEngine(t *testing.T) {
	if _, e := New("C:\\Windows\\System32\\cmd.exe"); e == nil {
		t.Fatal("shell accepted as slicing engine")
	}
}
func TestSupportRecipesTranslateToNative(t *testing.T) {
	ps, fs := config.ReferenceFixture()
	s := optimizer.SupportPolicy(optimizer.Settings{LayerHeight: .2, LineWidth: .45, Speed: 30, Supports: true}, config.Preferences{SupportMode: "easy_remove", SupportPlacement: "buildplate", SupportAngle: 50}, .4)
	ini, e := Translate(ps[0], fs[0], s)
	if e != nil {
		t.Fatal(e)
	}
	for _, want := range []string{"support_material_style = organic", "support_material_buildplate_only = 1", "support_material_contact_distance = 0.18000", "support_material_xy_spacing = 0.40000", "support_material_threshold = 50.00000", "dont_support_bridges = 1", "support_material_with_sheath = 0"} {
		if !strings.Contains(ini, want) {
			t.Fatalf("missing native setting %s", want)
		}
	}
}
