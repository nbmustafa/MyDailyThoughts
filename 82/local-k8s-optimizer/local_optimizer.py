#!/usr/bin/env python3
"""
EKS Cost Optimisation Report Generator
=======================================
Principal Engineer-grade script that:
  - Switches kubectl context to a target EKS cluster
  - Collects comprehensive resource usage & configuration data
  - Reads Karpenter NodePool / NodeClass manifests from a local repo path
  - Sends all findings to a corporate GenAI Studio endpoint
  - Produces a richly formatted Markdown + HTML cost-optimisation report

Usage:
  python eks_cost_optimizer.py \
    --cluster <eks-cluster-name> \
    --region  <aws-region> \
    --repo    <path/to/karpenter/manifests> \
    --ai-url  https://genai-studio.internal/v1/chat/completions \
    --ai-model gpt-4o \
    --output  ./cost-reports

Environment variables:
  GENAI_API_TOKEN   Bearer token for the GenAI Studio (alternative to --ai-token)

Dependencies:
  pip install requests pyyaml markdown
"""

from __future__ import annotations

import argparse
import json
import logging
import os
import subprocess
import sys
import textwrap
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any

import requests
import yaml

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
log = logging.getLogger("eks-cost-optimizer")


# ---------------------------------------------------------------------------
# Data containers
# ---------------------------------------------------------------------------

@dataclass
class ClusterSnapshot:
    cluster_name: str
    region: str
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat())

    # Raw kubectl / top output
    nodes: list[dict]               = field(default_factory=list)
    node_top: str                   = ""
    pods_all: list[dict]            = field(default_factory=list)
    pod_top: str                    = ""
    deployments: list[dict]         = field(default_factory=list)
    statefulsets: list[dict]        = field(default_factory=list)
    daemonsets: list[dict]          = field(default_factory=list)
    hpas: list[dict]                = field(default_factory=list)
    pvcs: list[dict]                = field(default_factory=list)
    services: list[dict]            = field(default_factory=list)
    namespaces: list[str]           = field(default_factory=list)

    # Karpenter config read from repo
    nodepools: list[dict]           = field(default_factory=list)
    nodeclasses: list[dict]         = field(default_factory=list)

    # Derived summaries (populated by analyse())
    node_summary: list[dict]        = field(default_factory=list)
    idle_workloads: list[dict]      = field(default_factory=list)
    oversized_workloads: list[dict] = field(default_factory=list)
    unbound_pvcs: list[dict]        = field(default_factory=list)


# ---------------------------------------------------------------------------
# kubectl / shell helpers
# ---------------------------------------------------------------------------

def run(cmd: list[str], *, check: bool = True) -> str:
    """Execute a shell command and return stdout."""
    log.debug("RUN: %s", " ".join(cmd))
    result = subprocess.run(cmd, capture_output=True, text=True, check=False)
    if check and result.returncode != 0:
        log.warning(
            "Command failed (rc=%d): %s\nSTDERR: %s",
            result.returncode, " ".join(cmd), result.stderr.strip(),
        )
        return ""
    return result.stdout.strip()


def kubectl(*args: str) -> str:
    return run(["kubectl", *args])


def kubectl_json(*args: str) -> list[dict]:
    """Run kubectl with -o json and return the items list (or the root object)."""
    raw = kubectl(*args, "-o", "json")
    if not raw:
        return []
    try:
        parsed = json.loads(raw)
        if isinstance(parsed, dict):
            return parsed.get("items", [parsed])
        return parsed
    except json.JSONDecodeError:
        log.warning("Failed to parse JSON from: kubectl %s", " ".join(args))
        return []


# ---------------------------------------------------------------------------
# 1. Context switch
# ---------------------------------------------------------------------------

def switch_context(cluster_name: str, region: str, profile: str | None) -> None:
    """Update kubeconfig to point at the target EKS cluster."""
    log.info("Switching kubectl context → cluster=%s  region=%s", cluster_name, region)
    cmd = [
        "aws", "eks", "update-kubeconfig",
        "--name", cluster_name,
        "--region", region,
    ]
    if profile:
        cmd += ["--profile", profile]
    out = run(cmd)
    log.info("%s", out or "Kubeconfig updated.")
    ctx = kubectl("config", "current-context")
    log.info("Active context: %s", ctx)


