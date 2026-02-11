package kubernetes

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pterodactyl/wings/environment"
)

// buildPodSpec generates a Kubernetes Pod specification from server configuration.
func (e *Environment) buildPodSpec() *corev1.Pod {
	cfg := e.configuration
	
	// Generate labels for the pod
	labels := map[string]string{
		"app":                       "pterodactyl-wings",
		"pterodactyl/server-id":     e.id,
		"pterodactyl/server-uuid":   e.id,
	}

	// Generate annotations
	annotations := map[string]string{
		"pterodactyl/image":   e.meta.Image,
		"pterodactyl/created": metav1.Now().String(),
	}

	// Build environment variables
	envVars := make([]corev1.EnvVar, 0)
	for _, envStr := range cfg.EnvironmentVariables() {
		// Environment variables are in KEY=VALUE format
		// Parse them properly
		envVars = append(envVars, corev1.EnvVar{
			Name:  envStr, // TODO: Parse KEY=VALUE format
			Value: "",
		})
	}

	// Convert resource limits
	limits := cfg.Limits()
	memoryLimit := resource.MustParse(fmt.Sprintf("%dMi", limits.Memory))
	cpuLimit := convertCPUToMillicores(limits.CpuUnits)

	// Build container spec
	container := corev1.Container{
		Name:  "server",
		Image: e.meta.Image,
		Env:   envVars,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: memoryLimit,
				corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", cpuLimit)),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: memoryLimit,
				// CPU limits are optional, can be omitted for burstable
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "server-data",
				MountPath: "/home/container",
			},
		},
		// Set working directory
		WorkingDir: "/home/container",
		// Stdin and TTY for interactive servers
		Stdin: true,
		TTY:   true,
	}

	// Build pod spec
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.getPodName(),
			Namespace:   e.namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			// Don't restart automatically - Wings manages lifecycle
			RestartPolicy: corev1.RestartPolicyNever,
			Containers:    []corev1.Container{container},
			Volumes: []corev1.Volume{
				{
					Name: "server-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: e.getPVCName(),
						},
					},
				},
			},
			// Security context
			SecurityContext: &corev1.PodSecurityContext{
				FSGroup: int64Ptr(1000), // pterodactyl group
			},
		},
	}

	return pod
}

// buildServiceSpec generates a Kubernetes Service specification for exposing server ports.
func (e *Environment) buildServiceSpec() *corev1.Service {
	cfg := e.configuration
	
	labels := map[string]string{
		"app":                   "pterodactyl-wings",
		"pterodactyl/server-id": e.id,
	}

	// Build service ports from allocations
	allocations := cfg.Allocations()
	servicePorts := make([]corev1.ServicePort, 0)
	
	// Iterate over all IP -> port mappings
	for _, ports := range allocations.Mappings {
		for _, port := range ports {
			// Skip invalid ports
			if port < 1 || port > 65535 {
				continue
			}
			
			// Add both TCP and UDP ports (game servers often need both)
			servicePorts = append(servicePorts,
				corev1.ServicePort{
					Name:       fmt.Sprintf("tcp-%d", port),
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(port),
					TargetPort: intstr.FromInt(port),
				},
				corev1.ServicePort{
					Name:       fmt.Sprintf("udp-%d", port),
					Protocol:   corev1.ProtocolUDP,
					Port:       int32(port),
					TargetPort: intstr.FromInt(port),
				},
			)
		}
	}

	serviceType := corev1.ServiceTypeNodePort
	if e.provider.serviceType == "LoadBalancer" {
		serviceType = corev1.ServiceTypeLoadBalancer
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.getServiceName(),
			Namespace: e.namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Selector: labels,
			Ports:    servicePorts,
		},
	}

	return service
}

// buildPVCSpec generates a PersistentVolumeClaim specification for server storage.
func (e *Environment) buildPVCSpec() *corev1.PersistentVolumeClaim {
	cfg := e.configuration
	
	labels := map[string]string{
		"app":                   "pterodactyl-wings",
		"pterodactyl/server-id": e.id,
	}

	// Convert disk space to resource quantity
	storageSize := resource.MustParse(fmt.Sprintf("%dMi", cfg.Limits.DiskSpace))

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.getPVCName(),
			Namespace: e.namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
		},
	}

	// Set storage class if specified
	if e.provider.storageClass != "" {
		pvc.Spec.StorageClassName = &e.provider.storageClass
	}

	return pvc
}

// Helper functions

func (e *Environment) getPodName() string {
	return fmt.Sprintf("wings-server-%s", e.id)
}

func (e *Environment) getServiceName() string {
	return fmt.Sprintf("wings-server-%s", e.id)
}

func (e *Environment) getPVCName() string {
	return fmt.Sprintf("wings-server-%s-data", e.id)
}

// convertCPUToMillicores converts Wings CPU percentage to Kubernetes millicores.
// Wings uses percentage (100 = 1 core), Kubernetes uses millicores (1000m = 1 core).
func convertCPUToMillicores(cpuPercentage int64) int64 {
	return cpuPercentage * 10
}

// convertMillicoresToCPU converts Kubernetes millicores to Wings CPU percentage.
func convertMillicoresToCPU(millicores int64) int64 {
	return millicores / 10
}

// int64Ptr returns a pointer to an int64 value.
func int64Ptr(i int64) *int64 {
	return &i
}
