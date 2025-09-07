Here's a step-by-step guide to deploy Argo CD with Terraform and Helm, and configure it to automatically sync with a dedicated Git repo and path for your app manifests.


---

✅ Objective:

Use Terraform to deploy Argo CD via Helm

Configure Argo CD to:

Watch a Git repository at a specific path

Automatically sync and deploy apps




---

🔧 Prerequisites:

A Kubernetes cluster (e.g., EKS, GKE, etc.)

kubectl, terraform, and helm installed

A Git repo with Kubernetes manifests (e.g., https://github.com/your-org/your-apps.git)



---

🗂 Directory Structure

terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── argocd-values.yaml       # Custom values for Helm chart
└── argocd-app.yaml          # App manifest to point Argo CD to the repo/path


---

1️⃣ Terraform Setup with Helm Provider

main.tf

provider "kubernetes" {
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}

resource "helm_release" "argocd" {
  name       = "argocd"
  namespace  = "argocd"
  create_namespace = true

  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "5.51.6"

  values = [
    file("${path.module}/argocd-values.yaml")
  ]
}


---

argocd-values.yaml

Customize Argo CD values. Here's a minimal example:

configs:
  cm:
    application.instanceLabelKey: argocd.argoproj.io/instance
    repositories: |
      - url: https://github.com/your-org/your-apps.git

server:
  service:
    type: LoadBalancer

> Replace your-org/your-apps.git with your Git repo URL.




---

2️⃣ Create Argo CD Application (Auto-managed via Terraform)

After Argo CD is deployed, we create an Application CR to point to your repo and path.

argocd-app.yaml

apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-apps
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/your-apps.git
    targetRevision: HEAD
    path: apps/production
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true


---

Add the argocd-app.yaml as a Terraform resource

In main.tf:

resource "kubernetes_manifest" "argocd_app" {
  manifest = yamldecode(file("${path.module}/argocd-app.yaml"))
  depends_on = [helm_release.argocd]
}

> Note: If your Terraform provider does not support kubernetes_manifest, you can apply it using a null_resource and kubectl.




---

✅ Optional: Auto-login to Argo CD CLI

After deployment, get Argo CD server address and admin password:

kubectl -n argocd get svc argocd-server -o wide
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

Then login:

argocd login <ARGOCD_SERVER> --username admin --password <PASSWORD>


---

🧪 Verify

1. After terraform apply, check:

kubectl get all -n argocd


2. Argo CD UI should be accessible via LoadBalancer IP


3. Your app in apps/production should be deployed automatically




---

🔁 Future Improvements

Use Helm chart for Argo CD Applications (argocd-apps chart)

Store Git repo credentials as secrets if private

Add monitoring & notifications



---

Would you like me to generate a ready-to-run GitHub repo structure with these files?

