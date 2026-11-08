# Evidence levels and limits

| Quantity | Classification | Practical limit |
|---|---|---|
| Input triangle dimensions / surface sums | Direct geometry calculation | Depends on model integrity and transforms. |
| Filament area and density conversion | Exact formula for supplied values | Diameter/density/calibration uncertainty remains. |
| Trapezoidal representative segment time | Exact idealized formula | Actual path, junctions and firmware differ. |
| Shell/infill/support volume | Analytical approximation | No exact offset union, bridging or contact topology. |
| Warm-up/travel overhead | Engineering heuristic | Fixed factors are disclosed, not benchmark calibrated. |
| Quality error | Engineering heuristic | Not validated surface roughness or dimensional accuracy. |
| Structural strength | Unverified engineering proxy | No physical strength guarantee. |
| Post-slice paths / extrusion | Parsed native-engine output | Static file interpretation, not measured physical output. |

No benchmark calibration data is bundled, and no performance, quality, reliability or strength improvement percentages are fabricated. Reference-machine values are clearly simulated. An optimizer report may show a lower calculated objective, but actual print improvements require controlled measurements against the recorded baseline.

The current release uses constant layer height, conservative shell minima, native organic/snug/grid supports and the user-selected orientation. A bounded spatial analysis recomputes unsupported/contact estimates for the placed mesh and labels bridge-like spans heuristically; see [support analysis](support-analysis.md). It does not claim exact support topology, adaptive layers, wall-thickness measurement or load-bearing validation. Geometry features guide settings only to the extent available from actual analysis. Missing detail measurements do not become invented precision claims.
