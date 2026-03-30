Here's a breakdown of what the script does across its 8 pipeline stages:

---

**Stage 1 — Context Switch**
Runs `aws eks update-kubeconfig --name <cluster> --region <region>` and verifies the active context. Supports `--profile` for multi-account setups.

**Stage 2 — Cluster Data Collection**
Collects via `kubectl -o json`: nodes, pods (all namespaces), deployments, statefulsets, daemonsets, HPAs, PVCs, services, namespaces, plus `kubectl top nodes` and `kubectl top pods` for real utilisation data.

**Stage 3 — Karpenter Manifest Reader**
Recursively scans the `--repo` path for `*.yaml`/`*.yml`, parses all documents, and extracts `NodePool` and `EC2NodeClass` objects into the snapshot.

**Stage 4 — Heuristic Pre-Analysis**
Before the AI call, derives: node summaries with instance type / zone / nodepool labels, idle/zero-replica workloads, containers with missing or poorly-set limits (request < 25% of limit), and unbound PVCs.

**Stage 5 — Prompt Engineering**
Builds a ~10-section structured prompt injecting all the above as a single JSON payload. The AI is instructed to produce: executive summary with a 1–10 cost-health score, node analysis, Karpenter config review with YAML diffs, workload right-sizing tables, HPA review, idle resource cleanup commands, storage optimisation advice, quick wins, and an estimated savings table.

**Stage 6 — GenAI Studio API Call**
Supports any OpenAI-compatible endpoint (`/v1/chat/completions`) or Anthropic-style response schema. Token via `--ai-token` or `GENAI_API_TOKEN` env var.

**Stages 7–8 — Report Output**
Writes three files: a `.md` report, a styled `.html` report with a dashboard meta-bar, and a `-snapshot.json` audit trail.

**Usage:**
```bash
pip install requests pyyaml markdown

export GENAI_API_TOKEN="your-token"

python eks_cost_optimizer.py \
  --cluster my-prod-cluster \
  --region eu-west-1 \
  --repo ./infra/karpenter \
  --ai-url https://genai-studio.internal/v1/chat/completions \
  --ai-model gpt-4o \
  --output ./reports

# Test without hitting the AI API
python eks_cost_optimizer.py ... --dry-run
```