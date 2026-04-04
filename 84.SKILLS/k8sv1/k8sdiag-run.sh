#!/usr/bin/env bash
# k8sdiag-run.sh — Wrapper called by the k8s-diagnostics Claude SKILL
#
# Usage:
#   ./k8sdiag-run.sh --cluster <name> [--namespace <ns>] [--output text|json|markdown]
#
# This script:
#   1. Locates the k8sdiag binary
#   2. Validates required arguments
#   3. Executes the diagnostic with appropriate flags
#   4. Exits with the tool's exit code for Claude to interpret

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────────
CLUSTER=""
NAMESPACE=""
OUTPUT="text"
TIMEOUT="90s"
NO_COLOR=""
SKIP=""
KUBECONFIG_PATH=""
KUBE_CONTEXT=""

# ── Parse args ────────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --cluster|-c)     CLUSTER="$2";        shift 2 ;;
    --namespace|-n)   NAMESPACE="$2";      shift 2 ;;
    --output|-o)      OUTPUT="$2";         shift 2 ;;
    --timeout)        TIMEOUT="$2";        shift 2 ;;
    --no-color)       NO_COLOR="--no-color"; shift ;;
    --skip)           SKIP="$2";           shift 2 ;;
    --kubeconfig)     KUBECONFIG_PATH="$2"; shift 2 ;;
    --context)        KUBE_CONTEXT="$2";   shift 2 ;;
    --help|-h)
      echo "Usage: $0 --cluster <n> [--namespace <ns>] [--output text|json|markdown]"
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

# ── Validate ──────────────────────────────────────────────────────────────────
if [[ -z "$CLUSTER" ]]; then
  echo "ERROR: --cluster is required" >&2
  exit 1
fi

# ── Locate binary ─────────────────────────────────────────────────────────────
BINARY=""
SEARCH_PATHS=(
  "/usr/local/bin/k8sdiag"
  "$HOME/.local/bin/k8sdiag"
  "$(dirname "$0")/../bin/k8sdiag"
  "$(dirname "$0")/../../k8s-diag-tool/bin/k8sdiag"
  "./bin/k8sdiag"
  "./k8sdiag"
)

for p in "${SEARCH_PATHS[@]}"; do
  if [[ -x "$p" ]]; then
    BINARY="$p"
    break
  fi
done

if [[ -z "$BINARY" ]]; then
  echo "ERROR: k8sdiag binary not found. Please run:" >&2
  echo "  cd k8s-diag-tool && make install" >&2
  exit 127
fi

# ── Build command ─────────────────────────────────────────────────────────────
CMD=("$BINARY" "--cluster" "$CLUSTER" "--output" "$OUTPUT" "--timeout" "$TIMEOUT")

[[ -n "$NAMESPACE" ]]       && CMD+=("--namespace" "$NAMESPACE")
[[ -n "$NO_COLOR" ]]        && CMD+=("--no-color")
[[ -n "$SKIP" ]]            && CMD+=("--skip" "$SKIP")
[[ -n "$KUBECONFIG_PATH" ]] && CMD+=("--kubeconfig" "$KUBECONFIG_PATH")
[[ -n "$KUBE_CONTEXT" ]]    && CMD+=("--context" "$KUBE_CONTEXT")

# ── Execute ───────────────────────────────────────────────────────────────────
"${CMD[@]}"
exit $?
