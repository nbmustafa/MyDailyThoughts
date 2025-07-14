Deploying Argo CD on Minikube on a Corporate MacBook Behind a VPN
Deploying Argo CD on Minikube within a corporate environment on a MacBook that's behind a VPN requires a few extra steps to navigate network restrictions. This guide will walk you through the process, from configuring your local environment to deploying a sample application with Argo CD.
1. Configure Your Local Environment
Before starting Minikube, you'll need to configure Docker Desktop and your terminal environment to work with your corporate proxy.
Docker Desktop Proxy Settings
Your corporate network likely uses a proxy to manage internet access. You'll need to configure Docker Desktop to use this proxy.
 * Open Docker Desktop and go to Preferences > Resources > Proxies.
 * Select Manual proxy configuration and enter the HTTP and HTTPS proxy settings provided by your IT department.
 * Add the following to the "Bypass proxy settings for these hosts & domains" section to ensure Minikube can communicate with its internal components:
   localhost,127.0.0.1,kubernetes.docker.internal,10.96.0.0/12,192.168.0.0/16,192.168.49.0/24,192.168.59.0/24,192.168.99.0/24

 * Apply and restart Docker Desktop.
Terminal Proxy Configuration
Set the following environment variables in your terminal session. It's recommended to add these to your shell profile (~/.zshrc, ~/.bashrc, etc.) for persistence.
export HTTP_PROXY="http://<your-proxy-url>:<port>"
export HTTPS_PROXY="http://<your-proxy-url>:<port>"
export NO_PROXY="localhost,127.0.0.1,kubernetes.docker.internal,10.96.0.0/12,192.168.0.0/16,192.168.49.0/24,192.168.59.0/24,192.168.99.0/24"

Replace <your-proxy-url>:<port> with your actual proxy server address and port.
2. Start Minikube with VPN Considerations
When connected to a corporate VPN, it's crucial to ensure that traffic to Minikube's internal network is not routed through the VPN.
 * Check VPN Settings: Some VPN clients have an option to "Allow local (LAN) access" or a similar feature. If available, enable it.
 * Start Minikube: Start Minikube with the appropriate driver (hyperkit is common on macOS). Pass the proxy environment variables to the Minikube VM during startup:
   minikube start --driver=hyperkit \
--docker-env HTTP_PROXY=$HTTP_PROXY \
--docker-env HTTPS_PROXY=$HTTPS_PROXY \
--docker-env NO_PROXY=$NO_PROXY

3. Install and Access Argo CD
With Minikube running, you can now deploy Argo CD.
 * Create the Argo CD Namespace:
   kubectl create namespace argocd

 * Apply the Argo CD Manifests:
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

 * Access the Argo CD UI: To access the Argo CD web interface, you'll need to forward the argocd-server service port to your local machine:
   kubectl port-forward svc/argocd-server -n argocd 8080:443

   You can now access the Argo CD UI at https://localhost:8080.
 * Get the Initial Admin Password: The initial password for the admin user is stored in a Kubernetes secret. Retrieve it with the following command:
   kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

   Use this password to log in to the UI.
4. Configure Argo CD for Your Corporate Git Repository
If your application source code is in a corporate Git repository that is also behind a firewall, you'll need to configure Argo CD to use your proxy to access it.
 * Edit the Argo CD Repo Server Deployment:
   kubectl edit deployment argocd-repo-server -n argocd

 * Add Proxy Environment Variables: In the spec.template.spec.containers section for the argocd-repo-server container, add the following environment variables:
   env:
- name: HTTP_PROXY
  value: "http://<your-proxy-url>:<port>"
- name: HTTPS_PROXY
  value: "http://<your-proxy-url>:<port>"
- name: NO_PROXY
  value: "kubernetes.default.svc,10.96.0.0/12,192.168.0.0/16"

   Save and exit the editor. The argocd-repo-server pod will restart with the new environment variables.
You can now connect to your corporate Git repository through the Argo CD UI or CLI and start deploying your applications.
This video provides a great visual guide to installing Argo CD on a local Kubernetes cluster, which can complement the steps outlined above.
 * Installing ArgoCD on a local Kubernetes Cluster | Minikube
