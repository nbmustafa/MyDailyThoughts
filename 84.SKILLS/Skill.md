Use skill creator to write a claude skill for my day to day eks cluster debugging supporting massive environemnt with vas components and feafures enabled.

**Skill Name:eks-debug

**Skill Description / Objective:**  
Create a highly specialized, production-grade Claude Skill that acts as my personal **day-to-day EKS debugging co-pilot** for massive, complex, enterprise-scale Amazon EKS environments (thousands of nodes, tens of thousands of pods, 100+ namespaces, dozens of AWS-native and third-party add-ons/features enabled simultaneously).

The skill must be optimized for real-world operational complexity, not simple lab clusters. It should handle noisy, high-scale environments with heavy use of:
- Networking: VPC CNI (custom networking, prefix delegation, security groups for pods), CoreDNS, AWS Load Balancer Controller, ExternalDNS, Ingress-NGINX, Istio/Linkerd, Cilium, etc.
- Storage: EBS CSI, EFS CSI, FSx, Velero, etc.
- IAM & Security: IRSA, Pod Identity, Kyverno/Gatekeeper, Pod Security Policies/Standards, Network Policies, AWS WAF, etc.
- Scaling & Scheduling: Karpenter, Cluster Autoscaler, Descheduler, Vertical Pod Autoscaler, KEDA, etc.
- Observability & Logging: AWS Distro for OpenTelemetry, Prometheus + Grafana + multi-region setups, etc.

**Core Capabilities the Skill Must Deliver:**

1. **Interactive, Context-Aware Debugging Workflow**  
   - Always start by asking smart clarifying questions to quickly understand cluster context, symptoms, recent changes, and scope (single pod, namespace, node, whole cluster, etc.).
   - Follow a structured, repeatable troubleshooting methodology (gather → observe → hypothesize → test → remediate).

2. **Comprehensive Troubleshooting Coverage**  
   - Pod lifecycle issues (Pending, CrashLoopBackOff, Evicted, ImagePullBackOff, OOMKilled, etc.)
   - Networking & DNS failures
   - IAM / IRSA / permission errors
   - Resource scheduling & node pressure problems
   - Scaling & autoscaling failures (Karpenter, Cluster Autoscaler)
   - Load balancer / Ingress / Service issues
   - Storage / PVC / CSI driver problems
   - CNI-specific issues in massive clusters
   - Observability & logging gaps
   - Performance / latency / throttling issues
   - Node termination / spot instance handling
   - Upgrade-related breakage (EKS control plane, add-ons, Kubernetes version skew)

3. **Tool & Command Library**  
   - Provide ready-to-run `kubectl`, `aws cli`, `eksctl`, `kubectx/kubens`, CloudWatch Logs Insights, Prometheus queries, etc.
   - Include one-liners, debug pods, ephemeral containers, and advanced techniques for large-scale clusters (e.g., `kubectl debug`, `kubectl top`, `aws eks describe-cluster`, etc.).

4. **Massive-Environment Specific Guidance**  
   - Highlight scale-specific gotchas (API server throttling, etcd performance, kubelet pressure, control-plane limits, VPC limits, etc.).
   - Recommend observability-first approaches and cost-effective debugging in huge clusters.
   - Include checklists for “noisy neighbor” problems and blast-radius awareness.

5. **Output Style & Behavior**  
   - Always be concise yet actionable — use markdown tables, numbered steps, and prioritized recommendations.
   - Provide risk ratings for remediation steps.
   - Suggest monitoring/alerting improvements after every major incident.
   - Maintain conversation memory across the session (reference previous context).
   - Offer “next best action” at every step.

**Format Instructions for the Skill Creator:**  
Output the complete skill in Claude’s native skill format (system prompt + example user/assistant turns if needed). Make the system prompt extremely detailed, robust, and self-contained so the skill works reliably every single time I invoke it.

---
