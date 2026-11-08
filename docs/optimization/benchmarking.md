# Reproducible benchmarking

Run benchmarks only in the final validation stage for this implementation, per the user's request. The automated `BenchmarkOptimizeCube` benchmark exercises a fixed 20 mm cube, reference printer/material and a 256-candidate balanced search. Use `go test -bench=BenchmarkOptimizeCube -benchmem ./internal/optimizer` after the ordinary test suite passes. Record Go version, OS, CPU, source revision, budget, mode, profile identifiers/content hashes and number of repetitions alongside output; do not copy a machine-specific number into a universal speed claim.

For native slicing comparisons, keep a model suite containing a calibration cube, overhang, bridge, thin wall, support-requiring shape, near-boundary volume and malformed geometry. Slice the **same transformed model** twice with the recorded baseline and selected settings. Record wall-clock native process duration, peak process memory where OS instrumentation is available, parsed estimated print seconds, positive extruded filament length/mass, deposited/travel path length and support extrusion separately. Do not infer support mass from total mass or claim a peak memory value when it was not measured.

The report's analytical baseline/selected values are suitable for checking optimization decisions, not for proving native toolpath improvements. When native baseline slicing has not been run, label the comparison “analytical estimate”. G-code time should identify whether it is a native-engine estimate or parser approximation. Physical print duration, finished mass, dimensional error, surface scoring and strength need independent measured records.

Recommended physical-measurement CSV fields:

```csv
timestamp,source_revision,model_sha256,printer_profile_sha256,filament_profile_sha256,mode,gcode_sha256,ambient_c,nozzle_c,bed_c,actual_seconds,finished_mass_g,x_error_mm,y_error_mm,z_error_mm,load_direction,load_at_failure_n,method,notes
```

Leave unmeasured values empty, not zero. Use repeated paired baseline/selected prints and documented instruments. Strength records must identify specimen geometry, conditioning, loading, fixture, rate and failure mode; one anecdotal print is not material validation. Update mathematical calibration factors only from traceable data and preserve the original measurement files.
