# Forge Slicer

A local optimization-first FDM desktop slicer built with Go, Wails 3, Svelte and Three.js. Uses native PrusaSlicer for genuine toolpaths. No AI, cloud service, account, or API key.

**Import your printer profile → review machine capabilities and compatible materials → add STL/3MF → review print readiness and supports → choose an objective → slice → inspect layers → export G-code.**

Forge starts with **no printer configured**. The main setup screen provides printer-profile import, a detailed in-app [profile creation guide](docs/configuration/getting-started.md), and a downloadable offline documentation/schema kit. Importing a profile defines the build plate, nozzle, firmware, motion/thermal/flow limits and material compatibility. Your imported profiles and active printer selection are saved locally.

Seven editable recipes are included: PLA, PETG, ABS, ASA, TPU 95A, PA12 and the user-requested custom PLA Plus at 115°C nozzle. The latter is explicitly unverified and remains unavailable when the printer's minimum nozzle temperature excludes it. Recipes must be reviewed against the actual spool and machine. See [material sources and assumptions](docs/configuration/material-catalogue.md).

[Print readiness and supports](docs/configuration/supports.md) inspect the transformed geometry for suspended surfaces, possible bridges, bed contact and small details. Choose Auto, Optimized, Easy removal, Stable or Off, with organic/snug/grid shapes and placement controls. Auto and Optimized compare two native support candidates when applicable, using validated toolpath material and time estimates. Advice is geometric and does not establish physical print success.

Five deterministic constrained objectives: Fast Draft, Balanced, High Quality, Structural Strength and Material Saver. Selected and baseline strategies are independently sliced; reports distinguish analytical search estimates from actual G-code-derived estimates and calculated extrusion mass. Mechanical strength remains an unvalidated proxy.

See [documentation](docs/README.md), [build instructions](docs/development/build.md), [configuration guide](docs/configuration/overview.md), [external assistant configuration guide](docs/configuration/external-ai-guide.md), [progress](PROGRESS.md) and [verification evidence](VERIFICATION.md).

No simulation printer is installed or selected. Fictional printer files under examples are developer test fixtures only. To create your own profile, start from the blank draft provided by the setup screen, fill the hardware fields from your machine specifications, use factory/default flow settings, validate, then import. Flow calibration and verification flags are not required. Initial adapter boundaries: one extruder, rectangular/front-left beds, Marlin 2/Klipper/RepRapFirmware linear G-code. Proprietary or unusual machines are not automatically supported.

AGPL-3.0; see [LICENSE](LICENSE) and [third-party notices](THIRD_PARTY_NOTICES.md). A subprocess boundary does not remove engine licensing obligations.
