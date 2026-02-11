package runtime

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pterodactyl/wings/internal/runtime/kubernetes"
)

func TestGetProvider(t *testing.T) {
	t.Run("returns docker provider for docker type", func(t *testing.T) {
		provider, err := GetProvider(ProviderDocker, nil)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "docker", provider.Type())
	})

	t.Run("returns docker provider for empty type", func(t *testing.T) {
		provider, err := GetProvider("", nil)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "docker", provider.Type())
	})

	t.Run("returns kubernetes provider for kubernetes type", func(t *testing.T) {
		cfg := &kubernetes.Config{
			Namespace:   "test",
			ServiceType: "NodePort",
		}
		provider, err := GetProvider(ProviderKubernetes, cfg)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "kubernetes", provider.Type())
	})

	t.Run("returns kubernetes provider with nil config", func(t *testing.T) {
		provider, err := GetProvider(ProviderKubernetes, nil)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "kubernetes", provider.Type())
	})

	t.Run("normalizes provider type case", func(t *testing.T) {
		provider, err := GetProvider("DOCKER", nil)
		require.NoError(t, err)
		assert.Equal(t, "docker", provider.Type())

		provider, err = GetProvider("  Kubernetes  ", nil)
		require.NoError(t, err)
		assert.Equal(t, "kubernetes", provider.Type())
	})

	t.Run("returns error for unknown provider type", func(t *testing.T) {
		provider, err := GetProvider("unknown", nil)
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unknown runtime provider")
	})
}

func TestGetProviderFromEnv(t *testing.T) {
	t.Run("returns docker when env not set", func(t *testing.T) {
		// Ensure env var is not set
		os.Unsetenv("WINGS_RUNTIME_PROVIDER")
		
		providerType := GetProviderFromEnv()
		assert.Equal(t, ProviderDocker, providerType)
	})

	t.Run("returns value from environment variable", func(t *testing.T) {
		os.Setenv("WINGS_RUNTIME_PROVIDER", "kubernetes")
		defer os.Unsetenv("WINGS_RUNTIME_PROVIDER")

		providerType := GetProviderFromEnv()
		assert.Equal(t, "kubernetes", providerType)
	})

	t.Run("normalizes env value", func(t *testing.T) {
		os.Setenv("WINGS_RUNTIME_PROVIDER", "  DOCKER  ")
		defer os.Unsetenv("WINGS_RUNTIME_PROVIDER")

		providerType := GetProviderFromEnv()
		assert.Equal(t, "docker", providerType)
	})
}

func TestGetKubernetesConfigFromEnv(t *testing.T) {
	t.Run("returns defaults when no env vars set", func(t *testing.T) {
		// Clear all relevant env vars
		os.Unsetenv("K8S_NAMESPACE")
		os.Unsetenv("KUBECONFIG")
		os.Unsetenv("K8S_SERVICE_TYPE")
		os.Unsetenv("STORAGE_CLASS")

		cfg := GetKubernetesConfigFromEnv()
		assert.Equal(t, "default", cfg.Namespace)
		assert.Equal(t, "", cfg.Kubeconfig)
		assert.Equal(t, "NodePort", cfg.ServiceType)
		assert.Equal(t, "", cfg.StorageClass)
	})

	t.Run("reads values from environment", func(t *testing.T) {
		os.Setenv("K8S_NAMESPACE", "pterodactyl-servers")
		os.Setenv("KUBECONFIG", "/path/to/kubeconfig")
		os.Setenv("K8S_SERVICE_TYPE", "LoadBalancer")
		os.Setenv("STORAGE_CLASS", "fast-ssd")
		defer func() {
			os.Unsetenv("K8S_NAMESPACE")
			os.Unsetenv("KUBECONFIG")
			os.Unsetenv("K8S_SERVICE_TYPE")
			os.Unsetenv("STORAGE_CLASS")
		}()

		cfg := GetKubernetesConfigFromEnv()
		assert.Equal(t, "pterodactyl-servers", cfg.Namespace)
		assert.Equal(t, "/path/to/kubeconfig", cfg.Kubeconfig)
		assert.Equal(t, "LoadBalancer", cfg.ServiceType)
		assert.Equal(t, "fast-ssd", cfg.StorageClass)
	})
}
