package api

import (
	"context"
	"errors"

	openai "github.com/openai/openai-go"
)

func SendOpenAIPrompt(prompt string, client *openai.Client) (string, error) {
	if client == nil {
		return "", errors.New("OpenAI client is not initialized")
	}

	chatCompletion, err := client.Chat.Completions.New(
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
