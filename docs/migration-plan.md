# Kubernetes Provider Migration Plan

## Overview

This document outlines the staged implementation approach for adding Kubernetes provider support to Pterodactyl Wings. The implementation is designed to be incremental, with each milestone producing compileable, testable code.

## Milestones

### Milestone 1: Foundation & Skeleton ✅

**Goal:** Establish the runtime provider abstraction and wire up both Docker and Kubernetes providers.

**Deliverables:**
- [x] Create `internal/runtime/` package structure
- [x] Define `Provider` interface in `internal/runtime/provider.go`
- [x] Move Docker implementation to `internal/runtime/docker/`
- [x] Create Kubernetes stub in `internal/runtime/kubernetes/`
- [x] Add configuration support (env vars + YAML)
- [x] Wire provider selection logic in main entry point
- [x] Documentation: `docs/k8s-provider.md`, `docs/migration-plan.md`

**Testing:**
- Unit test for provider selection logic
- Verify Docker provider still works (existing tests pass)

**Success Criteria:**
- Code compiles successfully
- Docker provider works as before (no regressions)
- Kubernetes provider loads but returns "not implemented" errors

---

### Milestone 2: Kubernetes Client & Resource Builders

**Goal:** Initialize Kubernetes client and implement resource spec generation.

**Deliverables:**
- [ ] Kubernetes client initialization (in-cluster + kubeconfig support)
- [ ] `buildPodSpec()` - Generate Pod specification from server config
- [ ] `buildServiceSpec()` - Generate Service specification from allocations
- [ ] `buildPVCSpec()` - Generate PersistentVolumeClaim specification
- [ ] Helper functions for labels, annotations, resource conversions

**Testing:**
- Unit tests for spec generation functions
- Validate generated YAML matches expected structure
- Use fake Kubernetes client for testing

**Success Criteria:**
- Generated specs are valid Kubernetes resources
- Labels and annotations are correctly applied
- Resource limits correctly converted (MB → Mi, CPU% → millicores)

---

### Milestone 3: Core Lifecycle - Create & Exists

**Goal:** Implement server creation and existence checking.

**Deliverables:**
- [ ] Implement `Create()` - Create PVC, Pod, Service in Kubernetes
- [ ] Implement `Exists()` - Check if Pod exists
- [ ] Implement `Destroy()` - Delete Pod, Service, PVC
- [ ] Error handling and logging

**Testing:**
- Unit tests with fake client
- Integration test: Create → Exists → Destroy cycle

**Success Criteria:**
- Can create a server (PVC + Pod + Service appear in cluster)
- Can query existence
- Can clean up all resources

---

### Milestone 4: State Management & Power Control

**Goal:** Implement server start, stop, and state querying.

**Deliverables:**
- [ ] Implement `Start()` - Apply Pod to cluster (if not exists, create)
- [ ] Implement `Stop()` - Delete Pod gracefully
- [ ] Implement `IsRunning()` - Check Pod phase
- [ ] Implement `State()` - Map Pod phase to Wings states
- [ ] Implement `SetState()` - Update internal state
- [ ] Implement `WaitForStop()` - Poll for Pod termination
- [ ] Implement `Terminate()` - Force delete Pod

**Testing:**
- Unit tests for state mapping
- Integration test: Start → IsRunning → Stop → State sequence

**Success Criteria:**
- Server transitions through correct states
- Pod lifecycle matches Wings semantics
- Graceful shutdown works correctly

---

### Milestone 5: Logs & Console Output

**Goal:** Implement log reading and streaming.

**Deliverables:**
- [ ] Implement `Readlog()` - Fetch tail of Pod logs
- [ ] Implement `SetLogCallback()` - Wire log streaming to callback
- [ ] Background goroutine for continuous log streaming
- [ ] Handle pod restart / log stream interruption

**Testing:**
- Unit test log parsing
- Integration test: Start server, read logs, verify callback fired

**Success Criteria:**
- Can read historical logs
- Real-time log streaming works
- Panel console displays server output

