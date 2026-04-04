package checker

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// PodChecker checks all pod-related issues
type PodChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewPodChecker(client kubernetes.Interface, namespace string) *PodChecker {
	return &PodChecker{client: client, namespace: namespace}
}

func (c *PodChecker) Category() types.Category { return types.CategoryPods }

func (c *PodChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to list pods: %w", err)
		return result
	}

	for _, pod := range pods.Items {
		result.Findings = append(result.Findings, c.checkPod(pod)...)
	}

	result.Duration = time.Since(start)
	return result
}

func (c *PodChecker) checkPod(pod corev1.Pod) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	// ── Phase checks ────────────────────────────────────────────────────────────
	switch pod.Status.Phase {
	case corev1.PodFailed:
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityCritical,
			Title:       "Pod in Failed phase",
			Description: fmt.Sprintf("Pod %s is in Failed state", ref),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Run `kubectl describe pod` and `kubectl logs` to investigate root cause",
		})
	case corev1.PodUnknown:
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityCritical,
			Title:       "Pod in Unknown phase",
			Description: fmt.Sprintf("Pod %s is in Unknown state — node communication may be lost", ref),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Check node health and kubelet status on the hosting node",
		})
	}

	// ── Container status checks ─────────────────────────────────────────────────
	for _, cs := range pod.Status.ContainerStatuses {
		findings = append(findings, c.checkContainerStatus(pod, cs)...)
	}
	for _, cs := range pod.Status.InitContainerStatuses {
		findings = append(findings, c.checkInitContainerStatus(pod, cs)...)
	}

	// ── Pending pod analysis ────────────────────────────────────────────────────
	if pod.Status.Phase == corev1.PodPending {
		findings = append(findings, c.checkPendingPod(pod)...)
	}

	// ── Resource limits ──────────────────────────────────────────────────────────
	for _, container := range pod.Spec.Containers {
		if container.Resources.Limits == nil {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityWarning,
				Title:       "Container missing resource limits",
				Description: fmt.Sprintf("Container %q in pod %s has no resource limits set", container.Name, ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Set CPU and memory limits to prevent resource starvation across pods",
			})
		}
		if container.Resources.Requests == nil {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityInfo,
				Title:       "Container missing resource requests",
				Description: fmt.Sprintf("Container %q in pod %s has no resource requests set", container.Name, ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Set resource requests for accurate scheduling decisions",
			})
		}
	}

	// ── Liveness / Readiness probes ─────────────────────────────────────────────
	for _, container := range pod.Spec.Containers {
		if container.LivenessProbe == nil {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityInfo,
				Title:       "Container missing liveness probe",
				Description: fmt.Sprintf("Container %q in pod %s has no liveness probe", container.Name, ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Add a liveness probe so Kubernetes can restart stuck containers",
			})
		}
		if container.ReadinessProbe == nil {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityInfo,
				Title:       "Container missing readiness probe",
				Description: fmt.Sprintf("Container %q in pod %s has no readiness probe", container.Name, ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Add a readiness probe so traffic is not routed to unready pods",
			})
		}
	}

	// ── Security context ─────────────────────────────────────────────────────────
	if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsNonRoot == nil {
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityInfo,
			Title:       "Pod may run as root",
			Description: fmt.Sprintf("Pod %s does not explicitly set runAsNonRoot", ref),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Set securityContext.runAsNonRoot: true to enforce least-privilege",
		})
	}

	// ── Image policy ─────────────────────────────────────────────────────────────
	for _, container := range pod.Spec.Containers {
		if strings.HasSuffix(container.Image, ":latest") || !strings.Contains(container.Image, ":") {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityWarning,
				Title:       "Container using latest or untagged image",
				Description: fmt.Sprintf("Container %q in pod %s uses image %q", container.Name, ref, container.Image),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Pin image tags to specific digests for reproducible deployments",
			})
		}
	}

	return findings
}

