# Kubernetes Provider Design

## Overview

This document describes the design and implementation of the Kubernetes provider for Pterodactyl Wings. The goal is to replace Docker as the container runtime while maintaining full compatibility with the Pterodactyl Panel API.

## Architecture

### High-Level Design

Wings will run as a Deployment (or DaemonSet) inside a Kubernetes cluster and manage game server workloads as Kubernetes resources in the same namespace. The existing Panel-facing API remains unchanged - all changes are internal to Wings' runtime implementation.

```
┌─────────────────┐
│ Pterodactyl     │
│ Panel           │
└────────┬────────┘
         │ HTTP API (unchanged)
         ▼
┌─────────────────────────────────┐
│ Wings (Kubernetes Pod)          │
│  ┌──────────────────────────┐   │
│  │ Runtime Provider Layer   │   │
│  │  - Docker Provider       │   │
│  │  - Kubernetes Provider   │◄──┼─ Config: WINGS_RUNTIME_PROVIDER
│  └──────────────────────────┘   │
└─────────────┬───────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│ Kubernetes API Server           │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│ Game Server Pods + Services     │
│ + PersistentVolumeClaims        │
└─────────────────────────────────┘
```

## Resource Mapping

### Server → Kubernetes Resources

Each Pterodactyl "Server" is represented as a set of Kubernetes resources:

| Resource | Purpose | Naming Convention |
|----------|---------|-------------------|
| **Pod** | Runs the game server container | `wings-server-{uuid}` |
| **Service** | Exposes game ports (UDP/TCP) | `wings-server-{uuid}` |
| **PersistentVolumeClaim** | Persistent storage for server files | `wings-server-{uuid}-data` |

**Why Pod instead of StatefulSet?**
- Wings acts as the controller, managing lifecycle directly
- StatefulSets add complexity with automatic restart/scheduling behavior
- Simpler to map existing Docker semantics to Pods
- Future: Could migrate to StatefulSet for enhanced stability

### Networking Strategy

**Port Exposure: NodePort (default)**
- Each server gets a Service with `type: NodePort`
- Pterodactyl allocations map to Service ports
- NodePort range: 30000-32767 (configurable via kube-apiserver)
- Suitable for homelab/on-premise deployments

**Alternative: LoadBalancer**
- Configure via `K8S_SERVICE_TYPE=LoadBalancer`
- Requires cloud provider or MetalLB
- Each server gets a dedicated external IP (cost consideration)

**Port Allocation Flow:**
```
Panel Allocation (e.g., 25565 TCP) 
    ↓
Service Port Definition (port: 25565, nodePort: auto or specified)
    ↓
Exposed on Node IP:NodePort
```

**Allocation Format:**
- Primary allocation: Used for Service `port` and `targetPort`
- Additional allocations: Additional ports in the same Service
- IP binding: Service `clusterIP` or external IP

### Storage Strategy

**PersistentVolumeClaim per Server:**
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: wings-server-{uuid}-data
  labels:
    app: pterodactyl-wings
    pterodactyl/server-id: {uuid}
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: {STORAGE_CLASS}  # Configurable
  resources:
    requests:
      storage: {disk_limit}Gi
```

**Mount Path:** `/home/container` (same as Docker implementation)

**StorageClass Configuration:**
- Environment variable: `STORAGE_CLASS` (default: cluster default)
- Must support `ReadWriteOnce` access mode
- For Talos: Consider Rook/Ceph or local-path-provisioner

**Init Container for Permissions:**
- Run as privileged to `chown` files to game server user
- Set correct permissions before main container starts

### Console/Logs Implementation

**Logs:**
- Use Kubernetes Logs API: `clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})`
- Stream logs via `TailLines` and `Follow` options
- Map to Wings' `Readlog()` and log callback system

**Interactive Console (Attach/Exec):**
- Use SPDY-based `exec` API from client-go
- `remotecommand.NewSPDYExecutor()` for WebSocket-like bidirectional I/O
- Attach to main container's stdin/stdout
- Implement `SendCommand()` by writing to exec stream

**Challenges:**
- Exec requires container to be running with shell (`/bin/sh` or `/bin/bash`)
- Game servers may not have interactive shells
- **Solution:** Use `kubectl exec` equivalent with game console piped through stdin

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WINGS_RUNTIME_PROVIDER` | `docker` | Runtime provider: `docker` or `kubernetes` |
| `KUBECONFIG` | In-cluster config | Path to kubeconfig file (optional) |
| `K8S_NAMESPACE` | `default` | Namespace for game server resources |
| `K8S_SERVICE_TYPE` | `NodePort` | Service type: `NodePort` or `LoadBalancer` |
| `K8S_NODEPORT_RANGE` | (none) | Custom NodePort range if configured on cluster |
| `STORAGE_CLASS` | (cluster default) | StorageClass for PVCs |

### Configuration File Extension

Add to `/etc/pterodactyl/config.yml`:

```yaml
runtime:
  provider: kubernetes  # or docker
  
kubernetes:
  namespace: pterodactyl-servers
  service_type: NodePort
  storage_class: fast-ssd
  # kubeconfig: /path/to/kubeconfig  # Optional, defaults to in-cluster
```

## Security Considerations

### RBAC Requirements

Wings needs the following Kubernetes permissions (see `k8s/manifests/rbac.yaml`):

**Minimum Required:**
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: wings-server-manager
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/log", "pods/exec", "pods/attach"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["pods/status"]
    verbs: ["get"]
