Based on your description, it seems like the Terraform state after import isn't fully capturing the annotations from the live Kubernetes Ingress resource, even though your code matches exactly. This is a known issue that can occur due to how the Kubernetes provider handles metadata during import and plan operations. The behavior where changing one annotation causes all others to show as added indicates that the state file likely doesn't contain any annotations, so Terraform compares an empty state against your config and wants to add everything.

Here's a step-by-step solution to resolve this issue:

Step 1: Verify the Current State and Live Resource

First, confirm what's in your Terraform state and the live resource to understand the discrepancy.

1. Check the Terraform state after import:
   ```bash
   terraform state show kubernetes_ingress_v1.main
   ```
   Look for the annotations field in the output. If it's missing or incomplete, that's the problem.
2. Check the live resource using kubectl:
   ```bash
   kubectl get ingress my-app-ingress -n production -o json | jq .metadata.annotations
   ```
   Ensure the annotations match your Terraform code exactly.

Step 2: Update Your Terraform Configuration with a Lifecycle Block

To work around this, we'll use a lifecycle block to ignore changes to annotations temporarily. This will allow Terraform to adopt the existing annotations into its state without trying to manage them immediately. Add this to your ingress.tf:

```hcl
resource "kubernetes_ingress_v1" "main" {
  metadata {
    name      = "my-app-ingress"
    namespace = "production"
    annotations = {
      "kubernetes.io/ingress.class"                 = "nginx"
      "cert-manager.io/cluster-issuer"              = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/rewrite-target"  = "/"
      "nginx.ingress.kubernetes.io/ssl-redirect"    = "true"
      # Include all other annotations from your live resource here
    }
  }

  spec {
    # ... your existing spec ...
  }

  lifecycle {
    ignore_changes = [
      metadata.annotations, # Ignore all changes to annotations
    ]
  }
}
```

Step 3: Apply the Configuration to Sync the State

Now, apply this configuration to update the Terraform state without changing the live resource:

1. Run terraform apply:
   ```bash
   terraform apply
   ```
   This should show no changes to apply because the lifecycle block ignores annotations, and the resource already exists. However, during apply, Terraform will refresh its state from the live resource, which should now include the annotations.
2. Verify the state after apply:
   ```bash
   terraform state show kubernetes_ingress_v1.main
   ```
   You should now see all annotations in the state output.

Step 4: Remove the Lifecycle Block and Revert

Once the state is correctly populated, remove the lifecycle block to allow Terraform to manage annotations normally:

1. Remove the lifecycle block from your ingress.tf.
2. Run terraform plan again:
   ```bash
   terraform plan
   ```
   This should now show no changes, indicating that the state and configuration are in sync.

Step 5: Handle Ongoing Management

If you still see issues after this, consider the following:

· Provider Version: Ensure you're using the latest version of the Kubernetes provider. Update your providers.tf:
  ```hcl
  terraform {
    required_providers {
      kubernetes = {
        source  = "hashicorp/kubernetes"
        version = "~> 2.27.0" # Use the latest version
      }
    }
  }
  ```
  Then run terraform init -upgrade.
· Specific Annotations: If certain annotations are managed by external controllers (e.g., cloud-generated ones), you might need to ignore them permanently to avoid drift. For example:
  ```hcl
  lifecycle {
    ignore_changes = [
      metadata.annotations["kubectl.kubernetes.io/last-applied-configuration"],
      metadata.annotations["ingress.k8s.aws/stack"], # Example for AWS
      # Add other controller-specific annotations here
    ]
  }
  ```

Why This Works

The lifecycle block with ignore_changes prevents Terraform from detecting differences in annotations during apply. When you run apply, Terraform refreshes its state from the live resource, including all annotations. Removing the lifecycle block then allows Terraform to manage annotations, but since the state now matches the config, no changes are proposed.

This approach effectively forces a state refresh without altering the live resource, resolving the import issue. If problems persist, please provide the output of the state show and plan commands for further diagnosis.