---
title: "Installing and Using Skills"
description: "Where to install skills, what collides with what, and what headless sessions can and cannot do, derived from platform reports."
date: 2026-09-25
weight: 2
showTableOfContents: true
---

This page is for people who install and run skills rather than write them. This guidance is derived from the automated findings in the [platform comparison](/platforms/) (check list 0.4, tested in headless mode against Antigravity CLI 1.2.11, Claude Code 2.1.274, Codex CLI 0.157.0, and GitHub Copilot CLI 1.0.88; each platform report lists the releases behind its findings). Untested platforms may behave differently, so this guidance is likely to change as we study more harness skill implementations. The terms used here (push harness, pull harness, canary phrase, and more) are defined in the [glossary](/guidance/glossary/).

## Where to install

**Install into each platform's native skills directory.** The `.agents/skills/` cross-client convention is not universally scanned. Codex and Copilot CLI read it, Antigravity uses it natively, and Claude Code does not scan it at all. A skill installed only at the convention path is invisible on Claude Code (`cross-client-directory-interop`).

**Install each skill as a direct child of the skills directory.** Codex and Copilot CLI find skills in grouping subfolders (`skills/data-tools/my-skill/`) because they scan recursively, but Claude Code and Antigravity list direct children only, and the grouped skills vanish from their catalogs with no error (`recursive-root-discovery`).

**Do not install the same skill under two roots in one repository unless the copies are identical.** Copilot CLI scans `.github/skills`, `.agents/skills`, and `.claude/skills` but lists a colliding name once, native root first, so the convention copies are shadowed with no warning. Codex lists both its native copy and the `.agents/skills` copy and the model picks one. Claude Code and Antigravity read only their own root, so the second copy is invisible there. If you keep a repository usable from several tools, keep one copy and symlink or document the rest (`multi-root-collision-precedence`).

**Never install two skills with the same name at different scopes.** The Agent Skill implementation guide calls project-over-user precedence "a universal convention." In practice, it's not. Antigravity and Copilot CLI follow the convention. Claude Code resolves a naming collision user-over-project. Codex does not resolve it at all; its catalog lists both entries, and the model picks one, so the winner may vary between conversations. If a name collides, which content loads depends on the platform and sometimes on the run (`name-collision-precedence`).

## Dependencies between skills

**A skill's dependencies must be installed too, but they can live at user scope.** When a skill's instructions activate another skill by name, every tested platform resolved a dependency installed at user scope from a project-level skill (`cross-scope-dependency`). When the dependency is missing, the failure surfaced at runtime on every tested platform, as an explicit tool error on Claude Code and Copilot CLI, and on Codex and Antigravity only because the model noticed and said so. None checks at install time, so the first activation is where you find out (`missing-dependency-behavior`).

## Sessions and permissions

**Scripts need an approval path.** Whether a skill's bundled script can run depends on the platform's permission posture, not on the skill. In headless sessions the script ran on Codex and Antigravity and was blocked on Claude Code and Copilot CLI, where the model could only report the denial (`bundled-script-execution`). If a skill relies on its scripts, run it interactively where you can approve the command, or with the platform's permission bypass, and do not expect the `allowed-tools` field to help: no tested platform demonstrably acted on it under any spelling (`allowed-tools-behavior`, `allowed-tools-name-matching`).

**Newly installed skills need a new session; edited skills usually do not.** Catalogs are enumerated when a session starts, so a skill installed mid-session is not discoverable until the next one. Edits to an already-installed skill were served fresh on the next activation on Claude Code, Copilot CLI, and Antigravity. On Codex the model kept using the copy already in context in every run and never saw the edit, so start a new session there after editing (`reactivation-freshness`).

**Activating a skill twice usually costs twice.** Claude Code and Copilot CLI re-inject the full body on every reactivation, and Antigravity's model re-read the file by choice in our runs; Codex's model reused its earlier copy instead (`reactivation-deduplication`). On Copilot CLI each activation also carries the full path of every bundled file (`bundled-file-enumeration-scale`). Prefer one activation per task over repeated ones in long sessions.
