To deploy a self-managed container registry on a Minikube Kubernetes cluster, you can use Minikube’s built-in registry add-on or set up a custom Docker registry. Below are the steps for both approaches, with the add-on method being the simplest. I'll focus on the add-on method first, as it’s officially supported by Minikube, and then briefly cover a custom setup for more control.

### Method 1: Using Minikube’s Registry Add-on

Minikube provides a registry add-on that deploys a Docker registry inside the cluster, ideal for local development and testing.

1. **Start Minikube with Insecure Registry Support**:
   Ensure Minikube is configured to allow pushing/pulling images to an insecure registry (since the default registry add-on doesn’t use TLS).
   ```bash
   minikube start --insecure-registry="10.0.0.0/24"
   ```
   - The `--insecure-registry` flag allows the Minikube kubelet to communicate with the registry without TLS. The `10.0.0.0/24` range covers the cluster’s service IPs. If you already have a running Minikube cluster, delete it first with `minikube delete` and then restart.[](https://minikube.sigs.k8s.io/docs/handbook/registry/)

2. **Enable the Registry Add-on**:
   Enable the built-in registry add-on, which deploys a Docker registry in the `kube-system` namespace.
   ```bash
   minikube addons enable registry
   ```
   - This creates a registry service listening on port 5000 within the cluster. You can verify it’s running:
     ```bash
     kubectl -n kube-system get svc registry
     ```
     Output will show the `CLUSTER-IP` and port (typically `10.x.x.x:80`).[](https://developers.redhat.com/blog/2019/07/11/deploying-an-internal-container-registry-with-minikube-add-ons)[](https://minikube.sigs.k8s.io/docs/handbook/registry/)

3. **Verify the Registry**:
   Check the registry pod’s status:
   ```bash
   kubectl -n kube-system get pods | grep registry
   ```
   Ensure the registry pod (e.g., `registry-xxx`) is in the `Running` state.

4. **Access the Registry from Your Local Machine**:
   To push/pull images from your host machine, set up port forwarding to access the registry:
   ```bash
   kubectl port-forward --namespace kube-system service/registry 5000:80
   ```
   - This maps `localhost:5000` on your host to the registry service’s port 80. Keep this terminal running for port forwarding.[](https://minikube.sigs.k8s.io/docs/handbook/registry/)

5. **Test Pushing an Image**:
   - Build a sample Docker image:
     ```bash
     docker build -t localhost:5000/my-image:latest .
     ```
   - Push the image to the registry:
     ```bash
     docker push localhost:5000/my-image:latest
     ```
   - Verify the image is in the registry:
     ```bash
     curl http://localhost:5000/v2/_catalog
     ```
     You should see `{"repositories":["my-image"]}`.[](https://minikube.sigs.k8s.io/docs/handbook/registry/)

6. **Use the Registry in Kubernetes**:
   - Deploy a pod using the image from the registry. Create a file `test-pod.yaml`:
     ```yaml
     apiVersion: v1
     kind: Pod
     metadata:
       name: test-pod
     spec:
       containers:
       - name: test-container
         image: registry.kube-system.svc.cluster.local:80/my-image:latest
     ```
   - Apply it:
     ```bash
     kubectl apply -f test-pod.yaml
     ```
   - The pod should pull the image from the internal registry. Use `kubectl get pods` to check its status.[](https://developers.redhat.com/blog/2019/07/11/deploying-an-internal-container-registry-with-minikube-add-ons)

7. **Optional: Update `/etc/hosts` for Aliases**:
   If you want to use a custom domain (e.g., `registry.dev`) for the registry, update the Minikube VM’s `/etc/hosts`:
   ```bash
   REGISTRY_IP=$(kubectl -n kube-system get svc registry -o jsonpath='{.spec.clusterIP}')
   minikube ssh "echo \"$REGISTRY_IP registry.dev\" | sudo tee -a /etc/hosts"
   ```
   This allows pods to reference the registry as `registry.dev:80`.[](https://developers.redhat.com/blog/2019/07/11/deploying-an-internal-container-registry-with-minikube-add-ons)

### Method 2: Deploy a Custom Docker Registry

If you need more control (e.g., persistent storage or custom configuration), deploy a Docker registry manually.

1. **Create a Registry Deployment**:
   Create a file `registry-deployment.yaml`:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: registry
     namespace: kube-system
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: registry
     template:
       metadata:
         labels:
           app: registry
       spec:
         containers:
         - name: registry
           image: registry:2
           ports:
           - containerPort: 5000
           volumeMounts:
           - name: registry-storage
             mountPath: /var/lib/registry
         volumes:
         - name: registry-storage
           emptyDir: {} # Use persistentVolumeClaim for production
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: registry
     namespace: kube-system
   spec:
     selector:
       app: registry
     ports:
     - protocol: TCP
       port: 80
       targetPort: 5000
   ```
   Apply it:
   ```bash
   kubectl apply -f registry-deployment.yaml
   ```

2. **Configure Insecure Registry**:
   Ensure Minikube is started with `--insecure-registry="10.0.0.0/24"` as in Method 1.[](https://minikube.sigs.k8s.io/docs/handbook/registry/)

3. **Port Forwarding and Testing**:
   Follow steps 4–6 from Method 1 to set up port forwarding, push images, and use the registry in pods.

### Notes
- **Insecure Registry**: The `--insecure-registry` flag is necessary for local development without TLS. For production, configure TLS certificates for security.[](https://minikube.sigs.k8s.io/docs/handbook/registry/)[](https://hasura.io/blog/sharing-a-local-registry-for-minikube-37c7240d0615)
- **Storage**: The add-on uses ephemeral storage by default. For persistence, configure a `PersistentVolumeClaim` in the custom deployment.[](https://developers.redhat.com/blog/2019/07/11/deploying-an-internal-container-registry-with-minikube-add-ons)
- **Accessing from Host**: If `localhost:5000` fails, ensure your Docker daemon trusts the insecure registry. On your host, add `"insecure-registries": ["localhost:5000"]` to `/etc/docker/daemon.json` and restart Docker (`sudo systemctl restart docker`).[](https://gist.github.com/trisberg/37c97b6cc53def9a3e38be6143786589)
- **CI/CD**: For CI/CD pipelines, ensure pods can resolve the registry using the cluster IP or a custom domain like `registry.dev`. Update CoreDNS or `/etc/hosts` as needed.[](https://developers.redhat.com/blog/2019/07/11/deploying-an-internal-container-registry-with-minikube-add-ons)

### Troubleshooting
- **Pod Fails to Pull Image**: Check if the registry service is running (`kubectl -n kube-system get svc registry`) and verify the image name matches exactly.
- **Connection Refused**: Ensure port forwarding is active or the registry’s cluster IP is reachable. Check `no_proxy` settings if behind a proxy.[](https://gist.github.com/trisberg/37c97b6cc53def9a3e38be6143786589)
- **ImagePullBackOff**: Confirm the `--insecure-registry` flag is set and the registry is accessible from the Minikube VM (`minikube ssh` and `curl registry.kube-system.svc.cluster.local:80/v2/_catalog`).

This setup should get you a functional self-managed registry on Minikube. If you need further customization (e.g., authentication, TLS), let me know!