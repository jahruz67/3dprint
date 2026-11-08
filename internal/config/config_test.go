package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultsRequirePrinterImportAndOfferMaterialChoices(t *testing.T) {
	printers, filaments := Defaults()
	if printers == nil || len(printers) != 0 {
		t.Fatal("fresh application must contain no configured or simulated printer")
	}
	if len(filaments) != 7 {
		t.Fatalf("expected six generic starters plus the custom PLA+ entry, got %d", len(filaments))
	}
	want := map[string]bool{"PLA": true, "PLA+": true, "PETG": true, "ABS": true, "ASA": true, "TPU": true, "PA": true}
	for _, f := range filaments {
		if err := ValidateFilament(f); err != nil {
			t.Fatalf("starter %s is invalid: %v", f.ID, err)
		}
		if f.ID != "user-pla-plus-115" && (!strings.HasPrefix(f.ID, "generic-") || !strings.Contains(f.Name, "review spool settings")) {
			t.Fatalf("starter must disclose generic scope: %+v", f)
		}
		delete(want, f.Material)
	}
	if len(want) != 0 {
		t.Fatalf("missing material families: %v", want)
	}
	filaments[0].NozzleTemperature = 999
	_, again := Defaults()
	if again[0].NozzleTemperature == 999 {
		t.Fatal("defaults share mutable profile state")
	}
}

func TestCustomPLAPlusKeepsRequestedNozzleWithoutWeakeningMachine(t *testing.T) {
	_, materials := Defaults()
	var custom Filament
	for _, f := range materials {
		if f.ID == "user-pla-plus-115" {
			custom = f
		}
	}
	if custom.Material != "PLA+" || custom.NozzleTemperature != 115 || len(custom.Notes) < 3 {
		t.Fatal("custom PLA+ request or warning notes missing")
	}
	printers, _ := ReferenceFixture()
	before := printers[0].Thermal.MinNozzle
	if err := Compose(printers[0], custom, Preferences{SchemaVersion: "1.0", Mode: "balanced", Budget: 64, MinWalls: 2, Supports: true}); err == nil {
		t.Fatal("custom low temperature overrode verified printer minimum")
	}
	if printers[0].Thermal.MinNozzle != before {
		t.Fatal("composition changed printer thermal limits")
	}
}

