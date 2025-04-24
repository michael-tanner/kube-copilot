package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const workspaceBasePath = ".kubecopilot/workspace"

// getWorkspacePath ensures a path is relative to the workspace base path
func (s *Service) getWorkspacePath(path string) string {
	// Clean the path first
	cleanPath := filepath.Clean(path)

	// If it's already an absolute path starting with the workspace path, leave it
	if strings.HasPrefix(cleanPath, workspaceBasePath) {
		return cleanPath
	}

	// For absolute paths, we need to be careful
	if filepath.IsAbs(cleanPath) {
		// Extract just the base filename/directory
		baseName := filepath.Base(cleanPath)
		return filepath.Join(workspaceBasePath, baseName)
	}

	// For relative paths, simply join with workspace base
	return filepath.Join(workspaceBasePath, cleanPath)
}

// listFiles returns a list of files in the specified directory.
func (s *Service) listFiles(path string) ([]string, error) {
	// Ensure workspace path
	workspacePath := s.getWorkspacePath(path)

	// Ensure directory exists
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	files, err := ioutil.ReadDir(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var fileList []string
	for _, file := range files {
		// Append '/' to directories for clarity
		name := file.Name()
		if file.IsDir() {
			name += "/"
		}
		fileList = append(fileList, name)
	}

	return fileList, nil
}

// readFile reads the content of a file.
func (s *Service) readFile(path string) (string, error) {
	// Ensure workspace path
	workspacePath := s.getWorkspacePath(path)

	content, err := ioutil.ReadFile(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// writeFile writes content to a file.
func (s *Service) writeFile(path string, content string) error {
	// Ensure workspace path
	workspacePath := s.getWorkspacePath(path)

	// Create parent directories if they don't exist
	dir := filepath.Dir(workspacePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write to file
	err = ioutil.WriteFile(workspacePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
