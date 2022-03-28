# Build
$ErrorActionPreference = "Stop"
. "$PSScriptRoot\build.ps1"

# Run
"$PSScriptRoot\bin\siemcraft.exe -rules $PSScriptRoot\rules"
