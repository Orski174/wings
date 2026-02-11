// Package runtime provides the abstraction layer for different container/server runtime providers.
// This allows Wings to support multiple backends (Docker, Kubernetes, etc.) while maintaining
// a consistent interface for server management.
package runtime

import (
	"context"

	"github.com/pterodactyl/wings/environment"
	"github.com/pterodactyl/wings/remote"
)

// Provider defines the interface that all runtime providers must implement.
// A runtime provider is responsible for creating and managing the execution environment
// for game servers (e.g., Docker containers, Kubernetes pods).
type Provider interface {
	// Type returns the name of the provider (e.g., "docker", "kubernetes").
	Type() string

	// Create creates a new environment for a server. This involves creating the necessary
	// runtime resources (container, pod, etc.) but does not start the server.
	//
	// Parameters:
	//   - id: Unique identifier for the server (typically UUID)
	//   - meta: Metadata specific to the provider (image, stop config, etc.)
	//   - cfg: Environment configuration (resource limits, allocations, etc.)
	//
	// Returns:
	//   - environment.ProcessEnvironment: The created environment
	//   - error: Any error that occurred during creation
	Create(id string, meta interface{}, cfg *environment.Configuration) (environment.ProcessEnvironment, error)

	// IsAvailable checks if the provider is available and properly configured.
	// For Docker, this checks if the Docker daemon is accessible.
	// For Kubernetes, this checks if the cluster connection is valid.
	IsAvailable() error
}

// Metadata defines common metadata that all providers need.
// Provider-specific metadata should embed this struct and add additional fields.
type Metadata struct {
	// Image is the container/pod image to use
	Image string

	// Stop configuration for graceful shutdown
	Stop remote.ProcessStopConfiguration
}
