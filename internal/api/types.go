package api

import (
	openai "github.com/openai/openai-go"
)

// Service represents your core API service
type Service struct {
	OpenAIClient *openai.Client
	OpenaiApiKey string
}

type Status struct {
	OpenaiApiKeyIsSet bool
	KubeClusterName   string
	KubeNamespaces    []string
	CurrentNamespace  string
}

type PromptResponse struct {
	InputPrompt string
	AIResponse  string
}
