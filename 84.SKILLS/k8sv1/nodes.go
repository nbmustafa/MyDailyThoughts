package checker

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// NodeChecker checks node health, capacity and taint issues
type NodeChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewNodeChecker(client kubernetes.Interface, namespace string) *NodeChecker {
	return &NodeChecker{client: client, namespace: namespace}
}

func (c *NodeChecker) Category() types.Category { return types.CategoryNodes }

func (c *NodeChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	// Node checks are always cluster-scoped
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to list nodes: %w", err)
		return result
	}

	for _, node := range nodes.Items {
		result.Findings = append(result.Findings, c.checkNode(ctx, node)...)
	}

	// Cluster-level: check if all masters are reachable
	result.Findings = append(result.Findings, c.checkControlPlane(nodes.Items)...)

	result.Duration = time.Since(start)
	return result
}

func (c *NodeChecker) checkNode(ctx context.Context, node corev1.Node) []types.Finding {
	var findings []types.Finding

	// ── Ready condition ──────────────────────────────────────────────────────────
	for _, cond := range node.Status.Conditions {
		switch cond.Type {
		case corev1.NodeReady:
			if cond.Status != corev1.ConditionTrue {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node NotReady",
					Description: fmt.Sprintf("Node %s is NotReady: %s", node.Name, cond.Message),
					Resource:    node.Name,
					Suggestion:  "Check kubelet logs on the node: `journalctl -u kubelet`. Verify network, disk and memory pressure",
				})
			}
		case corev1.NodeMemoryPressure:
			if cond.Status == corev1.ConditionTrue {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node under MemoryPressure",
					Description: fmt.Sprintf("Node %s is experiencing memory pressure", node.Name),
					Resource:    node.Name,
					Suggestion:  "Reduce pod density on this node or increase memory. The kubelet may begin evicting pods",
				})
			}
		case corev1.NodeDiskPressure:
			if cond.Status == corev1.ConditionTrue {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node under DiskPressure",
					Description: fmt.Sprintf("Node %s is experiencing disk pressure", node.Name),
					Resource:    node.Name,
					Suggestion:  "Free disk space: clean up unused images (`crictl rmi --prune`), logs or pods. Check storage mounts",
				})
			}
		case corev1.NodePIDPressure:
			if cond.Status == corev1.ConditionTrue {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node under PID Pressure",
					Description: fmt.Sprintf("Node %s is experiencing PID pressure", node.Name),
					Resource:    node.Name,
					Suggestion:  "Too many processes — reduce pod count or increase the PID limit on the node",
				})
			}
		case corev1.NodeNetworkUnavailable:
			if cond.Status == corev1.ConditionTrue {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node network unavailable",
					Description: fmt.Sprintf("Node %s network is marked unavailable: %s", node.Name, cond.Message),
					Resource:    node.Name,
					Suggestion:  "Check CNI plugin (Calico/Flannel etc.) and node networking configuration",
				})
			}
		}
	}

	// ── Cordoned / unschedulable ─────────────────────────────────────────────────
	if node.Spec.Unschedulable {
		findings = append(findings, types.Finding{
			Category:    types.CategoryNodes,
			Severity:    types.SeverityWarning,
			Title:       "Node cordoned (unschedulable)",
			Description: fmt.Sprintf("Node %s is cordoned — no new pods will be scheduled", node.Name),
			Resource:    node.Name,
			Suggestion:  "If not intentional, run `kubectl uncordon %s` to re-enable scheduling",
		})
	}

	// ── Capacity thresholds ──────────────────────────────────────────────────────
	findings = append(findings, c.checkCapacity(node)...)

	// ── Kernel version / OS ──────────────────────────────────────────────────────
	if node.Status.NodeInfo.KernelVersion == "" {
		findings = append(findings, types.Finding{
			Category:    types.CategoryNodes,
			Severity:    types.SeverityInfo,
			Title:       "Node kernel info unavailable",
			Description: fmt.Sprintf("Node %s is not reporting kernel version", node.Name),
			Resource:    node.Name,
			Suggestion:  "Node may be unreachable or kubelet may not be reporting system info",
		})
	}

	// ── NotReady age ────────────────────────────────────────────────────────────
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status != corev1.ConditionTrue {
			notReadyFor := time.Since(cond.LastTransitionTime.Time)
			if notReadyFor > 5*time.Minute {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNodes,
					Severity:    types.SeverityCritical,
					Title:       "Node NotReady for extended period",
					Description: fmt.Sprintf("Node %s has been NotReady for %.0f minutes", node.Name, notReadyFor.Minutes()),
					Resource:    node.Name,
					Suggestion:  "Consider draining and replacing the node. Pods on this node may be stuck in Terminating state",
				})
			}
		}
	}

	return findings
}

