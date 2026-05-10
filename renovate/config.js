/**
 * Renovate Self-Hosted Runner Configuration
 * -----------------------------------------
 * This file is loaded by the Renovate CLI at startup.
 * It controls WHICH repositories are managed and global runner behaviour.
 *
 * Docs: https://docs.renovatebot.com/self-hosted-configuration/
 */

module.exports = {
  // ── Platform ───────────────────────────────────────────────────────────────
  platform: "github",

  /**
   * GitHub endpoint – change only if you run GitHub Enterprise Server.
   * For github.com this is always the default and can be omitted, but
   * being explicit makes intent clear and eases future migrations.
   */
  endpoint: "https://api.github.com/",

  // ── Authentication ─────────────────────────────────────────────────────────
  /**
   * RENOVATE_TOKEN is injected at runtime by the GitHub Actions workflow
   * (see renovate.yml).  Never hard-code a token here.
   */
  token: process.env.RENOVATE_TOKEN,

  /**
   * Used by Renovate to author commits and PRs.
   * Must match the identity of the PAT owner, or a dedicated bot account.
   */
  gitAuthor: "Renovate Bot <renovate-bot@users.noreply.github.com>",

  // ── Target repositories ────────────────────────────────────────────────────
  /**
   * Explicit allow-list: only these four repos will ever be touched.
   * Format: "<org-or-user>/<repo>"
   */
  repositories: [
    "nbmustafa/fluent-bit",
    "nbmustafa/grafana",
    "nbmustafa/prometheus",
    "nbmustafa/argocd",
  ],

  // ── Onboarding ─────────────────────────────────────────────────────────────
  /**
   * Each repo already carries a renovate.json (provided separately).
   * Disable the first-run onboarding PR so Renovate starts working
   * immediately without waiting for a merge.
   */
  onboarding: false,

  /**
   * Still require a renovate.json to be present in the repo root.
   * This acts as a safety gate: Renovate will not touch a repo unless
   * an explicit config exists, preventing accidental changes.
   */
  requireConfig: "required",

  // ── Runtime behaviour ──────────────────────────────────────────────────────
  /**
   * Print verbose logs when the LOG_LEVEL env-var is set to "debug".
   * The workflow sets LOG_LEVEL=debug on workflow_dispatch so you can
   * trigger a manual debug run without changing this file.
   */
  logLevel: process.env.LOG_LEVEL || "info",

  /**
   * Never create more than 10 open PRs per repository at one time.
   * Keeps the PR queue manageable and avoids GitHub rate-limit spikes.
   */
  prConcurrentLimit: 10,

  /**
   * Wait 3 days before raising a PR for a brand-new release.
   * Filters out yanked / immediately-patched releases.
   */
  stabilityDays: 3,

  // ── Persistence (optional but recommended for large orgs) ──────────────────
  /**
   * Cache Renovate's internal state between runs so it does not
   * re-process every repository from scratch on every schedule tick.
   * The GitHub Actions workflow mounts this path via the "cache" action.
   */
  cacheDir: "/tmp/renovate/cache",
  baseDir:  "/tmp/renovate/base",

  // ── Dry-run override ───────────────────────────────────────────────────────
  /**
   * Set RENOVATE_DRY_RUN=true in the workflow environment to validate
   * config changes without actually opening PRs.
   */
  dryRun: process.env.RENOVATE_DRY_RUN === "true" ? "full" : null,
};
