import { useState, useEffect, useRef } from "react";

// ─── Design Tokens ────────────────────────────────────────────────────────────
const CSS = `
  @import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=IBM+Plex+Sans:wght@300;400;500;600;700&display=swap');

  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg-void: #080c10;
    --bg-base: #0d1117;
    --bg-raised: #131920;
    --bg-card: #161d26;
    --bg-hover: #1c2733;
    --bg-active: #1f2e3d;
    --border: #1e2d3d;
    --border-bright: #2a3f55;
    --text-primary: #e6edf3;
    --text-secondary: #8b949e;
    --text-muted: #484f58;
    --accent-cyan: #00d4ff;
    --accent-cyan-dim: rgba(0,212,255,0.12);
    --accent-green: #3fb950;
    --accent-green-dim: rgba(63,185,80,0.12);
    --accent-amber: #d29922;
    --accent-amber-dim: rgba(210,153,34,0.12);
    --accent-red: #f85149;
    --accent-red-dim: rgba(248,81,73,0.12);
    --accent-purple: #a371f7;
    --accent-purple-dim: rgba(163,113,247,0.12);
    --font-mono: 'IBM Plex Mono', monospace;
    --font-sans: 'IBM Plex Sans', sans-serif;
    --radius: 6px;
    --radius-lg: 10px;
    --shadow: 0 4px 24px rgba(0,0,0,0.4);
    --shadow-glow-cyan: 0 0 20px rgba(0,212,255,0.15);
    --shadow-glow-green: 0 0 20px rgba(63,185,80,0.15);
  }

  body { background: var(--bg-void); color: var(--text-primary); font-family: var(--font-sans); }

  ::-webkit-scrollbar { width: 6px; height: 6px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: var(--border-bright); border-radius: 3px; }

  .idp-root {
    display: flex;
    height: 100vh;
    overflow: hidden;
    background: var(--bg-void);
    position: relative;
  }

  /* Grid overlay for atmosphere */
  .idp-root::before {
    content: '';
    position: fixed;
    inset: 0;
    background-image: 
      linear-gradient(rgba(0,212,255,0.02) 1px, transparent 1px),
      linear-gradient(90deg, rgba(0,212,255,0.02) 1px, transparent 1px);
    background-size: 40px 40px;
    pointer-events: none;
    z-index: 0;
  }

  /* ─── Sidebar ─── */
  .sidebar {
    width: 220px;
    min-width: 220px;
    background: var(--bg-base);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    z-index: 10;
    position: relative;
  }

  .sidebar-logo {
    padding: 18px 20px 16px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .logo-mark {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, var(--accent-cyan), #0066ff);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-weight: 600;
    font-size: 14px;
    color: #000;
    box-shadow: 0 0 16px rgba(0,212,255,0.3);
  }

  .logo-text { display: flex; flex-direction: column; }
  .logo-title { font-size: 13px; font-weight: 600; letter-spacing: 0.04em; color: var(--text-primary); }
  .logo-sub { font-size: 10px; color: var(--text-muted); font-family: var(--font-mono); letter-spacing: 0.06em; }

  .cluster-selector {
    margin: 12px;
    padding: 8px 10px;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: all 0.15s;
  }
  .cluster-selector:hover { border-color: var(--border-bright); background: var(--bg-hover); }
  .cluster-dot { width: 7px; height: 7px; border-radius: 50%; background: var(--accent-green); box-shadow: 0 0 6px var(--accent-green); flex-shrink: 0; }
  .cluster-name { font-size: 11px; font-family: var(--font-mono); color: var(--text-secondary); flex: 1; }
  .cluster-region { font-size: 10px; color: var(--text-muted); }

  .nav-section { padding: 6px 8px 4px; }
  .nav-label { font-size: 9px; font-weight: 600; letter-spacing: 0.12em; color: var(--text-muted); text-transform: uppercase; padding: 6px 8px 4px; }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 7px 10px;
    border-radius: var(--radius);
    cursor: pointer;
    transition: all 0.12s;
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 400;
    position: relative;
  }
  .nav-item:hover { background: var(--bg-hover); color: var(--text-primary); }
  .nav-item.active { background: var(--accent-cyan-dim); color: var(--accent-cyan); font-weight: 500; }
  .nav-item.active::before {
    content: '';
    position: absolute;
    left: -8px;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 60%;
    background: var(--accent-cyan);
    border-radius: 2px;
    box-shadow: 0 0 8px var(--accent-cyan);
  }
  .nav-icon { font-size: 14px; width: 18px; text-align: center; flex-shrink: 0; }
  .nav-badge {
    margin-left: auto;
    background: var(--accent-red);
    color: #fff;
    font-size: 9px;
    font-weight: 600;
    padding: 1px 5px;
    border-radius: 10px;
    font-family: var(--font-mono);
  }

  .sidebar-footer {
    margin-top: auto;
    padding: 12px;
    border-top: 1px solid var(--border);
  }
  .tenant-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px;
    background: var(--accent-purple-dim);
    border: 1px solid rgba(163,113,247,0.2);
    border-radius: var(--radius);
  }
  .tenant-avatar {
    width: 24px; height: 24px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--accent-purple), #6e40c9);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 700;
    color: #fff;
    flex-shrink: 0;
  }
  .tenant-info { flex: 1; }
  .tenant-name { font-size: 11px; font-weight: 600; color: var(--text-primary); }
  .tenant-ns { font-size: 9px; color: var(--accent-purple); font-family: var(--font-mono); }

  /* ─── Main ─── */
  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    z-index: 1;
  }

  .topbar {
    height: 52px;
    background: var(--bg-base);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    padding: 0 24px;
    gap: 16px;
    flex-shrink: 0;
  }
  .topbar-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
  .topbar-breadcrumb { font-size: 12px; color: var(--text-muted); font-family: var(--font-mono); }
  .topbar-spacer { flex: 1; }

  .status-chip {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 11px;
    font-family: var(--font-mono);
    font-weight: 500;
    border: 1px solid;
  }
  .status-chip.healthy { background: var(--accent-green-dim); color: var(--accent-green); border-color: rgba(63,185,80,0.3); }
  .status-chip.warning { background: var(--accent-amber-dim); color: var(--accent-amber); border-color: rgba(210,153,34,0.3); }
  .status-chip.error { background: var(--accent-red-dim); color: var(--accent-red); border-color: rgba(248,81,73,0.3); }
  .status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; animation: pulse 2s infinite; }
  @keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }

  .content { flex: 1; overflow-y: auto; padding: 24px; }

  /* ─── Grid Cards ─── */
  .metric-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
  .stat-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 18px 20px;
    position: relative;
    overflow: hidden;
    transition: border-color 0.15s;
  }
  .stat-card:hover { border-color: var(--border-bright); }
  .stat-card::after {
    content: '';
    position: absolute;
    top: 0; left: 0; right: 0;
    height: 2px;
    background: var(--accent-color, var(--accent-cyan));
    opacity: 0.6;
  }
  .stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: 8px; }
  .stat-value { font-size: 28px; font-weight: 700; font-family: var(--font-mono); color: var(--text-primary); }
  .stat-sub { font-size: 11px; color: var(--text-secondary); margin-top: 4px; display: flex; align-items: center; gap: 4px; }
  .stat-up { color: var(--accent-green); }
  .stat-down { color: var(--accent-red); }

  /* ─── Section Headers ─── */
  .section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
  .section-title { font-size: 13px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.08em; display: flex; align-items: center; gap: 8px; }
  .section-title::before { content: '//'; font-family: var(--font-mono); color: var(--accent-cyan); opacity: 0.7; }

  /* ─── Buttons ─── */
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 14px;
    border-radius: var(--radius);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid;
    transition: all 0.15s;
    font-family: var(--font-sans);
    white-space: nowrap;
  }
  .btn-primary {
    background: var(--accent-cyan);
    color: #000;
    border-color: var(--accent-cyan);
    font-weight: 600;
  }
  .btn-primary:hover { background: #00b8d9; box-shadow: 0 0 16px rgba(0,212,255,0.3); }
  .btn-secondary { background: var(--bg-raised); color: var(--text-secondary); border-color: var(--border); }
  .btn-secondary:hover { border-color: var(--border-bright); color: var(--text-primary); background: var(--bg-hover); }
  .btn-danger { background: var(--accent-red-dim); color: var(--accent-red); border-color: rgba(248,81,73,0.3); }
  .btn-danger:hover { background: rgba(248,81,73,0.2); }
  .btn-green { background: var(--accent-green-dim); color: var(--accent-green); border-color: rgba(63,185,80,0.3); }
  .btn-green:hover { background: rgba(63,185,80,0.2); }
  .btn-sm { padding: 4px 10px; font-size: 11px; }

  /* ─── Table ─── */
  .data-table { width: 100%; border-collapse: collapse; }
  .data-table th {
    text-align: left;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-muted);
    padding: 8px 16px;
    border-bottom: 1px solid var(--border);
  }
  .data-table td { padding: 10px 16px; font-size: 12px; border-bottom: 1px solid rgba(30,45,61,0.5); vertical-align: middle; }
  .data-table tr:hover td { background: var(--bg-hover); }
  .data-table tr:last-child td { border-bottom: none; }

  /* ─── Workload Status Badges ─── */
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 600;
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border: 1px solid;
  }
  .badge-running { background: var(--accent-green-dim); color: var(--accent-green); border-color: rgba(63,185,80,0.25); }
  .badge-pending { background: var(--accent-amber-dim); color: var(--accent-amber); border-color: rgba(210,153,34,0.25); }
  .badge-failed { background: var(--accent-red-dim); color: var(--accent-red); border-color: rgba(248,81,73,0.25); }
  .badge-scaling { background: var(--accent-cyan-dim); color: var(--accent-cyan); border-color: rgba(0,212,255,0.25); }
  .badge-canary { background: var(--accent-purple-dim); color: var(--accent-purple); border-color: rgba(163,113,247,0.25); }

  /* ─── Deploy Form ─── */
  .deploy-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  .form-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .form-card-header {
    padding: 14px 20px;
    border-bottom: 1px solid var(--border);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .form-card-body { padding: 20px; }

  .form-row { margin-bottom: 16px; }
  .form-label { display: block; font-size: 11px; font-weight: 500; color: var(--text-secondary); margin-bottom: 6px; letter-spacing: 0.04em; }
  .form-label span { color: var(--accent-red); margin-left: 2px; }
  .form-input, .form-select, .form-textarea {
    width: 100%;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 12px;
    font-size: 12px;
    color: var(--text-primary);
    font-family: var(--font-sans);
    transition: all 0.15s;
    outline: none;
  }
  .form-input:focus, .form-select:focus, .form-textarea:focus {
    border-color: var(--accent-cyan);
    box-shadow: 0 0 0 3px rgba(0,212,255,0.1);
  }
  .form-input::placeholder { color: var(--text-muted); }
  .form-select { cursor: pointer; appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%238b949e' d='M6 8L1 3h10z'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 10px center; padding-right: 28px; }
  .form-select option { background: var(--bg-raised); }
  .form-textarea { resize: vertical; min-height: 80px; font-family: var(--font-mono); font-size: 11px; }
  .form-hint { font-size: 10px; color: var(--text-muted); margin-top: 4px; font-family: var(--font-mono); }

  .form-grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
  .form-grid-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 14px; }

  .slider-container { display: flex; align-items: center; gap: 10px; }
  .slider {
    flex: 1;
    -webkit-appearance: none;
    height: 4px;
    background: var(--border);
    border-radius: 2px;
    outline: none;
  }
  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: var(--accent-cyan);
    cursor: pointer;
    box-shadow: 0 0 8px rgba(0,212,255,0.4);
  }
  .slider-val { font-family: var(--font-mono); font-size: 12px; color: var(--accent-cyan); min-width: 28px; text-align: right; }

  .toggle-group { display: flex; gap: 4px; background: var(--bg-raised); border: 1px solid var(--border); border-radius: var(--radius); padding: 3px; }
  .toggle-opt {
    flex: 1;
    padding: 5px 10px;
    text-align: center;
    font-size: 11px;
    border-radius: 4px;
    cursor: pointer;
    color: var(--text-muted);
    transition: all 0.15s;
    font-weight: 500;
  }
  .toggle-opt.active { background: var(--accent-cyan); color: #000; font-weight: 600; }

  /* ─── Manifest Preview ─── */
  .manifest-panel {
    background: var(--bg-base);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  .manifest-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 10px;
    background: var(--bg-raised);
  }
  .manifest-tabs { display: flex; gap: 2px; flex: 1; }
  .manifest-tab {
    padding: 4px 10px;
    font-size: 11px;
    font-family: var(--font-mono);
    color: var(--text-muted);
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.12s;
  }
  .manifest-tab.active { background: var(--accent-cyan-dim); color: var(--accent-cyan); }
  .manifest-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }
  .yaml-code {
    font-family: var(--font-mono);
    font-size: 11px;
    line-height: 1.7;
    color: var(--text-secondary);
    white-space: pre;
  }
  .yaml-key { color: #79c0ff; }
  .yaml-val { color: #a5d6ff; }
  .yaml-str { color: #a8daff; }
  .yaml-comment { color: var(--text-muted); font-style: italic; }
  .yaml-num { color: var(--accent-amber); }
  .yaml-bool { color: var(--accent-purple); }
  .yaml-section { color: var(--accent-cyan); font-weight: 600; }

  /* ─── AI Panel ─── */
  .ai-panel {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    margin-top: 20px;
  }
  .ai-header {
    padding: 14px 18px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 10px;
    background: linear-gradient(90deg, rgba(0,212,255,0.05), transparent);
  }
  .ai-spark { font-size: 16px; }
  .ai-title { font-size: 12px; font-weight: 600; color: var(--accent-cyan); }
  .ai-subtitle { font-size: 10px; color: var(--text-muted); font-family: var(--font-mono); }
  .ai-body { padding: 16px 18px; }
  .ai-input-row { display: flex; gap: 8px; }
  .ai-input {
    flex: 1;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 12px;
    font-size: 12px;
    color: var(--text-primary);
    outline: none;
    font-family: var(--font-sans);
    transition: all 0.15s;
  }
  .ai-input:focus { border-color: var(--accent-cyan); box-shadow: 0 0 0 3px rgba(0,212,255,0.1); }
  .ai-input::placeholder { color: var(--text-muted); }
  .ai-response {
    margin-top: 14px;
    padding: 12px 14px;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.6;
    font-family: var(--font-mono);
    min-height: 60px;
    position: relative;
  }
  .ai-typing::after {
    content: '▋';
    animation: blink 1s infinite;
    color: var(--accent-cyan);
  }
  @keyframes blink { 0%,100% { opacity: 1; } 50% { opacity: 0; } }

  /* ─── Service Catalog ─── */
  .catalog-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }
  .catalog-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 18px;
    cursor: pointer;
    transition: all 0.15s;
    position: relative;
    overflow: hidden;
  }
  .catalog-card:hover {
    border-color: var(--border-bright);
    transform: translateY(-2px);
    box-shadow: var(--shadow);
  }
  .catalog-card-icon {
    font-size: 24px;
    margin-bottom: 10px;
    display: block;
  }
  .catalog-card-name { font-size: 13px; font-weight: 600; margin-bottom: 4px; }
  .catalog-card-desc { font-size: 11px; color: var(--text-muted); line-height: 1.5; margin-bottom: 12px; }
  .catalog-card-tags { display: flex; flex-wrap: wrap; gap: 4px; }
  .catalog-tag {
    font-size: 9px;
    padding: 2px 6px;
    border-radius: 3px;
    font-family: var(--font-mono);
    font-weight: 500;
  }

  /* ─── Topology / Namespace View ─── */
  .ns-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
  .ns-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    transition: border-color 0.15s;
  }
  .ns-card:hover { border-color: var(--border-bright); }
  .ns-card-header {
    padding: 12px 16px;
    background: var(--bg-raised);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .ns-icon { font-size: 14px; }
  .ns-name { font-size: 12px; font-weight: 600; font-family: var(--font-mono); flex: 1; }
  .ns-card-body { padding: 14px 16px; }

  .resource-bar { margin-bottom: 10px; }
  .resource-bar-label { display: flex; justify-content: space-between; font-size: 10px; color: var(--text-muted); margin-bottom: 4px; }
  .bar-track { height: 5px; background: var(--border); border-radius: 3px; overflow: hidden; }
  .bar-fill { height: 100%; border-radius: 3px; transition: width 0.6s ease; }

  /* ─── Pipelines ─── */
  .pipeline-list { display: flex; flex-direction: column; gap: 12px; }
  .pipeline-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 14px 18px;
    display: flex;
    align-items: center;
    gap: 16px;
    transition: border-color 0.15s;
  }
  .pipeline-card:hover { border-color: var(--border-bright); }
  .pipeline-status-icon { font-size: 18px; flex-shrink: 0; }
  .pipeline-info { flex: 1; }
  .pipeline-name { font-size: 13px; font-weight: 600; margin-bottom: 2px; }
  .pipeline-meta { font-size: 11px; color: var(--text-muted); font-family: var(--font-mono); }
  .pipeline-stages { display: flex; align-items: center; gap: 4px; margin-top: 10px; }
  .pipeline-stage {
    height: 4px;
    flex: 1;
    border-radius: 2px;
    background: var(--border);
    position: relative;
    overflow: hidden;
  }
  .pipeline-stage.done { background: var(--accent-green); }
  .pipeline-stage.active { background: var(--accent-cyan); animation: shimmer 1.5s infinite; }
  .pipeline-stage.fail { background: var(--accent-red); }
  @keyframes shimmer {
    0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; }
  }
  .pipeline-stage-label { font-size: 9px; color: var(--text-muted); margin-top: 4px; text-align: center; }

  /* ─── Observability ─── */
  .obs-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 16px; }
  .log-panel {
    background: var(--bg-base);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .log-header { padding: 10px 16px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 10px; background: var(--bg-raised); }
  .log-title { font-size: 11px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.08em; flex: 1; }
  .log-body { padding: 12px; height: 300px; overflow-y: auto; font-family: var(--font-mono); font-size: 10px; }
  .log-line { display: flex; gap: 12px; margin-bottom: 4px; line-height: 1.5; }
  .log-ts { color: var(--text-muted); flex-shrink: 0; }
  .log-level-info { color: var(--accent-cyan); }
  .log-level-warn { color: var(--accent-amber); }
  .log-level-error { color: var(--accent-red); }
  .log-msg { color: var(--text-secondary); }

  /* ─── Alert Toast ─── */
  .toast-container { position: fixed; bottom: 24px; right: 24px; z-index: 9999; display: flex; flex-direction: column; gap: 8px; }
  .toast {
    background: var(--bg-raised);
    border: 1px solid var(--border-bright);
    border-radius: var(--radius-lg);
    padding: 12px 16px;
    display: flex;
    align-items: flex-start;
    gap: 10px;
    min-width: 280px;
    max-width: 360px;
    box-shadow: var(--shadow);
    animation: slideIn 0.25s ease;
  }
  @keyframes slideIn { from { transform: translateX(40px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
  .toast.success { border-left: 3px solid var(--accent-green); }
  .toast.error { border-left: 3px solid var(--accent-red); }
  .toast.info { border-left: 3px solid var(--accent-cyan); }
  .toast-icon { font-size: 14px; flex-shrink: 0; margin-top: 1px; }
  .toast-content { flex: 1; }
  .toast-title { font-size: 12px; font-weight: 600; margin-bottom: 2px; }
  .toast-body { font-size: 11px; color: var(--text-muted); line-height: 1.4; }

  /* ─── Env Switcher ─── */
  .env-tabs { display: flex; gap: 6px; margin-bottom: 20px; }
  .env-tab {
    padding: 6px 14px;
    border-radius: 20px;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid var(--border);
    color: var(--text-muted);
    transition: all 0.15s;
    font-family: var(--font-mono);
  }
  .env-tab:hover { border-color: var(--border-bright); color: var(--text-secondary); }
  .env-tab.dev.active { background: rgba(63,185,80,0.12); color: var(--accent-green); border-color: rgba(63,185,80,0.3); }
  .env-tab.staging.active { background: rgba(210,153,34,0.12); color: var(--accent-amber); border-color: rgba(210,153,34,0.3); }
  .env-tab.prod.active { background: rgba(248,81,73,0.12); color: var(--accent-red); border-color: rgba(248,81,73,0.3); }

  /* ─── Responsive polish ─── */
  @media (max-width: 1280px) {
    .metric-grid { grid-template-columns: repeat(2, 1fr); }
    .deploy-layout { grid-template-columns: 1fr; }
    .catalog-grid { grid-template-columns: repeat(2, 1fr); }
  }
`;

