import { useState, useCallback, useRef } from "react";

// ── Palette & tokens ──────────────────────────────────────────────────────────
const C = {
  bg:       "#0a0d12",
  panel:    "#0f1318",
  border:   "#1e2530",
  borderHi: "#2e3d50",
  accent:   "#00d4ff",
  accentDim:"#0090b0",
  green:    "#00e5a0",
  yellow:   "#f5c542",
  red:      "#ff4d6d",
  orange:   "#ff8c42",
  muted:    "#4a5568",
  text:     "#c9d6e3",
  textDim:  "#6b7f94",
  textHi:   "#eaf2ff",
};

// ── Reusable micro-components ─────────────────────────────────────────────────
const Badge = ({ color, children }) => (
  <span style={{
    background: color + "22", color, border: `1px solid ${color}55`,
    borderRadius: 4, padding: "2px 8px", fontSize: 11, fontWeight: 700,
    letterSpacing: "0.06em", textTransform: "uppercase",
  }}>{children}</span>
);

const Pill = ({ label, value, color = C.accent }) => (
  <div style={{
    background: C.panel, border: `1px solid ${C.border}`,
    borderRadius: 8, padding: "10px 16px", minWidth: 130,
  }}>
    <div style={{ fontSize: 11, color: C.textDim, marginBottom: 4, letterSpacing: "0.08em" }}>{label}</div>
    <div style={{ fontSize: 20, fontWeight: 800, color, fontFamily: "monospace" }}>{value}</div>
  </div>
);

const SectionHeader = ({ icon, title, sub }) => (
  <div style={{ marginBottom: 18 }}>
    <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
      <span style={{ fontSize: 18 }}>{icon}</span>
      <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontWeight: 700, color: C.textHi, fontSize: 15 }}>{title}</span>
    </div>
    {sub && <div style={{ fontSize: 12, color: C.textDim, marginTop: 4, paddingLeft: 28 }}>{sub}</div>}
  </div>
);

const Input = ({ label, value, onChange, placeholder, type = "text", small }) => (
  <div style={{ marginBottom: small ? 8 : 14 }}>
    {label && <label style={{ display: "block", fontSize: 11, color: C.textDim, marginBottom: 5, letterSpacing: "0.06em" }}>{label}</label>}
    <input
      type={type}
      value={value}
      onChange={e => onChange(e.target.value)}
      placeholder={placeholder}
      style={{
        width: "100%", boxSizing: "border-box",
        background: "#070a0e", border: `1px solid ${C.border}`,
        borderRadius: 6, padding: small ? "6px 10px" : "9px 12px",
        color: C.text, fontFamily: "'IBM Plex Mono', monospace",
        fontSize: small ? 12 : 13, outline: "none",
      }}
    />
  </div>
);

const Textarea = ({ label, value, onChange, placeholder, rows = 6 }) => (
  <div style={{ marginBottom: 14 }}>
    {label && <label style={{ display: "block", fontSize: 11, color: C.textDim, marginBottom: 5, letterSpacing: "0.06em" }}>{label}</label>}
    <textarea
      value={value}
      onChange={e => onChange(e.target.value)}
      placeholder={placeholder}
      rows={rows}
      style={{
        width: "100%", boxSizing: "border-box",
        background: "#070a0e", border: `1px solid ${C.border}`,
        borderRadius: 6, padding: "9px 12px",
        color: C.text, fontFamily: "'IBM Plex Mono', monospace",
        fontSize: 12, outline: "none", resize: "vertical", lineHeight: 1.6,
      }}
    />
  </div>
);

