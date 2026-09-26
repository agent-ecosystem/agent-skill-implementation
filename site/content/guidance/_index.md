---
title: "Guidance"
description: "What the platform findings mean in practice, organized by who acts on them: skill authors, people installing skills, distributors and packagers, and harness implementers."
date: 2026-09-25
showTableOfContents: false
---

The same finding lands differently depending on who you are. A platform that reads only its native skills directory is a layout constraint for an author, an installation rule for a user, a packaging rule for a distributor, and an interoperability gap for whoever builds the harness. This section keeps each audience's guidance on its own page, with every recommendation citing the check that motivates it.

- **[Authoring](/guidance/authoring/)**: how to write a SKILL.md and its bundled files so they behave the same way everywhere.
- **[Installing and using](/guidance/using/)**: where to put skills, what collides with what, and what headless sessions can and cannot do.
- **[Distributing and packaging](/guidance/distributing/)**: what a catalog, marketplace, or plugin has to get right for a skill to load on every platform it ships to.
- **[Implementing a harness](/guidance/implementing/)**: where tested platforms contradict the specification or each other, and the choices that would narrow the gaps.
- **[What a bundled file costs on Copilot CLI](/guidance/enumeration-cost/)**: an audit of five public catalogs against the one platform that lists every bundled file at activation.
- **[Glossary](/guidance/glossary/)**: the terms used throughout.

This guidance is derived from the automated findings in the [platform comparison](/platforms/) (check list 0.4, tested in headless mode against Antigravity CLI 1.2.11, Claude Code 2.1.274, Codex CLI 0.157.0, and GitHub Copilot CLI 1.0.88; each platform report lists the releases behind its findings). Untested platforms may behave differently, so this guidance is likely to change as we study more harness skill implementations. The terms used here (push harness, pull harness, canary phrase, and more) are defined in the [glossary](/guidance/glossary/).