// ─── Data ─────────────────────────────────────────────────────────────────────
const WORKLOADS = [
  { name: "payments-api", namespace: "payments", type: "Deployment", replicas: "3/3", status: "running", cpu: "210m", memory: "256Mi", image: "payments-api:v2.3.1", env: "prod", hpa: true },
  { name: "auth-service", namespace: "auth", type: "Deployment", replicas: "2/2", status: "running", cpu: "85m", memory: "128Mi", image: "auth-svc:v1.8.4", env: "prod", hpa: false },
  { name: "data-pipeline", namespace: "analytics", type: "CronJob", replicas: "—", status: "running", cpu: "450m", memory: "512Mi", image: "data-pipeline:v3.1.0", env: "prod", hpa: false },
  { name: "frontend-web", namespace: "web", type: "Deployment", replicas: "4/4", status: "scaling", cpu: "120m", memory: "192Mi", image: "frontend:v5.2.0", env: "prod", hpa: true },
  { name: "notification-svc", namespace: "comms", type: "Deployment", replicas: "1/2", status: "pending", cpu: "30m", memory: "64Mi", image: "notif-svc:v0.9.1", env: "staging", hpa: false },
  { name: "ml-inference", namespace: "ml", type: "StatefulSet", replicas: "2/2", status: "running", cpu: "1200m", memory: "4Gi", image: "ml-serve:v2.0.0", env: "prod", hpa: true },
  { name: "canary-checkout", namespace: "checkout", type: "Deployment", replicas: "1/5", status: "canary", cpu: "60m", memory: "96Mi", image: "checkout:v3.0.0-rc1", env: "prod", hpa: false },
  { name: "worker-queue", namespace: "jobs", type: "Deployment", replicas: "0/3", status: "failed", cpu: "0m", memory: "0Mi", image: "worker:v1.2.3", env: "prod", hpa: false },
];

