# Validation rules

Parsing is deliberately stricter than permissive JSON readers. Profiles are limited to 1 MiB, contain exactly one JSON document, have no duplicate object keys, no unknown properties, no `NaN`/infinity or overflowing exponent, and satisfy the bundled JSON Schema Draft 2020-12 using the mature `santhosh-tekuri/jsonschema/v6` library. Schemas are embedded and compiled locally once. Profiles never select their own schemas or trigger a network fetch.

Required fields, object/array structure, enumerated strings, numerical bounds and optional three-component load vectors are checked by the schema. `additionalProperties:false` is intentional: spelling mistakes must not silently disable constraints. Numerical schema bounds are capability envelope limits, not recommendations for a physical machine. The public schemas are independent of the UI and slicing engine.

Cross-field semantic checks additionally require:

- Exactly one extruder, a rectangular bed and front-left origin.
- Nozzle diameter smaller than filament diameter; a positive numeric flow cap and documentation for build volume, motion and thermal limits. Factory/default flow values are allowed without verified-flow metadata or calibration evidence.
- Nozzle minimum below maximum; bed minimum at most maximum; unheated beds have both limits zero.
- Every excluded rectangle lies completely within the bed.
- Retraction speed does not exceed the configured machine speed.
- `marlin` pairs with `marlin2`; `klipper` with `klipper`; `reprap` with `reprapfirmware`.
- Supported and unsupported command sets never overlap. The core command contract includes `G0`, `G1`, `G21`, `G28`, `G90`, `G92`, `M84`, `M104`, `M109`, `M107`, `M204`, `M400`; the configured extrusion mode requires `M82` or `M83`. Heated beds require `M140`/`M190`, and part cooling requires `M106`.
- Firmware retraction is false for this adapter. All retractions are explicit extrusion moves.
- Startup/shutdown lines are literal single lines containing only explicitly supported G/M/T commands. Template braces, macros, shell substitutions, escapes and multiline strings are rejected. Full generated G-code is validated separately before export; profile validation alone is insufficient.
- Critical provenance values are nonempty and not `unknown`. An application cannot verify the truth of a source string: the human remains responsible for its evidence.

Composition checks material temperatures against machine ranges, rejects requested cooling without a fan, and requires an intersection between material layer range and `[0.2 nozzle diameter, 0.8 nozzle diameter]`. Machine flow, speed, travel and acceleration cap the optimizer. Model placement and actual G-code bounds are checked by the application and output validator, not by a profile-only JSON Schema.

Optional support preferences accept only the documented mode/style/placement enums and 20–80 degree angle. Omitted support mode preserves the legacy `supports` boolean behavior; omitted angle means 45 degrees. `Filament.notes` is an optional bounded list of human-readable warnings, and the `PLA+` material category does not imply a manufacturer temperature. The explicitly requested custom PLA+ 115 C nozzle recipe must still satisfy the printer's original thermal range. Its warnings do not grant an exception to composition or firmware safety checks.

`config.ParsePrinter`, `ParseFilament`, `ParsePreferences`, `ValidatePrinter`, `ValidateFilament`, `ValidatePreferences` are public internal APIs for services. `config.Compose` applies precedence and cross-profile checks. `ValidateJSON` performs formal schema checks; callers handling printer or filament data must also use their semantic parser. Error text names the offending schema path or cross-field dependency. Imported configurations are data and are never passed to a shell.

## Manual validation

The settings editor validates imports before saving. With the standalone local server running, send a same-origin JSON request to `POST /api/profiles/validate`:

```json
{"kind":"optimization","profile":{"schemaVersion":"1.0","mode":"balanced","budget":256,"minWalls":2,"minInfill":0.12,"supports":true}}
```

Expected structure is `{ "valid": true, "errors": [] }`, or `valid:false` with actionable errors. Use the printer or filament kind for the corresponding profile. Network clients must comply with the local server's origin/token policy. Do not relax the schema to make an unknown machine appear supported.
