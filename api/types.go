package api

import (
	openai "github.com/sashabaranov/go-openai"
)

// Service represents your core API service
type Service struct {
	ServiceContext *ServiceContext
}

type ServiceContext struct {
	OpenaiModel  string
	OpenaiApiKey string
	OpenAIClient *openai.Client
	AssistantId  string
	ThreadId     string
	Namespace    string
	KubeConfig   string
}

type CliStatus struct {
	OpenaiApiKeyIsSet bool
	KubeClusterName   string
	KubeNamespaces    []string
	CurrentNamespace  string
}

type PromptResponse struct {
	InputPrompt string
	AIResponse  string
	Content     string
}
