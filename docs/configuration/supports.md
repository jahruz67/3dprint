# Smart supports and print readiness

Forge analyzes the **placed, rotated and scaled mesh**, rather than enabling supports from a model name or a global preset. Downward-facing surfaces are sampled by area; spatial ray queries distinguish air gaps from lower model geometry and from coincident contact surfaces of touching solids. The report highlights original mesh triangle indices associated with sampled unsupported areas and explains placement, adhesion, detail and bridge-like concerns. These findings are engineering guidance; the native slicer still determines actual support regions and printable paths.

## Choose a support goal

| `supportMode` | Behavior | Tradeoff |
|---|---|---|
| `auto` | Enable support processing when placed-mesh analysis recommends it; when enabled and shape is Auto, compare organic and snug native outputs. | A normal cube on the bed needs no support comparison. Sampled geometry cannot prove that every tiny unsupported feature was found. |
| `optimized` | Permit support generation and compare organic/snug when shape is Auto, using actual validated native outputs. | Extra native slice work buys a calculated comparison, not a global optimum. |
| `easy_remove` | Use a larger contact gap and XY clearance, sparse interface spacing, and organic shape unless explicitly overridden. | Aims to ease removal; supported underside finish and small-detail support can suffer. Removal force is not measured. |
| `stable` | Use a smaller contact gap, denser spacing, three interface layers and support sheath; grid shape unless overridden. | Aims to make support structures more robust; can use more material and be harder to remove. Stability is not physically validated. |
| `off` | Explicitly disable support generation. | Unsupported-region findings remain visible and should be reviewed. |

An explicit mode takes precedence over the older `supports` boolean. For an older profile with no mode, `supports:true` means Auto and `supports:false` means Off. The original boolean remains required for version-1 compatibility. This additive profile revision retains `schemaVersion:"1.0"`.

The candidate comparison holds the model, layers, walls, infill and speed settings fixed. Each native result must pass G-code validation. Among valid candidates, selection minimizes `0.75 * supportFilament / referenceSupportFilament + 0.25 * printTime / referencePrintTime`, with small denominator guards; the first valid candidate is the reference. The report shows actual calculated support-filament length and native print-time estimates, identifies rejected candidates and explains the winner. It does not claim measured removal force, proven mechanical stability or exhaustively optimized supports.

## Shape, placement and threshold

`supportStyle` can be `auto`, `organic`, `snug` or `grid`. Organic uses branching native supports; snug follows the overhang footprint closely; grid uses a conventional support structure. Explicit shape choices are respected and skip automatic shape comparison. Only supported styles are sent to the installed native backend; rejected native results cannot be exported.

`supportPlacement` is `everywhere` by default or `buildplate`. Everywhere can anchor on lower model geometry. Build-plate-only can leave enclosed or obstructed overhangs without reachable support; inspect those regions rather than assuming the setting is universally better.

`supportAngle` is 20–80 degrees, default 45, **measured from the horizontal plane** to match PrusaSlicer's threshold convention. Increasing it includes steeper faces and can request more support. It is not an angle measured from vertical. The field is optional: omitted/internal zero normalizes to 45; an explicitly supplied JSON angle must be within 20–80.

For a baseline layer height `h` and nozzle diameter `d`, current starting recipes are:

| Recipe | Z contact gap | XY clearance | Support spacing | Interface layers / spacing | Sheath |
|---|---|---|---|---|---|
| Auto / Optimized | `max(0.12,0.65h)` mm | `max(0.25,0.7d)` mm | 3 mm | 2 / 0.2 mm | Off |
| Easy removal | `max(0.18,0.9h)` mm | `max(0.35,d)` mm | 3 mm | 2 / 0.3 mm | Off |
| Stable | `max(0.10,0.5h)` mm | `max(0.20,0.5d)` mm | 1.8 mm | 3 / 0 mm | On |

These are explicit engineering starting recipes, recalculated with the selected layer height, not calibrated guarantees for a particular material. Native recognized bridges are excluded from support generation by the backend's bridge policy; Forge's own bridge-like label remains a separate geometric heuristic.

## Readiness findings before and after slicing

The report can recommend: move the model onto the bed, try a broader orientation, consider a brim for limited contact/slender geometry, inspect detail relative to the nozzle, verify scale/units, or allow supports. A very short triangulation edge is **not** a measured thin wall. Contact area is a projected triangle estimate and can overcount overlapping solids. Brim and orientation recommendations do not silently change your placement or printer profile.

Review material-specific notes as well as geometry. In particular, the custom PLA+ 115 C **nozzle** preset is intentionally labeled unverified and does not alter machine thermal limits; see [material catalogue](material-catalogue.md). A geometric readiness finding cannot establish that a requested material temperature will extrude correctly.

The algorithm is bounded and cancellable. Low confidence or a work-limit finding means inspect the native preview, not “all clear.” For numerical definitions, bounds and assumptions see [support analysis](../optimization/support-analysis.md). No support choice is described as safe or optimal without corresponding evidence.
