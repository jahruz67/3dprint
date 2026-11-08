package geometry

import (
	"archive/zip"
	"bytes"
	"math"
	"testing"
)

func cube(size float64) *Mesh {
	v := []Vec{{0, 0, 0}, {size, 0, 0}, {size, size, 0}, {0, size, 0}, {0, 0, size}, {size, 0, size}, {size, size, size}, {0, size, size}}
	faces := [][3]int{{0, 2, 1}, {0, 3, 2}, {4, 5, 6}, {4, 6, 7}, {0, 1, 5}, {0, 5, 4}, {1, 2, 6}, {1, 6, 5}, {2, 3, 7}, {2, 7, 6}, {3, 0, 4}, {3, 4, 7}}
	var ts []Triangle
	for _, f := range faces {
		ts = append(ts, Triangle{v[f[0]], v[f[1]], v[f[2]]})
	}
	m, _ := Analyze(ts)
	return m
}
func TestBinarySTLRoundTrip(t *testing.T) {
	m := cube(20)
	var b bytes.Buffer
	if e := m.WriteSTL(&b); e != nil {
		t.Fatal(e)
	}
	got, e := Parse("cube.stl", b.Bytes())
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(got.Volume-8000) > 1e-6 || got.Area != 2400 || len(got.Triangles) != 12 {
		t.Fatalf("bad geometry: %+v", got)
	}
	if got.OverhangArea != 0 {
		t.Fatal("bottom face must not require supports")
	}
}
func TestTransformRebasesRotationAndScale(t *testing.T) {
	m, e := cube(20).Transformed(Transform{Rotation: Vec{90, 0, 45}, Scale: Vec{2, 1, 1}, Position: Vec{100, 100, 0}})
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(m.Bounds.Min[2]) > 1e-9 || math.Abs(m.Volume-16000) > 1e-6 || math.Abs(m.Bounds.Min[0]+m.Bounds.Max[0]-200) > 1e-9 {
		t.Fatalf("incorrect rebasing: %+v", m.Bounds)
	}
}
func TestMalformedMeshesRejected(t *testing.T) {
	for _, s := range []string{"", "solid bad\nvertex NaN 0 0\nvertex 1 0 0\nvertex 0 1 0\nendsolid", "solid bad\nvertex 0 0 0\nvertex 1 0 0\nvertex 0 1 0\nendsolid"} {
		if _, e := Parse("bad.stl", []byte(s)); e == nil {
			t.Fatalf("accepted malformed mesh %q", s)
		}
	}
}
func Test3MFBuildTransformAndUnit(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("3D/3dmodel.model")
	w.Write([]byte(`<model unit="centimeter"><resources><object id="1"><mesh><vertices><vertex x="0" y="0" z="0"/><vertex x="1" y="0" z="0"/><vertex x="0" y="1" z="1"/></vertices><triangles><triangle v1="0" v2="1" v3="2"/></triangles></mesh></object></resources><build><item objectid="1" transform="1 0 0 0 1 0 0 0 1 2 3 4"/></build></model>`))
	z.Close()
	m, e := Parse("part.3mf", b.Bytes())
	if e != nil {
		t.Fatal(e)
	}
	if m.Bounds.Min != (Vec{20, 30, 40}) || m.Bounds.Max != (Vec{30, 40, 50}) {
		t.Fatalf("wrong 3MF units or transform: %+v", m.Bounds)
	}
}
func Test3MFCycleRejected(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("3D/3dmodel.model")
	w.Write([]byte(`<model><resources><object id="1"><components><component objectid="1"/></components></object></resources><build><item objectid="1"/></build></model>`))
	z.Close()
	if _, e := Parse("cycle.3mf", b.Bytes()); e == nil {
		t.Fatal("cyclic components accepted")
	}
}
func BenchmarkAnalyze(b *testing.B) {
	m := cube(20)
	for i := 0; i < b.N; i++ {
		if _, e := Analyze(m.Triangles); e != nil {
			b.Fatal(e)
		}
	}
}
