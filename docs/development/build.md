# Reproducible build and run

Requirements: Go 1.26+ (development environment pins 1.27.3), Node 24 LTS, pnpm 11, and PrusaSlicer 2.9.6 native CLI. Wails is pinned to 3.0.0; frontend versions are pinned by pnpm-lock.yaml. Windows needs the Microsoft Edge WebView2 runtime. macOS needs Xcode command-line tooling; Linux needs GTK4/WebKitGTK 6 development libraries as described in the pinned Wails installation documentation. Native engine binaries are platform-specific.

Windows:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
.\bin\ForgeSlicer.exe
.\bin\ForgeSlicer.exe --serve --listen 127.0.0.1:8787
.\bin\ForgeSlicer.exe --validate-profile examples/profiles/reference-cartesian.json --kind printer
```

`build.ps1` uses portable tools/go and the bundled Codex Node runtime if Go/Node are unavailable in PATH; normal developers can use their own installations. The script installs lockfile dependencies without lifecycle scripts, checks frontend types, builds the offline frontend, then compiles the native executable. `scripts/verify.ps1` runs the final automated suites, including optional native slicing integration, and records logs under artifacts/.

Portable engine discovery: FORGE_ENGINE, PATH, tools/prusa/<distribution-directory>/prusa-slicer-console.exe, engine/prusa-slicer-console.exe next to the app. On Unix set FORGE_ENGINE to the installed prusa-slicer/PrusaSlicer executable. Settings can store its absolute executable path; a shell command or script is never accepted.

Generic source build after satisfying native Wails prerequisites:

```sh
cd frontend
pnpm install --frozen-lockfile --ignore-scripts
pnpm check
pnpm build
cd ..
go mod download
go build -trimpath -o bin/ForgeSlicer ./cmd/forge
```

Do not compile Go before generating frontend/dist: those assets are embedded. No network connection is required at runtime. Development dependency acquisition does require a network or prepopulated Go/pnpm caches.

`scripts/package.ps1` creates an app-only portable package with the built executable, source, documentation, schemas and license notices. It excludes downloaded engines and developer toolchains. Engine-inclusive public redistribution is a separate audited release step; see THIRD_PARTY_NOTICES.md.
