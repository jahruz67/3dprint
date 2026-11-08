// Package geometry imports bounded, inert mesh data. Imported 3MF printer settings
// and scripts are deliberately never forwarded to the native slicing engine.
package geometry

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
)

const MaxBytes = 64 << 20
const MaxTriangles = 500000

type Vec [3]float64
type Triangle [3]Vec
type Bounds struct {
	Min Vec `json:"min"`
	Max Vec `json:"max"`
}
type Mesh struct {
	Triangles    []Triangle `json:"-"`
	Bounds       Bounds     `json:"bounds"`
	Volume       float64    `json:"volumeMm3"`
	Area         float64    `json:"areaMm2"`
	OverhangArea float64    `json:"overhangAreaMm2"`
	MinFeature   float64    `json:"minFeatureMm"`
	Warnings     []string   `json:"warnings"`
}
type Transform struct {
	Rotation Vec `json:"rotation"`
	Scale    Vec `json:"scale"`
	Position Vec `json:"position"`
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func sub(a, b Vec) Vec      { return Vec{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func cross(a, b Vec) Vec {
	return Vec{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func dot(a, b Vec) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func length(a Vec) float64 { return math.Sqrt(dot(a, a)) }

func Parse(name string, data []byte) (*Mesh, error) {
	if len(data) > MaxBytes {
		return nil, errors.New("model exceeds 64 MiB import limit")
	}
	var ts []Triangle
	var err error
	switch strings.ToLower(filepath.Ext(name)) {
	case ".stl":
		ts, err = parseSTL(data)
	case ".3mf":
		ts, err = parse3MF(data)
	default:
		return nil, errors.New("supported models are STL and mesh-only 3MF")
	}
	if err != nil {
		return nil, err
	}
	return Analyze(ts)
}
func parseSTL(data []byte) ([]Triangle, error) {
	if len(data) >= 84 {
		n := int(binary.LittleEndian.Uint32(data[80:84]))
		if n > 0 && n <= MaxTriangles && 84+50*n == len(data) {
			out := make([]Triangle, n)
			for i := range out {
				for j := 0; j < 9; j++ {
					o := 84 + i*50 + 12 + j*4
					out[i][j/3][j%3] = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[o : o+4])))
				}
			}
			return out, nil
		}
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 4096), 1024*1024)
	var out []Triangle
	var vertices []Vec
	solid := false
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 0 {
			continue
		}
		if f[0] == "solid" {
			solid = true
		}
		if f[0] != "vertex" {
			continue
		}
		if len(f) != 4 {
			return nil, errors.New("malformed STL vertex")
		}
		var v Vec
		for j := 0; j < 3; j++ {
			x, e := strconv.ParseFloat(f[j+1], 64)
			if e != nil {
				return nil, e
			}
			v[j] = x
		}
		vertices = append(vertices, v)
		if len(vertices) == 3 {
			out = append(out, Triangle{vertices[0], vertices[1], vertices[2]})
			vertices = vertices[:0]
			if len(out) > MaxTriangles {
				return nil, errors.New("model exceeds 500,000 triangle limit")
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if !solid || len(vertices) != 0 || len(out) == 0 {
		return nil, errors.New("invalid or empty STL")
	}
	return out, nil
}

type xmlModel struct {
	Unit      string `xml:"unit,attr"`
	Resources struct {
		Objects []xmlObject `xml:"object"`
	} `xml:"resources"`
	Build struct {
		Items []xmlItem `xml:"item"`
	} `xml:"build"`
}
type xmlObject struct {
	ID   int `xml:"id,attr"`
	Mesh struct {
		Vertices []struct {
			X float64 `xml:"x,attr"`
			Y float64 `xml:"y,attr"`
			Z float64 `xml:"z,attr"`
		} `xml:"vertices>vertex"`
		Triangles []struct {
			A int `xml:"v1,attr"`
			B int `xml:"v2,attr"`
			C int `xml:"v3,attr"`
		} `xml:"triangles>triangle"`
	} `xml:"mesh"`
	Components []xmlItem `xml:"components>component"`
}
type xmlItem struct {
	ID        int    `xml:"objectid,attr"`
	Transform string `xml:"transform,attr"`
}
type matrix [12]float64

func identity() matrix { return matrix{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0} }
func parseMatrix(s string) (matrix, error) {
	m := identity()
	if s == "" {
		return m, nil
	}
	f := strings.Fields(s)
	if len(f) != 12 {
		return m, errors.New("3MF transform needs 12 numbers")
	}
	for i := range m {
		v, e := strconv.ParseFloat(f[i], 64)
		if e != nil || !finite(v) {
			return m, errors.New("invalid 3MF transform")
		}
		m[i] = v
	}
	return m, nil
}
func (m matrix) apply(v Vec) Vec {
	return Vec{v[0]*m[0] + v[1]*m[3] + v[2]*m[6] + m[9], v[0]*m[1] + v[1]*m[4] + v[2]*m[7] + m[10], v[0]*m[2] + v[1]*m[5] + v[2]*m[8] + m[11]}
}
func parse3MF(data []byte) ([]Triangle, error) {
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return nil, fmt.Errorf("invalid 3MF ZIP: %w", e)
	}
	var chosen *zip.File
	var total uint64
	for _, f := range z.File {
		total += f.UncompressedSize64
		if total > 256<<20 {
			return nil, errors.New("3MF expanded archive exceeds 256 MiB")
		}
		if strings.EqualFold(f.Name, "3D/3dmodel.model") {
			chosen = f
		}
	}
	if chosen == nil {
		return nil, errors.New("3MF requires 3D/3dmodel.model; externally referenced build objects are unsupported")
	}
	if chosen.UncompressedSize64 > MaxBytes {
		return nil, errors.New("3MF model XML exceeds 64 MiB")
	}
	r, e := chosen.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	d, e := io.ReadAll(io.LimitReader(r, MaxBytes+1))
	if e != nil {
		return nil, e
	}
	if len(d) > MaxBytes {
		return nil, errors.New("3MF XML exceeds limit")
	}
	var model xmlModel
	if e = xml.Unmarshal(d, &model); e != nil {
		return nil, fmt.Errorf("3MF XML: %w", e)
	}
	factor := 1.
	switch model.Unit {
	case "", "millimeter":
	case "micron":
		factor = .001
	case "centimeter":
		factor = 10
	case "meter":
		factor = 1000
	case "inch":
		factor = 25.4
	case "foot":
		factor = 304.8
	default:
		return nil, errors.New("unknown 3MF length unit")
	}
	objects := map[int]xmlObject{}
	for _, o := range model.Resources.Objects {
		if _, ok := objects[o.ID]; ok {
			return nil, errors.New("duplicate 3MF object ID")
		}
		objects[o.ID] = o
	}
	var out []Triangle
	var expand func(xmlItem, []matrix, map[int]bool) error
	expand = func(item xmlItem, parents []matrix, visiting map[int]bool) error {
		if len(parents) > 32 || visiting[item.ID] {
			return errors.New("cyclic or deeply nested 3MF components")
		}
		o, ok := objects[item.ID]
		if !ok {
			return errors.New("missing 3MF object reference")
		}
		m, e := parseMatrix(item.Transform)
		if e != nil {
			return e
		}
		chain := append([]matrix{m}, parents...)
		visiting[item.ID] = true
		defer delete(visiting, item.ID)
		for _, t := range o.Mesh.Triangles {
			if len(out) >= MaxTriangles {
				return errors.New("3MF exceeds triangle limit")
			}
			var tri Triangle
			for j, i := range []int{t.A, t.B, t.C} {
				if i < 0 || i >= len(o.Mesh.Vertices) {
					return errors.New("3MF triangle vertex index out of range")
				}
				v := o.Mesh.Vertices[i]
				p := Vec{v.X, v.Y, v.Z}
				for _, mat := range chain {
					p = mat.apply(p)
				}
				for k := range p {
					p[k] *= factor
				}
				tri[j] = p
			}
			out = append(out, tri)
		}
		for _, c := range o.Components {
			if e := expand(c, chain, visiting); e != nil {
				return e
			}
		}
		return nil
	}
	if len(model.Build.Items) == 0 {
		return nil, errors.New("3MF has no printable build items")
	}
	for _, i := range model.Build.Items {
		if e := expand(i, nil, map[int]bool{}); e != nil {
			return nil, e
		}
	}
	return out, nil
}
func Analyze(ts []Triangle) (*Mesh, error) {
	if len(ts) == 0 || len(ts) > MaxTriangles {
		return nil, errors.New("empty mesh or too many triangles")
	}
	m := &Mesh{Triangles: ts, Warnings: []string{}, Bounds: Bounds{Min: Vec{math.Inf(1), math.Inf(1), math.Inf(1)}, Max: Vec{math.Inf(-1), math.Inf(-1), math.Inf(-1)}}, MinFeature: math.Inf(1)}
	degenerate := 0
	vol := 0.
	for _, t := range ts {
		for _, v := range t {
			for k, x := range v {
				if !finite(x) || math.Abs(x) > 100000 {
					return nil, errors.New("mesh coordinates must be finite and within ±100,000 mm")
				}
				m.Bounds.Min[k] = math.Min(m.Bounds.Min[k], x)
				m.Bounds.Max[k] = math.Max(m.Bounds.Max[k], x)
			}
		}
		n := cross(sub(t[1], t[0]), sub(t[2], t[0]))
		a := length(n) * .5
		if a < 1e-12 {
			degenerate++
			continue
		}
		m.Area += a
		vol += dot(t[0], cross(t[1], t[2])) / 6
		for j := 0; j < 3; j++ {
			edge := length(sub(t[j], t[(j+1)%3]))
			if edge > 1e-6 {
				m.MinFeature = math.Min(m.MinFeature, edge)
			}
		}
	}
	if m.Area < 1e-9 {
		return nil, errors.New("mesh has no nondegenerate surface")
	}
	for k := 0; k < 3; k++ {
		if m.Bounds.Max[k]-m.Bounds.Min[k] < 1e-6 {
			return nil, errors.New("mesh has zero thickness on an axis")
		}
	}
	m.Volume = math.Abs(vol)
	if m.Volume < 1e-6 {
		m.Warnings = append(m.Warnings, "Signed mesh volume is near zero; mesh may be open or inconsistently wound.")
	}
	if degenerate > 0 {
		m.Warnings = append(m.Warnings, fmt.Sprintf("%d degenerate triangles; native engine may repair them.", degenerate))
	}
	for _, t := range ts {
		n := cross(sub(t[1], t[0]), sub(t[2], t[0]))
		l := length(n)
		if l > 0 && n[2]/l < -.70710678 && (t[0][2]+t[1][2]+t[2][2])/3 > m.Bounds.Min[2]+.1 {
			m.OverhangArea += l * .5
		}
	}
	m.Warnings = append(m.Warnings, "Overhang and feature-size analysis uses triangle normals and shortest mesh edges; thin walls and bridges are delegated to the native engine.")
	return m, nil
}

// Transformed re-bases rotated/scaled bounds so Position is XY centre and Z bottom.
func (m *Mesh) Transformed(t Transform) (*Mesh, error) {
	for i := 0; i < 3; i++ {
		if !finite(t.Scale[i]) || t.Scale[i] <= 0 || t.Scale[i] > 1000 || !finite(t.Rotation[i]) || !finite(t.Position[i]) {
			return nil, errors.New("invalid transform; scale must be positive and at most 1000")
		}
	}
	ts := make([]Triangle, len(m.Triangles))
	var c, s Vec
	for k := 0; k < 3; k++ {
		c[k] = math.Cos(t.Rotation[k] * math.Pi / 180)
		s[k] = math.Sin(t.Rotation[k] * math.Pi / 180)
	}
	for i, tri := range m.Triangles {
		for j, v := range tri {
			x, y, z := v[0]*t.Scale[0], v[1]*t.Scale[1], v[2]*t.Scale[2]
			y, z = y*c[0]-z*s[0], y*s[0]+z*c[0]
			x, z = x*c[1]+z*s[1], -x*s[1]+z*c[1]
			x, y = x*c[2]-y*s[2], x*s[2]+y*c[2]
			ts[i][j] = Vec{x, y, z}
		}
	}
	out, e := Analyze(ts)
	if e != nil {
		return nil, e
	}
	offset := Vec{t.Position[0] - (out.Bounds.Min[0]+out.Bounds.Max[0])/2, t.Position[1] - (out.Bounds.Min[1]+out.Bounds.Max[1])/2, t.Position[2] - out.Bounds.Min[2]}
	for i := range ts {
		for j := range ts[i] {
			for k := 0; k < 3; k++ {
				ts[i][j][k] += offset[k]
			}
		}
	}
	return Analyze(ts)
}
func (m *Mesh) Vertices() []float64 {
	out := make([]float64, 0, len(m.Triangles)*9)
	cx := (m.Bounds.Min[0] + m.Bounds.Max[0]) / 2
	cy := (m.Bounds.Min[1] + m.Bounds.Max[1]) / 2
	for _, t := range m.Triangles {
		for _, v := range t {
			out = append(out, v[0]-cx, v[1]-cy, v[2]-m.Bounds.Min[2])
		}
	}
	return out
}
func (m *Mesh) WriteSTL(w io.Writer) error {
	h := make([]byte, 84)
	copy(h, "Forge Slicer inert mesh export")
	binary.LittleEndian.PutUint32(h[80:], uint32(len(m.Triangles)))
	if _, e := w.Write(h); e != nil {
		return e
	}
	b := make([]byte, 50)
	for _, t := range m.Triangles {
		clear(b)
		for j := 0; j < 9; j++ {
			binary.LittleEndian.PutUint32(b[12+j*4:], math.Float32bits(float32(t[j/3][j%3])))
		}
		if _, e := w.Write(b); e != nil {
			return e
		}
	}
	return nil
}
