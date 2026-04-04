package checker

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// ── Resource Quota Checker ────────────────────────────────────────────────────

type ResourceChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewResourceChecker(client kubernetes.Interface, namespace string) *ResourceChecker {
	return &ResourceChecker{client: client, namespace: namespace}
}

func (c *ResourceChecker) Category() types.Category { return types.CategoryResources }

func (c *ResourceChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	result.Findings = append(result.Findings, c.checkResourceQuotas(ctx)...)
	result.Findings = append(result.Findings, c.checkLimitRanges(ctx)...)
	result.Findings = append(result.Findings, c.checkHPA(ctx)...)

	result.Duration = time.Since(start)
	return result
}

func (c *ResourceChecker) checkResourceQuotas(ctx context.Context) []types.Finding {
	var findings []types.Finding

	quotas, err := c.client.CoreV1().ResourceQuotas(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	for _, quota := range quotas.Items {
		ref := fmt.Sprintf("%s/%s", quota.Namespace, quota.Name)

		for resource, hard := range quota.Status.Hard {
			used, hasUsed := quota.Status.Used[resource]
			if !hasUsed {
				continue
			}

			hardVal := hard.Value()
			usedVal := used.Value()

			if hardVal == 0 {
				continue
			}

			ratio := float64(usedVal) / float64(hardVal)

			if ratio >= 1.0 {
				findings = append(findings, types.Finding{
					Category:    types.CategoryResources,
					Severity:    types.SeverityCritical,
					Title:       "Resource quota exhausted",
					Description: fmt.Sprintf("Quota %s: resource %q is at %d/%d (100%%) — new pods/resources will be rejected", ref, resource, usedVal, hardVal),
					Resource:    ref,
					Namespace:   quota.Namespace,
					Suggestion:  "Delete unused resources or increase the quota limit. New deployments will fail until quota is freed",
				})
			} else if ratio >= 0.85 {
				findings = append(findings, types.Finding{
					Category:    types.CategoryResources,
					Severity:    types.SeverityWarning,
					Title:       "Resource quota near limit",
					Description: fmt.Sprintf("Quota %s: resource %q is at %d/%d (%.0f%%)", ref, resource, usedVal, hardVal, ratio*100),
					Resource:    ref,
					Namespace:   quota.Namespace,
					Suggestion:  "Consider increasing quota or cleaning up unused resources before hitting the limit",
				})
			}
		}
	}

	return findings
}

func (c *ResourceChecker) checkLimitRanges(ctx context.Context) []types.Finding {
	var findings []types.Finding

	lrs, err := c.client.CoreV1().LimitRanges(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	ns := c.namespace
	if ns == "" {
		ns = "all namespaces"
	}

	if len(lrs.Items) == 0 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryResources,
			Severity:    types.SeverityInfo,
			Title:       "No LimitRange configured",
			Description: fmt.Sprintf("No LimitRange found in %s", ns),
			Resource:    ns,
			Namespace:   c.namespace,
			Suggestion:  "Consider adding a LimitRange to set default resource requests/limits for pods",
		})
	}

	return findings
}

func (c *ResourceChecker) checkHPA(ctx context.Context) []types.Finding {
	var findings []types.Finding

	hpas, err := c.client.AutoscalingV2().HorizontalPodAutoscalers(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// Try v1
		return findings
	}

	for _, hpa := range hpas.Items {
		ref := fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)

		// HPA at max replicas
		if hpa.Status.CurrentReplicas >= hpa.Spec.MaxReplicas {
			findings = append(findings, types.Finding{
				Category:    types.CategoryResources,
				Severity:    types.SeverityWarning,
				Title:       "HPA at maximum replicas",
				Description: fmt.Sprintf("HPA %s is at max replicas (%d/%d) — cannot scale further", ref, hpa.Status.CurrentReplicas, hpa.Spec.MaxReplicas),
				Resource:    ref,
				Namespace:   hpa.Namespace,
				Suggestion:  "Increase HPA maxReplicas or optimise application performance to reduce load",
			})
		}

		// HPA unable to scale (conditions)
		for _, cond := range hpa.Status.Conditions {
			if cond.Type == "AbleToScale" && cond.Status == corev1.ConditionFalse {
				findings = append(findings, types.Finding{
					Category:    types.CategoryResources,
					Severity:    types.SeverityCritical,
					Title:       "HPA unable to scale",
					Description: fmt.Sprintf("HPA %s cannot scale: %s", ref, cond.Message),
					Resource:    ref,
					Namespace:   hpa.Namespace,
					Suggestion:  "Check metrics-server availability and that the target deployment exists",
				})
			}
		}
	}

	return findings
}

