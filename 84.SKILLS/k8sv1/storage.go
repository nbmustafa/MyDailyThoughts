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

// StorageChecker checks PV, PVC, StorageClass and volume mount issues
type StorageChecker struct {
	client    kubernetes.Interface
	namespace string
}

func NewStorageChecker(client kubernetes.Interface, namespace string) *StorageChecker {
	return &StorageChecker{client: client, namespace: namespace}
}

func (c *StorageChecker) Category() types.Category { return types.CategoryStorage }

func (c *StorageChecker) Run(ctx context.Context) types.CheckResult {
	start := time.Now()
	result := types.CheckResult{Category: c.Category()}

	result.Findings = append(result.Findings, c.checkPVCs(ctx)...)
	result.Findings = append(result.Findings, c.checkPVs(ctx)...)
	result.Findings = append(result.Findings, c.checkStorageClasses(ctx)...)
	result.Findings = append(result.Findings, c.checkVolumeMounts(ctx)...)

	result.Duration = time.Since(start)
	return result
}

func (c *StorageChecker) checkPVCs(ctx context.Context) []types.Finding {
	var findings []types.Finding

	pvcs, err := c.client.CoreV1().PersistentVolumeClaims(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []types.Finding{{
			Category:    types.CategoryStorage,
			Severity:    types.SeverityWarning,
			Title:       "Failed to list PVCs",
			Description: err.Error(),
		}}
	}

	for _, pvc := range pvcs.Items {
		ref := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)

		switch pvc.Status.Phase {
		case corev1.ClaimPending:
			pendingFor := time.Since(pvc.CreationTimestamp.Time)
			sev := types.SeverityWarning
			if pendingFor > 10*time.Minute {
				sev = types.SeverityCritical
			}
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    sev,
				Title:       "PVC stuck in Pending",
				Description: fmt.Sprintf("PVC %s has been Pending for %.0f minutes", ref, pendingFor.Minutes()),
				Resource:    ref,
				Namespace:   pvc.Namespace,
				Suggestion:  "Check StorageClass exists, provisioner is running, and node/zone topology. Run `kubectl describe pvc` for events",
			})

		case corev1.ClaimLost:
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityCritical,
				Title:       "PVC in Lost state",
				Description: fmt.Sprintf("PVC %s is in Lost state — the underlying PV may have been deleted", ref),
				Resource:    ref,
				Namespace:   pvc.Namespace,
				Suggestion:  "The PV binding is gone. Check if PV still exists and reclaim policy. Data may be at risk",
			})
		}

		// Check for missing StorageClass
		if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName == "" {
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityInfo,
				Title:       "PVC using default StorageClass",
				Description: fmt.Sprintf("PVC %s does not specify a StorageClass — using cluster default", ref),
				Resource:    ref,
				Namespace:   pvc.Namespace,
				Suggestion:  "Explicitly set storageClassName to avoid unexpected behaviour if default changes",
			})
		}

		// Warn on PVC access modes mismatch for ReadWriteOnce with multiple pods
		if containsAccessMode(pvc.Spec.AccessModes, corev1.ReadWriteOnce) {
			findings = append(findings, c.checkRWOConflict(ctx, pvc)...)
		}
	}

	return findings
}

func (c *StorageChecker) checkRWOConflict(ctx context.Context, pvc corev1.PersistentVolumeClaim) []types.Finding {
	var findings []types.Finding
	if pvc.Status.Phase != corev1.ClaimBound {
		return findings
	}

	// Find pods using this PVC
	pods, err := c.client.CoreV1().Pods(pvc.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	var consumers []string
	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvc.Name {
				consumers = append(consumers, pod.Name)
			}
		}
	}

	if len(consumers) > 1 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryStorage,
			Severity:    types.SeverityWarning,
			Title:       "ReadWriteOnce PVC mounted by multiple pods",
			Description: fmt.Sprintf("PVC %s/%s (RWO) is referenced by pods: %v", pvc.Namespace, pvc.Name, consumers),
			Resource:    fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name),
			Namespace:   pvc.Namespace,
			Suggestion:  "RWO volumes can only be mounted by pods on the same node. Use ReadWriteMany if multi-node access is needed",
		})
	}

	return findings
}

