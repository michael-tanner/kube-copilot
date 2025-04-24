package api

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

// getKubeNamespaces returns a list of namespaces in the current Kubernetes cluster.
func (s *Service) getKubeNamespaces() ([]string, error) {
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

// kubectlProxy executes a kubectl-like command using the Go client.
// Supported: get pods, get services, get deployments, get nodes, get namespaces
func (s *Service) kubectlProxy(args []string) ([]string, error) {
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

// helmProxy executes a helm command using the system's helm binary.
func (s *Service) helmProxy(args []string) ([]string, error) {
	if len(args) < 1 {
		return nil, errors.New("no helm command provided")
	}
	cmdArgs := append([]string{}, args...)
	cmd := exec.Command("helm", cmdArgs...)
	// Optionally set KUBECONFIG for helm as well
	cmd.Env = append(os.Environ(), "KUBECONFIG="+s.ServiceContext.KubeConfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{string(output)}, err
	}
	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	return lines, nil
}
