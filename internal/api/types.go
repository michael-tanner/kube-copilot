package api

type Status struct {
	OpenaiApiKeyIsSet bool
	KubeClusterName   string
	KubeNamespaces    []string
	CurrentNamespace  string
}

type PromptResponse struct {
	InputPrompt string
}