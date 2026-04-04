package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"github.com/your-org/k8sdiag/internal/types"
)

// NetworkChecker checks Services, Endpoints, NetworkPolicies and Calico policies
type NetworkChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewNetworkChecker(client kubernetes.Interface, namespace string) *NetworkChecker {
	return &NetworkChecker{client: client, namespace: namespace}
}

func (c *NetworkChecker) Category() types.Category { return types.CategoryNetwork }

func (c *NetworkChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	result.Findings = append(result.Findings, c.checkServices(ctx)...)
	result.Findings = append(result.Findings, c.checkNetworkPolicies(ctx)...)
	result.Findings = append(result.Findings, c.checkCalicoPolicies(ctx)...)
	result.Findings = append(result.Findings, c.checkDNS(ctx)...)

	result.Duration = time.Since(start)
	return result
}

// ── Services & Endpoints ─────────────────────────────────────────────────────

func (c *NetworkChecker) checkServices(ctx context.Context) []types.Finding {
	var findings []types.Finding

	services, err := c.client.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []types.Finding{{
			Category:    types.CategoryNetwork,
			Severity:    types.SeverityWarning,
			Title:       "Failed to list services",
			Description: err.Error(),
		}}
	}

	for _, svc := range services.Items {
		if svc.Name == "kubernetes" && svc.Namespace == "default" {
			continue // skip the default k8s service
		}
		ref := fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)

		// Check endpoints
		ep, err := c.client.CoreV1().Endpoints(svc.Namespace).Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			findings = append(findings, types.Finding{
				Category:    types.CategoryServices,
				Severity:    types.SeverityWarning,
				Title:       "Service has no Endpoints object",
				Description: fmt.Sprintf("Service %s has no corresponding Endpoints resource", ref),
				Resource:    ref,
				Namespace:   svc.Namespace,
				Suggestion:  "Ensure pods matching the service selector exist and are ready",
			})
			continue
		}

		readyAddrs := 0
		notReadyAddrs := 0
		for _, subset := range ep.Subsets {
			readyAddrs += len(subset.Addresses)
			notReadyAddrs += len(subset.NotReadyAddresses)
		}

		if readyAddrs == 0 && svc.Spec.Type != corev1.ServiceTypeExternalName {
			sev := types.SeverityCritical
			if notReadyAddrs > 0 {
				sev = types.SeverityWarning
			}
			findings = append(findings, types.Finding{
				Category:    types.CategoryServices,
				Severity:    sev,
				Title:       "Service has no ready endpoints",
				Description: fmt.Sprintf("Service %s has 0 ready endpoints (%d not-ready)", ref, notReadyAddrs),
				Resource:    ref,
				Namespace:   svc.Namespace,
				Suggestion:  "Check pod readiness probes and label selectors match between Service and Pods. Run `kubectl get endpoints %s -n %s`",
			})
		}

		// Selector validation
		if len(svc.Spec.Selector) > 0 {
			findings = append(findings, c.checkSelectorMatch(ctx, svc)...)
		}

		// NodePort range check
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			for _, port := range svc.Spec.Ports {
				if port.NodePort < 30000 || port.NodePort > 32767 {
					findings = append(findings, types.Finding{
						Category:    types.CategoryServices,
						Severity:    types.SeverityWarning,
						Title:       "NodePort outside default range",
						Description: fmt.Sprintf("Service %s NodePort %d is outside 30000-32767", ref, port.NodePort),
						Resource:    ref,
						Namespace:   svc.Namespace,
						Suggestion:  "Verify the cluster allows custom NodePort ranges",
					})
				}
			}
		}

		// LoadBalancer pending
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			if len(svc.Status.LoadBalancer.Ingress) == 0 {
				pendingFor := time.Since(svc.CreationTimestamp.Time)
				if pendingFor > 3*time.Minute {
					findings = append(findings, types.Finding{
						Category:    types.CategoryServices,
						Severity:    types.SeverityWarning,
						Title:       "LoadBalancer service has no external IP",
						Description: fmt.Sprintf("Service %s has been pending an external IP for %.0f minutes", ref, pendingFor.Minutes()),
						Resource:    ref,
						Namespace:   svc.Namespace,
						Suggestion:  "Check cloud provider credentials, quota, and load balancer controller logs",
					})
				}
			}
		}
	}

	return findings
}

