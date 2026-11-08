// Package gcode validates the supported linear FDM dialect and builds a batched
// preview from exactly the file offered for export. It never rewrites toolpaths.
package gcode

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/forge-slicer/forge/internal/config"
)

const MaxSegments = 2000000

type Preview struct {
	Segments          []float64 `json:"segments"`
	Layers            []float64 `json:"layers"`
	EstimatedSeconds  float64   `json:"estimatedSeconds"`
	FilamentMM        float64   `json:"filamentMm"`
	MassG             float64   `json:"massG"`
	PathMM            float64   `json:"pathMm"`
	SupportFilamentMM float64   `json:"supportFilamentMm"`
	Warnings          []string  `json:"warnings"`
}
type token struct {
	key     byte
	value   float64
	present bool
}

var commandParameters = map[string]string{"G0": "XYZEF", "G1": "XYZEF", "G21": "", "G20": "", "G90": "", "G91": "", "M82": "", "M83": "", "G28": "XYZ", "G92": "E", "M104": "ST", "M109": "SRT", "M140": "S", "M190": "SR", "M106": "S", "M107": "", "M204": "PST", "G4": "PS", "M84": "", "M18": "", "M400": "", "M73": "PRQS", "T0": ""}

func parse(line string) (string, map[byte]token, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil, nil
	}
	if len(line) > 65536 {
		return "", nil, errors.New("line exceeds 64 KiB")
	}
	// Parenthesized comments, checksums, and line numbers are intentionally outside
	// the generated dialect; accepting them would require separate state semantics.
	if strings.ContainsAny(line, "()*") {
		return "", nil, errors.New("unsupported comment/checksum syntax")
	}
	i := 0
	read := func() (token, error) {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i >= len(line) {
			return token{}, io.EOF
		}
		key := line[i]
		if key < 'A' || key > 'Z' {
			return token{}, errors.New("expected uppercase command/parameter")
		}
		i++
		start := i
		for i < len(line) && ((line[i] >= '0' && line[i] <= '9') || line[i] == '.' || line[i] == '-' || line[i] == '+') {
			i++
		}
		t := token{key: key, present: i > start}
		if t.present {
			v, e := strconv.ParseFloat(line[start:i], 64)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return t, errors.New("invalid numeric parameter")
			}
			t.value = v
		}
		return t, nil
	}
	first, e := read()
	if e != nil {
		return "", nil, e
	}
	if !first.present || first.value != math.Trunc(first.value) || (first.key != 'G' && first.key != 'M' && first.key != 'T') {
		return "", nil, errors.New("invalid G/M/T command")
	}
	cmd := fmt.Sprintf("%c%d", first.key, int(first.value))
	p := map[byte]token{}
	for {
		t, e := read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", nil, e
		}
		if _, ok := p[t.key]; ok {
			return "", nil, errors.New("duplicate parameter")
		}
		p[t.key] = t
	}
	return cmd, p, nil
}

var timeRE = regexp.MustCompile(`(\d+)\s*([dhms])`)

