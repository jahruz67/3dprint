# Firmware compatibility

The initial G-code adapter uses PrusaSlicer's established `marlin2`, `klipper` and `reprapfirmware` flavors. The profile records the exact firmware family and verified version/range; local configuration and vendor customizations still need review. Commands must be explicitly listed as supported. The application never treats an absent unsupported-command entry as proof of support.

| Family | `gcodeFlavor` | Extrusion | Notes |
|---|---|---|---|
| `marlin` | `marlin2` | `absolute` / `relative` | Verify vendor-enabled motion and thermal commands. |
| `klipper` | `klipper` | `absolute` / `relative` | Custom PRINT_START/PRINT_END and other macros are outside the literal command contract. |
| `reprap` | `reprapfirmware` | `absolute` / `relative` | Verify configured drive and tool behavior; single tool only. |

Base output establishes millimeter units, absolute XYZ and a declared E mode. Absolute extrusion uses `M82`; relative extrusion uses `M83`. `G92 E0` changes the logical extrusion origin without physical extrusion. `G28` is machine homing, not an ordinary line move. Commanded speeds are in mm/min in G-code; profiles use mm/s, requiring multiplication by 60 in translation. Firmware acceleration and velocity commands can be firmware-specific; generated commands must be supported and never exceed profile limits.

The adapter generates thermal and motion initialization/shutdown and inserts only validated literal profile lines. Source snippets cannot use placeholders such as `{temperature}`, `[first_layer_temperature]` or `$VAR`. Optional flags for leveling/pressure advance do not automatically generate firmware commands. If a machine requires a macro that the validator cannot establish as safe, the initial release does not support that startup path. Do not replace a macro with an invented equivalent.

Organic, snug and grid support structures, ordinary native infill, perimeters, seam/travel planning and explicit retractions are handled by PrusaSlicer. Auto/Optimized can compare actual organic/snug outputs; see [smart supports](supports.md). Adaptive layer schedules, multi-tool changes, firmware retraction and proprietary printer packaging remain outside the enabled feature set. A real printer profile is a compatibility declaration for a particular firmware setup, not a universal machine certification.

Firmware command references for independent verification: [Marlin G-code](https://marlinfw.org/meta/gcode/), [Klipper G-codes](https://www.klipper3d.org/G-Codes.html), [RepRapFirmware G-code dictionary](https://docs.duet3d.com/en/User_manual/Reference/Gcodes). Save the relevant pages or vendor manual revision alongside the profile for offline evidence. These links are research starting points; no online access occurs in the application.
