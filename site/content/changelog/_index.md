---
title: "Changelog"
description: "What changed between running the checks: the harness releases tested, which check results changed, and whether the platform or the model moved them."
date: 2026-11-01
showTableOfContents: true
---

Every finding on this site is a dated observation of one harness release running one model. Harnesses ship updates frequently, the default models behind them change, and both of these things can change a check result. The [platform pages](/platforms/) always show the latest observation, and mark with 🔄 any result that differs from the previous observation of that platform, with the earlier result and release stated under the finding. This page records changes over time and what moved each one.

Each entry below represents one run of the automated checks. It lists the releases and models tested, and every check result that changed since the previous run:

- **Platform-level** changes are attributable to the harness: the content it injects, the catalog it builds, the files it reads. These are the changes that matter most to skill authors, because they apply regardless of which model is active.
- **Model-level** changes are the model making a different choice within the same platform behavior: whether to go looking for a directory the harness never showed the model, how to report a dependency it could not find, whether to re-read a file it already has in context. These are expected to vary, and a single verdict in either direction is one observed outcome that says nothing about how often it happens.

The raw material is the [verdict ledger](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner/ledger): one small JSON file per run of the checks with every finding's verdict, harness version, and model. The runner diffs two ledgers into a draft; the classification into platform-level and model-level is done by hand from the transcripts.

## 2026-11-01: Autumn releases, and the first two platform-level changes

The full check list (44 automated checks) was run against every supported harness, sandboxed, on the releases current at the end of October. Of 176 findings compared with the previous runs, 167 kept the same verdict and evidence grade and 9 moved. Two of the nine are attributable to the platform; they are the first platform-level changes this project has observed since building the automated check runner in July.

| Platform | Release tested | Model | Previously |
|---|---|---|---|
| Antigravity CLI | 1.4.12 | Gemini 4 Argon (Medium) | 1.2.11, Gemini 3.8 Flash (High) |
| Claude Code | 2.1.309 | claude-sonnet-5-5 | 2.1.274, claude-sonnet-5 |
| Codex CLI | 0.166.0 | gpt-6.1-sol | 0.157.0, gpt-6-astra |
| GitHub Copilot CLI | 1.0.105 | claude-sonnet-5.5 | 1.0.88, claude-sonnet-5 |

### Platform-level changes

