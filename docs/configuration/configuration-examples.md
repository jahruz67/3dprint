# Complete examples

Begin user setup with the blank template in the profile editor, or `GET /api/printer-template`. It preserves the exact structure while keeping unknown hardware values null. It is intentionally invalid until completed; see [getting started](getting-started.md).

Material recipes and developer fixtures live in [examples/profiles](../../examples/profiles/):

- `generic-pla.json`, `generic-petg.json`, `generic-abs.json`, `generic-asa.json`, `generic-tpu.json`, `generic-pa.json`: the six selectable material starting points, with [manufacturer sources and explicit engineering assumptions](material-catalogue.md).
- [reference-cartesian.json](../../examples/profiles/reference-cartesian.json): developer-only fictional 220 × 220 × 250 mm machine for automated tests. Never default-loaded or offered as a configured printer.
- [reference-pla.json](../../examples/profiles/reference-pla.json): matching developer test fixture; not a default material entry.
- [balanced.json](../../examples/profiles/balanced.json): balanced objective, 256 candidate budget, minimum two walls and 12% infill, supports permitted.

Complete examples include every required field and are available offline in the repository/documentation bundle. Fixture `provenance` labels identify simulated values. They are not user setup templates: do not relabel a fixture as a real machine. User printer setup starts empty and requires measured or published hardware values.

For a Bowden conversion, change `extruders[0].type` to `bowden` and independently calibrate retraction distance/speed; the type does not invent a preset. For an unheated bed, set `hasHeatedBed:false`, `minBed:0`, `maxBed:0` and material `bedTemperature:0`. For a firmware change, update family, version, matching flavor and verified command list as one reviewed change. For a changed nozzle, update nozzle diameter and reverify material layer range/flow; do not reuse a flow claim from a different nozzle without evidence.

Factory/default flow caps are accepted. New drafts start with a configurable 4 mm3/s cap; calibration is optional. `flowVerified` can be omitted, false or true without blocking import or slicing, and flow provenance is optional. Numeric cap bounds and the smaller machine/material flow limit remain enforced. If a command is unsupported, placing it in a startup line is still an error.
