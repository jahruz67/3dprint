package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"sync"

	"github.com/forge-slicer/forge/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const SchemaVersion = "1.0"

var once sync.Once
var compiled map[string]*jsonschema.Schema
var schemaError error

func loadSchemas() {
	compiled = make(map[string]*jsonschema.Schema)
	for _, kind := range []string{"printer", "filament", "optimization"} {
		raw, err := schemas.Files.ReadFile(kind + ".schema.json")
		if err != nil {
			schemaError = err
			return
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			schemaError = err
			return
		}
		c := jsonschema.NewCompiler()
		loc := "https://forge-slicer.local/schemas/" + kind + ".schema.json"
		if err = c.AddResource(loc, doc); err != nil {
			schemaError = err
			return
		}
		compiled[kind], err = c.Compile(loc)
		if err != nil {
			schemaError = err
			return
		}
	}
}

// ValidateJSON validates against the bundled Draft 2020-12 schema. It never fetches external resources.
func ValidateJSON(kind string, data []byte) error {
	if len(data) > 1024*1024 {
		return errors.New("profile exceeds 1 MiB limit")
	}
	once.Do(loadSchemas)
	if schemaError != nil {
		return fmt.Errorf("bundled schema: %w", schemaError)
	}
	sch, ok := compiled[kind]
	if !ok {
		return errors.New("unknown profile kind")
	}
	// Strict numeric decoding rejects infinities, oversized exponents and trailing documents before schema evaluation.
	var value any
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&value); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return errors.New("profile must contain exactly one JSON document")
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	if err := sch.Validate(value); err != nil {
		return fmt.Errorf("%s profile schema: %w", kind, err)
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	var parse func() error
	parse = func() error {
		t, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				key := k.(string)
				if seen[key] {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = true
				if e = parse(); e != nil {
					return e
				}
			}
			_, e = d.Token()
			return e
		case '[':
			for d.More() {
				if e = parse(); e != nil {
					return e
				}
			}
			_, e = d.Token()
			return e
		}
		return nil
	}
	return parse()
}

func ParsePrinter(data []byte) (Printer, error) {
	var p Printer
	if err := ValidateJSON("printer", data); err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, validatePrinterSemantic(p)
}
func ParseFilament(data []byte) (Filament, error) {
	var p Filament
	if err := ValidateJSON("filament", data); err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, validateFilamentSemantic(p)
}
func ParsePreferences(data []byte) (Preferences, error) {
	var p Preferences
	if err := ValidateJSON("optimization", data); err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, nil
}
func ValidatePrinter(p Printer) error {
	b, e := json.Marshal(p)
	if e != nil {
		return e
	}
	if e = ValidateJSON("printer", b); e != nil {
		return e
	}
	return validatePrinterSemantic(p)
}
func ValidateFilament(p Filament) error {
	b, e := json.Marshal(p)
	if e != nil {
		return e
	}
	if e = ValidateJSON("filament", b); e != nil {
		return e
	}
	return validateFilamentSemantic(p)
}
func ValidatePreferences(p Preferences) error {
	b, e := json.Marshal(p)
	if e != nil {
		return e
	}
	return ValidateJSON("optimization", b)
}

var commandRE = regexp.MustCompile(`^(?:G|M|T)[0-9]+$`)

