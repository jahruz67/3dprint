# Objective functions and measurable quantities

Forge implements conventional deterministic numerical search. It uses no model, training or cloud inference. Selected parameters actually reach the native engine: layer height `h`, line width `w`, walls `n`, infill fraction `r`, infill pattern and speed `v`; travel and acceleration obey machine caps. Orientation is supplied by the user's transformed model and is not silently optimized with unsupported geometric assumptions.

The baseline is explicit and persisted in every report: `h=clamp(0.5*nozzle, material/nozzle interval)`, `w=1.1*nozzle`, `n=max(2,minWalls)`, `r=max(0.20,minInfill)`, grid infill, speed `min(50,maxSpeed,flow/(w*h))`, machine-capped travel and acceleration. Supports are enabled only when allowed and overhang area is nonzero. Comparing against this baseline never compares arbitrary profile names.

The volume model uses actual supplied model volume `V`, surface area `A`, dimensions `W,D,H` and analyzed unsupported-area estimate `O` (legacy geometry-only calls may supply a broader overhang estimate):

```
shell = min(V, max(0,A-2WD)*n*w + 2WD*max(0.8,3h))
support = supports ? O*H*0.18*0.12 : 0
Vprint = shell + (V-shell)*r + support
mass_g = Vprint * density_g_cm3 / 1000
bead_area = h*(w-h) + pi*h*h/4
path_length = Vprint / bead_area
```

Shell/support geometry is a first-order analytical approximation; it is not exact union/offset geometry. The factors `0.18` and `0.12` are explicit uncalibrated support-shape/density assumptions. Layer count is `ceil(H/h)`. Representative extrusion segment length is `L=max(2,0.55*min(W,D))`. For commanded speed `v` and acceleration `a`, the rest-to-rest segment-time calculation is exact for that idealized segment:

```
if L >= v*v/a: dt = L/v + v/a
else:          dt = 2*sqrt(L/a)
T = path_length/L*dt
  + layers*(W+D)*(0.8+2*r)/travelSpeed
  + layers*0.6 + 90
```

The full job time remains an approximation because segment geometry and overhead terms are heuristic. It does not claim firmware junction/jerk simulation or measured warm-up. Quality proxy (mm-like error) is `Q=0.55h+0.08w+0.055(v/maxSpeed)^2+0.7max(0,w-minFeature)` when a minimum feature estimate is available. This is an engineering quality proxy, not a validated roughness prediction.

Strength proxy is dimensionless:

```
anisotropy = load known ? 1-0.45*abs(loadZ)/norm(load) : 0.78
bonding = clamp(1-0.22*h/nozzle,0.6,1)
S = (1-exp(-0.42*n))*(0.45+0.55*sqrt(r))*anisotropy*bonding
```

No measured constitutive material law is used. The proxy does not predict failure stress, fatigue, creep, interlayer tensile strength or a safe load. Explicit load direction is interpreted in build coordinates; an unspecified vector is reported as an assumption. Physical verification is mandatory for structural claims.

Normalize `t=T/Tbaseline`, `m=M/Mbaseline`, `q=Q/Qbaseline`, `s=Sbaseline/S` (with small denominator guards). Minimize:

| Mode | Objective |
|---|---|
| Fast Draft (`fast`) | `t + 0.015m + 0.01q` |
| Balanced (`balanced`) | `0.40t + 0.25m + 0.25q + 0.10s` |
| High Quality (`quality`) | `q + 0.025t + 0.015m` |
| Structural Strength (`strength`) | `s + 0.025m + 0.015t` |
| Material Saver (`material`) | `m + 0.02t + 0.01q` |

Secondary terms break trade-off ties and avoid extreme cost for marginal proxy benefit. Weights are explicit engineering choices, not universal empirical truth. The selected report includes baseline/selected settings and estimates, evaluated candidate count, Pareto count, enforced constraints, changes and limitations. Actual post-slice G-code metrics are separate. No displayed percentage should be described as a real-world improvement when it is only an analytical comparison.