func ParseTime(s string) float64 {
	total := 0.
	for _, m := range timeRE.FindAllStringSubmatch(s, -1) {
		n, _ := strconv.ParseFloat(m[1], 64)
		switch m[2] {
		case "d":
			n *= 86400
		case "h":
			n *= 3600
		case "m":
			n *= 60
		}
		total += n
	}
	return total
}
func intersects(x1, y1, x2, y2, x, y, w, h float64) bool {
	lo, hi := 0., 1.
	dx, dy := x2-x1, y2-y1
	for _, v := range [][2]float64{{-dx, x1 - x}, {dx, x + w - x1}, {-dy, y1 - y}, {dy, y + h - y1}} {
		p, q := v[0], v[1]
		if math.Abs(p) < 1e-12 {
			if q < 0 {
				return false
			}
			continue
		}
		r := q / p
		if p < 0 {
			lo = math.Max(lo, r)
		} else {
			hi = math.Min(hi, r)
		}
		if lo > hi {
			return false
		}
	}
	return true
}
func Validate(ctx context.Context, r io.Reader, p config.Printer, f config.Filament) (*Preview, error) {
	out := &Preview{Segments: []float64{}, Layers: []float64{}, Warnings: []string{"Static validation cannot establish physical machine safety. Inspect the first layer and verify the profile against your machine.", "Print time is a native-engine or acceleration-limited estimate; heating and actual firmware behavior can differ."}}
	sc := bufio.NewScanner(io.LimitReader(r, 128<<20+1))
	sc.Buffer(make([]byte, 4096), 65536)
	lineNo := 0
	bytesRead := 0
	kind := 4
	layer := -1
	var pos [3]float64
	known := [3]bool{}
	ePos, feed, retractDebt := 0., 0., 0.
	absoluteE := true
	units, xyzMode, eMode, homed := false, false, false, false
	readyNozzle, readyBed := false, !p.Thermal.HasHeatedBed || f.BedTemperature == 0
	nozzle, bed := 0., 0.
	shutdownNozzle, shutdownBed, motorsOff := false, !p.Thermal.HasHeatedBed, false
	actualExtrusion := false
	engineTime := 0.
	allow := map[string]bool{}
	for _, s := range p.Firmware.SupportedCommands {
		allow[strings.ToUpper(s)] = true
	}
	deny := map[string]bool{}
	for _, s := range p.Firmware.UnsupportedCommands {
		deny[strings.ToUpper(s)] = true
	}
	maxFlow := math.Min(p.Extruders[0].MaxVolumetricFlow, f.MaxVolumetricFlow)
	fa := math.Pi * math.Pow(p.Extruders[0].FilamentDiameter/2, 2)
	fail := func(s string) (*Preview, error) { return nil, fmt.Errorf("G-code line %d: %s", lineNo, s) }
	for sc.Scan() {
		lineNo++
		if lineNo%4096 == 0 {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
		}
		raw := sc.Text()
		bytesRead += len(raw) + 1
		if bytesRead > 128<<20 {
			return fail("file exceeds 128 MiB validation limit")
		}
		code, comment, _ := strings.Cut(raw, ";")
		comment = strings.TrimSpace(comment)
		if strings.HasPrefix(comment, "TYPE:") {
			t := strings.ToLower(strings.TrimPrefix(comment, "TYPE:"))
			switch {
			case strings.Contains(t, "support"):
				kind = 3
			case strings.Contains(t, "perimeter"):
				kind = 1
			case strings.Contains(t, "infill"), strings.Contains(t, "solid"):
				kind = 2
			default:
				kind = 4
			}
		}
		if strings.HasPrefix(comment, "LAYER_CHANGE") {
			layer++
			out.Layers = append(out.Layers, 0)
		}
		if strings.HasPrefix(comment, "Z:") && layer >= 0 {
			z, e := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(comment, "Z:")), 64)
			if e == nil {
				out.Layers[layer] = z
			}
		}
		if strings.Contains(comment, "estimated printing time (normal mode)") {
			if _, v, ok := strings.Cut(comment, "="); ok {
				engineTime = ParseTime(v)
			}
		}
		cmd, a, err := parse(strings.TrimSpace(code))
		if err != nil {
			return fail(err.Error())
		}
		if cmd == "" {
			continue
		}
		if deny[cmd] {
			return fail("explicitly unsupported command " + cmd)
		}
		if len(allow) > 0 && !allow[cmd] {
			return fail("command not declared supported by printer: " + cmd)
		}
		value := func(k byte, def float64) float64 {
			if t, ok := a[k]; ok && t.present {
				return t.value
			}
			return def
		}
		if permitted, ok := commandParameters[cmd]; ok {
			for k := range a {
				if !strings.ContainsRune(permitted, rune(k)) {
					return fail(fmt.Sprintf("parameter %c has no validated semantics for %s", k, cmd))
				}
			}
		}
		for k, t := range a {
			if !t.present && cmd != "G28" {
				return fail(fmt.Sprintf("parameter %c lacks a number", k))
			}
		}
		switch cmd {
		case "G21":
			units = true
		case "G20":
			return fail("inch units are unsupported")
		case "G90":
			xyzMode = true
			eMode = false // Firmware varies; require M82/M83 after G90.
		case "G91":
			return fail("relative XYZ movement is outside the supported export dialect")
		case "M82":
			absoluteE = true
			eMode = true
		case "M83":
			absoluteE = false
			eMode = true
		case "G28":
			if motorsOff {
				return fail("homing after motor release is unsupported")
			}
			if actualExtrusion {
				return fail("homing after deposition is unsupported")
			}
			homed = true
			any := false
			for i, k := range []byte{'X', 'Y', 'Z'} {
				if _, ok := a[k]; ok {
					known[i] = true
					pos[i] = 0
					any = true
				}
			}
			if !any {
				known = [3]bool{true, true, true}
				pos = [3]float64{}
			}
		case "G92":
			for k := range a {
				if k != 'E' {
					return fail("G92 XYZ/work offsets are unsupported")
				}
			}
			ePos = value('E', ePos)
		case "G0", "G1":
			if motorsOff {
				return fail("movement after motor release is unsupported")
			}
			if !units || !xyzMode || !eMode || !homed {
				return fail("movement requires G21, G90, explicit M82/M83 and homing")
			}
			next := pos
			moving := false
			for i, k := range []byte{'X', 'Y', 'Z'} {
				if t, ok := a[k]; ok {
					next[i] = t.value
					known[i] = true
					moving = true
				}
			}
			if !known[0] || !known[1] || !known[2] {
				return fail("position is not established")
			}
			for i, limit := range []float64{p.BuildVolume.Width, p.BuildVolume.Depth, p.BuildVolume.Height} {
				if next[i] < -.02 || next[i] > limit+.02 {
					return fail(fmt.Sprintf("axis %d coordinate %.3f exceeds [0, %.3f] mm", i, next[i], limit))
				}
			}
			if moving {
				for _, r := range p.BuildVolume.ExcludedRegions {
					if intersects(pos[0], pos[1], next[0], next[1], r.X, r.Y, r.Width, r.Depth) {
						return fail("movement crosses an excluded region")
					}
				}
			}
			feed = value('F', feed)
			if feed <= 0 || feed > math.Max(p.Motion.MaxTravelSpeed, p.Motion.MaxSpeed)*60*1.001 {
				return fail("missing or excessive feed rate")
			}
			de := 0.
			if t, ok := a['E']; ok {
				if absoluteE {
					de = t.value - ePos
					ePos = t.value
				} else {
					de = t.value
					ePos += de
				}
			}
			if de < 0 {
				if -de > p.Extruders[0].RetractionLength+.1 {
					return fail("retraction exceeds configured length")
				}
				retractDebt -= de
			}
			deposited := math.Max(0, de-retractDebt)
			if de > 0 {
				retractDebt = math.Max(0, retractDebt-de)
			}
			dx, dy, dz := next[0]-pos[0], next[1]-pos[1], next[2]-pos[2]
			dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
			speed := feed / 60
			if deposited > 1e-5 {
				if !readyNozzle || !readyBed || nozzle < p.Thermal.MinNozzle {
					return fail("extrusion before required temperature waits")
				}
				if dist < 1e-6 {
					return fail("unbounded stationary extrusion is unsupported")
				}
				if speed > p.Motion.MaxSpeed*1.001 {
					return fail("extrusion movement exceeds print speed")
				}
				flow := deposited * fa * speed / dist
				if flow > maxFlow*1.025 {
					return fail(fmt.Sprintf("volumetric flow %.3f exceeds %.3f mm³/s", flow, maxFlow))
				}
				actualExtrusion = true
				out.FilamentMM += deposited
				if kind == 3 {
					out.SupportFilamentMM += deposited
				}
			}
			if dist > 1e-7 {
				if len(out.Segments)/9 >= MaxSegments {
					return fail("preview exceeds 2,000,000 segment limit")
				}
				k := 0
				if deposited > 1e-5 {
					k = kind
				}
				if layer < 0 && deposited > 0 {
					layer = 0
					out.Layers = append(out.Layers, next[2])
				}
				li := layer
				if li < 0 {
					li = 0
				}
				out.Segments = append(out.Segments, pos[0], pos[1], pos[2], next[0], next[1], next[2], float64(k), float64(li), feed)
				out.PathMM += dist
				accel := p.Motion.MaxAcceleration
				if dist >= speed*speed/accel {
					out.EstimatedSeconds += dist/speed + speed/accel
				} else {
					out.EstimatedSeconds += 2 * math.Sqrt(dist/accel)
				}
			} else if math.Abs(de) > 0 {
				if speed > p.Extruders[0].RetractionSpeed*1.001 {
					return fail("extruder-only speed exceeds retraction capability")
				}
				out.EstimatedSeconds += math.Abs(de) / speed
			}
			pos = next
		case "M104", "M109":
			temp := value('S', value('R', -1))
			if temp < 0 || temp > p.Thermal.MaxNozzle || (temp > 0 && temp < p.Thermal.MinNozzle) {
				return fail("invalid nozzle temperature")
			}
			if value('T', 0) != 0 {
				return fail("only extruder T0 is supported")
			}
			nozzle = temp
			readyNozzle = cmd == "M109" && temp >= f.NozzleTemperature-1
			if actualExtrusion && temp == 0 {
				shutdownNozzle = true
			}
		case "M140", "M190":
			temp := value('S', value('R', -1))
			if temp < 0 || temp > p.Thermal.MaxBed || (temp > 0 && temp < p.Thermal.MinBed) || (!p.Thermal.HasHeatedBed && temp > 0) {
				return fail("invalid bed temperature")
			}
			bed = temp
			_ = bed
			readyBed = temp == 0 && f.BedTemperature == 0 || cmd == "M190" && temp >= f.BedTemperature-1
			if actualExtrusion && temp == 0 {
				shutdownBed = true
			}
		case "M106":
			if !p.Thermal.PartCooling {
				return fail("fan command without part cooling capability")
			}
			if value('S', 255) < 0 || value('S', 255) > 255 {
				return fail("fan PWM outside 0..255")
			}
		case "M107":
		case "M204":
			for _, k := range []byte{'P', 'T', 'S'} {
				if t, ok := a[k]; ok && (t.value < 0 || t.value > p.Motion.MaxAcceleration*1.001) {
					return fail("acceleration exceeds machine limit")
				}
			}
		case "G4":
			if value('P', 0) < 0 || value('S', 0) < 0 {
				return fail("negative dwell")
			}
			out.EstimatedSeconds += value('S', 0) + value('P', 0)/1000
		case "M84", "M18":
			motorsOff = true
		case "M400", "M73", "M117": // M117 text is rejected by the numeric tokenizer.
		case "T0":
		default:
			return fail("command lacks validated semantics: " + cmd)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if !actualExtrusion || out.FilamentMM <= 0 {
		return nil, errors.New("G-code has no genuine deposition paths")
	}
	if !shutdownNozzle || !shutdownBed || !motorsOff || nozzle != 0 || bed != 0 {
		return nil, errors.New("G-code lacks heater-off and motor-release shutdown")
	}
	if engineTime > 0 {
		out.EstimatedSeconds = engineTime
	}
	out.MassG = out.FilamentMM * fa * f.Density / 1000
	return out, nil
}