const NAMESPACES = [
  { name: "payments", team: "Platform", cpuUsed: 72, memUsed: 58, pods: 12, quota: "4 CPU / 8Gi" },
  { name: "auth", team: "Security", cpuUsed: 34, memUsed: 41, pods: 6, quota: "2 CPU / 4Gi" },
  { name: "analytics", team: "Data", cpuUsed: 89, memUsed: 76, pods: 18, quota: "8 CPU / 16Gi" },
  { name: "ml", team: "ML Ops", cpuUsed: 95, memUsed: 88, pods: 8, quota: "16 CPU / 32Gi" },
];

const CATALOG_ITEMS = [
  { icon: "🌐", name: "Web Service", desc: "HTTP/HTTPS service with Ingress, HPA, and PodDisruptionBudget pre-configured.", tags: ["deployment", "hpa", "ingress"], color: "var(--accent-cyan)" },
  { icon: "⚙️", name: "Background Worker", desc: "Queue consumer with autoscaling based on queue depth via KEDA.", tags: ["deployment", "keda", "kms"], color: "var(--accent-purple)" },
  { icon: "⏱️", name: "Cron Job", desc: "Scheduled workload with retry logic, job history, and alerting.", tags: ["cronjob", "monitoring"], color: "var(--accent-amber)" },
  { icon: "🗄️", name: "StatefulSet DB", desc: "Stateful workload with persistent volumes, anti-affinity, and backup hooks.", tags: ["statefulset", "pvc", "backup"], color: "var(--accent-green)" },
  { icon: "🔌", name: "gRPC Service", desc: "Internal gRPC service with mTLS via service mesh and load balancing.", tags: ["grpc", "mtls", "istio"], color: "#ff6b6b" },
  { icon: "📦", name: "ML Serving", desc: "Model serving endpoint with GPU node affinity, KNative autoscaling.", tags: ["gpu", "knative", "serving"], color: "var(--accent-purple)" },
];

