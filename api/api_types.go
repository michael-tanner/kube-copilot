package api

import (
	openai "github.com/sashabaranov/go-openai"
)

// Service represents your core API service
type Service struct {
	ServiceContext *ServiceContext
	OutputWriter   OutputWriter
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

// OutputWriter defines methods for writing output and errors.
type OutputWriter interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Errorf(format string, args ...interface{})
}

// NoopOutputWriter is an OutputWriter that does nothing.
type NoopOutputWriter struct{}

func (n *NoopOutputWriter) Print(args ...interface{})                 {}
func (n *NoopOutputWriter) Printf(format string, args ...interface{}) {}
func (n *NoopOutputWriter) Println(args ...interface{})               {}
func (n *NoopOutputWriter) Errorf(format string, args ...interface{}) {}
