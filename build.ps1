# Create Addon
$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force -Path "$PSScriptRoot\bin"
Compress-Archive -Path "$PSScriptRoot\siemcraft_addon_behavior" -DestinationPath "$PSScriptRoot\bin\siemcraft_addon_behavior.mcpack" -Force
Compress-Archive -Path "$PSScriptRoot\siemcraft_addon_gametest" -DestinationPath "$PSScriptRoot\bin\siemcraft_addon_gametest.mcpack" -Force
Compress-Archive -Path "$PSScriptRoot\siemcraft_addon_resource" -DestinationPath "$PSScriptRoot\bin\siemcraft_addon_resource.mcpack" -Force
Compress-Archive -Path "$PSScriptRoot\bin\*.mcpack" -DestinationPath "$PSScriptRoot\bin\siemcraft.mcaddon" -Force

# Build Server
go build -o "$PSScriptRoot\bin\siemcraft.exe" ./src
if (-not $?) {
    throw 'Go build failure'
}
