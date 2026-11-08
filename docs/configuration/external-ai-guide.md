# Generating a Printer Configuration Using an External AI Assistant

This document is a standalone workflow for asking an **external** assistant to research a printer profile. Forge has no AI model, assistant, API key, inference, cloud connection or AI-dependent behavior. The user independently supplies documentation to an external service and manually imports its output. Humans can follow exactly the same steps without AI.

## Supply a self-contained documentation bundle

Give the assistant this document, the three JSON files in `schemas/`, every file in `docs/configuration/`, and the blank printer draft exported by the profile editor. Material examples live in `examples/profiles/`; reference-machine files there are developer test fixtures and must not be treated as hardware specifications. Exact authoritative schemas are `printer.schema.json`, `filament.schema.json`, `optimization.schema.json`, Draft 2020-12; their version is the string `"1.0"`. The generated parameter references enumerate each field, type, units, required status, numerical bounds and dependency rules. No online schema resolution is required by Forge. Fresh Forge installations contain no printer profile.

Initial compatibility is intentionally bounded: single extruder, rectangular usable bed, front-left origin, verified Marlin 2 (`marlin2`), Klipper (`klipper`) or RepRapFirmware (`reprapfirmware`) command flavor. Multi-tool printers, circular/delta/belt arrangements, arbitrary firmware macros, firmware retraction and proprietary job containers are outside initial scope. Do not generate a profile that declares an unsupported architecture as rectangular to make validation pass.

## Identify the exact physical setup

Record manufacturer, exact model and revision, installed firmware and version, installed nozzle size, extruder architecture, filament diameter, thermal hardware and modifications. A model name alone does not identify vendor firmware features, upgrades or calibrations. Ask the human for missing installation details before generating final output.

Research manufacturer manuals and firmware configuration for the exact model. Keep sources for motion, usable build volume and thermal limits. For flow, use a suitable factory default or conservative preset and describe it as a starting value; physical calibration is optional. New Forge drafts use a 4 mm3/s operating cap.

Keep published hardware ratings separate from preset choices. Do not fabricate maximum temperatures, firmware command support or citations. A configured numeric flow cap does not require measured verification or a calibration certificate. Omit the legacy `flowVerified` field; old profiles containing false or true remain accepted. Flow provenance is optional. Unresolved required hardware fields still need numeric values and appropriate machine documentation; unknown optional hardware capabilities remain false.

## Produce portable JSON

Start from Forge's blank printer draft (`GET /api/printer-template`), whose unresolved hardware values are null. Replace each unknown with verified values; an unresolved draft must not pass validation. Do not copy capability values from the fictional developer fixture. Use a new lowercase identifier and name. Preserve object structure, field names and types; never add arbitrary engine flags, script hooks, external schema URLs or environment expansion. All required fields must be present, including empty arrays and false optional unverified capability booleans. `additionalProperties:false` rejects extras.

Use millimeters, mm/s, mm/s², mm³/s, °C and g/cm³. Infill is `0.20` for 20%, not `20`. Temperatures belong to the material layer and must fit printer thermal limits. Printer limits cannot be overridden by a filament or mode. Excluded rectangles and usable build dimensions must account for actual permanent obstacles and toolhead clearances.

Pair firmware family/flavor correctly. Supply a verified supported-command list, a non-overlapping unsupported list and literal startup/shutdown arrays. No macros, placeholders, template braces, scripts or shell commands are supported. Optional leveling/pressure advance flags do not configure those features. If the printer requires a custom macro whose effect cannot be established, explain that limitation and do not assert compatibility.

## Developer-only complete structural example

This structural example is an explicitly fictional **automated-test fixture**, retained to demonstrate a complete schema instance when this document is copied independently. It is never loaded into the application catalogue. It is not the user-facing template and must never be relabeled as verified physical hardware. User profile generation starts with the null-valued draft and researched evidence:

```json
{
 "schemaVersion":"1.0","id":"reference-cartesian","name":"Reference Cartesian 220 (simulation)",
 "manufacturer":"Forge software reference fixture","model":"Virtual Cartesian 220 — replace before physical printing",
 "firmware":{"family":"marlin","version":"Reference Marlin 2 command subset","gcodeFlavor":"marlin2","extrusionMode":"absolute",
  "supportedCommands":["G0","G1","G4","G21","G28","G90","G91","G92","M73","M82","M83","M84","M104","M106","M107","M109","M140","M190","M201","M203","M204","M205","M220","M221","M400"],
  "unsupportedCommands":["G2","G3","G10","G11"],"startGcode":[],"endGcode":[]},
 "buildVolume":{"width":220,"depth":220,"height":250,"shape":"rectangular","origin":"front-left","excludedRegions":[]},
 "motion":{"maxSpeed":120,"maxTravelSpeed":180,"maxAcceleration":1000},
 "extruders":[{"nozzleDiameter":0.4,"filamentDiameter":1.75,"maxVolumetricFlow":8,"retractionLength":0.8,"retractionSpeed":30,"type":"direct"}],
 "thermal":{"minNozzle":170,"maxNozzle":260,"minBed":0,"maxBed":100,"hasHeatedBed":true,"partCooling":true},
 "capabilities":{"autoBedLeveling":false,"firmwareRetraction":false,"pressureAdvance":false},
 "provenance":{"buildVolume":"SIMULATION ONLY: reference fixture dimensions","motion":"SIMULATION ONLY: bounded software fixture","thermal":"SIMULATION ONLY: no hardware rating","maxVolumetricFlow":"SIMULATION ONLY: reference assumption; calibrate real hardware"}
}
```

## Validate and import

Validate locally with `forge --validate-profile printer.json --kind printer` (use `filament` or `optimization` for other profile kinds). This performs both schema and semantic checks. You can also validate against the bundled formal JSON Schema using a trusted offline Draft 2020-12 validator. Then use Forge's profile import/editor, which additionally performs firmware compatibility, thermal cross-field, command, provenance and flow-verification checks. The same application check is available at `POST /api/profiles/validate` with `{ "kind":"printer", "profile": { ... } }` on the local application origin. A valid schema alone cannot establish compatibility or physical safety.

Read every validation error. Correct the source data or report unsupported capabilities; never remove constraints or increase limits just to silence an error. Save the validated JSON, import it through the printer profile editor, select compatible filament and slice a small representative model. Review startup/shutdown, placement, generated layers, speeds and temperatures before physical use. Static validation does not replace supervised calibration and physical acceptance prints.

## Copyable instructions for an external assistant

> Using the attached Forge 1.0 schemas, complete examples and validation documentation, research a portable profile for my exact printer: [manufacturer, model, revision, modifications, firmware/version, installed nozzle]. Use primary manufacturer/firmware sources and my supplied calibration records. Cite each critical capability in provenance. Separate documented ratings, calibrated limits and unknowns. Never invent commands, maximum flow, temperatures, offsets, motion capabilities or optional hardware. Ask for missing installation details. If any critical value is unknown, return an unresolved-items list and an explicitly incomplete draft rather than a supposedly importable final profile. Respect initial single-extruder/rectangular/front-left restrictions. Return only the final JSON once all critical values are verified, followed separately by sources and human review instructions. Do not include API keys, shell commands, macros, templates or AI integrations. Validate against the exact schema and all documented semantic rules; never weaken those rules to fit unsupported hardware.
