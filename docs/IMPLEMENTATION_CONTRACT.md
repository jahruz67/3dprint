# Implementation contract

Module: `github.com/forge-slicer/forge`. Product: Forge Slicer. Go backend + Wails v3; Svelte/TypeScript/Three frontend. Native engine: external PrusaSlicer. No builds or tests until final integration (user instruction). Dependency downloading and source inspection are allowed.

## HTTP API (same origin, Wails asset middleware + standalone loopback server)
All errors: HTTP 4xx/5xx `{error:string}`. JSON keys camelCase. POST JSON except file upload.
- GET /api/state -> {printers:Printer[],filaments:Filament[],engine:{available:boolean,path:string,version:string},version:string}
- POST /api/models multipart field `file` (.stl or .3mf) -> Model
- GET /api/models/{id}/mesh -> {vertices:number[]} (triangle soup XYZ in mm)
- POST /api/jobs -> {modelId:string,printer:Printer,filament:Filament,mode:string,transform:{rotation:[number,number,number],scale:[number,number,number],position:[number,number,number]},budget:number,loadDirection?:[number,number,number]} returns Job. Rotation degrees XYZ; scale defaults [1,1,1]. Model coordinates normalized with centered XY and bottom Z=0 before transform; position is bed XY center and Z bottom. Default position bed center. Backend applies transform and verifies bounds.
- GET /api/jobs/{id} -> Job
- DELETE /api/jobs/{id} -> cancel
- GET /api/jobs/{id}/preview -> {segments:number[],layers:number[],estimatedSeconds:number,filamentMm:number,massG:number,warnings:string[]} Segments flat stride 9: x1,y1,z1,x2,y2,z2,kind,layer,feedMmMin. kind 0 travel,1 perimeter,2 infill,3 support,4 other extrusion. layer zero based.
- GET /api/jobs/{id}/gcode -> attachment only after validation passes.
- POST /api/profiles/validate {kind:'printer'|'filament'|'optimization',profile:object} -> {valid:boolean,errors:string[]}
- POST /api/profiles {kind,profile} -> saved validated profile
- POST /api/engine {path:string} -> engine status (path must identify executable; never a shell command)

Model: {id,name,triangles,bounds:{min:[x,y,z],max:[x,y,z]},volumeMm3,areaMm2,warnings:string[]}
Job: {id,status:'queued'|'optimizing'|'slicing'|'validating'|'complete'|'cancelled'|'failed',progress:number,message:string,error?:string,report?:object,previewReady:boolean}
Mode identifiers: fast,balanced,quality,strength,material.

## Shared Go contracts
`internal/config` exports Printer, Filament, Preferences, Defaults() ([]Printer, []Filament), ParsePrinter([]byte)(Printer,error), ParseFilament([]byte)(Filament,error), ParsePreferences([]byte)(Preferences,error), ValidatePrinter(Printer) error, ValidateFilament(Filament) error. Use concrete types below (may extend with fields):
Printer {SchemaVersion string; ID string; Name string; Manufacturer string; Model string; Firmware Firmware; BuildVolume BuildVolume; Motion Motion; Extruders []Extruder; Thermal Thermal; Capabilities Capabilities; Provenance map[string]string}
Firmware {Family string; Version string; GCodeFlavor string; ExtrusionMode string; SupportedCommands []string; UnsupportedCommands []string; StartGCode []string; EndGCode []string}
BuildVolume {Width,Depth,Height float64; Shape,Origin string; ExcludedRegions []Region}; Region {X,Y,Width,Depth float64}
Motion {MaxSpeed,MaxTravelSpeed,MaxAcceleration float64}
Extruder {NozzleDiameter,FilamentDiameter,MaxVolumetricFlow,RetractionLength,RetractionSpeed float64; FlowVerified bool; Type string}
Thermal {MinNozzle,MaxNozzle,MinBed,MaxBed float64; HasHeatedBed,PartCooling bool}
Capabilities {AutoBedLeveling,FirmwareRetraction,PressureAdvance bool}
Filament {SchemaVersion,ID,Name,Material string; Density,NozzleTemperature,BedTemperature,MaxVolumetricFlow,MinLayerHeight,MaxLayerHeight float64; FanPercent int}
Preferences {SchemaVersion,Mode string; Budget int; MinWalls int; MinInfill float64; Supports bool}
All JSON fields lower camelCase. Dimensions mm, speeds mm/s, acceleration mm/s2, flow mm3/s, density g/cm3, temperatures Celsius, infill fraction 0..1.
`internal/optimizer` exports Geometry {Width,Depth,Height,Volume,Area,OverhangArea,MinFeature float64}; Settings {LayerHeight,LineWidth float64; Walls int; Infill float64; InfillPattern string; Speed,TravelSpeed,Acceleration float64; Supports bool; SupportStyle string; Rotation [3]float64}; Estimate {Seconds,MassG,VolumeMm3,QualityError,StrengthProxy float64}; Report {Mode string; Selected Settings; Baseline Settings; Predicted Estimate; BaselinePredicted Estimate; Evaluated int; ParetoCount int; Constraints,Assumptions,Changes []string}; Optimize(ctx context.Context,g Geometry,p config.Printer,f config.Filament,pref config.Preferences)(Report,error).
Settings are analytical optimizer estimates. Actual sliced metrics returned from G-code separately. No claim of global optimum or actual improvement unless both sliced.

## Printer-first revision (0.2)
`config.Defaults()` returns no printers and six material starters. `ReferenceFixture()` is development-test-only. `PrinterTemplate()` returns a null-valued unresolved draft.
- GET /api/state additionally returns activePrinterId, persisted across restarts.
- POST /api/printers/select {id} activates an already imported printer.
- GET /api/materials?printerId=... returns [{profile:Filament,compatible:boolean,reasons:string[]}]. Compatibility is declared temperature/cooling/layer constraint compatibility, not a physical material guarantee.
- GET /api/printer-template returns the editable draft JSON object.
- GET /api/docs?path=configuration/getting-started.md returns {title,content,path} for an embedded Markdown file.
- GET /api/profile-kit downloads the complete offline printer-creation kit.
- Saving a printer validates/persists/activates it. Slicing requires its exact saved hardware profile. First-run UI is printer import/creation/help; model importing is a secondary control.
