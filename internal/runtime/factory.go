package runtime

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterodactyl/wings/internal/runtime/docker"
	"github.com/pterodactyl/wings/internal/runtime/kubernetes"
)

const (
	// ProviderDocker is the Docker runtime provider
	ProviderDocker = "docker"
	// ProviderKubernetes is the Kubernetes runtime provider
	ProviderKubernetes = "kubernetes"
)

// GetProvider returns the configured runtime provider based on environment variables
// and configuration. It checks WINGS_RUNTIME_PROVIDER environment variable first,
// then falls back to the configuration file, and finally defaults to Docker.
func GetProvider(providerType string, kubeCfg *kubernetes.Config) (Provider, error) {
	// Normalize provider type
	providerType = strings.ToLower(strings.TrimSpace(providerType))

	switch providerType {
	case ProviderDocker, "":
		// Default to Docker if not specified
		return docker.NewProvider(), nil

	case ProviderKubernetes:
		return kubernetes.NewProvider(kubeCfg)

	default:
		return nil, fmt.Errorf("unknown runtime provider: %s (supported: docker, kubernetes)", providerType)
	}
}

// GetProviderFromEnv returns the provider type from environment variable.
// It checks WINGS_RUNTIME_PROVIDER and defaults to "docker" if not set.
func GetProviderFromEnv() string {
	provider := os.Getenv("WINGS_RUNTIME_PROVIDER")
	if provider == "" {
		return ProviderDocker
	}
	return strings.ToLower(strings.TrimSpace(provider))
}

// GetKubernetesConfigFromEnv returns Kubernetes configuration from environment variables.
func GetKubernetesConfigFromEnv() *kubernetes.Config {
	return &kubernetes.Config{
		Namespace:    getEnvOrDefault("K8S_NAMESPACE", "default"),
		Kubeconfig:   os.Getenv("KUBECONFIG"),
		ServiceType:  getEnvOrDefault("K8S_SERVICE_TYPE", "NodePort"),
		StorageClass: os.Getenv("STORAGE_CLASS"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
