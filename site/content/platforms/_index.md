---
title: "Platform Reports"
description: "Where agent platforms agree and diverge on skill loading, from automated transcript-cited checks (check list 0.2)."
date: 2026-08-01
showTableOfContents: true
---

We ran the same 38 automated checks against each platform and compared what actually happened. Each row below asks one question about platform behavior; the cells say in plain language what each platform did. Rows marked 📌 are where platforms disagree: the cases where a skill that works on one platform behaves differently on another.

Full detail for every finding (the exact verdict, how content reached the model, confidence, and notes) lives on the per-platform pages. Each page opens with a spec alignment summary: where that platform's observed behavior contradicts or matches what the [Agent Skills specification](https://agentskills.io/specification) prescribes, and how it handles skills that violate the spec's format rules.

- [Antigravity CLI (headless)](/platforms/antigravity/)
- [Claude Code (headless)](/platforms/claude-code/)
- [Codex CLI (headless)](/platforms/codex/)

For what these findings mean when writing a skill, see the [cross-platform authoring guidance](/guidance/authoring-guidance/). The [check list](/checks/) has the full rationale behind every question. All findings come from headless sessions, which can differ from interactive use; outcomes marked † rest on behavioral inference rather than direct transcript evidence. Two checks (context compaction protection, trust gating) need interactive sessions and are marked manual.

## Loading Timing

When skill content enters the model's context, and how much loads at each stage.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| Does the harness read only SKILL.md metadata at discovery, or the full body? | Metadata only † | Metadata only | Metadata only |
| On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle? | Body only | Body only | Body only |
| Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text? | No pre-fetching † | No pre-fetching | No pre-fetching |

## Directory Recognition

Which directories a platform treats as part of a skill, and what happens to ones it doesn't recognize.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation? | Nothing enumerated † | Nothing enumerated | Nothing enumerated |
| 📌 Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible? | Readable when the model looks | Not surfaced; model never looked | Not surfaced; model never looked |
| What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible? | Not surfaced; model never looked † | Not surfaced; model never looked | Not surfaced; model never looked |

## Resource Access Patterns

How supporting files (scripts, references, assets) become available to the model.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored? | Nothing enumerated † | Nothing enumerated | Nothing enumerated |
| 📌 When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written? | Bare path fails; model recovers | Bare path fails; model recovers | Model used full paths (base untested) † |
| With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's? | Got its own file † | Got its own file † | Got its own file † |
| Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked? | Reads outside the skill allowed | Reads outside the skill allowed | Reads outside the skill allowed |
| How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels. | All depths reachable (tested to 5) | All depths reachable (tested to 5) | All depths reachable (tested to 5) |
| 📌 Can the agent run a bundled scripts/ file and receive its output? | Script ran; output returned | Blocked with an error | Script ran; output returned |

## Content Presentation

What the model actually sees, at discovery and at activation, and how it's formatted.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| 📌 Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations? | Also surfaces location † | Name and description only | Also surfaces location |
| 📌 Does the SKILL.md YAML frontmatter reach the model at activation, or only the body? | Visible (model reads the raw file) | Stripped before injection | Visible (model reads the raw file) |
| 📌 Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses? | Raw file via model read | Raw markdown, no wrapper tags | Raw file via model read |

## Lifecycle Management

How skill content is managed over the course of a conversation.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| 📌 When the same skill is activated twice in one session, is its content loaded again or deduplicated? | Model re-reads each time | Full content re-injected every time | Model re-reads each time |
| After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy? | Edits picked up immediately | Edits picked up immediately | Edits picked up immediately |
| Is skill content protected when the context window fills up? | _manual_ | _manual_ | _manual_ |

## Access Control

How platforms gate skill loading and handle control-related frontmatter.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| Do project-level skills require trust approval before loading? | _manual_ | _manual_ | _manual_ |
| Does a compatibility field naming another platform gate loading, get surfaced to the model, or get ignored? | Loads normally, no gating | Loads normally, no gating | Loads normally, no gating |
| Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it? | Ran with and without the field | Ran with and without the field | Ran with and without the field |

## Skill-to-Skill Invocation

Whether one skill's instructions can activate another skill.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| Can one skill's instructions get a second installed skill activated by name? | Second skill activated | Second skill activated | Second skill activated |
| How deep does a skill→skill→skill activation chain run before something stops it? | Full three-skill chain completed | Full three-skill chain completed | Full three-skill chain completed |
| When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice? | Model stopped the loop itself † | Model stopped the loop itself | Model stopped the loop itself |
| Does the invoke chain still complete when the activation prompt is Japanese? | Full three-skill chain completed | Full three-skill chain completed | Full three-skill chain completed |

## Skill Dependencies

What happens when one skill depends on another, formally or in prose.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime? | Second skill activated | Second skill activated | Second skill activated |
| 📌 When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped? | Reported missing without attempting | Failed with a visible error | Reported missing without attempting |
| Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)? | Ignored † | Ignored | Ignored |
| 📌 Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent? | Resolved across scopes; missing dependency reported, not attempted | Resolved across scopes; missing dependency fails visibly | Resolved across scopes; missing dependency reported, not attempted |

## Discovery Scope

Where platforms look for skills: which directories are scanned, how deep, and what happens when two discovered skills claim the same name.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| 📌 Is a skill installed only at the cross-client .agents/skills convention path discovered? | Convention path is the native directory † | Convention path not scanned | Convention path scanned |
| 📌 Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered? | Direct children only; stray file ignored † | Direct children only; stray file ignored | Scans the root recursively; stray file ignored |
| 📌 Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill? | Not discovered † | Not discovered | Discovered as a separate skill |
| 📌 With the same skill name installed at project and user scope, which variant's content activates? | Project scope wins † | User scope wins | Lists both; model chose the project variant † |

## Validation Strictness

How strictly platforms judge skills that break the spec's format rules: rejected, repaired, or loaded anyway.

| Question | Antigravity CLI | Claude Code | Codex CLI |
|---|---|---|---|
| 📌 Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable? | Not cataloged; file still readable † | Tolerated and loaded | Tolerated and loaded |
| 📌 Is a skill with no description field skipped (as the guide prescribes), or loaded anyway? | Loaded anyway † | Loaded anyway | Skipped (as the guide prescribes) |
| 📌 Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable? | All three invalid names tolerated † | All three invalid names tolerated | Tolerated only: uppercase, double-hyphen |
| 📌 When directory name and frontmatter name disagree, which identity is the skill listed and invocable under? | Frontmatter name wins † | Directory name wins | Listed under both names |
| Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model? | Loaded fine | Loaded fine | Loaded fine |
| Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated? | Loaded anyway † | Loaded anyway | Loaded anyway |
| Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable? | Loaded anyway † | Loaded anyway | Loaded anyway |