const PIPELINES = [
  { name: "payments-api", branch: "main", commit: "a3f9c2d", status: "success", stages: ["done","done","done","done","done"], time: "4m 22s", ago: "12min ago" },
  { name: "frontend-web", branch: "feature/checkout-v2", commit: "b7e1a09", status: "running", stages: ["done","done","active","",""], time: "2m 14s", ago: "live" },
  { name: "ml-inference", branch: "main", commit: "c1d8f44", status: "failed", stages: ["done","done","fail","",""], time: "1m 08s", ago: "1h ago" },
  { name: "auth-service", branch: "main", commit: "f2a3b56", status: "success", stages: ["done","done","done","done","done"], time: "3m 51s", ago: "2h ago" },
];

const LOGS = [
  { ts: "10:42:18.221", level: "INFO", msg: "payments-api: request processed in 42ms [POST /v1/charge]", pod: "payments-api-7b9f-p2x4k" },
  { ts: "10:42:18.334", level: "INFO", msg: "auth-service: JWT validated for user_id=u_938fba [scope=payments:write]", pod: "auth-svc-5dc9-m1z2j" },
  { ts: "10:42:19.001", level: "WARN", msg: "notification-svc: pod evicted due to OOMKilled — restarting (restart_count=3)", pod: "notif-svc-2ab1-q8w5p" },
  { ts: "10:42:19.445", level: "INFO", msg: "frontend-web: HPA scaled replicas 3→4 [cpu_util=82%]", pod: "kube-system/hpa-controller" },
  { ts: "10:42:20.112", level: "ERROR", msg: "worker-queue: failed to connect to Redis — ECONNREFUSED — CrashLoopBackOff", pod: "worker-7xc3-r9k2m" },
  { ts: "10:42:20.789", level: "INFO", msg: "ml-inference: model request latency p99=210ms [endpoint=/predict]", pod: "ml-serve-0" },
  { ts: "10:42:21.003", level: "INFO", msg: "data-pipeline: batch completed — 248,192 records processed in 18.4s", pod: "data-pipeline-28431-wk9f2" },
  { ts: "10:42:21.556", level: "WARN", msg: "canary-checkout: error rate 2.1% above baseline [threshold=1%]", pod: "argo-rollouts/checkout-canary" },
];

// ─── YAML Renderer ────────────────────────────────────────────────────────────
function YamlLine({ line }) {
  if (line.startsWith("#")) return <div><span className="yaml-comment">{line}</span></div>;
  if (line.startsWith("---") || line.match(/^[a-zA-Z]/)) {
    const colonIdx = line.indexOf(":");
    if (colonIdx === -1) return <div><span className="yaml-section">{line}</span></div>;
    const key = line.slice(0, colonIdx + 1);
    const val = line.slice(colonIdx + 1);
    return <div><span className="yaml-section">{key}</span><span className="yaml-val">{val}</span></div>;
  }
  const match = line.match(/^(\s*)([\w-]+):\s*(.*)$/);
  if (!match) return <div><span className="yaml-val">{line}</span></div>;
  const [, indent, key, val] = match;
  let valEl;
  if (!val) valEl = null;
  else if (val === "true" || val === "false") valEl = <span className="yaml-bool"> {val}</span>;
  else if (!isNaN(val.trim()) && val.trim() !== "") valEl = <span className="yaml-num"> {val}</span>;
  else if (val.startsWith('"') || val.startsWith("'")) valEl = <span className="yaml-str"> {val}</span>;
  else valEl = <span className="yaml-val"> {val}</span>;
  return <div>{indent}<span className="yaml-key">{key}:</span>{valEl}</div>;
}

// ─── Generate Manifest ────────────────────────────────────────────────────────
function generateManifest(form) {
  const { name, namespace, image, port, replicas, cpu, memory, type, env } = form;
  const lines = [
    `# ── Generated by EKS-IDP ─────────────────────────────────────────`,
    `# Tenant: ${namespace} | Env: ${env} | ${new Date().toISOString().slice(0,19)}Z`,
    `---`,
    `apiVersion: apps/v1`,
    `kind: ${type || "Deployment"}`,
    `metadata:`,
    `  name: ${name || "my-app"}`,
    `  namespace: ${namespace || "default"}`,
    `  labels:`,
    `    app.kubernetes.io/name: "${name || "my-app"}"`,
    `    app.kubernetes.io/managed-by: "eks-idp"`,
    `    idp.platform/tenant: "${namespace || "default"}"`,
    `    idp.platform/env: "${env || "dev"}"`,
    `spec:`,
    `  replicas: ${replicas || 2}`,
    `  selector:`,
    `    matchLabels:`,
    `      app: ${name || "my-app"}`,
    `  template:`,
    `    metadata:`,
    `      labels:`,
    `        app: ${name || "my-app"}`,
    `      annotations:`,
    `        prometheus.io/scrape: "true"`,
    `        prometheus.io/port: "${port || 8080}"`,
    `    spec:`,
    `      serviceAccountName: ${namespace || "default"}-sa`,
    `      securityContext:`,
    `        runAsNonRoot: true`,
    `        fsGroup: 1000`,
    `      containers:`,
    `        - name: ${name || "my-app"}`,
    `          image: ${image || "my-registry/my-app:latest"}`,
    `          imagePullPolicy: Always`,
    `          ports:`,
    `            - containerPort: ${port || 8080}`,
    `              protocol: TCP`,
    `          resources:`,
    `            requests:`,
    `              cpu: "${Math.round((cpu||500)*0.5)}m"`,
    `              memory: "${Math.round((memory||256)*0.5)}Mi"`,
    `            limits:`,
    `              cpu: "${cpu || 500}m"`,
    `              memory: "${memory || 256}Mi"`,
    `          readinessProbe:`,
    `            httpGet:`,
    `              path: /healthz`,
    `              port: ${port || 8080}`,
    `            initialDelaySeconds: 10`,
    `            periodSeconds: 5`,
    `          livenessProbe:`,
    `            httpGet:`,
    `              path: /healthz`,
    `              port: ${port || 8080}`,
    `            initialDelaySeconds: 30`,
    `            periodSeconds: 10`,
    `          envFrom:`,
    `            - secretRef:`,
    `                name: ${name || "my-app"}-secrets`,
    `            - configMapRef:`,
    `                name: ${name || "my-app"}-config`,
    `      topologySpreadConstraints:`,
    `        - maxSkew: 1`,
    `          topologyKey: topology.kubernetes.io/zone`,
    `          whenUnsatisfiable: DoNotSchedule`,
  ];
  return lines;
}