---

### Milestone 6: Interactive Console (Exec/Attach)

**Goal:** Implement command sending and interactive console.

**Deliverables:**
- [ ] Implement `Attach()` - Use Kubernetes exec API with SPDY
- [ ] Implement `SendCommand()` - Write to exec stdin
- [ ] Handle exec connection lifecycle (reconnect on failure)
- [ ] Buffer management for console I/O

**Testing:**
- Integration test: Send command, verify execution
- Test reconnection on pod restart

**Success Criteria:**
- Can send commands to game server
- Console input/output works bidirectionally
- Handles connection errors gracefully

---

### Milestone 7: Resource Monitoring & Stats

**Goal:** Implement resource usage monitoring.

**Deliverables:**
- [ ] Implement `Uptime()` - Calculate from Pod start time
- [ ] Implement resource metrics collection (CPU, memory)
- [ ] Emit `ResourceEvent` with stats
- [ ] Background polling goroutine

**Testing:**
- Unit test uptime calculation
- Integration test: Verify stats are reported

**Success Criteria:**
- Panel displays CPU/memory usage
- Uptime is accurate
- Stats update every N seconds (configurable)

---

### Milestone 8: Advanced Lifecycle Features

**Goal:** Implement remaining lifecycle methods.

**Deliverables:**
- [ ] Implement `InSituUpdate()` - Update Pod resources without restart
- [ ] Implement `ExitState()` - Get exit code and OOMKilled status
- [ ] Implement `OnBeforeStart()` - Pre-start validation (image pull check)

**Testing:**
- Unit test exit code extraction
- Integration test: Resource limit update

**Success Criteria:**
- Can update limits on running servers
- Exit codes are correctly reported
- OOMKilled detection works

---

### Milestone 9: Kubernetes Manifests & RBAC

**Goal:** Provide deployment manifests and RBAC configuration.

**Deliverables:**
- [ ] `k8s/manifests/rbac.yaml` - ServiceAccount, Role, RoleBinding
- [ ] `k8s/manifests/wings-deployment.yaml` - Wings deployment template
- [ ] `k8s/manifests/configmap.yaml` - Wings configuration
- [ ] `k8s/manifests/example-server.yaml` - Example server Pod

**Testing:**
- Deploy to test cluster
- Verify RBAC permissions are sufficient

**Success Criteria:**
- Wings deploys successfully in Kubernetes
- Can manage servers without permission errors
- Manifests follow Kubernetes best practices

---

### Milestone 10: Documentation & Examples

**Goal:** Complete user-facing documentation.

**Deliverables:**
- [ ] Update README.md with Kubernetes setup instructions
- [ ] Create deployment guide (`docs/kubernetes-deployment.md`)
- [ ] Add troubleshooting guide (`docs/kubernetes-troubleshooting.md`)
- [ ] Document environment variable reference

**Success Criteria:**
- User can deploy Wings to Kubernetes following docs
- Common issues are documented

---

### Milestone 11: Testing & Validation

**Goal:** Comprehensive test coverage and validation.

**Deliverables:**
- [ ] Unit test coverage >70% for Kubernetes provider
- [ ] Integration tests with Kind cluster
- [ ] End-to-end test: Full server lifecycle
- [ ] Performance benchmarks (vs Docker provider)

**Success Criteria:**
- All tests pass
- No regressions in Docker provider
- Kubernetes provider is production-ready

---

### Milestone 12: CI/CD Integration

**Goal:** Automate testing and builds.

**Deliverables:**
- [ ] GitHub Actions workflow for Kubernetes tests
- [ ] Docker image with Kubernetes support
- [ ] Release artifacts (binaries, manifests)

**Success Criteria:**
- CI tests pass on every commit
- Release process includes Kubernetes manifests

---

## Phase Summary

### Phase 1: Foundation (Milestones 1-2)
**Duration:** 1-2 days  
**Focus:** Architecture and scaffolding

