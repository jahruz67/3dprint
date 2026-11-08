# Third-party software and distribution

Forge Slicer source is licensed under GNU AGPL version 3 (see LICENSE). Copyright 2026 Forge Slicer contributors. It is provided without warranty. The application contains no AI, cloud inference, analytics, or account requirement.

Direct dependencies:

| Software | Pinned version | License | Upstream |
|---|---|---|---|
| Wails | 3.0.0 | MIT | https://github.com/wailsapp/wails |
| Go JSON Schema | 6.0.3 | Apache-2.0 | https://github.com/santhosh-tekuri/jsonschema |
| Svelte | See frontend/package.json and pnpm-lock.yaml | MIT | https://github.com/sveltejs/svelte |
| Three.js | See frontend/package.json and pnpm-lock.yaml | MIT | https://github.com/mrdoob/three.js |
| Vite (build only) | See frontend/package.json and pnpm-lock.yaml | MIT | https://github.com/vitejs/vite |
| TypeScript (build only) | See frontend/package.json and pnpm-lock.yaml | Apache-2.0 | https://github.com/microsoft/TypeScript |
| PrusaSlicer (separate native process) | Validated with 2.9.6 | AGPL-3.0; additional bundled dependency notices | https://github.com/prusa3d/PrusaSlicer/tree/version_2.9.6 |

The development tools directory contains unmodified upstream downloads for local development. The packaging script deliberately does not bundle the native engine or development tools. Running an engine out of process does not itself settle licensing obligations or remove copyleft obligations. Before redistributing an engine-inclusive product, preserve upstream copyright/license notices, audit every included dependency, provide the applicable corresponding source and build instructions, and comply with all applicable AGPL terms, including modified-network-service obligations where applicable. Engine correspondence must match the exact shipped binary; an arbitrary source link is not a replacement for source provision where required.

For app distribution, include this file, LICENSE, the complete Forge source including Go and frontend lockfiles/build instructions, and the dependency notices collected by scripts/package.ps1. That script collects available package LICENSE/NOTICE/COPYING files; review the resulting notices and transitive dependencies before publishing. No legal conclusion about aggregation is implied by this architecture.

PrusaSlicer comparison: CuraEngine also uses AGPL-3.0 (https://github.com/Ultimaker/CuraEngine/blob/main/LICENSE). PrusaSlicer was selected because it provides a self-contained desktop CLI, broad firmware output, mesh repair, mature toolpath planning, conventional/snug/organic support, and portable native binaries without recreating polygon slicing. CuraEngine would require a second settings translation and engine definition packaging. The Backend interface permits another adapter later.