func (c *NetworkChecker) checkSelectorMatch(ctx context.Context, svc corev1.Service) []types.Finding {
	var findings []types.Finding

	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: svc.Spec.Selector})
	pods, err := c.client.CoreV1().Pods(svc.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil || len(pods.Items) == 0 {
		ref := fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
		findings = append(findings, types.Finding{
			Category:    types.CategoryLabels,
			Severity:    types.SeverityCritical,
			Title:       "Service selector matches no pods",
			Description: fmt.Sprintf("Service %s selector %v matches no pods in namespace %s", ref, svc.Spec.Selector, svc.Namespace),
			Resource:    ref,
			Namespace:   svc.Namespace,
			Suggestion:  "Verify pod labels match the service selector exactly. Labels are case-sensitive",
		})
	}

	return findings
}

// ── Kubernetes NetworkPolicies ───────────────────────────────────────────────

func (c *NetworkChecker) checkNetworkPolicies(ctx context.Context) []types.Finding {
	var findings []types.Finding

	policies, err := c.client.NetworkingV1().NetworkPolicies(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	// Check for deny-all policies that might isolate pods unexpectedly
	for _, pol := range policies.Items {
		ref := fmt.Sprintf("%s/%s", pol.Namespace, pol.Name)

		// Detect default-deny-all (empty podSelector + empty ingress/egress)
		if len(pol.Spec.PodSelector.MatchLabels) == 0 && len(pol.Spec.PodSelector.MatchExpressions) == 0 {
			if len(pol.Spec.Ingress) == 0 && containsPolicyType(pol.Spec.PolicyTypes, netv1.PolicyTypeIngress) {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNetwork,
					Severity:    types.SeverityInfo,
					Title:       "Default-deny ingress NetworkPolicy active",
					Description: fmt.Sprintf("NetworkPolicy %s denies all ingress to namespace %s", ref, pol.Namespace),
					Resource:    ref,
					Namespace:   pol.Namespace,
					Suggestion:  "Ensure all required ingress rules are explicitly allowed by other NetworkPolicies",
				})
			}
			if len(pol.Spec.Egress) == 0 && containsPolicyType(pol.Spec.PolicyTypes, netv1.PolicyTypeEgress) {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNetwork,
					Severity:    types.SeverityInfo,
					Title:       "Default-deny egress NetworkPolicy active",
					Description: fmt.Sprintf("NetworkPolicy %s denies all egress from namespace %s", ref, pol.Namespace),
					Resource:    ref,
					Namespace:   pol.Namespace,
					Suggestion:  "Ensure DNS (UDP 53), required services, and external endpoints are explicitly allowed",
				})
			}
		}

		// Warn if egress policy doesn't allow DNS
		if containsPolicyType(pol.Spec.PolicyTypes, netv1.PolicyTypeEgress) {
			hasDNS := false
			for _, egress := range pol.Spec.Egress {
				for _, port := range egress.Ports {
					if port.Port != nil && port.Port.IntValue() == 53 {
						hasDNS = true
					}
				}
			}
			if !hasDNS && len(pol.Spec.Egress) > 0 {
				findings = append(findings, types.Finding{
					Category:    types.CategoryNetwork,
					Severity:    types.SeverityWarning,
					Title:       "Egress NetworkPolicy may block DNS",
					Description: fmt.Sprintf("NetworkPolicy %s restricts egress but does not explicitly allow UDP/TCP port 53 (DNS)", ref),
					Resource:    ref,
					Namespace:   pol.Namespace,
					Suggestion:  "Add egress rule to allow UDP 53 to kube-dns/CoreDNS to prevent DNS resolution failures",
				})
			}
		}
	}

	// Check for pods with no NetworkPolicy (in namespaces that have at least one NP)
	if len(policies.Items) > 0 && c.namespace != "" {
		findings = append(findings, c.checkUnprotectedPods(ctx, policies.Items)...)
	}

	return findings
}

