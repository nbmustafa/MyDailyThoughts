# Diagnostic Coverage Reference

## Pods (`checker/pods.go`)

| Check | Severity |
|-------|---------|
| Pod in Failed phase | CRITICAL |
| Pod in Unknown phase | CRITICAL |
| CrashLoopBackOff | CRITICAL |
| OOMKilled | CRITICAL |
| ImagePullBackOff / ErrImagePull | CRITICAL |
| CreateContainerConfigError | CRITICAL |
| Init container stuck/failing | CRITICAL |
| Unschedulable pod (>15 min) | CRITICAL |
| Pending pod (>2 min) | WARNING |
| High restart count (>20) | CRITICAL |
| Moderate restart count (>5) | WARNING |
| Running but not ready | WARNING |
| Missing resource limits | WARNING |
| Missing resource requests | INFO |
| Missing liveness probe | INFO |
| Missing readiness probe | INFO |
| Pod may run as root | INFO |
| Latest/untagged image | WARNING |

## Nodes (`checker/nodes.go`)

| Check | Severity |
|-------|---------|
| NotReady | CRITICAL |
| NotReady > 5 minutes | CRITICAL |
| MemoryPressure | CRITICAL |
| DiskPressure | CRITICAL |
| PIDPressure | CRITICAL |
| NetworkUnavailable | CRITICAL |
| Control plane node not ready | CRITICAL |
| Cordoned/unschedulable | WARNING |
| Low allocatable CPU (<50%) | WARNING |
| Low allocatable memory (<50%) | WARNING |
| Single control plane (no HA) | INFO |
| Kernel info unavailable | INFO |

## Storage (`checker/storage.go`)

| Check | Severity |
|-------|---------|
| PVC Pending >10 minutes | CRITICAL |
| PVC in Lost state | CRITICAL |
| PV in Failed state | CRITICAL |
| VolumeMount references undefined volume | CRITICAL |
| PVC Pending <10 minutes | WARNING |
| ReadWriteOnce PVC on multiple pods | WARNING |
| PV in Released state | INFO |
| PV available/unbound | INFO |
| PV reclaim policy is Delete | INFO |
| No default StorageClass | WARNING |
| Multiple default StorageClasses | WARNING |
| PVC using default StorageClass | INFO |

## Networking (`checker/network.go`)

### Services & Endpoints
| Check | Severity |
|-------|---------|
| Service has no ready endpoints (no not-ready) | CRITICAL |
| Service selector matches no pods | CRITICAL |
| Service has no Endpoints object | WARNING |
| Service has endpoints but some not-ready | WARNING |
| LoadBalancer pending IP >3 min | WARNING |
| NodePort outside default range | WARNING |

### NetworkPolicy
| Check | Severity |
|-------|---------|
| Default-deny-all ingress | INFO |
| Default-deny-all egress | INFO |
| Egress policy may block DNS | WARNING |
| Pod not covered by any NetworkPolicy | INFO |

### Calico
| Check | Severity |
|-------|---------|
| Calico CRDs not installed | INFO |
| Calico deny-all policy | WARNING |
| GlobalNetworkPolicy without order | INFO |

### DNS
| Check | Severity |
|-------|---------|
| No DNS pods found | CRITICAL |
| All DNS pods not running | CRITICAL |
| Some DNS pods not running | WARNING |

## Scheduling / Affinity (`checker/affinity.go`)

| Check | Severity |
|-------|---------|
| NodeSelector matches no nodes | CRITICAL |
| Required node affinity matches no nodes | CRITICAL |
| Pod on node with unmatched NoExecute/NoSchedule taint | WARNING |
| Wildcard toleration | WARNING |
| PodAntiAffinity term missing topologyKey | WARNING |
| TopologySpreadConstraint topologyKey not on any node | WARNING |

## Events (`checker/events.go`)

Events from the last **2 hours** are analysed. Warning events are deduped by
`(namespace, object, reason)`. Classified reasons include:

| Reason | Severity |
|--------|---------|
| OOMKilling | CRITICAL |
| BackOff | CRITICAL |
| Failed | CRITICAL |
| FailedMount / FailedAttach | CRITICAL |
| FailedScheduling | CRITICAL |
| Evicted / EvictionThresholdMet | CRITICAL |
| NodeNotReady | CRITICAL |
| NetworkNotReady | CRITICAL |
| FailedCreatePodSandBox | CRITICAL |
| Unhealthy | WARNING |
| ProbeWarning | WARNING |
| NodeHasDiskPressure / NodeHasMemoryPressure | WARNING |
| TopologyAffinityError | WARNING |

## Resources & HPA (`checker/resources.go`)

| Check | Severity |
|-------|---------|
| ResourceQuota at 100% | CRITICAL |
| HPA unable to scale | CRITICAL |
| ResourceQuota at 85%+ | WARNING |
| HPA at maximum replicas | WARNING |
| No LimitRange configured | INFO |

## RBAC (`checker/resources.go`)

| Check | Severity |
|-------|---------|
| ServiceAccount bound to cluster-admin | CRITICAL |
| User/Group bound to cluster-admin (non-kube-system) | WARNING |
| ClusterRole with wildcard verbs or resources | WARNING |

## Config / Secrets (`checker/resources.go`)

| Check | Severity |
|-------|---------|
| Pod references missing Secret (volume) | CRITICAL |
| Container references missing Secret (envFrom) | CRITICAL |
| Orphaned ConfigMap | INFO |

## Ingress (`checker/resources.go`)

| Check | Severity |
|-------|---------|
| Ingress references missing IngressClass | CRITICAL |
| Ingress references missing backend Service | CRITICAL |

## Namespace (`checker/resources.go`)

| Check | Severity |
|-------|---------|
| Namespace stuck in Terminating | CRITICAL (scoped) / WARNING (cluster) |