func (c *StorageChecker) checkPVs(ctx context.Context) []types.Finding {
	var findings []types.Finding

	pvs, err := c.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []types.Finding{{
			Category:    types.CategoryStorage,
			Severity:    types.SeverityWarning,
			Title:       "Failed to list PVs",
			Description: err.Error(),
		}}
	}

	for _, pv := range pvs.Items {
		switch pv.Status.Phase {
		case corev1.VolumeFailed:
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityCritical,
				Title:       "PV in Failed state",
				Description: fmt.Sprintf("PersistentVolume %s is in Failed state: %s", pv.Name, pv.Status.Message),
				Resource:    pv.Name,
				Suggestion:  "Check the volume backend (EBS/NFS/Ceph etc.) and reclaim policy. Manual intervention may be required",
			})
		case corev1.VolumeReleased:
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityInfo,
				Title:       "PV in Released state",
				Description: fmt.Sprintf("PersistentVolume %s has been released but not reclaimed", pv.Name),
				Resource:    pv.Name,
				Suggestion:  "If reclaim policy is Retain, manually reclaim or delete the PV. Data on the volume is still intact",
			})
		case corev1.VolumeAvailable:
			// Unbound PVs sitting available — possibly orphaned
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityInfo,
				Title:       "PV available but unbound",
				Description: fmt.Sprintf("PersistentVolume %s is Available (unbound) — may be unused", pv.Name),
				Resource:    pv.Name,
				Suggestion:  "Verify this PV is intended to remain unbound, or clean it up to free storage quota",
			})
		}

		// Check reclaim policy — Delete on cloud volumes can cause data loss
		if pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			findings = append(findings, types.Finding{
				Category:    types.CategoryStorage,
				Severity:    types.SeverityInfo,
				Title:       "PV reclaim policy is Delete",
				Description: fmt.Sprintf("PV %s has reclaimPolicy=Delete — volume will be destroyed when PVC is deleted", pv.Name),
				Resource:    pv.Name,
				Suggestion:  "Ensure this is intentional. Consider Retain for critical data volumes",
			})
		}
	}

	return findings
}

func (c *StorageChecker) checkStorageClasses(ctx context.Context) []types.Finding {
	var findings []types.Finding

	scs, err := c.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	defaultCount := 0
	for _, sc := range scs.Items {
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultCount++
		}
	}

	if defaultCount == 0 && len(scs.Items) > 0 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryStorage,
			Severity:    types.SeverityWarning,
			Title:       "No default StorageClass configured",
			Description: "No StorageClass is marked as default — PVCs without explicit storageClassName will fail",
			Resource:    "cluster/storageclass",
			Suggestion:  "Set a default StorageClass: `kubectl patch storageclass <name> -p '{\"metadata\":{\"annotations\":{\"storageclass.kubernetes.io/is-default-class\":\"true\"}}}'`",
		})
	}

	if defaultCount > 1 {
		findings = append(findings, types.Finding{
			Category:    types.CategoryStorage,
			Severity:    types.SeverityWarning,
			Title:       "Multiple default StorageClasses",
			Description: fmt.Sprintf("%d StorageClasses are marked as default — PVC behaviour may be unpredictable", defaultCount),
			Resource:    "cluster/storageclass",
			Suggestion:  "Only one StorageClass should be set as default",
		})
	}

	return findings
}

func (c *StorageChecker) checkVolumeMounts(ctx context.Context) []types.Finding {
	var findings []types.Finding

	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return findings
	}

	for _, pod := range pods.Items {
		ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		// Build set of volumes defined in spec
		definedVols := map[string]bool{}
		for _, vol := range pod.Spec.Volumes {
			definedVols[vol.Name] = true
		}

		// Check all containers reference existing volumes
		for _, container := range pod.Spec.Containers {
			for _, mount := range container.VolumeMounts {
				if !definedVols[mount.Name] {
					findings = append(findings, types.Finding{
						Category:    types.CategoryStorage,
						Severity:    types.SeverityCritical,
						Title:       "VolumeMount references undefined volume",
						Description: fmt.Sprintf("Container %q in pod %s mounts volume %q which is not defined in pod spec", container.Name, ref, mount.Name),
						Resource:    ref,
						Namespace:   pod.Namespace,
						Suggestion:  "Add the missing volume definition to pod spec volumes",
					})
				}
			}
		}
	}

	return findings
}

func containsAccessMode(modes []corev1.PersistentVolumeAccessMode, target corev1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == target {
			return true
		}
	}
	return false
}
