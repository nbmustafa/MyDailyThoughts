# k8sdiag — Kubernetes Diagnostic Tool

A comprehensive, fast Kubernetes cluster diagnostic tool written in Go.  
Designed to be called by Claude CLI via the `k8s-diagnostics` skill.

---

## Features

| Category | What is checked |
|----------|----------------|
| **Pods** | CrashLoopBackOff, OOMKilled, ImagePullBackOff, pending pods, restart counts, missing probes, resource limits, security context |
| **Nodes** | NotReady, MemoryPressure, DiskPressure, PIDPressure, NetworkUnavailable, cordoned, allocatable headroom, control-plane HA |
| **Storage** | PVC pending/lost, PV failed/released, StorageClass default, RWO multi-mount, orphaned volumes, missing volume definitions |
| **Networking** | Services with no endpoints, selector mismatches, LoadBalancer pending, NetworkPolicy deny-all, DNS egress blocking |
| **Calico** | Calico NetworkPolicy + GlobalNetworkPolicy — deny-all detection, missing order |
| **Scheduling** | NodeSelector/affinity matches zero nodes, wildcard tolerations, pod on tainted node, TopologySpreadConstraint |
| **Events** | Warning events (last 2 hours) — deduped and classified by severity |
| **Resources** | ResourceQuota exhaustion/near-limit, HPA at max, LimitRange gaps |
| **RBAC** | cluster-admin bindings, wildcard ClusterRole permissions |
| **Config** | Missing Secrets, orphaned ConfigMaps |
| **Ingress** | Missing IngressClass, missing backend services |
| **DNS** | CoreDNS/kube-dns pod health |
| **Namespace** | Terminating namespaces |

---

## Installation

### From source
```bash
git clone https://github.com/your-org/k8sdiag
cd k8sdiag
make build
make install   # installs to /usr/local/bin
```

### Cross-compile
```bash
make cross-build
# Outputs: bin/k8sdiag-linux-amd64, bin/k8sdiag-darwin-arm64, etc.
```

---

## Usage

```
k8sdiag --cluster <name> [--namespace <ns>] [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--cluster` | `-c` | required | Cluster name or kubeconfig context |
| `--namespace` | `-n` | all | Scope to a specific namespace |
| `--kubeconfig` | | ~/.kube/config | Path to kubeconfig |
| `--context` | | | Override kubeconfig context |
| `--output` | `-o` | `text` | Output format: `text`, `json`, `markdown` |
| `--timeout` | | `60s` | Global timeout for all checks |
| `--no-color` | | false | Disable colored output |
| `--verbose` | `-v` | false | Show additional detail |
| `--skip` | | | Skip categories e.g. `--skip RBAC,Ingress` |

### Examples

```bash
# Full cluster scan
k8sdiag --cluster prod-cluster

# Namespace-scoped (faster, focused)
k8sdiag --cluster prod-cluster --namespace payments

# JSON for CI pipelines or alerting
k8sdiag --cluster prod-cluster --output json | jq '.Summary'

# Markdown report
k8sdiag --cluster prod-cluster --output markdown > diag-$(date +%Y%m%d).md

# Skip noisy categories
k8sdiag --cluster prod-cluster --skip "Config,RBAC"

# Use explicit kubeconfig context
k8sdiag --cluster prod-cluster --context arn:aws:eks:us-east-1:123:cluster/prod
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No issues found |
| `1` | Warnings found (no criticals) |
| `2` | Critical issues found |

---

## Repository Structure

```
k8sdiag/
├── cmd/
│   └── k8sdiag/
│       └── main.go          # CLI entry point, Cobra commands, checker orchestration
├── internal/
│   ├── checker/
│   │   ├── pods.go          # Pod phase, container status, restarts, probes
│   │   ├── nodes.go         # Node conditions, capacity, control-plane HA
│   │   ├── storage.go       # PV, PVC, StorageClass, volume mount checks
│   │   ├── network.go       # Services, endpoints, NetworkPolicy, Calico, DNS
│   │   ├── affinity.go      # Scheduling: nodeSelector, affinity, taints, topology
│   │   ├── events.go        # Warning event analysis and classification
│   │   └── resources.go     # Quotas, HPA, RBAC, ConfigMaps, Secrets, Ingress, Namespace
│   ├── config/
│   │   └── client.go        # Kubernetes client builder and context resolver
│   ├── reporter/
│   │   └── reporter.go      # Colored text, JSON and Markdown output
│   └── types/
│       └── types.go         # Shared types: Finding, CheckResult, DiagnosticReport
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

---

## Adding a New Checker

1. Create `internal/checker/mycheck.go`
2. Implement the `Checker` interface:
   ```go
   type MyChecker struct { ... }
   func (c *MyChecker) Category() types.Category { return types.CategoryMyArea }
   func (c *MyChecker) Run(ctx context.Context) types.CheckResult { ... }
   ```
3. Register in `cmd/k8sdiag/main.go` within the `allCheckers` slice
4. Add a `CategoryMyArea` constant to `internal/types/types.go`

---

## Claude SKILL Integration

This tool is designed to be invoked by the Claude `k8s-diagnostics` skill.  
The skill handles:
- Parsing user intent (cluster name, optional namespace)
- Resolving the correct binary path
- Running the tool and streaming output
- Presenting findings in a structured, conversational format

See `../k8s-diagnostics-skill/SKILL.md` for the skill definition.