**Claude Code 2.1.309 now lists a skill under both the directory name and the frontmatter name when they differ.** The [`name-directory-mismatch`](/platforms/claude-code/#name-directory-mismatch) result changed from `directory-name-identity` to `listed-under-both`. The fixture lives in a directory called `probe-mismatch-dir` while its frontmatter declares `name: probe-name-mismatch`. On 2.1.274 the skill catalog injected at session start read `- probe-mismatch-dir: <description>`; on 2.1.309 the same entry reads `- probe-mismatch-dir (probe-name-mismatch): <description>` with the mismatched frontmatter name presented in parentheses. Skills whose directory and frontmatter names agree are listed exactly as before, so the parenthetical appears only when there is a mismatch to show. Activating by either name loads the skill. The evidence is transcript-direct: the listing is harness-injected text, recorded verbatim.

What it means for authors: the spec requires the two names to match, and that is still the portable choice. On Claude Code a mismatch is now surfaced to the model, where earlier releases resolved it to the directory name with no indication that another name existed, which makes the skill reachable by its declared name but also puts the inconsistency in front of the model on every turn.

**Antigravity CLI 1.4.12 now lists a skill whose frontmatter does not parse.** The [`malformed-yaml-tolerance`](/platforms/antigravity/#malformed-yaml-tolerance) verdict moved from `unlisted-but-reachable` to `tolerated-and-loaded`, consistent across three runs. The fixture's description contains an unquoted colon, which is invalid YAML. On 1.1.9 and 1.2.11 the skill was not presented to the model in the available skill catalog, and it loaded only when the model read the file directly. On 1.4.12 the model reports the skill in its catalog with the description intact, not truncated at the colon and not quoted. Antigravity records no injected context in its transcripts, so this rests on the model's account of its catalog (behavioral inference), but the account did not vary across three runs and two earlier releases gave the opposite answer just as consistently.

What it means for authors: a frontmatter error that used to make a skill invisible on Antigravity no longer does. That is more forgiving, and it also means the error will not be caught by the platform. Validate frontmatter before shipping; see [authoring guidance](/guidance/authoring/).

### Model-level changes

Every harness shipped with a different default model than in September, and the other changes in skill behavior track model changes.

- **Antigravity CLI** (Gemini 3.8 Flash to Gemini 4 Argon): [`directory-naming-divergence`](/platforms/antigravity/#directory-naming-divergence) and [`unrecognized-directory-handling`](/platforms/antigravity/#unrecognized-directory-handling) moved from readable on demand to untouched. The harness never surfaces those directories; what changed is that the new model did not go looking. [`circular-invocation-handling`](/platforms/antigravity/#circular-invocation-handling) moved from `reinvocation-no-reload` to `cycle-stopped-model-choice`, and [`missing-dependency-behavior`](/platforms/antigravity/#missing-dependency-behavior) moved from reporting the missing skill to attempting it and failing visibly. Both are the model deciding how to handle an instruction it cannot complete.
- **Claude Code** (claude-sonnet-5 to claude-sonnet-5-5): [`path-resolution-base`](/platforms/claude-code/#path-resolution-base) moved from `cwd-base-model-requalified` to `model-preemptively-qualified`. The model wrote fully qualified paths from the start, so the platform's resolution base was never exercised; the earlier verdict came from a model that tried the bare path first. [`unrecognized-directory-handling`](/platforms/claude-code/#unrecognized-directory-handling) moved from readable on demand back to untouched, the same pattern as Antigravity.
- **GitHub Copilot CLI** (claude-sonnet-5 to claude-sonnet-5.5): [`path-resolution-base`](/platforms/copilot/#path-resolution-base) moved the same way as Claude Code's, for the same reason.
- **Codex CLI** (gpt-6-astra to gpt-6.1-sol): all 44 verdicts unchanged.

Two checks were rerun three times to separate noise from signal. Copilot's [`reactivation-deduplication`](/platforms/copilot/#reactivation-deduplication) came back inconclusive on the first run because the model answered the second turn from memory; three further runs all showed the platform reinjecting the skill, the same verdict as September. Antigravity's malformed-yaml result above is the other.

Tooling: this run moved the runner to agentsummons 0.4.1, agentminutes 0.9.0, and skillxp 0.4.1, which read Codex 0.166.0's compressed transcripts and Claude Code 2.1.309's request log.

## 2026-09-25: Check list 0.4, GitHub Copilot CLI joins

Check list [0.4](/checks/#changelog) added five checks, prompted by adding GitHub Copilot CLI as the fourth automated harness. The three existing harnesses were checked on their current releases, and Copilot was observed for the first time (44 findings, no baseline to compare).

| Platform | Release tested | Model | Previously |
|---|---|---|---|
| Antigravity CLI | 1.2.11 | Gemini 3.8 Flash (High) | 1.1.9, Gemini 3.6 Flash (High) |
| Claude Code | 2.1.274 | claude-sonnet-5 | 2.1.212, claude-sonnet-5 |
| Codex CLI | 0.157.0 | gpt-6-astra | 0.146.0, gpt-5.6-sol |
| GitHub Copilot CLI | 1.0.88 | claude-sonnet-5 | first run |

Of 117 findings compared, 9 verdicts changed. None was attributable to the platform.

- **Antigravity CLI**: `unrecognized-directory-handling` moved from untouched to readable on demand (the model chose to read `evals/` and `templates/`); `circular-invocation-handling` moved from `cycle-stopped-model-choice` to `reinvocation-no-reload`; the missing-dependency tier in `cross-scope-dependency` changed. All three are the model's choices on a pull harness.
- **Claude Code**: `unrecognized-directory-handling` moved from untouched to readable on demand. The harness injects only SKILL.md and never surfaces those directories; the model read them with tools.
- **Codex CLI**: with the default model moving from gpt-5.6-sol to gpt-6-astra, five verdicts moved. `reactivation-deduplication` went from `re-read-each-activation` to `not-re-read` and `reactivation-freshness` from `fresh-content-served` to `edit-missed-no-reload`, consistent across three runs: the new model reused the copy of SKILL.md already in its context instead of re-reading the file, so a mid-session edit never reached it. On a pull harness that is the model's decision, but the consequence for authors is real and is noted in the [using guidance](/guidance/using/). `missing-description-handling` moved from `skipped-as-guide-prescribes` to `unlisted-but-reachable`: the skill stayed out of the catalog, and the model read the file anyway. The missing-dependency tiers in `cross-scope-dependency` and `missing-dependency-behavior` changed.

## 2026-09-19: Check list 0.3, description length unit

Check list [0.3](/checks/#changelog) added `description-length-unit`, with two multibyte fixtures at the 1024 boundary, after a validator bug report showed that enforcing implementations disagree on whether the limit counts code points, UTF-16 units, or bytes. Only the new check ran.

| Platform | Release tested | Model | Result |
|---|---|---|---|
| Antigravity CLI | 1.2.7 | Gemini 3.8 Flash (High) | No length enforcement |
| Claude Code | 2.1.267 | claude-fable-5-1 | No length enforcement |
| Codex CLI | 0.154.0 | gpt-6-astra | Counts code points (ASCII overrun truncated at 1024; both multibyte fixtures delivered whole) |

## 2026-08-01: Check list 0.2, first full run

Check list [0.2](/checks/#changelog) grew the automated set from five checks to the full list, adding the spec-alignment and validation-strictness checks. This run produced the first observation for 99 harness-and-check pairs. Of the 15 pairs that had a July baseline, 2 moved: the missing-dependency tier in `cross-scope-dependency` on Antigravity and on Codex, which is the model's choice of how to report a skill it cannot find.

| Platform | Release tested | Model | Previously |
|---|---|---|---|
| Antigravity CLI | 1.1.9 | Gemini 3.6 Flash (High) | 1.1.5, Gemini 3.5 Flash (Medium) |
| Claude Code | 2.1.212 | claude-sonnet-5 (claude-fable-5 on some checks) | 2.1.205, claude-sonnet-5 |
| Codex CLI | 0.146.0 | gpt-5.6-sol | 0.144.6, gpt-5.6-sol |

## 2026-07-21: First automated run

The runner's first sandboxed run covered five checks on three harnesses: `discovery-reading-depth`, `activation-loading-scope`, `cross-scope-dependency`, `reactivation-deduplication`, and `reactivation-freshness`. It established the baseline the later runs are measured against: all three read only metadata at discovery and only the SKILL.md body at activation, with Claude Code injecting the body (harness-push) and Antigravity and Codex having the model read the file (model-pull).

| Platform | Release tested | Model |
|---|---|---|
| Antigravity CLI | 1.1.5 | Gemini 3.5 Flash (Medium) |
| Claude Code | 2.1.205 | claude-sonnet-5 |
| Codex CLI | 0.144.6 | gpt-5.6-sol |