# ---------------------------------------------------------------------------
# 2. Cluster data collection
# ---------------------------------------------------------------------------

def collect_cluster_data(snap: ClusterSnapshot) -> None:
    """Collect live cluster data via kubectl into the snapshot."""
    log.info("─── Collecting cluster resources ───────────────────────────────")

    log.info("  nodes …")
    snap.nodes = kubectl_json("get", "nodes")

    log.info("  kubectl top nodes …")
    snap.node_top = kubectl("top", "nodes", "--no-headers")

    log.info("  pods (all namespaces) …")
    snap.pods_all = kubectl_json("get", "pods", "-A")

    log.info("  kubectl top pods …")
    snap.pod_top = kubectl("top", "pods", "-A", "--no-headers")

    log.info("  deployments …")
    snap.deployments = kubectl_json("get", "deployments", "-A")

    log.info("  statefulsets …")
    snap.statefulsets = kubectl_json("get", "statefulsets", "-A")

    log.info("  daemonsets …")
    snap.daemonsets = kubectl_json("get", "daemonsets", "-A")

    log.info("  HPAs …")
    snap.hpas = kubectl_json("get", "hpa", "-A")

    log.info("  PVCs …")
    snap.pvcs = kubectl_json("get", "pvc", "-A")

    log.info("  services …")
    snap.services = kubectl_json("get", "svc", "-A")

    log.info("  namespaces …")
    raw_ns = kubectl_json("get", "namespaces")
    snap.namespaces = [
        n["metadata"]["name"] for n in raw_ns
        if isinstance(n, dict) and "metadata" in n
    ]

    log.info(
        "Collection complete — %d nodes, %d pods, %d deployments, %d HPAs, %d PVCs.",
        len(snap.nodes), len(snap.pods_all), len(snap.deployments),
        len(snap.hpas), len(snap.pvcs),
    )


# ---------------------------------------------------------------------------
# 3. Karpenter manifest reader
# ---------------------------------------------------------------------------

def collect_karpenter_config(snap: ClusterSnapshot, repo_path: Path) -> None:
    """Parse Karpenter NodePool / EC2NodeClass YAMLs from a local repository."""
    log.info("─── Reading Karpenter manifests from: %s ───────────────────────", repo_path)

    if not repo_path.exists():
        log.warning("Repository path does not exist: %s", repo_path)
        return

    yaml_files = sorted(repo_path.rglob("*.yaml")) + sorted(repo_path.rglob("*.yml"))
    if not yaml_files:
        log.warning("No YAML files found under: %s", repo_path)
        return

    for manifest in yaml_files:
        try:
            docs = list(yaml.safe_load_all(manifest.read_text()))
            for doc in docs:
                if not isinstance(doc, dict):
                    continue
                kind = doc.get("kind", "")
                name = doc.get("metadata", {}).get("name", "<unnamed>")
                if kind == "NodePool":
                    snap.nodepools.append(doc)
                    log.debug("  ✔ NodePool: %s", name)
                elif kind in ("EC2NodeClass", "AWSNodeTemplate"):
                    snap.nodeclasses.append(doc)
                    log.debug("  ✔ NodeClass: %s", name)
        except yaml.YAMLError as exc:
            log.warning("YAML parse error in %s: %s", manifest, exc)

    log.info(
        "Karpenter manifests: %d NodePool(s), %d NodeClass(es).",
        len(snap.nodepools), len(snap.nodeclasses),
    )


# ---------------------------------------------------------------------------
# 4. Heuristic analysis (runs before AI to enrich prompt)
# ---------------------------------------------------------------------------

def _parse_cpu(val: str) -> float:
    """CPU value → millicores."""
    if not val or val == "0":
        return 0.0
    val = str(val)
    return float(val[:-1]) if val.endswith("m") else float(val) * 1000


