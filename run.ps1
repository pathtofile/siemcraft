# Build
$ErrorActionPreference = "Stop"
. "$PSScriptRoot\build.ps1"

# Run
# .\bin\siemcraft.exe -noKill
# .\bin\siemcraft.exe -fakeEvents -rules C:\code\siemcraft\sigma\rules
.\bin\siemcraft.exe -channels "Sealighter/Operational"
