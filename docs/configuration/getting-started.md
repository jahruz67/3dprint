# Make Forge fit your printer

Forge starts with **no printer configured**. It does not select a fictional machine, invent a bed size or infer hardware limits from a brand name. Import a verified printer profile, review a material starting point, then slice your model. Everything below is available offline in the application; only independent research may require internet access.

## 1. Identify the machine you actually have

Record the manufacturer, exact model and revision, firmware family and installed version, installed nozzle diameter, filament diameter, extruder type and modifications. Check the machine's own firmware/configuration screens or configuration files. A similarly named machine, different firmware build or changed hot end can have different limits. Keep the manufacturer's manual and firmware configuration alongside your exported profile.

The initial adapter supports a **single extruder, rectangular usable bed and front-left coordinate origin** with verified Marlin 2, Klipper or RepRapFirmware commands. Multi-tool, belt, circular/delta arrangements, proprietary job containers and arbitrary firmware macros are outside this release. Read [firmware compatibility](firmware-compatibility.md) before investing time in a specialized setup. If your printer falls outside the boundary, do not declare false capabilities just to make a file import.

## 2. Start a draft, without inventing unknowns

Open the printer profile editor and choose its blank template. The backend supplies the same draft at `GET /api/printer-template`. It contains all exact field names and nested objects; unknown hardware values are JSON `null`, identity/evidence fields are empty, optional unavailable capabilities are false. A numeric 4 mm3/s flow cap is already filled in; no verification flag is required. This is valid JSON but **intentionally not a valid printer profile**. Replace unknowns with verified values before validation can pass.

Use the [printer field reference](printer-schema.md) and [formal printer schema](../../schemas/printer.schema.json) while filling it. Keep `schemaVersion` as the string `"1.0"`. Give the profile a unique lowercase `id` such as `my-printer-main` and a useful display name identifying its nozzle/firmware revision. Do not paste a different slicer's configuration file into this format: field names and meaning differ.

## 3. Collect evidence for the required hardware fields

| Profile group | What to establish | Where and how to verify |
|---|---|---|
| `manufacturer`, `model`, `firmware` | Exact identity, installed firmware/version, matching flavor, extrusion mode and supported commands. | Manufacturer manual plus actual firmware configuration. Family/flavor pairs are `marlin`/`marlin2`, `klipper`/`klipper`, `reprap`/`reprapfirmware`. Choose absolute or relative E deliberately; the adapter establishes it. |
| `buildVolume` | Usable width/depth/height, rectangular bed, front-left origin, excluded rectangles. | Documented coordinates and mechanical checks. Subtract permanent toolhead/fixture clearances from usable dimensions. Bed clips/obstacles belong in excluded regions; the validator conservatively prevents paths across them. |
| `motion` | Conservative maximum print speed, travel speed and acceleration. | Exact machine/firmware limits, including modifications. A marketing speed is not automatically a verified process speed. |
| `extruders[0]` | Installed nozzle and filament diameters, direct/Bowden architecture, configured volumetric flow cap, retraction length/speed. | Use a factory/default flow cap or the provided 4 mm3/s starting cap. Calibration is optional. Nozzle/filament dimensions come from the installed hardware. |
| `thermal` | Permitted nozzle/bed temperature ranges, whether the bed is heated, controllable part cooling. | Exact thermal hardware ratings and firmware limits. An unheated bed requires both bed limits zero. Do not raise ratings merely to accommodate a material. |
| `capabilities` | Verified leveling, pressure advance and other listed hardware flags. | Exact firmware/hardware documentation. Unknown optional flags stay false. Firmware retraction must be false in this adapter. These flags do not invent startup macros. |
| `provenance` | Sources for `buildVolume`, `motion`, `thermal`; an optional note for `maxVolumetricFlow`. | Record title/revision/page or URL and access date; for calibration, record date, method, nozzle, material and temperature. Unknown critical values must be resolved before slicing. |

Units are mm, mm/s, mm/s2, mm3/s and C. The profile's limits describe capability; the optimizer determines actual settings within them. Never enter a G-code feed rate in mm/min directly into a mm/s field. For example, a documented feed rate must be converted deliberately; the application later multiplies a profile speed by 60 when writing G-code.

The verified command list must include the adapter's required commands described in [validation rules](validation-rules.md). List only commands known to work on the actual firmware. `startGcode` and `endGcode` are arrays of literal single lines; empty arrays are acceptable when the generated initializer/shutdown meets the machine's needs. No shell commands, scripts, engine placeholders, template expressions or custom firmware macros are executed. Machines requiring an uninspectable macro are outside the initial export dialect. Review [G-code boundaries](gcode-specification.md).

## 4. Validate, repair errors and save the printer

