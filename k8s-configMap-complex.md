# Complex Conditional Kubernetes ConfigMap with Terraform

This Terraform configuration creates a ConfigMap that varies by environment (dev, nonprod, prep, prod) with secrets pulled from AWS Secrets Manager, where each environment has different secret names.

## Prerequisites

- Terraform installed
- AWS provider configured
- Kubernetes provider configured
- Appropriate IAM permissions to access AWS Secrets Manager

## Configuration

```hcl
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = "us-east-1" # Change to your region
}

provider "kubernetes" {
  # Configure your Kubernetes provider
  # This could be via kubeconfig file or other authentication methods
}

# Variable for environment (dev, nonprod, prep, prod)
variable "environment" {
  description = "The deployment environment (dev, nonprod, prep, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "nonprod", "prep", "prod"], var.environment)
    error_message = "Environment must be one of: dev, nonprod, prep, prod"
  }
}

# Map of secret names for each environment
locals {
  secret_names = {
    dev     = "dev/app-secrets"
    nonprod = "nonprod/app-secrets"
    prep    = "prep/app-secrets"
    prod    = "prod/app-secrets"
  }
}

# Retrieve secrets from AWS Secrets Manager
data "aws_secretsmanager_secret" "app_secrets" {
  name = local.secret_names[var.environment]
}

data "aws_secretsmanager_secret_version" "current" {
  secret_id = data.aws_secretsmanager_secret.app_secrets.id
}

# Parse the secret JSON (assuming secrets are stored as JSON)
locals {
  secrets = jsondecode(data.aws_secretsmanager_secret_version.current.secret_string)
}

# Common configuration for all environments
locals {
  common_config = {
    APP_NAME          = "my-application"
    LOG_LEVEL         = "info"
    MAX_RETRIES       = 3
    TIMEOUT           = 30
    FEATURE_FLAG_FILE = "/etc/config/feature-flags.json"
  }
}

# Environment-specific configuration
locals {
  env_specific_config = {
    dev = {
      DEBUG_MODE          = "true"
      CACHE_ENABLED       = "false"
      DATABASE_POOL_SIZE  = "5"
      SERVICE_ENDPOINT    = "https://dev.api.example.com"
    }
    nonprod = {
      DEBUG_MODE          = "false"
      CACHE_ENABLED       = "true"
      DATABASE_POOL_SIZE  = "10"
      SERVICE_ENDPOINT    = "https://nonprod.api.example.com"
    }
    prep = {
      DEBUG_MODE          = "false"
      CACHE_ENABLED       = "true"
      DATABASE_POOL_SIZE  = "15"
      SERVICE_ENDPOINT    = "https://prep.api.example.com"
    }
    prod = {
      DEBUG_MODE          = "false"
      CACHE_ENABLED       = "true"
      DATABASE_POOL_SIZE  = "20"
      SERVICE_ENDPOINT    = "https://api.example.com"
    }
  }
}

# Merge all configurations
locals {
  full_config = merge(
    local.common_config,
    local.env_specific_config[var.environment],
    {
      DB_USERNAME     = local.secrets["db_username"]
      DB_PASSWORD     = local.secrets["db_password"]
      API_KEY         = local.secrets["api_key"]
      ENCRYPTION_KEY  = local.secrets["encryption_key"]
    }
  )
}

# Create the ConfigMap
resource "kubernetes_config_map" "app_config" {
  metadata {
    name      = "app-config-${var.environment}"
    namespace = "default"
    labels = {
      app     = "my-application"
      env     = var.environment
      version = "v1.0"
    }
  }

  data = local.full_config
}

# Output the ConfigMap name for reference
output "config_map_name" {
  value = kubernetes_config_map.app_config.metadata[0].name
}
```

## Explanation

1. **Providers**: Configures AWS and Kubernetes providers.

2. **Environment Variable**: Defines the environment (dev, nonprod, prep, prod) with validation.

3. **Secret Names Mapping**: Maps each environment to its corresponding secret name in AWS Secrets Manager.

4. **Secret Retrieval**: Fetches the secret from AWS Secrets Manager based on the current environment.

5. **Configuration Structure**:
   - `common_config`: Settings shared across all environments
   - `env_specific_config`: Settings that vary by environment
   - `full_config`: Merges common, environment-specific, and secret values

6. **ConfigMap Creation**: Creates a Kubernetes ConfigMap with:
   - Environment-specific name
   - Proper labels for identification
   - All merged configuration values

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Apply for a specific environment (e.g., prod):
   ```bash
   terraform apply -var="environment=prod"
   ```

3. To deploy to another environment, change the environment variable:
   ```bash
   terraform apply -var="environment=dev"
   ```

## Notes

- This assumes your secrets are stored in AWS Secrets Manager as JSON strings.
- The secret structure should include keys like `db_username`, `db_password`, etc.
- Adjust the configuration keys and values according to your application's needs.
- Consider using Terraform workspaces for more complex multi-environment management.