func (c *NetworkChecker) checkUnprotectedPods(ctx context.Context, policies []netv1.NetworkPolicy) []types.Finding {
	var findings []types.Finding

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	for _, pod := range pods.Items {
		covered := false
		for _, pol := range policies {
			if pol.Namespace != pod.Namespace {
				continue
			}
			if labelsMatch(pod.Labels, pol.Spec.PodSelector.MatchLabels) {
				covered = true
				break
			}
		}
		if !covered {
			findings = append(findings, types.Finding{
				Category:    types.CategoryNetwork,
				Severity:    types.SeverityInfo,
				Title:       "Pod not covered by any NetworkPolicy",
				Description: fmt.Sprintf("Pod %s/%s has no NetworkPolicy selecting it", pod.Namespace, pod.Name),
				Resource:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				Namespace:   pod.Namespace,
				Suggestion:  "Consider applying a NetworkPolicy for defence in depth",
			})
		}
	}

	return findings
}

// ── Calico Policies ──────────────────────────────────────────────────────────

// calicoPolicy is a minimal struct to parse Calico GlobalNetworkPolicy / NetworkPolicy CRDs
type calicoPolicy struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Selector string        `json:"selector"`
		Ingress  []interface{} `json:"ingress"`
		Egress   []interface{} `json:"egress"`
		Order    *float64      `json:"order"`
	} `json:"spec"`
}

var (
	calicoNetPolGVR = schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "networkpolicies",
	}
	calicoGlobalNetPolGVR = schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "globalnetworkpolicies",
	}
)

func (c *NetworkChecker) checkCalicoPolicies(ctx context.Context) []types.Finding {
	var findings []types.Finding

	// Use the dynamic client through REST discovery — Calico CRDs may or may not exist
	dc, err := c.dynamicList(ctx, calicoNetPolGVR, c.namespace)
	if err != nil {
		// Calico may not be installed — this is not an error
		findings = append(findings, types.Finding{
			Category:    types.CategoryCalico,
			Severity:    types.SeverityInfo,
			Title:       "Calico NetworkPolicy CRD not found",
			Description: "Calico NetworkPolicy CRDs are not installed on this cluster",
			Resource:    "cluster/calico",
			Suggestion:  "If Calico is the CNI, check that Calico CRDs are properly installed",
		})
		return findings
	}

	findings = append(findings, c.analyseCalicoList(dc, false)...)

	// Global policies (cluster-scoped)
	gdc, err := c.dynamicList(ctx, calicoGlobalNetPolGVR, "")
	if err == nil {
		findings = append(findings, c.analyseCalicoList(gdc, true)...)
	}

	return findings
}

