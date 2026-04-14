# Phase 37: Optimize Container Parsers - Ultra-Compact JSON Format

## Goal

Optimize container and orchestration parsers (docker ps, docker images, kubectl get, helm list) to use ultra-compact array-based JSON format. These commands are common in CI/CD and can return dozens of resources with repeated field structure.

## Context

**Current Problem:**
- `docker ps` lists containers with repeated field names
- `docker images` lists images with repeated metadata
- `kubectl get` lists K8s resources with repeated structure
- `helm list` shows releases with repeated fields
- Result: Significant token waste on repeated `{"id": "...", "name": "...", "status": "..."}`

**Solution:**
- Use array tuples for resource lists
- Group by resource type (for kubectl)
- Add smart truncation (100 resources max)
- Target: 30% token savings

## Affected Parsers

1. **docker ps** - Container listing
2. **docker images** - Image listing
3. **kubectl get pods** - Pod listing
4. **kubectl get services** - Service listing
5. **kubectl get deployments** - Deployment listing
6. **helm list** - Helm release listing

## Tasks

### 1. Update Container Types (internal/adapters/parsers/docker/types.go, kubernetes/types.go)

- [ ] Create `DockerPsCompact` struct with array-based results
- [ ] Create `DockerImagesCompact` struct with array-based results
- [ ] Create `KubectlGetCompact` struct with array-based results
- [ ] Create `HelmListCompact` struct with array-based results
- [ ] Add `ContainerTuple` type: `[id, image, status, name, ports]`
- [ ] Add `ImageTuple` type: `[repository, tag, image_id, size, created]`
- [ ] Add `PodTuple` type: `[name, ready, status, restarts, age]`
- [ ] Add `ServiceTuple` type: `[name, type, cluster_ip, external_ip, ports]`
- [ ] Add `DeploymentTuple` type: `[name, ready, up_to_date, available, age]`
- [ ] Add `ReleaseTuple` type: `[name, namespace, revision, status, chart, app_version]`
- [ ] Keep old types temporarily for backward compatibility

### 2. Update Docker Ps Parser (internal/adapters/parsers/docker/ps.go)

- [ ] Modify `Parse()` to use array tuples
- [ ] Extract container ID (short or full), image, status, name
- [ ] Parse port mappings into compact string format
- [ ] Parse created/uptime duration
- [ ] Implement truncation: 100 containers max
- [ ] Filter out header line
- [ ] Update schema to reflect compact structure

### 3. Update Docker Images Parser (internal/adapters/parsers/docker/images.go)

- [ ] Use array tuples for image list
- [ ] Extract repository, tag, image ID, size, created date
- [ ] Handle `<none>` tags appropriately
- [ ] Parse size with units (MB, GB)
- [ ] Implement truncation: 100 images max
- [ ] Filter out header line
- [ ] Update schema

### 4. Update Kubectl Get Pods Parser (internal/adapters/parsers/kubernetes/pods.go)

- [ ] Use array tuples for pod list
- [ ] Extract name, ready count (e.g., "2/2"), status, restarts, age
- [ ] Parse namespace if `-A` or `--all-namespaces` used
- [ ] Handle pod status (Running, Pending, Failed, etc.)
- [ ] Implement truncation: 100 pods max
- [ ] Filter out header line and warnings
- [ ] Update schema

### 5. Update Kubectl Get Services Parser (internal/adapters/parsers/kubernetes/services.go)

- [ ] Use array tuples for service list
- [ ] Extract name, type (ClusterIP, NodePort, LoadBalancer), cluster-ip, external-ip, ports
- [ ] Parse port format (80:30080/TCP)
- [ ] Handle `<none>` or `<pending>` values
- [ ] Implement truncation: 100 services max
- [ ] Update schema

### 6. Update Kubectl Get Deployments Parser (internal/adapters/parsers/kubernetes/deployments.go)

- [ ] Use array tuples for deployment list
- [ ] Extract name, ready replicas, up-to-date count, available count, age
- [ ] Parse replica counts (e.g., "3/3")
- [ ] Implement truncation: 100 deployments max
- [ ] Update schema

### 7. Update Helm List Parser (internal/adapters/parsers/helm/list.go)