// ── RBAC Checker ─────────────────────────────────────────────────────────────

type RBACChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewRBACChecker(client kubernetes.Interface, namespace string) *RBACChecker {
	return &RBACChecker{client: client, namespace: namespace}
}

func (c *RBACChecker) Category() types.Category { return types.CategoryRBAC }

func (c *RBACChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	result.Findings = append(result.Findings, c.checkClusterAdminBindings(ctx)...)
	result.Findings = append(result.Findings, c.checkWildcardRoles(ctx)...)

	result.Duration = time.Since(start)
	return result
}

func (c *RBACChecker) checkClusterAdminBindings(ctx context.Context) []types.Finding {
	var findings []types.Finding

	crbs, err := c.client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	for _, crb := range crbs.Items {
		if crb.RoleRef.Name != "cluster-admin" {
			continue
		}

		for _, subject := range crb.Subjects {
			// Skip system service accounts that legitimately need cluster-admin
			if subject.Namespace == "kube-system" {
				continue
			}

			sev := types.SeverityWarning
			if subject.Kind == "ServiceAccount" {
				sev = types.SeverityCritical
			}

			findings = append(findings, types.Finding{
				Category:    types.CategoryRBAC,
				Severity:    sev,
				Title:       "cluster-admin binding detected",
				Description: fmt.Sprintf("ClusterRoleBinding %q grants cluster-admin to %s %s/%s", crb.Name, subject.Kind, subject.Namespace, subject.Name),
				Resource:    crb.Name,
				Suggestion:  "Apply least-privilege principle — grant only the specific roles needed",
			})
		}
	}

	return findings
}

func (c *RBACChecker) checkWildcardRoles(ctx context.Context) []types.Finding {
	var findings []types.Finding

	// Check ClusterRoles
	crs, err := c.client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	for _, cr := range crs.Items {
		if isSystemRole(cr.Name) {
			continue
		}
		for _, rule := range cr.Rules {
			if containsWildcard(rule.Verbs) || containsWildcard(rule.Resources) {
				findings = append(findings, types.Finding{
					Category:    types.CategoryRBAC,
					Severity:    types.SeverityWarning,
					Title:       "ClusterRole uses wildcard permissions",
					Description: fmt.Sprintf("ClusterRole %q has wildcard (*) in verbs or resources", cr.Name),
					Resource:    cr.Name,
					Suggestion:  "Replace wildcards with explicit verbs and resources to enforce least-privilege",
				})
				break
			}
		}
	}

	return findings
}

func isSystemRole(name string) bool {
	prefixes := []string{"system:", "kubeadm:", "calico"}
	for _, p := range prefixes {
		if len(name) >= len(p) && name[:len(p)] == p {
			return true
		}
	}
	return false
}

func containsWildcard(list []string) bool {
	for _, s := range list {
		if s == "*" {
			return true
		}
	}
	return false
}

// ── Config / Secrets Checker ──────────────────────────────────────────────────

type ConfigChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewConfigChecker(client kubernetes.Interface, namespace string) *ConfigChecker {
	return &ConfigChecker{client: client, namespace: namespace}
}

func (c *ConfigChecker) Category() types.Category { return types.CategoryConfig }

