# Kubernetes Platform Engineer: Scalable EKS Provisioning with Terraform

## Solution Overview

As a Kubernetes Platform Engineer, I'll design a scalable Terraform solution to provision EKS clusters based on tenant profiles stored as YAML files in separate Git repositories. This approach enables self-service for tenants while maintaining centralized control and standardization.

## Architecture Components

1. **Tenant Profile Repositories**: Each tenant maintains their cluster configurations in their own Git repo
2. **Configuration Management**: Centralized component to process tenant YAML profiles
3. **Terraform Execution**: Dynamic EKS cluster provisioning based on processed configurations
4. **State Management**: Secure storage for Terraform state files per tenant/cluster

## Implementation Approach

### 1. Tenant Profile Structure (YAML)

Each tenant repository would contain files like:
```yaml
# tenant-a/production-cluster.yaml
cluster:
  name: tenant-a-prod
  region: us-west-2
  kubernetes_version: "1.27"
  node_groups:
    - name: general-purpose
      instance_type: m5.large
      min_size: 3
      max_size: 10
      desired_size: 3
    - name: memory-optimized
      instance_type: r5.xlarge
      min_size: 1
      max_size: 5
  network:
    vpc_cidr: "10.10.0.0/16"
    private_subnets: ["10.10.1.0/24", "10.10.2.0/24", "10.10.3.0/24"]
    public_subnets: ["10.10.101.0/24", "10.10.102.0/24", "10.10.103.0/24"]
  addons:
    - name: aws-ebs-csi-driver
      version: "v1.20.0"
    - name: aws-load-balancer-controller
      version: "v2.6.0"
```

### 2. Terraform Module Structure

```
modules/
├── eks-cluster/
│   ├── main.tf          # EKS cluster resources
│   ├── variables.tf     # Input variables
│   └── outputs.tf      # Output values
└── vpc/
    ├── main.tf          # VPC resources
    ├── variables.tf
    └── outputs.tf
```

### 3. Root Module for Dynamic Provisioning

```hcl
# main.tf
terraform {
  required_version = ">= 1.3.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

locals {
  tenant_configs = {
    for file in fileset("${path.module}/tenants/", "**/*.yaml") :
    trimsuffix(file, ".yaml") => yamldecode(file("${path.module}/tenants/${file}"))
  }
}

module "tenant_clusters" {
  for_each = local.tenant_configs

  source = "./modules/eks-cluster"

  tenant_name         = each.key
  cluster_config      = each.value.cluster
  vpc_config         = each.value.network
  addons_config      = each.value.addons
  node_groups_config = each.value.node_groups
}
```

### 4. EKS Cluster Module (simplified)

```hcl
# modules/eks-cluster/main.tf
module "vpc" {
  source = "../vpc"

  name                 = "${var.tenant_name}-vpc"
  cidr                 = var.vpc_config.vpc_cidr
  private_subnets      = var.vpc_config.private_subnets
  public_subnets       = var.vpc_config.public_subnets
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true
}

resource "aws_eks_cluster" "this" {
  name     = var.cluster_config.name
  role_arn = aws_iam_role.cluster.arn
  version  = var.cluster_config.kubernetes_version

  vpc_config {
    subnet_ids = concat(
      module.vpc.private_subnets,
      module.vpc.public_subnets
    )
  }
}

resource "aws_eks_node_group" "this" {
  for_each = { for ng in var.node_groups_config : ng.name => ng }

  cluster_name    = aws_eks_cluster.this.name
  node_group_name = each.key
  node_role_arn   = aws_iam_role.nodes.arn
  subnet_ids      = module.vpc.private_subnets

  scaling_config {
    desired_size = each.value.desired_size
    max_size     = each.value.max_size
    min_size     = each.value.min_size
  }

  instance_types = [each.value.instance_type]
}
```

### 5. GitOps Integration

To automatically trigger provisioning when tenants update their YAML files:

1. **Option 1**: Webhook-based approach
   - Set up a webhook in each tenant repo to trigger your Terraform pipeline
   - Validate changes before applying

2. **Option 2**: Periodic sync
   - Use a CI/CD pipeline that periodically checks for changes in tenant repos
   - Process only changed configurations

### 6. State Management

```hcl
# backend.tf (per tenant or per cluster)
terraform {
  backend "s3" {
    bucket         = "eks-tfstate-${var.tenant_name}"
    key            = "${var.cluster_config.name}/terraform.tfstate"
    region         = var.cluster_config.region
    dynamodb_table = "eks-tfstate-lock"
    encrypt        = true
  }
}
```

## Deployment Workflow

1. Tenant creates/updates YAML file in their repo
2. Git webhook triggers pipeline or pipeline detects change
3. System validates YAML against schema
4. Terraform plan/apply executes for affected clusters
5. Status reported back to tenant (via PR comments, Slack, etc.)

## Scaling Considerations

1. **Isolation**: Each tenant's infrastructure is managed separately
2. **Modularity**: Reusable Terraform modules for consistent deployments
3. **Configuration Drift Prevention**: State files are strictly managed
4. **Automation**: Full CI/CD pipeline reduces manual intervention

## Security Best Practices

1. Implement OPA/Gatekeeper policies for tenant clusters
2. Use IAM roles for service accounts (IRSA)
3. Enforce network policies
4. Implement centralized logging and monitoring

Would you like me to elaborate on any specific part of this solution?