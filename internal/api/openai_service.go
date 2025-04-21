package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func SendOpenAIPrompt(prompt string, client *openai.Client) (string, error) {
	if client == nil {
		return "", errors.New("OpenAI client is not initialized")
	}

	// Initial prompt to ChatGPT
	chatCompletion, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Functions: []openai.FunctionDefinition{
				{
					Name:        "KubectlProxy",
					Description: "Executes kubectl-like commands using the Go client",
					Parameters: json.RawMessage(`{
						"type": "object",
						"properties": {
							"args": {
								"type": "array",
								"items": {
									"type": "string"
								}
							}
						},
						"required": ["args"]
					}`),
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	fmt.Printf("Thread ID: %s\n", chatCompletion.ID)

	if len(chatCompletion.Choices) > 0 && chatCompletion.Choices[0].Message.FunctionCall != nil {
		functionCall := chatCompletion.Choices[0].Message.FunctionCall
		if functionCall.Name == "KubectlProxy" {
			var argsPayload struct {
				Args []string `json:"args"`
			}
			if err := json.Unmarshal([]byte(functionCall.Arguments), &argsPayload); err != nil {
				return "", err
			}

			service := NewService()
			kubectlOutput, err := service.KubectlProxy(argsPayload.Args)
			if err != nil {
				return "", err
			}

			functionResponse, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT4,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleFunction,
							Name:    "KubectlProxy",
							Content: fmt.Sprintf(`{"output": %q}`, kubectlOutput),
						},
					},
				},
			)
			if err != nil {
				return "", err
			}

			if len(functionResponse.Choices) > 0 {
				return functionResponse.Choices[0].Message.Content, nil
			}
		}
	}

	if len(chatCompletion.Choices) > 0 {
		return chatCompletion.Choices[0].Message.Content, nil
	}

	return "", errors.New("no response from OpenAI")
}