func (c *NetworkChecker) analyseCalicoList(items []map[string]interface{}, global bool) []types.Finding {
	var findings []types.Finding
	kind := "NetworkPolicy"
	if global {
		kind = "GlobalNetworkPolicy"
	}

	for _, item := range items {
		raw, _ := json.Marshal(item)
		var pol calicoPolicy
		if err := json.Unmarshal(raw, &pol); err != nil {
			continue
		}

		ref := fmt.Sprintf("calico/%s/%s", kind, pol.Metadata.Name)

		// Warn about pass/deny-all Calico policies (empty selector)
		if pol.Spec.Selector == "" || pol.Spec.Selector == "all()" {
			if len(pol.Spec.Ingress) == 0 && len(pol.Spec.Egress) == 0 {
				findings = append(findings, types.Finding{
					Category:    types.CategoryCalico,
					Severity:    types.SeverityWarning,
					Title:       fmt.Sprintf("Calico %s deny-all policy", kind),
					Description: fmt.Sprintf("Calico policy %s selects all endpoints with no ingress/egress rules", ref),
					Resource:    ref,
					Namespace:   pol.Metadata.Namespace,
					Suggestion:  "Ensure this is intentional. This policy blocks all traffic matching the selector",
				})
			}
		}

		// Warn about missing order (undefined evaluation precedence)
		if pol.Spec.Order == nil && global {
			findings = append(findings, types.Finding{
				Category:    types.CategoryCalico,
				Severity:    types.SeverityInfo,
				Title:       "Calico GlobalNetworkPolicy has no order set",
				Description: fmt.Sprintf("Policy %s has no order — evaluation order is non-deterministic", ref),
				Resource:    ref,
				Suggestion:  "Set an explicit order value to control policy evaluation precedence",
			})
		}
	}

	return findings
}

func (c *NetworkChecker) dynamicList(ctx context.Context, gvr schema.GroupVersionResource, namespace string) ([]map[string]interface{}, error) {
	// Use RESTClient to dynamically list CRDs without importing dynamic client
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", gvr.Group, gvr.Version, namespace, gvr.Resource)
	} else {
		path = fmt.Sprintf("/apis/%s/%s/%s", gvr.Group, gvr.Version, gvr.Resource)
	}

	result := c.client.CoreV1().RESTClient().Get().AbsPath(path).Do(ctx)
	raw, err := result.Raw()
	if err != nil {
		return nil, err
	}

	var list struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

// ── DNS ──────────────────────────────────────────────────────────────────────

func (c *NetworkChecker) checkDNS(ctx context.Context) []types.Finding {
	var findings []types.Finding

	// Check CoreDNS / kube-dns is running in kube-system
	pods, err := c.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=kube-dns",
	})
	if err != nil {
		return findings
	}

	if len(pods.Items) == 0 {
		// Try CoreDNS label
		pods, err = c.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
			LabelSelector: "app=coredns",
		})
		if err != nil {
			return findings
		}
	}

	readyDNS := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			readyDNS++
		}
	}

	if len(pods.Items) == 0 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryDNS,
			Severity:    types.SeverityCritical,
			Title:       "No DNS pods found",
			Description: "Neither kube-dns nor CoreDNS pods were found in kube-system",
			Resource:    "kube-system/dns",
			Suggestion:  "DNS resolution in the cluster will fail. Deploy CoreDNS or kube-dns",
		})
	} else if readyDNS == 0 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryDNS,
			Severity:    types.SeverityCritical,
			Title:       "All DNS pods are not running",
			Description: fmt.Sprintf("%d DNS pods exist but none are in Running state", len(pods.Items)),
			Resource:    "kube-system/dns",
			Suggestion:  "Cluster DNS resolution is broken. Investigate CoreDNS pod logs immediately",
		})
	} else if readyDNS < len(pods.Items) {
		findings = append(findings, types.Finding{
			Category:    types.CategoryDNS,
			Severity:    types.SeverityWarning,
			Title:       "Some DNS pods are not running",
			Description: fmt.Sprintf("%d/%d DNS pods are running", readyDNS, len(pods.Items)),
			Resource:    "kube-system/dns",
			Suggestion:  "DNS capacity is reduced. Check failing pod logs",
		})
	}

	return findings
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func containsPolicyType(types []netv1.PolicyType, target netv1.PolicyType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

func labelsMatch(podLabels, selectorLabels map[string]string) bool {
	if len(selectorLabels) == 0 {
		return true // empty selector matches all
	}
	for k, v := range selectorLabels {
		if podLabels[k] != v {
			return false
		}
	}
	return true
}

// Unused but exported for completeness
var _ = strings.Contains
