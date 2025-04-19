package api

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"
)

// SendOpenAIPrompt sends a prompt to OpenAI's ChatGPT-4 and returns the response.
func SendOpenAIPrompt(prompt string, openaiClient *openai.Client) (string, error) {
	// Try to get the API key from environment if not set in config

	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}
