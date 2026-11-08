# printer schema parameter reference

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

Human-readable machine name.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `manufacturer`

Verified manufacturer identity.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `model`

Exact installed model/revision, not a guessed family.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware`

Firmware identity and explicitly supported G-code contract.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.family`

Verified firmware family.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: marlin, klipper, reprap.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.version`

Exact installed version or documented verified compatibility range.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.gcodeFlavor`

PrusaSlicer adapter flavor. Must agree with family.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: marlin2, klipper, reprapfirmware.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.extrusionMode`

Extrusion coordinate interpretation. Initializer sets matching M82 or M83.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: absolute, relative.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.supportedCommands`

Explicitly verified firmware commands. Command names only; never executable shell commands.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 0 to 256.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.unsupportedCommands`

Explicitly verified firmware commands. Command names only; never executable shell commands.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 0 to 256.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.startGcode`

Literal G-code lines. No template expansion, macros, scripts or shell execution. Commands must be supported and output must pass full G-code validation.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 0 to 128.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `firmware.endGcode`

Literal G-code lines. No template expansion, macros, scripts or shell execution. Commands must be supported and output must pass full G-code validation.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 0 to 128.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume`

Usable Cartesian printing volume after permanent mechanical clearances. Initial adapter supports front-left rectangular beds.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.width`

Usable X extent after toolhead clearances.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 10 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.depth`

Usable Y extent after toolhead clearances.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 10 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.height`

Usable Z extent.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 10 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.shape`

Supported bed geometry.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: rectangular.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.origin`

Coordinate origin used by model placement and G-code.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: front-left.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.excludedRegions`

Rectangular regions unavailable for printing or motion.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 0 to 64.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.excludedRegions[].x`

Lower-left X coordinate.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.excludedRegions[].y`

Lower-left Y coordinate.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.excludedRegions[].width`

Excluded rectangle width.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.01 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `buildVolume.excludedRegions[].depth`

Excluded rectangle depth.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.01 to 2000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `motion`

Non-overridable machine motion limits. Profile speeds are capabilities, not print presets.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `motion.maxSpeed`

Maximum permitted extrusion motion speed.

- Type: number. Unit: mm/s. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 1000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `motion.maxTravelSpeed`

Maximum permitted non-extruding motion speed.

- Type: number. Unit: mm/s. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 1500.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `motion.maxAcceleration`

Maximum permitted commanded acceleration.

- Type: number. Unit: mm/s2. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 50000.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders`

Exactly one extruder supported in initial adapter.

- Type: array. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Array length: 1 to 1.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].nozzleDiameter`

Installed nozzle bore diameter, measured or manufacturer specified.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.1 to 2.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].filamentDiameter`

Nominal input filament diameter.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 4.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].maxVolumetricFlow`

Configured hot-end flow cap. Factory defaults and conservative preset values are accepted without calibration or verification; material profiles may reduce this cap. New drafts start at 4 mm3/s.

- Type: number. Unit: mm3/s. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0.1 to 100.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].retractionLength`

Explicit extrusion retraction distance.

- Type: number. Unit: mm. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 15.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].retractionSpeed`

Retraction feed rate, never above machine speed cap.

- Type: number. Unit: mm/s. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 150.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].flowVerified`

Legacy informational metadata. Omitted, false and true are all accepted; this field never blocks import or slicing.

- Type: boolean. Unit: none. Optional.
- Default behavior: Optional compatibility field; ignored by validation and slicing policy..
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `extruders[].type`

Extruder feed architecture, used for documentation and retraction interpretation.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Accepted values: direct, bowden.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal`

Verified thermal hardware range and cooling availability.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.minNozzle`

Minimum supported printing temperature.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 500.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.maxNozzle`

Maximum permitted nozzle temperature.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 1 to 500.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.minBed`

Minimum permitted bed temperature; zero for unheated.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 200.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.maxBed`

Maximum permitted bed temperature; zero for unheated.

- Type: number. Unit: C. Required.
- Default behavior: Required; no implicit default.
- Inclusive bounds: 0 to 200.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.hasHeatedBed`

Verified heated bed; false requires zero minBed and maxBed.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `thermal.partCooling`

Verified controllable part cooling fan.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `capabilities`

Verified optional hardware. Unknown capabilities must be false. Flags do not emit unreviewed macros.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `capabilities.autoBedLeveling`

Verified automatic bed leveling hardware. This does not automatically run a leveling command.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `capabilities.firmwareRetraction`

Must be false in initial adapter; explicit E retraction is used.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Required constant: False.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `capabilities.pressureAdvance`

Records verified support only. No pressure advance commands are generated.

- Type: boolean. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `provenance`

Documentation for machine capability groups. A flow source/calibration record is optional; default flow caps do not require certification.

- Type: object. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `provenance.buildVolume`

Evidence for usable dimensions.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `provenance.motion`

Evidence for motion limits.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `provenance.thermal`

Evidence for thermal limits.

- Type: string. Unit: none. Required.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..

## `provenance.maxVolumetricFlow`

Optional note describing a default flow cap or calibration.

- Type: string. Unit: none. Optional.
- Default behavior: Required; no implicit default.
- Example: see the complete examples/profiles files; retain the field type and units.
- Hardware / firmware dependencies and interactions: See validation-rules.md and firmware-compatibility.md.
- Common mistakes: Do not infer a machine capability from a similarly named model..
- Safety: Verify values against the exact hardware and firmware before printing..
