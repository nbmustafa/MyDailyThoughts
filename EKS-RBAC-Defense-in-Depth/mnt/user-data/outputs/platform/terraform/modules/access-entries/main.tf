###############################################################################
# EKS Access Entries Module
#
# Replaces the deprecated aws-auth ConfigMap with the EKS Access Entries API.
# Authentication mode must be set to "API" on the cluster.
#
# Access Entries bind IAM ARNs → Kubernetes username/groups, then
# Kubernetes RBAC (ClusterRole/Role + Bindings) handles authorization.
###############################################################################

###############################################################################
# PLATFORM ADMIN — Full cluster write access
###############################################################################
resource "aws_eks_access_entry" "platform_admin" {
  cluster_name      = var.cluster_name
  principal_arn     = var.platform_admin_role_arn
  kubernetes_groups = ["platform-admins"]
  type              = "STANDARD"

  tags = merge(var.tags, {
    Role   = "platform-admin"
    Access = "cluster-write"
  })
}

###############################################################################
# PLATFORM READONLY — Full cluster read access
###############################################################################
resource "aws_eks_access_entry" "platform_readonly" {
  cluster_name      = var.cluster_name
  principal_arn     = var.platform_readonly_role_arn
  kubernetes_groups = ["platform-readers"]
  type              = "STANDARD"

  tags = merge(var.tags, {
    Role   = "platform-readonly"
    Access = "cluster-read"
  })
}

###############################################################################
# Node IAM Role Access Entry
# Required so worker nodes can register with the cluster.
# Type EC2_LINUX skips K8s user/group — nodes authenticate via their role.
###############################################################################
resource "aws_eks_access_entry" "node_group" {
  cluster_name  = var.cluster_name
  principal_arn = var.node_iam_role_arn
  type          = "EC2_LINUX"

  tags = var.tags
}

###############################################################################
# Karpenter Node IAM Role Access Entry
# Karpenter-launched nodes also need cluster access.
###############################################################################
resource "aws_eks_access_entry" "karpenter_node" {
  cluster_name  = var.cluster_name
  principal_arn = var.karpenter_node_role_arn
  type          = "EC2_LINUX"

  tags = var.tags
}

###############################################################################
# TENANT WRITE — One Access Entry per tenant
# Maps to K8s group "tenant-{name}-writers" → namespace-scoped Role via Helm
###############################################################################
resource "aws_eks_access_entry" "tenant_write" {
  for_each = var.tenants

  cluster_name = var.cluster_name
  principal_arn = each.value.write_role_arn

  # Group naming convention: tenant-{name}-writers
  # This group is bound to a namespace-scoped Role by the tenant-rbac Helm chart
  kubernetes_groups = ["tenant-${each.key}-writers"]
  type              = "STANDARD"

  tags = merge(var.tags, {
    Role   = "tenant-write"
    Tenant = each.key
    Access = "namespace-write"
  })
}

###############################################################################
# TENANT READONLY — One Access Entry per tenant
# Maps to K8s group "tenant-readers" → ClusterRole (read-only cluster-wide)
###############################################################################
resource "aws_eks_access_entry" "tenant_read" {
  for_each = var.tenants

  cluster_name      = var.cluster_name
  principal_arn     = each.value.read_role_arn
  kubernetes_groups = ["tenant-readers"]
  type              = "STANDARD"

  tags = merge(var.tags, {
    Role   = "tenant-readonly"
    Tenant = each.key
    Access = "cluster-read"
  })
}
