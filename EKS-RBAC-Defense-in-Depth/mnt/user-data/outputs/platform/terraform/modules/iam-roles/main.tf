###############################################################################
# IAM Roles Module — Platform + Tenant RBAC Roles
#
# Creates four categories of IAM roles:
#   1. platform-admin-role    → maps to K8s ClusterRole: platform-admin
#   2. platform-readonly-role → maps to K8s ClusterRole: platform-readonly
#   3. tenant-{name}-write    → maps to K8s Role (NS-scoped): tenant-writer
#   4. tenant-{name}-read     → maps to K8s ClusterRole: tenant-readonly
###############################################################################

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
}

###############################################################################
# PLATFORM ADMIN ROLE
###############################################################################
resource "aws_iam_role" "platform_admin" {
  name        = "eks-${var.cluster_name}-platform-admin"
  description = "Platform engineers — full write access to all cluster resources"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowPlatformEngineerAssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = var.platform_admin_principal_arns
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "sts:ExternalId" = var.external_id
          }
          Bool = {
            "aws:MultiFactorAuthPresent" = "true"
          }
        }
      }
    ]
  })

  max_session_duration = 3600

  tags = merge(var.tags, {
    Name    = "eks-${var.cluster_name}-platform-admin"
    Purpose = "eks-rbac"
    Role    = "platform-admin"
  })
}

resource "aws_iam_role_policy" "platform_admin_eks" {
  name = "eks-cluster-access"
  role = aws_iam_role.platform_admin.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DescribeCluster"
        Effect = "Allow"
        Action = ["eks:DescribeCluster", "eks:ListClusters"]
        Resource = "arn:aws:eks:${local.region}:${local.account_id}:cluster/${var.cluster_name}"
      }
    ]
  })
}

###############################################################################
# PLATFORM READONLY ROLE
###############################################################################
resource "aws_iam_role" "platform_readonly" {
  name        = "eks-${var.cluster_name}-platform-readonly"
  description = "Platform engineers — read-only access to all cluster resources"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowPlatformReadonlyAssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = var.platform_readonly_principal_arns
        }
        Action = "sts:AssumeRole"
        Condition = {
          Bool = {
            "aws:MultiFactorAuthPresent" = "true"
          }
        }
      }
    ]
  })

  max_session_duration = 3600

  tags = merge(var.tags, {
    Name    = "eks-${var.cluster_name}-platform-readonly"
    Purpose = "eks-rbac"
    Role    = "platform-readonly"
  })
}

resource "aws_iam_role_policy" "platform_readonly_eks" {
  name = "eks-cluster-access"
  role = aws_iam_role.platform_readonly.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DescribeCluster"
        Effect = "Allow"
        Action = ["eks:DescribeCluster", "eks:ListClusters"]
        Resource = "arn:aws:eks:${local.region}:${local.account_id}:cluster/${var.cluster_name}"
      }
    ]
  })
}

###############################################################################
# TENANT WRITE ROLES (one per tenant)
###############################################################################
resource "aws_iam_role" "tenant_write" {
  for_each = var.tenants

  name        = "eks-${var.cluster_name}-tenant-${each.key}-write"
  description = "Tenant ${each.key} — write access to own namespaces only"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowTenantWriteAssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = each.value.write_principal_arns
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:PrincipalTag/tenant" = each.key
          }
        }
      }
    ]
  })

  max_session_duration = 3600

  tags = merge(var.tags, {
    Name    = "eks-${var.cluster_name}-tenant-${each.key}-write"
    Purpose = "eks-rbac"
    Role    = "tenant-write"
    Tenant  = each.key
  })
}

resource "aws_iam_role_policy" "tenant_write_eks" {
  for_each = var.tenants

  name = "eks-cluster-access"
  role = aws_iam_role.tenant_write[each.key].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DescribeCluster"
        Effect = "Allow"
        Action = ["eks:DescribeCluster"]
        Resource = "arn:aws:eks:${local.region}:${local.account_id}:cluster/${var.cluster_name}"
      }
    ]
  })
}

###############################################################################
# TENANT READONLY ROLES (one per tenant, or a shared one)
###############################################################################
resource "aws_iam_role" "tenant_read" {
  for_each = var.tenants

  name        = "eks-${var.cluster_name}-tenant-${each.key}-read"
  description = "Tenant ${each.key} — read-only access cluster-wide"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowTenantReadAssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = each.value.read_principal_arns
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  max_session_duration = 3600

  tags = merge(var.tags, {
    Name    = "eks-${var.cluster_name}-tenant-${each.key}-read"
    Purpose = "eks-rbac"
    Role    = "tenant-read"
    Tenant  = each.key
  })
}

resource "aws_iam_role_policy" "tenant_read_eks" {
  for_each = var.tenants

  name = "eks-cluster-access"
  role = aws_iam_role.tenant_read[each.key].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DescribeCluster"
        Effect = "Allow"
        Action = ["eks:DescribeCluster"]
        Resource = "arn:aws:eks:${local.region}:${local.account_id}:cluster/${var.cluster_name}"
      }
    ]
  })
}