func validatePrinterSemantic(p Printer) error {
	var errs []string
	if len(p.Extruders) != 1 {
		errs = append(errs, "initial release supports exactly one extruder")
	}
	if p.BuildVolume.Shape != "rectangular" || p.BuildVolume.Origin != "front-left" {
		errs = append(errs, "initial release requires rectangular bed with front-left origin")
	}
	if p.Thermal.MinNozzle >= p.Thermal.MaxNozzle || p.Thermal.MinBed > p.Thermal.MaxBed {
		errs = append(errs, "thermal minimum must be below maximum")
	}
	if !p.Thermal.HasHeatedBed && (p.Thermal.MinBed != 0 || p.Thermal.MaxBed != 0) {
		errs = append(errs, "unheated bed must have zero minimum and maximum temperatures")
	}
	for _, e := range p.Extruders {
		if e.NozzleDiameter >= e.FilamentDiameter {
			errs = append(errs, "nozzleDiameter must be smaller than filamentDiameter")
		}
		if e.RetractionSpeed > p.Motion.MaxSpeed {
			errs = append(errs, "retractionSpeed exceeds configured maxSpeed")
		}
	}
	for _, r := range p.BuildVolume.ExcludedRegions {
		if r.X < 0 || r.Y < 0 || r.X+r.Width > p.BuildVolume.Width || r.Y+r.Depth > p.BuildVolume.Depth {
			errs = append(errs, "excluded region must lie inside build volume")
		}
	}
	flavors := map[string]string{"marlin": "marlin2", "klipper": "klipper", "reprap": "reprapfirmware"}
	if flavors[p.Firmware.Family] != p.Firmware.GCodeFlavor {
		errs = append(errs, "firmware family and gcodeFlavor disagree")
	}
	supported := map[string]bool{}
	unsupported := map[string]bool{}
	for _, c := range p.Firmware.SupportedCommands {
		if !commandRE.MatchString(c) {
			errs = append(errs, "invalid supported command "+c)
		}
		supported[c] = true
	}
	for _, c := range p.Firmware.UnsupportedCommands {
		if !commandRE.MatchString(c) {
			errs = append(errs, "invalid unsupported command "+c)
		}
		unsupported[c] = true
		if supported[c] {
			errs = append(errs, "command appears as both supported and unsupported: "+c)
		}
	}
	for _, c := range []string{"G0", "G1", "G21", "G28", "G90", "G92", "M84", "M104", "M109", "M107", "M204", "M400"} {
		if !supported[c] || unsupported[c] {
			errs = append(errs, "required command unavailable: "+c)
		}
	}
	extrusionCommand := "M82"
	if p.Firmware.ExtrusionMode == "relative" {
		extrusionCommand = "M83"
	}
	if !supported[extrusionCommand] {
		errs = append(errs, "extrusion mode requires "+extrusionCommand)
	}
	if p.Thermal.HasHeatedBed && (!supported["M140"] || !supported["M190"]) {
		errs = append(errs, "heated bed requires M140 and M190")
	}
	if p.Thermal.PartCooling && !supported["M106"] {
		errs = append(errs, "part cooling requires M106")
	}
	if p.Capabilities.FirmwareRetraction {
		errs = append(errs, "firmware retraction is not enabled by this adapter; set capability false and use explicit retraction")
	}
	// Imported snippets are data only. Unsupported macros are rejected instead of executed or silently trusted.
	for _, lines := range [][]string{p.Firmware.StartGCode, p.Firmware.EndGCode} {
		for _, line := range lines {
			if strings.ContainsAny(line, "\r\n{}[]`$\\") {
				errs = append(errs, "G-code snippets must be single literal lines without templates, escapes or macros")
				continue
			}
			s := strings.TrimSpace(strings.SplitN(line, ";", 2)[0])
			if s == "" {
				continue
			}
			cmd := strings.ToUpper(strings.Fields(s)[0])
			if !commandRE.MatchString(cmd) || !supported[cmd] || unsupported[cmd] {
				errs = append(errs, "unreviewable or unsupported snippet command: "+cmd)
			}
		}
	}
	for _, key := range []string{"buildVolume", "motion", "thermal"} {
		v := strings.TrimSpace(p.Provenance[key])
		if v == "" || strings.EqualFold(v, "unknown") {
			errs = append(errs, "provenance."+key+": supply documentation source or calibration record")
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
func validateFilamentSemantic(f Filament) error {
	if f.MinLayerHeight > f.MaxLayerHeight {
		return errors.New("minLayerHeight exceeds maxLayerHeight")
	}
	return nil
}

// Compose checks the material against non-overridable machine limits. Material never increases machine capabilities.
func Compose(p Printer, f Filament, pref Preferences) error {
	if e := ValidatePrinter(p); e != nil {
		return e
	}
	if e := ValidateFilament(f); e != nil {
		return e
	}
	if e := ValidatePreferences(pref); e != nil {
		return e
	}
	if f.NozzleTemperature < p.Thermal.MinNozzle || f.NozzleTemperature > p.Thermal.MaxNozzle {
		return errors.New("filament nozzle temperature is outside machine limits")
	}
	if f.BedTemperature < p.Thermal.MinBed || f.BedTemperature > p.Thermal.MaxBed {
		return errors.New("filament bed temperature is outside machine limits")
	}
	if !p.Thermal.PartCooling && f.FanPercent > 0 {
		return errors.New("filament requests cooling but machine has no verified part cooling")
	}
	e := p.Extruders[0]
	if math.Max(f.MinLayerHeight, e.NozzleDiameter*.2) > math.Min(f.MaxLayerHeight, e.NozzleDiameter*.8) {
		return errors.New("material and nozzle have no feasible layer-height interval")
	}
	return nil
}

// Migrate explicitly rejects unknown versions; 1.0 currently has no predecessor.
func Migrate(kind string, data []byte) ([]byte, error) {
	if err := ValidateJSON(kind, data); err != nil {
		return nil, fmt.Errorf("migration unavailable: only schemaVersion 1.0 is implemented: %w", err)
	}
	return append([]byte(nil), data...), nil
}