function generateHPA(form) {
  const { name, namespace } = form;
  return [
    `---`,
    `apiVersion: autoscaling/v2`,
    `kind: HorizontalPodAutoscaler`,
    `metadata:`,
    `  name: ${name || "my-app"}-hpa`,
    `  namespace: ${namespace || "default"}`,
    `spec:`,
    `  scaleTargetRef:`,
    `    apiVersion: apps/v1`,
    `    kind: Deployment`,
    `    name: ${name || "my-app"}`,
    `  minReplicas: 2`,
    `  maxReplicas: 20`,
    `  metrics:`,
    `    - type: Resource`,
    `      resource:`,
    `        name: cpu`,
    `        target:`,
    `          type: Utilization`,
    `          averageUtilization: 70`,
    `    - type: Resource`,
    `      resource:`,
    `        name: memory`,
    `        target:`,
    `          type: Utilization`,
    `          averageUtilization: 80`,
  ];
}

function generateIngress(form) {
  const { name, namespace, port } = form;
  return [
    `---`,
    `apiVersion: networking.k8s.io/v1`,
    `kind: Ingress`,
    `metadata:`,
    `  name: ${name || "my-app"}-ingress`,
    `  namespace: ${namespace || "default"}`,
    `  annotations:`,
    `    kubernetes.io/ingress.class: alb`,
    `    alb.ingress.kubernetes.io/scheme: internet-facing`,
    `    alb.ingress.kubernetes.io/target-type: ip`,
    `    alb.ingress.kubernetes.io/ssl-redirect: "443"`,
    `    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-1:...`,
    `spec:`,
    `  rules:`,
    `    - host: ${name || "my-app"}.internal.company.com`,
    `      http:`,
    `        paths:`,
    `          - path: /`,
    `            pathType: Prefix`,
    `            backend:`,
    `              service:`,
    `                name: ${name || "my-app"}`,
    `                port:`,
    `                  number: ${port || 8080}`,
  ];
}

// ─── Components ───────────────────────────────────────────────────────────────

function StatusBadge({ status }) {
  const map = { running: "badge-running", pending: "badge-pending", failed: "badge-failed", scaling: "badge-scaling", canary: "badge-canary" };
  const icon = { running: "●", pending: "◌", failed: "✕", scaling: "↑", canary: "◈" };
  return <span className={`badge ${map[status] || "badge-pending"}`}>{icon[status]} {status}</span>;
}

