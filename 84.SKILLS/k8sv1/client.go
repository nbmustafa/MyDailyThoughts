package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Clients holds all Kubernetes API clients
type Clients struct {
	Kubernetes kubernetes.Interface
	Metrics    metricsv1beta1.Interface
	RestConfig *rest.Config
}

// BuildClients creates K8s API clients for the given context
func BuildClients(kubeconfig, context string) (*Clients, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			defaultPath := filepath.Join(home, ".kube", "config")
			if _, err := os.Stat(defaultPath); err == nil {
				loadingRules.ExplicitPath = defaultPath
			}
		}
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// Increase QPS and burst for parallel diagnostics
	restConfig.QPS = 50
	restConfig.Burst = 100

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(restConfig)
	if err != nil {
		// Metrics server is optional — don't fail hard
		metricsClient = nil
	}

	return &Clients{
		Kubernetes: k8sClient,
		Metrics:    metricsClient,
		RestConfig: restConfig,
	}, nil
}

// ResolveContext finds the kubeconfig context for a given cluster name
func ResolveContext(kubeconfig, clusterName string) (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	rawConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).RawConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Try exact context name match first
	if _, ok := rawConfig.Contexts[clusterName]; ok {
		return clusterName, nil
	}

	// Try matching cluster name within context definitions
	for ctxName, ctx := range rawConfig.Contexts {
		if ctx.Cluster == clusterName {
			return ctxName, nil
		}
	}

	// Try partial match on context name
	for ctxName := range rawConfig.Contexts {
		if containsCI(ctxName, clusterName) {
			return ctxName, nil
		}
	}

	return "", fmt.Errorf("no kubeconfig context found for cluster %q — available contexts: %v",
		clusterName, contextNames(rawConfig.Contexts))
}

func containsCI(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			(func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if equalFold(s[i:i+len(substr)], substr) {
						return true
					}
				}
				return false
			})())
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func contextNames(m map[string]*clientcmd.NamedContext) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	return names
}
