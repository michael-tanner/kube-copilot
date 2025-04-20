package api

import (
	"context"
	"errors"

	openai "github.com/openai/openai-go"
)

// SendOpenAIPrompt sends a prompt to OpenAI's ChatGPT-4 and returns the response.
func SendOpenAIPrompt(prompt string, openaiClient *openai.Client) (string, error) {
	if openaiClient == nil {
		return "", errors.New("OpenAI client is not initialized")
	}

	chatCompletion, err := openaiClient.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT4,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
	)

	if err != nil {
		return "", err
	}
	if len(chatCompletion.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