func (c *NodeChecker) checkCapacity(node corev1.Node) []types.Finding {
	var findings []types.Finding

	// Allocatable vs capacity — if allocatable is < 80% of capacity something is consuming headroom
	capCPU := node.Status.Capacity.Cpu()
	allocCPU := node.Status.Allocatable.Cpu()
	if capCPU != nil && allocCPU != nil && capCPU.Cmp(resource.MustParse("0")) > 0 {
		cpuRatio := float64(allocCPU.MilliValue()) / float64(capCPU.MilliValue())
		if cpuRatio < 0.5 {
			findings = append(findings, types.Finding{
				Category:    types.CategoryNodes,
				Severity:    types.SeverityWarning,
				Title:       "Low allocatable CPU headroom",
				Description: fmt.Sprintf("Node %s: only %.0f%% of CPU is allocatable (system/kube reserved may be high)", node.Name, cpuRatio*100),
				Resource:    node.Name,
				Suggestion:  "Review kube-reserved and system-reserved kubelet flags",
			})
		}
	}

	capMem := node.Status.Capacity.Memory()
	allocMem := node.Status.Allocatable.Memory()
	if capMem != nil && allocMem != nil && capMem.Cmp(resource.MustParse("0")) > 0 {
		memRatio := float64(allocMem.Value()) / float64(capMem.Value())
		if memRatio < 0.5 {
			findings = append(findings, types.Finding{
				Category:    types.CategoryNodes,
				Severity:    types.SeverityWarning,
				Title:       "Low allocatable memory headroom",
				Description: fmt.Sprintf("Node %s: only %.0f%% of memory is allocatable", node.Name, memRatio*100),
				Resource:    node.Name,
				Suggestion:  "Review kube-reserved and system-reserved memory settings",
			})
		}
	}

	return findings
}

func (c *NodeChecker) checkControlPlane(nodes []corev1.Node) []types.Finding {
	var findings []types.Finding
	masterCount := 0
	readyMasters := 0

	for _, node := range nodes {
		_, isMaster := node.Labels["node-role.kubernetes.io/master"]
		_, isControlPlane := node.Labels["node-role.kubernetes.io/control-plane"]
		if isMaster || isControlPlane {
			masterCount++
			for _, cond := range node.Status.Conditions {
				if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
					readyMasters++
				}
			}
		}
	}

	if masterCount > 0 && readyMasters < masterCount {
		findings = append(findings, types.Finding{
			Category:    types.CategoryNodes,
			Severity:    types.SeverityCritical,
			Title:       "Control plane node(s) not ready",
			Description: fmt.Sprintf("%d/%d control plane nodes are ready", readyMasters, masterCount),
			Resource:    "cluster/control-plane",
			Suggestion:  "Investigate control plane node health immediately — API server, etcd and scheduler may be impacted",
		})
	}

	if masterCount == 1 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryNodes,
			Severity:    types.SeverityInfo,
			Title:       "Single control plane node (no HA)",
			Description: "Cluster has only one control plane node — no high availability",
			Resource:    "cluster/control-plane",
			Suggestion:  "Consider adding additional control plane nodes for production resilience",
		})
	}

	return findings
}
