# benchmark-runner

Automated runner for this repository's skill-loading checks. It owns what
makes these runs a benchmark (which skills to install, which prompts to
send, and how observed facts map to verdicts), while the invocation and
observation machinery lives in
[skillxp](https://github.com/agent-ecosystem/skillxp): install skills in a
fixture, invoke the harness headlessly, locate and parse the transcript,
trace canary phrases.

Check IDs match [checks.md](../checks.md) (check list
version 0.2); canary phrases match the
[canary index](../benchmark-skills/README.md#canary-phrase-index).
Automated findings complement, not replace, the manual results in
[platform-findings/](../platform-findings/):
the runner covers the three harnesses with headless modes (Antigravity
CLI, Claude Code, Codex CLI), and **headless behavior may differ from
interactive use**. Manual testing remains the only path for the other
20+ platforms.

## Usage

Requires Go and at least one supported harness installed and
authenticated.

```bash
cd benchmark-runner

# List implemented checks
go run . -list

# Everything, all installed harnesses
go run . -out results

# One harness, one check
go run . -harness claude-code -check discovery-reading-depth -out results

# Repeat a model-dependent check and report verdict rates
go run . -harness codex -check reactivation-deduplication -runs 5 -out results

# Run against isolated homes so your own installed skills can't
# contaminate listings (see skillxp's README for one-time auth setup)
go run . -sandbox -out results

# Generate per-platform markdown reports from findings (latest finding
# wins when the same harness+check appears in several dirs)
go run . -report results/batch2-2026-08-01,results/batch3-2026-08-01 -out reports
```

Output: `results/<harness>/<check>/finding.json` plus the archived
transcript(s) its evidence cites by line number.

## Reading a finding

- `verdict`: the check's conclusion (e.g. `metadata-only`, `body-only`).
- `vehicle`: how content reached the model, either `harness-push` (platform
  injected it) or `model-pull` (model fetched it with a tool). The two
  vehicles produce different author-facing behavior (frontmatter
  visibility, wrapping) even under the same verdict.
- `confidence`: `transcript-direct` when the transcript records the
  evidence; `behavioral-inference` when it must be inferred (Antigravity
  records no injected context).
- `evidence`: event indices and 1-based line numbers into the archived
  transcript.
- `runs` / `verdict_counts` / `run_errors`: present on repeated checks
  (`-runs` > 1): `verdict` holds the modal verdict, `verdict_counts` the
  full distribution, and per-run findings sit in `run-NN/` next to their
  transcripts. Verdict variation across runs proves model-level behavior;
  consistency suggests (but does not prove) platform-level behavior.

## Implemented checks

All checks from checks.md are automated except two that stay
manual by nature: `context-compaction-protection` (needs a long
conversation) and `trust-gating-behavior` (needs an interactive trust
prompt).

| Check | Status |
|---|---|
| `discovery-reading-depth` | Automated |
| `activation-loading-scope` | Automated |
| `eager-link-resolution` | Automated |
| `recognized-directory-set` | Automated |
| `directory-naming-divergence` | Automated |
| `unrecognized-directory-handling` | Automated |
| `resource-enumeration-behavior` | Automated |
| `path-resolution-base` | Automated |
| `cross-skill-resource-shadowing` | Automated |
| `path-traversal-boundary` | Automated (installs a sibling skill as the traversal target) |
| `discovery-listing-fields` | Automated (passive verbatim-catalog turn) |
| `frontmatter-handling` | Automated |
| `metadata-value-edge-cases` | Automated |
| `content-wrapping-format` | Automated |
| `reactivation-deduplication` | Automated (two-turn session) |
| `reactivation-freshness` | Automated (two-turn session, edits SKILL.md between activations) |
| `compatibility-field-behavior` | Automated |
| `nested-skill-discovery` | Automated (passive listing turn) |
| `resource-nesting-depth` | Automated |
| `name-directory-mismatch` | Automated (three turns: listing, then activation by each name) |
| `recursive-root-discovery` | Automated (installs a stray SKILL.md outside the skills root) |
| `cross-skill-invocation` | Automated |
| `invocation-depth-limit` | Automated |
| `circular-invocation-handling` | Automated |
| `invocation-language-sensitivity` | Automated (Japanese prompt; use `-runs` for rates) |
| `informal-dependency-resolution` | Automated |
| `missing-dependency-behavior` | Automated |
| `nonstandard-dependency-fields` | Automated (installs the named dependencies so the platform could resolve them) |
| `cross-scope-dependency` | Automated (two sessions: dependency at user scope, then absent; requires `-sandbox`) |
| `cross-client-directory-interop` | Automated (skill installed only at `.agents/skills/` via project overlay) |
| `malformed-yaml-tolerance` | Automated |
| `missing-description-handling` | Automated |
| `name-collision-precedence` | Automated (same name at project and user scope; requires `-sandbox`) |
| `context-compaction-protection` | Manual |
| `trust-gating-behavior` | Manual |

## Note on the module

This directory is a Go module so the repository stays a content-first
project; the module depends on the published
[skillxp](https://github.com/agent-ecosystem/skillxp) release (v0.1.1 as
of this writing). Unlike the repository's CC-BY-4.0 content, code
under this directory is intended to be MIT-licensed to match the rest of
the tooling; see the repository maintainers if that matters to your use.
