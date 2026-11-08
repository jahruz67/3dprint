// Package engine isolates the replaceable native slicing backend. No imported
// project settings or shell commands are executed by this adapter.
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/forge-slicer/forge/internal/config"
	"github.com/forge-slicer/forge/internal/geometry"
	"github.com/forge-slicer/forge/internal/optimizer"
)

type Status struct {
	Available bool   `json:"available"`
	Path      string `json:"path"`
	Version   string `json:"version"`
}
type Request struct {
	Mesh      *geometry.Mesh
	Printer   config.Printer
	Filament  config.Filament
	Settings  optimizer.Settings
	Directory string
}
type Result struct {
	Path    string
	Log     string
	Elapsed time.Duration
}
type Backend interface {
	Status() Status
	Slice(context.Context, Request, func(float64, string)) (Result, error)
}
type Prusa struct{ path string }

func New(path string) (*Prusa, error) {
	if path == "" {
		return nil, errors.New("PrusaSlicer executable path is required")
	}
	abs, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	st, e := os.Stat(abs)
	if e != nil || st.IsDir() {
		return nil, errors.New("engine executable does not exist")
	}
	name := strings.ToLower(filepath.Base(abs))
	if name != "prusa-slicer-console.exe" && name != "prusa-slicer.exe" && name != "prusa-slicer" && name != "prusaslicer" {
		return nil, errors.New("choose a PrusaSlicer executable, not a shell or script")
	}
	return &Prusa{abs}, nil
}
func Discover() *Prusa {
	if p, e := New(os.Getenv("FORGE_ENGINE")); e == nil {
		return p
	}
	for _, n := range []string{"prusa-slicer-console.exe", "prusa-slicer.exe", "prusa-slicer", "PrusaSlicer"} {
		if s, e := exec.LookPath(n); e == nil {
			if p, e := New(s); e == nil {
				return p
			}
		}
	}
	exe, _ := os.Executable()
	roots := []string{".", filepath.Dir(exe), filepath.Dir(filepath.Dir(exe))}
	for _, root := range roots {
		for _, pattern := range []string{"tools/prusa/**/prusa-slicer-console.exe", "tools/prusa/prusa-slicer-console.exe", "tools/prusa/**/prusa-slicer.exe", "engine/prusa-slicer-console.exe", "engine/prusa-slicer", "PrusaSlicer.app/Contents/MacOS/PrusaSlicer"} {
			matches, _ := filepath.Glob(filepath.Join(root, pattern))
			for _, path := range matches {
				if p, e := New(path); e == nil {
					return p
				}
			}
		}
	}
	return nil
}
func (p *Prusa) Status() Status {
	if p == nil {
		return Status{Version: "Not configured"}
	}
	return Status{Available: true, Path: p.path, Version: "PrusaSlicer CLI"}
}

type logWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
	cb  func(float64, string)
}