def _parse_mem(val: str) -> float:
    """Memory value → MiB."""
    if not val or val == "0":
        return 0.0
    val = str(val)
    for suffix, factor in [("Ki", 1 / 1024), ("Mi", 1), ("Gi", 1024),
                            ("Ti", 1024 ** 2), ("k", 1 / 1024),
                            ("M", 1), ("G", 1024)]:
        if val.endswith(suffix):
            return float(val[: -len(suffix)]) * factor
    return float(val) / (1024 * 1024)


def analyse(snap: ClusterSnapshot) -> None:
    """Populate derived heuristic findings in the snapshot."""

    # ── Node summary ──────────────────────────────────────────────────────
    for node in snap.nodes:
        meta   = node.get("metadata", {})
        spec   = node.get("spec", {})
        status = node.get("status", {})
        labels = meta.get("labels", {})
        cap    = status.get("capacity", {})
        alloc  = status.get("allocatable", {})
        snap.node_summary.append({
            "name":           meta.get("name"),
            "instance_type":  labels.get("node.kubernetes.io/instance-type", "unknown"),
            "zone":           labels.get("topology.kubernetes.io/zone", "unknown"),
            "nodepool":       labels.get("karpenter.sh/nodepool", "—"),
            "capacity_cpu":   cap.get("cpu"),
            "capacity_mem":   cap.get("memory"),
            "allocatable_cpu": alloc.get("cpu"),
            "allocatable_mem": alloc.get("memory"),
            "taints":         spec.get("taints", []),
        })

    # ── Idle workloads ────────────────────────────────────────────────────
    for dep in snap.deployments:
        desired = dep.get("spec", {}).get("replicas", 1)
        ready   = dep.get("status", {}).get("readyReplicas", 0)
        if desired == 0 or ready == 0:
            snap.idle_workloads.append({
                "kind":      "Deployment",
                "namespace": dep["metadata"]["namespace"],
                "name":      dep["metadata"]["name"],
                "desired":   desired,
                "ready":     ready,
            })

    for sts in snap.statefulsets:
        desired = sts.get("spec", {}).get("replicas", 1)
        ready   = sts.get("status", {}).get("readyReplicas", 0)
        if desired == 0 or ready == 0:
            snap.idle_workloads.append({
                "kind":      "StatefulSet",
                "namespace": sts["metadata"]["namespace"],
                "name":      sts["metadata"]["name"],
                "desired":   desired,
                "ready":     ready,
            })

    # ── Oversized / missing-limits containers ────────────────────────────
    for dep in snap.deployments:
        ns   = dep["metadata"]["namespace"]
        name = dep["metadata"]["name"]
        containers = (
            dep.get("spec", {})
               .get("template", {})
               .get("spec", {})
               .get("containers", [])
        )
        for c in containers:
            res     = c.get("resources", {})
            reqs    = res.get("requests", {})
            limits  = res.get("limits", {})
            cname   = c.get("name", "")

            if not limits:
                snap.oversized_workloads.append({
                    "kind": "Deployment", "namespace": ns, "name": name,
                    "container": cname, "issue": "No resource limits defined",
                    "requests": reqs, "limits": {},
                })
                continue

            cpu_req = _parse_cpu(reqs.get("cpu", "0"))
            cpu_lim = _parse_cpu(limits.get("cpu", "0"))
            mem_req = _parse_mem(reqs.get("memory", "0"))
            mem_lim = _parse_mem(limits.get("memory", "0"))

            if cpu_lim > 0 and cpu_req > 0 and (cpu_req / cpu_lim) < 0.25:
                snap.oversized_workloads.append({
                    "kind": "Deployment", "namespace": ns, "name": name,
                    "container": cname,
                    "issue": (
                        f"CPU request ({cpu_req:.0f}m) is less than 25% of limit "
                        f"({cpu_lim:.0f}m) — potential over-provisioning"
                    ),
                    "requests": reqs, "limits": limits,
                })
            if mem_lim > 0 and mem_req > 0 and (mem_req / mem_lim) < 0.25:
                snap.oversized_workloads.append({
                    "kind": "Deployment", "namespace": ns, "name": name,
                    "container": cname,
                    "issue": (
                        f"Memory request ({mem_req:.0f} MiB) is less than 25% of "
                        f"limit ({mem_lim:.0f} MiB) — potential over-provisioning"
                    ),
                    "requests": reqs, "limits": limits,
                })

    # ── Unbound PVCs ──────────────────────────────────────────────────────
    for pvc in snap.pvcs:
        if pvc.get("status", {}).get("phase") != "Bound":
            snap.unbound_pvcs.append({
                "namespace":    pvc["metadata"]["namespace"],
                "name":         pvc["metadata"]["name"],
                "phase":        pvc["status"].get("phase", "Unknown"),
                "storage":      pvc["spec"].get("resources", {}).get("requests", {}).get("storage"),
                "storageClass": pvc["spec"].get("storageClassName"),
            })

    log.info(
        "Heuristics: %d idle workloads | %d oversized/no-limit containers | %d unbound PVCs",
        len(snap.idle_workloads), len(snap.oversized_workloads), len(snap.unbound_pvcs),
    )


