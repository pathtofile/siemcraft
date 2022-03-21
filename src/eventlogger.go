package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/0xrawsec/golang-win32/win32/wevtapi"
	"github.com/bradleyjkemp/sigma-go"
	"golang.org/x/sys/windows"
)

var waitGroup sync.WaitGroup
var eventProvider *wevtapi.PullEventProvider = nil

type SealighterMessage struct {
	Header        map[string]interface{} `json:"header"`
	Properties    map[string]interface{} `json:"properties"`
	PropertyTypes map[string]interface{} `json:"property_types"`
}

func IsAdmin() (bool, error) {
	// Check if current process has privliges
	// to get event logs as a fail-fast
	var token windows.Token
	err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_QUERY,
		&token,
	)
	if err != nil {
		return false, err
	}

	return token.IsElevated(), nil
}

func StartEventSubscription(channels []string) error {
	fmt.Println("[e] Starting event subscription")

	// Only run if user is Elevated
	isAdmin, err := IsAdmin()
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("getting Event logs requires elevation")
	}

	fmt.Printf("[e] Subscribing to Event Log channels:\n")
	for _, channel := range channels {
		fmt.Printf("[e]   - %s\n", channel)
	}

	// Creat waitGroup to keep track of goroutine
	waitGroup = sync.WaitGroup{}
	eventProvider = wevtapi.NewPullEventProvider()
	waitGroup.Add(1)
	go func() {
		// Make a call to wevtapi to subscive to event log channels
		fetchFlags := wevtapi.EvtSubscribeToFutureEvents
		for e := range eventProvider.FetchEvents(channels, fetchFlags) {
			// Check event against SIGMA rules
			j := e.ToJSONEvent()
			event := j.Event.EventData
			channel := strings.ToLower(j.Event.System.Channel)
			var rule *sigma.Rule

			// Add fields that Sigma might want to check
			event["Computer"] = j.Event.System.Computer
			event["EventID"] = j.Event.System.EventID
			event["Provider_Name"] = j.Event.System.Provider.Name

			if channel == "application" || channel == "security" || channel == "system" {
				// Check standard Event Log Channels
				rule, err = CheckRules(event, channel)
				if err != nil {
					continue
				}
			} else if strings.Contains(channel, "sealighter") {
				// Need to convert event to be like Sysmon
				message := SealighterMessage{}
				err := json.Unmarshal([]byte(event["json"]), &message)
				if err != nil {
					continue
				}

				// For now only add the 'Properties' and not the ETW Header
				for k, v := range message.Properties {
					// Rename 'ImageFileName' to 'Image'
					// to match with Sysmon
					if k == "ImageFileName" {
						k = "Image"
						v = fmt.Sprintf("\\%s", v)
					}
					event[k] = fmt.Sprint(v)
				}
				// Delete the json string as we no longet need it
				delete(event, "json")

				// Now we can check the rule
				// First check if there's any sealighter specific
				// SIGMA rules (unlikley), then if not assume a process-creation
				// event
				rule, err = CheckRules(event, "sealighter")
				if err != nil {
					continue
				}
				if rule == nil {
					rule, err = CheckRules(event, "process_creation")
					if err != nil {
						continue
					}
				}
			} else {
				// Otherwise assume Sysmon
				channel = "sysmon"
				eventID, err := strconv.Atoi(j.Event.System.EventID)
				if err != nil {
					continue
				}
				switch eventID {
				case 1:
					// Process
					rule, err = CheckRules(event, "process_creation")
					if err != nil {
						return
					}
				default:
					// Unhandled event type
					continue
				}
			}

			// If a match, raise alert in Minecraft
			if rule != nil {
				// Remove fields we don't need to see in Minecraft
				delete(event, "EventID")
				delete(event, "Provider_Name")
				RaiseAlert(event, rule, channel)
			}
		}
		waitGroup.Done()
	}()

	return nil
}

func StopEventSubscription() {
	if eventProvider != nil {
		fmt.Println("[e] Stopping event subscription")
		eventProvider.Stop()
		waitGroup.Wait()
	}
}