### Phase 2: Core Implementation (Milestones 3-4)
**Duration:** 3-5 days  
**Focus:** Basic server lifecycle

### Phase 3: Console & Monitoring (Milestones 5-7)
**Duration:** 3-4 days  
**Focus:** User-facing features

### Phase 4: Polish & Deploy (Milestones 8-12)
**Duration:** 2-3 days  
**Focus:** Production readiness

**Total Estimated Duration:** 9-14 days

---

## Risk Mitigation

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Kubernetes API breaking changes | High | Pin client-go version, test against multiple K8s versions |
| SPDY deprecation (exec/attach) | Medium | Monitor WebSocket migration path, plan future upgrade |
| Storage provisioner unavailable | High | Document requirements, provide local-path example |
| NodePort conflicts | Medium | Allow custom port mapping, document range configuration |
| PVC binding delays | Medium | Implement retry logic, clear error messages |

### Operational Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| RBAC permission errors | High | Provide comprehensive RBAC manifest, clear error logs |
| Resource quota exceeded | Medium | Document quota requirements, fail gracefully |
| Network policy blocking | Medium | Document required network policies |
| Init container failures | Medium | Verbose logging, fail fast with clear errors |

---

## Success Metrics

### Functional Requirements
- ✅ Panel API unchanged (100% backward compatible)
- ✅ Server CRUD operations work
- ✅ Console input/output functional
- ✅ Resource limits enforced
- ✅ Logs accessible from Panel

### Non-Functional Requirements
- ✅ Performance within 10% of Docker provider
- ✅ No memory leaks (tested with multiple servers)
- ✅ Graceful error handling (no panics)
- ✅ Clear logging for debugging

### Documentation
- ✅ Setup guide complete
- ✅ Troubleshooting guide available
- ✅ API reference updated
- ✅ Migration guide from Docker

---

## Post-Implementation Roadmap

### Short-term (1-3 months)
- [ ] StatefulSet support (optional)
- [ ] LoadBalancer service type support
- [ ] Metrics server integration
- [ ] Prometheus exporter

### Medium-term (3-6 months)
- [ ] Multi-node scheduling with ReadWriteMany PVCs
- [ ] Network policies for isolation
- [ ] CSI snapshot integration for backups
- [ ] HPA based on player count

### Long-term (6-12 months)
- [ ] Custom CRD for Pterodactyl servers
- [ ] Operator pattern (full reconciliation)
- [ ] Gitops integration (ArgoCD/Flux)
- [ ] Multi-cluster support

---

## Testing Checklist

### Manual Testing
- [ ] Create server via Panel
- [ ] Start server
- [ ] Send console commands
- [ ] View console output
- [ ] Stop server
- [ ] Restart server
- [ ] Delete server
- [ ] View resource usage
- [ ] Upload/download files
- [ ] Create backup

### Automated Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] E2E tests pass
- [ ] Load test (100+ servers)
- [ ] Security scan (CodeQL)

---

## Rollout Plan

### Development
1. Feature branch: `feature/kubernetes-provider`
2. PR reviews for each milestone
3. Merge to `develop` after milestone completion

### Testing
1. Internal testing cluster (Kind)
2. Staging environment (real cluster)
3. Beta testers (opt-in)

### Production
1. Release candidate (RC)
2. Documentation freeze
3. Final testing
4. Release v2.0.0 (major version bump)

---

## Support Strategy

### Documentation
- Quick start guide
- FAQ
- Common issues
- Best practices

### Community
- Discord channel: #kubernetes-provider
- GitHub Discussions for Q&A
- Example configurations repository

### Maintenance
- Bug triage within 24 hours
- Security patches within 48 hours
- Feature requests via GitHub Issues
- Monthly release cycle

---

## Conclusion

This migration plan provides a clear path to implementing Kubernetes provider support while maintaining backward compatibility and ensuring production readiness. Each milestone is independently testable and brings the project closer to the final goal of running Pterodactyl Wings in Kubernetes environments.