func (c *ConfigChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	result.Findings = append(result.Findings, c.checkOrphanedConfigMaps(ctx)...)
	result.Findings = append(result.Findings, c.checkMissingSecrets(ctx)...)

	result.Duration = time.Since(start)
	return result
}

func (c *ConfigChecker) checkOrphanedConfigMaps(ctx context.Context) []types.Finding {
	var findings []types.Finding

	cms, err := c.client.CoreV1().ConfigMaps(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	// Build set of referenced ConfigMaps
	referencedCMs := map[string]bool{}
	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.ConfigMap != nil {
				referencedCMs[vol.Namespace+"/"+vol.ConfigMap.Name] = true
			}
		}
		for _, c := range pod.Spec.Containers {
			for _, envFrom := range c.EnvFrom {
				if envFrom.ConfigMapRef != nil {
					referencedCMs[pod.Namespace+"/"+envFrom.ConfigMapRef.Name] = true
				}
			}
		}
	}

	for _, cm := range cms.Items {
		key := cm.Namespace + "/" + cm.Name
		if !referencedCMs[key] && !isSystemConfigMap(cm.Name) {
			findings = append(findings, types.Finding{
				Category:    types.CategoryConfig,
				Severity:    types.SeverityInfo,
				Title:       "Potentially orphaned ConfigMap",
				Description: fmt.Sprintf("ConfigMap %s/%s is not referenced by any pod in the namespace", cm.Namespace, cm.Name),
				Resource:    key,
				Namespace:   cm.Namespace,
				Suggestion:  "Verify if this ConfigMap is still needed. Orphaned ConfigMaps add noise and consume quota",
			})
		}
	}

	return findings
}

func (c *ConfigChecker) checkMissingSecrets(ctx context.Context) []types.Finding {
	var findings []types.Finding

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	secrets, err := c.client.CoreV1().Secrets(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	existingSecrets := map[string]bool{}
	for _, s := range secrets.Items {
		existingSecrets[s.Namespace+"/"+s.Name] = true
	}

	for _, pod := range pods.Items {
		ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		for _, vol := range pod.Spec.Volumes {
			if vol.Secret != nil {
				key := pod.Namespace + "/" + vol.Secret.SecretName
				if !existingSecrets[key] {
					findings = append(findings, types.Finding{
						Category:    types.CategoryConfig,
						Severity:    types.SeverityCritical,
						Title:       "Pod references missing Secret",
						Description: fmt.Sprintf("Pod %s references secret %q which does not exist", ref, vol.Secret.SecretName),
						Resource:    ref,
						Namespace:   pod.Namespace,
						Suggestion:  "Create the missing secret or update the pod spec to reference an existing secret",
					})
				}
			}
		}

		for _, container := range pod.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.SecretRef != nil {
					key := pod.Namespace + "/" + envFrom.SecretRef.Name
					if !existingSecrets[key] {
						findings = append(findings, types.Finding{
							Category:    types.CategoryConfig,
							Severity:    types.SeverityCritical,
							Title:       "Container references missing Secret",
							Description: fmt.Sprintf("Container %q in pod %s references missing secret %q via envFrom", container.Name, ref, envFrom.SecretRef.Name),
							Resource:    ref,
							Namespace:   pod.Namespace,
							Suggestion:  "Create the missing secret or remove the envFrom reference",
						})
					}
				}
			}
		}
	}

	return findings
}

func isSystemConfigMap(name string) bool {
	systemNames := map[string]bool{
		"kube-root-ca.crt": true,
		"kubeadm-config":   true,
		"kubelet-config":   true,
	}
	return systemNames[name]
}

// ── Namespace Checker ─────────────────────────────────────────────────────────

type NamespaceChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewNamespaceChecker(client kubernetes.Interface, namespace string) *NamespaceChecker {
	return &NamespaceChecker{client: client, namespace: namespace}
}

func (c *NamespaceChecker) Category() types.Category { return types.CategoryNamespace }

