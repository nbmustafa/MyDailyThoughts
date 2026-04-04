package checker

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// AffinityChecker checks affinity, anti-affinity, node selectors and taints/tolerations
type AffinityChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewAffinityChecker(client kubernetes.Interface, namespace string) *AffinityChecker {
	return &AffinityChecker{client: client, namespace: namespace}
}

func (c *AffinityChecker) Category() types.Category { return types.CategoryAffinity }

func (c *AffinityChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to list pods: %w", err)
		return result
	}

	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to list nodes: %w", err)
		return result
	}

	for _, pod := range pods.Items {
		result.Findings = append(result.Findings, c.checkPodScheduling(pod, nodes.Items)...)
	}

	result.Duration = time.Since(start)
	return result
}

func (c *AffinityChecker) checkPodScheduling(pod corev1.Pod, nodes []corev1.Node) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	// ── NodeSelector validation ──────────────────────────────────────────────────
	if len(pod.Spec.NodeSelector) > 0 {
		matchingNodes := 0
		for _, node := range nodes {
			if nodeMatchesSelector(node, pod.Spec.NodeSelector) {
				matchingNodes++
			}
		}
		if matchingNodes == 0 {
			findings = append(findings, types.Finding{
				Category:    types.CategoryAffinity,
				Severity:    types.SeverityCritical,
				Title:       "NodeSelector matches no available nodes",
				Description: fmt.Sprintf("Pod %s nodeSelector %v does not match any node in the cluster", ref, pod.Spec.NodeSelector),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Check node labels with `kubectl get nodes --show-labels`. Ensure nodes have the required labels",
			})
		}
	}

	// ── Node Affinity validation ─────────────────────────────────────────────────
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		na := pod.Spec.Affinity.NodeAffinity
		if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			matchingNodes := 0
			for _, node := range nodes {
				if nodeMatchesAffinityTerms(node, na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) {
					matchingNodes++
				}
			}
			if matchingNodes == 0 {
				findings = append(findings, types.Finding{
					Category:    types.CategoryAffinity,
					Severity:    types.SeverityCritical,
					Title:       "Required node affinity matches no nodes",
					Description: fmt.Sprintf("Pod %s requiredNodeAffinity rules match 0 nodes — pod will never schedule", ref),
					Resource:    ref,
					Namespace:   pod.Namespace,
					Suggestion:  "Review node affinity rules and available node labels. Use `kubectl describe pod` for scheduler rejection reason",
				})
			}
		}
	}

	// ── Toleration checks ────────────────────────────────────────────────────────
	if pod.Spec.Tolerations != nil {
		findings = append(findings, c.checkTolerations(pod, nodes)...)
	}

	// ── Taint check: pods on tainted nodes ──────────────────────────────────────
	if pod.Spec.NodeName != "" {
		for _, node := range nodes {
			if node.Name == pod.Spec.NodeName {
				findings = append(findings, c.checkPodNodeTaints(pod, node)...)
				break
			}
		}
	}

	// ── Pod anti-affinity analysis ───────────────────────────────────────────────
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.PodAntiAffinity != nil {
		findings = append(findings, c.checkPodAntiAffinity(pod)...)
	}

	// ── TopologySpreadConstraints ────────────────────────────────────────────────
	if len(pod.Spec.TopologySpreadConstraints) > 0 {
		findings = append(findings, c.checkTopologySpread(pod, nodes)...)
	}

	return findings
}

func (c *AffinityChecker) checkTolerations(pod corev1.Pod, nodes []corev1.Node) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	// Warn on wildcard toleration (tolerates everything)
	for _, tol := range pod.Spec.Tolerations {
		if tol.Operator == corev1.TolerationOpExists && tol.Key == "" {
			findings = append(findings, types.Finding{
				Category:    types.CategoryAffinity,
				Severity:    types.SeverityWarning,
				Title:       "Pod has wildcard toleration",
				Description: fmt.Sprintf("Pod %s tolerates ALL taints — it may schedule on any node including tainted/restricted nodes", ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Scope tolerations to specific taint keys to avoid unexpected scheduling",
			})
			break
		}
	}

	return findings
}

