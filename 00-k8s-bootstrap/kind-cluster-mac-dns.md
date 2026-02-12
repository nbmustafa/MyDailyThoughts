Fix DNS resolution issues in your Kind cluster on macOS. This is a common problem where Kind containers can't use the host's DNS settings properly.

## Quick Solutions

### 1. **Create Kind Cluster with Custom DNS Config**

Create a Kind configuration file that uses your Mac's DNS servers:

```yaml
# kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
networking:
  # This helps with DNS resolution
  disableDefaultCNI: false
```

Then create the cluster:
```bash
kind create cluster --config kind-config.yaml
```

### 2. **Use Docker Desktop's DNS Settings**

Configure Docker Desktop to use your Mac's DNS:

1. Open Docker Desktop → Settings → Resources → Network
2. Note the DNS servers or set them manually to match your Mac's DNS
3. Common options:
   - `8.8.8.8, 8.8.4.4` (Google DNS)
   - Your corporate DNS if on VPN
   - Or find your Mac's DNS with: `scutil --dns | grep 'nameserver\[[0-9]*\]'`

### 3. **Patch CoreDNS in Kind Cluster**

If the cluster is already created, update CoreDNS to use specific nameservers:

```bash
kubectl -n kube-system edit configmap coredns
```

Add/modify the `forward` section:
```
.:53 {
    errors
    health
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      pods insecure
      fallthrough in-addr.arpa ip6.arpa
    }
    prometheus :9153
    forward . 8.8.8.8 8.8.4.4  # Add your DNS servers here
    cache 30
    loop
    reload
    loadbalance
}
```

Then restart CoreDNS:
```bash
kubectl -n kube-system rollout restart deployment coredns
```

### 4. **Check Docker DNS Settings**

Verify Docker is using correct DNS:
```bash
docker run --rm alpine nslookup docker.io
```

If this fails, edit `~/.docker/daemon.json`:
```json
{
  "dns": ["8.8.8.8", "8.8.4.4"]
}
```

Then restart Docker Desktop.

### 5. **Test DNS Resolution**

After applying fixes, test inside a Kind node:
```bash
docker exec -it kind-control-plane nslookup docker.io
```

Or create a test pod:
```bash
kubectl run test --image=busybox --rm -it --restart=Never -- nslookup docker.io
```

## Most Common Fix for Mac

The most reliable approach is usually **option 3** (patching CoreDNS) combined with ensuring Docker Desktop has proper DNS configured.

Would you like me to help you troubleshoot further or create a more specific configuration based on your setup?