package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var channels stringSet

func main() { os.Exit(mainReturnWithCode()) }
func mainReturnWithCode() int {
	rulesDir := flag.String("rules", ".\\rules", "Folder containing SIGMA rules")
	bindAddress := flag.String("bind", "127.0.0.1", "Address to bind websocket to")
	bindPort := flag.String("port", "8000", "Port to bind websocket to")
	fakeEvents := flag.Bool("fakeEvents", false, "Don't subscript to event logs, just fake generate them")
	noKill := flag.Bool("noKill", false, "Never attempt to kill a process")
	flag.Var(&channels, "channels", "Comma-seperated list of event logs to subscribe to\n(default [\"Microsoft-Windows-Sysmon/Operational\", \"Security\"])")
	flag.Parse()

	// Set default channels if none specified
	if len(channels) == 0 {
		channels = append(channels, []string{"Microsoft-Windows-Sysmon/Operational", "Security"}...)
	}

	// Load SIGMA rules
	err := ParseRules(*rulesDir)
	if err != nil {
		fmt.Printf("[*] Error Parsing Rules: %s\n", err.Error())
		return 1
	}

	if *fakeEvents {
		// Just fake them, and don't kill
		*noKill = true
		FireFakeEvents()
	} else {
		// Start getting events from Event logs
		err = StartEventSubscription(channels)
		if err != nil {
			fmt.Printf("[*] Error Starting Event Subscription: %s\n", err.Error())
			return 2
		}
		defer StopEventSubscription()
	}

	// Start Minecraft websocket to handle requests
	err = StartMinecraftWebsocket(*bindAddress, *bindPort, *noKill)
	if err != nil {
		fmt.Printf("[*] Error Starting Minecraft websocket: %s\n", err.Error())
		return 1
	}
	return 0
}

func init() {
	c := make(chan os.Signal)
	// This works on Windows, weird...
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Ensure we cleanup as we quit
		StopEventSubscription()
		os.Exit(0)
	}()
}
