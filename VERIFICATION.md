# Verification — default flow revision 0.3.1

Observed 2026-11-01. Implementation and documentation finished before checks/builds began.

- Removed the verified-flow gate. Legacy `flowVerified` metadata is optional; false, true and omission all permit validation, import and slicing. A flow provenance/calibration record is optional. New profile drafts include a configurable 4 mm3/s cap and no verification flag.
- Numeric flow bounds, optimizer speed/flow constraints, material composition and G-code flow validation remain enforced. Regression coverage exercises all three metadata forms, omitted flow evidence, rejected invalid numeric caps and capped optimization without verification.
- Frontend check passed with zero errors/warnings; production bundle and native Windows executable built. Go checks: 44 top-level tests/examples passed (native test skipped there and invoked separately). Log: artifacts/default-flow-unit-tests.txt.
- All seven real native fixtures passed with unverified 4 mm3/s printer flow caps and no flow provenance. Log: artifacts/default-flow-native-integration.txt.
- The actual Documents/Anycubic Kobra Neo.json returned valid:true with no errors from the rebuilt application's validator and imported successfully through the browser UI into an isolated test workspace. It has no flow verification flag. The active-machine dashboard showed its configured 4 mm3/s cap and compatible material recipes.
- The same Kobra Neo browser workflow sliced a 24-triangle overhang fixture to validated output, 86 preview layers and enabled G-code export. The inspected browser warning/error console was empty.
- The live review application runs version 0.3.1. Main workspace profiles were not replaced by test data. No physical printing or printer calibration was performed.

---

# Previous verification — smart support revision 0.3.0

Observed on 2026-11-01. Two sub-agents worked on this revision. All implementation and documentation finished before builds/tests began.

- Frontend check: zero errors and warnings. Production build: passed, 123 modules, existing Three.js bundle-size advisory. Native Windows build: passed, 21,280,256 bytes.
- Go: 43 top-level unit/example checks passed; native integration skipped in this invocation and run separately. Log: artifacts/supports-unit-tests.txt.
- Native PrusaSlicer: all seven fixtures passed (cube, thin wall, boundary, bridge, optimized overhang, easy-removal organic and stable grid). Auto left the cube unsupported; the overhang compared two validated shapes. Log: artifacts/supports-native-integration.txt. An initial test run reached the six-model session limit on fixture seven; fixtures now use isolated services. The app's model limit was retained.
- Geometry tests cover reversed winding, coincident touching solids, cantilevers, bridge anchors, contact/slenderness/detail advice, angle convention, sample bounds, cancellation and invalid input. Profile tests cover strict support fields, legacy profiles and custom 115°C retaining machine constraints.
- Native WebView startup logged successful embedded frontend/API connection. Log: artifacts/supports-native-smoke.err.log. A WebView window-class cleanup warning followed the successful smoke result.
- Isolated browser workflow: imported a 24-triangle cantilever; readiness estimated 283 mm² unsupported area, 90 mm² bed contact and 20 mm sampled gap. Off showed a warning; the recommendation action restored Auto. Risk highlighting and the full in-app support guide were available. All seven recipes appeared, with custom PLA+115 notes and the fixture's 170°C minimum incompatibility explained.
- That browser workflow reached genuine validated output and 118 preview layers. Both organic and snug passed. The report selected snug at approximately 16 minutes, 4.2 g total mass and 0.27 m support filament, versus organic at 19 minutes, 4.4 g and 0.33 m. Inspected browser warning/error console was empty. These are toolpath estimates/calculated extrusion, not measured print results.
- Live review state: http://127.0.0.1:8787, version 0.3.0, zero printers, empty active printer ID, seven recipes. Test profiles stayed in the separate port/data directory.

Support readiness is sampled geometry; native PrusaSlicer determines exact regions, branches and recognized bridge omission. Candidate selection holds primary print settings fixed, with a documented 75% support-filament / 25% print-time score. Removal force, stability and print success remain unmeasured. The confirmed custom 115°C PLA Plus nozzle preset is unverified and cannot override machine limits. No physical printing was performed.

---

# Previous verification — printer-first revision 0.2.0

Observed on 2026-11-01. Builds and tests waited until the redesign was implemented, preserving the user's instruction.

## Product behavior verified

- Fresh state: zero configured printers, empty active printer ID, six editable material recipes. No placeholder machine or assumed build plate is shown.
- The live application opens with printer import, an in-app creation guide, downloadable documentation/schema kit, and an unresolved draft containing null hardware values.
- In a separate isolated test workspace, importing an explicitly labeled test profile changed the workspace to a 180 x 160 x 150 mm bed, 0.6 mm nozzle, 65 mm/s print speed, 170-220 C nozzle range and 0-70 C bed range. The user's workspace was left unconfigured.
- Material selection offered PLA, PETG, ABS, ASA, TPU95A and PA12. Recipes exceeding that test machine's nozzle/bed limits were disabled and explained. Selecting TPU changed displayed temperatures and effective flow ceiling; material management opened the selected material, independently of printer settings.
- The guide rendered actual offline Markdown, links, field tables and instructions in the application. The complete ZIP kit downloaded successfully and includes all documentation, three formal schemas and a null-valued draft.
- The imported-machine UI workflow reached genuine native sliced output, 69 layers, calculated filament consumption and guarded G-code export. No warnings/errors appeared in the inspected browser console.
- Printer import/activation, saved active-printer restoration, material compatibility, preference persistence, unregistered/unsaved-printer rejection and documentation path/kit integrity have automated regression coverage.

## Checks

- Frontend: svelte-check 0 errors, 0 warnings; Vite production build passed, 120 modules. The existing large Three.js bundle advisory remains.
- Go: 31 top-level unit/example checks passed, including the offline documentation kit and printer-first regressions. Log: artifacts/printer-first-all-tests.txt.
- Real native engine: all five representative fixtures/objectives passed after explicit printer import, including cube, thin wall, near-boundary, bridge and supported overhang. Log: artifacts/printer-first-native-integration.txt.
- Application: Windows native executable builds successfully. The final native WebView startup smoke check reported that embedded UI connected to Go services; see artifacts/printer-first-native-smoke.err.log. The final live workspace was reloaded at version0.2.0 with zero configured printers and six recipes.
- Live source review found and corrected a displayed layer interval mismatch: the UI and backend now both intersect material limits with 0.2-0.8 times nozzle diameter. Placement tolerance also matches the backend's 1e-5 mm check.

One initial concurrent native integration invocation encountered an external PrusaSlicer post-export crash (0xc0000409). The job failed and output was not treated as successful. Its log is preserved at artifacts/printer-first-native-integration-first-run.txt. An isolated rerun passed every fixture. Native-process failure remains explicit; the application does not silently accept a crashed engine's output.

## Environment and remaining boundaries

Windows amd64, Go1.27.3, Wails3.0.0, Svelte5.57.3, Three0.187.0, Vite8.3.5, PrusaSlicer2.9.6. macOS/Linux builds were not verified here. Existing adapter boundaries remain: one extruder, rectangular front-left beds and supported Marlin2/Klipper/RepRapFirmware command subsets. Adaptive layers, broader macros, multi-extruder jobs, persistent model projects and validated mechanics are outside this revision. There are no physical print measurements.

Generic material recipes are starting points to review against the actual spool and machine; their sources and engineering assumptions are recorded in docs/configuration/material-catalogue.md. Simulation fixtures are development test data only and are never loaded as application defaults. The user's printer profile must be imported before slicing.
