# Create Addon
$ErrorActionPreference = "Stop"
Compress-Archive -Path .\siemcraft_addon_behavior -DestinationPath .\bin\siemcraft_addon_behavior.mcpack -Force
Compress-Archive -Path .\siemcraft_addon_resource -DestinationPath .\bin\siemcraft_addon_resource.mcpack -Force
Compress-Archive -Path .\bin\*.mcpack -DestinationPath .\bin\siemcraft.mcaddon -Force

# Build Server
go build -o .\bin\siemcraft.exe ./src
if (-not $?) {
    throw 'Go build failure'
}