func TestOptionalSupportPreferencesAreStrictAndBackwardCompatible(t *testing.T) {
	base := `{"schemaVersion":"1.0","mode":"balanced","budget":64,"minWalls":2,"minInfill":0.1,"supports":true`
	for _, mode := range []string{"auto", "optimized", "easy_remove", "stable", "off"} {
		raw := base + `,"supportMode":"` + mode + `","supportStyle":"organic","supportPlacement":"buildplate","supportAngle":60}`
		if _, err := ParsePreferences([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, tail := range []string{`,"supportMode":"magic"}`, `,"supportStyle":"tree-shaped"}`, `,"supportPlacement":"nowhere"}`, `,"supportAngle":0}`, `,"supportAngle":81}`} {
		if _, err := ParsePreferences([]byte(base + tail)); err == nil {
			t.Fatalf("invalid support preference accepted: %s", tail)
		}
	}
	legacy, err := ParsePreferences([]byte(base + `}`))
	if err != nil || legacy.SupportMode != "" || legacy.SupportAngle != 0 {
		t.Fatal("legacy supports profile no longer parses with optional zero values")
	}
}

func TestPrinterTemplateKeepsUnknownHardwareUnresolved(t *testing.T) {
	draft := PrinterTemplate()
	data, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParsePrinter(data); err == nil {
		t.Fatal("blank template must not pass as a verified printer")
	}
	bed := draft["buildVolume"].(map[string]any)
	for _, field := range []string{"width", "depth", "height", "shape", "origin"} {
		if bed[field] != nil {
			t.Fatalf("unknown build field %s must be null", field)
		}
	}
	extruder := draft["extruders"].([]any)[0].(map[string]any)
	if extruder["maxVolumetricFlow"] != 4.0 {
		t.Fatal("draft must provide the default operating flow cap")
	}
	if _, present := extruder["flowVerified"]; present {
		t.Fatal("draft must not ask users to certify flow")
	}
	bed["width"] = 123
	if PrinterTemplate()["buildVolume"].(map[string]any)["width"] != nil {
		t.Fatal("template drafts share mutable hardware values")
	}
}

func TestProfilesRoundTripAndComposition(t *testing.T) {
	p, f := ReferenceFixture()
	b, e := json.Marshal(p[0])
	if e != nil {
		t.Fatal(e)
	}
	round, e := ParsePrinter(b)
	if e != nil {
		t.Fatal(e)
	}
	if round.Extruders[0].NozzleDiameter != .4 {
		t.Fatal("round trip changed nozzle")
	}
	if e := Compose(round, f[0], Preferences{SchemaVersion: "1.0", Mode: "balanced", Budget: 128, MinWalls: 2, MinInfill: .12, Supports: true}); e != nil {
		t.Fatal(e)
	}
}

func TestRejectUnsafeAndAmbiguousProfiles(t *testing.T) {
	printers, _ := ReferenceFixture()
	original, _ := json.Marshal(printers[0])
	cases := map[string]func(*Printer){
		"firmware mismatch":           func(p *Printer) { p.Firmware.Family = "klipper" },
		"unknown provenance":          func(p *Printer) { p.Provenance["motion"] = "unknown" },
		"excluded region outside bed": func(p *Printer) { p.BuildVolume.ExcludedRegions = []Region{{X: 210, Y: 0, Width: 20, Depth: 20}} },
		"unknown version":             func(p *Printer) { p.SchemaVersion = "99" },
		"script macro":                func(p *Printer) { p.Firmware.StartGCode = []string{"RUN_SHELL_COMMAND CMD=example"} },
		"templating":                  func(p *Printer) { p.Firmware.StartGCode = []string{"M104 S{temperature}"} },
		"contradictory commands":      func(p *Printer) { p.Firmware.UnsupportedCommands = append(p.Firmware.UnsupportedCommands, "G1") },
		"multiple extruders":          func(p *Printer) { p.Extruders = append(p.Extruders, p.Extruders[0]) },
		"unsafe temperature interval": func(p *Printer) { p.Thermal.MinNozzle = 280 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			var p Printer
			if e := json.Unmarshal(original, &p); e != nil {
				t.Fatal(e)
			}
			change(&p)
			if ValidatePrinter(p) == nil {
				t.Fatal("unsafe profile accepted")
			}
		})
	}
	for _, raw := range []string{strings.Replace(string(original), `"schemaVersion":"1.0"`, `"schemaVersion":"1.0","schemaVersion":"1.0"`, 1), strings.TrimSuffix(string(original), "}") + `,"unrecognized":true}`, string(original) + " {}", strings.Replace(string(original), `"maxSpeed":120`, `"maxSpeed":1e9999999`, 1)} {
		if _, e := ParsePrinter([]byte(raw)); e == nil {
			t.Fatalf("ambiguous JSON accepted: %.80s", raw)
		}
	}
}

func TestCompositionCannotOverrideHardware(t *testing.T) {
	p, f := ReferenceFixture()
	pref := Preferences{SchemaVersion: "1.0", Mode: "material", Budget: 64, MinWalls: 2, MinInfill: .1, Supports: true}
	f[0].NozzleTemperature = 300
	if Compose(p[0], f[0], pref) == nil {
		t.Fatal("material exceeded thermal limit")
	}
	f[0].NozzleTemperature = 205
	p[0].Thermal.PartCooling = false
	if Compose(p[0], f[0], pref) == nil {
		t.Fatal("material requested unavailable fan")
	}
}

func TestPreferencesSchema(t *testing.T) {
	for _, mode := range []string{"fast", "balanced", "quality", "strength", "material"} {
		raw := []byte(`{"schemaVersion":"1.0","mode":"` + mode + `","budget":64,"minWalls":2,"minInfill":0.1,"supports":true}`)
		if _, e := ParsePreferences(raw); e != nil {
			t.Fatal(e)
		}
	}
	for _, raw := range []string{`{"schemaVersion":"1.0","mode":"magic","budget":64,"minWalls":2,"minInfill":0.1,"supports":true}`, `{"schemaVersion":"1.0","mode":"fast","budget":99999,"minWalls":2,"minInfill":0.1,"supports":true}`} {
		if _, e := ParsePreferences([]byte(raw)); e == nil {
			t.Fatal("invalid preference accepted")
		}
	}
}

func TestMigrationDoesNotInventValues(t *testing.T) {
	if _, e := Migrate("printer", []byte(`{"schemaVersion":"0.9"}`)); e == nil {
		t.Fatal("unknown old version silently migrated")
	}
}