# ---------------------------------------------------------------------------
# 5. Build GenAI prompt
# ---------------------------------------------------------------------------

def build_prompt(snap: ClusterSnapshot) -> str:
    """Assemble a comprehensive, structured prompt for the GenAI Studio."""

    def _trim(data: list, max_items: int = 40) -> list:
        if len(data) > max_items:
            return data[:max_items] + [
                {"_note": f"{len(data) - max_items} additional items omitted for brevity"}
            ]
        return data

    # Compact deployment view (avoid token explosion)
    deployment_view = _trim([
        {
            "namespace": d["metadata"]["namespace"],
            "name":      d["metadata"]["name"],
            "replicas":  d["spec"].get("replicas"),
            "ready":     d.get("status", {}).get("readyReplicas"),
            "containers": [
                {
                    "name":      c["name"],
                    "image":     c.get("image", ""),
                    "resources": c.get("resources", {}),
                }
                for c in (
                    d.get("spec", {})
                     .get("template", {})
                     .get("spec", {})
                     .get("containers", [])
                )
            ],
        }
        for d in snap.deployments
    ])

    hpa_view = _trim([
        {
            "namespace":       h["metadata"]["namespace"],
            "name":            h["metadata"]["name"],
            "minReplicas":     h["spec"].get("minReplicas"),
            "maxReplicas":     h["spec"].get("maxReplicas"),
            "currentReplicas": h.get("status", {}).get("currentReplicas"),
            "metrics":         h["spec"].get("metrics", []),
        }
        for h in snap.hpas
    ])

    pvc_view = _trim([
        {
            "namespace":    p["metadata"]["namespace"],
            "name":         p["metadata"]["name"],
            "phase":        p["status"].get("phase"),
            "storageClass": p["spec"].get("storageClassName"),
            "capacity":     p["spec"].get("resources", {}).get("requests", {}).get("storage"),
        }
        for p in snap.pvcs
    ])

    data_payload = {
        "cluster_metadata": {
            "name":             snap.cluster_name,
            "region":           snap.region,
            "collected_at_utc": snap.timestamp,
            "node_count":       len(snap.nodes),
            "namespace_count":  len(snap.namespaces),
            "namespaces":       snap.namespaces,
        },
        "node_summary":              snap.node_summary,
        "kubectl_top_nodes":         snap.node_top,
        "kubectl_top_pods_sample":   "\n".join(snap.pod_top.splitlines()[:80]),
        "deployments":               deployment_view,
        "hpas":                      hpa_view,
        "pvcs":                      pvc_view,
        "karpenter_nodepools":       snap.nodepools,
        "karpenter_nodeclasses":     snap.nodeclasses,
        "heuristic_analysis": {
            "idle_workloads":              snap.idle_workloads,
            "oversized_or_missing_limits": snap.oversized_workloads,
            "unbound_pvcs":                snap.unbound_pvcs,
        },
    }

    json_block = json.dumps(data_payload, indent=2, default=str)

    prompt = textwrap.dedent(f"""
    You are a Principal Cloud and Platform Engineer specialising in Kubernetes cost
    optimisation on AWS EKS with deep expertise in Karpenter autoscaling.

    You have been provided with a complete automated snapshot of a live EKS cluster.
    Analyse ALL the data below and produce the cost optimisation report described.

    ═══════════════════════════════════════════════════════════════════════
    REPORT STRUCTURE  (respond ONLY with the report — no preamble)
    ═══════════════════════════════════════════════════════════════════════

    # EKS Cost Optimisation Report — {snap.cluster_name}

    ## 1. Executive Summary
    - Overall cost-health score 1–10 with one-paragraph justification.
    - Top 3 highest-impact actions (bold each one).

    ## 2. Node & Compute Analysis
    - Node count, instance type distribution table, zone spread.
    - Utilisation analysis from `kubectl top nodes` data.
    - Commentary on Spot vs On-Demand ratio and Graviton adoption.

    ## 3. Karpenter Configuration Review
    - Review each NodePool: instance families, weight, limits, disruption policy,
      consolidation settings, expiry / drift.
    - Review each EC2NodeClass: AMI family, subnet/SG selectors, storage config.
    - Provide specific YAML patches (```yaml diff blocks) for any recommended changes.

    ## 4. Workload Right-Sizing
    - Table: namespace | workload | container | issue | recommended action
    - Coverage: missing limits, over-provisioned requests, zero-replica deployments.
    - Mention VPA / KEDA where appropriate.

    ## 5. HPA & Autoscaling Review
    - HPAs with high minReplicas relative to observed demand.
    - Stateless Deployments lacking HPA.
    - Concrete recommendations per workload.

    ## 6. Idle & Zombie Resources
    - Idle workloads (table with recommended `kubectl` cleanup commands).
    - Unbound PVCs (table with estimated wasted storage cost at $0.10/GiB/month).
    - Unused LoadBalancer Services.

    ## 7. Storage Optimisation
    - StorageClass mix (gp2 vs gp3 vs io2) — cost delta if migrated to gp3.
    - Oversized PVC allocations.
    - Migration recommendations.

    ## 8. Quick Wins  (achievable today, < 1 hour each)
    Numbered list — include the exact `kubectl` / `helm` command for each.

    ## 9. Medium-Term Actions  (1–4 weeks)
    Architectural or process improvements with effort estimate (S/M/L).

    ## 10. Estimated Savings Summary
    | Category | Current Est. Waste | Potential Saving | Confidence |
    |---|---|---|---|
    | ... | ... | ... | Low/Medium/High |

    ═══════════════════════════════════════════════════════════════════════
    CLUSTER SNAPSHOT DATA
    ═══════════════════════════════════════════════════════════════════════

    ```json
    {json_block}
    ```

    ───────────────────────────────────────────────────────────────────────
    Rules: Reference actual namespace/workload names from the data.
    Use Markdown tables and fenced code blocks. Be specific and actionable.
    Do NOT add any text outside the report structure above.
    ───────────────────────────────────────────────────────────────────────
    """).strip()

    return prompt


