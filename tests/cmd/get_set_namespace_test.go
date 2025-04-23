package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/michael-tanner/kube-copilot/cmd"
)

func executeCommand(args ...string) (string, error) {
	root := cmd.GetRootCmd() // You may need to export a function GetRootCmd() in cmd/root.go
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestGetSetNamespace(t *testing.T) {
	// Get original namespace
	out, err := executeCommand("get", "namespace")
	if err != nil {
		t.Fatalf("failed to get namespace: %v", err)
	}
	var origNS string
	if strings.Contains(out, "not found") {
		origNS = ""
	} else {
		parts := strings.SplitN(out, ":", 2)
		if len(parts) < 2 {
			t.Fatalf("unexpected output from get namespace: %q", out)
		}
		origNS = strings.TrimSpace(parts[1])
	}

	// Set to UnitTestNamespace
	_, err = executeCommand("set", "namespace", "UnitTestNamespace")
	if err != nil {
		t.Fatalf("failed to set namespace: %v", err)
	}

	// Get and check new value
	out, err = executeCommand("get", "namespace")
	if err != nil {
		t.Fatalf("failed to get namespace after set: %v", err)
	}
	parts := strings.SplitN(out, ":", 2)
	if len(parts) < 2 {
		t.Fatalf("unexpected output from get namespace after set: %q", out)
	}
	gotNS := strings.TrimSpace(parts[1])
	if gotNS != "UnitTestNamespace" {
		t.Errorf("expected namespace to be UnitTestNamespace, got %s", gotNS)
	}

	// Restore original namespace
	_, err = executeCommand("set", "namespace", origNS)
	if err != nil {
		t.Fatalf("failed to restore namespace: %v", err)
	}
}
