###############################################################################
# Production Environment — EKS Platform
# Orchestrates: EKS cluster, IAM roles, Access Entries, Helm releases
###############################################################################

terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.13"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.30"
    }
  }

  backend "s3" {
    bucket         = "your-tfstate-bucket"
    key            = "eks/platform/production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "eks-platform"
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "platform-team"
    }
  }
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_ca_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_ca_data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

###############################################################################
# EKS Cluster
###############################################################################
module "eks" {
  source = "../../modules/eks"

  cluster_name              = var.cluster_name
  environment               = var.environment
  kubernetes_version        = "1.32"
  vpc_id                    = var.vpc_id
  private_subnet_ids        = var.private_subnet_ids
  endpoint_public_access    = false

  platform_node_instance_types = ["m6i.2xlarge", "m6a.2xlarge"]
  platform_node_desired        = 3
  platform_node_min            = 2
  platform_node_max            = 6

  log_retention_days = 90

  tags = local.common_tags
}

###############################################################################
# IAM Roles (platform + tenant)
###############################################################################
module "iam_roles" {
  source = "../../modules/iam-roles"

  cluster_name  = module.eks.cluster_name
  external_id   = var.platform_admin_external_id

  platform_admin_principal_arns = [
    "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/platform-engineers-sso",
    "arn:aws:iam::${data.aws_caller_identity.current.account_id}:group/platform-engineers",
  ]

  platform_readonly_principal_arns = [
    "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/platform-oncall-sso",
    "arn:aws:iam::${data.aws_caller_identity.current.account_id}:group/platform-oncall",
  ]

  tenants = var.tenants

  tags = local.common_tags
}

###############################################################################
# Karpenter IRSA
###############################################################################
module "karpenter_irsa" {
  source = "../../modules/eks"

  cluster_name          = module.eks.cluster_name
  oidc_provider_arn     = module.eks.oidc_provider_arn
  oidc_issuer_url       = module.eks.cluster_oidc_issuer_url
  karpenter_sqs_arn     = module.eks.karpenter_sqs_arn
  tags                  = local.common_tags
}

###############################################################################
# EKS Access Entries
###############################################################################
module "access_entries" {
  source = "../../modules/access-entries"

  cluster_name               = module.eks.cluster_name
  platform_admin_role_arn    = module.iam_roles.platform_admin_role_arn
  platform_readonly_role_arn = module.iam_roles.platform_readonly_role_arn
  node_iam_role_arn          = module.eks.node_iam_role_arn
  karpenter_node_role_arn    = module.karpenter_irsa.karpenter_node_role_arn

  tenants = {
    for name, config in var.tenants : name => {
      write_role_arn = module.iam_roles.tenant_write_role_arns[name]
      read_role_arn  = module.iam_roles.tenant_read_role_arns[name]
      namespaces     = config.namespaces
    }
  }

  tags = local.common_tags

  depends_on = [module.eks, module.iam_roles]
}

###############################################################################
# Helm — Platform RBAC (ClusterRoles, bindings, tenant namespace RBAC)
###############################################################################
resource "helm_release" "platform_rbac" {
  name             = "platform-rbac"
  chart            = "${path.module}/../../../../helm/platform-rbac"
  namespace        = "kube-system"
  create_namespace = false
  atomic           = true
  cleanup_on_fail  = true

  values = [
    templatefile("${path.module}/helm-values/platform-rbac.yaml", {
      tenants = var.tenants
    })
  ]

  depends_on = [module.access_entries]
}

###############################################################################
# Helm — Karpenter
###############################################################################
resource "helm_release" "karpenter" {
  name             = "karpenter"
  repository       = "oci://public.ecr.aws/karpenter"
  chart            = "karpenter"
  version          = "1.2.1"
  namespace        = "karpenter"
  create_namespace = true
  atomic           = true
  cleanup_on_fail  = true

  set {
    name  = "settings.clusterName"
    value = module.eks.cluster_name
  }
  set {
    name  = "settings.interruptionQueue"
    value = module.eks.karpenter_sqs_url
  }
  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.karpenter_irsa.karpenter_controller_role_arn
  }
  set {
    name  = "controller.resources.requests.cpu"
    value = "250m"
  }
  set {
    name  = "controller.resources.requests.memory"
    value = "512Mi"
  }

  depends_on = [module.eks]
}

###############################################################################
# Helm — ArgoCD
###############################################################################
resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = "7.7.3"
  namespace        = "argocd"
  create_namespace = true
  atomic           = true
  cleanup_on_fail  = true

  values = [file("${path.module}/helm-values/argocd.yaml")]

  depends_on = [module.eks]
}

###############################################################################
# Locals + data
###############################################################################
data "aws_caller_identity" "current" {}

locals {
  common_tags = {
    Project     = "eks-platform"
    Environment = var.environment
    ClusterName = var.cluster_name
  }
}