# ---------------------------------------------------------------------------
# 6. GenAI Studio API call
# ---------------------------------------------------------------------------

def call_genai_studio(
    prompt: str,
    api_url: str,
    api_token: str,
    model: str,
    timeout: int = 180,
) -> str:
    """POST the prompt to the corporate GenAI Studio and return the AI response."""
    log.info("─── Calling GenAI Studio ────────────────────────────────────────")
    log.info("  URL   : %s", api_url)
    log.info("  Model : %s", model)

    headers = {
        "Authorization": f"Bearer {api_token}",
        "Content-Type":  "application/json",
    }
    body = {
        "model": model,
        "messages": [
            {
                "role": "system",
                "content": (
                    "You are a Principal Platform Engineer specialising in Kubernetes "
                    "and AWS EKS cost optimisation. Provide detailed, actionable, "
                    "technically precise reports in clean Markdown."
                ),
            },
            {"role": "user", "content": prompt},
        ],
        "temperature": 0.15,
        "max_tokens":  8192,
    }

    try:
        resp = requests.post(api_url, headers=headers, json=body, timeout=timeout)
        resp.raise_for_status()
        data = resp.json()

        # Support OpenAI-compatible, Anthropic, and plain text response schemas
        if "choices" in data:
            return data["choices"][0]["message"]["content"]
        if "content" in data:
            # Anthropic-style: content is a list of blocks
            if isinstance(data["content"], list):
                return "\n".join(
                    block.get("text", "") for block in data["content"]
                    if block.get("type") == "text"
                )
            return str(data["content"])
        if "response" in data:
            return str(data["response"])

        log.warning("Unexpected response schema — dumping raw JSON.")
        return json.dumps(data, indent=2)

    except requests.exceptions.HTTPError as exc:
        log.error("HTTP error from GenAI Studio: %s\n%s", exc, exc.response.text[:500])
        return f"**ERROR: GenAI Studio HTTP error — {exc}**"
    except requests.exceptions.RequestException as exc:
        log.error("Network error calling GenAI Studio: %s", exc)
        return f"**ERROR: GenAI Studio request failed — {exc}**"


