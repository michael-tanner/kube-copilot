/*
This file (api_main.go) is the main entry point for the API service.
It initializes the service, sets up the OpenAI client, and handles
sending prompts to the OpenAI API.
The service context is initialized with values from the environment
variables or defaults.

All api structs must be defined in api_types.go
The service context is initialized with values from the environment
variables or defaults.

*/

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
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
		OpenaiModel:  s.getValueOrDefault("OPENAI_MODEL", "gpt-4.1-mini", true),
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


func (s *Service) SendPrompt(prompt string) (*PromptResponse, error) {
	cliDebug := viper.GetBool("cli_debug")

	// Generate log file path
	logFilePath, err := generateLogFilePath()
	if err != nil {
		s.OutputWriter.Errorf("Failed to generate log file path: %v\n", err)
		return nil, err
	}

	// Open the log file for appending
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.OutputWriter.Errorf("Failed to create log file: %v\n", err)
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Use log helpers from utils.go
	logPrintf(logFile, cliDebug, "[DEBUG] Sending prompt: %q\n", prompt)

	apiKey := s.ServiceContext.OpenaiApiKey
	if apiKey == "" {
		return nil, errors.New("OpenAI API key not set. Please configure with the 'config openai' command")
	}
	client := openai.NewClient(apiKey)

	assistantID, err := s.CreateOrGetAssistant()
	if err != nil {
		logErrorf(logFile, cliDebug, "[DEBUG] failed to get assistant: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}
	logPrintf(logFile, cliDebug, "[DEBUG] Using AssistantID: %s\n", assistantID)

	threadId, err := s.ensureThreadID(client)
	if err != nil {
		logErrorf(logFile, cliDebug, "[DEBUG] failed to ensure thread ID: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, err
	}
	logPrintf(logFile, cliDebug, "[DEBUG] Using ThreadID: %s\n", threadId)

	logPrintf(logFile, cliDebug, "[DEBUG] Sending prompt: %q\n", prompt)
	_, err = client.CreateMessage(context.Background(), threadId, openai.MessageRequest{
		Role:    string(openai.ThreadMessageRoleUser),
		Content: prompt,
	})
	if err != nil {
		logErrorf(logFile, cliDebug, "[DEBUG] CreateMessage error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	runRequest := openai.RunRequest{
		AssistantID: assistantID,
	}
	logPrintf(logFile, cliDebug, "[DEBUG] Creating run with request: %+v\n", runRequest)
	run, err := client.CreateRun(context.Background(), threadId, runRequest)
	if err != nil {
		logErrorf(logFile, cliDebug, "[DEBUG] CreateRun error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	for run.Status != openai.RunStatusCompleted {
		logPrintf(logFile, cliDebug, "[DEBUG] Run status: %s\n", run.Status)
		run, err = client.RetrieveRun(context.Background(), threadId, run.ID)
		if err != nil {
			logErrorf(logFile, cliDebug, "[DEBUG] RetrieveRun error: %v\n", err)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("failed to retrieve run: %w", err)
		}
		if run.Status == openai.RunStatusFailed {
			logErrorf(logFile, cliDebug, "[DEBUG] Run failed: %v\n", run.LastError)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("run failed: %s", run.LastError.Message)
		}
		if run.Status == openai.RunStatusExpired || run.Status == openai.RunStatusCancelled {
			logErrorf(logFile, cliDebug, "[DEBUG] Run ended with status: %s\n", run.Status)
			s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
			return nil, fmt.Errorf("run ended with status: %s", run.Status)
		}
		// Handle tool calls
		if run.Status == openai.RunStatusRequiresAction &&
			run.RequiredAction != nil &&
			run.RequiredAction.Type == "submit_tool_outputs" {
			logPrintf(logFile, cliDebug, "[DEBUG] Run requires action: %+v\n", run.RequiredAction)
			var toolOutputs []openai.ToolOutput
			for _, toolCall := range run.RequiredAction.SubmitToolOutputs.ToolCalls {
				logPrintf(logFile, cliDebug, "[DEBUG] ToolCall (%s): %+v\n", toolCall.Function.Name, toolCall)
				switch {
				case strings.HasPrefix(toolCall.Function.Name, "kubectl_proxy"):
					var args struct {
						Args        []string `json:"args"`
						Description string   `json:"description"`
					}
					logPrintf(logFile, cliDebug, "[DEBUG] ToolCall.Function.Arguments (%s): %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logErrorf(logFile, cliDebug, "[DEBUG] Failed to parse tool arguments: %v\n", err)
						toolOutputs = append(toolOutputs, openai.ToolOutput{
							ToolCallID: toolCall.ID,
							Output:     fmt.Sprintf("Failed to parse arguments: %v", err),
						})
						continue
					}
					logPrintf(logFile, cliDebug, "[DEBUG] Parsed tool args (%s): %+v\n", toolCall.Function.Name, args)
					s.OutputWriter.Println("\nAI Function Call: " + args.Description)
					logPrintln(logFile, cliDebug, "\nAI Function Call: "+args.Description)
					out, err := s.KubectlProxy(args.Args)
					outStr := strings.Join(out, "\n")
					outStr = sanitizeOutput(outStr)
					const maxOutputLen = 72000 // ToDo: make this configurable
					if len(outStr) > maxOutputLen {
						outStr = outStr[:maxOutputLen] + "\n[...truncated...]"
					}
					if outStr != "" && !strings.HasPrefix(outStr, "```") {
						outStr = "```\n" + outStr + "\n```"
					}
					if err != nil {
						outStr = fmt.Sprintf("Error: %v\nOutput:\n%s", err, outStr)
					}
					logPrintln(logFile, cliDebug, "[AI Function Response]:\n"+outStr)
					logPrintf(logFile, cliDebug, "[DEBUG] Tool output length: %d\n", len(outStr))
					toolOutputs = append(toolOutputs, openai.ToolOutput{
						ToolCallID: toolCall.ID,
						Output:     outStr,
					})
				case strings.HasPrefix(toolCall.Function.Name, "helm_proxy"):
					var args struct {
						Args        []string `json:"args"`
						Description string   `json:"description"`
					}
					logPrintf(logFile, cliDebug, "[DEBUG] ToolCall.Function.Arguments (%s): %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logErrorf(logFile, cliDebug, "[DEBUG] Failed to parse tool arguments: %v\n", err)
						toolOutputs = append(toolOutputs, openai.ToolOutput{
							ToolCallID: toolCall.ID,
							Output:     fmt.Sprintf("Failed to parse arguments: %v", err),
						})
						continue
					}
					logPrintf(logFile, cliDebug, "[DEBUG] Parsed tool args (%s): %+v\n", toolCall.Function.Name, args)
					s.OutputWriter.Println("\nAI Function Call: " + args.Description)
					logPrintln(logFile, cliDebug, "\nAI Function Call: "+args.Description)
					out, err := s.HelmProxy(args.Args)
					outStr := strings.Join(out, "\n")
					outStr = sanitizeOutput(outStr)
					const maxOutputLen = 72000 // ToDo: make this configurable
					if len(outStr) > maxOutputLen {
						outStr = outStr[:maxOutputLen] + "\n[...truncated...]"
					}
					if outStr != "" && !strings.HasPrefix(outStr, "```") {
						outStr = "```\n" + outStr + "\n```"
					}
					if err != nil {
						outStr = fmt.Sprintf("Error: %v\nOutput:\n%s", err, outStr)
					}
					logPrintln(logFile, cliDebug, "[AI Function Response]:\n"+outStr)
					logPrintf(logFile, cliDebug, "[DEBUG] Tool output length: %d\n", len(outStr))
					toolOutputs = append(toolOutputs, openai.ToolOutput{
						ToolCallID: toolCall.ID,
						Output:     outStr,
					})
				default:
					logPrintf(logFile, cliDebug, "[DEBUG] Unhandled tool function: %s\n", toolCall.Function.Name)
					toolOutputs = append(toolOutputs, openai.ToolOutput{
						ToolCallID: toolCall.ID,
						Output:     fmt.Sprintf("Unhandled tool function: %s", toolCall.Function.Name),
					})
				}
			}
			logPrintf(logFile, cliDebug, "[DEBUG] Submitting tool outputs: %+v\n", toolOutputs)
			run, err = client.SubmitToolOutputs(
				context.Background(),
				threadId,
				run.ID,
				openai.SubmitToolOutputsRequest{
					ToolOutputs: toolOutputs,
				},
			)
			if err != nil {
				logErrorf(logFile, cliDebug, "[DEBUG] SubmitToolOutputs error: %v\n", err)
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
		logErrorf(logFile, cliDebug, "[DEBUG] getLatestAssistantMessage error: %v\n", err)
		s.OutputWriter.Errorf("An error occurred. See debug log: %s\n", logFilePath)
		return nil, err
	}
	logPrintf(logFile, cliDebug, "[DEBUG] Assistant response: %q\n", content)

	return &PromptResponse{
		InputPrompt: prompt,
		Content:     content,
	}, nil
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

// GetKubeNamespaces returns a list of namespaces in the current Kubernetes cluster.
func (s *Service) GetKubeNamespaces() ([]string, error) {
	return s.getKubeNamespaces()
}

// KubectlProxy executes a kubectl command and returns the output.
func (s *Service) KubectlProxy(args []string) ([]string, error) {
	return s.kubectlProxy(args)
}

// HelmProxy executes a helm command and returns the output.
func (s *Service) HelmProxy(args []string) ([]string, error) {
	return s.helmProxy(args)
}
