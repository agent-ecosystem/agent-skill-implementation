---
title: "What a Bundled File Costs on Copilot CLI"
description: "An audit of five public skills repositories against Copilot CLI's activation file list: what each shipped file costs per activation, and where evals and packaging metadata land."
date: 2026-09-26
weight: 6
showTableOfContents: true
---

Copilot CLI is the only tested platform that lists a skill's bundled files when it activates the skill. It wraps the injected SKILL.md body in a header that states the skill's base directory and then names the absolute path of every file under the skill directory. It sends that header again on every activation in a session (`bundled-file-enumeration-scale`, `reactivation-deduplication`). Antigravity, Claude Code, and Codex list nothing. So on Copilot CLI, and only there, every file an author ships has a per-activation token cost whether or not the model ever reads it.

This page puts numbers on that cost using real skills rather than the benchmark fixture, because two common packaging habits looked like they might matter: Claude Code's skill-creator saves evaluation cases to `evals/evals.json` inside the skill directory, and some organizations ship signing, benchmark, and card files alongside every skill.

## Method

The header has a fixed shape. For a skill installed at `BASE`:

```
<skill-context name="NAME">
Base directory for this skill: BASE

Related files (use view tool to read):
  - BASE/references/api-patterns.md
  - BASE/evals/evals.json
  ...

```

A script in the repository (`research/copilot-enumeration-audit.py`) reconstructs this header for every `SKILL.md` in a checked-out repository and counts its tokens. Before auditing anything it rebuilds the header for the benchmark's own 43-file fixture and compares it with the header Copilot CLI 1.0.88 actually recorded in the transcript; the reconstruction matches byte for byte (6,362 characters, 42 files listed, the `vendor/` tree omitted as observed). Skills are assumed to be installed at `/Users/dev/project/.github/skills/<name>`. Tokens are counted with the `o200k_base` tokenizer as an approximation, since the tokenizer of the model Copilot ran (claude-sonnet-5) is not public; treat the figures as indicative, not exact.

