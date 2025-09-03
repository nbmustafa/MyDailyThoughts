Configuring ECR Cross-Account Access for PCCS-Deployer Roles

I'll help you configure an ECR policy that allows any AWS account with an IAM role named "PCCS-Deployer" to pull images and Helm charts from your ECR repository.

ECR Repository Policy

Create this JSON policy for your ECR repository:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowPCCSDeployerCrossAccountPull",
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages"
      ],
      "Condition": {
        "ArnLike": {
          "aws:PrincipalArn": "arn:aws:iam::*:role/PCCS-Deployer"
        }
      }
    }
  ]
}
```

Apply the Policy

Use the AWS CLI to set this policy on your ECR repository:

```bash
aws ecr set-repository-policy \
    --repository-name <YOUR_REPO_NAME> \
    --policy-text file://ecr-policy.json \
    --region <YOUR_REGION>
```

IAM Policy for PCCS-Deployer Roles

Each account's PCCS-Deployer role needs these permissions. Attach this IAM policy to the PCCS-Deployer role in each consumer account:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchCheckLayerAvailability"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages"
      ],
      "Resource": "arn:aws:ecr:<REGION>:<OWNER_ACCOUNT_ID>:repository/<REPO_NAME>"
    }
  ]
}
```

Cross-Account Authentication & Pull Commands

From any account with a PCCS-Deployer role, use these commands to authenticate and pull:

For Docker Images:

```bash
# Authenticate to ECR
aws ecr get-login-password --region <REGION> | docker login --username AWS --password-stdin <OWNER_ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com

# Pull image
docker pull <OWNER_ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/<REPO_NAME>:<TAG>
```

For Helm OCI Charts:

```bash
# Authenticate to ECR
aws ecr get-login-password --region <REGION> | helm registry login --username AWS --password-stdin <OWNER_ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com

# Pull Helm chart
helm pull oci://<OWNER_ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/<REPO_NAME>/<CHART_NAME>
```

Important Considerations

1. Security: This configuration allows any AWS account with a role named "PCCS-Deployer" to access your repository. Ensure this meets your security requirements.
2. Role Names: The policy specifically looks for the exact role name "PCCS-Deployer". Any variations (different cases, paths, etc.) won't match.
3. Region Specificity: The ECR policy is region-specific. If your repository exists in multiple regions, you'll need to apply this policy in each region.
4. Monitoring: Consider enabling AWS CloudTrail to monitor access patterns and detect any unauthorized access attempts.
5. Alternative Approach: For enhanced security, you could implement a more restrictive condition that only allows specific accounts:

```json
"Condition": {
  "ArnLike": {
    "aws:PrincipalArn": [
      "arn:aws:iam::123456789012:role/PCCS-Deployer",
      "arn:aws:iam::345678901234:role/PCCS-Deployer"
    ]
  }
}
```

This configuration allows any AWS account with a properly configured PCCS-Deployer role to pull from your ECR repository while maintaining security through IAM role-based access control.