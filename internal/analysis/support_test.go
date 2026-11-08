package analysis

import (
	"context"
	"reflect"
	"testing"

	"github.com/forge-slicer/forge/internal/geometry"
)

func boxTriangles(min, max geometry.Vec) []geometry.Triangle {
	v := [8]geometry.Vec{{min[0], min[1], min[2]}, {max[0], min[1], min[2]}, {max[0], max[1], min[2]}, {min[0], max[1], min[2]}, {min[0], min[1], max[2]}, {max[0], min[1], max[2]}, {max[0], max[1], max[2]}, {min[0], max[1], max[2]}}
	faces := [][3]int{{0, 3, 2}, {0, 2, 1}, {4, 5, 6}, {4, 6, 7}, {0, 1, 5}, {0, 5, 4}, {1, 2, 6}, {1, 6, 5}, {2, 3, 7}, {2, 7, 6}, {3, 0, 4}, {3, 4, 7}}
	triangles := make([]geometry.Triangle, len(faces))
	for i, f := range faces {
		triangles[i] = geometry.Triangle{v[f[0]], v[f[1]], v[f[2]]}
	}
	return triangles
}
func mesh(t *testing.T, triangles []geometry.Triangle) *geometry.Mesh {
	t.Helper()
	m, err := geometry.Analyze(triangles)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func finding(r Report, code string) bool {
	for _, f := range r.Findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

func TestCubeHasContactWithoutSupportIncludingReversedWinding(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		ts := boxTriangles(geometry.Vec{10, 10, 0}, geometry.Vec{30, 30, 20})
		if reverse {
			for i := range ts {
				ts[i][0], ts[i][1] = ts[i][1], ts[i][0]
			}
		}
		r, err := Analyze(context.Background(), mesh(t, ts), .4, 45)
		if err != nil {
			t.Fatal(err)
		}
		if r.SupportsRecommended || r.UnsupportedAreaMm2 != 0 || r.OverhangAreaMm2 != 0 || r.BedContactAreaMm2 != 400 || r.BedContactRatio != 1 {
			t.Fatalf("cube incorrectly classified (reversed=%v): %+v", reverse, r)
		}
	}
}
func TestCoincidentTopSurfaceOfTouchingSolidCountsAsSupport(t *testing.T) {
	ts := boxTriangles(geometry.Vec{0, 0, 0}, geometry.Vec{20, 20, 10})
	ts = append(ts, boxTriangles(geometry.Vec{0, 0, 10}, geometry.Vec{20, 20, 12})...)
	r, err := Analyze(context.Background(), mesh(t, ts), .4, 45)
	if err != nil {
		t.Fatal(err)
	}
	if r.OverhangAreaMm2 < 399 || r.Samples == 0 || r.SupportsRecommended || r.UnsupportedAreaMm2 > 1e-7 {
		t.Fatalf("coincident contact treated as air: %+v", r)
	}
}
func TestCantileverFindsUnsupportedAreaAndStableTriangleIndices(t *testing.T) {
	ts := boxTriangles(geometry.Vec{0, 0, 0}, geometry.Vec{5, 10, 20})
	ts = append(ts, boxTriangles(geometry.Vec{0, 0, 20}, geometry.Vec{20, 10, 22})...)
	m := mesh(t, ts)
	r, err := Analyze(context.Background(), m, .4, 45)
	if err != nil {
		t.Fatal(err)
	}
	if !r.SupportsRecommended || r.UnsupportedAreaMm2 < 100 || r.MaxGapMm < 19 || len(r.RiskTriangles) == 0 || !finding(r, "unsupported_regions") {
		t.Fatalf("cantilever risk missed: %+v", r)
	}
	for _, id := range r.RiskTriangles {
		if id < 12 || id >= len(ts) {
			t.Fatalf("risk index does not match original bar triangle ordering: %d", id)
		}
	}
	again, err := Analyze(context.Background(), m, .4, 45)
	if err != nil || !reflect.DeepEqual(r, again) {
		t.Fatal("analysis is not deterministic")
	}
}
func TestBridgeLikeSpanFindsOpposingPillarAnchors(t *testing.T) {
	ts := boxTriangles(geometry.Vec{0, 0, 0}, geometry.Vec{5, 10, 10})
	ts = append(ts, boxTriangles(geometry.Vec{25, 0, 0}, geometry.Vec{30, 10, 10})...)
	ts = append(ts, boxTriangles(geometry.Vec{0, 0, 10}, geometry.Vec{30, 10, 12})...)
	r, err := Analyze(context.Background(), mesh(t, ts), .4, 45)
	if err != nil {
		t.Fatal(err)
	}
	if r.UnsupportedAreaMm2 < 150 || r.BridgeLikeAreaMm2 < 50 || r.BridgeLikeAreaMm2 > r.UnsupportedAreaMm2 || !finding(r, "bridge_like_regions") {
		t.Fatalf("opposing bridge anchors missed: %+v", r)
	}
}
func TestPlacementSlenderContactAndFeatureAdvice(t *testing.T) {
	floating := mesh(t, boxTriangles(geometry.Vec{0, 0, 2}, geometry.Vec{10, 10, 12}))
	r, err := Analyze(context.Background(), floating, .4, 45)
	if err != nil {
		t.Fatal(err)
	}
	if !finding(r, "off_bed") || !r.SupportsRecommended {
		t.Fatalf("floating model not identified: %+v", r)
	}
	slender := mesh(t, boxTriangles(geometry.Vec{0, 0, 0}, geometry.Vec{2, 2, 30}))
	slender.MinFeature = .1
	r, err = Analyze(context.Background(), slender, .4, 45)
	if err != nil {
		t.Fatal(err)
	}
	if !finding(r, "adhesion_stability") || !finding(r, "fine_mesh_detail") {
		t.Fatalf("readiness advice missed: %+v", r)
	}
}
func TestHorizontalSlopeAngleAndSampleBudget(t *testing.T) {
	m := mesh(t, boxTriangles(geometry.Vec{}, geometry.Vec{20, 20, 20}))
	rotated, err := m.Transformed(geometry.Transform{Rotation: geometry.Vec{0, 60, 0}, Scale: geometry.Vec{1, 1, 1}, Position: geometry.Vec{30, 30, 0}})
	if err != nil {
		t.Fatal(err)
	}
	low, err := Analyze(context.Background(), rotated, .4, 20)
	if err != nil {
		t.Fatal(err)
	}
	high, err := Analyze(context.Background(), rotated, .4, 80)
	if err != nil {
		t.Fatal(err)
	}
	if high.OverhangAreaMm2 <= low.OverhangAreaMm2 {
		t.Fatal("higher horizontal-angle threshold must include more steep slopes")
	}
	if high.Samples > MaxSamples {
		t.Fatal("sample budget exceeded")
	}
}
func TestCancellationAndInvalidInputs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Analyze(ctx, nil, .4, 45); err != context.Canceled {
		t.Fatalf("cancellation not propagated: %v", err)
	}
	if _, err := Analyze(context.Background(), nil, .4, 45); err == nil {
		t.Fatal("nil mesh accepted")
	}
	m := mesh(t, boxTriangles(geometry.Vec{}, geometry.Vec{10, 10, 10}))
	if _, err := Analyze(context.Background(), m, 0, 45); err == nil {
		t.Fatal("unknown nozzle accepted")
	}
	if _, err := Analyze(context.Background(), m, .4, 90); err == nil {
		t.Fatal("invalid angle accepted")
	}
}