func (c *AffinityChecker) checkPodNodeTaints(pod corev1.Pod, node corev1.Node) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	for _, taint := range node.Spec.Taints {
		if taint.Effect == corev1.TaintEffectNoExecute || taint.Effect == corev1.TaintEffectNoSchedule {
			if !podToleratesTaint(pod, taint) {
				findings = append(findings, types.Finding{
					Category:    types.CategoryAffinity,
					Severity:    types.SeverityWarning,
					Title:       "Pod on node with unmatched taint",
					Description: fmt.Sprintf("Pod %s on node %s has taint %s=%s:%s that pod does not tolerate", ref, node.Name, taint.Key, taint.Value, taint.Effect),
					Resource:    ref,
					Namespace:   pod.Namespace,
					Suggestion:  "Add a toleration for this taint or move the pod to an untainted node",
				})
			}
		}
	}

	return findings
}

func (c *AffinityChecker) checkPodAntiAffinity(pod corev1.Pod) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	paa := pod.Spec.Affinity.PodAntiAffinity
	if paa.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		for _, term := range paa.RequiredDuringSchedulingIgnoredDuringExecution {
			if term.TopologyKey == "" {
				findings = append(findings, types.Finding{
					Category:    types.CategoryAffinity,
					Severity:    types.SeverityWarning,
					Title:       "PodAntiAffinity term missing topologyKey",
					Description: fmt.Sprintf("Pod %s has a required pod anti-affinity term with empty topologyKey", ref),
					Resource:    ref,
					Namespace:   pod.Namespace,
					Suggestion:  "Set topologyKey (e.g. kubernetes.io/hostname or topology.kubernetes.io/zone) for anti-affinity to work correctly",
				})
			}
		}
	}

	return findings
}

func (c *AffinityChecker) checkTopologySpread(pod corev1.Pod, nodes []corev1.Node) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	for _, tsc := range pod.Spec.TopologySpreadConstraints {
		// Verify topologyKey exists on at least one node
		keyFound := false
		for _, node := range nodes {
			if _, ok := node.Labels[tsc.TopologyKey]; ok {
				keyFound = true
				break
			}
		}

		if !keyFound {
			findings = append(findings, types.Finding{
				Category:    types.CategoryAffinity,
				Severity:    types.SeverityWarning,
				Title:       "TopologySpreadConstraint topologyKey not found on any node",
				Description: fmt.Sprintf("Pod %s uses topologyKey %q but no nodes have this label", ref, tsc.TopologyKey),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  fmt.Sprintf("Add label %q to nodes or correct the topologyKey in the spread constraint", tsc.TopologyKey),
			})
		}
	}

	return findings
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func nodeMatchesSelector(node corev1.Node, selector map[string]string) bool {
	for k, v := range selector {
		if node.Labels[k] != v {
			return false
		}
	}
	return true
}

func nodeMatchesAffinityTerms(node corev1.Node, terms []corev1.NodeSelectorTerm) bool {
	for _, term := range terms {
		if nodeMatchesTerm(node, term) {
			return true
		}
	}
	return false
}

func nodeMatchesTerm(node corev1.Node, term corev1.NodeSelectorTerm) bool {
	for _, req := range term.MatchExpressions {
		switch req.Operator {
		case corev1.NodeSelectorOpIn:
			val, ok := node.Labels[req.Key]
			if !ok {
				return false
			}
			found := false
			for _, v := range req.Values {
				if v == val {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case corev1.NodeSelectorOpNotIn:
			val, ok := node.Labels[req.Key]
			if ok {
				for _, v := range req.Values {
					if v == val {
						return false
					}
				}
			}
		case corev1.NodeSelectorOpExists:
			if _, ok := node.Labels[req.Key]; !ok {
				return false
			}
		case corev1.NodeSelectorOpDoesNotExist:
			if _, ok := node.Labels[req.Key]; ok {
				return false
			}
		}
	}
	return true
}

func podToleratesTaint(pod corev1.Pod, taint corev1.Taint) bool {
	for _, tol := range pod.Spec.Tolerations {
		if tol.ToleratesTaint(&taint) {
			return true
		}
	}
	return false
}
