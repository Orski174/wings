package docker

import (
	"context"

	"github.com/pterodactyl/wings/environment"
	dockerenv "github.com/pterodactyl/wings/environment/docker"
	"github.com/pterodactyl/wings/internal/runtime"
)

// Ensure DockerProvider implements the Provider interface
var _ runtime.Provider = (*DockerProvider)(nil)

// DockerProvider is a runtime provider that uses Docker containers.
type DockerProvider struct{}

// NewProvider creates a new Docker runtime provider.
func NewProvider() *DockerProvider {
	return &DockerProvider{}
}

// Type returns the provider type name.
func (p *DockerProvider) Type() string {
	return "docker"
}

// Create creates a new Docker environment for a server.
func (p *DockerProvider) Create(id string, meta interface{}, cfg *environment.Configuration) (environment.ProcessEnvironment, error) {
	// Extract Docker-specific metadata
	dockerMeta, ok := meta.(*dockerenv.Metadata)
	if !ok {
		// Fallback: try to convert from generic metadata
		if m, ok := meta.(*runtime.Metadata); ok {
			dockerMeta = &dockerenv.Metadata{
				Image: m.Image,
				Stop:  m.Stop,
			}
		} else {
			// If neither works, create a new one with defaults
			dockerMeta = &dockerenv.Metadata{}
		}
	}

	// Use the existing Docker environment constructor
	return dockerenv.New(id, dockerMeta, cfg)
}

// IsAvailable checks if Docker is available and accessible.
func (p *DockerProvider) IsAvailable() error {
	// Try to get the Docker client
	_, err := environment.Docker()
	return err
}
