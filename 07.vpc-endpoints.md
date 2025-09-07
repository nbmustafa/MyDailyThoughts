Background:

AWS Service VPC Endpoints are deployed centrally in network common account
Hosted zone IDs associated to the endpoints are exported via SSM parameter
Other network types read this SSM parameter and associate their VPCs to it (stored in separate state)

Reference: https://github.com/nbmustafa/terraform-shared-vpc-endpoint


```
# --------------------------------------------------
# VPC Endpoint - Interface Type
# --------------------------------------------------

# Over-ride map for aws service end points where dns entry for service name
# does not match the service name
# ex:
# "kinesis-firehose" = "firehose"
locals {
  endpoint_name_override = {
    "kinesis-streams" = "kinesis"
  }
}

resource "aws_security_group" "interface_endpoints" {
  provider = aws.network_common
  for_each = local.common_networks
  # Terraform resource name change is destructive, platform resources prefixed with "core" for IAM policy simplification
  name_prefix = "core-interface-endpoints-"
  vpc_id      = aws_vpc.common_vpc[each.key].id

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-sg-${each.key}" },
    local.tags
  )
}

resource "aws_security_group_rule" "interface_endpoints_rule_1" {
  provider  = aws.network_common
  for_each  = local.common_networks
  type      = "ingress"
  from_port = 443
  to_port   = 443
  protocol  = "tcp"
  cidr_blocks = distinct(concat(
    local.aws_internal_cidr, // should be aws internals
    lookup(local.network_aliases, "on_prem_cidr_uk_for_endpoint_subnets", [])
  ))
  security_group_id = aws_security_group.interface_endpoints[each.key].id
}

resource "aws_vpc_endpoint" "interface" {
  provider           = aws.network_common
  for_each           = local.interface_services
  vpc_id             = aws_vpc.common_vpc[split(".", each.value)[0]].id
  service_name       = "com.amazonaws.${var.common_master_region}.${each.key}"
  vpc_endpoint_type  = "Interface"
  security_group_ids = [aws_security_group.interface_endpoints[split(".", each.value)[0]].id]
  subnet_ids = [
    aws_subnet.common_private_subnet["${each.value}_a0"].id,
    aws_subnet.common_private_subnet["${each.value}_b0"].id,
    aws_subnet.common_private_subnet["${each.value}_c0"].id
  ]
  private_dns_enabled = false

  policy = <<POLICY
{
  "Statement": [
    {
      "Action": "*",
      "Effect": "Allow",
      "Resource": "*",
      "Principal": "*"%{if each.key == "mgn"},
      "Condition": {
        "StringEquals": {
          "aws:PrincipalOrgID": "${local.core_organisation_id}"
        }
      }%{endif}
    }
  ]
}
POLICY

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-${each.key}" },
    local.tags
  )
}

# Defined ignore_changes to allow downstream VPCs to add associations without Terraform deleting via drift alignment
resource "aws_route53_zone" "interface" {
  provider = aws.network_common
  for_each = local.interface_services
  name     = join(".", reverse(split(".", "com.amazonaws.${var.common_master_region}.${lookup(local.endpoint_name_override, each.key, each.key)}")))
  comment  = "Hosted Zone for Interface Endpoints"

  dynamic "vpc" {
    for_each = local.common_networks
    content {
      vpc_id = aws_vpc.common_vpc[vpc.key].id
    }
  }

  lifecycle {
    ignore_changes = [vpc]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-interface-private-hosted-zone" },
    local.tags
  )
}

# AWSBS11-927 Due to ignore_changes argument in the above resource (aws_route53_zone.interface), adding additional dynamic blocks will not take effect
locals {
  /* Resultant map after merge(flatten([local.sscontainer_vpc_to_interface])...) looks like:
    interface_to_sscontainer = {
      "vpc0_ec2" = "Z0344866W2B9O1ABC123",
      "vpc0_kms" = "Z0344866W2B9OMABC234",
      "vpc0_ecr.dkr" = "Z0344866W2B9OMABC234",
    }

    Note: using "_" delimiter to prevent issues with string split when service is dot-separated like ecr.dkr
  */
  interface_to_sscontainer = [
    for vpc_alias in keys(local.sscontainer_networks) : {
      for service in keys(local.interface_services) :
      "${vpc_alias}_${service}" => true
    }
  ]
}

resource "aws_route53_zone_association" "interface_to_sscontainer" {
  provider = aws.network_common
  for_each = merge(flatten([local.interface_to_sscontainer])...) == null ? {} : merge(flatten([local.interface_to_sscontainer])...)
  zone_id  = aws_route53_zone.interface[split("_", each.key)[1]].zone_id
  vpc_id   = aws_vpc.sscontainer_vpc[split("_", each.key)[0]].id
}

resource "aws_route53_record" "interface" {
  provider = aws.network_common
  for_each = local.interface_services
  zone_id  = aws_route53_zone.interface[each.key].zone_id
  name     = join(".", reverse(split(".", "com.amazonaws.${var.common_master_region}.${lookup(local.endpoint_name_override, each.key, each.key)}")))
  type     = "A"

  alias {
    name                   = lookup(aws_vpc_endpoint.interface[each.key].dns_entry[0], "dns_name")
    zone_id                = lookup(aws_vpc_endpoint.interface[each.key].dns_entry[0], "hosted_zone_id")
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "interface_wildcard" {
  provider = aws.network_common
  for_each = local.interface_services
  zone_id  = aws_route53_zone.interface[each.key].zone_id
  name     = join(".", reverse(split(".", "com.amazonaws.${var.common_master_region}.${lookup(local.endpoint_name_override, each.key, each.key)}.*")))
  type     = "CNAME"
  ttl      = "60"
  records  = [join(".", reverse(split(".", "com.amazonaws.${var.common_master_region}.${lookup(local.endpoint_name_override, each.key, each.key)}")))]
}

# --------------------------------------------------------------------------
# VPC Endpoint - Interface Type (Sagemaker notebook)
# - Sagemaker notebook endpoint does not follow same service naming format
# - only created once in vpc0
# - uncomment when required
# --------------------------------------------------------------------------
resource "aws_security_group" "sagemaker_studio" {
  count       = local.sagemaker_required ? 1 : 0
  provider    = aws.network_common
  name_prefix = "core-interface-endpoints-sagemaker"
  vpc_id      = aws_vpc.common_vpc["vpc0"].id

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-sg-sagemaker-studio" },
    local.tags
  )
}

resource "aws_security_group_rule" "sagemaker_studio" {
  count             = local.sagemaker_required ? 1 : 0
  provider          = aws.network_common
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = local.aws_internal_cidr // should be aws internals
  security_group_id = aws_security_group.sagemaker_studio[count.index].id
}

resource "aws_route53_zone" "interface_sagemaker_studio" {
  count    = local.sagemaker_required ? 1 : 0
  provider = aws.network_common
  name     = "studio.${var.common_master_region}.sagemaker.aws"
  comment  = "Hosted Zone for Sagemaker studio Interface Endpoint"

  vpc {
    vpc_id = aws_vpc.common_vpc["vpc0"].id
  }

  lifecycle {
    ignore_changes = [vpc]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-interface-sagemaker-studio" },
    local.tags
  )
}

resource "aws_vpc_endpoint" "interface_sagemaker_studio" {
  count              = local.sagemaker_required ? 1 : 0
  provider           = aws.network_common
  vpc_id             = aws_vpc.common_vpc["vpc0"].id
  service_name       = "aws.sagemaker.${var.common_master_region}.studio"
  vpc_endpoint_type  = "Interface"
  security_group_ids = [aws_security_group.sagemaker_studio[count.index].id]
  subnet_ids = [
    aws_subnet.common_private_subnet["vpc0.endpoint_a0"].id,
    aws_subnet.common_private_subnet["vpc0.endpoint_b0"].id,
    aws_subnet.common_private_subnet["vpc0.endpoint_c0"].id,
  ]
  private_dns_enabled = false

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-sagemaker-studio" },
    local.tags
  )
}

resource "aws_route53_record" "interface_sagemaker_studio" {
  count    = local.sagemaker_required ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.interface_sagemaker_studio[count.index].zone_id
  name     = "studio.${var.common_master_region}.sagemaker.aws"
  type     = "A"

  alias {
    name                   = lookup(aws_vpc_endpoint.interface_sagemaker_studio[count.index].dns_entry[4], "dns_name") #needs to be index of 4 because there are 8 entries all together
    zone_id                = lookup(aws_vpc_endpoint.interface_sagemaker_studio[count.index].dns_entry[4], "hosted_zone_id")
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "interface_sagemaker_studio_wildcard" {
  count    = local.sagemaker_required ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.interface_sagemaker_studio[count.index].zone_id
  name     = "*.studio.${var.common_master_region}.sagemaker.aws"
  type     = "CNAME"
  ttl      = 3600
  records  = ["studio.${var.common_master_region}.sagemaker.aws"]
}


# --------------------------------------------------------------------------
# VPC Endpoint - Interface Type (S3 Privatelink endpoint)
# --------------------------------------------------------------------------
resource "aws_security_group" "interface_onprem_s3" {
  provider    = aws.network_common
  name_prefix = "core-${local.common_prefix}-interface-endpoints-s3-onprem"
  vpc_id      = aws_vpc.common_vpc["vpc0"].id
  lifecycle {
    create_before_destroy = true
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-sg-s3-onprem" },
    local.tags
  )
}

resource "aws_security_group_rule" "interface_onprem_s3" {
  provider          = aws.network_common
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = local.anz_internal_cidrs
  security_group_id = aws_security_group.interface_onprem_s3.id
}

resource "aws_vpc_endpoint" "interface_onprem_s3" {
  provider           = aws.network_common
  vpc_id             = aws_vpc.common_vpc["vpc0"].id
  service_name       = "com.amazonaws.${var.common_master_region}.s3"
  vpc_endpoint_type  = "Interface"
  security_group_ids = [aws_security_group.interface_onprem_s3.id]
  subnet_ids = [
    aws_subnet.common_private_subnet["vpc0.endpoint_a0"].id,
    aws_subnet.common_private_subnet["vpc0.endpoint_b0"].id,
    aws_subnet.common_private_subnet["vpc0.endpoint_c0"].id,
  ]
  private_dns_enabled = false

  # Ignore changes to policy, policy updates are triggered by av-orchestration pipeline execution
  lifecycle {
    ignore_changes = [
      policy,
    ]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-onprem-s3" },
    local.tags
  )
}

resource "aws_route53_record" "interface_onprem_s3" {
  provider = aws.sharedservices
  zone_id  = aws_route53_zone.shared_services.zone_id
  name     = "s3-privatelink-${local.common_prefix}"
  type     = "CNAME"
  ttl      = "60"
  records  = [aws_vpc_endpoint.interface_onprem_s3.dns_entry[0]["dns_name"]]
}

resource "aws_s3_bucket_object" "interface_onprem_s3_vpce_id" {
  provider     = aws.network_common
  bucket       = local.facts_bucket
  content_type = "application/json"
  acl          = "bucket-owner-full-control"
  key          = "/core/v2/network-common/s3-common-vpce-id"
  content = jsonencode(
    aws_vpc_endpoint.interface_onprem_s3.id,
  )
}

# Overriding Hosted Zone for requirements to use S3 Interface Endpoint for all workloads e.g. Rehost requirements
resource "aws_route53_zone" "interface_onprem_s3_for_rehost" {
  count    = local.enable_rehost ? 1 : 0
  provider = aws.network_common
  name     = "s3.${var.common_master_region}.amazonaws.com"
  comment  = "Hosted Zone for On-Prem S3 Interface Endpoints"

  dynamic "vpc" {
    for_each = local.common_networks
    content {
      vpc_id = aws_vpc.common_vpc[vpc.key].id
    }
  }

  lifecycle {
    ignore_changes = [vpc]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-interface-private-hosted-zone-s3-onprem" },
    local.tags
  )
}

resource "aws_route53_record" "interface_onprem_s3_for_rehost" {
  count    = local.enable_rehost ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.interface_onprem_s3_for_rehost[count.index].zone_id
  name     = "s3.${var.common_master_region}.amazonaws.com"
  type     = "A"

  alias {
    name                   = lookup(aws_vpc_endpoint.interface_onprem_s3.dns_entry[0], "dns_name")
    zone_id                = lookup(aws_vpc_endpoint.interface_onprem_s3.dns_entry[0], "hosted_zone_id")
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "interface_onprem_s3_for_rehost_wildcard" {
  count    = local.enable_rehost ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.interface_onprem_s3_for_rehost[count.index].zone_id
  name     = "*.s3.${var.common_master_region}.amazonaws.com"
  type     = "CNAME"
  ttl      = "60"
  records  = ["s3.${var.common_master_region}.amazonaws.com"]
}

# --------------------------------------------------
# VPC Endpoint - Gateway Type - S3
# --------------------------------------------------
resource "aws_vpc_endpoint" "common_s3" {
  provider     = aws.network_common
  for_each     = local.common_networks
  vpc_id       = aws_vpc.common_vpc[each.key].id
  service_name = "com.amazonaws.${var.common_master_region}.s3"

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-s3-common-${each.key}" },
    local.tags
  )
}

resource "aws_vpc_endpoint_route_table_association" "common_private_rt_s3" {
  provider        = aws.network_common
  for_each        = toset(local.common_vpc_to_trustlevels)
  vpc_endpoint_id = aws_vpc_endpoint.common_s3[split(".", each.value)[0]].id
  route_table_id  = aws_route_table.common_private_route_table[each.key].id
}

resource "aws_vpc_endpoint" "sscontainer_s3" {
  provider     = aws.network_common
  for_each     = local.sscontainer_networks
  vpc_id       = aws_vpc.sscontainer_vpc[each.key].id
  service_name = "com.amazonaws.${var.common_master_region}.s3"

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-s3-sscontainer-${each.key}" },
    local.tags
  )
}

resource "aws_vpc_endpoint_route_table_association" "sscontainer_private_rt_s3" {
  provider        = aws.network_common
  for_each        = toset(local.sscontainer_vpc_to_trustlevels)
  vpc_endpoint_id = aws_vpc_endpoint.sscontainer_s3[split(".", each.value)[0]].id
  route_table_id  = aws_route_table.sscontainer_private_route_table[each.key].id
}

# --------------------------------------------------
# VPC Endpoint - Gateway Type - DynamoDB
# --------------------------------------------------
resource "aws_vpc_endpoint" "common_dynamodb" {
  provider     = aws.network_common
  for_each     = local.common_networks
  vpc_id       = aws_vpc.common_vpc[each.key].id
  service_name = "com.amazonaws.${var.common_master_region}.dynamodb"

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-ddb-common-${each.key}" },
    local.tags
  )
}

resource "aws_vpc_endpoint_route_table_association" "common_private_rt_dynamodb" {
  provider        = aws.network_common
  for_each        = toset(local.common_vpc_to_trustlevels)
  vpc_endpoint_id = aws_vpc_endpoint.common_dynamodb[split(".", each.value)[0]].id
  route_table_id  = aws_route_table.common_private_route_table[each.key].id
}

resource "aws_vpc_endpoint" "sscontainer_dynamodb" {
  provider     = aws.network_common
  for_each     = local.sscontainer_networks
  vpc_id       = aws_vpc.sscontainer_vpc[each.key].id
  service_name = "com.amazonaws.${var.common_master_region}.dynamodb"

  tags = merge(
    { "Name" = "${local.common_prefix}-endpoint-ddb-sscontainer-${each.key}" },
    local.tags
  )
}

resource "aws_vpc_endpoint_route_table_association" "sscontainer_private_rt_dynamodb" {
  provider        = aws.network_common
  for_each        = toset(local.sscontainer_vpc_to_trustlevels)
  vpc_endpoint_id = aws_vpc_endpoint.sscontainer_dynamodb[split(".", each.value)[0]].id
  route_table_id  = aws_route_table.sscontainer_private_route_table[each.key].id
}

# --------------------------------------------------
# SSM output for Endpoint IDs
# --------------------------------------------------
resource "aws_ssm_parameter" "network_common_endpoint_hosted_zones" {
  provider = aws.network_common
  name     = "core-network-common-${var.network_environment}-endpoint-hosted-zones"
  type     = "String"
  value = jsonencode(
    merge(
      zipmap(keys(local.interface_services), [for key, value in aws_route53_zone.interface : value.zone_id]),
      { "sagemaker_studio" = length(aws_route53_zone.interface_sagemaker_studio) > 0 ? aws_route53_zone.interface_sagemaker_studio[0].zone_id : null },
      local.enable_rehost ? { "onprem_s3" = aws_route53_zone.interface_onprem_s3_for_rehost[0].zone_id } : {}
    )
  )
  tags = merge(
    { "Name" = "core-network-common-${var.network_environment}-endpoint-hosted-zones" },
    local.tags
  )
}

resource "aws_ssm_parameter" "network_common_endpoint_ids" {
  provider = aws.network_common
  name     = "core-network-common-${var.network_environment}-endpoint-ids"
  type     = "String"
  value = jsonencode(
    merge(
      zipmap(keys(local.interface_services), [for key, value in aws_vpc_endpoint.interface : value.id]),
      { "sagemaker_studio" = length(aws_vpc_endpoint.interface_sagemaker_studio) > 0 ? aws_vpc_endpoint.interface_sagemaker_studio[0].id : null }
    )
  )
  tags = merge(
    { "Name" = "core-network-common-${var.network_environment}-endpoint-ids" },
    local.tags
  )
}
```