- [ ] Use array tuples for release list
- [ ] Extract name, namespace, revision, status, chart, app version
- [ ] Parse deployment status (deployed, failed, pending-install, etc.)
- [ ] Handle different output formats (table, json, yaml)
- [ ] Implement truncation: 100 releases max
- [ ] Filter out header line
- [ ] Update schema

### 8. Update Tests for All Container Parsers

#### Docker Ps Tests (internal/adapters/parsers/docker/ps_test.go)
- [ ] Update `TestDockerPsParser_BasicList` for tuple format
- [ ] Update `TestDockerPsParser_WithPorts` for port parsing
- [ ] Add `TestDockerPsParser_Truncation` for 100-container limit
- [ ] Add `TestDockerPsParser_EmptyList` for no containers
- [ ] Update existing tests for new format

#### Docker Images Tests (internal/adapters/parsers/docker/images_test.go)
- [ ] Update for tuple format
- [ ] Test `<none>` tag handling
- [ ] Add truncation tests
- [ ] Test size parsing

#### Kubectl Pods Tests (internal/adapters/parsers/kubernetes/pods_test.go)
- [ ] Update for tuple format
- [ ] Test namespace parsing with `-A`
- [ ] Test various pod statuses
- [ ] Add truncation tests

#### Kubectl Services Tests
- [ ] Update for tuple format
- [ ] Test service types
- [ ] Test port parsing
- [ ] Add truncation tests

#### Kubectl Deployments Tests
- [ ] Update for tuple format
- [ ] Test replica count parsing
- [ ] Add truncation tests

#### Helm List Tests (internal/adapters/parsers/helm/list_test.go)
- [ ] Update for tuple format
- [ ] Test various release statuses
- [ ] Test different namespaces
- [ ] Add truncation tests

### 9. Update JSON Schemas

- [ ] Update `schemas/docker-ps.json` for compact format
- [ ] Update `schemas/docker-images.json` for compact format
- [ ] Update `schemas/kubectl-get-pods.json` for compact format
- [ ] Update `schemas/kubectl-get-services.json` for compact format
- [ ] Update `schemas/kubectl-get-deployments.json` for compact format
- [ ] Update `schemas/helm-list.json` for compact format
- [ ] Document tuple formats in each schema
- [ ] Add examples showing array structure

### 10. Update Feature Tests

- [ ] Update docker scenarios in `features/docker.feature`
- [ ] Update kubectl scenarios in `features/kubernetes.feature`
- [ ] Update helm scenarios in `features/helm.feature`
- [ ] Add truncation scenarios for large clusters
- [ ] Verify namespace handling

### 11. Token Savings Validation

- [ ] Benchmark docker ps with 20 containers
- [ ] Benchmark docker images with 50 images
- [ ] Benchmark kubectl get pods with 50 pods
- [ ] Benchmark with 100+ resources (truncation)
- [ ] Verify 30% token savings target
- [ ] Measure before/after for common use cases

### 12. Documentation

- [ ] Update CLAUDE.md with container parser examples
- [ ] Update README.md docker/kubernetes sections
- [ ] Document truncation limits
- [ ] Document tuple formats
- [ ] Add migration guide

## Implementation Notes

### Docker Ps Compact Format

```json
{
  "total_containers": 5,
  "containers": [
    ["abc123def456", "nginx:latest", "Up 2 hours", "web-server", "80:8080/tcp, 443:8443/tcp"],
    ["def789ghi012", "redis:7", "Up 3 days", "cache", "6379/tcp"],
    ["ghi345jkl678", "postgres:15", "Up 1 week", "database", "5432/tcp"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[container_id, image, status, name, ports]`

### Docker Images Compact Format

```json
{
  "total_images": 15,
  "images": [
    ["nginx", "latest", "sha256:abc123", "187MB", "2 weeks ago"],
    ["redis", "7", "sha256:def456", "138MB", "3 weeks ago"],
    ["postgres", "15", "sha256:ghi789", "412MB", "1 month ago"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[repository, tag, image_id, size, created]`

### Kubectl Get Pods Compact Format

