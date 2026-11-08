// Package analysis provides bounded geometric readiness advice. It does not
// replace the native engine's exact support/toolpath generation or physical tests.
package analysis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/forge-slicer/forge/internal/geometry"
)

const MaxSamples = 4096
const maxTriangleChecks = 4000000
const maxNodeVisits = 8000000

type Finding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Action   string `json:"action"`
}
type Report struct {
	SupportsRecommended bool      `json:"supportsRecommended"`
	OverhangAreaMm2     float64   `json:"overhangAreaMm2"`
	UnsupportedAreaMm2  float64   `json:"unsupportedAreaMm2"`
	BridgeLikeAreaMm2   float64   `json:"bridgeLikeAreaMm2"`
	BedContactAreaMm2   float64   `json:"bedContactAreaMm2"`
	BedContactRatio     float64   `json:"bedContactRatio"`
	MaxGapMm            float64   `json:"maxGapMm"`
	Samples             int       `json:"samples"`
	Confidence          string    `json:"confidence"`
	Findings            []Finding `json:"findings"`
	Assumptions         []string  `json:"assumptions"`
	RiskTriangles       []int     `json:"riskTriangles"`
}

type face struct {
	id       int
	box      geometry.Bounds
	center   geometry.Vec
	area, nz float64
}
type node struct {
	box                     geometry.Bounds
	start, end, left, right int
}
type index struct {
	ctx            context.Context
	mesh           *geometry.Mesh
	faces          []face
	order          []int
	nodes          []node
	checks, visits int
	limited        bool
	err            error
	sign           float64
}
type sample struct {
	p      geometry.Vec
	face   int
	weight float64
}

