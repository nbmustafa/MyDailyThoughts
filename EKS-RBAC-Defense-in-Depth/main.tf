###############################################################################
# EKS Cluster Module
# - Creates EKS cluster with API auth mode (Access Entries, no aws-auth)
# - Managed node group for platform system workloads
# - Karpenter IRSA + SQS for interruption handling
###############################################################################

locals {
  cluster_name = "${var.cluster_name}-${var.environment}"
  common_tags = merge(var.tags, {
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
    ManagedBy                                      = "terraform"
    Environment                                    = var.environment
  })
}

###############################################################################
# EKS Cluster
###############################################################################
resource "aws_eks_cluster" "this" {
  name     = local.cluster_name
  version  = var.kubernetes_version
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = var.endpoint_public_access
    public_access_cidrs     = var.public_access_cidrs
    security_group_ids      = [aws_security_group.cluster.id]
  }

  # ── CRITICAL: Use API mode — eliminates aws-auth ConfigMap entirely ──
  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = false # we manage this explicitly
  }

  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }

  enabled_cluster_log_types = [
    "api", "audit", "authenticator", "controllerManager", "scheduler"
  ]

  kubernetes_network_config {
    service_ipv4_cidr = var.service_cidr
    ip_family         = "ipv4"
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster_policy,
    aws_iam_role_policy_attachment.cluster_vpc_policy,
    aws_cloudwatch_log_group.eks,
  ]

  tags = local.common_tags
}

###############################################################################
# OIDC Provider (required for IRSA)
###############################################################################
data "tls_certificate" "cluster" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "cluster" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cluster.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer

  tags = local.common_tags
}

###############################################################################
# Platform System Managed Node Group
###############################################################################
resource "aws_eks_node_group" "platform_system" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "platform-system"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.private_subnet_ids
  instance_types  = var.platform_node_instance_types
  capacity_type   = "ON_DEMAND"
  ami_type        = "AL2_x86_64"

  scaling_config {
    desired_size = var.platform_node_desired
    min_size     = var.platform_node_min
    max_size     = var.platform_node_max
  }

  update_config {
    max_unavailable_percentage = 25
  }

  labels = {
    role                     = "platform-system"
    "platform.io/node-group" = "platform-system"
  }

  taint {
    key    = "platform.io/platform-only"
    value  = "true"
    effect = "NO_SCHEDULE"
  }

  launch_template {
    id      = aws_launch_template.platform_system.id
    version = aws_launch_template.platform_system.latest_version
  }

  depends_on = [
    aws_iam_role_policy_attachment.node_worker_policy,
    aws_iam_role_policy_attachment.node_cni_policy,
    aws_iam_role_policy_attachment.node_ecr_policy,
  ]

  tags = local.common_tags
}

resource "aws_launch_template" "platform_system" {
  name_prefix   = "${local.cluster_name}-platform-system-"
  instance_type = var.platform_node_instance_types[0]

  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size           = 50
      volume_type           = "gp3"
      encrypted             = true
      kms_key_id            = aws_kms_key.eks.arn
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required" # IMDSv2 enforced
    http_put_response_hop_limit = 1
  }

  monitoring { enabled = true }

  tag_specifications {
    resource_type = "instance"
    tags          = merge(local.common_tags, { Name = "${local.cluster_name}-platform-system" })
  }

  tags = local.common_tags
}

###############################################################################
# KMS Key for EKS secrets encryption
###############################################################################
resource "aws_kms_key" "eks" {
  description             = "EKS cluster ${local.cluster_name} secrets encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = local.common_tags
}

resource "aws_kms_alias" "eks" {
  name          = "alias/eks-${local.cluster_name}"
  target_key_id = aws_kms_key.eks.key_id
}

###############################################################################
# CloudWatch Log Group
###############################################################################
resource "aws_cloudwatch_log_group" "eks" {
  name              = "/aws/eks/${local.cluster_name}/cluster"
  retention_in_days = var.log_retention_days
  kms_key_id        = aws_kms_key.eks.arn
  tags              = local.common_tags
}

