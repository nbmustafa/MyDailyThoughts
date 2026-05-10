# Renovate Bot — Self-Hosted Setup Guide

## Repository / folder layout

```
.                               ← A dedicated "renovate-config" repo
├── .github/
│   └── workflows/
│       └── renovate.yml        ← GitHub Actions workflow (runs the bot)
├── renovate/
│   ├── config.js               ← Self-hosted runner config (which repos to scan)
│   └── renovate.json           ← Per-repo Renovate rules (copy into EACH target repo)
└── Chart.yaml.examples         ← Reference Chart.yaml snippets (do NOT commit as-is)
```

> **Two-repo model** — `config.js` lives in this *runner* repo. `renovate.json` is
> committed to the **root of each target repo** (`fluent-bit`, `grafana`,
> `prometheus`, `argocd`).

---

## 1. Create the GitHub Personal Access Token (PAT)

### Required scopes

| Scope | Reason |
|---|---|
| `repo` | Read code, open PRs, post comments |
| `read:org` | Enumerate repos under the `nbmustafa` org |
| `workflow` | Allow Renovate to update `.github/workflows/` files |

### Steps

1. Go to **GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens**.
2. Set *Resource owner* to `nbmustafa`.
3. Under *Repository access*, choose **Only select repositories** and tick the four target repos.
4. Grant the scopes in the table above.
5. Click **Generate token** and copy the value immediately — it is shown only once.

---

## 2. Store the PAT as a GitHub Actions secret

> GitHub Actions automatically **masks** every registered secret value in
> all log output — even if your code accidentally prints it to stdout.
> The raw token value is never visible in workflow logs, artifacts, or
> the GitHub UI after initial creation.

### In the *runner* repository (where the workflow lives)

```
Settings → Secrets and variables → Actions → New repository secret

Name:  RENOVATE_TOKEN
Value: <paste your PAT here>
```

If you host multiple runner repos inside an organisation, you can instead
store the secret at the **organisation** level and grant access to specific
repositories — avoiding duplication.

### Environment reference in the workflow

```yaml
env:
  RENOVATE_TOKEN: ${{ secrets.RENOVATE_TOKEN }}
```

`${{ secrets.RENOVATE_TOKEN }}` is resolved at runtime by the Actions runner
and injected as an environment variable. The literal string `RENOVATE_TOKEN`
is stored in `config.js` as `process.env.RENOVATE_TOKEN` — again, never
hard-coded.

---

## 3. Copy renovate.json into each target repo

```bash
for repo in fluent-bit grafana prometheus argocd; do
  gh repo clone nbmustafa/$repo /tmp/$repo
  cp renovate/renovate.json /tmp/$repo/renovate.json
  cd /tmp/$repo
  git checkout -b add-renovate-config
  git add renovate.json
  git commit -m "chore: add Renovate bot configuration"
  gh pr create --title "chore: add Renovate bot configuration" \
               --body  "Adds renovate.json so the self-hosted Renovate bot can track upstream Helm chart releases." \
               --base  main
  cd -
done
```

---

## 4. Update Chart.yaml in each target repo

Follow the annotated `Chart.yaml.examples` file. The critical parts are:

1. Add `dependencies[]` with the correct upstream `repository` URL.
2. Add the `# renovate: datasource=helm registryUrl=... depName=...` comment
   on the line **immediately above** the dependency entry.
3. Run `helm dependency update` to generate / refresh `Chart.lock`.

Commit `Chart.yaml` **and** `Chart.lock` together. Renovate will update both
files in its PRs.

---

## 5. Trigger a manual dry-run to validate

```
GitHub → Actions → 🤖 Renovate Bot → Run workflow
  dry_run:   true
  log_level: debug
```

Check the workflow logs. You should see Renovate discover the Chart.yaml
dependencies and report which PRs *would* be opened without actually opening
them.

---

## 6. Schedule

The workflow runs on `cron: "0 */6 * * *"` — 00:00, 06:00, 12:00, 18:00 UTC.
Change the cron expression in `.github/workflows/renovate.yml` to suit your
team's cadence.

---

## Chart.yaml ↔ values.yaml synchronisation

| File | Field | Updated by |
|---|---|---|
| `Chart.yaml` | `dependencies[].version` | Renovate helmv3 manager ✅ (automatic) |
| `Chart.yaml` | `appVersion` | Renovate regexManager ✅ (if configured) or manual |
| `values.yaml` | `image.tag` | Renovate regexManager ✅ (if configured) or manual |
| `Chart.lock` | `dependencies[].version` / `digest` | Renovate post-upgrade task ✅ (automatic) |

See the `regexManagers` comment block in `renovate.json` to enable full
automatic synchronisation of `appVersion` and `image.tag` in the same PR.

---

## Upgrading the pinned Renovate version

The workflow pins `renovate-version: "37.440.0"`. To bump it:

1. Check the [Renovate releases page](https://github.com/renovatebot/renovate/releases).
2. Update the version string in `renovate.yml`.
3. Open a PR, let CI pass, then merge.

You can also let Renovate manage its own version by adding this to `renovate.json`:

```json
"extends": ["helpers:pinGitHubActionDigests"]
```
