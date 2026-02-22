# EKS Managed Platform — RBAC Architecture

> **Principal Security Architect + Principal Platform Engineer Design**  
> AWS EKS · Access Entries (no aws-auth) · ArgoCD · Karpenter · Terraform · Helm

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  AWS IAM                                                    │
│  ┌──────────────────────┐  ┌──────────────────────────────┐ │
│  │ platform-admin-role  │  │ tenant-{name}-write-role     │ │
│  │ platform-read-role   │  │ tenant-{name}-read-role      │ │
│  └──────────┬───────────┘  └──────────────┬───────────────┘ │
└─────────────┼────────────────────────────┼─────────────────┘
              │ EKS Access Entries (API mode — no aws-auth)
              ▼
┌─────────────────────────────────────────────────────────────┐
│  EKS Cluster                                                │
│  ┌──────────────────┐  ┌──────────────────────────────────┐ │
│  │ Platform NS      │  │ Tenant Namespaces                │ │
│  │ - argocd         │  │ - tenant-alpha-{dev,stg,prod}    │ │
│  │ - karpenter      │  │ - tenant-beta-{dev,prod}         │ │
│  │ - monitoring     │  │                                  │ │
│  │ - cert-manager   │  │ [Write: tenant-writer Role]      │ │
│  │ - etc.           │  │ [Read:  tenant-readonly CR]      │ │
│  └──────────────────┘  └──────────────────────────────────┘ │
│                                                             │
│  K8s RBAC (Helm-managed)                                    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ platform-admin ClusterRole     → platform-admins grp  │  │
│  │ platform-readonly ClusterRole  → platform-readers grp │  │
│  │ tenant-readonly ClusterRole    → tenant-readers grp   │  │
│  │ tenant-writer Role (per NS)    → tenant-{x}-writers   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## RBAC Matrix

| Role | Scope | Secrets | Exec | Platform NS | Tenant NS |
|------|-------|---------|------|-------------|-----------|
| Platform Admin | Cluster | ✅ Full | ✅ | ✅ Write | ✅ Write |
| Platform ReadOnly | Cluster | 📋 List only | ❌ | ✅ Read | ✅ Read |
| Tenant Write | Namespace | ✅ Own NS only | ✅ Own NS | ❌ None | ✅ Write own |
| Tenant ReadOnly | Cluster | ❌ None | ❌ | 🔍 Minimal | ✅ Read |

## File Structure

```
platform/
├── terraform/
│   ├── modules/
│   │   ├── eks/                      # EKS cluster, node groups, Karpenter SQS
│   │   │   ├── main.tf
│   │   │   ├── karpenter_irsa.tf     # Karpenter controller + node IAM
│   │   │   ├── variables.tf
│   │   │   └── outputs.tf
│   │   ├── iam-roles/                # Platform + tenant IAM roles
│   │   │   ├── main.tf
│   │   │   └── variables_outputs.tf
│   │   └── access-entries/           # EKS Access Entries (replaces aws-auth)
│   │       ├── main.tf
│   │       └── variables_outputs.tf
│   └── environments/
│       └── production/
│           ├── main.tf               # Wires all modules + Helm releases
│           ├── variables.tf
│           └── terraform.tfvars
├── helm/
│   └── platform-rbac/
│       ├── Chart.yaml
│       ├── values.yaml               # Tenants config lives here
│       └── templates/
│           ├── cluster-roles.yaml              # 3 ClusterRoles
│           ├── cluster-role-bindings.yaml      # 3 ClusterRoleBindings
│           └── tenant-namespaces-and-roles.yaml # Per NS: NS + Role + RoleBinding + Quota + NP
└── argocd/
    ├── platform/apps/
    │   ├── app-of-apps.yaml          # Platform App-of-Apps + AppProjects
    │   ├── platform-rbac-app.yaml    # Deploys the Helm chart above
    │   └── tenant-team-alpha-app.yaml # Points at tenant gitops repo
    └── tenant-template/
        └── karpenter/
            └── nodepool.yaml         # Tenant NodePool + EC2NodeClass template
```

## Deployment Order

```bash
# 1. Apply Terraform (EKS + IAM + Access Entries)
cd terraform/environments/production
terraform init
terraform plan -var="platform_admin_external_id=$EXTERNAL_ID"
terraform apply

# 2. Bootstrap ArgoCD (one-time)
helm upgrade --install argocd argo/argo-cd \
  --namespace argocd --create-namespace \
  -f helm-values/argocd.yaml

# 3. Apply the App-of-Apps — ArgoCD takes over from here
kubectl apply -f argocd/platform/apps/app-of-apps.yaml

# 4. Onboard a new tenant (add to values.yaml + tfvars, then PR)
# → platform-rbac Helm auto-syncs via ArgoCD
# → Terraform access-entries auto-provisions via CI/CD
```

## Adding a New Tenant

1. Add to `terraform.tfvars`:
```hcl
tenants = {
  "team-gamma" = {
    write_principal_arns = ["arn:aws:iam::ACCOUNT:role/team-gamma-sso"]
    read_principal_arns  = ["arn:aws:iam::ACCOUNT:role/team-gamma-read-sso"]
    namespaces           = ["tenant-gamma-dev", "tenant-gamma-prod"]
  }
}
```

2. Add to `helm/platform-rbac/values.yaml`:
```yaml
tenants:
  - name: team-gamma
    k8sGroup: tenant-team-gamma-writers
    namespaces:
      - tenant-gamma-dev
      - tenant-gamma-prod
```

3. Add ArgoCD Application pointing at their gitops repo.

4. Merge PR → CI runs `terraform apply` → ArgoCD syncs Helm chart → done.

## Security Decisions

- **No aws-auth ConfigMap**: Fully replaced by EKS Access Entries API (`authenticationMode: API`)
- **IMDSv2 enforced**: `http_tokens: required` on all launch templates
- **Secrets encryption**: KMS with key rotation enabled
- **Node isolation**: Platform nodes have `NoSchedule` taint — tenant workloads cannot run there
- **NetworkPolicy**: Default-deny ingress on all tenant namespaces
- **Tenant namespace scope**: Tenant write Role is namespace-scoped, never ClusterRole
- **Secrets access**: Tenant readers cannot `get`/`watch` secrets — list only for platform readers
- **MFA required**: Platform admin role requires `aws:MultiFactorAuthPresent: true`
- **ArgoCD AppProject**: Tenants cannot deploy to platform namespaces via ArgoCD
- **Karpenter limits**: Hard CPU/memory limits on each tenant NodePool
