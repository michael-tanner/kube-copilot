package api

import openai "github.com/sashabaranov/go-openai"

// Service represents your core API service
type Service struct {
	OpenAIClient *openai.Client
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
