package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"emperror.dev/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/pterodactyl/wings/environment"
	"github.com/pterodactyl/wings/events"
	"github.com/pterodactyl/wings/remote"
	"github.com/pterodactyl/wings/system"
)

// KubernetesProvider is a runtime provider that uses Kubernetes pods.
type KubernetesProvider struct {
	// namespace is the Kubernetes namespace where server pods will be created
	namespace string

	// kubeconfig path (optional, defaults to in-cluster config)
	kubeconfig string

	// serviceType is the type of Service to create (NodePort, LoadBalancer)
	serviceType string

	// storageClass for PersistentVolumeClaims
	storageClass string
}

// Config holds configuration for the Kubernetes provider.
type Config struct {
	Namespace    string
	Kubeconfig   string
	ServiceType  string
	StorageClass string
}

// NewProvider creates a new Kubernetes runtime provider.
func NewProvider(cfg *Config) (*KubernetesProvider, error) {
	if cfg == nil {
		cfg = &Config{
			Namespace:   "default",
			ServiceType: "NodePort",
		}
	}

	// Set defaults
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}
	if cfg.ServiceType == "" {
		cfg.ServiceType = "NodePort"
	}

	return &KubernetesProvider{
		namespace:    cfg.Namespace,
		kubeconfig:   cfg.Kubeconfig,
		serviceType:  cfg.ServiceType,
		storageClass: cfg.StorageClass,
	}, nil
}

// Type returns the provider type name.
func (p *KubernetesProvider) Type() string {
	return "kubernetes"
}

// Create creates a new Kubernetes environment for a server.
func (p *KubernetesProvider) Create(id string, meta interface{}, cfg *environment.Configuration) (environment.ProcessEnvironment, error) {
	// Extract Kubernetes-specific metadata
	kubeMeta, ok := meta.(*Metadata)
	if !ok {
		// Create default metadata if not provided
		kubeMeta = &Metadata{
			Image: "ubuntu:latest", // Default image
		}
		// Try to extract image if generic metadata provided
		if m, ok := meta.(map[string]interface{}); ok {
			if img, ok := m["image"].(string); ok {
				kubeMeta.Image = img
			}
		}
	}

	return p.newEnvironment(id, kubeMeta, cfg)
}

// IsAvailable checks if Kubernetes cluster connection is available.
func (p *KubernetesProvider) IsAvailable() error {
	clientset, err := p.initKubernetesClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.testConnection(ctx, clientset)
}

// Metadata defines Kubernetes-specific metadata for servers.
type Metadata struct {
	Image string
	Stop  remote.ProcessStopConfiguration

	// Additional Kubernetes-specific fields can be added here
	// For example: NodeSelector, Tolerations, etc.
}

// Environment implements the ProcessEnvironment interface for Kubernetes.
type Environment struct {
	id            string
	namespace     string
	configuration *environment.Configuration
	meta          *Metadata
	provider      *KubernetesProvider

	// Kubernetes client
	clientset *kubernetes.Clientset

	// State tracking
	st *system.AtomicString

	// Event bus
	emitter *events.Bus

	// Log callback
	logCallbackMx sync.Mutex
	logCallback   func([]byte)

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Ensure Environment implements ProcessEnvironment interface
var _ environment.ProcessEnvironment = (*Environment)(nil)

// newEnvironment creates a new Kubernetes environment for a server.
func (p *KubernetesProvider) newEnvironment(id string, meta *Metadata, cfg *environment.Configuration) (*Environment, error) {
	// Initialize Kubernetes client
	clientset, err := p.initKubernetesClient()
	if err != nil {
		return nil, err
	}

	e := &Environment{
		id:            id,
		namespace:     p.namespace,
		configuration: cfg,
		meta:          meta,
		provider:      p,
		clientset:     clientset,
		st:            system.NewAtomicString(environment.ProcessOfflineState),
		emitter:       events.NewBus(),
	}

	return e, nil
}

// Stub implementation to satisfy compilation
func (e *Environment) Type() string {
	return "kubernetes"
}

func (e *Environment) Config() *environment.Configuration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.configuration
}

func (e *Environment) Events() *events.Bus {
	return e.emitter
}

func (e *Environment) OnBeforeStart(ctx context.Context) error {
	// TODO: Implement pre-start validation (e.g., image pull check)
	return nil
}

func (e *Environment) WaitForStop(ctx context.Context, duration time.Duration, terminate bool) error {
	// TODO: Implement waiting for pod to stop
	return fmt.Errorf("kubernetes: WaitForStop() not implemented")
}

func (e *Environment) Terminate(ctx context.Context, signal string) error {
	// For Kubernetes, termination is the same as stopping (deleting the pod)
	return e.Stop(ctx)
}

func (e *Environment) InSituUpdate() error {
	// TODO: Implement updating pod resources without restart
	return fmt.Errorf("kubernetes: InSituUpdate() not implemented")
}

func (e *Environment) ExitState() (uint32, bool, error) {
	// TODO: Get pod exit code and OOMKilled status
	return 0, false, fmt.Errorf("kubernetes: ExitState() not implemented")
}

func (e *Environment) Attach(ctx context.Context) error {
	// TODO: Implement attaching to pod console
	return fmt.Errorf("kubernetes: Attach() not implemented")
}

func (e *Environment) SendCommand(cmd string) error {
	// TODO: Implement sending commands to pod
	return fmt.Errorf("kubernetes: SendCommand() not implemented")
}

func (e *Environment) Readlog(lines int) ([]string, error) {
	// TODO: Implement reading pod logs
	return nil, fmt.Errorf("kubernetes: Readlog() not implemented")
}

func (e *Environment) State() string {
	return e.st.Load()
}

// SetState sets the state of the environment. This emits an event that server's
// can hook into to take their own actions and track their own state based on
// the environment.
func (e *Environment) SetState(state string) {
	// Validate state
	if state != environment.ProcessOfflineState &&
		state != environment.ProcessStartingState &&
		state != environment.ProcessRunningState &&
		state != environment.ProcessStoppingState {
		panic(errors.New(fmt.Sprintf("kubernetes: invalid server state received: %s", state)))
	}

	// Emit the event to any listeners that are currently registered.
	if e.State() != state {
		// If the state changed make sure we update the internal tracking to note that.
		e.st.Store(state)
		e.Events().Publish(environment.StateChangeEvent, state)
	}
}

func (e *Environment) Uptime(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("kubernetes: Uptime() not implemented")
}

func (e *Environment) SetLogCallback(f func([]byte)) {
	e.logCallbackMx.Lock()
	defer e.logCallbackMx.Unlock()
	e.logCallback = f
}
