package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/michael-tanner/kube-copilot/cmd"
	"github.com/stretchr/testify/assert"
)

func TestHelpCommand(t *testing.T) {
	// Capture the output of the help command
	output := new(bytes.Buffer)
	rootCmd := cmd.GetRootCmd()       // Get the root command
	rootCmd.SetOut(output)            // Set the output to the buffer
	rootCmd.SetArgs([]string{"help"}) // Set the command arguments to "help"

	// Execute the command
	err := rootCmd.Execute()
	assert.NoError(t, err)

	// Verify the output
	expectedOutput := "Welcome to kube-copilot 😎"
	assert.True(t, strings.Contains(output.String(), expectedOutput), "Help command output should contain expected text")
}