func sub(a, b geometry.Vec) geometry.Vec { return geometry.Vec{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func cross(a, b geometry.Vec) geometry.Vec {
	return geometry.Vec{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func dot(a, b geometry.Vec) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func finite(v float64) bool         { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func emptyBounds() geometry.Bounds {
	return geometry.Bounds{Min: geometry.Vec{math.Inf(1), math.Inf(1), math.Inf(1)}, Max: geometry.Vec{math.Inf(-1), math.Inf(-1), math.Inf(-1)}}
}
func include(b *geometry.Bounds, v geometry.Vec) {
	for k := 0; k < 3; k++ {
		b.Min[k] = math.Min(b.Min[k], v[k])
		b.Max[k] = math.Max(b.Max[k], v[k])
	}
}

// Analyze inspects an already transformed/placed triangle mesh. Angle is the
// overhang threshold measured from the horizontal plane (higher includes steeper
// faces). Ray samples are deterministic and weighted by true triangle area.
func Analyze(ctx context.Context, m *geometry.Mesh, nozzle, angle float64) (Report, error) {
	r := Report{Confidence: "moderate", Findings: []Finding{}, Assumptions: []string{}, RiskTriangles: []int{}}
	if ctx == nil {
		return r, errors.New("analysis requires a context")
	}
	if err := ctx.Err(); err != nil {
		return r, err
	}
	if m == nil || len(m.Triangles) == 0 || len(m.Triangles) > geometry.MaxTriangles {
		return r, errors.New("analysis requires a bounded nonempty mesh")
	}
	if !finite(nozzle) || nozzle < .1 || nozzle > 2 {
		return r, errors.New("analysis nozzle must be within 0.1–2 mm")
	}
	if angle == 0 {
		angle = 45
	}
	if !finite(angle) || angle < 20 || angle > 80 {
		return r, errors.New("support angle must be within 20–80 degrees")
	}
	x := index{ctx: ctx, mesh: m, faces: make([]face, len(m.Triangles)), order: make([]int, len(m.Triangles)), nodes: make([]node, 0, len(m.Triangles)/4), sign: 1}
	box := emptyBounds()
	reference := m.Triangles[0][0]
	signedVolume, totalArea := 0., 0.
	for i, t := range m.Triangles {
		if i%2048 == 0 {
			if err := ctx.Err(); err != nil {
				return r, err
			}
		}
		b := emptyBounds()
		c := geometry.Vec{}
		for _, v := range t {
			for k, a := range v {
				if !finite(a) || math.Abs(a) > 1000000 {
					return r, errors.New("analysis coordinates must be finite and bounded to ±1,000,000 mm")
				}
				c[k] += a / 3
			}
			include(&b, v)
			include(&box, v)
		}
		n := cross(sub(t[1], t[0]), sub(t[2], t[0]))
		length := math.Sqrt(dot(n, n))
		nz := 0.
		if length > 1e-12 {
			nz = n[2] / length
		}
		x.faces[i] = face{id: i, box: b, center: c, area: length * .5, nz: nz}
		x.order[i] = i
		totalArea += length * .5
		signedVolume += dot(sub(t[0], reference), cross(sub(t[1], reference), sub(t[2], reference))) / 6
	}
	if totalArea <= 1e-9 {
		return r, errors.New("mesh has no usable surface area")
	}
	if signedVolume < 0 {
		x.sign = -1
	}
	if math.Abs(signedVolume) < 1e-7 {
		r.Confidence = "low"
		r.Findings = append(r.Findings, Finding{Code: "winding_uncertain", Severity: "warning", Title: "Surface orientation is uncertain", Detail: "Signed volume is near zero. Open or inconsistently wound meshes can invalidate normal-based support classification.", Action: "inspect"})
	}
	tol := math.Max(.025, nozzle*.1)
	threshold := -math.Cos(angle * math.Pi / 180)
	eligible := make([]int, 0)
	cumulative := make([]float64, 0)
	for i, f := range x.faces {
		nz := f.nz * x.sign
		if nz < -.05 && f.box.Min[2] >= -tol && f.box.Max[2] <= tol {
			r.BedContactAreaMm2 += f.area * (-nz)
		}
		if nz < threshold && f.area > 1e-12 && f.box.Max[2] > tol {
			r.OverhangAreaMm2 += f.area
			eligible = append(eligible, i)
			cumulative = append(cumulative, r.OverhangAreaMm2)
		}
	}
	w, d, h := box.Max[0]-box.Min[0], box.Max[1]-box.Min[1], box.Max[2]-box.Min[2]
	footprint := w * d
	if footprint > 1e-10 {
		r.BedContactRatio = math.Min(1, r.BedContactAreaMm2/footprint)
	}
	r.Assumptions = []string{
		fmt.Sprintf("Overhang angle %.1f° is measured from the horizontal plane; bed-coincident faces within %.3f mm are excluded.", angle, tol),
		"Normals are corrected for globally reversed winding using signed volume; mixed-winding, intersecting or open meshes remain uncertain.",
		"Unsupported area is a bounded deterministic area-weighted sample estimate, not an exact union or native support region.",
		"Vertical rays seek upward-facing geometry below each sample, including coincident contact faces of separate touching solids; the bed is the final fallback.",
		"Bridge-like regions use opposing lateral anchor probes at similar height; this heuristic does not establish physically printable bridging.",
		"Bed contact sums projected near-bed triangles, capped only in its footprint ratio; overlapping/nonmanifold solids can overcount contact area.",
		"Shortest mesh edges are a tessellation/detail proxy and are not measured wall thickness. The native engine determines actual printable features and supports.",
	}
	if box.Min[2] > tol {
		r.Findings = append(r.Findings, Finding{Code: "off_bed", Severity: "warning", Title: "Model is above the build plate", Detail: fmt.Sprintf("The lowest model point is %.3f mm above Z=0. Place it on the bed or deliberately review supported printing.", box.Min[2]), Action: "place_on_bed"})
	}
	if box.Min[2] < -tol {
		r.Findings = append(r.Findings, Finding{Code: "below_bed", Severity: "error", Title: "Model extends below the bed", Detail: fmt.Sprintf("The model reaches Z=%.3f mm; correct its placement before slicing.", box.Min[2]), Action: "place_on_bed"})
	}
	if box.Min[2] <= tol && r.BedContactAreaMm2 <= nozzle*nozzle*.25 {
		r.Findings = append(r.Findings, Finding{Code: "minimal_contact", Severity: "warning", Title: "Very little flat bed contact", Detail: "The mesh has little horizontal contact at Z=0. Consider a flatter orientation and inspect the first layer; curved contact is approximated by triangle sampling.", Action: "orient"})
	}
	if r.BedContactAreaMm2 > 0 && (r.BedContactRatio < .12 || h/math.Max(nozzle, math.Min(w, d)) > 8) {
		r.Findings = append(r.Findings, Finding{Code: "adhesion_stability", Severity: "warning", Title: "Review adhesion and tipping resistance", Detail: fmt.Sprintf("Estimated bed contact is %.1f%% of the XY bounding footprint; model height is %.1f mm. A broader orientation or a brim may help, but stability is not physically simulated.", r.BedContactRatio*100, h), Action: "brim"})
	}
	if m.MinFeature > 0 && m.MinFeature < nozzle*.8 {
		r.Findings = append(r.Findings, Finding{Code: "fine_mesh_detail", Severity: "info", Title: "Small mesh detail relative to the nozzle", Detail: fmt.Sprintf("The shortest reported mesh edge is %.3f mm versus a %.3f mm nozzle. This can be tessellation and is not a wall-thickness measurement; inspect thin features in the native layer preview.", m.MinFeature, nozzle), Action: "inspect"})
	}
	if math.Min(w, math.Min(d, h)) < nozzle {
		r.Findings = append(r.Findings, Finding{Code: "small_extent", Severity: "warning", Title: "A model dimension is smaller than the nozzle", Detail: "Inspect dimensions and units. Scaling or a verified smaller nozzle may be appropriate; do not enlarge a dimension that is functionally constrained.", Action: "scale"})
	}
	if len(eligible) == 0 {
		return r, nil
	}
	if _, err := x.build(0, len(x.order)); err != nil {
		return r, err
	}
	n := int(math.Ceil(r.OverhangAreaMm2 / math.Max(.25, nozzle*nozzle*4)))
	n = max(128, min(MaxSamples, n))
	weight := r.OverhangAreaMm2 / float64(n)
	risk := map[int]bool{}
	unsupported := make([]sample, 0, n)
	for i := 0; i < n; i++ {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		target := (float64(i) + .5) * weight
		j := sort.SearchFloat64s(cumulative, target)
		if j >= len(eligible) {
			j = len(eligible) - 1
		}
		faceID := eligible[j]
		p := point(m.Triangles[faceID], i+1)
		z, found := x.below(p, tol, faceID)
		if x.err != nil {
			return r, x.err
		}
		if x.limited {
			break
		}
		r.Samples++
		if !found {
			z = 0
		}
		gap := math.Max(0, p[2]-z)
		r.MaxGapMm = math.Max(r.MaxGapMm, gap)
		if gap > tol {
			r.UnsupportedAreaMm2 += weight
			risk[faceID] = true
			unsupported = append(unsupported, sample{p: p, face: faceID, weight: weight})
		}
	}
	// Probe a uniform deterministic subsample of unsupported samples. Count the
	// fraction classified as bridge-like, then scale by sampled unsupported area.
	bridges, probed := 0, 0
	stride := max(1, (len(unsupported)+511)/512)
	for i := 0; i < len(unsupported) && !x.limited; i += stride {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		s := unsupported[i]
		ok := x.bridge(s.p, tol, nozzle, s.face)
		if x.err != nil {
			return r, x.err
		}
		if x.limited {
			break
		}
		probed++
		if ok {
			bridges++
		}
	}
	if probed > 0 {
		r.BridgeLikeAreaMm2 = r.UnsupportedAreaMm2 * float64(bridges) / float64(probed)
	}
	if x.limited {
		r.Confidence = "low"
		remaining := r.OverhangAreaMm2 - float64(r.Samples)*weight
		r.UnsupportedAreaMm2 = math.Min(r.OverhangAreaMm2, r.UnsupportedAreaMm2+math.Max(0, remaining))
		r.Findings = append(r.Findings, Finding{Code: "analysis_budget", Severity: "warning", Title: "Support analysis reached its work limit", Detail: "Dense overlapping geometry exhausted the bounded ray budget. Unexamined overhang area is conservatively counted as potentially unsupported; inspect the native preview.", Action: "inspect"})
	}
	for id := range risk {
		r.RiskTriangles = append(r.RiskTriangles, id)
	}
	sort.Ints(r.RiskTriangles)
	r.SupportsRecommended = r.UnsupportedAreaMm2 > math.Max(nozzle*nozzle*.5, r.OverhangAreaMm2*.002)
	if r.SupportsRecommended {
		r.Findings = append(r.Findings, Finding{Code: "unsupported_regions", Severity: "warning", Title: "Unsupported regions need review", Detail: fmt.Sprintf("Area-weighted probes estimate %.1f mm² of downward surface above an air gap (maximum sampled vertical gap %.2f mm). Allow automatic supports or try a different orientation; the native engine determines exact placement.", r.UnsupportedAreaMm2, r.MaxGapMm), Action: "supports"})
	}
	if r.BridgeLikeAreaMm2 > nozzle*nozzle {
		r.Findings = append(r.Findings, Finding{Code: "bridge_like_regions", Severity: "info", Title: "Some spans have opposing anchors", Detail: fmt.Sprintf("About %.1f mm² of sampled unsupported surface has nearby opposing-height anchors. It may be bridge-like; cooling, material, span length and native toolpaths still determine whether supports can be omitted.", r.BridgeLikeAreaMm2), Action: "inspect"})
	}
	return r, nil
}

func radical(n, base int) float64 {
	v, f := 0., 1./float64(base)
	for n > 0 {
		v += float64(n%base) * f
		n /= base
		f /= float64(base)
	}
	return v
}
func point(t geometry.Triangle, i int) geometry.Vec {
	a := math.Sqrt(radical(i, 2))
	b := radical(i, 3)
	weights := [3]float64{1 - a, a * (1 - b), a * b}
	p := geometry.Vec{}
	for j, v := range t {
		for k := 0; k < 3; k++ {
			p[k] += v[k] * weights[j]
		}
	}
	return p
}

func (x *index) build(start, end int) (int, error) {
	if err := x.ctx.Err(); err != nil {
		return 0, err
	}
	b := emptyBounds()
	cb := emptyBounds()
	for _, id := range x.order[start:end] {
		f := x.faces[id]
		include(&b, f.box.Min)
		include(&b, f.box.Max)
		include(&cb, f.center)
	}
	id := len(x.nodes)
	x.nodes = append(x.nodes, node{box: b, start: start, end: end, left: -1, right: -1})
	if end-start <= 12 {
		return id, nil
	}
	axis := 0
	for k := 1; k < 3; k++ {
		if cb.Max[k]-cb.Min[k] > cb.Max[axis]-cb.Min[axis] {
			axis = k
		}
	}
	sort.Slice(x.order[start:end], func(a, b int) bool {
		fa, fb := x.faces[x.order[start+a]], x.faces[x.order[start+b]]
		if fa.center[axis] == fb.center[axis] {
			return fa.id < fb.id
		}
		return fa.center[axis] < fb.center[axis]
	})
	mid := (start + end) / 2
	left, err := x.build(start, mid)
	if err != nil {
		return 0, err
	}
	right, err := x.build(mid, end)
	if err != nil {
		return 0, err
	}
	x.nodes[id].left = left
	x.nodes[id].right = right
	return id, nil
}

// below finds the highest upward-facing surface at or below the query height.
// Coincident top faces count as contact, unlike a ray started strictly below the
// current face, which would miss touching multi-solid bridge/pillar geometry.
func (x *index) below(p geometry.Vec, tol float64, exclude int) (float64, bool) {
	best, found := 0., false
	var visit func(int)
	visit = func(id int) {
		if x.limited || x.err != nil {
			return
		}
		x.visits++
		if x.visits > maxNodeVisits {
			x.limited = true
			return
		}
		if x.visits%2048 == 0 {
			x.err = x.ctx.Err()
			if x.err != nil {
				return
			}
		}
		n := x.nodes[id]
		b := n.box
		if p[0] < b.Min[0]-1e-9 || p[0] > b.Max[0]+1e-9 || p[1] < b.Min[1]-1e-9 || p[1] > b.Max[1]+1e-9 || b.Min[2] > p[2]+tol || (found && b.Max[2] < best) {
			return
		}
		if n.left >= 0 {
			visit(n.left)
			visit(n.right)
			return
		}
		for _, faceID := range x.order[n.start:n.end] {
			x.checks++
			if x.checks > maxTriangleChecks {
				x.limited = true
				return
			}
			if faceID == exclude {
				continue
			}
			f := x.faces[faceID]
			if f.nz*x.sign <= .025 {
				continue
			}
			t := x.mesh.Triangles[faceID]
			den := (t[1][1]-t[2][1])*(t[0][0]-t[2][0]) + (t[2][0]-t[1][0])*(t[0][1]-t[2][1])
			if math.Abs(den) < 1e-15 {
				continue
			}
			a := ((t[1][1]-t[2][1])*(p[0]-t[2][0]) + (t[2][0]-t[1][0])*(p[1]-t[2][1])) / den
			bb := ((t[2][1]-t[0][1])*(p[0]-t[2][0]) + (t[0][0]-t[2][0])*(p[1]-t[2][1])) / den
			c := 1 - a - bb
			if a < -1e-8 || bb < -1e-8 || c < -1e-8 {
				continue
			}
			z := a*t[0][2] + bb*t[1][2] + c*t[2][2]
			if z <= p[2]+tol && z >= -tol && (!found || z > best) {
				best = z
				found = true
			}
		}
	}
	visit(0)
	return best, found
}
func (x *index) bridge(p geometry.Vec, tol, nozzle float64, exclude int) bool {
	for axis := 0; axis < 2; axis++ {
		if x.anchor(p, axis, -1, tol, nozzle, exclude) && x.anchor(p, axis, 1, tol, nozzle, exclude) {
			return true
		}
		if x.limited || x.err != nil {
			return false
		}
	}
	return false
}

// anchor searches a narrow lateral slab for upward geometry at a similar height.
// Intersecting the projected triangle with the probe line avoids missing narrow
// pillars between a fixed set of discrete ray offsets. The 16 mm reach in each
// direction is a disclosed heuristic boundary, not a certified bridge limit.
func (x *index) anchor(p geometry.Vec, axis int, sign, tol, nozzle float64, exclude int) bool {
	other := 1 - axis
	near, far := math.Max(.25, nozzle*.5), 16.
	band := math.Max(tol, nozzle*.5)
	lo, hi := p[axis]+sign*near, p[axis]+sign*far
	if lo > hi {
		lo, hi = hi, lo
	}
	var visit func(int) bool
	visit = func(id int) bool {
		if x.limited || x.err != nil {
			return false
		}
		x.visits++
		if x.visits > maxNodeVisits {
			x.limited = true
			return false
		}
		if x.visits%2048 == 0 {
			x.err = x.ctx.Err()
			if x.err != nil {
				return false
			}
		}
		n := x.nodes[id]
		b := n.box
		if b.Max[axis] < lo || b.Min[axis] > hi || p[other] < b.Min[other]-1e-9 || p[other] > b.Max[other]+1e-9 || b.Max[2] < p[2]-band || b.Min[2] > p[2]+band {
			return false
		}
		if n.left >= 0 {
			return visit(n.left) || visit(n.right)
		}
		for _, faceID := range x.order[n.start:n.end] {
			x.checks++
			if x.checks > maxTriangleChecks {
				x.limited = true
				return false
			}
			if faceID == exclude || x.faces[faceID].nz*x.sign <= .025 {
				continue
			}
			t := x.mesh.Triangles[faceID]
			type hit struct{ a, z float64 }
			hits := make([]hit, 0, 6)
			for i := 0; i < 3; i++ {
				a, b := t[i], t[(i+1)%3]
				delta := b[other] - a[other]
				if math.Abs(delta) < 1e-12 {
					if math.Abs(p[other]-a[other]) < 1e-8 {
						hits = append(hits, hit{a[axis], a[2]}, hit{b[axis], b[2]})
					}
					continue
				}
				u := (p[other] - a[other]) / delta
				if u >= -1e-9 && u <= 1+1e-9 {
					hits = append(hits, hit{a[axis] + u*(b[axis]-a[axis]), a[2] + u*(b[2]-a[2])})
				}
			}
			if len(hits) < 2 {
				continue
			}
			first, last := hits[0], hits[0]
			for _, q := range hits[1:] {
				if q.a < first.a {
					first = q
				}
				if q.a > last.a {
					last = q
				}
			}
			a, c := math.Max(lo, first.a), math.Min(hi, last.a)
			if a > c || last.a-first.a < 1e-12 {
				continue
			}
			za := first.z + (a-first.a)/(last.a-first.a)*(last.z-first.z)
			zc := first.z + (c-first.a)/(last.a-first.a)*(last.z-first.z)
			if math.Max(za, zc) >= p[2]-band && math.Min(za, zc) <= p[2]+band {
				return true
			}
		}
		return false
	}
	return visit(0)
}