Repositories audited on 2026-09-26: [NVIDIA/skills](https://github.com/NVIDIA/skills) (commit d8519c5), [anthropics/skills](https://github.com/anthropics/skills) (3337550), [melodic-software/claude-code-plugins](https://github.com/melodic-software/claude-code-plugins) (943db97), [magnus919/agent-skills](https://github.com/magnus919/agent-skills) (9a5adbc), and [TerminalSkills/skills](https://github.com/TerminalSkills/skills) (511ec20). The first is a large vendor catalog that ships evals with every skill, the second is the reference set from the format's authors, the third and fourth are community catalogs built with skill-creator, and the fifth is a large catalog of single-file skills that serves as a floor.

## Results

| Repository | Skills | Ship `evals/` | Files per skill (median, p90, max) | Header tokens per activation (median, p90, max) | Share from `evals/` | Share from other non-spec files | Header costs more than the body |
|---|---|---|---|---|---|---|---|
| NVIDIA/skills | 383 | 383 | 7, 26, 130 | 204, 724, 4,349 | 13% | 35% | 6 of 383 |
| anthropics/skills | 20 | 0 | 6, 69, 82 | 149, 1,729, 2,036 | 0% | 43% | 2 of 20 |
| melodic-software/claude-code-plugins | 278 | 267 | 3, 15, 364 | 88, 342, 10,615 | 31% | 33% | 5 of 278 |
| magnus919/agent-skills | 182 | 182 | 11, 29, 101 | 273, 663, 2,192 | 9% | 26% | 1 of 182 |
| TerminalSkills/skills | 1,020 | 0 | 1, 1, 31 | 53, 59, 735 | 0% | 31% | 0 of 1,020 |

"Other non-spec files" are entries outside `SKILL.md`, `scripts/`, `references/`, `assets/`, and `evals/`. The SKILL.md body medians run from 1,400 to 2,800 tokens across these repositories, so a median skill's header is a 5 to 15 percent surcharge on each activation, and the tail is where it hurts: the heaviest skills pay more for the file list than for their instructions.

### Installing a whole catalog

Enumeration is paid at activation, not at install, so installing every skill in a repository costs nothing in file paths until skills are activated. Two totals bound what it can cost. The first is the file-list cost of activating every skill once, the sum of every skill's header. The second is what a whole-catalog install does add to every session: the discovery listing, one `<skill>` block of name, description, and location per installed skill in Copilot CLI's system prompt (`discovery-listing-fields`), reconstructed here from the format the transcript records.

| Repository | Skills | Files | File-list tokens, every skill activated once | Of which `evals/` | Of which other non-spec | Discovery listing, every session |
|---|---|---|---|---|---|---|
| NVIDIA/skills | 383 | 4,817 | 141,109 | 18,177 | 48,778 | 24,739 |
| anthropics/skills | 20 | 394 | 10,628 | 0 | 4,547 | 1,895 |
| melodic-software/claude-code-plugins | 278 | 2,352 | 61,581 | 18,917 | 20,516 | 49,742 |
| magnus919/agent-skills | 182 | 2,733 | 65,768 | 6,142 | 17,224 | 6,946 |
| TerminalSkills/skills | 1,020 | 1,122 | 56,599 | 0 | 17,586 | 39,102 |

Activating all of NVIDIA's skills once would spend about 141,000 tokens on file paths alone, 67,000 of them naming evals and packaging material which the spec does not call for. Nobody activates a whole catalog, though. The listing is the real whole-catalog cost, and it is paid before the first prompt of every session: about 25,000 tokens for NVIDIA's 383 skills, 50,000 for melodic-software's 278 (its descriptions are long), and 39,000 for TerminalSkills' 1,020. That figure is set by description length rather than file count, and every tested platform carries some form of it, not only Copilot CLI. Install only the skills you use.

## What drives the cost

**Every file included in the skill bundle costs about 20 tokens. Half of that is the install path.** A line like `  - /Users/dev/project/.github/skills/my-skill/references/api-patterns.md` is 20 tokens, 11 of them the install prefix. The same skill installed deep in a monorepo (`/Users/firstname.lastname/Developer/company/platform-monorepo/services/billing/.github/skills/my-skill`) pays 31 tokens per line. Authors control the file count; users control the path.

In a bundle that ships 7 files, filepath enumeration costs an average of 140 tokens. In a bundle that ships 130 files, as in one example in the audit above, the token cost in filepath enumeration just for activating the skill is 4,349, whether or not the agent needs any of the supplemental files.

**`evals/evals.json` on its own is cheap.** In all three catalogs that ship evals, the median skill has exactly one file under `evals/`, costing about 20 tokens per activation. The evals share of the totals above (9 to 31 percent) comes from the tail, where evaluation fixtures live inside the skill: `warp-compile-time-optimizer` in NVIDIA/skills carries 2,231 tokens of `evals/files/` paths, and `setup` in melodic-software/claude-code-plugins carries 9,071, more than its 4,791-token body. Keep the JSON; move fixture trees out of the skill directory.

**Packaging metadata adds up across a catalog.** Every one of NVIDIA's 383 skills ships `skill.oms.sig`, `BENCHMARK.md`, and `skill-card.md` beside SKILL.md, about 60 tokens per activation per skill that no model needs, and 199 of them ship a `schemas/` tree. The community catalogs' non-spec cost comes from `context/`, `reference/` (singular, so not a spec directory), `extraction/`, `templates/`, and `tests/` directories. Anthropic's `canvas-design` lists 81 font files, 2,001 tokens of paths, at every activation. The 1,020 single-file skills in TerminalSkills/skills show the floor: about 50 tokens for the fixed lines alone.

**Reactivation multiplies it.** Copilot CLI re-injects the whole wrapper each time a skill is activated in a session (`reactivation-deduplication`), so a skill activated three times in one conversation pays the list three times.

## What to do with this

- **Authors**: ship only what the model needs, in the spec-enumerated directories. Each extra file is roughly 20 tokens per activation on Copilot CLI, dotfiles and binaries included; the only exclusion observed is a top-level `vendor/` tree. See the [authoring guidance](/guidance/authoring/).
- **Distributors**: strip signing, benchmark, card, schema, and test material from the installed skill directory, or install it beside the skill rather than inside it. A three-file metadata convention costs a catalog's users 60 tokens per activation per skill. See the [distributing guidance](/guidance/distributing/).
- **Harness implementers**: if you list files, cap the list and document the exclusions; on real catalogs the median list is 50 to 370 tokens and the tail passes 10,000. See the [implementer notes](/guidance/implementing/).

## Limitations

The token counts are approximations from a public tokenizer, not the model's own. The install path is assumed, and it is about half the cost per line, so a different layout shifts every number. Only one Copilot CLI release (1.0.88) has been observed; the exclusion list may be longer than `vendor/`, and node_modules was not tested. The header was reconstructed rather than captured for the audited skills, which is why the calibration against a captured header matters: the reconstruction is exact for the shape observed, and the audit's only inputs beyond that shape are the file inventories of public repositories.