func (l *logWriter) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.buf.Len() < 256<<10 {
		n := len(b)
		if n > (256<<10)-l.buf.Len() {
			n = (256 << 10) - l.buf.Len()
		}
		l.buf.Write(b[:n])
	}
	for _, s := range strings.Split(string(b), "\n") {
		s = strings.TrimSpace(s)
		if len(s) > 0 && len(s) < 300 && l.cb != nil {
			l.cb(.45, s)
		}
	}
	return len(b), nil
}
func (p *Prusa) Slice(ctx context.Context, r Request, progress func(float64, string)) (Result, error) {
	if p == nil {
		return Result{}, errors.New("PrusaSlicer is not installed; configure its executable in Settings")
	}
	if e := config.ValidatePrinter(r.Printer); e != nil {
		return Result{}, e
	}
	if e := config.ValidateFilament(r.Filament); e != nil {
		return Result{}, e
	}
	if e := os.MkdirAll(r.Directory, 0700); e != nil {
		return Result{}, e
	}
	model := filepath.Join(r.Directory, "model.stl")
	f, e := os.Create(model)
	if e != nil {
		return Result{}, e
	}
	e = r.Mesh.WriteSTL(f)
	ce := f.Close()
	if e != nil {
		return Result{}, e
	}
	if ce != nil {
		return Result{}, ce
	}
	ini, e := Translate(r.Printer, r.Filament, r.Settings)
	if e != nil {
		return Result{}, e
	}
	settings := filepath.Join(r.Directory, "settings.ini")
	if e = os.WriteFile(settings, []byte(ini), 0600); e != nil {
		return Result{}, e
	}
	out := filepath.Join(r.Directory, "output.gcode")
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.path, "--export-gcode", "--config-compatibility", "disable", "--datadir", filepath.Join(r.Directory, "engine-data"), "--threads", strconv.Itoa(min(4, runtime.NumCPU())), "--load", settings, "--dont-arrange", "--no-ensure-on-bed", "--output", out, model)
	cmd.Dir = r.Directory
	hideWindow(cmd)
	log := &logWriter{cb: progress}
	cmd.Stdout = log
	cmd.Stderr = log
	start := time.Now()
	e = cmd.Run()
	elapsed := time.Since(start)
	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}
	if e != nil {
		return Result{}, fmt.Errorf("native slicer failed: %w\n%s", e, log.buf.String())
	}
	st, e := os.Stat(out)
	if e != nil || st.Size() < 100 {
		return Result{}, errors.New("native slicer produced no G-code")
	}
	if st.Size() > 128<<20 {
		return Result{}, errors.New("native output exceeds 128 MiB")
	}
	return Result{Path: out, Log: log.buf.String(), Elapsed: elapsed}, nil
}
func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\r", "")
	return strings.ReplaceAll(s, "\n", "\\n")
}
func Translate(p config.Printer, f config.Filament, s optimizer.Settings) (string, error) {
	if len(p.Extruders) != 1 {
		return "", errors.New("native adapter supports exactly one extruder")
	}
	ex := p.Extruders[0]
	var b strings.Builder
	set := func(k string, v any) { fmt.Fprintf(&b, "%s = %v\n", k, v) }
	num := func(v float64) string { return strconv.FormatFloat(v, 'f', 5, 64) }
	start := []string{"G21", "G90"}
	mode := "M82"
	if p.Firmware.ExtrusionMode == "relative" {
		mode = "M83"
	}
	start = append(start, mode)
	if p.Thermal.HasHeatedBed {
		start = append(start, fmt.Sprintf("M140 S%s", num(f.BedTemperature)))
	}
	start = append(start, fmt.Sprintf("M104 S%s", num(f.NozzleTemperature)), "G28")
	// Only validated numeric G/M-code is allowed in profile sequences. No engine
	// placeholders, expressions, filename substitutions, or postprocessing hooks.
	start = append(start, p.Firmware.StartGCode...)
	if p.Thermal.HasHeatedBed && f.BedTemperature > 0 {
		start = append(start, fmt.Sprintf("M190 S%s", num(f.BedTemperature)))
	}
	start = append(start, fmt.Sprintf("M109 S%s", num(f.NozzleTemperature)), "G21", "G90", mode, "G92 E0")
	end := append([]string{}, p.Firmware.EndGCode...)
	end = append(end, "M400", "M104 S0")
	if p.Thermal.HasHeatedBed {
		end = append(end, "M140 S0")
	}
	if p.Thermal.PartCooling {
		end = append(end, "M107")
	}
	end = append(end, "M84")
	set("start_gcode", escape(strings.Join(start, "\n")))
	set("end_gcode", escape(strings.Join(end, "\n")))
	set("before_layer_gcode", "")
	set("layer_gcode", "")
	set("toolchange_gcode", "")
	set("post_process", "")
	set("printer_technology", "FFF")
	set("gcode_flavor", p.Firmware.GCodeFlavor)
	set("bed_shape", fmt.Sprintf("0x0,%sx0,%sx%s,0x%s", num(p.BuildVolume.Width), num(p.BuildVolume.Width), num(p.BuildVolume.Depth), num(p.BuildVolume.Depth)))
	set("max_print_height", num(p.BuildVolume.Height))
	set("nozzle_diameter", num(ex.NozzleDiameter))
	set("filament_diameter", num(ex.FilamentDiameter))
	set("filament_density", num(f.Density))
	material := f.Material
	if material == "PLA+" {
		material = "PLA"
	}
	set("filament_type", material)
	set("temperature", num(f.NozzleTemperature))
	set("first_layer_temperature", num(f.NozzleTemperature))
	set("bed_temperature", num(f.BedTemperature))
	set("first_layer_bed_temperature", num(f.BedTemperature))
	set("use_relative_e_distances", boolInt(p.Firmware.ExtrusionMode == "relative"))
	set("use_firmware_retraction", 0)
	set("use_volumetric_e", 0)
	set("extrusion_multiplier", 1)
	set("layer_height", num(s.LayerHeight))
	set("first_layer_height", num(s.LayerHeight))
	set("extrusion_width", num(s.LineWidth))
	set("first_layer_extrusion_width", num(s.LineWidth))
	set("perimeter_extrusion_width", num(s.LineWidth))
	set("external_perimeter_extrusion_width", num(s.LineWidth))
	set("infill_extrusion_width", num(s.LineWidth))
	set("solid_infill_extrusion_width", num(s.LineWidth))
	set("top_infill_extrusion_width", num(s.LineWidth))
	set("support_material_extrusion_width", num(s.LineWidth))
	set("perimeters", s.Walls)
	set("top_solid_layers", int(math.Max(3, math.Ceil(.8/s.LayerHeight))))
	set("bottom_solid_layers", int(math.Max(3, math.Ceil(.8/s.LayerHeight))))
	set("fill_density", num(s.Infill*100)+"%")
	set("fill_pattern", s.InfillPattern)
	set("top_fill_pattern", "monotonic")
	set("bottom_fill_pattern", "monotonic")
	set("infill_every_layers", 1)
	set("spiral_vase", 0)
	speed := math.Min(s.Speed, math.Min(ex.MaxVolumetricFlow, f.MaxVolumetricFlow)/(s.LayerHeight*s.LineWidth))
	for _, k := range []string{"perimeter_speed", "infill_speed", "solid_infill_speed", "support_material_speed", "gap_fill_speed", "bridge_speed"} {
		set(k, num(speed))
	}
	set("external_perimeter_speed", num(speed*.75))
	set("small_perimeter_speed", num(speed*.65))
	set("top_solid_infill_speed", num(speed*.65))
	set("support_material_interface_speed", num(speed*.8))
	set("first_layer_speed", num(math.Min(20, speed)))
	set("travel_speed", num(math.Min(s.TravelSpeed, p.Motion.MaxTravelSpeed)))
	set("travel_speed_z", num(math.Min(10, p.Motion.MaxTravelSpeed)))
	set("max_print_speed", num(speed))
	set("max_volumetric_speed", num(math.Min(ex.MaxVolumetricFlow, f.MaxVolumetricFlow)))
	set("filament_max_volumetric_speed", num(math.Min(ex.MaxVolumetricFlow, f.MaxVolumetricFlow)))
	set("default_acceleration", num(s.Acceleration))
	set("travel_acceleration", num(s.Acceleration))
	for _, k := range []string{"perimeter_acceleration", "infill_acceleration", "bridge_acceleration", "first_layer_acceleration", "top_solid_infill_acceleration"} {
		set(k, num(s.Acceleration))
	}
	set("machine_limits_usage", "time_estimate_only")
	for _, k := range []string{"machine_max_acceleration_x", "machine_max_acceleration_y", "machine_max_acceleration_z", "machine_max_acceleration_extruding", "machine_max_acceleration_travel"} {
		set(k, num(p.Motion.MaxAcceleration))
	}
	for _, k := range []string{"machine_max_feedrate_x", "machine_max_feedrate_y", "machine_max_feedrate_z"} {
		set(k, num(p.Motion.MaxTravelSpeed))
	}
	set("retract_length", num(ex.RetractionLength))
	set("retract_speed", num(ex.RetractionSpeed))
	set("deretract_speed", num(ex.RetractionSpeed))
	set("retract_lift", 0)
	set("retract_before_travel", 2)
	set("retract_layer_change", 1)
	set("wipe", 0)
	set("avoid_crossing_perimeters", 1)
	set("seam_position", "nearest")
	set("skirts", 0)
	set("brim_width", 0)
	set("raft_layers", 0)
	set("support_material", boolInt(s.Supports))
	style := s.SupportStyle
	if style == "" {
		style = "snug"
	}
	set("support_material_style", style)
	set("support_material_auto", 1)
	set("support_material_buildplate_only", boolInt(s.SupportBuildplateOnly))
	threshold := s.SupportThreshold
	if threshold == 0 {
		threshold = 45
	}
	gap := s.SupportContactZ
	if gap == 0 {
		gap = math.Max(.15, s.LayerHeight)
	}
	spacing := s.SupportSpacing
	if spacing == 0 {
		spacing = 2.5
	}
	xy := s.SupportXYDistance
	if xy == 0 {
		xy = .5 * ex.NozzleDiameter
	}
	interfaces := s.SupportInterfaceLayers
	if interfaces == 0 {
		interfaces = 2
	}
	set("support_material_threshold", num(threshold))
	set("support_material_contact_distance", num(gap))
	set("support_material_bottom_contact_distance", num(gap))
	set("support_material_xy_spacing", num(xy))
	set("support_material_interface_layers", interfaces)
	set("support_material_interface_spacing", num(s.SupportInterfaceSpacing))
	set("support_material_with_sheath", boolInt(s.SupportSheath))
	set("dont_support_bridges", 1)
	set("support_material_pattern", "rectilinear")
	set("support_material_spacing", num(spacing))
	fan := f.FanPercent
	if !p.Thermal.PartCooling {
		fan = 0
	}
	set("cooling", boolInt(p.Thermal.PartCooling))
	set("fan_always_on", boolInt(fan > 0))
	set("min_fan_speed", fan)
	set("max_fan_speed", fan)
	set("bridge_fan_speed", fan)
	set("disable_fan_first_layers", 1)
	set("slowdown_below_layer_time", 5)
	set("min_print_speed", num(math.Min(10, speed)))
	set("binary_gcode", 0)
	set("gcode_comments", 0)
	set("remaining_times", 0)
	set("label_objects", 0)
	set("complete_objects", 0)
	set("thumbnails", "")
	set("silent_mode", 0)
	return b.String(), nil
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Keep io referenced here to make the output sink contract explicit.
var _ io.Writer = (*logWriter)(nil)