### R53 Hosted zone association:
```
resource "aws_route53_zone" "shared_services" {
  provider = aws.sharedservices
  name     = "svc.${local.regional_hosted_zone}"
  comment  = "Hosted Zone for Shared Services"
  vpc {
    vpc_id = aws_vpc.common_vpc["vpc0"].id
  }

  lifecycle {
    ignore_changes = [vpc]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-svc-private-hosted-zone" },
    local.tags
  )
}

resource "aws_route53_zone_association" "shared_services_to_sscontainer" {
  provider = aws.sharedservices
  for_each = local.sscontainer_networks
  zone_id  = aws_route53_zone.shared_services.zone_id
  vpc_id   = aws_vpc.sscontainer_vpc[each.key].id
}

resource "aws_ssm_parameter" "network_sharedservices_hosted_zones" {
  provider = aws.sharedservices
  name     = "core-network-sharedservices-${var.network_environment}-hosted-zones"
  type     = "String"

  value = jsonencode({
    "svc" : {
      "zone_id" : aws_route53_zone.shared_services.zone_id
    }
  })
  tags = merge(
    { "Name" = "core-network-sharedservices-${var.network_environment}-hosted-zones" },
    local.tags
  )
}

resource "aws_route53_zone" "googleapis" {
  provider = aws.network_common
  name     = "googleapis.com"
  comment  = "Hosted Zone for googleapis"

  dynamic "vpc" {
    for_each = aws_vpc.common_vpc
    content {
      vpc_id = vpc.value.id
    }
  }

  lifecycle {
    ignore_changes = [vpc]
  }

  tags = merge(
    { "Name" = "${local.common_prefix}-googleapis-private-hosted-zone" },
    local.tags
  )
}

resource "aws_route53_zone_association" "googleapis_to_sscontainer" {
  provider = aws.network_common
  for_each = aws_vpc.sscontainer_vpc
  zone_id  = aws_route53_zone.googleapis.zone_id
  vpc_id   = aws_vpc.sscontainer_vpc[each.key].id
}

resource "aws_route53_record" "googleapis_wildcard" {
  count    = length(module.common_networks_map.restricted_googleapis_ips) > 0 ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.googleapis.zone_id
  name     = "*"
  type     = "A"
  ttl      = "60"
  records  = module.common_networks_map.restricted_googleapis_ips
}

resource "aws_route53_record" "googleapis_pubsub" {
  count    = length(module.common_networks_map.restricted_googleapis_ips) > 0 ? 1 : 0
  provider = aws.network_common
  zone_id  = aws_route53_zone.googleapis.zone_id
  name     = "pubsub"
  type     = "A"
  ttl      = "60"
  records  = module.common_networks_map.restricted_googleapis_ips
}
```

