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

‐--------------------------------------------
act as a principal engineer to answer below tasks amd questions

we have github action pipeline in step 1 assume GHA-Deploy role with github action assume plugin,

then then in step 2 use aws cli to assume another role  PCCS-Deployer role to login to ecr (i belive this assume role is only active for this step correct me if i am wrong)

then in step 3 it run terraform init/plan/ apply and inside terraform provide it again assumes PCCS-Deployer role to run terraform, part of terraform code is helm_release which pulls oci helm from ecr, will ecr login in step 2 still work for step 3 when it again assume same eole but inside terraform provider


Great questions — let’s walk through this carefully step by step as if I’m reviewing the pipeline design as a principal engineer.


---

🔹 Step 1: GitHub Action assumes GHA-Deploy role

Using the GitHub OIDC federation plugin, your workflow assumes GHA-Deploy in your AWS account.

This role is what gives your pipeline its initial AWS credentials (short-lived STS creds).


✅ This part is correct.


---

🔹 Step 2: Use AWS CLI to assume PCCS-Deployer role to log in to ECR

From inside the workflow, you’re running aws sts assume-role (or aws configure sso equivalent) to switch from GHA-Deploy → PCCS-Deployer.

These credentials are temporary (15min–1h depending on your role/session settings).

When you run aws ecr get-login-password | docker login ..., you’re writing a Docker credential into ~/.docker/config.json that is valid for 12 hours by default (ECR login token TTL).

The role assumption credentials themselves are only used to request that login password from ECR.


👉 Answer to your assumption: Yes, the aws sts assume-role credentials you obtained in Step 2 are only active for the duration of that shell/session or until you overwrite the env vars. Once the step is done, unless you export them to GitHub Actions environment, they are not implicitly carried forward.


---

🔹 Step 3: Terraform init/plan/apply

Your Terraform AWS provider is configured to assume PCCS-Deployer again. That’s fine — it will do its own STS assume role calls.

The helm provider inside Terraform uses helm (or helm registry login) under the hood when dealing with OCI registries like ECR.


⚠️ Key point:

Terraform assuming PCCS-Deployer role does not automatically reuse the ECR docker login you performed in Step 2.

However, since ECR login is stored in ~/.docker/config.json (or $DOCKER_CONFIG), that credential will still be valid (up to 12h) when terraform helm_release pulls the chart.

So as long as your job runner persists the workspace (same container/VM, not a new job), the Step 2 ECR login is still valid in Step 3.



---

✅ Final Answer

Yes, your understanding is mostly correct: the Step 2 assume-role creds are only scoped to that execution (not globally available).

But the ECR login created in Step 2 will still work in Step 3, because it’s a Docker credential written to disk and valid for ~12 hours, independent of the Terraform AWS provider’s role assumption.

As long as you don’t switch runners between steps, Terraform’s helm provider will pick up that credential and be able to pull from ECR.



---

👉 Best practice recommendation:
Instead of relying on Step 2 to “bootstrap” Docker login, you can also:

Use aws ecr get-login-password inside Terraform’s null_resource or pre_hook for helm_release to guarantee freshness.

Or configure Terraform helm provider with an explicit registry_login command so you’re not relying on external step state.



---

Do you want me to sketch a recommended GitHub Actions pipeline structure (with minimal redundant role assumes and a clean ECR login flow inside Terraform)?