# ---------------------------------------------------------------------------
# 7. Report rendering
# ---------------------------------------------------------------------------

REPORT_CSS = """
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    max-width: 1200px; margin: 48px auto; padding: 0 24px;
    background: #f8fafc; color: #1e293b; line-height: 1.7;
  }
  h1 { color: #0f172a; border-bottom: 3px solid #3b82f6; padding-bottom: 10px; font-size: 2rem; }
  h2 { color: #1d4ed8; margin-top: 2.5em; font-size: 1.4rem; }
  h3 { color: #374151; }
  a  { color: #2563eb; }
  table { border-collapse: collapse; width: 100%; margin: 1.2em 0; font-size: .92em; }
  th { background: #1d4ed8; color: #fff; padding: 9px 14px; text-align: left; }
  td { border: 1px solid #cbd5e1; padding: 8px 14px; vertical-align: top; }
  tr:nth-child(even) td { background: #eff6ff; }
  code { background: #e2e8f0; border-radius: 4px; padding: 2px 6px; font-size: .88em; }
  pre  { background: #0f172a; color: #e2e8f0; padding: 20px; border-radius: 10px;
         overflow-x: auto; font-size: .84em; line-height: 1.5; }
  pre code { background: none; padding: 0; }
  blockquote { border-left: 4px solid #3b82f6; margin: 1em 0; padding: 6px 18px;
               background: #eff6ff; border-radius: 0 8px 8px 0; }
  hr { border: none; border-top: 1px solid #cbd5e1; margin: 2em 0; }
  .meta-bar {
    display: flex; flex-wrap: wrap; gap: 20px;
    background: #0f172a; color: #94a3b8;
    padding: 14px 22px; border-radius: 10px;
    font-size: .84em; margin-bottom: 2.5em;
  }
  .meta-bar span b { color: #f1f5f9; }
  footer { color: #94a3b8; font-size: .78em; text-align: center; margin-top: 3em; }
"""


def _md_to_html(md: str) -> str:
    """Convert Markdown to HTML — uses 'markdown' package if available."""
    try:
        import markdown  # type: ignore
        return markdown.markdown(
            md, extensions=["tables", "fenced_code", "toc", "nl2br"]
        )
    except ImportError:
        pass

    # Lightweight fallback
    import re
    html = md
    html = re.sub(
        r"```(\w*)\n(.*?)```",
        lambda m: f"<pre><code class='language-{m.group(1)}'>{m.group(2)}</code></pre>",
        html, flags=re.DOTALL,
    )
    for lvl in range(6, 0, -1):
        html = re.sub(
            rf"^{'#' * lvl}\s+(.+)$", rf"<h{lvl}>\1</h{lvl}>",
            html, flags=re.MULTILINE,
        )
    html = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", html)
    html = re.sub(r"\*(.+?)\*",     r"<em>\1</em>",          html)
    html = re.sub(r"`([^`]+)`",     r"<code>\1</code>",       html)
    html = re.sub(r"^---+$", "<hr>", html, flags=re.MULTILINE)
    paras = re.split(r"\n{2,}", html)
    html = "\n".join(
        p if p.strip().startswith("<") else f"<p>{p.strip()}</p>"
        for p in paras if p.strip()
    )
    return html


