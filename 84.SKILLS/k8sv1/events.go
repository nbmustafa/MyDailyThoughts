package checker

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// EventChecker analyses Kubernetes events for warnings and patterns
type EventChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewEventChecker(client kubernetes.Interface, namespace string) *EventChecker {
	return &EventChecker{client: client, namespace: namespace}
}

func (c *EventChecker) Category() types.Category { return types.CategoryEvents }

func (c *EventChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	// Fetch events — scoped or cluster-wide
	ns := c.namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}

	events, err := c.client.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to list events: %w", err)
		return result
	}

	// Sort by last timestamp descending
	sort.Slice(events.Items, func(i, j int) bool {
		ti := events.Items[i].LastTimestamp.Time
		tj := events.Items[j].LastTimestamp.Time
		return ti.After(tj)
	})

	result.Findings = append(result.Findings, c.analyseEvents(events.Items)...)
	result.Duration = time.Since(start)
	return result
}

func (c *EventChecker) analyseEvents(events []corev1.Event) []types.Finding {
	var findings []types.Finding

	// Deduplicate by (involvedObject + reason + message) to avoid noise
	type eventKey struct {
		Namespace string
		ObjName   string
		ObjKind   string
		Reason    string
	}
	seen := map[eventKey]bool{}

	// Only look at events from the last 2 hours
	cutoff := time.Now().Add(-2 * time.Hour)

	for _, evt := range events {
		if evt.Type == corev1.EventTypeNormal {
			continue // skip normal events
		}

		lastTime := evt.LastTimestamp.Time
		if lastTime.IsZero() {
			lastTime = evt.EventTime.Time
		}
		if lastTime.Before(cutoff) {
			continue
		}

		key := eventKey{
			Namespace: evt.Namespace,
			ObjName:   evt.InvolvedObject.Name,
			ObjKind:   evt.InvolvedObject.Kind,
			Reason:    evt.Reason,
		}

		if seen[key] {
			continue
		}
		seen[key] = true

		ref := fmt.Sprintf("%s/%s(%s)", evt.Namespace, evt.InvolvedObject.Name, evt.InvolvedObject.Kind)
		sev, suggestion := c.classifyEvent(evt)

		findings = append(findings, types.Finding{
			Category:    types.CategoryEvents,
			Severity:    sev,
			Title:       fmt.Sprintf("Event: %s", evt.Reason),
			Description: fmt.Sprintf("[%s] %s — %s (count: %d)", ref, evt.Reason, evt.Message, evt.Count),
			Resource:    ref,
			Namespace:   evt.Namespace,
			Suggestion:  suggestion,
			Timestamp:   lastTime,
		})
	}

	return findings
}

func (c *EventChecker) classifyEvent(evt corev1.Event) (types.Severity, string) {
	reason := evt.Reason

	criticalReasons := map[string]string{
		"OOMKilling":                  "Increase memory limits or fix memory leak in the application",
		"BackOff":                     "Container is crash-looping. Check logs with `kubectl logs -p`",
		"Failed":                      "Check pod/container events and logs for root cause",
		"FailedCreate":                "Controller failed to create resource — check RBAC and quota",
		"FailedMount":                 "Volume mount failed — check PVC status and StorageClass provisioner",
		"FailedAttach":                "Volume attach failed — check node and cloud provider CSI driver",
		"FailedScheduling":            "No eligible nodes — check resources, taints, affinity and namespace quotas",
		"Evicted":                     "Pod was evicted — check node pressure conditions",
		"EvictionThresholdMet":        "Node is under pressure and may evict pods — check disk/memory",
		"NodeNotReady":                "Node is NotReady — investigate node health",
		"NetworkNotReady":             "CNI plugin may be failing — check network pod health",
		"FailedCreatePodSandBox":      "Pod sandbox creation failed — check containerd/CRI and CNI",
		"FailedKillPod":               "Pod could not be killed cleanly — manual cleanup may be needed",
		"ExceededGracePeriod":         "Pod exceeded termination grace period — check for stuck processes",
		"ContainerGCFailed":           "Container garbage collection failed — check disk space on node",
		"ImageGCFailed":               "Image GC failed — disk pressure likely on node",
		"FailedNodeAllocatableEnforcement": "Kubelet failed to enforce allocatable — check cgroup config",
	}

	warningReasons := map[string]string{
		"Pulling":             "Image is being pulled — may cause startup delay",
		"Unhealthy":           "Probe failing — check liveness/readiness probe configuration",
		"ProbeWarning":        "Probe is slow to respond — consider increasing timeoutSeconds",
		"NodeHasDiskPressure": "Reduce pod density or free disk space on the node",
		"NodeHasMemoryPressure": "Reduce pod density or add memory to the node",
		"Rescheduled":         "Pod was rescheduled — may indicate node instability",
		"HostPortConflict":    "Two pods trying to bind the same host port — check DaemonSet specs",
		"MissingClusterDNS":   "CoreDNS may be missing or misconfigured",
		"FreeDiskSpaceFailed": "Failed to free disk space — manual intervention needed",
		"InvalidDiskCapacity": "Node disk capacity reporting issue — check kubelet",
		"TopologyAffinityError": "Topology spread constraint cannot be satisfied",
	}

	if suggestion, ok := criticalReasons[reason]; ok {
		return types.SeverityCritical, suggestion
	}
	if suggestion, ok := warningReasons[reason]; ok {
		return types.SeverityWarning, suggestion
	}

	// Default for unknown Warning events
	return types.SeverityWarning, "Investigate with `kubectl describe` on the affected resource"
}
