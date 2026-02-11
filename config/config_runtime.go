package config

// RuntimeConfiguration defines the configuration for the container runtime provider.
type RuntimeConfiguration struct {
	// Provider specifies which runtime to use: "docker" or "kubernetes"
	Provider string `default:"docker" yaml:"provider"`

	// Kubernetes-specific configuration
	Kubernetes KubernetesConfiguration `yaml:"kubernetes"`
}

// KubernetesConfiguration defines settings specific to the Kubernetes provider.
type KubernetesConfiguration struct {
	// Namespace where server pods will be created
	Namespace string `default:"default" yaml:"namespace"`

	// Kubeconfig path (optional, defaults to in-cluster config)
	Kubeconfig string `yaml:"kubeconfig"`

	// ServiceType for exposing server ports (NodePort or LoadBalancer)
	ServiceType string `default:"NodePort" yaml:"service_type"`

	// StorageClass for PersistentVolumeClaims (empty = cluster default)
	StorageClass string `yaml:"storage_class"`

	// NodePortRange specifies the custom NodePort range if configured
	// Format: "30000-32767" (optional)
	NodePortRange string `yaml:"nodeport_range"`
}
