package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
func NewService(writers ...OutputWriter) *Service {
	var writer OutputWriter
	if len(writers) > 0 && writers[0] != nil {
		writer = writers[0]
	} else {
		writer = &NoopOutputWriter{}
	}
	s := &Service{
		OutputWriter: writer,
	}
	s.ServiceContext = &ServiceContext{
		OpenaiModel:  s.getValueOrDefault("OPENAI_MODEL", openai.GPT4, true),
		OpenaiApiKey: s.getValueOrDefault("OPENAI_API_KEY", ""),
		OpenAIClient: nil,
		AssistantId:  s.getValueOrDefault("ASSISTANT_ID", "auto", true),
		ThreadId:     s.getValueOrDefault("THREAD_ID", "", true),
		Namespace:    s.getValueOrDefault("NAMESPACE", "default", true),
		KubeConfig:   s.getValueOrDefault("KUBECONFIG", clientcmd.RecommendedHomeFile),
	}
	if s.ServiceContext.OpenaiApiKey != "" {
		s.ServiceContext.OpenAIClient = openai.NewClient(s.ServiceContext.OpenaiApiKey)
	}
	return s
}

// getValueOrDefault is now a method so it can use s.OutputWriter
func (s *Service) getValueOrDefault(key string, defaultValue string, setIfNotSet ...bool) string {
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
				s.OutputWriter.Errorf("Error writing config: %v\n", err)
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
	instructions := "You are a Kubernetes expert assistant. Use the kubectl_proxy function to make calls to kubectl. When you call kubectl_proxy, you MUST always provide a clear, human-readable description of what function or tool call you are performing in the 'description' parameter. This description will be shown to the user prefixed with 'AI Function: '."

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
		s.OutputWriter.Errorf("Warning: Failed to save assistant ID to config: %v\n", err)
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

func (s *Service) SendPrompt(prompt string) (*PromptResponse, error) {
	cliDebug := viper.GetBool("cli_debug")
	var logFilePath string

	// Generate a sortable datetime log filename
	now := time.Now().UTC()
	logFileName := now.Format("20060102T150405.000000") + ".log"
	logDir := ".kubecopilot/logs"
	logFilePath = filepath.Join(logDir, logFileName)

	// Ensure log directory exists
	_ = os.MkdirAll(logDir, 0755)

	// Open the log file for appending
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.OutputWriter.Errorf("Failed to create log file: %v\n", err)
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Helper functions to write logs progressively
	logPrintf := func(format string, args ...interface{}) {
		logLine := fmt.Sprintf(format, args...)
		logFile.WriteString(logLine)
		logFile.WriteString("\n")
		if cliDebug {
			s.OutputWriter.Printf(format, args...)
		}
	}
	logPrintln := func(args ...interface{}) {
		logLine := fmt.Sprintln(args...)
		logFile.WriteString(logLine)
		if cliDebug {
			s.OutputWriter.Println(args...)
		}
	}
	logErrorf := func(format string, args ...interface{}) {
		logLine := fmt.Sprintf(format, args...)
		logFile.WriteString(logLine)
		logFile.WriteString("\n")
		if cliDebug {
			s.OutputWriter.Errorf(format, args...)
		}
	}

	apiKey := s.ServiceContext.OpenaiApiKey
	if apiKey == "" {
		return nil, errors.New("OpenAI API key not set. Please configure with the 'config openai' command")
	}
	client := openai.NewClient(apiKey)

	assistantID, err := s.CreateOrGetAssistant()
	if err != nil {
		logErrorf("[DEBUG] failed to get assistant: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}
	logPrintf("[DEBUG] Using AssistantID: %s\n", assistantID)

	threadId, err := s.ensureThreadID(client)
	if err != nil {
		logErrorf("[DEBUG] failed to ensure thread ID: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, err
	}
	logPrintf("[DEBUG] Using ThreadID: %s\n", threadId)

	logPrintf("[DEBUG] Sending prompt: %q\n", prompt)
	_, err = client.CreateMessage(context.Background(), threadId, openai.MessageRequest{
		Role:    string(openai.ThreadMessageRoleUser),
		Content: prompt,
	})
	if err != nil {
		logErrorf("[DEBUG] CreateMessage error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	runRequest := openai.RunRequest{
		AssistantID: assistantID,
	}
	logPrintf("[DEBUG] Creating run with request: %+v\n", runRequest)
	run, err := client.CreateRun(context.Background(), threadId, runRequest)
	if err != nil {
		logErrorf("[DEBUG] CreateRun error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	for run.Status != openai.RunStatusCompleted {
		logPrintf("[DEBUG] Run status: %s\n", run.Status)
		run, err = client.RetrieveRun(context.Background(), threadId, run.ID)
		if err != nil {
			logErrorf("[DEBUG] RetrieveRun error: %v\n", err)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("failed to retrieve run: %w", err)
		}
		if run.Status == openai.RunStatusFailed {
			logErrorf("[DEBUG] Run failed: %v\n", run.LastError)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("run failed: %s", run.LastError.Message)
		}
		if run.Status == openai.RunStatusExpired || run.Status == openai.RunStatusCancelled {
			logErrorf("[DEBUG] Run ended with status: %s\n", run.Status)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("run ended with status: %s", run.Status)
		}
		// Handle tool calls
		if run.Status == openai.RunStatusRequiresAction &&
			run.RequiredAction != nil &&
			run.RequiredAction.Type == "submit_tool_outputs" {
			logPrintf("[DEBUG] Run requires action: %+v\n", run.RequiredAction)
			var toolOutputs []openai.ToolOutput
			for _, toolCall := range run.RequiredAction.SubmitToolOutputs.ToolCalls {
				logPrintf("[DEBUG] ToolCall: %+v\n", toolCall)
				if toolCall.Function.Name == "kubectl_proxy" {
					var args struct {
						Args        []string `json:"args"`
						Description string   `json:"description"`
					}
					logPrintf("[DEBUG] ToolCall.Function.Arguments: %s\n", toolCall.Function.Arguments)
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logErrorf("[DEBUG] Failed to parse tool arguments: %v\n", err)
						toolOutputs = append(toolOutputs, openai.ToolOutput{
							ToolCallID: toolCall.ID,
							Output:     fmt.Sprintf("Failed to parse arguments: %v", err),
						})
						continue
					}
					logPrintf("[DEBUG] Parsed tool args: %+v\n", args)
					logPrintln("\nAI Function: " + args.Description)
					out, err := s.KubectlProxy(args.Args)
					outStr := strings.Join(out, "\n")

					// Sanitize and format the output
					outStr = sanitizeOutput(outStr)

					// Ensure the output is properly formatted and truncated if necessary
					const maxOutputLen = 4000
					if len(outStr) > maxOutputLen {
						outStr = outStr[:maxOutputLen] + "\n[...truncated...]"
					}
					// Wrap the output in a code block for clarity
					if outStr != "" && !strings.HasPrefix(outStr, "```") {
						outStr = "```\n" + outStr + "\n```"
					}
					if err != nil {
						outStr = fmt.Sprintf("Error: %v\nOutput:\n%s", err, outStr)
					}
					logPrintf("[DEBUG] Tool output length: %d\n", len(outStr))
					logPrintf("[DEBUG] Tool output preview:\n%s\n", outStr[:min(500, len(outStr))])
					toolOutputs = append(toolOutputs, openai.ToolOutput{
						ToolCallID: toolCall.ID,
						Output:     outStr,
					})
				}
			}
			logPrintf("[DEBUG] Submitting tool outputs: %+v\n", toolOutputs)
			run, err = client.SubmitToolOutputs(
				context.Background(),
				threadId,
				run.ID,
				openai.SubmitToolOutputsRequest{
					ToolOutputs: toolOutputs,
				},
			)
			if err != nil {
				logErrorf("[DEBUG] SubmitToolOutputs error: %v\n", err)
				s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
				return nil, fmt.Errorf("failed to submit tool outputs: %w", err)
			}
		}
		if run.Status == openai.RunStatusInProgress || run.Status == openai.RunStatusQueued {
			// Ensure period is written immediately
			if f, ok := s.OutputWriter.(interface{ Flush() }); ok {
				s.OutputWriter.Print(".")
				f.Flush()
			} else {
				s.OutputWriter.Print(".")
			}
			// Also flush stdout for non-cobra writers
			if w, ok := s.OutputWriter.(*NoopOutputWriter); ok {
				// do nothing
				_ = w
			} else {
				// Try to flush os.Stdout if possible
				os.Stdout.Sync()
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	s.OutputWriter.Println("")

	content, err := getLatestAssistantMessage(client, threadId)
	if err != nil {
		logErrorf("[DEBUG] getLatestAssistantMessage error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, err
	}
	logPrintf("[DEBUG] Assistant response: %q\n", content)

	return &PromptResponse{
		InputPrompt: prompt,
		Content:     content,
	}, nil
}

// Helper function to sanitize output
func sanitizeOutput(output string) string {
	// Remove excessive newlines and trim spaces
	output = strings.TrimSpace(output)
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")
	// Replace multiple newlines with a single newline
	output = strings.ReplaceAll(output, "\n\n", "\n")
	return output
}

// Helper to save debug log to file
func saveDebugLog(path string, buf *bytes.Buffer) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(buf.Bytes())
	return err
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
