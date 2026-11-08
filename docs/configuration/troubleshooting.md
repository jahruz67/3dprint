# Configuration troubleshooting

| Error or symptom | Meaning and action |
|---|---|
| Unknown property / wrong type | Check the exact camelCase field name against the schema reference. Numbers are JSON numbers, booleans are `true`/`false`. |
| Unknown schema version | Only string `"1.0"` is implemented; no silent migration from unrelated slicer formats exists. |
| Duplicate JSON key | Remove the duplicate; different tools disagree which value wins. |
| Old verified-flow error | Update Forge. Flow verification and calibration evidence are no longer required; old false/true flags and omitted flags are accepted. Use a numeric flow cap, such as the 4 mm3/s draft default. |
| Firmware/flavor disagreement | Use marlin/marlin2, klipper/klipper or reprap/reprapfirmware as verified on the actual machine. |
| Unsupported or unreviewable snippet | Remove templates/macros and use only verified literal commands; machines requiring uninspectable macros are outside initial scope. |
| Material temperature outside range | Select a compatible material or correct the machine profile from evidence; never increase a hardware rating to suppress the error. |
| No feasible layer interval | Material min/max must intersect 0.2–0.8 of installed nozzle diameter. |
| Model exceeds volume | Change placement, scale or orientation deliberately; verify units and mechanical usable region. |
| Native engine missing | Configure a genuine PrusaSlicer console executable; this is a file path, not a command string. See development/build.md. |
| G-code export unavailable | Read the job error and output-validation findings; a sliced preview does not automatically make a file valid. |
| Strength result seems unusual | Strength is an engineering proxy; supply load direction if known and validate with representative physical tests. |

The app does not calibrate a printer by reading its model name. Compare actual installed nozzle, firmware revision, bed clips, offsets, thermal hardware and motion configuration with the profile. Keep original evidence with the exported JSON for future review. Review the initial capability boundary before attempting a specialized or proprietary printer.
