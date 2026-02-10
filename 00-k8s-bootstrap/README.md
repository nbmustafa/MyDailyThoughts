Kubernetes sandbox bootstrap with all the features you requested. Here's what you have:
🎯 Complete Package Delivered
Core Scripts (All Executable)

bootstrap.sh - Main bootstrap script with Helm-based deployment
bootstrap-gitops.sh - Advanced version with local Git repository for true GitOps
cleanup.sh - Safe cluster teardown
validate.sh - Comprehensive health checks and smoke tests
Makefile - Convenient commands (make bootstrap, make status, etc.)

Documentation

00-START-HERE.md - Quick orientation guide
README.md - Complete documentation (9.7KB)
QUICKSTART.md - 5-minute setup guide
ARCHITECTURE.md - Deep technical architecture (14KB)
PROJECT_STRUCTURE.md - File organization and usage patterns

Example Configurations

custom-alerts.yaml - Prometheus alerting rules examples
sample-argocd-app.yaml - ArgoCD application templates
grafana-dashboards.md - Dashboard configuration guide

🚀 Key Features
✅ Single-script deployment - Everything in one command
✅ Local Helm charts - Downloaded and stored locally (argocd 7.7.12, prometheus-stack 67.6.0)
✅ GitOps workflow - ArgoCD manages all deployments
✅ Complete monitoring - Prometheus + Grafana + Alertmanager
✅ Localhost access - All services exposed on localhost ports
✅ Self-contained - No external dependencies during runtime
✅ Production patterns - Enterprise platform engineering best practices
📦 What Gets Deployed

Kind Cluster with custom port mappings
ArgoCD (GitOps engine) → http://localhost:8080
Prometheus (metrics) → http://localhost:9090
Grafana (dashboards) → http://localhost:3000
Alertmanager (alerts) → http://localhost:9093
Node Exporter + Kube State Metrics

🎓 Usage
bash# Quick start
chmod +x bootstrap.sh
./bootstrap.sh

# Or use GitOps version
./bootstrap-gitops.sh

# Or use Make
make bootstrap

# Validate
./validate.sh

# Cleanup
./cleanup.sh
💡 Two Versions Available

Standard (bootstrap.sh) - Direct Helm deployment, fastest setup
GitOps (bootstrap-gitops.sh) - Creates local Git repo, true GitOps workflow

Both versions include the same monitoring stack, but the GitOps version demonstrates production GitOps patterns with a local repository that ArgoCD watches.
The entire package is ready to download and use! Start with 00-START-HERE.md for quick orientation, or dive into QUICKSTART.md to get running in 5 minutes.
