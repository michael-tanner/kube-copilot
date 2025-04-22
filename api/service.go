package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/michael-tanner/kube-copilot/api/config"
)

func init() {
	viper.AutomaticEnv()
	config.SetupViperConfig()
	_ = viper.ReadInConfig() // ignore error if config file not found
}

// NewService creates a new API service instance
func NewService() *Service {
	rtn := &Service{
		ServiceContext: &ServiceContext{
			OpenaiModel:  getValueOrDefault("OPENAI_MODEL", openai.GPT4, true),
			OpenaiApiKey: getValueOrDefault("OPENAI_API_KEY", ""),
			OpenAIClient: nil,
			AssistantId:  getValueOrDefault("ASSISTANT_ID", "auto", true),
			ThreadId:     getValueOrDefault("THREAD_ID", "", true),
			Namespace:    getValueOrDefault("NAMESPACE", "default", true),
			KubeConfig:   getValueOrDefault("KUBECONFIG", clientcmd.RecommendedHomeFile),
		},
	}
	if rtn.ServiceContext.OpenaiApiKey != "" {
		rtn.ServiceContext.OpenAIClient = openai.NewClient(rtn.ServiceContext.OpenaiApiKey)
	}
	return rtn
}

func getValueOrDefault(key string, defaultValue string, setIfNotSet ...bool) string {
	value := viper.GetString(key)
	if value == "" {
		value = os.Getenv(key)
	}
	if value == "" {
		value = defaultValue
		if len(setIfNotSet) > 0 && setIfNotSet[0] && defaultValue != "" {
			viper.Set(key, defaultValue)
			err := viper.WriteConfig()
			if err != nil {
				fmt.Printf("Error writing config: %v\n", err)
			}
		}
	}
	return value
}

// CheckStatus reads config/context and returns status info
func (s *Service) CheckStatus() (*CliStatus, error) {
	namespace := s.ServiceContext.Namespace

	// Get namespaces and sort them
	nsList, err := s.GetKubeNamespaces()
	if err != nil {
		return nil, err
	}
	// Sort namespaces alphabetically
	if len(nsList) > 1 {
		sort.Strings(nsList)
	}

	clusterName := ""
	config, err := clientcmd.LoadFromFile(s.ServiceContext.KubeConfig)
	if err == nil && config != nil && config.CurrentContext != "" {
		ctx := config.Contexts[config.CurrentContext]
		if ctx != nil {
			clusterName = ctx.Cluster
		}
	}

	status := &CliStatus{
		OpenaiApiKeyIsSet: s.ServiceContext.OpenaiApiKey != "",
		KubeClusterName:   clusterName,
		KubeNamespaces:    nsList,
		CurrentNamespace:  namespace,
	}

	return status, nil
}

// getKubeClientset returns a Kubernetes clientset.
func (s *Service) getKubeClientset() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", s.ServiceContext.KubeConfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// GetKubeNamespaces returns a list of namespaces in the current Kubernetes cluster.
func (s *Service) GetKubeNamespaces() ([]string, error) {
	clientset, err := s.getKubeClientset()
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

// KubectlProxy executes a kubectl-like command using the Go client.
// Supported: get pods, get services, get deployments, get nodes, get namespaces
func (s *Service) KubectlProxy(args []string) ([]string, error) {
	if len(args) < 1 {
		return nil, errors.New("no kubectl command provided")
	}

	cmdArgs := append([]string{}, args...)
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+s.ServiceContext.KubeConfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{string(output)}, err
	}
	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	return lines, nil
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
	client := s.getOpenAIClient()
	if client == nil {
		return "", errors.New("OpenAI client not configured")
	}

	name := "Kube Copilot Assistant"
	description := "Kubernetes assistant to help manage and troubleshoot clusters"
	instructions := "You are a Kubernetes expert assistant. Use the kubectl_proxy function to make calls to kubectl."

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
			},
			"required": []string{"args"},
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
		fmt.Printf("Warning: Failed to save assistant ID to config: %v\n", err)
	}

	return assistant.ID, nil
}

