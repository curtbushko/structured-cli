# Phase 25: Kubernetes (kubectl) Parsers

Kubernetes CLI integration for cluster management.

## Parsers

- [x] `kubectl get pods` - Pod listing with status, restarts, age
- [x] `kubectl get services` - Service listing with type, cluster IP, ports
- [x] `kubectl get deployments` - Deployment listing with replicas
- [x] `kubectl get nodes` - Node listing with status, roles
- [x] `kubectl describe pod` - Pod details with events
- [x] `kubectl logs` - Container logs with timestamps
- [x] `kubectl top pods` - Resource usage metrics
- [x] `kubectl top nodes` - Node resource metrics

## E2E Tests

- [x] `kubectl get pods` returns structured pod data
- [x] `kubectl get services` returns service data
- [x] Graceful handling when kubectl not configured
- [x] Handles multiple namespaces
