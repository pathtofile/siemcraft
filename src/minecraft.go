package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bradleyjkemp/sigma-go"
	"github.com/sandertv/mcwss"
	"github.com/sandertv/mcwss/protocol/event"
)

var MinecraftPlayers []*mcwss.Player
var neverKillProcesses bool
var hostname string

// Used to tag messages from and to MineCraft addon
const messagePrefix = "[SIEMCRAFT]"

// Sent from the add-on when an entity is killed
type SIEMCraftMessage struct {
	// Item used to kill the entity
	Item string `json:"item"`
	// If a Projectile was used, it'd be here
	Projectile string `json:"projectile"`
	// The saved event
	EventB64 string `json:"eventb64"`
}

func getRandom(slice []string) string {
	return slice[rand.Intn(len(slice))]
}

func RaiseAlert(event map[string]string, rule *sigma.Rule, channel string) {
	// If there's not a player don't do anything
	if len(MinecraftPlayers) == 0 {
		return
	}
	fmt.Printf("[m] Rule Hit: %s\n", rule.Title)

	for _, p := range MinecraftPlayers {
		actionbar(p, fmt.Sprintf("Rule Hit: %s", rule.Title))
	}

	// Seed randomness for entity spawning
	rand.Seed(time.Now().UnixNano())

	// Select random player to spawn near
	player := MinecraftPlayers[rand.Intn(len(MinecraftPlayers))]

	name := ""
	jsonTag := ""
	if channel == "sysmon" {
		// Only display certain feilds
		simpleFields := []string{
			// Common
			"Computer",
			"ProcessId",
			"Image",
			"CommandLine",
			"User",
			// "UtcTime",
			// Process
			"ParentProcessId",
			"ParentImage",
			// "ParentCommandLine",
			// Network connection
			"SourceIp",
			"SourcePort",
			"DestinationIp",
			"DestinationPort",
			// Image and Driver Load
			"ImageLoaded",
			// File Create
			"TargetFilename",
			// Registry
			"TargetObject",
			// DNS
			"QueryName",
			"QueryStatus",
			"QueryResults",
		}
		simpleEvent := map[string]string{}
		for _, field := range simpleFields {
			if event[field] != "" {
				simpleEvent[field] = strings.ReplaceAll(event[field], "\n", "")
			}
		}
		eventData, _ := json.MarshalIndent(simpleEvent, "", "  ")
		name = string(eventData)

		// Now store event as a base64 JSON in the entity's tag
		evantDataJSON, _ := json.Marshal(simpleEvent)
		jsonTag = base64.StdEncoding.EncodeToString(evantDataJSON)
	} else {
		// For everything else, print full event because I cbf custom parsing everything
		eventData, _ := json.MarshalIndent(event, "", "  ")
		name = string(eventData)

		// Now store event as a base64 JSON in the entity's tag
		evantDataJSON, _ := json.Marshal(event)
		jsonTag = base64.StdEncoding.EncodeToString(evantDataJSON)
	}

	// Remove or escape characters that can't be easily displayed in minecraft
	name = strings.ReplaceAll(name, "\\", "\\\\")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	name = strings.ReplaceAll(name, "\"", "")

	// Add Rule header
	name = fmt.Sprintf("Rule: %s\n----------\n%s", rule.Title, name)

	// The event severity determins the entity spawned
	entity := ""
	if rule.Level == "low" {
		entity = "chicken"
	} else if rule.Level == "medium" {
		entity = getRandom([]string{"pig", "cow"})
	} else if rule.Level == "high" {
		entity = getRandom([]string{"spider", "panda", "siemcraft:bear"})
	}

	// Generate locations and spawn command
	xPos := rand.Intn(18) + 2
	zPos := rand.Intn(18) + 2
	if rand.Float32() < 0.5 {
		xPos = xPos * -1
	}
	if rand.Float32() < 0.5 {
		zPos = zPos * -1
	}
	fmt.Printf("[m] Spawning %s at %d:%d from player %s:\n", entity, xPos, zPos, player.Name())
	spawnCmd := fmt.Sprintf("summon %s \"%s\" ~%d ~ ~%d", entity, name, xPos, zPos)
	player.Exec(spawnCmd, nil)

	// Add the event as JSON in a tag
	// This command selects any un-tagged entity within a small radious of where we jsut summoned it
	// (in case it was spawned in the air and has started to fall)
	// We have to do these tag shenanigans to select the target, as we can't get a multi-line name
	// And we can't rename after the fact (so can't 'summon as simpleName->tag->rename to multiline')
	jsonTag = fmt.Sprintf("%s%s", messagePrefix, jsonTag)
	tagCmd := fmt.Sprintf("tag @e[tag=,type=%s,x=~%d,y=~,z=~%d,r=50,rm=0] add \"%s\"", entity, xPos, zPos, jsonTag)
	player.Exec(tagCmd, nil)

	// Teleport to a safe place, as they may have spawned into the wall
	spreadCmd := fmt.Sprintf("spreadplayers ~%d ~%d 0 5 @e[tag=\"%s\"]", xPos, zPos, jsonTag)
	player.Exec(spreadCmd, nil)
}

// actionbar will display a message to the player
func actionbar(player *mcwss.Player, message string) {
	player.Exec(fmt.Sprintf("title %s actionbar %s", player.Name(), message), nil)
}

