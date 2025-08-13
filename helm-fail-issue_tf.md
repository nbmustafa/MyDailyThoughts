Helm_release exists after a failed install

When a Helm install fails under Terraform, Helm still records a release with that name in the cluster (status: failed). On the next terraform apply, the provider tries a fresh install and Helm rejects it because a release with that name already exists. The fix is twofold: clean up once, then configure the resource to auto-rollback on future failures.

---

One‑time recovery

Use one of these paths to get back to green.

- Uninstall the failed release:
  - helm uninstall <release> -n <namespace>
  - Verify it’s gone: helm ls -n <namespace>
- Or import it into state (if you want Terraform to “own” it now):
  - terraform import helmrelease.<resourcename> <namespace>/<release>
  - Follow with terraform apply to upgrade/fix it.
- If uninstall hangs due to finalizers:
  - kubectl get ns/<namespace> or problematic resources, remove stuck finalizers, then retry uninstall.

> Tip: If Terraform has a half-created state for the release and you want to start clean, terraform state rm helmrelease.<resourcename> and then uninstall the release, or import it and apply.

---

Make future applies resilient

Configure helm_release so failures roll back and don’t leave orphaned releases.

`hcl
resource "helmrelease" "myapp" {
  name       = "my-app"
  namespace  = "my-namespace"
  repository = "https://charts.example.com"
  chart      = "my-chart"
  version    = "1.2.3"

  create_namespace   = true
  wait               = true          # wait for resources to become ready
  waitforjobs      = true          # include Helm hook jobs in 'wait'
  timeout            = 600           # tune to your environment
  atomic             = true          # uninstall on failure (install) / rollback (upgrade)
  cleanuponfail    = true          # clean partial resources on upgrade failure
  force_update       = true          # replace immutable fields on upgrade
  dependency_update  = true          # refresh chart deps before install/upgrade
  max_history        = 5             # keep Helm history short to avoid clutter

values = [file("values.yaml")]

set { name = "image.tag"; value = "..." }
}
`

- atomic: Ensures failed installs are automatically uninstalled, so the name isn’t left “taken.”
- wait + waitforjobs + timeout: Reduce timing flakiness that causes false failures.
- cleanuponfail: On upgrade failures, cleans up new resources created during the attempt.
- force_update: Useful when charts change immutable fields (e.g., Service type/selector, PVC templates).
- create_namespace: Avoids races on first install.

---

Common root causes of failed installs (and quick checks)

- CRDs not applied or out of order:
  - Ensure CRDs are installed first. If the chart bundles CRDs, consider a separate helmrelease for CRDs and set the main release with skipcrds = true.
- Validating webhooks blocking new resources:
  - During upgrades, webhooks can reject transient states. atomic + wait helps; otherwise, stage changes to satisfy validators.
- RBAC/service account issues:
  - Verify the chart’s serviceAccount and Roles match your cluster policies.
- Resource quotas or PDBs preventing rollout:
  - Check namespace quotas and PodDisruptionBudgets.
- Init/Hook jobs failing:
  - Use waitforjobs = true; inspect hook job logs on failure.

---

Operational guardrails

- Pre-flight check: helm lint and dry-run in CI before terraform apply.
- Serializing dependencies: Use dependson between helmrelease resources (e.g., CRDs -> operator -> workloads).
- Consistent naming: Keep release name stable; prefer values changes over renames to avoid churn.
- Observability: On failure, immediately check helm status <name> -n <ns> and kubectl describe for the failing pods/jobs.

---

If it happens again

- Fast clean: helm uninstall <name> -n <ns>, then terraform apply.
- Take ownership instead: terraform import helm_release.<res> <ns>/<name>, then apply.

If you share the specific chart, Helm/Terraform provider versions, and the failure message, I can tailor the exact flags and a minimal, reproducible config for your setup.