func (c *PodChecker) checkContainerStatus(pod corev1.Pod, cs corev1.ContainerStatus) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s[%s]", pod.Namespace, pod.Name, cs.Name)

	// CrashLoopBackOff
	if cs.State.Waiting != nil {
		switch cs.State.Waiting.Reason {
		case "CrashLoopBackOff":
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityCritical,
				Title:       "CrashLoopBackOff detected",
				Description: fmt.Sprintf("Container %s is in CrashLoopBackOff (restart count: %d)", ref, cs.RestartCount),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Check logs: `kubectl logs -p` for previous crash logs. Common causes: wrong command, missing env vars, config errors",
			})
		case "OOMKilled":
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityCritical,
				Title:       "Container OOMKilled",
				Description: fmt.Sprintf("Container %s was OOMKilled — memory limit exceeded", ref),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Increase memory limit or optimise application memory usage",
			})
		case "ImagePullBackOff", "ErrImagePull":
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityCritical,
				Title:       "Image pull failure",
				Description: fmt.Sprintf("Container %s cannot pull image: %s", ref, cs.State.Waiting.Message),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Check image name/tag, registry credentials (imagePullSecrets), and network access to registry",
			})
		case "CreateContainerConfigError":
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    types.SeverityCritical,
				Title:       "Container config error",
				Description: fmt.Sprintf("Container %s: %s", ref, cs.State.Waiting.Message),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Check referenced ConfigMaps and Secrets exist in the namespace",
			})
		}
	}

	// High restart count
	if cs.RestartCount > 5 {
		sev := types.SeverityWarning
		if cs.RestartCount > 20 {
			sev = types.SeverityCritical
		}
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    sev,
			Title:       "High container restart count",
			Description: fmt.Sprintf("Container %s has restarted %d times", ref, cs.RestartCount),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Investigate crash logs and liveness probe configuration",
		})
	}

	// Not ready
	if !cs.Ready && pod.Status.Phase == corev1.PodRunning {
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityWarning,
			Title:       "Container running but not ready",
			Description: fmt.Sprintf("Container %s is running but readiness probe is failing", ref),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Check readiness probe endpoint and application startup time",
		})
	}

	return findings
}

func (c *PodChecker) checkInitContainerStatus(pod corev1.Pod, cs corev1.ContainerStatus) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s[init:%s]", pod.Namespace, pod.Name, cs.Name)

	if !cs.Ready && cs.State.Waiting != nil {
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityCritical,
			Title:       "Init container not completing",
			Description: fmt.Sprintf("Init container %s is stuck: %s — %s", ref, cs.State.Waiting.Reason, cs.State.Waiting.Message),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Init containers must complete successfully before main containers start. Check dependencies (DBs, external services)",
		})
	}

	if cs.RestartCount > 3 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    types.SeverityCritical,
			Title:       "Init container repeatedly failing",
			Description: fmt.Sprintf("Init container %s has failed %d times", ref, cs.RestartCount),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Check init container logs and the availability of any prerequisite services",
		})
	}

	return findings
}

func (c *PodChecker) checkPendingPod(pod corev1.Pod) []types.Finding {
	var findings []types.Finding
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	pendingDuration := time.Since(pod.CreationTimestamp.Time)
	if pendingDuration < 2*time.Minute {
		return findings // too early to flag
	}

	sev := types.SeverityWarning
	if pendingDuration > 15*time.Minute {
		sev = types.SeverityCritical
	}

	// Check pod conditions for scheduling clues
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
			findings = append(findings, types.Finding{
				Category:    types.CategoryPods,
				Severity:    sev,
				Title:       "Pod unschedulable",
				Description: fmt.Sprintf("Pod %s cannot be scheduled (pending %.0fm): %s", ref, pendingDuration.Minutes(), cond.Message),
				Resource:    ref,
				Namespace:   pod.Namespace,
				Suggestion:  "Check node capacity, taints/tolerations, node selectors, and affinity rules. Use `kubectl describe pod` for scheduler messages",
			})
		}
	}

	// No conditions populated yet — generic pending warning
	if len(pod.Status.Conditions) == 0 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryPods,
			Severity:    sev,
			Title:       "Pod pending with no conditions",
			Description: fmt.Sprintf("Pod %s has been Pending for %.0f minutes with no scheduler feedback", ref, pendingDuration.Minutes()),
			Resource:    ref,
			Namespace:   pod.Namespace,
			Suggestion:  "Check if the scheduler is running and there are available nodes",
		})
	}

	return findings
}