func (c *NamespaceChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	if c.namespace != "" {
		// Check specific namespace health
		ns, err := c.client.CoreV1().Namespaces().Get(ctx, c.namespace, metav1.GetOptions{})
		if err != nil {
			result.Error = fmt.Errorf("namespace %q not found: %w", c.namespace, err)
			return result
		}

		if ns.Status.Phase == corev1.NamespaceTerminating {
			result.Findings = append(result.Findings, types.Finding{
				Category:    types.CategoryNamespace,
				Severity:    types.SeverityCritical,
				Title:       "Namespace stuck in Terminating",
				Description: fmt.Sprintf("Namespace %q has been in Terminating state", ns.Name),
				Resource:    ns.Name,
				Suggestion:  "Check for finalizers: `kubectl get ns %s -o json | jq .spec.finalizers`. Remove stuck finalizers if needed",
			})
		}
	} else {
		// Cluster-wide: check all namespaces
		nsList, err := c.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			result.Error = err
			return result
		}

		for _, ns := range nsList.Items {
			if ns.Status.Phase == corev1.NamespaceTerminating {
				stuckFor := time.Since(ns.DeletionTimestamp.Time)
				result.Findings = append(result.Findings, types.Finding{
					Category:    types.CategoryNamespace,
					Severity:    types.SeverityWarning,
					Title:       "Namespace stuck in Terminating",
					Description: fmt.Sprintf("Namespace %q has been Terminating for %.0f minutes", ns.Name, stuckFor.Minutes()),
					Resource:    ns.Name,
					Suggestion:  "Check for custom finalizers blocking namespace deletion",
				})
			}
		}
	}

	result.Duration = time.Since(start)
	return result
}

// ── Ingress Checker ───────────────────────────────────────────────────────────

type IngressChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewIngressChecker(client kubernetes.Interface, namespace string) *IngressChecker {
	return &IngressChecker{client: client, namespace: namespace}
}

func (c *IngressChecker) Category() types.Category { return types.CategoryIngress }

func (c *IngressChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	ingresses, err := c.client.NetworkingV1().Ingresses(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Duration = time.Since(start)
		return result
	}

	for _, ing := range ingresses.Items {
		ref := fmt.Sprintf("%s/%s", ing.Namespace, ing.Name)

		// Check if IngressClass exists
		if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName != "" {
			_, err := c.client.NetworkingV1().IngressClasses().Get(ctx, *ing.Spec.IngressClassName, metav1.GetOptions{})
			if err != nil {
				result.Findings = append(result.Findings, types.Finding{
					Category:    types.CategoryIngress,
					Severity:    types.SeverityCritical,
					Title:       "Ingress references missing IngressClass",
					Description: fmt.Sprintf("Ingress %s references IngressClass %q which does not exist", ref, *ing.Spec.IngressClassName),
					Resource:    ref,
					Namespace:   ing.Namespace,
					Suggestion:  "Create the IngressClass or update the Ingress to reference an existing one",
				})
			}
		}

		// Check backend service exists
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service == nil {
					continue
				}
				svcName := path.Backend.Service.Name
				_, err := c.client.CoreV1().Services(ing.Namespace).Get(ctx, svcName, metav1.GetOptions{})
				if err != nil {
					result.Findings = append(result.Findings, types.Finding{
						Category:    types.CategoryIngress,
						Severity:    types.SeverityCritical,
						Title:       "Ingress backend service not found",
						Description: fmt.Sprintf("Ingress %s references backend service %q which does not exist in namespace %s", ref, svcName, ing.Namespace),
						Resource:    ref,
						Namespace:   ing.Namespace,
						Suggestion:  "Create the missing service or update the Ingress backend reference",
					})
				}
			}
		}
	}

	result.Duration = time.Since(start)
	return result
}

// Ensure rbacv1 import is used
var _ = rbacv1.ClusterRoleBinding{}