def write_reports(
    snap: ClusterSnapshot,
    ai_report: str,
    output_dir: Path,
) -> tuple[Path, Path, Path]:
    """Write Markdown, HTML, and raw snapshot JSON to output_dir."""
    output_dir.mkdir(parents=True, exist_ok=True)
    ts        = datetime.utcnow().strftime("%Y%m%d-%H%M%S")
    base_name = f"cost-report_{snap.cluster_name}_{ts}"

    # ── Markdown ──────────────────────────────────────────────────────────
    md_header = textwrap.dedent(f"""
    # EKS Cost Optimisation Report — {snap.cluster_name}

    | Field | Value |
    |---|---|
    | **Cluster** | `{snap.cluster_name}` |
    | **Region** | `{snap.region}` |
    | **Generated (UTC)** | {snap.timestamp} |
    | **Nodes** | {len(snap.nodes)} |
    | **Namespaces** | {len(snap.namespaces)} |
    | **Deployments** | {len(snap.deployments)} |
    | **HPAs** | {len(snap.hpas)} |
    | **PVCs** | {len(snap.pvcs)} |
    | **Karpenter NodePools** | {len(snap.nodepools)} |
    | **Karpenter NodeClasses** | {len(snap.nodeclasses)} |
    | **Idle workloads** | {len(snap.idle_workloads)} |
    | **Oversized/no-limit containers** | {len(snap.oversized_workloads)} |
    | **Unbound PVCs** | {len(snap.unbound_pvcs)} |

    ---

    """).lstrip()

    md_path = output_dir / f"{base_name}.md"
    md_path.write_text(md_header + ai_report, encoding="utf-8")
    log.info("Markdown report : %s", md_path)

    # ── HTML ──────────────────────────────────────────────────────────────
    meta_bar = textwrap.dedent(f"""
    <div class="meta-bar">
      <span>🗂 Cluster: <b>{snap.cluster_name}</b></span>
      <span>🌍 Region: <b>{snap.region}</b></span>
      <span>🕒 Generated: <b>{snap.timestamp} UTC</b></span>
      <span>🖥 Nodes: <b>{len(snap.nodes)}</b></span>
      <span>📦 Namespaces: <b>{len(snap.namespaces)}</b></span>
      <span>🚀 Deployments: <b>{len(snap.deployments)}</b></span>
      <span>⚡ HPAs: <b>{len(snap.hpas)}</b></span>
      <span>⚠️ Idle workloads: <b>{len(snap.idle_workloads)}</b></span>
      <span>🔴 Oversized containers: <b>{len(snap.oversized_workloads)}</b></span>
      <span>💾 Unbound PVCs: <b>{len(snap.unbound_pvcs)}</b></span>
    </div>
    """)

    body_html = _md_to_html(ai_report)
    html_content = textwrap.dedent(f"""<!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>EKS Cost Optimisation — {snap.cluster_name}</title>
      <style>{REPORT_CSS}</style>
    </head>
    <body>
      <h1>📊 EKS Cost Optimisation Report</h1>
      {meta_bar}
      {body_html}
      <hr>
      <footer>
        Generated by <strong>eks-cost-optimizer</strong> &bull;
        Powered by Corporate GenAI Studio &bull;
        {snap.timestamp} UTC
      </footer>
    </body>
    </html>""")

    html_path = output_dir / f"{base_name}.html"
    html_path.write_text(html_content, encoding="utf-8")
    log.info("HTML report     : %s", html_path)

    # ── Raw snapshot JSON (audit trail) ──────────────────────────────────
    json_path = output_dir / f"{base_name}-snapshot.json"
    json_path.write_text(
        json.dumps(
            {
                "cluster":            snap.cluster_name,
                "region":             snap.region,
                "timestamp":          snap.timestamp,
                "node_summary":       snap.node_summary,
                "idle_workloads":     snap.idle_workloads,
                "oversized_workloads": snap.oversized_workloads,
                "unbound_pvcs":       snap.unbound_pvcs,
                "karpenter_nodepools": [
                    np.get("metadata", {}).get("name") for np in snap.nodepools
                ],
                "karpenter_nodeclasses": [
                    nc.get("metadata", {}).get("name") for nc in snap.nodeclasses
                ],
            },
            indent=2,
            default=str,
        ),
        encoding="utf-8",
    )
    log.info("Snapshot JSON   : %s", json_path)

    return md_path, html_path, json_path


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(
        description="EKS Cost Optimisation Report Generator",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    p.add_argument(
        "--cluster", required=True,
        help="EKS cluster name",
    )
    p.add_argument(
        "--region", default="eu-west-1",
        help="AWS region (default: eu-west-1)",
    )
    p.add_argument(
        "--profile", default=None,
        help="AWS CLI profile name (optional)",
    )
    p.add_argument(
        "--repo", required=True,
        help="Local path to Karpenter NodePool / NodeClass YAML manifests",
    )
    p.add_argument(
        "--ai-url", required=True,
        help="Corporate GenAI Studio completions endpoint URL",
    )
    p.add_argument(
        "--ai-token", default=None,
        help="Bearer token for the GenAI Studio API (or set GENAI_API_TOKEN env var)",
    )
    p.add_argument(
        "--ai-model", default="gpt-4o",
        help="Model identifier to use (default: gpt-4o)",
    )
    p.add_argument(
        "--output", default="./cost-reports",
        help="Output directory for reports (default: ./cost-reports)",
    )
    p.add_argument(
        "--skip-context-switch", action="store_true",
        help="Skip `aws eks update-kubeconfig` (useful if context is already set)",
    )
    p.add_argument(
        "--dry-run", action="store_true",
        help="Collect data and write prompt to disk; skip the AI API call",
    )
    p.add_argument(
        "--debug", action="store_true",
        help="Enable DEBUG-level logging",
    )
    return p.parse_args()


def main() -> None:
    args = parse_args()

    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)

    # Resolve API token
    api_token = args.ai_token or os.environ.get("GENAI_API_TOKEN", "")
    if not api_token and not args.dry_run:
        log.error(
            "No API token supplied. Use --ai-token <TOKEN> or "
            "export GENAI_API_TOKEN=<TOKEN>."
        )
        sys.exit(1)

    log.info("═══════════════════════════════════════════════════════════")
    log.info(" EKS Cost Optimisation Report Generator")
    log.info(" Cluster : %s  |  Region : %s", args.cluster, args.region)
    log.info("═══════════════════════════════════════════════════════════")

    # Step 1 — Switch kubectl context
    if not args.skip_context_switch:
        switch_context(args.cluster, args.region, args.profile)

    # Step 2 — Initialise snapshot
    snap = ClusterSnapshot(cluster_name=args.cluster, region=args.region)

    # Step 3 — Collect live cluster data
    collect_cluster_data(snap)

    # Step 4 — Read Karpenter manifests from repository
    collect_karpenter_config(snap, Path(args.repo))

    # Step 5 — Local heuristic analysis
    log.info("─── Running heuristic analysis ─────────────────────────────────")
    analyse(snap)

    # Step 6 — Build AI prompt
    log.info("─── Building AI prompt ──────────────────────────────────────────")
    prompt = build_prompt(snap)
    log.info("Prompt size: %d characters / ~%d tokens", len(prompt), len(prompt) // 4)

    if args.dry_run:
        out = Path(args.output)
        out.mkdir(parents=True, exist_ok=True)
        prompt_file = out / f"prompt-dry-run_{args.cluster}.txt"
        prompt_file.write_text(prompt, encoding="utf-8")
        log.info("[DRY RUN] Prompt written to %s — AI call skipped.", prompt_file)
        return

    # Step 7 — Call GenAI Studio
    ai_report = call_genai_studio(
        prompt=prompt,
        api_url=args.ai_url,
        api_token=api_token,
        model=args.ai_model,
    )

    # Step 8 — Write reports
    log.info("─── Writing reports ─────────────────────────────────────────────")
    md_path, html_path, json_path = write_reports(
        snap, ai_report, Path(args.output)
    )

    print()
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║  ✅  Cost Optimisation Report Generated Successfully          ║")
    print("╠══════════════════════════════════════════════════════════════╣")
    print(f"║  📄  Markdown  →  {str(md_path):<44} ║")
    print(f"║  🌐  HTML      →  {str(html_path):<44} ║")
    print(f"║  🗄  Snapshot  →  {str(json_path):<44} ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print()


if __name__ == "__main__":
    main()