const Select = ({ label, value, onChange, options }) => (
  <div style={{ marginBottom: 14 }}>
    {label && <label style={{ display: "block", fontSize: 11, color: C.textDim, marginBottom: 5, letterSpacing: "0.06em" }}>{label}</label>}
    <select
      value={value}
      onChange={e => onChange(e.target.value)}
      style={{
        width: "100%", background: "#070a0e", border: `1px solid ${C.border}`,
        borderRadius: 6, padding: "9px 12px", color: C.text,
        fontFamily: "'IBM Plex Mono', monospace", fontSize: 13, outline: "none",
      }}
    >
      {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
    </select>
  </div>
);

const Btn = ({ onClick, children, disabled, variant = "primary", small }) => {
  const bg = variant === "primary" ? C.accent : variant === "danger" ? C.red : C.border;
  const fg = variant === "ghost" ? C.text : "#000";
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        background: disabled ? C.muted : bg, color: disabled ? "#888" : fg,
        border: "none", borderRadius: 6, padding: small ? "6px 14px" : "10px 22px",
        fontSize: small ? 12 : 13, fontWeight: 700, cursor: disabled ? "not-allowed" : "pointer",
        fontFamily: "'IBM Plex Mono', monospace", letterSpacing: "0.04em",
        transition: "opacity 0.15s",
      }}
    >
      {children}
    </button>
  );
};

// ── Cost severity ─────────────────────────────────────────────────────────────
function severityColor(sev) {
  if (!sev) return C.accent;
  const s = sev.toLowerCase();
  if (s.includes("critical") || s.includes("high")) return C.red;
  if (s.includes("medium")) return C.yellow;
  if (s.includes("low")) return C.green;
  return C.accent;
}

