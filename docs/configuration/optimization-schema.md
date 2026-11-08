# optimization schema parameter reference

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

## `mode`

Objective function, not a preset. See optimization/mathematics.md.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: fast, balanced, quality, strength, material.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `budget`

Maximum unique feasible candidate evaluations; deterministic search terminates when exhausted.

- Type: integer. Unit: candidates. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 16 to 4096.
- Example: `256`.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `minWalls`

User functional minimum perimeter count; optimizer may increase it.

- Type: integer. Unit: count. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 12.
- Example: `2`.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `minInfill`

User functional minimum interior fill fraction.

- Type: number. Unit: fraction. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 1.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `supports`

Permit automatic support generation when geometry analysis finds overhangs. False with substantial overhangs produces a warning.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `loadDirection`

Optional nominal unitless XYZ load vector. Zero vector means unspecified; directional strength remains an unvalidated engineering proxy.

- Type: array. Unit: none. Optional.
- Default behavior: Optional; omitted or zero vector means unknown load direction..
- Array length: 3 to 3.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `supportMode`

Support objective: auto uses geometry readiness; optimized compares actual native alternatives; easy_remove favors contact removal; stable favors support robustness; off disables supports explicitly.

- Type: string. Unit: none. Optional.
- Default behavior: Optional; when absent legacy supports true means auto and false means off. An explicit mode takes precedence..
- Accepted values: auto, optimized, easy_remove, stable, off.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `supportStyle`

Native support structure. Auto permits supported native style comparison; organic, snug and grid select explicit native styles.

- Type: string. Unit: none. Optional.
- Default behavior: Optional; omitted means auto..
- Accepted values: auto, organic, snug, grid.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `supportPlacement`

Everywhere can anchor on lower model geometry; buildplate restricts supports to bed-connected placement and can leave internal overhangs unsupported.

- Type: string. Unit: none. Optional.
- Default behavior: Optional; omitted means everywhere..
- Accepted values: everywhere, buildplate.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `supportAngle`

Overhang slope threshold from the horizontal plane. Higher thresholds include steeper faces and can request more support. This is not an angle measured from vertical.

- Type: number. Unit: degrees. Optional.
- Default behavior: Optional; omitted means 45 degrees. Internal Go zero value is normalized to 45; explicit JSON values must be 20 through 80..
- Inclusive bounds: 20 to 80.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..
