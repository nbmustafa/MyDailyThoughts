package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/your-org/k8sdiag/internal/checker"
	"github.com/your-org/k8sdiag/internal/config"
	"github.com/your-org/k8sdiag/internal/reporter"
	"github.com/your-org/k8sdiag/internal/types"
)

// Checker is the common interface all diagnostic checkers implement
type Checker interface {
	Category() types.Category
	Run(ctx context.Context) types.CheckResult
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var (
	clusterName    string
	namespace      string
	kubeconfig     string
	kubeContext    string
	outputFormat   string
	timeout        time.Duration
	noColor        bool
	verbose        bool
	skipCategories []string
)

var rootCmd = &cobra.Command{
	Use:   "k8sdiag",
	Short: "Comprehensive Kubernetes cluster diagnostic tool",
	Long: `k8sdiag — Kubernetes Diagnostic Tool

Performs comprehensive health checks across your Kubernetes cluster covering:
  • Pods (crash loops, OOM, image issues, probe misconfigurations)
  • Nodes (NotReady, pressure conditions, capacity)
  • Storage (PV/PVC, StorageClass, volume mounts)
  • Networking (Services, Endpoints, NetworkPolicies, Calico)
  • Scheduling (affinity, taints, topology spread)
  • Events (namespace-level warning event analysis)
  • Resources (quota exhaustion, HPA health)
  • RBAC (excessive permissions, wildcard roles)
  • Config (missing Secrets, orphaned ConfigMaps)
  • Ingress (missing backends, IngressClass)
  • DNS (CoreDNS health)
  • Namespace (terminating namespaces)

Examples:
  # Cluster-wide diagnostics
  k8sdiag --cluster my-prod-cluster

  # Namespace-scoped diagnostics
  k8sdiag --cluster my-prod-cluster --namespace payments

  # Output as JSON for piping
  k8sdiag --cluster my-prod-cluster --output json

  # Output as Markdown (for reports)
  k8sdiag --cluster my-prod-cluster --output markdown > report.md
`,
	RunE: runDiagnostics,
}

func init() {
	rootCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "Cluster name or kubeconfig context (required)")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Scope diagnostics to a specific namespace (default: all namespaces)")
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	rootCmd.Flags().StringVar(&kubeContext, "context", "", "Override kubeconfig context directly")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text|json|markdown")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "Timeout for all checks combined")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show additional detail including OK results")
	rootCmd.Flags().StringSliceVar(&skipCategories, "skip", nil, "Skip specific check categories (e.g. --skip RBAC,Ingress)")

	_ = rootCmd.MarkFlagRequired("cluster")
}

func runDiagnostics(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	report := types.DiagnosticReport{
		ClusterName: clusterName,
		Namespace:   namespace,
		StartTime:   time.Now(),
	}

	// ── Resolve context and build clients ────────────────────────────────────────
	resolvedCtx := kubeContext
	if resolvedCtx == "" {
		var err error
		resolvedCtx, err = config.ResolveContext(kubeconfig, clusterName)
		if err != nil {
			return fmt.Errorf("cannot resolve cluster context: %w", err)
		}
	}

	if outputFormat == "text" && !noColor {
		color.New(color.FgHiCyan).Fprintf(os.Stderr, "  ☸  Connecting to cluster %q (context: %s)...\n", clusterName, resolvedCtx)
	}

	clients, err := config.BuildClients(kubeconfig, resolvedCtx)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	if outputFormat == "text" && !noColor {
		color.New(color.FgGreen).Fprintln(os.Stderr, "  ✓  Connected. Running diagnostics...\n")
	}

	// ── Build skip set ────────────────────────────────────────────────────────────
	skipSet := map[types.Category]bool{}
	for _, s := range skipCategories {
		skipSet[types.Category(s)] = true
	}

	// ── Register all checkers ─────────────────────────────────────────────────────
	allCheckers := []Checker{
		checker.NewNamespaceChecker(clients.Kubernetes, namespace),
		checker.NewNodeChecker(clients.Kubernetes, namespace),
		checker.NewPodChecker(clients.Kubernetes, namespace),
		checker.NewStorageChecker(clients.Kubernetes, namespace),
		checker.NewNetworkChecker(clients.Kubernetes, namespace),
		checker.NewAffinityChecker(clients.Kubernetes, namespace),
		checker.NewEventChecker(clients.Kubernetes, namespace),
		checker.NewResourceChecker(clients.Kubernetes, namespace),
		checker.NewRBACChecker(clients.Kubernetes, namespace),
		checker.NewConfigChecker(clients.Kubernetes, namespace),
		checker.NewIngressChecker(clients.Kubernetes, namespace),
	}

	// ── Run checkers concurrently ─────────────────────────────────────────────────
	var (
		mu      sync.Mutex
		results []types.CheckResult
		wg      sync.WaitGroup
	)

	spinner := newSpinner(outputFormat == "text" && !noColor)
	spinner.Start()

	for _, chk := range allCheckers {
		if skipSet[chk.Category()] {
			continue
		}

		wg.Add(1)
		go func(chk Checker) {
			defer wg.Done()
			spinner.Update(fmt.Sprintf("Checking %s...", chk.Category()))
			result := chk.Run(ctx)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(chk)
	}

	wg.Wait()
	spinner.Stop()

	report.Results = results
	report.EndTime = time.Now()
	report.Summary = reporter.ComputeSummary(results)

	// ── Render report ─────────────────────────────────────────────────────────────
	rep := reporter.New(outputFormat, noColor)
	rep.Print(report)

	// Exit with non-zero code if there are critical findings
	if report.Summary.Critical > 0 {
		os.Exit(2)
	}
	if report.Summary.Warning > 0 {
		os.Exit(1)
	}

	return nil
}

// ── Spinner ───────────────────────────────────────────────────────────────────

type spinner struct {
	enabled bool
	ch      chan string
	done    chan struct{}
}

func newSpinner(enabled bool) *spinner {
	return &spinner{
		enabled: enabled,
		ch:      make(chan string, 10),
		done:    make(chan struct{}),
	}
}

func (s *spinner) Start() {
	if !s.enabled {
		return
	}
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		msg := "Running diagnostics..."
		spinColor := color.New(color.FgHiYellow)
		for {
			select {
			case <-s.done:
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case newMsg, ok := <-s.ch:
				if ok {
					msg = newMsg
				}
			default:
			}
			spinColor.Fprintf(os.Stderr, "\r  %s  %s", frames[i%len(frames)], msg)
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}()
}

func (s *spinner) Update(msg string) {
	if !s.enabled {
		return
	}
	select {
	case s.ch <- msg:
	default:
	}
}

func (s *spinner) Stop() {
	if !s.enabled {
		return
	}
	close(s.done)
	time.Sleep(100 * time.Millisecond)
}
