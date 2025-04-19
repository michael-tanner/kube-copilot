package api

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Service represents your core API service
type Service struct {
	// add dependencies here (e.g., logger, kubeClient) if needed
	OpenAIClient *openai.Client
}

// NewService creates a new API service instance
func NewService() *Service {
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	var openaiClient *openai.Client
	if openaiApiKey != "" {
		openaiClient = openai.NewClient(openaiApiKey)
	}
	return &Service{
		OpenAIClient: openaiClient,
	}
}

// CheckStatus reads config/context and returns status info
func (s *Service) CheckStatus() (*Status, error) {
	openaiKey := viper.GetString("OPENAI_API_KEY")
	namespace := viper.GetString("namespace")

	// Get namespaces and sort them
	nsList, err := s.GetKubeNamespaces()
	if err != nil {
		return nil, err
	}
	// Sort namespaces alphabetically
	if len(nsList) > 1 {
		sort.Strings(nsList)
	}

	// Get current cluster name from kubeconfig (if available)
	clusterName := ""
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err == nil && config != nil && config.CurrentContext != "" {
		ctx := config.Contexts[config.CurrentContext]
		if ctx != nil {
			clusterName = ctx.Cluster
		}
	}

	status := &Status{
		OpenaiApiKeyIsSet: openaiKey != "",
		KubeClusterName:   clusterName,
		KubeNamespaces:    nsList,
		CurrentNamespace:  namespace,
	}

	return status, nil
}

// GetKubeNamespaces returns a list of namespaces in the current Kubernetes cluster.
func (s *Service) GetKubeNamespaces() ([]string, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig file
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}
	return nsNames, nil
}

// SendPrompt sends a prompt to the AI and returns a PromptResponse.
// For now, this is a stub that just echoes the input prompt.
func (s *Service) SendPrompt(prompt string) (*PromptResponse, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return nil, errors.New("prompt must not be empty")
	}
	return &PromptResponse{
		InputPrompt: trimmed,
	}, nil
}
