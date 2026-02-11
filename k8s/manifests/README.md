# Kubernetes Manifests for Pterodactyl Wings

This directory contains Kubernetes manifests for deploying Wings with Kubernetes runtime provider support.

## Files

- **`rbac.yaml`** - ServiceAccount, Role, and RoleBinding for Wings to manage game servers
- **`wings-deployment.yaml`** - Wings Deployment, ConfigMap, and Service
- **`example-server.yaml`** - Example game server Pod (reference only, created automatically by Wings)

## Quick Start

### 1. Prerequisites

- Kubernetes cluster (v1.24+)
- `kubectl` configured to access the cluster
- A StorageClass available for PersistentVolumeClaims (e.g., `standard`, `local-path`)
- Panel URL and authentication token

### 2. Create Namespace

```bash
kubectl create namespace pterodactyl-servers
```

### 3. Apply RBAC Configuration

```bash
kubectl apply -f rbac.yaml
```

This creates:
- ServiceAccount: `wings-server-manager`
- Role: `wings-server-manager` (namespaced permissions)
- RoleBinding: Links ServiceAccount to Role

### 4. Configure Wings

Edit `wings-deployment.yaml` and update the ConfigMap:

```yaml
data:
  config.yml: |
    # Update these values:
    remote: https://your-panel.example.com
    token_id: your-token-id
    token: your-token-secret
    
    # ... rest of configuration
```

**Important Configuration Fields:**
- `remote` - Your Pterodactyl Panel URL
- `token_id` & `token` - Authentication credentials from Panel
- `runtime.kubernetes.namespace` - Namespace for game servers (default: `pterodactyl-servers`)
- `runtime.kubernetes.service_type` - `NodePort` or `LoadBalancer`
- `runtime.kubernetes.storage_class` - StorageClass for PVCs

### 5. Deploy Wings

```bash
kubectl apply -f wings-deployment.yaml
```

### 6. Verify Deployment

```bash
# Check Wings pod status
kubectl get pods -n pterodactyl-servers -l app=pterodactyl-wings

# View Wings logs
kubectl logs -n pterodactyl-servers -l app=pterodactyl-wings -f

# Check service
kubectl get svc -n pterodactyl-servers pterodactyl-wings
```

### 7. Access Wings API

Get the NodePort:

```bash
kubectl get svc pterodactyl-wings -n pterodactyl-servers
```

Access Wings at: `http://<node-ip>:<nodeport>`

## Configuration

### Environment Variables

Wings can be configured via environment variables (overrides config file):

| Variable | Description | Default |
|----------|-------------|---------|
| `WINGS_RUNTIME_PROVIDER` | Runtime provider (`docker` or `kubernetes`) | `docker` |
| `K8S_NAMESPACE` | Namespace for game server resources | `default` |
| `K8S_SERVICE_TYPE` | Service type (`NodePort` or `LoadBalancer`) | `NodePort` |
| `STORAGE_CLASS` | StorageClass for PVCs | (cluster default) |
| `KUBECONFIG` | Path to kubeconfig (optional) | In-cluster config |

### Storage Classes

Wings requires a StorageClass for game server storage. Common options:

- **local-path** - Local node storage (single-node clusters)
- **standard** - Default cloud provider storage
- **fast-ssd** - High-performance SSD storage
- **nfs-client** - Network File System (multi-node compatible)

Check available StorageClasses:

```bash
kubectl get storageclass
```

### Service Types

**NodePort (Default)**
- Exposes servers on all nodes
- Port range: 30000-32767
- Suitable for homelab/on-premise
- Access: `<node-ip>:<nodeport>`

**LoadBalancer**
- Requires cloud provider or MetalLB
- Dedicated external IP per server
- Automatic external DNS (if supported)
- Higher cost (cloud environments)

## Game Server Management

### Listing Game Servers

```bash
# List all game server pods
kubectl get pods -n pterodactyl-servers -l pterodactyl/server-id

# List all game server services
kubectl get svc -n pterodactyl-servers -l pterodactyl/server-id

# List all game server PVCs
kubectl get pvc -n pterodactyl-servers -l pterodactyl/server-id
```

### Accessing Game Server Console

Wings manages console access via the Panel UI. However, you can also access directly:

