# Forge desktop interface

Svelte 5, TypeScript, and Three.js. All assets are local; the application uses no CDN, external fonts, analytics, or network APIs. See the root development guide for application builds.

Install with `pnpm install --frozen-lockfile --ignore-scripts` (pnpm 11) or generate an npm lockfile with `npm install --ignore-scripts`. Run `pnpm check` and `pnpm build` during final verification. The checked-in pnpm lockfile pins the complete frontend dependency graph. This checkout was initialized using the bundled Codex Node/pnpm runtime; frontend builds and checks were deferred until the final verification phase.

`pnpm dev` proxies `/api` to the standalone Go development server at `127.0.0.1:8765`. The production bundle uses relative, same-origin API routes and works in the Wails native window.

## Viewport behavior

- The machine uses an XY plate and a Z-up camera. The model is centered in XY and placed at Z=0 before scaling and rotation. After transformation it is re-centered in XY and re-grounded at its configured position; this matches the backend convention.
- G-code uses one buffered `LineSegments` object per feature class. Layer and print-order sliders change buffer draw ranges rather than allocating UI objects per segment. Per-feature segment order must remain monotonic by layer and original G-code order.
- The auto-orient worker compares up to 24 large-face-normal orientations using up to 24,000 sampled triangles. Its overhang/contact-area score is a bounded engineering heuristic, not a guarantee of printability or minimum support. The result is returned as XYZ angle components applied in extrinsic X-then-Y-then-Z order (Three.js Euler order ZYX) and clearly identified in the interface.
- The dimensions in the object panel are the transformed bounding envelope and may conservatively exceed the actual rotated mesh dimensions. Exact placement validity is checked by the backend before slicing.
- Hardware-accelerated WebGL2 is needed for the viewport. A driver failure produces a message; profile editing, native slicing, and export remain separate services.

## Profile and result handling

Profiles are edited as JSON, validated by the backend schema and semantic rules, and persisted only after successful validation. Imported optimization preferences apply to the current setup after saving. Importing a printer or material selects that saved profile.

The optimization report separates two comparisons: native G-code metrics from two actual slicing runs, and analytical search-model estimates. Physical print duration, strength, and surface quality require calibration and real-world validation. Native desktop export uses a system Save dialog; browser development mode uses a normal G-code download.