function Toast({ toasts }) {
  return (
    <div className="toast-container">
      {toasts.map(t => (
        <div key={t.id} className={`toast ${t.type}`}>
          <span className="toast-icon">{t.type === "success" ? "✅" : t.type === "error" ? "❌" : "ℹ️"}</span>
          <div className="toast-content">
            <div className="toast-title">{t.title}</div>
            <div className="toast-body">{t.body}</div>
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Dashboard View ───────────────────────────────────────────────────────────
function Dashboard({ onDeploy }) {
  return (
    <div>
      <div className="metric-grid">
        {[
          { label: "Total Workloads", value: "24", sub: "+2 this week", color: "var(--accent-cyan)", icon: "📦" },
          { label: "Pods Running", value: "147", sub: "3 pending", color: "var(--accent-green)", icon: "🟢" },
          { label: "CPU Utilization", value: "68%", sub: "↑ 4% from avg", color: "var(--accent-amber)", icon: "⚡" },
          { label: "Incidents Open", value: "2", sub: "1 critical", color: "var(--accent-red)", icon: "🔥" },
        ].map(m => (
          <div key={m.label} className="stat-card" style={{ "--accent-color": m.color }}>
            <div className="stat-label">{m.label}</div>
            <div className="stat-value">{m.value}</div>
            <div className="stat-sub">{m.sub}</div>
          </div>
        ))}
      </div>

      <div className="section-header">
        <span className="section-title">Active Workloads</span>
        <button className="btn btn-primary" onClick={onDeploy}>⊕ Deploy Workload</button>
      </div>

      <div className="form-card" style={{ marginBottom: 24 }}>
        <div className="env-tabs" style={{ padding: "14px 16px 0", marginBottom: 0 }}>
          {["all", "prod", "staging", "dev"].map(e => (
            <div key={e} className={`env-tab ${e} ${e === "all" ? "active" : ""}`}
              style={e === "all" ? { background: "var(--accent-cyan-dim)", color: "var(--accent-cyan)", borderColor: "rgba(0,212,255,0.3)" } : {}}>
              {e}
            </div>
          ))}
        </div>
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th><th>Namespace</th><th>Type</th><th>Replicas</th>
              <th>Status</th><th>CPU</th><th>Memory</th><th>HPA</th><th></th>
            </tr>
          </thead>
          <tbody>
            {WORKLOADS.map(w => (
              <tr key={w.name}>
                <td>
                  <span style={{ fontFamily: "var(--font-mono)", fontSize: 12, color: "var(--text-primary)", fontWeight: 600 }}>{w.name}</span>
                  <br /><span style={{ fontSize: 10, color: "var(--text-muted)", fontFamily: "var(--font-mono)" }}>{w.image}</span>
                </td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--accent-cyan)" }}>{w.namespace}</span></td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--text-muted)" }}>{w.type}</span></td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 12 }}>{w.replicas}</span></td>
                <td><StatusBadge status={w.status} /></td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>{w.cpu}</span></td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>{w.memory}</span></td>
                <td>{w.hpa ? <span className="badge" style={{ fontSize: 9, background: "var(--accent-purple-dim)", color: "var(--accent-purple)", borderColor: "rgba(163,113,247,0.25)" }}>HPA</span> : "—"}</td>
                <td>
                  <div style={{ display: "flex", gap: 4 }}>
                    <button className="btn btn-secondary btn-sm">Logs</button>
                    <button className="btn btn-secondary btn-sm">Scale</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
        <div>
          <div className="section-header"><span className="section-title">Namespace Resources</span></div>
          <div className="ns-grid">
            {NAMESPACES.map(ns => (
              <div key={ns.name} className="ns-card">
                <div className="ns-card-header">
                  <span className="ns-icon">🗂</span>
                  <span className="ns-name">{ns.name}</span>
                  <span className="status-chip healthy" style={{ fontSize: 9, padding: "2px 7px" }}><span className="status-dot"></span>{ns.team}</span>
                </div>
                <div className="ns-card-body">
                  {[["CPU", ns.cpuUsed, ns.cpuUsed > 80 ? "var(--accent-red)" : ns.cpuUsed > 65 ? "var(--accent-amber)" : "var(--accent-green)"],
                    ["Memory", ns.memUsed, ns.memUsed > 80 ? "var(--accent-red)" : ns.memUsed > 65 ? "var(--accent-amber)" : "var(--accent-cyan)"]
                  ].map(([label, pct, color]) => (
                    <div key={label} className="resource-bar">
                      <div className="resource-bar-label"><span>{label}</span><span style={{ color }}>{pct}%</span></div>
                      <div className="bar-track"><div className="bar-fill" style={{ width: `${pct}%`, background: color }}></div></div>
                    </div>
                  ))}
                  <div style={{ display: "flex", justifyContent: "space-between", marginTop: 8, fontSize: 10, color: "var(--text-muted)", fontFamily: "var(--font-mono)" }}>
                    <span>{ns.pods} pods</span><span>Quota: {ns.quota}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div>
          <div className="section-header"><span className="section-title">Recent Pipelines</span></div>
          <div className="pipeline-list">
            {PIPELINES.map(p => (
              <div key={p.name} className="pipeline-card">
                <span className="pipeline-status-icon">{p.status === "success" ? "✅" : p.status === "running" ? "🔄" : "❌"}</span>
                <div className="pipeline-info">
                  <div className="pipeline-name">{p.name}</div>
                  <div className="pipeline-meta">{p.branch} · {p.commit} · {p.ago}</div>
                  <div className="pipeline-stages">
                    {["Build","Test","Scan","Push","Deploy"].map((s, i) => (
                      <div key={s} style={{ flex: 1, textAlign: "center" }}>
                        <div className={`pipeline-stage ${p.stages[i]}`}></div>
                        <div className="pipeline-stage-label">{s}</div>
                      </div>
                    ))}
                  </div>
                </div>
                <div style={{ fontSize: 10, color: "var(--text-muted)", fontFamily: "var(--font-mono)", textAlign: "right", flexShrink: 0 }}>{p.time}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Deploy View ──────────────────────────────────────────────────────────────
function DeployView({ addToast }) {
  const [form, setForm] = useState({
    name: "", namespace: "payments", image: "", port: 8080,
    replicas: 2, cpu: 500, memory: 256, type: "Deployment",
    env: "prod", strategy: "RollingUpdate", ingress: true, hpa: true,
    serviceType: "ClusterIP", protocol: "HTTP"
  });
  const [activeTab, setActiveTab] = useState("deployment");
  const [aiQuery, setAiQuery] = useState("");
  const [aiResponse, setAiResponse] = useState("");
  const [aiLoading, setAiLoading] = useState(false);
  const [deploying, setDeploying] = useState(false);

  const set = (k, v) => setForm(f => ({ ...f, [k]: v }));

  const manifestLines = generateManifest(form);
  const hpaLines = generateHPA(form);
  const ingressLines = generateIngress(form);

  const tabLines = activeTab === "deployment" ? manifestLines : activeTab === "hpa" ? hpaLines : ingressLines;

  const handleAI = async () => {
    if (!aiQuery.trim()) return;
    setAiLoading(true);
    setAiResponse("");
    try {
      const response = await fetch("https://api.anthropic.com/v1/messages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          model: "claude-sonnet-4-20250514",
          max_tokens: 1000,
          system: `You are an expert Kubernetes and EKS platform engineer. Help developers configure their workloads. 
Be concise, precise, and practical. Give direct, actionable recommendations. 
When suggesting resource limits, be specific with numbers (e.g. 250m CPU, 512Mi memory).
Keep responses under 150 words. Format as plain text — no markdown, no asterisks.`,
          messages: [{ role: "user", content: `Workload context: name=${form.name || "my-app"}, type=${form.type}, env=${form.env}, replicas=${form.replicas}, cpu=${form.cpu}m, memory=${form.memory}Mi.\n\nQuestion: ${aiQuery}` }],
        }),
      });
      const data = await response.json();
      setAiResponse(data.content?.[0]?.text || "Unable to get recommendation.");
    } catch {
      setAiResponse("AI advisor unavailable. Please configure resources manually.");
    }
    setAiLoading(false);
  };

  const handleDeploy = async () => {
    if (!form.name || !form.image) {
      addToast({ type: "error", title: "Validation Error", body: "Workload name and image are required." });
      return;
    }
    setDeploying(true);
    await new Promise(r => setTimeout(r, 1800));
    setDeploying(false);
    addToast({ type: "success", title: "Workload Deployed", body: `${form.name} deployed to ${form.namespace}/${form.env} — 3 manifests applied.` });
  };

  return (
    <div>
      <div className="section-header">
        <span className="section-title">Deploy Workload</span>
        <div style={{ display: "flex", gap: 8 }}>
          <button className="btn btn-secondary">⬆ Import YAML</button>
          <button className="btn btn-primary" onClick={handleDeploy} disabled={deploying}>
            {deploying ? "⟳ Deploying..." : "🚀 Deploy to EKS"}
          </button>
        </div>
      </div>

      <div className="deploy-layout">
        {/* Left: Form */}
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          {/* Identity */}
          <div className="form-card">
            <div className="form-card-header">📋 Workload Identity</div>
            <div className="form-card-body">
              <div className="form-grid-2">
                <div className="form-row">
                  <label className="form-label">Workload Name <span>*</span></label>
                  <input className="form-input" placeholder="payments-api" value={form.name} onChange={e => set("name", e.target.value)} />
                  <div className="form-hint">DNS-safe, lowercase, hyphen-separated</div>
                </div>
                <div className="form-row">
                  <label className="form-label">Namespace / Tenant <span>*</span></label>
                  <select className="form-select" value={form.namespace} onChange={e => set("namespace", e.target.value)}>
                    <option value="payments">payments</option>
                    <option value="auth">auth</option>
                    <option value="analytics">analytics</option>
                    <option value="ml">ml</option>
                    <option value="web">web</option>
                  </select>
                </div>
              </div>
              <div className="form-grid-2">
                <div className="form-row">
                  <label className="form-label">Workload Type</label>
                  <div className="toggle-group">
                    {["Deployment","StatefulSet","CronJob"].map(t => (
                      <div key={t} className={`toggle-opt ${form.type === t ? "active" : ""}`} onClick={() => set("type", t)}>{t}</div>
                    ))}
                  </div>
                </div>
                <div className="form-row">
                  <label className="form-label">Environment</label>
                  <div className="toggle-group">
                    {["dev","staging","prod"].map(e => (
                      <div key={e} className={`toggle-opt ${form.env === e ? "active" : ""}`} onClick={() => set("env", e)}>{e}</div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Container */}
          <div className="form-card">
            <div className="form-card-header">🐳 Container Config</div>
            <div className="form-card-body">
              <div className="form-row">
                <label className="form-label">Container Image <span>*</span></label>
                <input className="form-input" placeholder="123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.0.0" value={form.image} onChange={e => set("image", e.target.value)} />
                <div className="form-hint">ECR URI recommended — latest tag discouraged in prod</div>
              </div>
              <div className="form-grid-3">
                <div className="form-row">
                  <label className="form-label">Container Port</label>
                  <input className="form-input" type="number" value={form.port} onChange={e => set("port", +e.target.value)} />
                </div>
                <div className="form-row">
                  <label className="form-label">Protocol</label>
                  <select className="form-select" value={form.protocol} onChange={e => set("protocol", e.target.value)}>
                    <option>HTTP</option><option>gRPC</option><option>TCP</option>
                  </select>
                </div>
                <div className="form-row">
                  <label className="form-label">Service Type</label>
                  <select className="form-select" value={form.serviceType} onChange={e => set("serviceType", e.target.value)}>
                    <option>ClusterIP</option><option>NodePort</option><option>LoadBalancer</option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          {/* Resources */}
          <div className="form-card">
            <div className="form-card-header">⚡ Resources & Scaling</div>
            <div className="form-card-body">
              <div className="form-row">
                <label className="form-label">Replicas — <span style={{ color: "var(--accent-cyan)", fontFamily: "var(--font-mono)" }}>{form.replicas}</span></label>
                <div className="slider-container">
                  <input type="range" className="slider" min={1} max={20} value={form.replicas} onChange={e => set("replicas", +e.target.value)} />
                  <span className="slider-val">{form.replicas}</span>
                </div>
              </div>
              <div className="form-row">
                <label className="form-label">CPU Limit — <span style={{ color: "var(--accent-amber)", fontFamily: "var(--font-mono)" }}>{form.cpu}m</span></label>
                <div className="slider-container">
                  <input type="range" className="slider" min={50} max={4000} step={50} value={form.cpu} onChange={e => set("cpu", +e.target.value)} />
                  <span className="slider-val" style={{ color: "var(--accent-amber)" }}>{form.cpu}m</span>
                </div>
              </div>
              <div className="form-row">
                <label className="form-label">Memory Limit — <span style={{ color: "var(--accent-purple)", fontFamily: "var(--font-mono)" }}>{form.memory}Mi</span></label>
                <div className="slider-container">
                  <input type="range" className="slider" min={64} max={8192} step={64} value={form.memory} onChange={e => set("memory", +e.target.value)} />
                  <span className="slider-val" style={{ color: "var(--accent-purple)" }}>{form.memory}Mi</span>
                </div>
              </div>
              <div className="form-grid-2">
                <div className="form-row">
                  <label className="form-label">Rollout Strategy</label>
                  <select className="form-select" value={form.strategy} onChange={e => set("strategy", e.target.value)}>
                    <option>RollingUpdate</option><option>Recreate</option><option>Canary (Argo)</option><option>Blue/Green</option>
                  </select>
                </div>
                <div className="form-row">
                  <label className="form-label">Features</label>
                  <div style={{ display: "flex", gap: 8, paddingTop: 4 }}>
                    {[["HPA", "hpa"], ["Ingress", "ingress"]].map(([label, key]) => (
                      <div key={key} onClick={() => set(key, !form[key])}
                        style={{ padding: "5px 12px", borderRadius: "var(--radius)", border: "1px solid", cursor: "pointer", fontSize: 11, fontWeight: 500, transition: "all 0.15s",
                          background: form[key] ? "var(--accent-cyan-dim)" : "var(--bg-raised)",
                          color: form[key] ? "var(--accent-cyan)" : "var(--text-muted)",
                          borderColor: form[key] ? "rgba(0,212,255,0.3)" : "var(--border)" }}>
                        {label}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* AI Advisor */}
          <div className="ai-panel">
            <div className="ai-header">
              <span className="ai-spark">✦</span>
              <div>
                <div className="ai-title">Platform AI Advisor</div>
                <div className="ai-subtitle">Powered by Claude · Ask anything about this deployment</div>
              </div>
            </div>
            <div className="ai-body">
              <div className="ai-input-row">
                <input className="ai-input" placeholder="e.g. What resource limits should I set for a Node.js API serving 500 rps?" value={aiQuery} onChange={e => setAiQuery(e.target.value)} onKeyDown={e => e.key === "Enter" && handleAI()} />
                <button className="btn btn-primary" onClick={handleAI} disabled={aiLoading}>{aiLoading ? "..." : "Ask"}</button>
              </div>
              {(aiResponse || aiLoading) && (
                <div className={`ai-response ${aiLoading ? "ai-typing" : ""}`}>
                  {aiResponse || (aiLoading ? "Analyzing workload context" : "")}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right: Manifest preview */}
        <div className="manifest-panel">
          <div className="manifest-header">
            <div className="manifest-tabs">
              {[["deployment", "Deployment"], ["hpa", "HPA"], ["ingress", "Ingress"]].map(([id, label]) => (
                <div key={id} className={`manifest-tab ${activeTab === id ? "active" : ""}`} onClick={() => setActiveTab(id)}>{label}</div>
              ))}
            </div>
            <button className="btn btn-secondary btn-sm">⎘ Copy</button>
            <button className="btn btn-secondary btn-sm">⬇ Export</button>
          </div>
          <div className="manifest-body">
            <div className="yaml-code">
              {tabLines.map((line, i) => <YamlLine key={i} line={line} />)}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Catalog View ─────────────────────────────────────────────────────────────
function CatalogView({ onSelect }) {
  const tagColors = { deployment: "var(--accent-cyan)", hpa: "var(--accent-green)", ingress: "var(--accent-amber)", keda: "var(--accent-purple)", gpu: "var(--accent-red)", mtls: "var(--accent-green)", monitoring: "var(--accent-cyan)" };
  return (
    <div>
      <div className="section-header">
        <span className="section-title">Service Catalog</span>
        <input className="form-input" style={{ width: 240 }} placeholder="🔍 Search templates..." />
      </div>
      <div style={{ marginBottom: 20, fontSize: 12, color: "var(--text-muted)", lineHeight: 1.6 }}>
        Pre-hardened workload templates that encode platform best practices — security contexts, resource policies, observability hooks, and network policies — all pre-configured.
      </div>
      <div className="catalog-grid">
        {CATALOG_ITEMS.map(item => (
          <div key={item.name} className="catalog-card" onClick={() => onSelect(item)}>
            <span className="catalog-card-icon">{item.icon}</span>
            <div className="catalog-card-name" style={{ color: item.color }}>{item.name}</div>
            <div className="catalog-card-desc">{item.desc}</div>
            <div className="catalog-card-tags">
              {item.tags.map(tag => (
                <span key={tag} className="catalog-tag" style={{ background: `${tagColors[tag] || "var(--accent-cyan)"}22`, color: tagColors[tag] || "var(--accent-cyan)" }}>{tag}</span>
              ))}
            </div>
            <div style={{ marginTop: 14, paddingTop: 12, borderTop: "1px solid var(--border)", display: "flex", justifyContent: "flex-end" }}>
              <button className="btn btn-secondary btn-sm">Use Template →</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Observe View ─────────────────────────────────────────────────────────────
function ObserveView() {
  return (
    <div>
      <div className="section-header"><span className="section-title">Observability</span>
        <div style={{ display: "flex", gap: 8 }}>
          <select className="form-select" style={{ width: 130, padding: "5px 10px", fontSize: 11 }}><option>All namespaces</option></select>
          <select className="form-select" style={{ width: 110, padding: "5px 10px", fontSize: 11 }}><option>Last 15m</option></select>
        </div>
      </div>
      <div className="metric-grid" style={{ marginBottom: 20 }}>
        {[
          { label: "p99 Latency", value: "142ms", color: "var(--accent-cyan)" },
          { label: "Error Rate", value: "0.4%", color: "var(--accent-green)" },
          { label: "Throughput", value: "4.2k rps", color: "var(--accent-amber)" },
          { label: "Pod Restarts", value: "7", color: "var(--accent-red)" },
        ].map(m => (
          <div key={m.label} className="stat-card" style={{ "--accent-color": m.color }}>
            <div className="stat-label">{m.label}</div>
            <div className="stat-value" style={{ fontSize: 22 }}>{m.value}</div>
          </div>
        ))}
      </div>

      {/* Sparkline bars visualization */}
      <div className="form-card" style={{ marginBottom: 20 }}>
        <div className="form-card-header">📈 Request Rate (rps) — last 15 minutes</div>
        <div className="form-card-body">
          <div style={{ display: "flex", alignItems: "flex-end", gap: 3, height: 80 }}>
            {[42,38,55,61,48,72,88,95,84,76,91,103,98,87,94,110,107,115,122,108,96,112,118,125,119,128,134,126,121,135].map((v, i) => (
              <div key={i} style={{
                flex: 1, background: i === 29 ? "var(--accent-cyan)" : "rgba(0,212,255,0.25)",
                borderRadius: "2px 2px 0 0", height: `${(v / 140) * 100}%`,
                transition: "height 0.3s ease"
              }}></div>
            ))}
          </div>
          <div style={{ display: "flex", justifyContent: "space-between", fontSize: 9, color: "var(--text-muted)", fontFamily: "var(--font-mono)", marginTop: 6 }}>
            <span>10:28</span><span>10:35</span><span>10:42</span>
          </div>
        </div>
      </div>

      <div className="obs-grid">
        <div className="log-panel">
          <div className="log-header">
            <span className="log-title">Live Logs</span>
            <div className="status-chip healthy"><span className="status-dot"></span>streaming</div>
            <button className="btn btn-secondary btn-sm">⬇ Download</button>
          </div>
          <div className="log-body">
            {LOGS.map((l, i) => (
              <div key={i} className="log-line">
                <span className="log-ts">{l.ts}</span>
                <span className={`log-level-${l.level.toLowerCase()}`}>{l.level}</span>
                <span className="log-msg">{l.msg}</span>
              </div>
            ))}
          </div>
        </div>

        <div>
          <div className="form-card">
            <div className="form-card-header">🔔 Active Alerts</div>
            <div className="form-card-body" style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {[
                { sev: "critical", name: "WorkerCrashLoopBackOff", ns: "jobs", since: "18m" },
                { sev: "warning", name: "HighMemoryUsage", ns: "ml", since: "4m" },
                { sev: "info", name: "CanaryErrorRateElevated", ns: "checkout", since: "2m" },
              ].map(a => (
                <div key={a.name} style={{ padding: "10px 12px", background: "var(--bg-raised)", border: "1px solid var(--border)", borderRadius: "var(--radius)", borderLeft: `3px solid ${a.sev === "critical" ? "var(--accent-red)" : a.sev === "warning" ? "var(--accent-amber)" : "var(--accent-cyan)"}` }}>
                  <div style={{ fontSize: 12, fontWeight: 600, marginBottom: 2 }}>{a.name}</div>
                  <div style={{ fontSize: 10, color: "var(--text-muted)", fontFamily: "var(--font-mono)" }}>{a.ns} · {a.since} ago</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Settings / Tenants ───────────────────────────────────────────────────────
function TenantsView() {
  return (
    <div>
      <div className="section-header"><span className="section-title">Tenant Management</span>
        <button className="btn btn-primary">⊕ New Tenant</button>
      </div>
      <div className="form-card">
        <table className="data-table">
          <thead><tr><th>Tenant</th><th>Namespace</th><th>Team</th><th>CPU Quota</th><th>Mem Quota</th><th>Workloads</th><th>Network Policy</th><th>Status</th><th></th></tr></thead>
          <tbody>
            {[
              { name: "Payments Platform", ns: "payments", team: "Platform Eng", cpu: "8", mem: "16Gi", wl: 6, net: "Strict", status: "active" },
              { name: "Auth & Identity", ns: "auth", team: "Security", cpu: "4", mem: "8Gi", wl: 3, net: "Strict", status: "active" },
              { name: "Data & Analytics", ns: "analytics", team: "Data Eng", cpu: "16", mem: "32Gi", wl: 8, net: "Permissive", status: "active" },
              { name: "ML Operations", ns: "ml", team: "ML Platform", cpu: "32", mem: "64Gi", wl: 4, net: "Strict", status: "active" },
              { name: "Web Frontend", ns: "web", team: "Product", cpu: "4", mem: "8Gi", wl: 2, net: "Permissive", status: "active" },
            ].map(t => (
              <tr key={t.ns}>
                <td style={{ fontWeight: 600, fontSize: 12 }}>{t.name}</td>
                <td><span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--accent-cyan)" }}>{t.ns}</span></td>
                <td style={{ fontSize: 12 }}>{t.team}</td>
                <td style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>{t.cpu} cores</td>
                <td style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>{t.mem}</td>
                <td style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>{t.wl}</td>
                <td><span className="badge" style={{ fontSize: 9, background: t.net === "Strict" ? "var(--accent-green-dim)" : "var(--accent-amber-dim)", color: t.net === "Strict" ? "var(--accent-green)" : "var(--accent-amber)", borderColor: "transparent" }}>{t.net}</span></td>
                <td><StatusBadge status="running" /></td>
                <td><button className="btn btn-secondary btn-sm">Manage</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── Root App ─────────────────────────────────────────────────────────────────
export default function IDPApp() {
  const [view, setView] = useState("dashboard");
  const [toasts, setToasts] = useState([]);

  const addToast = (toast) => {
    const id = Date.now();
    setToasts(prev => [...prev, { ...toast, id }]);
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 5000);
  };

  const NAV = [
    { id: "dashboard", label: "Overview", icon: "⬡", section: "main" },
    { id: "deploy", label: "Deploy", icon: "🚀", section: "main" },
    { id: "catalog", label: "Service Catalog", icon: "📦", section: "main" },
    { id: "observe", label: "Observability", icon: "📡", section: "ops" },
    { id: "tenants", label: "Tenants", icon: "🏢", section: "ops" },
    { id: "settings", label: "Settings", icon: "⚙", section: "ops", badge: null },
  ];

  const TITLES = { dashboard: "Platform Overview", deploy: "Deploy Workload", catalog: "Service Catalog", observe: "Observability", tenants: "Tenant Management", settings: "Settings" };

  return (
    <>
      <style>{CSS}</style>
      <div className="idp-root">
        {/* Sidebar */}
        <div className="sidebar">
          <div className="sidebar-logo">
            <div className="logo-mark">K8</div>
            <div className="logo-text">
              <div className="logo-title">EKS Platform</div>
              <div className="logo-sub">IDP v2.4.1</div>
            </div>
          </div>
          <div className="cluster-selector">
            <span className="cluster-dot"></span>
            <span className="cluster-name">prod-eks-us-east-1</span>
            <span className="cluster-region">1.29</span>
          </div>

          {["main", "ops"].map(section => (
            <div key={section} className="nav-section">
              <div className="nav-label">{section === "main" ? "Workloads" : "Platform"}</div>
              {NAV.filter(n => n.section === section).map(n => (
                <div key={n.id} className={`nav-item ${view === n.id ? "active" : ""}`} onClick={() => setView(n.id)}>
                  <span className="nav-icon">{n.icon}</span>
                  <span>{n.label}</span>
                  {n.badge && <span className="nav-badge">{n.badge}</span>}
                </div>
              ))}
            </div>
          ))}

          <div className="sidebar-footer">
            <div className="tenant-pill">
              <div className="tenant-avatar">PA</div>
              <div className="tenant-info">
                <div className="tenant-name">Payments Team</div>
                <div className="tenant-ns">ns: payments</div>
              </div>
            </div>
          </div>
        </div>

        {/* Main area */}
        <div className="main">
          <div className="topbar">
            <div style={{ display: "flex", flexDirection: "column" }}>
              <span className="topbar-title">{TITLES[view]}</span>
              <span className="topbar-breadcrumb">eks-idp / {view}</span>
            </div>
            <div className="topbar-spacer"></div>
            <div className="status-chip healthy"><span className="status-dot"></span>Cluster Healthy</div>
            <div className="status-chip healthy" style={{ fontFamily: "var(--font-mono)", fontSize: 10 }}>147 / 200 pods</div>
            <div className="status-chip warning"><span className="status-dot"></span>2 alerts</div>
          </div>

          <div className="content">
            {view === "dashboard" && <Dashboard onDeploy={() => setView("deploy")} />}
            {view === "deploy" && <DeployView addToast={addToast} />}
            {view === "catalog" && <CatalogView onSelect={() => setView("deploy")} />}
            {view === "observe" && <ObserveView />}
            {view === "tenants" && <TenantsView />}
            {view === "settings" && (
              <div style={{ padding: 40, textAlign: "center", color: "var(--text-muted)" }}>
                <div style={{ fontSize: 48, marginBottom: 12 }}>⚙️</div>
                <div style={{ fontSize: 14 }}>Platform Settings</div>
                <div style={{ fontSize: 12, marginTop: 8 }}>RBAC, ArgoCD, Vault, Service Mesh configuration</div>
              </div>
            )}
          </div>
        </div>

        <Toast toasts={toasts} />
      </div>
    </>
  );
}
