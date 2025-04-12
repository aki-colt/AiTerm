package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TerminalController manages the tmux right pane
type TerminalController struct {
	session    string // tmux session name
	pane       string // right pane ID
	outputChan chan string
	mutex      sync.Mutex
	running    bool
}

func NewTerminalController() *TerminalController {
	tc := &TerminalController{
		outputChan: make(chan string, 100),
		running:    true,
	}

	// Get the current tmux session
	session, err := getCurrentTmuxSession()
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		os.Exit(1)
	}
	tc.session = session

	// Initialize tmux pane
	tc.setupTmux()

	return tc
}

func (tc *TerminalController) setupTmux() {
	// Split the current window into left and right panes
	cmd := exec.Command("tmux", "split-window", "-h", "-d", "-P", "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error splitting tmux window:", err)
		os.Exit(1)
	}

	// Get the right pane ID
	tc.pane = strings.TrimSpace(string(output))
}

// getCurrentTmuxSession gets the current tmux session name
func getCurrentTmuxSession() (string, error) {
	// Check if in tmux
	if os.Getenv("TMUX") == "" {
		return "", fmt.Errorf("not in tmux, please install tmux and run in a tmux session")
	}

	// Use display-message to get the current session name
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	session := strings.TrimSpace(string(output))
	if session == "" {
		return "", fmt.Errorf("failed to get session name")
	}
	return session, nil
}

// tmuxCommand executes a tmux command
func (tc *TerminalController) tmuxCommand(args ...string) error {
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// WriteCommand writes a command to the right pane
func (tc *TerminalController) WriteCommand(command string) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	if !tc.running {
		return os.ErrClosed
	}
	err := tc.tmuxCommand("send-keys", "-t", tc.pane, command, "Enter")
	if err != nil {
		fmt.Println("WriteCommand error:", err)
	}
	return err
}

// capturePane captures the output of the right pane
func (tc *TerminalController) capturePane() (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", tc.pane, "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ReadOutput reads output from the right pane
func (tc *TerminalController) ReadOutput() <-chan string {
	return tc.outputChan
}

// Stop cleans up tmux configuration
func (tc *TerminalController) Stop() {
	tc.mutex.Lock()
	tc.running = false
	tc.mutex.Unlock()
	tc.tmuxCommand("kill-pane", "-t", tc.pane) // Close the right pane
}

// ExecuteAndGetResult clears the screen, executes a command, and immediately gets the result
func (tc *TerminalController) ExecuteAndGetResult(command string) (string, error) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	if !tc.running {
		return "", os.ErrClosed
	}

	// Clear the screen
	err := tc.tmuxCommand("send-keys", "-t", tc.pane, "clear", "Enter")
	if err != nil {
		return "", fmt.Errorf("failed to clear screen: %v", err)
	}

	// Wait for the screen to clear
	time.Sleep(100 * time.Millisecond)

	// Execute the command
	err = tc.tmuxCommand("send-keys", "-t", tc.pane, command, "Enter")
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	// Wait for the command to execute
	time.Sleep(100 * time.Millisecond)

	// Capture the output
	cmd := exec.Command("tmux", "capture-pane", "-t", tc.pane, "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture output: %v", err)
	}

	// Return the cleaned output
	return strings.TrimSpace(string(output)), nil
}
