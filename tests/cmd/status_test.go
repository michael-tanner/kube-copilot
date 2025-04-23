package cmd

import (
	"strings"
	"testing"

	"bytes"

	"github.com/michael-tanner/kube-copilot/cmd"
)

func executeStatusCommand(args ...string) (string, error) {
	root := cmd.GetRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestStatusCommand(t *testing.T) {
	out, err := executeStatusCommand("status")
	if err != nil {
		t.Fatalf("failed to run status command: %v", err)
	}
	if strings.Contains(out, "Error checking status: stat") && strings.Contains(out, ".kube/config") {
		// Acceptable in CI where kube config is missing
		return
	}
	if !strings.Contains(out, "Kube Copilot CLI Status") {
		t.Errorf("expected output to contain 'Kube Copilot CLI Status', got: %q", out)
	}
	if !strings.Contains(out, "Current Namespace:") {
		t.Errorf("expected output to contain 'Current Namespace:', got: %q", out)
	}
	if !strings.Contains(out, "OPENAI_API_KEY:") {
		t.Errorf("expected output to contain 'OPENAI_API_KEY:', got: %q", out)
	}
}
