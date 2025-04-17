package cmd_test

import (
	"bytes"
	"testing"

	"github.com/michael-tanner/kube-copilot/cmd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestStatusCommand_KeyAndNamespaceSet(t *testing.T) {
	// Setup
	viper.Set("OPENAI_API_KEY", "test-key")
	viper.Set("namespace", "test-ns")

	output := new(bytes.Buffer)
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(output)
	rootCmd.SetArgs([]string{"status"})

	// Execute
	err := rootCmd.Execute()
	assert.NoError(t, err)

	outStr := output.String()
	assert.Contains(t, outStr, "OPENAI_API_KEY: set")
	assert.Contains(t, outStr, "namespace: test-ns")
}

func TestStatusCommand_KeyUnset(t *testing.T) {
	// Setup
	viper.Set("OPENAI_API_KEY", "")
	viper.Set("namespace", "default")

	output := new(bytes.Buffer)
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(output)
	rootCmd.SetArgs([]string{"status"})

	// Execute
	err := rootCmd.Execute()
	assert.NoError(t, err)

	outStr := output.String()
	assert.Contains(t, outStr, "OPENAI_API_KEY: not set")
	assert.Contains(t, outStr, "namespace: default")
}

func TestStatusCommand_NamespaceUnset(t *testing.T) {
	// Setup
	viper.Set("OPENAI_API_KEY", "some-key")
	viper.Set("namespace", "")

	output := new(bytes.Buffer)
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(output)
	rootCmd.SetArgs([]string{"status"})

	// Execute
	err := rootCmd.Execute()
	assert.NoError(t, err)

	outStr := output.String()
	assert.Contains(t, outStr, "OPENAI_API_KEY: set")
	assert.Contains(t, outStr, "namespace: ")
}
