---
name: k8s-diagnostics
description: >
  Kubernetes cluster diagnostic skill. Use this skill whenever a user asks to
  diagnose, investigate, check, audit, or troubleshoot a Kubernetes cluster or
  namespace. Triggers on: "check my cluster", "diagnose namespace X", "what's
  wrong with my k8s", "run diagnostics on cluster", "investigate pod issues",
  "check storage/networking/nodes", or any mention of cluster health, pod
  crashes, CrashLoopBackOff, node pressure, PVC issues, NetworkPolicy problems,
  Calico issues, or event analysis. Always use this skill — even for partial
  phrases like "something's wrong in my cluster" or "pods are failing in prod".
---

# Kubernetes Diagnostics Skill

This skill runs the `k8sdiag` Go tool to perform comprehensive Kubernetes
cluster diagnostics and presents findings clearly to the user.

## Scope

| Trigger | Behaviour |
|---------|-----------|
| User provides cluster name only | Cluster-wide scan across all namespaces |
| User provides cluster name + namespace | Scoped scan within that namespace only |

---

## Step 1 — Gather Intent

Parse the user's message to extract:

- **cluster_name** (required) — the cluster context name or alias
- **namespace** (optional) — specific namespace to scope the investigation

If `cluster_name` is missing, ask:
> "Which cluster would you like me to diagnose? Please provide the cluster name or kubeconfig context name."

If `namespace` is present, confirm the scoped investigation:
> "I'll run a targeted diagnostic on the `<namespace>` namespace in cluster `<cluster_name>`."

Otherwise confirm cluster-wide:
> "I'll run a full cluster-wide diagnostic on `<cluster_name>`."

---

## Step 2 — Locate the Binary

```bash
# Check standard installation paths in order
BINARY=""
for p in \
  /usr/local/bin/k8sdiag \
  $HOME/.local/bin/k8sdiag \
  ./bin/k8sdiag \
  ./k8sdiag; do
  if [ -x "$p" ]; then
    BINARY="$p"
    break
  fi
done

if [ -z "$BINARY" ]; then
  echo "NOT_FOUND"
fi
echo "$BINARY"
```

**If binary is NOT found**, tell the user:

> "The `k8sdiag` binary is not installed. To install it:
> ```bash
> git clone https://github.com/your-org/k8sdiag
> cd k8sdiag
> make install
> ```
> Once installed, ask me to run the diagnostic again."

Do not proceed without the binary.

---

## Step 3 — Run the Diagnostic

### Cluster-wide scan (no namespace)
```bash
k8sdiag \
  --cluster "<cluster_name>" \
  --output text \
  --timeout 90s
```

### Namespace-scoped scan
```bash
k8sdiag \
  --cluster "<cluster_name>" \
  --namespace "<namespace>" \
  --output text \
  --timeout 60s
```

### For CI / structured output (when user asks for JSON/report)
```bash
k8sdiag \
  --cluster "<cluster_name>" \
  [--namespace "<namespace>"] \
  --output json \
  --no-color
```

**Capture both stdout and exit code:**
```bash
output=$(k8sdiag --cluster "$CLUSTER" [--namespace "$NS"] --output json --no-color 2>/dev/null)
exit_code=$?
```

Exit codes:
- `0` = clean (no findings)
- `1` = warnings only
- `2` = critical findings present

---

## Step 4 — Parse and Present Findings

### If `--output text` (default interactive use)
Stream the tool output directly to the user — the tool already applies colour
formatting. After the output, provide a conversational summary:

> Here's what I found on cluster `<cluster_name>`:
>
> - **X critical issues** require immediate attention
> - **Y warnings** should be reviewed  
> - **Z informational** notes for best-practice improvements
>
> [Highlight the top 3 most severe findings in plain language]

### If `--output json` (structured analysis)
Parse the JSON and build a structured response:

```
report.Summary       → overall counts
report.Results[]     → per-category findings
  .Category
  .Findings[]
    .Severity        → CRITICAL | WARNING | INFO
    .Title
    .Description
    .Resource        → namespace/name
    .Suggestion      → remediation hint
```

Present findings grouped by severity, highest first.

### If no findings
> "✅ Cluster `<cluster_name>` looks healthy! No issues were detected
> [in namespace `<namespace>`]."

---

## Step 5 — Offer Next Steps

After presenting findings, offer relevant follow-up actions based on what
was found. Choose from the list below — only suggest what's relevant:

**If pod issues found:**
> "Would you like me to pull logs for any of the failing pods, or describe
> the pod events in more detail?"

**If storage issues found:**
> "I can show you the current PVC/PV status or help you debug the provisioner."

**If node issues found:**
> "I can help you cordon/drain the affected node or check kubelet logs."

**If network/Calico issues found:**
> "Would you like me to trace the NetworkPolicy rules affecting a specific pod
> or check Calico node status?"

**If scheduling/affinity issues found:**
> "I can explain the exact scheduler constraints blocking the pod, or help
> you adjust affinity rules."

**If RBAC issues found:**
> "I can audit the full set of permissions for a specific service account
> or suggest a least-privilege role."

---

## Diagnostic Coverage Reference

Read `references/coverage.md` for the full list of what each checker covers.
Reference this when users ask "does it check X?" or when customising skip flags.

---

## Flags Cheatsheet

| Scenario | Flags to add |
|----------|-------------|
| Disable colour (CI) | `--no-color` |
| Skip slow checks | `--skip RBAC,Config` |
| Short timeout | `--timeout 30s` |
| Markdown report | `--output markdown` |
| Explicit kubeconfig | `--kubeconfig /path/to/config` |
| Explicit context | `--context arn:aws:eks:...` |

---

## Error Handling

| Error | Cause | Resolution to suggest |
|-------|-------|----------------------|
| `cannot resolve cluster context` | Cluster name not in kubeconfig | List available contexts: `kubectl config get-contexts` |
| `failed to connect` | Auth or network issue | Check kubeconfig credentials and cluster endpoint reachability |
| `context deadline exceeded` | Cluster too slow / large | Retry with `--namespace` to scope checks, or `--timeout 120s` |
| `permission denied` | RBAC | Tool needs at least `get/list` on all core resources. Check ServiceAccount |
| Binary not found | Not installed | Guide user through `make install` |