###############################################################################
# Security Groups
###############################################################################
resource "aws_security_group" "cluster" {
  name_prefix = "${local.cluster_name}-cluster-"
  description = "EKS cluster control plane security group"
  vpc_id      = var.vpc_id

  tags = merge(local.common_tags, { Name = "${local.cluster_name}-cluster-sg" })

  lifecycle { create_before_destroy = true }
}

resource "aws_security_group_rule" "cluster_ingress_nodes" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.cluster.id
  source_security_group_id = aws_security_group.nodes.id
  description              = "Node groups to cluster API"
}

resource "aws_security_group_rule" "cluster_egress_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.cluster.id
}

resource "aws_security_group" "nodes" {
  name_prefix = "${local.cluster_name}-nodes-"
  description = "EKS nodes shared security group"
  vpc_id      = var.vpc_id

  tags = merge(local.common_tags, {
    Name                                           = "${local.cluster_name}-nodes-sg"
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  })

  lifecycle { create_before_destroy = true }
}

resource "aws_security_group_rule" "nodes_ingress_self" {
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "-1"
  self              = true
  security_group_id = aws_security_group.nodes.id
  description       = "Node to node all ports"
}

resource "aws_security_group_rule" "nodes_egress_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.nodes.id
}

###############################################################################
# Cluster IAM Role
###############################################################################
resource "aws_iam_role" "cluster" {
  name_prefix        = "${local.cluster_name}-cluster-"
  assume_role_policy = data.aws_iam_policy_document.cluster_assume.json
  tags               = local.common_tags
}

data "aws_iam_policy_document" "cluster_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["eks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_vpc_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.cluster.name
}

###############################################################################
# Node IAM Role
###############################################################################
resource "aws_iam_role" "node" {
  name_prefix        = "${local.cluster_name}-node-"
  assume_role_policy = data.aws_iam_policy_document.node_assume.json
  tags               = local.common_tags
}

data "aws_iam_policy_document" "node_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "node_worker_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_ecr_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_ssm_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.node.name
}

###############################################################################
# Karpenter SQS (for interruption handling)
###############################################################################
resource "aws_sqs_queue" "karpenter" {
  name                      = "${local.cluster_name}-karpenter"
  message_retention_seconds = 300
  sqs_managed_sse_enabled   = true
  tags                      = local.common_tags
}

resource "aws_sqs_queue_policy" "karpenter" {
  queue_url = aws_sqs_queue.karpenter.id
  policy    = data.aws_iam_policy_document.karpenter_sqs.json
}

data "aws_iam_policy_document" "karpenter_sqs" {
  statement {
    sid    = "EC2InterruptionPolicy"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com", "sqs.amazonaws.com"]
    }
    actions   = ["sqs:SendMessage"]
    resources = [aws_sqs_queue.karpenter.arn]
  }
}

# EventBridge rules → SQS for Karpenter
resource "aws_cloudwatch_event_rule" "karpenter_spot_interrupt" {
  name        = "${local.cluster_name}-karpenter-spot-interrupt"
  description = "Karpenter spot interruption"
  event_pattern = jsonencode({
    source      = ["aws.ec2"]
    detail-type = ["EC2 Spot Instance Interruption Warning"]
  })
  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "karpenter_spot_interrupt" {
  rule      = aws_cloudwatch_event_rule.karpenter_spot_interrupt.name
  target_id = "KarpenterSQS"
  arn       = aws_sqs_queue.karpenter.arn
}

resource "aws_cloudwatch_event_rule" "karpenter_health" {
  name        = "${local.cluster_name}-karpenter-health"
  description = "Karpenter health events"
  event_pattern = jsonencode({
    source      = ["aws.health"]
    detail-type = ["AWS Health Event"]
  })
  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "karpenter_health" {
  rule      = aws_cloudwatch_event_rule.karpenter_health.name
  target_id = "KarpenterSQSHealth"
  arn       = aws_sqs_queue.karpenter.arn
}
