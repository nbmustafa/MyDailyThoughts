package types

import "time"

// Severity levels for findings
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
	SeverityOK       Severity = "OK"
)

// Category of diagnostic check
type Category string

const (
	CategoryPods        Category = "Pods"
	CategoryNodes       Category = "Nodes"
	CategoryStorage     Category = "Storage"
	CategoryNetwork     Category = "Network"
	CategoryServices    Category = "Services"
	CategoryEvents      Category = "Events"
	CategoryAffinity    Category = "Affinity/Scheduling"
	CategoryLabels      Category = "Labels/Selectors"
	CategoryCalico      Category = "Calico"
	CategoryResources   Category = "Resources/Quotas"
	CategoryRBAC        Category = "RBAC"
	CategoryConfig      Category = "Config/Secrets"
	CategoryNamespace   Category = "Namespace"
	CategoryDNS         Category = "DNS"
	CategoryIngress     Category = "Ingress"
	CategoryCertificates Category = "Certificates"
)

// Finding represents a single diagnostic finding
type Finding struct {
	Category    Category
	Severity    Severity
	Title       string
	Description string
	Resource    string
	Namespace   string
	Suggestion  string
	Timestamp   time.Time
}

// CheckResult holds results from a single checker
type CheckResult struct {
	Category Category
	Findings []Finding
	Duration time.Duration
	Error    error
}

// DiagnosticReport is the full report from all checkers
type DiagnosticReport struct {
	ClusterName string
	Namespace   string // empty = cluster-wide
	StartTime   time.Time
	EndTime     time.Time
	Results     []CheckResult
	Summary     Summary
}

// Summary aggregates finding counts
type Summary struct {
	Total    int
	Critical int
	Warning  int
	Info     int
	OK       int
}

// DiagnosticOptions configures what and how to check
type DiagnosticOptions struct {
	ClusterName   string
	Namespace     string
	KubeContext   string
	KubeconfigPath string
	Timeout       time.Duration
	Verbose       bool
	SkipCategories []Category
	OutputFormat  string // "text", "json", "markdown"
	NoColor       bool
}
