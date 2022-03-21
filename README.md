# SIEMCraft

Idea from [https://github.com/erjadi/kubecraftadmin](https://github.com/erjadi/kubecraftadmin)

## Links
http://www.s-anand.net/blog/programming-minecraft-with-websockets/
https://github.com/sanand0/minecraft-websocket/blob/dev/tutorial/mineserver2.py
https://bedrock.dev/docs/stable/Scripting#Slash%20Commands

# Still TODO
SIGMA:
  - More example rules?

MINECRAFT:
  - Low:
    - Chicken
  - Medium
    - Pig, Cow
  -  Hist
     -  Panda, Bear, Spider
  - Add pandas, bears, spiders etc to enemies spawned on high
    - Make them aggressive
  - Check can you make+spawn a new type of player entity?
    - And use that to '/tell' to make chat messages invisible

DOCO:
  - how to get remote sysmon logs from WEF
  - what SIGMA rules are and aren't supported (sigma-go bug)
  - Understand effort to add ImageLoad and other events
  - Details Sigma event types:
    - process_creation
    - file_create
    - image_load
    - driver_load
    - network_connection
    - dns_query / dns
    - registry_event

%localappdata%\Packages\Microsoft.MinecraftUWP_8wekyb3d8bbwe\LocalState\games\com.mojang\development_behavior_packs


# GameTest Development
npm i @types/mojang-minecraft
npm i @types/mojang-gametest
CheckNetIsolation.exe LoopbackExempt -a -p=S-1-15-2-1958404141-86561845-1752920682-3514627264-368642714-62675701-733520436

/script debugger listen 19144