Choose Validate in the profile editor. Validation checks the formal schema, cross-field relationships, critical evidence, command compatibility and supported hardware boundary. Read the exact path named in an error and fix its source data. A number must be a JSON number, not a quoted string. `null` means unresolved and will be rejected. Unknown properties and duplicate keys are errors. Factory/default numeric flow caps are accepted without calibration. The optional legacy `flowVerified` field does not affect validation or slicing.

After validation passes, save/import the profile. It appears in the printer selector and persists locally. Export a copy for backup. To create a variant, duplicate/export it, change the identifier, document what changed, verify the affected limits and import the new file. A nozzle change, firmware update or new hot end needs a reviewed variant, not just a renamed profile.

You can also validate offline from a terminal:

```text
forge --validate-profile printer.json --kind printer
```

Use `filament` or `optimization` as the kind for those files. The same application validator is available at `POST /api/profiles/validate` with `{ "kind":"printer", "profile": { ... } }`; requests must follow the local application's origin policy. A JSON Schema-only tool cannot replace semantic validation. [Troubleshooting](troubleshooting.md) maps common errors to repairs.

## 5. Choose and review a material

Select PLA, PETG, ABS, ASA, TPU 95A or PA12 from the material selector. These six generic entries are **starting recipes**, not proof that your printer or spool supports them. Review actual spool nozzle/bed temperatures, density, cooling and material-specific throughput; duplicate the closest recipe and save a spool-specific profile when needed. Read [material catalogue and primary sources](material-catalogue.md) for exactly which numbers came from manufacturer documents and which are uncalibrated engineering assumptions.

The additional custom PLA+ entry stores the explicitly requested **115 C nozzle** temperature and is labeled unverified. This is substantially below the cited manufacturer's PLA+ reference. Read its material notes; the recipe does not override machine thermal minimums or firmware safeguards and may correctly be rejected by your printer profile.

Material temperatures must fit the imported printer's ranges. Positive cooling requires a part-cooling fan. Effective volumetric flow is `min(printer limit, material limit)`; a material can reduce a machine limit but cannot raise it. Layer ranges must intersect the installed nozzle's supported interval. The schema cannot verify drying state, enclosure requirements, fumes, adhesion treatment or actual spool formulation; check the manufacturer's instructions for those conditions. For all material fields see the [filament reference](filament-schema.md).

## 6. Import, place and slice a model

Import a real STL, or a supported 3MF geometry file. Confirm its millimeter dimensions, place it on the build plate, then move, rotate, scale, center or auto-orient it deliberately. Check that the transformed model fits the usable volume and avoids excluded regions. Inspect warnings about malformed/non-closed geometry instead of ignoring them.

Choose the objective that matches the job: Fast Draft, Balanced, High Quality, Structural Strength or Material Saver. The optimizer calculates bounded candidate settings; these names are objective functions, not five disguised printer presets. Set functional minimum walls/infill, support permission and search budget in advanced settings as appropriate. Structural strength uses an explicitly unvalidated proxy; known load direction helps state assumptions but does not establish a safe load.

Review the placed-model readiness report and unsupported highlights. [Smart supports](supports.md) offers Auto, Optimized, Easy removal, Stable and Off goals, plus explicit organic/snug/grid shapes, placement restrictions and a threshold measured from horizontal. Auto can leave a normal bed-supported cube without supports, while unsupported overhangs trigger native support processing. Contact, orientation, brim and detail findings explain what to inspect; they do not silently change your model. Organic/snug comparisons use actual sliced support consumption and print-time estimates, not a preset claim that one shape always wins.

Slice. Expensive work runs asynchronously and can be cancelled. A genuine native PrusaSlicer executable must be configured; missing engine errors identify that requirement. The app passes its own geometry/settings to the isolated process and does not run imported project scripts.

## 7. Review output and export

Inspect the generated layer/toolpath preview, perimeters, infill, supports and travel. Read the optimizer's selected settings, constraints and assumptions, and keep analytical predictions separate from actual native G-code estimates. Review calculated time/filament use and the baseline comparison; neither is a guarantee of physical quality or strength.

Export becomes available only after output validation passes. Check first-layer placement, temperatures, startup/shutdown and the printer profile against the physical machine. Static validation cannot establish homing accuracy, wiring, adhesion, material conditioning or actual firmware behavior. Begin with a supervised small calibration print and record results before relying on a new profile.

## Optional external-assistant workflow

You can prepare the profile entirely by hand. If you independently use an external assistant to help research specifications, give it [Generating a Printer Configuration Using an External AI Assistant](external-ai-guide.md), the schema and exact hardware details. Require primary evidence and unresolved-items reporting. Forge itself contains no AI integration, subscriptions or cloud inference, and its import/validation workflow remains fully local.
