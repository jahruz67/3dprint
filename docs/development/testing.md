# Verification

User instruction for this implementation: no builds or tests until implementation integration is complete. Test source and fixtures may be written earlier. Actual results are recorded in VERIFICATION.md after the final verification phase; never infer passing from test presence.

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
powershell -ExecutionPolicy Bypass -File scripts/verify.ps1 -Integration
```

Unit suites cover strict schemas/duplicate keys, profile composition/firmware constraints, optimizer objectives/bounds/caching/cancellation, STL area/volume/transforms and 3MF units/build components, G-code mode/temperature/feed/flow/bounds/excluded-region/shutdown semantics and retraction accounting. HTTP suites check origin rejection, invalid profile feedback, guarded export and cancellation.

Set FORGE_INTEGRATION=1 and FORGE_ENGINE to a supported native executable to run the real import → profile → optimize → selected/baseline native slicing → validate → preview → export suite. It includes calibration, thin-wall and near-bed-boundary fixtures. Run `go test ./...` without the environment flag for unit-only verification. No fake backend stands in for the native integration test.

Additional checked-in meshes cover bridging, overhang/supports and malformed input. These can be imported interactively or used for further physical calibration. This initial release has no physical print measurements; automated results establish software behavior, not strength, dimensional accuracy, visual quality or guaranteed physical safety. Native computation time and G-code path/filament estimates are recorded in integration logs. The optimizer benchmark records allocations; external OS profiling is required for native peak process memory.

Browser verification should cover model import, viewport transformation, all modes, cancellation, actual preview layers/feature colors, profile editing validation, native-engine failure and G-code download. Native desktop verification additionally checks window launch and OS save dialog. Cross-platform builds and signed installers require their native CI runners and are not established by a Windows build alone.

`forge-console.exe --smoke-test` opens a hidden native Wails window, waits up to 20 seconds for the embedded frontend to call the Go state endpoint, then exits with a success/failure code. This checks native startup/connectivity without sending anything to a printer. It does not automate the OS save dialog.
