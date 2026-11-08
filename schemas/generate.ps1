$ErrorActionPreference = 'Stop'
$utf8 = New-Object System.Text.UTF8Encoding($false)
function Field($type,$description,$unit='none',$extra=@{}) {
 $f=[ordered]@{type=$type;description=$description;'x-unit'=$unit;'x-defaultBehavior'='Required; no implicit default';'x-dependencies'='See validation-rules.md and firmware-compatibility.md';'x-commonMistake'='Do not infer a machine capability from a similarly named model.';'x-safety'='Verify values against the exact hardware and firmware before printing.'}
 foreach($k in $extra.Keys){$f[$k]=$extra[$k]};return $f
}
function Num($description,$unit,$min,$max){return (Field 'number' $description $unit @{minimum=[double]$min;maximum=[double]$max})}
function Enum($description,$choices){return (Field 'string' $description 'none' @{enum=@($choices)})}
function Obj($description,$properties){return (Field 'object' $description 'none' @{properties=$properties;required=@($properties.Keys);additionalProperties=$false})}
function Arr($description,$items,$min=0,$max=256){return (Field 'array' $description 'none' @{items=$items;minItems=$min;maxItems=$max})}
function Text($description,$max=256){return (Field 'string' $description 'none' @{minLength=1;maxLength=$max})}
$id=Field 'string' 'Portable identifier, unique in its profile category.' 'none' @{pattern='^[a-z0-9][a-z0-9._-]{0,63}$';examples=@('my-printer-main')}
$version=Field 'string' 'Schema compatibility version. Unknown versions require an explicit future migration.' 'none' @{const='1.0';examples=@('1.0')}
$commands=Arr 'Explicitly verified firmware commands. Command names only; never executable shell commands.' (Field 'string' 'G-code command name.' 'none' @{pattern='^(G|M|T)[0-9]+$'}) 0 256
$snippet=Arr 'Literal G-code lines. No template expansion, macros, scripts or shell execution. Commands must be supported and output must pass full G-code validation.' (Field 'string' 'Single literal G-code line; semicolon comments permitted.' 'none' @{maxLength=256;pattern='^[^\r\n]*$'}) 0 128
$region=Obj 'Excluded rectangular bed area. Model and generated motions must avoid it.' ([ordered]@{x=(Num 'Lower-left X coordinate.' 'mm' 0 2000);y=(Num 'Lower-left Y coordinate.' 'mm' 0 2000);width=(Num 'Excluded rectangle width.' 'mm' 0.01 2000);depth=(Num 'Excluded rectangle depth.' 'mm' 0.01 2000)})
$extruder=Obj 'One extrusion system. Multi-tool hardware is outside initial release scope.' ([ordered]@{
 nozzleDiameter=(Num 'Installed nozzle bore diameter, measured or manufacturer specified.' 'mm' 0.1 2)
 filamentDiameter=(Num 'Nominal input filament diameter.' 'mm' 1 4)
 maxVolumetricFlow=(Num 'Configured hot-end flow cap. Factory defaults and conservative preset values are accepted without calibration or verification; material profiles may reduce this cap. New drafts start at 4 mm3/s.' 'mm3/s' 0.1 100)
 retractionLength=(Num 'Explicit extrusion retraction distance.' 'mm' 0 15)
 retractionSpeed=(Num 'Retraction feed rate, never above machine speed cap.' 'mm/s' 1 150)
 flowVerified=(Field 'boolean' 'Legacy informational metadata. Omitted, false and true are all accepted; this field never blocks import or slicing.')
 type=(Enum 'Extruder feed architecture, used for documentation and retraction interpretation.' @('direct','bowden'))
})
$extruder.required=@($extruder.required | Where-Object {$_ -ne 'flowVerified'})
$extruder.properties.flowVerified['x-defaultBehavior']='Optional compatibility field; ignored by validation and slicing policy.'
$printer=[ordered]@{
 schemaVersion=$version;id=$id;name=(Text 'Human-readable machine name.');manufacturer=(Text 'Verified manufacturer identity.');model=(Text 'Exact installed model/revision, not a guessed family.')
 firmware=(Obj 'Firmware identity and explicitly supported G-code contract.' ([ordered]@{
 family=(Enum 'Verified firmware family.' @('marlin','klipper','reprap'))
 version=(Text 'Exact installed version or documented verified compatibility range.')
 gcodeFlavor=(Enum 'PrusaSlicer adapter flavor. Must agree with family.' @('marlin2','klipper','reprapfirmware'))
 extrusionMode=(Enum 'Extrusion coordinate interpretation. Initializer sets matching M82 or M83.' @('absolute','relative'))
 supportedCommands=$commands;unsupportedCommands=$commands;startGcode=$snippet;endGcode=$snippet
 }))
 buildVolume=(Obj 'Usable Cartesian printing volume after permanent mechanical clearances. Initial adapter supports front-left rectangular beds.' ([ordered]@{
 width=(Num 'Usable X extent after toolhead clearances.' 'mm' 10 2000);depth=(Num 'Usable Y extent after toolhead clearances.' 'mm' 10 2000);height=(Num 'Usable Z extent.' 'mm' 10 2000)
 shape=(Enum 'Supported bed geometry.' @('rectangular'));origin=(Enum 'Coordinate origin used by model placement and G-code.' @('front-left'));excludedRegions=(Arr 'Rectangular regions unavailable for printing or motion.' $region 0 64)
 }))
 motion=(Obj 'Non-overridable machine motion limits. Profile speeds are capabilities, not print presets.' ([ordered]@{maxSpeed=(Num 'Maximum permitted extrusion motion speed.' 'mm/s' 1 1000);maxTravelSpeed=(Num 'Maximum permitted non-extruding motion speed.' 'mm/s' 1 1500);maxAcceleration=(Num 'Maximum permitted commanded acceleration.' 'mm/s2' 1 50000)}))
 extruders=(Arr 'Exactly one extruder supported in initial adapter.' $extruder 1 1)
 thermal=(Obj 'Verified thermal hardware range and cooling availability.' ([ordered]@{minNozzle=(Num 'Minimum supported printing temperature.' 'C' 0 500);maxNozzle=(Num 'Maximum permitted nozzle temperature.' 'C' 1 500);minBed=(Num 'Minimum permitted bed temperature; zero for unheated.' 'C' 0 200);maxBed=(Num 'Maximum permitted bed temperature; zero for unheated.' 'C' 0 200);hasHeatedBed=(Field 'boolean' 'Verified heated bed; false requires zero minBed and maxBed.');partCooling=(Field 'boolean' 'Verified controllable part cooling fan.')}))
 capabilities=(Obj 'Verified optional hardware. Unknown capabilities must be false. Flags do not emit unreviewed macros.' ([ordered]@{autoBedLeveling=(Field 'boolean' 'Verified automatic bed leveling hardware. This does not automatically run a leveling command.');firmwareRetraction=(Field 'boolean' 'Must be false in initial adapter; explicit E retraction is used.' 'none' @{const=$false});pressureAdvance=(Field 'boolean' 'Records verified support only. No pressure advance commands are generated.')}))
 provenance=(Field 'object' 'Documentation for machine capability groups. A flow source/calibration record is optional; default flow caps do not require certification.' 'none' @{required=@('buildVolume','motion','thermal');additionalProperties=(Text 'Documentation source or calibration record; identify reference/simulation fixtures explicitly.' 2048);properties=[ordered]@{buildVolume=(Text 'Evidence for usable dimensions.' 2048);motion=(Text 'Evidence for motion limits.' 2048);thermal=(Text 'Evidence for thermal limits.' 2048);maxVolumetricFlow=(Text 'Optional note describing a default flow cap or calibration.' 2048)}})
}
$filament=[ordered]@{
 schemaVersion=$version;id=$id;name=(Text 'Human-readable filament identity.');material=(Enum 'Polymer family; does not infer temperatures. PLA+ identifies a modified PLA family and is mapped to PLA for native engine categorization.' @('PLA','PLA+','PETG','ABS','ASA','TPU','PA','PC','other'))
 density=(Num 'Manufacturer or measured bulk polymer density.' 'g/cm3' 0.5 3)
 nozzleTemperature=(Num 'Chosen nozzle setpoint, constrained by machine thermal limits. Custom unverified values require explicit notes and process review; they do not weaken machine limits.' 'C' 100 450)
 bedTemperature=(Num 'Chosen bed setpoint; must be zero for unheated beds.' 'C' 0 180)
 maxVolumetricFlow=(Num 'Material-specific conservative throughput; effective flow is min(machine,material).' 'mm3/s' 0.1 100)
 minLayerHeight=(Num 'Minimum acceptable layer height for this material.' 'mm' 0.02 1.5)
 maxLayerHeight=(Num 'Maximum acceptable layer height, also capped at 0.8 nozzle diameters.' 'mm' 0.02 1.5)
 fanPercent=(Field 'integer' 'Part cooling setpoint; positive requires verified machine fan.' '%' @{minimum=0;maximum=100;examples=@(100)})
 notes=(Arr 'Optional material-specific warnings, provenance notes and unverified custom assumptions. Display these before slicing; never interpret them as executable instructions.' (Text 'Human-readable material warning or assumption.' 1024) 0 24)
}
$optimization=[ordered]@{
 schemaVersion=$version
 mode=(Enum 'Objective function, not a preset. See optimization/mathematics.md.' @('fast','balanced','quality','strength','material'))
 budget=(Field 'integer' 'Maximum unique feasible candidate evaluations; deterministic search terminates when exhausted.' 'candidates' @{minimum=16;maximum=4096;examples=@(256)})
 minWalls=(Field 'integer' 'User functional minimum perimeter count; optimizer may increase it.' 'count' @{minimum=1;maximum=12;examples=@(2)})
 minInfill=(Num 'User functional minimum interior fill fraction.' 'fraction' 0 1)
 supports=(Field 'boolean' 'Permit automatic support generation when geometry analysis finds overhangs. False with substantial overhangs produces a warning.')
 loadDirection=(Arr 'Optional nominal unitless XYZ load vector. Zero vector means unspecified; directional strength remains an unvalidated engineering proxy.' (Num 'Load vector component in build coordinates.' 'unitless' -1e6 1e6) 3 3)
 supportMode=(Enum 'Support objective: auto uses geometry readiness; optimized compares actual native alternatives; easy_remove favors contact removal; stable favors support robustness; off disables supports explicitly.' @('auto','optimized','easy_remove','stable','off'))
 supportStyle=(Enum 'Native support structure. Auto permits supported native style comparison; organic, snug and grid select explicit native styles.' @('auto','organic','snug','grid'))
 supportPlacement=(Enum 'Everywhere can anchor on lower model geometry; buildplate restricts supports to bed-connected placement and can leave internal overhangs unsupported.' @('everywhere','buildplate'))
 supportAngle=(Num 'Overhang slope threshold from the horizontal plane. Higher thresholds include steeper faces and can request more support. This is not an angle measured from vertical.' 'degrees' 20 80)
}
$optimization.loadDirection['x-defaultBehavior']='Optional; omitted or zero vector means unknown load direction.'
$optimization.supportMode['x-defaultBehavior']='Optional; when absent legacy supports true means auto and false means off. An explicit mode takes precedence.'
$optimization.supportStyle['x-defaultBehavior']='Optional; omitted means auto.'
$optimization.supportPlacement['x-defaultBehavior']='Optional; omitted means everywhere.'
$optimization.supportAngle['x-defaultBehavior']='Optional; omitted means 45 degrees. Internal Go zero value is normalized to 45; explicit JSON values must be 20 through 80.'
$filament.notes['x-defaultBehavior']='Optional; omitted means no extra material-specific notes, not evidence of calibration.'
foreach($kind in @('printer','filament','optimization')){
 $props=Get-Variable -Name $kind -ValueOnly
 $optional=@('loadDirection','supportMode','supportStyle','supportPlacement','supportAngle','notes')
 $required=@($props.Keys | Where-Object {$optional -notcontains $_})
 $schema=[ordered]@{'$schema'='https://json-schema.org/draft/2020-12/schema';'$id'="https://forge-slicer.local/schemas/$kind.schema.json";title="Forge $kind profile 1.0";type='object';additionalProperties=$false;required=$required;properties=$props}
 [IO.File]::WriteAllText((Join-Path $PSScriptRoot "$kind.schema.json"),($schema | ConvertTo-Json -Depth 40),$utf8)
 $lines=New-Object System.Collections.Generic.List[string]
 $lines.Add("# $kind schema parameter reference")
 $lines.Add('')
 $lines.Add('Generated from the bundled JSON Schema by `schemas/generate.ps1`. Normative rules are the schema plus validation-rules.md. Every field is required unless marked optional. No implicit safety-critical defaults exist. All examples are illustrative and require hardware verification.')
 function Walk($object,$prefix,$required){foreach($name in $object.Keys){$p=$object[$name];$path=if($prefix){"$prefix.$name"}else{$name};$status=if($required -contains $name){'Required'}else{'Optional'};$lines.Add('');$lines.Add("## ``$path``");$lines.Add('');$lines.Add("$($p.description)");$lines.Add('');$lines.Add("- Type: $($p.type). Unit: $($p['x-unit']). $status.");$lines.Add("- Default behavior: $($p['x-defaultBehavior']).");if($p.Contains('enum')){$lines.Add("- Accepted values: $($p.enum -join ', ').")};if($p.Contains('const')){$lines.Add("- Required constant: $($p.const).")};if($p.Contains('minimum')){$lines.Add("- Inclusive bounds: $($p.minimum) to $($p.maximum).")};if($p.Contains('pattern')){$lines.Add("- Pattern: ``$($p.pattern)``.")};if($p.Contains('minItems')){$lines.Add("- Array length: $($p.minItems) to $($p.maxItems).")};if($p.Contains('examples')){$lines.Add("- Example: ``$($p.examples -join ', ')``.")}else{$lines.Add('- Example: see the complete examples/profiles files; retain the field type and units.')};$lines.Add("- Hardware / firmware dependencies and interactions: $($p['x-dependencies']).");$lines.Add("- Common mistakes: $($p['x-commonMistake']).");$lines.Add("- Safety: $($p['x-safety']).");if($p.Contains('properties')){Walk $p.properties $path $p.required};if($p.type -eq 'array' -and $p.items.Contains('properties')){Walk $p.items.properties "$path[]" $p.items.required}}}
 Walk $props '' $required
 [IO.File]::WriteAllLines((Join-Path $PSScriptRoot "../docs/configuration/$kind-schema.md"),$lines,$utf8)
}
