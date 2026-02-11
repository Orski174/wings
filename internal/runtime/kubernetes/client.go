package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"emperror.dev/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// initKubernetesClient initializes the Kubernetes client using either in-cluster config or kubeconfig file.
func (p *KubernetesProvider) initKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// If kubeconfig is specified, use it
	if p.kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", p.kubeconfig)
		if err != nil {
			return nil, errors.Wrap(err, "kubernetes: failed to build config from kubeconfig")
		}
	} else {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fallback to default kubeconfig location
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return nil, errors.Wrap(err, "kubernetes: failed to get in-cluster config and cannot find home directory")
			}
			
			kubeconfigPath := filepath.Join(home, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				return nil, errors.Wrap(err, "kubernetes: failed to initialize client (tried in-cluster and ~/.kube/config)")
			}
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "kubernetes: failed to create clientset")
	}

	return clientset, nil
}

// testConnection tests the Kubernetes connection by making a simple API call.
func (p *KubernetesProvider) testConnection(ctx context.Context, clientset *kubernetes.Clientset) error {
	// Try to list namespaces to verify connection
	_, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return errors.Wrap(err, "kubernetes: failed to connect to cluster")
	}
	return nil
}
