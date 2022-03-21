package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bradleyjkemp/sigma-go"
	"github.com/bradleyjkemp/sigma-go/evaluator"
)

var RuleCategories []string = []string{"application", "security", "system", "process_creation", "file_create", "image_load", "driver_load", "network_connection", "dns", "registry_event", "sealighter"}
var Rules map[string][]sigma.Rule

func checkRule(event map[string]string, category string, rule sigma.Rule) (evaluator.Result, error) {
	// Sigma-Go uses runtime panics for things like "unsupported modifier re"
	defer func() { recover() }()

	ruleEvaluator := evaluator.ForRule(rule)
	result, err := ruleEvaluator.Matches(context.Background(), event)
	if err != nil {
		fmt.Printf("Failed to evaluate rule: %s", err.Error())
		return evaluator.Result{Match: false}, err
	}
	return result, nil
}

func CheckRules(event map[string]string, category string) (*sigma.Rule, error) {
	for _, rule := range Rules[category] {
		// Check rule, handle runtime panics from library
		result, err := checkRule(event, category, rule)
		if err != nil {
			return nil, err
		}

		if result.Match {
			return &rule, nil
		}
	}
	return nil, nil
}

func ParseRules(rulesDir string) error {
	fmt.Printf("[r] Parsing SIGMA rules from: %s\n", rulesDir)

	// Create global rules slice
	Rules = make(map[string][]sigma.Rule, len(RuleCategories))
	for _, category := range RuleCategories {
		Rules[category] = make([]sigma.Rule, 0, 10)
	}
	ruleCount := 0

	// Get rules from directory
	err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("Failed to open rule %s: %s\n", path, err.Error())
				return nil
			}
			rule, err := sigma.ParseRule(contents)
			if err != nil {
				fmt.Printf("Failed to parse rule %s: %s\n", path, err.Error())
				return nil
			}
			// Only Windows rules count
			if rule.Logsource.Product != "windows" {
				return nil
			}

			// Rules either have a Service or a Category
			ruleType := strings.TrimSpace(rule.Logsource.Service + rule.Logsource.Category)
			if ruleType == "dns_query" {
				ruleType = "dns"
			}
			switch ruleType {
			case
				// Services
				"application",
				"security",
				"system",
				// Categories
				"process_creation",
				"file_create",
				"image_load",
				"driver_load",
				"network_connection",
				"dns",
				"registry_event":
				//
				fmt.Printf("[r]    Found rule: %s\n", rule.Title)
				Rules[ruleType] = append(Rules[ruleType], rule)
				ruleCount++
			default:
				fmt.Printf("Ignoring Rule for bad service/category: %s\n", rule.Title)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Only run if we have at least one valid rule
	if ruleCount == 0 {
		errString := fmt.Sprintf("No valid rules found in %s", rulesDir)
		return errors.New(errString)
	}
	fmt.Printf("[r] Number of rules found: %d\n", ruleCount)
	return nil
}

// Used for Debugging only
func FireFakeEvents() {
	// First get a list of all the rules
	allRules := make([]sigma.Rule, 0)
	for _, category := range RuleCategories {
		for _, rule := range Rules[category] {
			// Overwrite category (which we don't use)
			// for our own purposes
			rule.Logsource.Category = category
			allRules = append(allRules, rule)
		}
	}

	// Get list of possible fake values
	selfImage, _ := os.Executable()
	imageNames := []string{
		"C:\\Windows\\System32\\whoami.exe",
		"C:\\Windows\\System32\\lsass.exe",
		"C:\\Windows\\System32\\svchost.exe",
		"C:\\Windows\\System32\\cmd.exe",
		"C:\\Windows\\System32\\powershell.exe",
		"C:\\Windows\\System32\\inetsrv\\w3wp.exe",
		"C:\\Python27\\python.exe",
		"C:\\dodgy\\mimikatz.exe",
		"C:\\bonza\\netcat.exe",
		"C\\Program Files (x86)\\Microsoft Office\\Office14\\WINWORD.EXE",
		selfImage,
	}
	commandLines := []string{
		"",
		"",
		"",
		"/c echo [*]& pwd & echo [*]",
		"-ep bypass -window hidden -enc 'AAA='",
		"dpapi::masterkey",
		"-k netsvcs -p -s Schedule",
		"-c 'print(\"hyper\")'",
		"-lvp 8000",
	}
	users := []string{
		"PATHDOWS\\PATH",
		"PATHDOWS\\DODGY",
		"NT AUTHORITY\\SYSTEM",
		"LocalService",
		"NetworkService",
		"Guest",
	}

	fmt.Printf("[f] Faking event generation\n")
	go func() {
		for {
			// Only fire events if there are players
			rand.Seed(time.Now().UnixNano())
			if len(MinecraftPlayers) != 0 {
				image := getRandom(imageNames)
				parentImage := getRandom(imageNames)
				event := map[string]string{
					"ProcessId":         "-1",
					"Image":             image,
					"CommandLine":       fmt.Sprintf("%s %s", image, getRandom(commandLines)),
					"User":              getRandom(users),
					"UtcTime":           time.Now().Format("2006-01-02 15:04:05.000"),
					"ParentProcessId":   "-1",
					"ParentImage":       parentImage,
					"ParentCommandLine": fmt.Sprintf("%s %s", parentImage, getRandom(commandLines)),
					"Computer":          "PATHDOWS",
				}
				rule := allRules[rand.Intn(len(allRules))]
				RaiseAlert(event, &rule, rule.Logsource.Category)
			}
			time.Sleep(time.Duration(rand.Intn(15)) * time.Second)
		}
	}()
}