func onMessage(event *event.PlayerMessage) {
	if event.Sender != "Script Engine" || event.MessageType != "tell" {
		return
	}
	// First trim quotes, then check and trim prefix
	msgString := event.Message[1 : len(event.Message)-1]
	if !strings.HasPrefix(msgString, messagePrefix) {
		return
	}
	msgString = msgString[len(messagePrefix):]

	// De-serialise from JSON
	message := SIEMCraftMessage{}
	err := json.Unmarshal([]byte(msgString), &message)
	if err != nil {
		// Not our message
		fmt.Printf("[m] Not our message? %s\n", err.Error())
		return
	}
	// Decode event from base64
	eventString, err := base64.StdEncoding.DecodeString(message.EventB64)
	if err != nil {
		fmt.Printf("[m] Failed to decode Evnet from message: %s\n", err.Error())
		return
	}
	messageEvent := map[string]string{}
	err = json.Unmarshal([]byte(eventString), &messageEvent)
	if err != nil {
		fmt.Printf("[m] Failed to unmarshal event: %s\n", err.Error())
		return
	}

	// Check Item used to kill, determine if we're killing process
	// if neverKillProcesses is set, then also never attempt to kill it
	// Also check event is from this machine and not a remote one (i.e. a WEF forwarded event)
	if !neverKillProcesses && messageEvent["Computer"] == hostname && strings.Contains(message.Item, "diamond_sword") {
		// Diamond weapon, kill a process!

		// kill if image is one of these
		killableImages := []string{"cmd.exe", "pwsh.exe", "powershell.exe", "wword.exe"}
		pidToKill := 0
		for _, killableImage := range killableImages {
			if strings.HasSuffix(messageEvent["Image"], killableImage) {
				pidToKill, _ = strconv.Atoi(messageEvent["ProcessId"])
				fmt.Printf("[*] Killing Pid %d (%s)\n", pidToKill, messageEvent["CommandLine"])
			}
		}
		if pidToKill == 0 {
			// Check parent, maybe we can kill it?
			for _, killableImage := range killableImages {
				if strings.HasSuffix(messageEvent["ParentImage"], killableImage) {
					pidToKill, _ = strconv.Atoi(messageEvent["ParentProcessId"])
					fmt.Printf("[*] Killing Parent Pid %d (%s)\n", pidToKill, messageEvent["ParentCommandLine"])
				}
			}
		}
		// Pid could be < 0 in testing, and 0 if there was no process ID found
		if pidToKill <= 0 {
			fmt.Printf("[m] Event handled\n")
		} else {
			process, err := os.FindProcess(pidToKill)
			if err != nil {
				fmt.Printf("[*] Failed to find process %d: %s\n", pidToKill, err.Error())
				return
			}
			err = process.Kill()
			if err != nil {
				fmt.Printf("[*] Failed to kill process %d: %s\n", pidToKill, err.Error())
				return
			}
		}
	} else {
		fmt.Printf("[m] Event handled for PID %s\n", messageEvent["ProcessId"])
	}
}

func onConnection(player *mcwss.Player) {
	playerName := player.Name()
	fmt.Printf("[m] Player %s has connected\n", playerName)

	// Send welcome message
	player.Exec(fmt.Sprintf("title %s title SIEMCraft", playerName), nil)

	// Provide player with equipment
	player.Exec("give @s diamond_sword", nil)
	player.Exec("give @s netherite_sword", nil)

	// TODO: remove?
	player.Exec("time set noon", nil)
	player.Exec("weather clear", nil)
	player.Exec("alwaysday", nil)

	// Register to recieve messages, we'll
	// use them to get data back from minecraft addon
	player.OnPlayerMessage(onMessage)

	// Add player
	MinecraftPlayers = append(MinecraftPlayers, player)
}

func onDisconnection(player *mcwss.Player) {
	// Log player disconnect
	// Find and remove player from list
	for i, p := range MinecraftPlayers {
		if p == player {
			fmt.Printf("[m] Main player %s has disconnected\n", player.Name())
			MinecraftPlayers = append(MinecraftPlayers[:i], MinecraftPlayers[i+1:]...)
			break
		}
	}
}

func setHostName() error {
	// Gotta do this the dodgy way
	cmd := exec.Command("powershell", "-c", `[System.Net.Dns]::GetHostByName($env:computerName).HostName`)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	hostname = strings.TrimSpace(out.String())
	return nil
}

func StartMinecraftWebsocket(bindAddress string, bindPort string, noKill bool) error {
	MinecraftPlayers = make([]*mcwss.Player, 0, 5)

	// Create a new server using the default configuration. To use specific configuration, pass a *wss.Config{} in here.
	address := fmt.Sprintf("%s:%s", bindAddress, bindPort)
	var c = mcwss.Config{HandlerPattern: "/ws", Address: address}
	server := mcwss.NewServer(&c)
	if server == nil {
		return errors.New("failed to create minecraft server config")
	}

	neverKillProcesses = noKill
	err := setHostName()
	if err != nil {
		return err
	}

	server.OnConnection(onConnection)
	server.OnDisconnection(onDisconnection)

	// Start websocket server
	fmt.Println("[m] starting SIEMCraft, run this command to connect:")
	fmt.Printf("    /connect %s/ws\n", address)
	var errorGroup errgroup.Group
	errorGroup.Go(func() error {
		return server.Run()
	})

	// Run forever (or until ctrl+c)
	return errorGroup.Wait()
}