```json
{
  "total_pods": 23,
  "namespace": "default",
  "pods": [
    ["web-deployment-7d4f8c9b5-x7k2m", "2/2", "Running", 0, "5d"],
    ["web-deployment-7d4f8c9b5-9p3ql", "2/2", "Running", 1, "5d"],
    ["cache-deployment-6c8f9d7b4-h4k9n", "1/1", "Running", 0, "10d"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[name, ready, status, restarts, age]`

### Kubectl Get Services Compact Format

```json
{
  "total_services": 8,
  "namespace": "default",
  "services": [
    ["kubernetes", "ClusterIP", "10.96.0.1", "<none>", "443/TCP"],
    ["web-service", "LoadBalancer", "10.98.123.45", "52.12.34.56", "80:30080/TCP,443:30443/TCP"],
    ["cache-service", "ClusterIP", "10.99.234.56", "<none>", "6379/TCP"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[name, type, cluster_ip, external_ip, ports]`

### Kubectl Get Deployments Compact Format

```json
{
  "total_deployments": 5,
  "namespace": "default",
  "deployments": [
    ["web-deployment", "3/3", "3", "3", "10d"],
    ["api-deployment", "2/2", "2", "2", "8d"],
    ["worker-deployment", "5/5", "5", "5", "12d"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[name, ready, up_to_date, available, age]`

### Helm List Compact Format

```json
{
  "total_releases": 12,
  "releases": [
    ["ingress-nginx", "ingress-nginx", "1", "deployed", "ingress-nginx-4.8.3", "1.9.5"],
    ["cert-manager", "cert-manager", "2", "deployed", "cert-manager-v1.13.2", "v1.13.2"],
    ["prometheus", "monitoring", "5", "deployed", "kube-prometheus-stack-54.2.2", "v0.69.1"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[name, namespace, revision, status, chart, app_version]`

### Truncation Strategy

1. **Global limit**: 100 resources per command
2. **Priority**: No special priority (first N resources)
3. **Metadata**: Track total vs displayed
4. **Warning**: Log when truncation occurs

### Token Savings Math

**Docker ps - 20 containers**:

Old format (~2,400 tokens):
```json
{"containers": [
  {"id": "abc123def456", "image": "nginx:latest", "status": "Up 2 hours", "name": "web-server", "ports": "80:8080/tcp"},
  // ... repeated 20 times
]}
```

New format (~1,680 tokens):
```json
{"total_containers": 20, "containers": [
  ["abc123def456", "nginx:latest", "Up 2 hours", "web-server", "80:8080/tcp"],
  // ... 20 entries
]}
```

**Savings**: ~30% reduction

**Kubectl get pods - 50 pods**:

Old format (~6,000 tokens):
```json
{"pods": [
  {"name": "web-7d4f8c9b5-x7k2m", "ready": "2/2", "status": "Running", "restarts": 0, "age": "5d"},
  // ... repeated 50 times
]}
```

New format (~4,200 tokens):
```json
{"total_pods": 50, "pods": [
  ["web-7d4f8c9b5-x7k2m", "2/2", "Running", 0, "5d"],
  // ... 50 entries
]}
```

**Savings**: ~30% reduction

### Edge Cases to Handle

**Docker**:
- No containers/images: `{"containers": [], "truncated": 0}`
- Exited containers: Include in status
- Port ranges: Compact format like "8000-8010/tcp"
- No ports: Empty string or omit

**Kubectl**:
- Multiple namespaces (`-A`): Add namespace field to tuple or group by namespace
- CrashLoopBackOff: Handle in status field
- Pending pods: Include in results
- No resources: Empty array

**Helm**:
- Failed deployments: Include in status
- Different namespaces: Include in tuple
- No releases: Empty array

## Acceptance Criteria

- [ ] All 6 container parsers use compact array format
- [ ] Token savings are ~30% for typical outputs
- [ ] Truncation works for large clusters (100+ resources)
- [ ] All existing tests pass with new format
- [ ] Feature tests pass
- [ ] Schemas are updated and validated
- [ ] Header lines are filtered
- [ ] No regressions in functionality

## Success Metrics

- 30% token savings on container listings with 20+ resources
- Large clusters (100+ resources) truncated appropriately
- Consistent tuple format across all parsers
- Token cost for 20 docker containers: ~1,680 tokens (vs ~2,400 currently)
- Token cost for 50 kubectl pods: ~4,200 tokens (vs ~6,000 currently)