```bash
# Exec into a game server pod
kubectl exec -it -n pterodactyl-servers wings-server-<uuid> -- /bin/bash

# View game server logs
kubectl logs -n pterodactyl-servers wings-server-<uuid> -f
```

### Troubleshooting Game Servers

```bash
# Describe pod for events
kubectl describe pod -n pterodactyl-servers wings-server-<uuid>

# Check resource usage
kubectl top pod -n pterodactyl-servers wings-server-<uuid>

# Check PVC status
kubectl get pvc -n pterodactyl-servers wings-server-<uuid>-data
```

## Security

### RBAC Permissions

The `wings-server-manager` Role has the following permissions:

- **Pods**: Full lifecycle management, log access, exec/attach
- **Services**: Full management for port exposure
- **PVCs**: Full management for persistent storage
- **ConfigMaps/Secrets**: Future use for configuration

Permissions are **namespace-scoped** (not cluster-wide).

### Pod Security

Game server pods run with:
- **Non-root user** (UID 1000)
- **No privilege escalation**
- **All capabilities dropped**
- **ReadWriteOnce PVC** (single-node attachment)

### Network Policies (Optional)

For additional security, apply network policies:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: game-server-isolation
  namespace: pterodactyl-servers
spec:
  podSelector:
    matchLabels:
      pterodactyl/server-id: ""
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
      - podSelector: {}  # Allow from all pods in namespace
  egress:
    - to:
      - podSelector: {}
    - to:  # Allow internet access
      - namespaceSelector: {}
```

## Monitoring

### Prometheus Metrics (Future)

Wings will expose Prometheus metrics at `/metrics`:

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: pterodactyl-wings
  namespace: pterodactyl-servers
spec:
  selector:
    matchLabels:
      app: pterodactyl-wings
  endpoints:
    - port: http
      path: /metrics
```

### Health Checks

Wings deployment includes:
- **Liveness Probe**: `/api/system` endpoint (30s interval)
- **Readiness Probe**: `/api/system` endpoint (10s interval)

## Upgrading

### Upgrade Wings

```bash
# Update image in deployment
kubectl set image deployment/pterodactyl-wings -n pterodactyl-servers \
  wings=ghcr.io/pterodactyl/wings:v1.11.0

# Or edit deployment
kubectl edit deployment pterodactyl-wings -n pterodactyl-servers

# Check rollout status
kubectl rollout status deployment/pterodactyl-wings -n pterodactyl-servers
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/pterodactyl-wings -n pterodactyl-servers

# Check rollout history
kubectl rollout history deployment/pterodactyl-wings -n pterodactyl-servers
```

## Uninstalling

```bash
# Delete all game servers first (via Panel)

# Delete Wings deployment
kubectl delete -f wings-deployment.yaml

# Delete RBAC
kubectl delete -f rbac.yaml

# Delete namespace (WARNING: Deletes all resources)
kubectl delete namespace pterodactyl-servers
```

## Troubleshooting

### Wings Pod Not Starting

```bash
# Check pod events
kubectl describe pod -n pterodactyl-servers -l app=pterodactyl-wings

# Check logs
kubectl logs -n pterodactyl-servers -l app=pterodactyl-wings

# Common issues:
# - Invalid Panel URL/token
# - Insufficient RBAC permissions
# - Config file syntax errors
```

### Game Server Pod Stuck in Pending

```bash
# Check PVC binding
kubectl get pvc -n pterodactyl-servers wings-server-<uuid>-data

# Check events
kubectl get events -n pterodactyl-servers --sort-by='.lastTimestamp'

# Common issues:
# - No available StorageClass
# - Insufficient cluster resources
# - PVC binding delays
```

### Cannot Connect to Game Server

```bash
# Check service
kubectl get svc -n pterodactyl-servers wings-server-<uuid>

# Check pod status
kubectl get pod -n pterodactyl-servers wings-server-<uuid>

# Check if port is exposed
kubectl describe svc -n pterodactyl-servers wings-server-<uuid>

# Common issues:
# - NodePort not accessible from external network
# - Firewall blocking NodePort range
# - Pod not in Running state
```

## Support

- **Documentation**: See `docs/k8s-provider.md` in repository
- **GitHub Issues**: https://github.com/pterodactyl/wings/issues
- **Discord**: https://discord.gg/pterodactyl

## License

Wings is licensed under the MIT License.
