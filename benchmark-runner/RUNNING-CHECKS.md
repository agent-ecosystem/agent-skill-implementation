# Running the checks against the harnesses

This is the maintainer procedure for re-running the automated checks after
a harness release, a dependency bump, or a check list change, and for
publishing what changed. It is written for a person; the project skill at
[`.claude/skills/run-checks/`](../.claude/skills/run-checks/SKILL.md)
points an agent at the same steps. Keep the two in step: the facts live
here, the skill only points at them.

A run has five stages. Each is cheap to redo, and the tooling is designed
so that regenerating from recorded findings is byte-for-byte repeatable.

1. [Check the tooling and releases](#1-check-the-tooling-and-releases)
2. [Run the checks](#2-run-the-checks)
3. [Validate the findings](#3-validate-the-findings)
4. [Record and compare](#4-record-and-compare)
5. [Publish](#5-publish)

## 1. Check the tooling and releases

The runner depends on three published Go modules that encode harness
lore: [agentsummons](https://github.com/agent-ecosystem/agentsummons)
(flag surfaces), [agentminutes](https://github.com/agent-ecosystem/agentminutes)
(transcript parsing), and [skillxp](https://github.com/agent-ecosystem/skillxp)
(skill install, invoke, locate). When a harness ships a release that
changes its transcript format or flags, the old parser fails or, worse,
parses partially. Bump first, then run.

```bash
cd benchmark-runner
go list -m -versions github.com/agent-ecosystem/skillxp   # and the other two
go get github.com/agent-ecosystem/agentsummons@vX github.com/agent-ecosystem/agentminutes@vY github.com/agent-ecosystem/skillxp@vZ
go mod tidy && go build ./... && go vet ./... && go test ./...
```

Then confirm the bump changed no verdict logic: regenerate the reports
from the existing results directories into a scratch location and diff
against `reports/`. The output should be identical.

Note the installed harness versions (`claude --version`, `codex --version`,
`agy --version`, `copilot --version`); they will appear in the findings and
the changelog entry.

## 2. Run the checks

Pick a results directory name for this run, dated:
`results/checks-YYYY-MM-DD`. Every harness writes into the same directory,
so the comparison and ledger steps see one run.

Run the harnesses **sandboxed** (`-sandbox`): isolated homes keep your own
installed skills out of the listings and are required by the two
user-scope checks (`cross-scope-dependency`, `name-collision-precedence`).
Each harness can run as its own process; fixtures are per-run temporary
directories, so parallel processes do not interfere. A full run of all
checks on one harness takes roughly 25 minutes.

```bash
cd benchmark-runner
go run . -sandbox -harness antigravity -out results/checks-YYYY-MM-DD &
go run . -sandbox -harness codex       -out results/checks-YYYY-MM-DD &
go run . -sandbox -harness copilot     -out results/checks-YYYY-MM-DD &
```

Sandbox authentication is a one-time setup per harness; see skillxp's
[sandboxing guide](https://skillxp.dev/docs/sandboxing/). In short: codex
and copilot need nothing extra on macOS, antigravity needs a seed home
under `~/.skillxp/seeds/antigravity/`, and Claude Code needs the token
below.

### Claude Code: run it yourself, in your own shell

Claude Code's sandbox cannot reach the macOS keychain, so it needs a
subscription OAuth token exported in the environment. Two consequences:

- An agent session on this machine does not have that token, and must
  not mint one.
- A Claude Code agent session and a headless `claude` run on the same
  machine share one subscription. Running the checks unsandboxed from
  inside an agent session once consumed the session limit mid-run and
  produced thirty findings scored against a synthetic "You've hit your
  session limit" message. See [Validate the findings](#3-validate-the-findings).

So the maintainer runs the Claude Code portion from a terminal outside
any agent session:

```bash
# once per token lifetime; opens a browser, bills to your subscription
claude setup-token
export CLAUDE_CODE_OAUTH_TOKEN="<paste the token>"

# or keep it in 1Password and let op resolve it per command:
# export CLAUDE_CODE_OAUTH_TOKEN="op://<vault>/<item>/credential"
# op run -- go run . -sandbox -harness claude-code -out results/checks-YYYY-MM-DD

cd benchmark-runner
go run . -sandbox -harness claude-code -out results/checks-YYYY-MM-DD
```

Use the same `-out` directory as the other three harnesses. When an agent
is driving the rest of the run, it should hand you this command with the
directory filled in, and wait for your results before comparing.

### Reruns

Model-level behavior varies between runs. When a verdict looks like it
could be either, or a check comes back `inconclusive`, rerun it with
repetitions and the finding records the verdict distribution:

```bash
go run . -sandbox -harness copilot -check reactivation-deduplication -runs 3 -out results/checks-YYYY-MM-DD
```

A consistent 3/3 suggests platform-level behavior; variation proves
model-level.

## 3. Validate the findings

Before trusting a run, look for the failure modes that have produced
plausible-looking garbage:

- **Synthetic error turns.** A rate limit or auth failure makes the
  harness return an assistant message with no model behind it; the runner
  currently scores it as `inconclusive` or even `observed`. Grep every
  transcript for `<synthetic>` and `isApiErrorMessage`, and look for
  findings whose `started` to `ended` span is under about four seconds
  (listing-only checks are legitimately quick; activation checks are not).
- **Missing model names.** Every observed finding should carry a
  `model`. An empty one means the transcript was not fully parsed.
- **Status errors.** `runner-error` findings name their cause in `notes`.
  The two user-scope checks error without `-sandbox` by design.

A contaminated run should be moved out of `results/` (not deleted; keep it
for diagnosis) and redone.

## 4. Record and compare

Results directories hold transcripts and stay out of version control. The
**ledger** is the committed record: one JSON file per results directory
with every finding's verdict, vehicle, confidence, run count, harness
version, model, and end time.

```bash
go run . -ledger results/checks-YYYY-MM-DD -out ledger
go run . -ledger-diff ledger/checks-YYYY-MM-DD.json \
  -baseline ledger/<previous run>.json[,<other previous files>]
```

The diff prints a markdown draft: releases and models tested, verdicts
that moved, same-verdict evidence-grade changes, first observations, and
baseline checks this run did not repeat. The baseline is every ledger
that contributed to the current published state; when the previous run
was partial, list all the files that together cover every harness and
check.

Regenerate the ledger before the reports. The reports read `ledger/` to
mark changed verdicts, and a run that is not yet in the ledger carries no
markers.

## 5. Publish

**Classify each moved verdict** with the transcripts open. The draft sorts
the facts and stops there. The judgment that matters to readers is
whether the harness moved (injected content, catalog, files read) or the
model did (chose not to look, reported a failure differently, reused its
context). The runner's own `notes` on a finding often say which; model
swaps between runs are a strong hint toward model-level.

**Write the changelog entry** at the top of
[`site/content/changelog/_index.md`](../site/content/changelog/_index.md):
releases and models tested with the previous ones, platform-level changes
with before and after verdicts and what they mean for authors,
model-level changes grouped by harness, and any reruns.

**Regenerate reports and site**, listing every results directory that
contributes to the current state (latest finding per harness and check
wins):

```bash
ALL=results/<dir1>,results/<dir2>,...,results/checks-YYYY-MM-DD
go run . -report $ALL -out reports
go run . -report $ALL -site ../site
```

Platform pages, the comparison index, and the reports then carry
"changed since previous observation" markers for every verdict that
differs from the platform's previous run.

**Update hand-written pages that quote releases.** The guidance index
names the releases the guidance was derived from; guidance pages that
state a platform's behavior may now be stale (the authoring page's
name-mismatch paragraph was, after Claude Code 2.1.309). Set `lastmod` on
any page you edit.

**Build the site** (`cd site && hugo`) and check the changelog, the
comparison index markers, and the guidance list before committing. Commit
the ledger, reports, site content, and any dependency changes together.
Results directories are ignored and stay local.
