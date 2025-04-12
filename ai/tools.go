package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
)

type ToolRequest struct {
	Cmd string `json:"cmd"`
}

var tools = []openai.ChatCompletionToolParam{
	{
		Function: openai.FunctionDefinitionParam{
			Name:        "checkCommand",
			Description: openai.String("Check if commant exists on user's machine. Return bool."),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"cmd": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"cmd"},
			},
		},
	},
	{
		Function: openai.FunctionDefinitionParam{
			Name:        "executeCommand",
			Description: openai.String("execute the command on user's machine. Return result or error."),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"cmd": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"cmd"},
			},
		},
	},
	{
		Function: openai.FunctionDefinitionParam{
			Name:        "getAvailableCommands",
			Description: openai.String("search commands that available on user's machine. Return commands or error"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"cmd": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"cmd"},
			},
		},
	},
}

func (c *AiClient) dealTool(toolCall openai.FinishedChatCompletionToolCall) openai.ChatCompletionMessageParamUnion {
	switch toolCall.Name {
	case `checkCommand`:
		var args ToolRequest
		err := json.Unmarshal([]byte(toolCall.Arguments), &args)
		if err != nil {
			return openai.ToolMessage(fmt.Sprintf("unmarshal param error: %v", err.Error()), toolCall.Id)
		}
		res := c.checkCommand(args.Cmd)
		return openai.ToolMessage(strconv.FormatBool(res), toolCall.Id)
	case `executeCommand`:
		var args ToolRequest
		err := json.Unmarshal([]byte(toolCall.Arguments), &args)
		if err != nil {
			return openai.ToolMessage(fmt.Sprintf("unmarshal param error: %v", err.Error()), toolCall.Id)
		}
		res, err := c.executeCommand(args.Cmd)
		if err != nil {
			res = fmt.Sprintf("error in executing executeCommand, %s", err.Error())
		}
		return openai.ToolMessage(res, toolCall.Id)
	case `getAvailableCommands`:
		var args ToolRequest
		err := json.Unmarshal([]byte(toolCall.Arguments), &args)
		if err != nil {
			return openai.ToolMessage(fmt.Sprintf("unmarshal param error: %v", err.Error()), toolCall.Id)
		}
		res := ""
		cmds, err := c.getAvailableCommands(args.Cmd)
		if err != nil {
			res = fmt.Sprintf("error in executing getAvailableCommands, %s", err.Error())
		} else {
			res = strings.Join(cmds, ",")
		}
		return openai.ToolMessage(res, toolCall.Id)
	default:
		return openai.ToolMessage(fmt.Sprintf("no tool named %s", toolCall.Name), toolCall.Id)
	}
}

// Tool function: Check if the command is available
func (c *AiClient) checkCommand(command string) bool {
	_, err := exec.LookPath(strings.Split(command, " ")[0])
	return err == nil
}

// Tool function: Execute the command and return the result
func (c *AiClient) executeCommand(command string) (string, error) {
	return c.Tc.ExecuteAndGetResult(command)
}

// Tool function: Get available commands
func (c *AiClient) getAvailableCommands(query string) ([]string, error) {
	// Get PATH environment variable
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, fmt.Errorf("PATH environment variable is empty")
	}

	// Split PATH into a list of directories
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))
	seen := make(map[string]struct{}) // Avoid duplicate commands
	var commands []string

	// Convert the query string to lowercase for case-insensitive matching
	query = strings.ToLower(query)

	// Iterate through each directory
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// Read directory contents
		files, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip inaccessible directories
		}

		// Check each file
		for _, file := range files {
			if file.IsDir() {
				continue // Skip subdirectories
			}

			// Get file information
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Check if it is executable
			if info.Mode()&0111 != 0 { // Check if it has execute permissions (rwx)
				name := file.Name()
				// Fuzzy matching: Check if the command name contains the query string (case-insensitive)
				if strings.Contains(strings.ToLower(name), query) {
					if _, exists := seen[name]; !exists {
						seen[name] = struct{}{}
						commands = append(commands, name)
					}
				}
			}
		}
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("no matching executable commands found for query: %s", query)
	}

	return commands, nil
}
