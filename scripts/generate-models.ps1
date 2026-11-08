# Deterministic inert ASCII STL fixtures; generates data, not application code.
$ErrorActionPreference = 'Stop'
$modelRoot = Join-Path (Split-Path -Parent $PSScriptRoot) 'examples/models'
New-Item -ItemType Directory -Force $modelRoot | Out-Null
function Add-Box([System.Text.StringBuilder]$Text,[double]$X,[double]$Y,[double]$Z,[double]$W,[double]$D,[double]$H) {
 $v = @(@($X,$Y,$Z),@(($X+$W),$Y,$Z),@(($X+$W),($Y+$D),$Z),@($X,($Y+$D),$Z),@($X,$Y,($Z+$H)),@(($X+$W),$Y,($Z+$H)),@(($X+$W),($Y+$D),($Z+$H)),@($X,($Y+$D),($Z+$H)))
 $faces = @(@(0,2,1),@(0,3,2),@(4,5,6),@(4,6,7),@(0,1,5),@(0,5,4),@(1,2,6),@(1,6,5),@(2,3,7),@(2,7,6),@(3,0,4),@(3,4,7))
 foreach ($face in $faces) { [void]$Text.AppendLine('facet normal 0 0 0'); [void]$Text.AppendLine('outer loop'); foreach ($i in $face) { $line = [string]::Format([Globalization.CultureInfo]::InvariantCulture,'vertex {0} {1} {2}',$v[$i][0],$v[$i][1],$v[$i][2]); [void]$Text.AppendLine($line) }; [void]$Text.AppendLine('endloop'); [void]$Text.AppendLine('endfacet') }
}
function Write-Model([string]$Name,[scriptblock]$Geometry) { $text = [Text.StringBuilder]::new(); [void]$text.AppendLine("solid $Name"); & $Geometry $text; [void]$text.AppendLine("endsolid $Name"); [IO.File]::WriteAllText((Join-Path $modelRoot ($Name+'.stl')),$text.ToString(),[Text.Encoding]::ASCII) }
Write-Model calibration-cube { param($t) Add-Box $t 0 0 0 20 20 20 }
Write-Model thin-wall { param($t) Add-Box $t 0 0 0 20 0.9 10 }
Write-Model boundary { param($t) Add-Box $t 0 0 0 210 5 2 }
Write-Model bridge { param($t) Add-Box $t 0 0 0 5 15 15; Add-Box $t 25 0 0 5 15 15; Add-Box $t 0 0 15 30 15 3 }
Write-Model overhang-support { param($t) Add-Box $t 0 0 0 6 15 20; Add-Box $t 0 0 20 25 15 3 }
[IO.File]::WriteAllText((Join-Path $modelRoot 'invalid.stl'),"solid invalid`nvertex NaN 0 0`nendsolid invalid",[Text.Encoding]::ASCII)
