package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/colt-halden/aiterm/ai"
	"github.com/colt-halden/aiterm/terminal"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/ini.v1"
)

var version = "v0.01.0"

func main() {
	showVersion()
	startApp()
}

func showVersion() {
	showVersion := flag.Bool("v", false, "Show version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("ai-terminal version", version)
		os.Exit(0)
	}
}

func startApp() {
	// Check and get config first
	cfg := checkAIConfig()

	// Start app
	app := tview.NewApplication()

	dialogView, dialogInput := terminal.NewDialogComponents()

	tc := terminal.NewTerminalController()

	dialogView.SetText(dialogView.GetText(true) + "\nAI: " + "Tell me what you want to do and I will execute the cmd on the right pane.")

	aiClient := ai.Init(tc, dialogView, cfg)

	generating := false
	var currentCancel context.CancelFunc

	dialogInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			input := dialogInput.GetText()
			if input != "" {
				generating = true
				fmt.Fprint(dialogView, "\n\n[green]You: "+input+"[-]\n\nAI: ")
				dialogView.ScrollToEnd()
				dialogInput.SetDisabled(true)
				dialogInput.SetText("generating.")
				ctx, cancel := context.WithCancel(context.Background())
				currentCancel = cancel
				go func() {
					defer func() {
						if e := recover(); e != nil {
							app.QueueUpdateDraw(func() {
								fmt.Fprintf(dialogView, "/n[red]%v[-]", e)
							})
						}
						currentCancel()
					}()
					err := aiClient.Run(ctx, input, app)
					if err != nil {
						app.QueueUpdateDraw(func() {
							fmt.Fprint(dialogView, "/n[red]"+err.Error()+"[-]")
						})
					}
					generating = false
				}()
				// Animation effect
				go generatingAnime(ctx, dialogInput, app)
			}
		}
	})

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dialogView, 0, 1, false).
		AddItem(dialogInput, 1, 0, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			if generating {
				generating = false
				currentCancel()
				return nil // 消费事件
			} else {
				tc.Stop()
				app.Stop()
				return nil
			}
		case tcell.KeyUp:
			app.SetFocus(dialogView)
		case tcell.KeyDown:
			app.SetFocus(dialogInput)
		}
		return event
	})
	if err := app.SetRoot(mainFlex, true).SetFocus(dialogInput).Run(); err != nil {
		panic(err)
	}
}

func checkAIConfig() ai.AiConfig {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %s\n", err.Error())
		os.Exit(1)
	}
	configPath := filepath.Join(homeDir, ".aitermrc")

	// Detect and load configuration
	config, err := loadConfig(configPath)
	if err != nil || config.URL == "" || config.Token == "" || config.Model == "" {
		// Configuration missing, start configuration UI and get config from user
		config = configureAI()
		// check if config valid
		if err := config.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error AI config: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("Your config will be saved in ~/.aitermrc. Edit it if you want change provider information.")
		// Save configuration
		if err := saveConfig(configPath, config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %s\n", err.Error())
			os.Exit(1)
		}
	}
	return config
}

// loadConfig load config from .aitermrc
func loadConfig(path string) (ai.AiConfig, error) {
	config := ai.AiConfig{}
	cfg, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // File does not exist, return empty config
		}
		return config, err
	}

	section := cfg.Section("ai")
	config.URL = section.Key("url").String()
	config.Token = section.Key("token").String()
	config.Model = section.Key("model").String()
	return config, nil
}

// saveConfig save config to .aitermrc
func saveConfig(path string, config ai.AiConfig) error {
	cfg := ini.Empty()
	section, err := cfg.NewSection("ai")
	if err != nil {
		return err
	}
	section.NewKey("url", config.URL)
	section.NewKey("token", config.Token)
	section.NewKey("model", config.Model)
	return cfg.SaveTo(path)
}

// configureAI get config from user
func configureAI() ai.AiConfig {
	config := ai.AiConfig{}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Your AI configuration was not found in the configuration file. Please enter the information for an OpenAI-compatible AI provider.")
	fmt.Print("Enter Base URL: ")
	if scanner.Scan() {
		config.URL = strings.TrimSpace(scanner.Text())
	}

	fmt.Print("Enter API Token: ")
	if scanner.Scan() {
		config.Token = strings.TrimSpace(scanner.Text())
	}

	fmt.Print("Enter Model Name: ")
	if scanner.Scan() {
		config.Model = strings.TrimSpace(scanner.Text())
	}

	return config
}

// show generating anime
func generatingAnime(ctx context.Context, dialogInput *tview.InputField, app *tview.Application) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	states := []string{".", "..", "..."}
	i := 0
	for {
		select {
		case <-ticker.C:
			app.QueueUpdateDraw(func() {
				dialogInput.SetText("generating" + states[i%3])
			})
			i++
		case <-ctx.Done():
			app.QueueUpdateDraw(func() {
				dialogInput.SetText("")
				dialogInput.SetDisabled(false)
			})
			return
		}
	}
}
