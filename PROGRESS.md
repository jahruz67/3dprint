# Default flow revision (0.3.1)

Removed mandatory verified-flow certification. Default/preset numeric caps work without a flag or flow evidence; old profiles remain compatible. New drafts include a configurable 4 mm3/s cap. Numeric bounds, flow limiting and G-code validation remain in force. Updated documentation and the user's standalone Kobra Neo profile in Documents. Final checks passed: frontend validation/build, 44 Go checks and all seven native fixtures with unverified flow. The actual Kobra Neo file validates and imports in an isolated browser workspace. Main review version is 0.3.1.

# Smart support revision (0.3)

Implemented bounded geometric print-readiness analysis, sampled risk highlights, support objectives (Auto/Optimized/Easy removal/Stable/Off), organic/snug/grid shapes, placement/threshold controls, and actual native support-candidate comparisons. Auto uses geometry advice to switch support generation on/off. Explicit modes are respected. Contact and adhesion guidance is documented as heuristic; exact regions and bridge exclusion use the native engine. The user's confirmed custom PLA Plus 115°C nozzle recipe is included as unverified, without changing machine temperature constraints. Final verification passed: frontend checks/build, 43 top-level Go checks, seven native integration fixtures, native startup and an isolated browser recommendation-to-slice/report workflow. Builds/tests waited until implementation finished. See VERIFICATION.md. The live workspace remains printer-first with zero printers and seven recipes.

# Printer-first correction (0.2)

The user rejected the model-first onboarding, placeholder machine and single material default. This revision starts with zero printers and puts importing/documenting the user's machine first. Source work: blank unresolved draft, six material starters, persisted active printer, machine-specific material availability, in-app documentation reader and complete downloadable profile-creation kit. The model import becomes secondary. Previous simulation data remains developer-fixture-only. Implementation and final verification are complete: frontend checks/build pass, 31 top-level Go checks pass, all five native integration fixtures pass, the final native startup smoke reports success, and browser testing verified profile-driven adaptation plus the corrected material/nozzle limits. Builds/tests waited until final integration. The live review workspace remains at printer setup with zero imported machines.

---
# Forge Slicer progress

## Current milestone
Working initial release implemented and verified on Windows. Final packaging is the delivery step.

## Completed
- Go/Wails 3 desktop app, loopback review server, embedded offline Svelte/Three interface and documentation.
- Bounded STL/3MF mesh imports, units/build transforms, geometry metrics, interactive plate and move/rotate/scale/center/worker-based auto-orient.
- Versioned JSON schemas, semantic validation/composition, local profile save/import/export/duplicate/editor, provenance and capability boundaries.
- Five deterministic constrained objectives, cached bounded search, Pareto reporting, cancellation and explicit analytical assumptions.
- Isolated native PrusaSlicer integration, settings translation, inert mesh handoff, bounded threads and cancellation/timeouts.
- Genuine selected/baseline G-code, command/state/temperature/flow/feed/bounds/shutdown validation, guarded export and batched layer/feature/travel/order preview.
- G-code-derived estimates, calculated filament mass and independently sliced baseline comparisons.
- Detailed generated parameter references, standalone external-assistant configuration guide, architecture/build/testing/licensing documentation.
- Unit/integration/benchmark suites, representative fixtures, native WebView smoke check, browser import/slice/preview/report/invalid-profile/export verification.

## Verification
All builds/tests waited until final integration per user instruction. Frontend checks/build, Go units, five-objective native integration, Windows launch and benchmarks pass. See VERIFICATION.md and artifacts logs. Two sub-agents completed the frontend and profile/optimizer workstreams.

## Decisions
PrusaSlicer process isolation reuses mature slicing/support algorithms. Source uses AGPL-3.0; app-only packaging excludes native engines. No AI integration, cloud dependency, SQLite or Electron. Machine limits constrain materials/preferences and unknown critical data is rejected. Included hardware values are labeled simulation fixtures.

## Known limits / next work
One extruder and rectangular/front-left beds. Only Windows built/launched here. Adaptive layers, broad firmware macros, leveling/pressure-advance setup, validated mechanics, multi-extruder output, project persistence and signed multi-platform installers remain future work. Native peak-memory and physical-print calibration remain unmeasured. Keep source/schema/native-validator behavior synchronized when expanding capabilities.

## Delivered artifacts
Native executable: bin/ForgeSlicer.exe (20,939,776 bytes). App-only/source package: artifacts/ForgeSlicer-windows-20261101-012910.zip (7,455,341 bytes). Archive inventory verified: executable, complete source, frontend lockfile and 71 dependency notice entries present; downloaded engines/toolchains excluded. Release executable native smoke passed with exit 0. Local review UI remains at http://127.0.0.1:8787.

## Revision0.2 delivered
Updated native executable: bin/ForgeSlicer.exe. Updated source/app archive: artifacts/ForgeSlicer-windows-20261101-015614.zip. Downloadable offline profile creation kit: artifacts/forge-printer-profile-kit.zip. Archive contents checked. Review server16964 runs0.2.0 at http://127.0.0.1:8787 with zero configured printers; isolated UI-test servers have been stopped.