// ── Recommendation card ───────────────────────────────────────────────────────
const RecCard = ({ rec, idx }) => {
  const [open, setOpen] = useState(false);
  const col = severityColor(rec.priority);
  return (
    <div style={{
      border: `1px solid ${col}44`, borderRadius: 10, marginBottom: 12,
      background: col + "08", overflow: "hidden",
    }}>
      <div
        onClick={() => setOpen(v => !v)}
        style={{
          display: "flex", alignItems: "center", gap: 12, padding: "14px 18px",
          cursor: "pointer",
        }}
      >
        <span style={{
          width: 28, height: 28, borderRadius: "50%",
          background: col + "33", color: col,
          display: "flex", alignItems: "center", justifyContent: "center",
          fontSize: 12, fontWeight: 800, flexShrink: 0,
        }}>{idx + 1}</span>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontWeight: 700, color: C.textHi, fontSize: 14 }}>{rec.title}</div>
          <div style={{ fontSize: 12, color: C.textDim, marginTop: 2 }}>{rec.category}</div>
        </div>
        <div style={{ textAlign: "right", flexShrink: 0 }}>
          {rec.estimated_savings && (
            <div style={{ color: C.green, fontWeight: 800, fontSize: 15, fontFamily: "monospace" }}>
              {rec.estimated_savings}
            </div>
          )}
          <Badge color={col}>{rec.priority}</Badge>
        </div>
        <span style={{ color: C.muted, marginLeft: 8 }}>{open ? "▲" : "▼"}</span>
      </div>
      {open && (
        <div style={{ padding: "0 18px 16px 18px", borderTop: `1px solid ${C.border}` }}>
          <p style={{ color: C.text, fontSize: 13, lineHeight: 1.7, margin: "12px 0 8px" }}>{rec.description}</p>
          {rec.action_steps && rec.action_steps.length > 0 && (
            <>
              <div style={{ fontSize: 11, color: C.textDim, marginBottom: 6, letterSpacing: "0.06em" }}>ACTION STEPS</div>
              <ol style={{ margin: 0, paddingLeft: 20 }}>
                {rec.action_steps.map((s, i) => (
                  <li key={i} style={{ color: C.text, fontSize: 12, lineHeight: 1.7, marginBottom: 4 }}>{s}</li>
                ))}
              </ol>
            </>
          )}
          {rec.risks && (
            <div style={{ marginTop: 10, padding: "8px 12px", background: C.red + "11", borderRadius: 6, border: `1px solid ${C.red}33` }}>
              <span style={{ fontSize: 11, color: C.red, fontWeight: 700 }}>⚠ RISK: </span>
              <span style={{ fontSize: 12, color: C.text }}>{rec.risks}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// ── Default sample cluster ────────────────────────────────────────────────────
const SAMPLE = {
  cluster_name: "prod-eks-us-east-1",
  region: "us-east-1",
  kubernetes_version: "1.28",
  node_groups: [
    {
      name: "on-demand-general",
      instance_type: "m5.2xlarge",
      desired_size: 12,
      min_size: 4,
      max_size: 20,
      current_utilization_cpu: "23%",
      current_utilization_memory: "31%",
      pricing_model: "on-demand",
      monthly_cost: 3200,
    },
    {
      name: "gpu-workloads",
      instance_type: "p3.2xlarge",
      desired_size: 3,
      min_size: 1,
      max_size: 6,
      current_utilization_cpu: "18%",
      current_utilization_memory: "45%",
      pricing_model: "on-demand",
      monthly_cost: 2700,
    },
    {
      name: "spot-batch",
      instance_type: "c5.xlarge",
      desired_size: 8,
      min_size: 0,
      max_size: 30,
      current_utilization_cpu: "67%",
      current_utilization_memory: "55%",
      pricing_model: "spot",
      monthly_cost: 420,
    },
  ],
  workloads: [
    { namespace: "production", deployments: 24, avg_cpu_request: "250m", avg_mem_request: "512Mi", avg_cpu_limit: "2000m", avg_mem_limit: "4Gi" },
    { namespace: "staging", deployments: 18, avg_cpu_request: "500m", avg_mem_request: "1Gi", avg_cpu_limit: "2000m", avg_mem_limit: "4Gi" },
    { namespace: "monitoring", deployments: 8, avg_cpu_request: "100m", avg_mem_request: "256Mi", avg_cpu_limit: "1000m", avg_mem_limit: "2Gi" },
  ],
  addons: ["aws-load-balancer-controller", "cluster-autoscaler", "kube-proxy", "coredns", "vpc-cni"],
  networking: { load_balancers: 7, nat_gateways: 3, data_transfer_monthly_gb: 2800 },
  storage: { ebs_volumes: 45, avg_volume_size_gb: 80, unused_snapshots: 120, total_storage_monthly_cost: 890 },
  observability: { cloudwatch_log_groups: 34, log_retention_days: 90, monthly_logging_cost: 540 },
  monthly_total_estimate: 8200,
};

// ── Main App ──────────────────────────────────────────────────────────────────
export default function EKSCostOptimizer() {
  const [tab, setTab] = useState("config");   // config | input | results
  const [token, setToken] = useState("");
  const [baseUrl, setBaseUrl] = useState("https://api.anthropic.com");
  const [model, setModel] = useState("claude-sonnet-4-20250514");
  const [clusterJson, setClusterJson] = useState(JSON.stringify(SAMPLE, null, 2));
  const [focus, setFocus] = useState(["all"]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState(null);
  const [streamLog, setStreamLog] = useState("");
  const logRef = useRef(null);

  const toggleFocus = (f) => {
    setFocus(prev => {
      if (f === "all") return ["all"];
      const without = prev.filter(x => x !== "all" && x !== f);
      const added = prev.includes(f) ? without : [...without, f];
      return added.length === 0 ? ["all"] : added;
    });
  };

  const totalMonthlyCost = useCallback(() => {
    try { return JSON.parse(clusterJson)?.monthly_total_estimate || 0; }
    catch { return 0; }
  }, [clusterJson]);

  const runAnalysis = async () => {
    setError("");
    setResult(null);
    setStreamLog("");
    setLoading(true);
    setTab("results");

    let clusterData;
    try { clusterData = JSON.parse(clusterJson); }
    catch (e) { setError("Invalid JSON in cluster data."); setLoading(false); return; }

    const focusAreas = focus.includes("all")
      ? "all areas (node sizing, spot instances, reserved instances, autoscaling, storage, networking, observability, workload right-sizing, scheduling)"
      : focus.join(", ");

    const systemPrompt = `You are a Principal Platform Engineer and AWS EKS Cost Optimization expert with 15+ years of experience. You perform deep cost analysis on EKS clusters and produce structured, actionable recommendations.

ALWAYS respond with ONLY a valid JSON object (no markdown fences) matching this schema:
{
  "summary": {
    "cluster_name": string,
    "analysis_date": string,
    "current_monthly_cost": number,
    "potential_monthly_savings": number,
    "potential_annual_savings": number,
    "savings_percentage": number,
    "overall_health": "Critical|Poor|Fair|Good|Excellent",
    "top_issues": [string]
  },
  "cost_breakdown": [
    { "category": string, "monthly_cost": number, "percentage": number, "status": "optimized|acceptable|needs_attention|critical" }
  ],
  "recommendations": [
    {
      "id": string,
      "title": string,
      "category": string,
      "priority": "Critical|High|Medium|Low",
      "estimated_savings": string,
      "estimated_savings_monthly": number,
      "description": string,
      "action_steps": [string],
      "effort": "Low|Medium|High",
      "risks": string,
      "aws_services": [string]
    }
  ],
  "quick_wins": [string],
  "architecture_insights": string,
  "next_review_date": string
}`;

    const userPrompt = `Analyze this EKS cluster and produce a comprehensive cost optimization report focused on: ${focusAreas}.

CLUSTER DATA:
${JSON.stringify(clusterData, null, 2)}

Provide specific, actionable recommendations. Include precise dollar estimates where possible. Prioritize by ROI. Consider: Spot vs On-Demand mix, Reserved Instance coverage, Savings Plans, node right-sizing, Karpenter vs Cluster Autoscaler, bin-packing efficiency, unused resources, storage optimization, data transfer costs, and observability cost reduction.`;

    try {
      const response = await fetch(`${baseUrl}/v1/messages`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { "x-api-key": token } : {}),
        },
        body: JSON.stringify({
          model,
          max_tokens: 4000,
          system: systemPrompt,
          messages: [{ role: "user", content: userPrompt }],
        }),
      });

      if (!response.ok) {
        const err = await response.text();
        throw new Error(`API error ${response.status}: ${err}`);
      }

      const data = await response.json();
      const raw = data.content?.map(b => b.text || "").join("") || "";

      setStreamLog(raw);

      // Parse JSON, stripping any accidental fences
      const clean = raw.replace(/```json|```/g, "").trim();
      const parsed = JSON.parse(clean);
      setResult(parsed);
    } catch (e) {
      setError(e.message || "Analysis failed. Check your token and cluster data.");
    } finally {
      setLoading(false);
    }
  };

  const healthColor = (h) => {
    if (!h) return C.muted;
    if (h === "Critical") return C.red;
    if (h === "Poor") return C.orange;
    if (h === "Fair") return C.yellow;
    if (h === "Good") return C.green;
    if (h === "Excellent") return C.accent;
    return C.muted;
  };

  const statusColor = (s) => {
    if (s === "critical") return C.red;
    if (s === "needs_attention") return C.orange;
    if (s === "acceptable") return C.yellow;
    if (s === "optimized") return C.green;
    return C.muted;
  };

  // ── Sidebar tabs ────────────────────────────────────────────────────────────
  const navItems = [
    { id: "config", label: "⚙ Config", desc: "API & model" },
    { id: "input",  label: "📡 Cluster", desc: "Input data" },
    { id: "results", label: "📊 Analysis", desc: "AI insights" },
  ];

  return (
    <div style={{
      fontFamily: "'IBM Plex Mono', 'Fira Code', monospace",
      background: C.bg, color: C.text, minHeight: "100vh",
      display: "flex", flexDirection: "column",
    }}>
      <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;600;700&family=IBM+Plex+Sans:wght@400;600;700;800&display=swap" rel="stylesheet" />

      {/* ── Top bar ── */}
      <div style={{
        background: C.panel, borderBottom: `1px solid ${C.border}`,
        padding: "0 28px", display: "flex", alignItems: "center", gap: 16, height: 56,
      }}>
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <div style={{
            width: 32, height: 32, borderRadius: 8,
            background: `linear-gradient(135deg, ${C.accent}, #0060ff)`,
            display: "flex", alignItems: "center", justifyContent: "center",
            fontSize: 16,
          }}>⚡</div>
          <div>
            <div style={{ fontFamily: "'IBM Plex Sans', sans-serif", fontWeight: 800, fontSize: 15, color: C.textHi, letterSpacing: "-0.02em" }}>
              EKS Cost Intelligence
            </div>
            <div style={{ fontSize: 10, color: C.textDim, letterSpacing: "0.1em" }}>POWERED BY GENERATIVE AI</div>
          </div>
        </div>

        <div style={{ flex: 1 }} />

        {/* Nav */}
        <div style={{ display: "flex", gap: 4 }}>
          {navItems.map(n => (
            <button key={n.id} onClick={() => setTab(n.id)} style={{
              background: tab === n.id ? C.accent + "22" : "transparent",
              border: `1px solid ${tab === n.id ? C.accent + "88" : "transparent"}`,
              borderRadius: 6, padding: "6px 16px", cursor: "pointer",
              color: tab === n.id ? C.accent : C.textDim,
              fontFamily: "'IBM Plex Mono', monospace", fontSize: 12, fontWeight: 600,
            }}>{n.label}</button>
          ))}
        </div>

        {result && (
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginLeft: 8 }}>
            <div style={{ width: 8, height: 8, borderRadius: "50%", background: C.green, boxShadow: `0 0 8px ${C.green}` }} />
            <span style={{ fontSize: 11, color: C.green }}>ANALYSIS READY</span>
          </div>
        )}
      </div>

      {/* ── Body ── */}
      <div style={{ flex: 1, padding: 28, maxWidth: 1100, margin: "0 auto", width: "100%", boxSizing: "border-box" }}>

        {/* ══ CONFIG TAB ══════════════════════════════════════════════════════ */}
        {tab === "config" && (
          <div style={{ maxWidth: 640 }}>
            <SectionHeader icon="⚙" title="API Configuration"
              sub="Connect your corporate Gen AI Studio or Anthropic endpoint." />

            <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 12, padding: 24, marginBottom: 20 }}>
              <Input
                label="CORPORATE GEN AI STUDIO / ANTHROPIC API TOKEN"
                value={token}
                onChange={setToken}
                placeholder="sk-ant-... or your enterprise token"
                type="password"
              />
              <Input
                label="BASE URL  (leave default for anthropic.com)"
                value={baseUrl}
                onChange={setBaseUrl}
                placeholder="https://api.anthropic.com"
              />
              <Select
                label="MODEL"
                value={model}
                onChange={setModel}
                options={[
                  { value: "claude-sonnet-4-20250514", label: "claude-sonnet-4-20250514 (recommended)" },
                  { value: "claude-opus-4-5", label: "claude-opus-4-5 (most capable)" },
                  { value: "claude-haiku-4-5-20251001", label: "claude-haiku-4-5-20251001 (fastest)" },
                ]}
              />
            </div>

            <div style={{
              background: C.accent + "0d", border: `1px solid ${C.accent}33`,
              borderRadius: 10, padding: 16, marginBottom: 20,
            }}>
              <div style={{ fontSize: 12, color: C.accent, fontWeight: 700, marginBottom: 6 }}>ℹ ENTERPRISE NOTE</div>
              <div style={{ fontSize: 12, color: C.textDim, lineHeight: 1.7 }}>
                If your corporate Gen AI Studio proxies Anthropic, set your studio base URL above and use your enterprise bearer token.
                The token is sent as <code style={{ color: C.accent }}>x-api-key</code>. No data is stored — all requests go directly from your browser to your configured endpoint.
              </div>
            </div>

            <Btn onClick={() => setTab("input")}>Continue → Cluster Data ▸</Btn>
          </div>
        )}

        {/* ══ INPUT TAB ═══════════════════════════════════════════════════════ */}
        {tab === "input" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 20, flexWrap: "wrap", gap: 12 }}>
              <SectionHeader icon="📡" title="Cluster Data Input"
                sub="Paste your EKS cluster JSON, or edit the pre-loaded sample." />
              <Btn small variant="ghost" onClick={() => setClusterJson(JSON.stringify(SAMPLE, null, 2))}>
                ↺ Load Sample
              </Btn>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 20 }}>
              <div>
                <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 12, padding: 20 }}>
                  <Textarea
                    label="CLUSTER JSON  (node groups · workloads · addons · networking · storage · observability)"
                    value={clusterJson}
                    onChange={setClusterJson}
                    rows={24}
                  />
                </div>
              </div>

              <div>
                <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 12, padding: 20, marginBottom: 16 }}>
                  <div style={{ fontSize: 11, color: C.textDim, marginBottom: 12, letterSpacing: "0.06em" }}>ANALYSIS FOCUS AREAS</div>
                  {[
                    ["all",          "🔭", "Full Analysis"],
                    ["compute",      "💻", "Compute / Node Groups"],
                    ["spot",         "⚡", "Spot Instance Mix"],
                    ["reservations", "📋", "Reserved / Savings Plans"],
                    ["autoscaling",  "📈", "Autoscaling Efficiency"],
                    ["storage",      "💾", "Storage & Snapshots"],
                    ["networking",   "🌐", "Networking & Data Transfer"],
                    ["observability","👁",  "Observability Costs"],
                    ["workloads",    "📦", "Workload Right-sizing"],
                  ].map(([id, ico, lbl]) => (
                    <div
                      key={id}
                      onClick={() => toggleFocus(id)}
                      style={{
                        display: "flex", alignItems: "center", gap: 10,
                        padding: "8px 12px", borderRadius: 6, marginBottom: 4,
                        background: focus.includes(id) ? C.accent + "18" : "transparent",
                        border: `1px solid ${focus.includes(id) ? C.accent + "55" : C.border}`,
                        cursor: "pointer",
                      }}
                    >
                      <span style={{ fontSize: 14 }}>{ico}</span>
                      <span style={{ fontSize: 12, color: focus.includes(id) ? C.accent : C.textDim }}>{lbl}</span>
                      {focus.includes(id) && <span style={{ marginLeft: "auto", color: C.accent, fontSize: 14 }}>✓</span>}
                    </div>
                  ))}
                </div>

                <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 12, padding: 20, marginBottom: 16 }}>
                  <div style={{ fontSize: 11, color: C.textDim, marginBottom: 10 }}>DETECTED MONTHLY COST</div>
                  <div style={{ fontSize: 28, fontWeight: 800, color: C.yellow }}>
                    ${totalMonthlyCost().toLocaleString()}
                  </div>
                  <div style={{ fontSize: 11, color: C.textDim, marginTop: 4 }}>
                    ${(totalMonthlyCost() * 12).toLocaleString()} / year
                  </div>
                </div>

                <Btn onClick={runAnalysis} disabled={loading || !clusterJson.trim()}>
                  {loading ? "⟳ Analysing…" : "⚡ Run AI Analysis"}
                </Btn>
              </div>
            </div>
          </div>
        )}

        {/* ══ RESULTS TAB ═════════════════════════════════════════════════════ */}
        {tab === "results" && (
          <div>
            {loading && (
              <div style={{ textAlign: "center", padding: "60px 0" }}>
                <div style={{ fontSize: 40, marginBottom: 16, animation: "spin 1s linear infinite" }}>⚙</div>
                <div style={{ color: C.accent, fontSize: 14, marginBottom: 8 }}>Running AI cost analysis…</div>
                <div style={{ color: C.textDim, fontSize: 12 }}>Evaluating compute, storage, networking & workload efficiency</div>
                <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
              </div>
            )}

            {error && (
              <div style={{
                background: C.red + "11", border: `1px solid ${C.red}44`,
                borderRadius: 10, padding: 20, marginBottom: 20,
              }}>
                <div style={{ color: C.red, fontWeight: 700, marginBottom: 8 }}>⚠ Analysis Error</div>
                <div style={{ color: C.text, fontSize: 13, whiteSpace: "pre-wrap", wordBreak: "break-word" }}>{error}</div>
                <div style={{ marginTop: 14, display: "flex", gap: 10 }}>
                  <Btn small onClick={() => { setTab("config"); setError(""); }}>Fix Config</Btn>
                  <Btn small variant="ghost" onClick={runAnalysis}>Retry</Btn>
                </div>
              </div>
            )}

            {result && (
              <div>
                {/* ── Summary ── */}
                <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 14, padding: 24, marginBottom: 20 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", flexWrap: "wrap", gap: 12, marginBottom: 20 }}>
                    <div>
                      <div style={{ fontFamily: "'IBM Plex Sans', sans-serif", fontWeight: 800, fontSize: 20, color: C.textHi }}>
                        {result.summary?.cluster_name}
                      </div>
                      <div style={{ fontSize: 12, color: C.textDim, marginTop: 4 }}>
                        Analysis • {result.summary?.analysis_date}
                      </div>
                    </div>
                    <Badge color={healthColor(result.summary?.overall_health)}>
                      {result.summary?.overall_health} Health
                    </Badge>
                  </div>

                  <div style={{ display: "flex", gap: 12, flexWrap: "wrap", marginBottom: 20 }}>
                    <Pill label="CURRENT MONTHLY" value={`$${(result.summary?.current_monthly_cost || 0).toLocaleString()}`} color={C.yellow} />
                    <Pill label="POTENTIAL SAVINGS / MO" value={`$${(result.summary?.potential_monthly_savings || 0).toLocaleString()}`} color={C.green} />
                    <Pill label="ANNUAL SAVINGS" value={`$${(result.summary?.potential_annual_savings || 0).toLocaleString()}`} color={C.green} />
                    <Pill label="SAVINGS %" value={`${result.summary?.savings_percentage || 0}%`} color={C.accent} />
                  </div>

                  {/* Savings progress bar */}
                  <div style={{ marginBottom: 16 }}>
                    <div style={{ fontSize: 11, color: C.textDim, marginBottom: 6 }}>OPTIMISATION OPPORTUNITY</div>
                    <div style={{ height: 8, background: C.border, borderRadius: 4, overflow: "hidden" }}>
                      <div style={{
                        height: "100%", borderRadius: 4,
                        width: `${result.summary?.savings_percentage || 0}%`,
                        background: `linear-gradient(90deg, ${C.green}, ${C.accent})`,
                        transition: "width 1s ease",
                      }} />
                    </div>
                  </div>

                  {result.summary?.top_issues?.length > 0 && (
                    <div>
                      <div style={{ fontSize: 11, color: C.textDim, marginBottom: 8, letterSpacing: "0.06em" }}>TOP ISSUES DETECTED</div>
                      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                        {result.summary.top_issues.map((iss, i) => (
                          <span key={i} style={{
                            fontSize: 12, color: C.orange,
                            background: C.orange + "15", border: `1px solid ${C.orange}44`,
                            borderRadius: 6, padding: "4px 10px",
                          }}>⚡ {iss}</span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                {/* ── Cost Breakdown ── */}
                {result.cost_breakdown?.length > 0 && (
                  <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 14, padding: 24, marginBottom: 20 }}>
                    <SectionHeader icon="💰" title="Cost Breakdown" sub="Monthly spend by category" />
                    <div style={{ display: "grid", gap: 10 }}>
                      {result.cost_breakdown.map((c, i) => (
                        <div key={i} style={{ display: "flex", alignItems: "center", gap: 12 }}>
                          <div style={{ width: 130, fontSize: 12, color: C.text, flexShrink: 0 }}>{c.category}</div>
                          <div style={{ flex: 1, height: 8, background: C.border, borderRadius: 4, overflow: "hidden" }}>
                            <div style={{
                              height: "100%", width: `${Math.min(c.percentage, 100)}%`,
                              background: statusColor(c.status), borderRadius: 4,
                              transition: "width 0.8s ease",
                            }} />
                          </div>
                          <div style={{ width: 80, textAlign: "right", fontFamily: "monospace", fontSize: 13, color: C.textHi }}>${(c.monthly_cost || 0).toLocaleString()}</div>
                          <div style={{ width: 60, textAlign: "right", fontSize: 11, color: C.textDim }}>{c.percentage}%</div>
                          <Badge color={statusColor(c.status)}>{c.status?.replace("_", " ")}</Badge>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* ── Quick Wins ── */}
                {result.quick_wins?.length > 0 && (
                  <div style={{
                    background: C.green + "0a", border: `1px solid ${C.green}33`,
                    borderRadius: 14, padding: 24, marginBottom: 20,
                  }}>
                    <SectionHeader icon="🚀" title="Quick Wins" sub="Low-effort, high-impact actions you can do today" />
                    <div style={{ display: "grid", gap: 8 }}>
                      {result.quick_wins.map((w, i) => (
                        <div key={i} style={{ display: "flex", gap: 10, alignItems: "flex-start" }}>
                          <span style={{ color: C.green, fontSize: 14, marginTop: 1 }}>✓</span>
                          <span style={{ fontSize: 13, color: C.text, lineHeight: 1.6 }}>{w}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* ── Recommendations ── */}
                {result.recommendations?.length > 0 && (
                  <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 14, padding: 24, marginBottom: 20 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 18 }}>
                      <SectionHeader icon="📋" title="Recommendations"
                        sub={`${result.recommendations.length} actionable optimisations identified`} />
                      <div style={{ fontSize: 12, color: C.textDim }}>Click to expand ↓</div>
                    </div>
                    {result.recommendations.map((r, i) => <RecCard key={r.id || i} rec={r} idx={i} />)}
                  </div>
                )}

                {/* ── Architecture Insights ── */}
                {result.architecture_insights && (
                  <div style={{ background: C.panel, border: `1px solid ${C.border}`, borderRadius: 14, padding: 24, marginBottom: 20 }}>
                    <SectionHeader icon="🏗" title="Architecture Insights" />
                    <p style={{ fontSize: 13, color: C.text, lineHeight: 1.8, margin: 0 }}>{result.architecture_insights}</p>
                  </div>
                )}

                {/* ── Footer ── */}
                <div style={{ display: "flex", gap: 12, marginTop: 24, flexWrap: "wrap" }}>
                  <Btn small onClick={runAnalysis} disabled={loading}>⟳ Re-run Analysis</Btn>
                  <Btn small variant="ghost" onClick={() => setTab("input")}>← Edit Cluster Data</Btn>
                  <Btn small variant="ghost" onClick={() => {
                    const blob = new Blob([JSON.stringify(result, null, 2)], { type: "application/json" });
                    const a = document.createElement("a"); a.href = URL.createObjectURL(blob);
                    a.download = `eks-cost-report-${Date.now()}.json`; a.click();
                  }}>⬇ Export JSON</Btn>
                </div>
              </div>
            )}

            {!loading && !result && !error && (
              <div style={{ textAlign: "center", padding: "80px 0" }}>
                <div style={{ fontSize: 48, marginBottom: 16 }}>📊</div>
                <div style={{ color: C.textDim, fontSize: 14, marginBottom: 20 }}>No analysis yet. Configure your cluster and run the AI analysis.</div>
                <Btn onClick={() => setTab("input")}>← Go to Cluster Input</Btn>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
