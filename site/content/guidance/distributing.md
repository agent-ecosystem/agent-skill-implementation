---
title: "Distributing and Packaging Skills"
description: "What a catalog, marketplace, or plugin has to get right for a skill to load on every platform it ships to, derived from platform reports."
date: 2026-09-25
weight: 3
showTableOfContents: true
---

This page is for people who package skills for others: skill directories and marketplaces, plugin authors, and teams that roll skills out across an organization. It overlaps with the [installing page](/guidance/using/) wherever a package decides where files land, and adds the concerns a catalog has that a single user does not. This guidance is derived from the automated findings in the [platform comparison](/platforms/) (check list 0.4, tested in headless mode against Antigravity CLI 1.2.11, Claude Code 2.1.274, Codex CLI 0.157.0, and GitHub Copilot CLI 1.0.88; each platform report lists the releases behind its findings). Untested platforms may behave differently, so this guidance is likely to change as we study more harness skill implementations. The terms used here (push harness, pull harness, canary phrase, and more) are defined in the [glossary](/guidance/glossary/).

## Validate to the strictest platform

**Accept into a catalog only what the strictest tested platform would load.** Validation posture differs by rule and by platform, and a skill that loads on a lenient platform vanishes with no error on a strict one. The union of enforced rules across tested platforms is: a description is required (Codex skips a skill without one), it must be valid YAML with colons quoted (the Antigravity and Copilot CLI catalogs reject an unquoted colon), the description must be under 1024 characters (Copilot CLI drops the skill entirely, Codex truncates), and the name must be ASCII and at most 64 characters (Codex and Copilot CLI reject longer names, and Copilot CLI drops any non-ASCII name regardless of length). Uppercase letters and consecutive hyphens in names were tolerated everywhere but are still invalid (`missing-description-handling`, `malformed-yaml-tolerance`, `oversize-description-handling`, `invalid-name-tolerance`).

**Count the description limit in UTF-16 code units, not characters.** The two enforcing platforms disagree on the unit: Codex counts Unicode code points and Copilot CLI counts UTF-16 code units, so a description heavy in emoji or CJK text can pass one and fail the other. Enforcing the stricter unit at intake means an entry that passes your catalog loads on both (`description-length-unit`). For names the same question dissolves if the catalog requires ASCII: two of the validators in circulation both count code points but differ on the character set (the spec's skills-ref accepts Unicode alphanumerics, skill-validator rejects non-ASCII names), and requiring ASCII satisfies both and every tested platform (`name-length-unit`).

**Require the frontmatter name to match the directory name.** Codex, Antigravity, and Copilot CLI activate a mismatched skill by its frontmatter name while Claude Code activates it by the directory name, so a mismatched entry answers to different names on different platforms and any documentation that names it is wrong somewhere (`name-directory-mismatch`).

## Names in a catalog

**Keep names unique across everything you distribute, and across the scopes you install to.** No tested platform warns on a name collision, and each resolves it differently: project over user on Antigravity and Copilot CLI, user over project on Claude Code, and both listed with the model choosing on Codex. Two catalog entries with the same name, or a catalog entry that matches something a user already installed, produce a different winner per platform (`name-collision-precedence`).

## Packaging and install layout

**Install a skill into exactly one root per project.** A package or plugin that writes the same skill into several roots to cover several tools creates a collision on the platforms that scan more than one: Copilot CLI lists the name once and shadows the convention copies, Codex lists both and the model picks. Write to one root and document the others, or generate per-platform packages (`multi-root-collision-precedence`, `cross-client-directory-interop`).

**Package the skill as a direct child of the skills directory, with nothing nested.** Grouping folders hide skills from Claude Code and Antigravity, and a SKILL.md shipped inside another skill's `references/` becomes a live, invocable skill on Codex (`recursive-root-discovery`, `nested-skill-discovery`). Example skills in a package need to live outside any skills root.

**Ship lean packages, and strip build and test material.** Copilot CLI injects the full path of every file under the skill directory on every activation, dotfiles and binaries included; a `vendor/` tree was the one thing it left out. Fixtures, evaluation data, and vendored dependencies in a package are a per-activation token cost there (`bundled-file-enumeration-scale`). On five public catalogs each shipped file costs about 20 tokens per activation; a three-file packaging convention (a signature, a benchmark, a card beside every SKILL.md) costs every user of a 383-skill catalog 60 tokens per activation per skill, and a `schemas/` or `tests/` tree costs more. Install such material beside the skill directory, not inside it ([the audit](/guidance/enumeration-cost/)).

**Do not encode requirements in frontmatter that a platform will never show.** `compatibility`, `metadata`, and nonstandard fields never reach the model on Claude Code or Copilot CLI, and reach it on Codex and Antigravity only when the model reads the raw file. A catalog can display these fields to people; the skill body has to carry anything the model must know (`frontmatter-handling`, `compatibility-field-behavior`, `discovery-listing-fields`, `nonstandard-dependency-fields`).
