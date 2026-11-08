# Portable profiles

Forge operates entirely locally. It never calls an AI service, and profiles cannot load code, launch programs or add network services. JSON profiles describe verified machine capabilities, material choices and functional preferences. The optimizer chooses bounded print settings within those capabilities.

Start with [Make Forge fit your printer](getting-started.md). A fresh installation has **no printer configured** and requires importing a verified printer profile before slicing. The six selectable generic material starting points are PLA, PETG, ABS, ASA, TPU 95A and PA12, plus a separately labeled user-requested custom PLA+ 115 C nozzle entry; review the actual spool and [source/assumption record](material-catalogue.md) before using them. [Smart supports and readiness](supports.md) explains geometry-driven support recommendations and shape/removal tradeoffs.

Use three separate files:

1. **Printer**: exact hardware, firmware, usable printing volume and verified limits. Schema: [printer.schema.json](../../schemas/printer.schema.json); [complete parameter reference](printer-schema.md).
2. **Filament**: polymer identity, density, chosen thermal setpoints and conservative flow/layer ranges. Schema: [filament.schema.json](../../schemas/filament.schema.json); [complete reference](filament-schema.md).
3. **Optimization preferences**: objective, workload budget and user functional minima. Schema: [optimization.schema.json](../../schemas/optimization.schema.json); [complete reference](optimization-schema.md).

The application bundles these schemas, examples and documentation. Validation needs no internet access. Schema version is the string `"1.0"`, not a number. Unknown properties and unsupported versions are rejected. The schema includes field descriptions, units, constraints, dependencies, safety notes and default behavior. Regenerate its references with `powershell -NoProfile -ExecutionPolicy Bypass -File schemas/generate.ps1` after editing the source generator.

## Composition and precedence

Hardware constraints always win. A filament profile can lower throughput through `min(printer.extruders[0].maxVolumetricFlow, filament.maxVolumetricFlow)` but cannot raise the printer's limit. Its temperatures must already fit the machine's range. Functional preferences constrain the optimizer (`walls >= minWalls`, `infill >= minInfill`) and never override hardware limits. The backend validates both individual profiles and their composition before optimization. A mode changes the objective function rather than replacing a printer profile with a static preset.

All dimensions are millimeters; area mm²; volume mm³; linear speeds mm/s; G-code feed rates mm/min; acceleration mm/s²; flow mm³/s; temperatures °C; density g/cm³; infill is a fraction in `[0,1]`. A value of `20` is **not** 20% infill: use `0.20`. Identifiers use lowercase ASCII letters, digits, dots, underscores and hyphens, at most 64 characters.

## Evidence and unknown values

Printer groups require nonempty `provenance` entries for `buildVolume`, `motion` and `thermal`. Record the source used for those hardware limits. Flow caps can come from factory defaults or conservative presets; a flow source or calibration record is optional. `flowVerified` is optional legacy metadata: false, true and omission all permit import and slicing. New drafts supply a 4 mm3/s operating cap, which can be adjusted without a verification step. The numeric cap is still enforced, together with the material cap. Optional hardware capability flags remain false when unavailable.

The `reference-cartesian.json` and `reference-pla.json` files are explicitly fictional **developer test fixtures**. They are never loaded by `config.Defaults()`, never preinstalled as a selectable printer and never used as a fallback for missing hardware information. `config.ReferenceFixture()` exists only for deterministic automated tests. The user-facing blank template instead keeps unknown hardware values null and cannot pass validation until completed.

## Capability boundary

Initial adapter scope: single extruder, rectangular bed, front-left Cartesian coordinates, Marlin 2 / Klipper / RepRapFirmware G-code families, literal reviewed initialization, ordinary explicit extrusion retraction. Firmware names are not promises that every firmware configuration is compatible. Delta, belt, circular beds, multi-tool/multi-material, proprietary containers, custom firmware macros and firmware retraction are rejected or unavailable. Clearances must already be subtracted from usable build-volume dimensions; excluded rectangular regions model permanent obstacles. Pressure advance and automatic leveling flags record verified availability but do not invent initialization commands.

## Creating and maintaining profiles

Start from the blank printer template, complete every unknown field from evidence, validate and then import. The template is a JSON draft, not an importable verified profile. Use the profile editor to inspect errors before saving. To duplicate an existing verified profile, export its JSON, assign a new identifier and import it after reviewing changed capabilities. Export creates a portable JSON file; no account is required. The exact local storage path is owned by the application service and appears in development documentation.

Migration is explicit. Version 1.0 has no predecessor; `Migrate` returns a copy of a valid current profile and rejects unknown formats. Never silently rename fields from a different slicer's configuration and assume equivalence.
