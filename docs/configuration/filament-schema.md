# filament schema parameter reference

Generated from the bundled JSON Schema by `schemas/generate.ps1`. Normative rules are the schema plus validation-rules.md. Every field is required unless marked optional. No implicit safety-critical defaults exist. All examples are illustrative and require hardware verification.

## `schemaVersion`

Schema compatibility version. Unknown versions require an explicit future migration.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Required constant: 1.0.
- Example: `1.0`.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `id`

Portable identifier, unique in its profile category.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Pattern: `^[a-z0-9][a-z0-9._-]{0,63}$`.
- Example: `my-printer-main`.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `name`

Human-readable filament identity.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `material`

Polymer family; does not infer temperatures. PLA+ identifies a modified PLA family and is mapped to PLA for native engine categorization.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: PLA, PLA+, PETG, ABS, ASA, TPU, PA, PC, other.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `density`

Manufacturer or measured bulk polymer density.

- Type: number. Unit: g/cm3. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.5 to 3.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `nozzleTemperature`

Chosen nozzle setpoint, constrained by machine thermal limits. Custom unverified values require explicit notes and process review; they do not weaken machine limits.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 100 to 450.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `bedTemperature`

Chosen bed setpoint; must be zero for unheated beds.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 180.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `maxVolumetricFlow`

Material-specific conservative throughput; effective flow is min(machine,material).

- Type: number. Unit: mm3/s. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.1 to 100.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `minLayerHeight`

Minimum acceptable layer height for this material.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.02 to 1.5.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `maxLayerHeight`

Maximum acceptable layer height, also capped at 0.8 nozzle diameters.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.02 to 1.5.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `fanPercent`

Part cooling setpoint; positive requires verified machine fan.

- Type: integer. Unit: %. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 100.
- Example: `100`.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `notes`

Optional material-specific warnings, provenance notes and unverified custom assumptions. Display these before slicing; never interpret them as executable instructions.

- Type: array. Unit: none. Optional.
- Default behavior: Optional; omitted means no extra material-specific notes, not evidence of calibration..
- Array length: 0 to 24.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..
