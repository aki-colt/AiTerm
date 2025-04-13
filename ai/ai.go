package ai

import (
	"context"
	"fmt"

	"github.com/aki-colt/aiterm/terminal"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rivo/tview"
)

type AiConfig struct {
	URL   string
	Token string
	Model string
}

func (c AiConfig) Validate() error {
	client := openai.NewClient(
		option.WithBaseURL(c.URL),
		option.WithAPIKey(c.Token),
	)

	// 尝试列出模型，验证配置
	_, err := client.Models.List(context.Background())
	if err != nil {
		return fmt.Errorf("invalid AI configuration: %w", err)
	}

	// 验证指定模型是否存在（可选）
	models, err := client.Models.List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}
	for _, m := range models.Data {
		if m.ID == c.Model {
			return nil
		}
	}
	return fmt.Errorf("model %s not found", c.Model)
}

type AiClient struct {
	client     openai.Client
	params     openai.ChatCompletionNewParams
	Tc         *terminal.TerminalController
	DialogView *tview.TextView
}

func Init(tc *terminal.TerminalController, dv *tview.TextView, cfg AiConfig) *AiClient {
	return &AiClient{
		client: openai.NewClient(
			option.WithBaseURL(cfg.URL),
			option.WithAPIKey(cfg.Token),
		),
		params: openai.ChatCompletionNewParams{
			Model:    cfg.Model,
			Messages: []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(prompt)},
			Seed:     openai.Int(0),
			Tools:    tools,
		},
		Tc:         tc,
		DialogView: dv,
	}
}

func (c *AiClient) Run(ctx context.Context, input string, app *tview.Application) error {
	if input != "" {
		c.params.Messages = append(c.params.Messages, openai.UserMessage(input))
	}
	stream := c.client.Chat.Completions.NewStreaming(ctx, c.params)
	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		if content, ok := acc.JustFinishedContent(); ok {
			c.params.Messages = append(c.params.Messages, openai.AssistantMessage(content))
		}

		// if using tool calls
		if tool, ok := acc.JustFinishedToolCall(); ok {
			assistantMsg := acc.Choices[0].Message
			if len(assistantMsg.ToolCalls) > 0 {
				c.params.Messages = append(c.params.Messages, assistantMsg.ToParam())
			}

			res := c.dealTool(tool)
			c.params.Messages = append(c.params.Messages, res)

			c.Run(ctx, "", app)
		}

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			println("Refusal stream finished:", refusal)
		}

		if len(chunk.Choices) > 0 {
			app.QueueUpdateDraw(func() {
				fmt.Fprint(c.DialogView, chunk.Choices[0].Delta.Content)
				c.DialogView.ScrollToEnd()
			})
		}
	}

	if stream.Err() != nil {
		return stream.Err()
	}
	return nil
}
