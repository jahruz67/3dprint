# Contributing

Keep the offline pipeline functional. Never introduce AI services, telemetry or cloud dependencies. Keep UI/services/geometry/profiles/optimizer/native-engine/G-code concerns separate. Changes to capability support must update schema metadata, semantic validation, engine translation, G-code state handling and documentation together.

Regenerate schemas and exhaustive field references with `powershell -File schemas/generate.ps1`. This generator writes JSON/Markdown from metadata; review both semantic validators and generated outputs. Tests must establish behavior or an important failure condition, not merely mirror implementation.

Do not claim global optimality, print quality, physical strength or benchmark improvements without corresponding evidence. Distinguish exact calculations, approximations, heuristics and measurements. Always retain native engine isolation, bounded imports/searches, cancellation, conservative unsupported-feature rejection and same-origin local API protection.

Before release, build on supported native OS runners, run all relevant automated and real-engine integration checks, inspect toolpath rendering, record evidence in VERIFICATION.md, and audit distribution licenses/source provision. Source contributions are under AGPL-3.0.
