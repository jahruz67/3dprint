# Geometric support analysis: method and limits

`internal/analysis.Analyze(ctx, mesh, nozzle, angle)` accepts already transformed build coordinates. It preserves original triangle ordering in reported `riskTriangles`. It neither edits the model nor generates supports. The output includes findings, assumptions, confidence, overhang/unsupported/bridge-like estimates, projected bed contact, sampled vertical gap and sample count.

## Surface orientation and bed contact

For each triangle, compute its cross-product normal and area. A signed-volume sum relative to a mesh vertex establishes the sign of globally consistent winding. If negative, normal classification is reversed consistently. This prevents an entirely reversed cube from being labeled as an unsupported top face. Mixed winding, nonmanifold/intersecting solids and open meshes remain uncertain; a near-zero signed volume produces a low-confidence warning.

Let `a` be the threshold measured from horizontal. A downward face is overhang-eligible when `normalZ < -cos(a)`. Thus a larger threshold includes more steeply inclined faces, matching native Prusa threshold convention. Faces entirely within `max(0.025 mm, 0.1*nozzle)` of the physical bed are excluded. Their projected downward area contributes to bed-contact area. Contact ratio is `min(1, contactArea / XYBoundingFootprint)`; projected triangle areas are summed, not unioned, so overlapping solids may overcount.

## Bounded unsupported-area estimation

A deterministic area-stratified sequence chooses at most 4096 downward-surface samples (at least 128 when eligible faces exist). Triangle points use square-root barycentric mapping and radical-inverse sequences. Each primary sample carries equal surface-area weight. A balanced spatial bounding-volume hierarchy prevents a normal query from scanning every mesh triangle.

Vertical projected-triangle intersection finds the highest upward-facing surface at or just below the sample, excluding the queried face. A coincident upward face from a lower touching solid **counts as contact**; starting every ray strictly below the sample would incorrectly miss the top of a pillar touching a separate bridge bar. If no model surface lies below, bed Z=0 is the fallback. A gap above the contact tolerance contributes to unsupported area. `maxGapMm` is the maximum **sampled vertical** gap; it is not bridge span length.

The input is bounded by the mesh import triangle limit. Ray work is capped at four million triangle checks and eight million node visits, shared by unsupported and bridge analysis. Cancellation is checked during preprocessing, BVH construction and queries. If the work cap is exhausted, confidence becomes low, unexamined eligible area is conservatively counted as potentially unsupported, and a visible inspection finding is returned. The report never treats an incomplete pass as a proof of support.

## Bridge-like and readiness heuristics

At most 512 deterministic unsupported samples receive opposing-anchor checks. For each XY axis, lateral BVH slab queries look for upward geometry near the sample's underside height on both sides, up to 16 mm away per side. Intersecting projected triangle footprints avoids skipping narrow pillars between discrete probe distances. The classified fraction estimates bridge-like area; it cannot establish material cooling, sag, contact strength, toolpath direction, extrusion suitability or physical bridging success. Automatic support recommendation remains conservative even where some region appears bridge-like; the native backend determines actual supports and recognized bridge omission.

Flat bed-contact, height-to-narrow-dimension and shortest-mesh-edge heuristics produce actionable advice. Small projected contact or a height/narrow-dimension ratio above eight suggests reviewing adhesion/orientation/brim. A shortest mesh edge below 0.8 nozzle diameters suggests inspecting detail, but may simply reflect fine tessellation. A complete model extent below nozzle diameter suggests checking scale/units or nozzle choice. These are disclosed heuristics, not a wall-thickness solver, static stability simulation or automatic geometry correction.

The general optimizer receives estimated unsupported area for its analytical support-volume approximation. Actual support comparison is a separate bounded native slicing step over organic and snug variants with identical primary print settings. Output validation and parsed extrusion provide the candidate comparison; no analytical support-area reduction is presented as a measured print improvement.