```

**Scope:** Role (namespaced), NOT ClusterRole - least privilege

### Pod Security

**Avoid Privileged Pods:**
- Game servers should run as non-root
- Use `securityContext` with `runAsNonRoot: true`
- Drop unnecessary capabilities
- Set `allowPrivilegeEscalation: false`

**Init Container Exception:**
- Init container may need elevated permissions for `chown`
- Runs before main container, isolated from game server process

**Network Policies (Future):**
- Restrict pod-to-pod communication
- Allow only necessary egress (game APIs, mod downloads)

## Implementation Details

### Client Library Choice: `client-go`

**Rationale:**
- Standard Kubernetes client library
- Direct access to typed resources (corev1.Pod, corev1.Service)
- Well-documented, stable API
- Supports SPDY exec/attach out of the box

**Not using controller-runtime:**
- Controller-runtime is designed for building operators/controllers
- Wings is not a reconciling controller - it directly creates/manages resources
- controller-runtime adds complexity and dependencies (controller-manager, manager)
- client-go is lighter and more appropriate for this use case

### State Mapping

Map Kubernetes Pod phases to Wings states:

| Pod Phase | Wings State | Notes |
|-----------|-------------|-------|
| `Pending` | `starting` | Pod created, waiting for scheduling |
| `Running` | `running` | Container is running |
| `Succeeded` | `offline` | Container exited with code 0 |
| `Failed` | `offline` | Container exited with error |
| `Unknown` | `offline` | Cannot determine pod status |

**Additional Checks:**
- `Ready` condition: Container passed readiness probes
- `ContainerStatuses[].State.Running`: Confirm main container is active
- Exit code from `ContainerStatuses[].State.Terminated.ExitCode`

### Resource Limits

Map Wings configuration to Pod resource requests/limits:

```yaml
spec:
  containers:
    - name: server
      resources:
        requests:
          memory: "{memory}Mi"
          cpu: "{cpu}m"
        limits:
          memory: "{memory}Mi"
          cpu: "{cpu}m"  # Or omit for burstable
```

**CPU Units:**
- Wings: CPU percentage (e.g., 100 = 1 core)
- Kubernetes: millicores (1000m = 1 core)
- Conversion: `cpu_percent * 10` = millicores

**Memory Units:**
- Wings: MB
- Kubernetes: `{value}Mi`

### Labels and Annotations

**Standard Labels:**
```yaml
metadata:
  labels:
    app: pterodactyl-wings
    pterodactyl/server-id: "{uuid}"
    pterodactyl/server-name: "{name}"
```

**Annotations for Metadata:**
```yaml
metadata:
  annotations:
    pterodactyl/image: "{docker_image}"
    pterodactyl/startup-command: "{startup}"
    pterodactyl/created-at: "{timestamp}"
```

## Tradeoffs and Limitations

### Tradeoffs

| Aspect | Docker Provider | Kubernetes Provider |
|--------|-----------------|---------------------|
| **Setup Complexity** | Low (docker.sock) | Medium (RBAC, cluster access) |
| **Isolation** | Container-level | Pod-level (same, but scheduled) |
| **Networking** | Direct port binding | NodePort/LoadBalancer indirection |
| **Storage** | Host bind mounts | PVC (requires storage provisioner) |
| **Observability** | Docker stats | Kubernetes metrics (native) |
| **Scaling** | Single-node focused | Multi-node capable |

### Known Limitations (Phase 1)

1. **No Multi-Node Scheduling:**
   - Wings assumes local storage (PVC on same node)
   - Future: Use ReadWriteMany PVCs or distributed storage

2. **No Automatic Restart:**
   - Wings manages lifecycle explicitly (no restartPolicy)
   - Crash detection remains Wings responsibility

3. **Init Container Overhead:**
   - Permission setup adds startup time
   - ~2-5 seconds per server start

4. **NodePort Range Constraints:**
   - Limited to 30000-32767 by default
   - May conflict with existing services
   - Consider allocating specific ranges for game servers

5. **No Docker Socket:**
   - Cannot use Docker-specific features (e.g., BuildKit)
   - Egg images must be pre-built and pushed to registry

## Migration Path

See `docs/migration-plan.md` for detailed milestones.

**Summary:**
1. ✅ Skeleton: Interface + stubbed providers
2. 🔄 Core: Create/Start/Stop/Status
3. ⏳ Console: Logs + Exec
4. ⏳ Networking: Service management
5. ⏳ Advanced: File operations, backups

## Testing Strategy

**Unit Tests:**
- Provider selection logic
- Resource spec generation (Pod/Service/PVC builders)
- State mapping functions
- Fake Kubernetes client for API interactions

**Integration Tests (Future):**
- Kind cluster or similar
- Full server lifecycle test
- Console interaction test

## Open Questions / Future Work

1. **StatefulSet Migration:** When/if to move from Pods to StatefulSets
2. **Ingress Support:** Expose HTTP-based servers via Ingress instead of NodePort
3. **HPA Integration:** Auto-scaling based on player count
4. **Multi-Tenancy:** Multiple Wings instances in different namespaces
5. **Volume Snapshots:** Backup integration with CSI snapshots
6. **Network Policies:** Fine-grained network segmentation
7. **Pod Disruption Budgets:** Graceful node maintenance

## References

- [client-go Documentation](https://github.com/kubernetes/client-go)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Pterodactyl Wings](https://github.com/pterodactyl/wings)