// getOpenAIClient returns an initialized OpenAI client
func (s *Service) getOpenAIClient() *openai.Client {
	apiKey := s.ServiceContext.OpenaiApiKey
	if apiKey == "" {
		return nil
	}
	return openai.NewClient(apiKey)
}

// SendPrompt sends a prompt to OpenAI and returns the response, handling kubectl_proxy tool calls
func (s *Service) SendPrompt(prompt string) (*PromptResponse, error) {
	apiKey := s.ServiceContext.OpenaiApiKey
	if apiKey == "" {
		return nil, errors.New("OpenAI API key not set. Please configure with the 'config openai' command")
	}
	client := openai.NewClient(apiKey)

	assistantID, err := s.CreateOrGetAssistant()
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}

	threadId := s.ServiceContext.ThreadId
	if threadId == "" {
		threadObj, err := client.CreateThread(context.Background(), openai.ThreadRequest{})
		if err != nil {
			return nil, fmt.Errorf("failed to create thread: %w", err)
		}
		threadId = threadObj.ID
		s.ServiceContext.ThreadId = threadId
		err = s.UpdateCurrentThreadID(threadId)
		if err != nil {
			return nil, fmt.Errorf("failed to update thread ID: %w", err)
		}
	}

	_, err = client.CreateMessage(context.Background(), threadId, openai.MessageRequest{
		Role:    string(openai.ThreadMessageRoleUser),
		Content: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	runRequest := openai.RunRequest{
		AssistantID: assistantID,
	}
	run, err := client.CreateRun(context.Background(), threadId, runRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	for run.Status != openai.RunStatusCompleted {
		run, err = client.RetrieveRun(context.Background(), threadId, run.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve run: %w", err)
		}
		if run.Status == openai.RunStatusFailed {
			return nil, fmt.Errorf("run failed: %s", run.LastError.Message)
		}
		if run.Status == openai.RunStatusExpired || run.Status == openai.RunStatusCancelled {
			return nil, fmt.Errorf("run ended with status: %s", run.Status)
		}
		// Handle tool calls
		if run.Status == openai.RunStatusRequiresAction &&
			run.RequiredAction != nil &&
			run.RequiredAction.Type == "submit_tool_outputs" {
			var toolOutputs []openai.ToolOutput
			for _, toolCall := range run.RequiredAction.SubmitToolOutputs.ToolCalls {
				if toolCall.Function.Name == "kubectl_proxy" {
					var args struct {
						Args []string `json:"args"`
					}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						toolOutputs = append(toolOutputs, openai.ToolOutput{
							ToolCallID: toolCall.ID,
							Output:     fmt.Sprintf("Failed to parse arguments: %v", err),
						})
						continue
					}
					out, err := s.KubectlProxy(args.Args)
					outStr := strings.Join(out, "\n")
					if err != nil {
						outStr = fmt.Sprintf("Error: %v\nOutput: %s", err, outStr)
					}
					toolOutputs = append(toolOutputs, openai.ToolOutput{
						ToolCallID: toolCall.ID,
						Output:     outStr,
					})
				}
			}
			run, err = client.SubmitToolOutputs(
				context.Background(),
				threadId,
				run.ID,
				openai.SubmitToolOutputsRequest{
					ToolOutputs: toolOutputs,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("failed to submit tool outputs: %w", err)
			}
		}
		if run.Status == openai.RunStatusInProgress || run.Status == openai.RunStatusQueued {
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
		}
	}
	fmt.Println()

	limit := 1
	order := "desc"
	messages, err := client.ListMessage(context.Background(), threadId, &limit, &order, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	if len(messages.Messages) > 0 && len(messages.Messages[0].Content) > 0 {
		if messages.Messages[0].Content[0].Type == "text" {
			return &PromptResponse{
				InputPrompt: prompt,
				Content:     messages.Messages[0].Content[0].Text.Value,
			}, nil
		}
	}
	return nil, errors.New("no response received from assistant")
}

func (s *Service) UpdateCurrentThreadID(threadId string) error {
	viper.Set("thread_id", threadId)
	s.ServiceContext.ThreadId = threadId

	// Write the updated config back to the file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	return nil
}
