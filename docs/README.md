# Forge Slicer documentation

Forge is a local desktop slicer using Go, Wails 3, Svelte, Three.js, and an isolated native PrusaSlicer engine. It imports STL and mesh-only 3MF, searches constrained printing parameters, slices both selected and baseline strategies, validates output, previews real toolpaths, and exports G-code. No application feature uses AI or cloud services.

- [Start here: make Forge fit your printer](configuration/getting-started.md)
- [Material catalogue and recipe sources](configuration/material-catalogue.md)
- [Print readiness and support choices](configuration/supports.md)
- [Architecture](architecture/overview.md)
- [Configuration overview](configuration/overview.md)
- [Printer parameter reference](configuration/printer-schema.md)
- [Filament parameter reference](configuration/filament-schema.md)
- [Optimization preference reference](configuration/optimization-schema.md)
- [Validation](configuration/validation-rules.md)
- [Firmware compatibility](configuration/firmware-compatibility.md)
- [G-code specification](configuration/gcode-specification.md)
- [Examples](configuration/configuration-examples.md)
- [Troubleshooting](configuration/troubleshooting.md)
- [Generating a Printer Configuration Using an External AI Assistant](configuration/external-ai-guide.md)
- [Mathematics](optimization/mathematics.md), [algorithms](optimization/algorithms.md), [assumptions](optimization/assumptions.md), [benchmarks](optimization/benchmarking.md)
- [Build and run](development/build.md), [testing](development/testing.md), [contributing](development/contributing.md)

## First use

1. Launch ForgeSlicer.exe. With no saved machine, Forge opens printer setup. No simulation printer or assumed build plate is selected.
2. Import your verified printer JSON. If you need to create one, open the profile guide, download the offline kit, and fill the blank draft from actual manufacturer/firmware specifications and your calibration. Unknown machine values remain null until resolved; the draft cannot be used to slice.
3. Review your imported machine dashboard and choose a compatible material recipe. PLA, PETG, ABS, ASA, TPU95A and PA12 starters are provided; check your spool settings and edit/import material profiles as needed. Then add an STL/3MF using the small model control. The bed, nozzle and limits now come from your printer. PrusaSlicer is discovered locally or configured under engine settings.
4. Review Print readiness. Choose Auto, Optimized, Easy removal, Stable or Off support mode; inspect contact, bridge and placement advice. Then slice. Auto/Optimized can compare two actual native support shapes, followed by a separate baseline. Every output is statically validated; Cancel stops the subprocess.
5. Inspect layer, toolpath kind and printing order in Preview. The report distinguishes analytical search estimates from estimates calculated from the actual generated G-code.
6. Export G-code. Desktop mode opens a native save dialog; loopback mode uses a browser download. Output is available only after validation succeeds.

## Initial capability boundaries

One model per job, one extruder, rectangular front-left coordinate beds, Marlin 2/Klipper/RepRapFirmware linear G-code. Other kinematics can use these coordinate constraints if verified; no claim is made for belt, nonplanar, proprietary binary, conveyor, resin or multi-extruder machines. 3MF build components and affine transforms are read; external mesh references, embedded process profiles and scripts are not executed. Mesh complexity is capped at 500,000 triangles and 64 MiB per import; preview is capped at 2,000,000 segments. Three completed jobs are retained per session. Unsaved models/results are session-local.

Adaptive layer-height optimization, automatic load-aware 3D reorientation, validated mechanical material models, project persistence, automatic bed-level/pressure-advance macros and signed installers are future work. Readiness bridge classification is a bounded opposing-anchor heuristic. Actual thin-wall handling, bridging, seams, travel, retractions, infill and support geometry are delegated to PrusaSlicer. The interface does not present unimplemented capabilities as complete.
