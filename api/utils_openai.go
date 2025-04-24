package api

import (
	"context"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

// getOpenAIClient returns an initialized OpenAI client
func getOpenAIClient(apiKey string) *openai.Client {
	if apiKey == "" {
		return nil
	}
	return openai.NewClient(apiKey)
}

// CreateOrGetAssistant ensures an OpenAI assistant exists for this application
func (s *Service) CreateOrGetAssistant() (string, error) {
	rtn := s.ServiceContext.AssistantId
	if rtn != "" && rtn != "auto" {
		return rtn, nil
	}
	return s.CreateNewAssistant()
}

// CreateNewAssistant creates a new OpenAI assistant with the kubectl_proxy tool
func (s *Service) CreateNewAssistant() (string, error) {
	client := getOpenAIClient(s.ServiceContext.OpenaiApiKey)
	if client == nil {
		return "", errors.New("OpenAI client not configured")
	}

	name := "Kube Copilot Assistant"
	description := "Kubernetes assistant to help manage and troubleshoot clusters"
	instructions := "You are a Kubernetes expert assistant. Use the kubectl_proxy and helm_proxy functions to make calls to kubectl or helm. When you call these functions, you MUST always provide a clear, human-readable description of what function or tool call you are performing in the 'description' parameter. If you get back [...truncated...] from a function call then you're missing important data. If there's a way to make a different function without getting truncated data, then do that. If that's not possible, then let the user know your working on truncated data."

	kubectlTool := openai.FunctionDefinition{
		Name:        "kubectl_proxy",
		Description: "Executes kubectl commands to interact with the Kubernetes cluster",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"args": map[string]interface{}{
					"type":        "array",
					"description": "The kubectl command arguments (e.g. ['get', 'pods'])",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "A description of what function/tool call is being performed.",
				},
			},
			"required": []string{"args", "description"},
		},
	}

	helmTool := openai.FunctionDefinition{
		Name:        "helm_proxy",
		Description: "Executes helm commands to interact with Helm releases and charts in the Kubernetes cluster",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"args": map[string]interface{}{
					"type":        "array",
					"description": "The helm command arguments (e.g. ['list', '-A'])",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "A description of what function/tool call is being performed.",
				},
			},
			"required": []string{"args", "description"},
		},
	}

	assistantRequest := openai.AssistantRequest{
		Name:         &name,
		Description:  &description,
		Model:        s.ServiceContext.OpenaiModel,
		Instructions: &instructions,
		Tools: []openai.AssistantTool{
			{
				Type:     openai.AssistantToolTypeFunction,
				Function: &kubectlTool,
			},
			{
				Type:     openai.AssistantToolTypeFunction,
				Function: &helmTool,
			},
		},
	}

	assistant, err := client.CreateAssistant(context.Background(), assistantRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create assistant: %w", err)
	}

	s.ServiceContext.AssistantId = assistant.ID
	viper.Set("ASSISTANT_ID", assistant.ID)
	err = viper.WriteConfig()
	if err != nil {
		// log warning but continue
		s.OutputWriter.Errorf("Warning: Failed to save assistant ID to config: %v\n", err)
	}

	return assistant.ID, nil
}

// ensureThreadID ensures a thread exists and returns its ID, updating the context if needed.
func (s *Service) ensureThreadID(client *openai.Client) (string, error) {
	threadId := s.ServiceContext.ThreadId
	if threadId != "" {
		return threadId, nil
	}
	threadObj, err := client.CreateThread(context.Background(), openai.ThreadRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to create thread: %w", err)
	}
	threadId = threadObj.ID
	s.ServiceContext.ThreadId = threadId
	if err := s.UpdateCurrentThreadID(threadId); err != nil {
		return "", fmt.Errorf("failed to update thread ID: %w", err)
	}
	return threadId, nil
}

// getLatestAssistantMessage fetches the latest assistant message text from a thread.
func getLatestAssistantMessage(client *openai.Client, threadId string) (string, error) {
	limit := 1
	order := "desc"
	messages, err := client.ListMessage(context.Background(), threadId, &limit, &order, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list messages: %w", err)
	}
	if len(messages.Messages) > 0 && len(messages.Messages[0].Content) > 0 {
		if messages.Messages[0].Content[0].Type == "text" {
			return messages.Messages[0].Content[0].Text.Value, nil
		}
	}
	return "", errors.New("no response received from assistant")
}
