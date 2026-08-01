---
title: "Cross-Platform Authoring Guidance"
description: "What skill authors should do differently based on observed platform divergence, derived from the automated platform reports."
date: 2026-08-01
showTableOfContents: true
weight: 1
---

This guidance is derived from the automated findings in the [platform comparison](/platforms/) (check list 0.2, tested against Antigravity CLI 1.1.9, Claude Code 2.1.212, and Codex CLI 0.146.0 in headless mode). Each recommendation cites the divergence that motivates it. Where all tested platforms agree, we say so; agreement across three platforms is encouraging but not proof that the other 20+ behave the same way.

## Naming and identity

**Make the frontmatter `name` match the directory name exactly.** The spec requires it, but platforms neither enforce it nor resolve it the same way: when the two disagree, Codex and Antigravity index the skill by its frontmatter name, while Claude Code lists it under the directory name only. A mismatched skill answers to different names on different platforms, and anything that references it by name (documentation, other skills, user habit) breaks on whichever platform chose the other identity (`name-directory-mismatch`).

**Never install two skills with the same name at different scopes.** The implementation guide calls project-over-user precedence a universal convention. It is not: Claude Code resolved the collision user-over-project, and Codex does not resolve it at all (its catalog lists both entries and the model picks one, so the winner can vary between conversations). If a name collides, which content loads depends on the platform and sometimes on the run (`name-collision-precedence`).

## Frontmatter

**Always include a description, and quote any value containing a colon.** Codex skips skills with no description entirely, while Claude Code and Antigravity load them anyway; a missing description means your skill does not exist for some users (`missing-description-handling`). An unquoted colon in a description is invalid YAML that Codex and Claude Code tolerate but Antigravity's catalog rejects (`malformed-yaml-tolerance`).

**Do not put load-bearing information only in frontmatter.** Claude Code strips frontmatter before injecting skill content, so `compatibility` notes and `metadata` values never reach the model there unless it happens to read the raw file. On pull harnesses the model sees frontmatter only because it reads the file itself. If the model must know something (requirements, warnings, version constraints), state it in the body (`frontmatter-handling`, `compatibility-field-behavior`).

**Expect nonstandard fields to be ignored.** No tested platform acted on `requires`, `depends-on`, or `priority`, even with the named dependencies installed (`nonstandard-dependency-fields`).

## Layout and installation

**Keep each skill a direct child of the skills directory.** Grouping skills in subfolders (`skills/data-tools/my-skill/`) works on Codex, which scans recursively, but Claude Code and Antigravity only see direct children; grouped skills vanish from their catalogs (`recursive-root-discovery`).

**Do not ship a SKILL.md inside another skill's directories.** Codex registers any SKILL.md it finds under the skills root as a separate, invocable skill, including one inside your `references/` folder. An example or vendored skill shipped as documentation becomes a live skill on Codex and stays invisible on the other platforms (`nested-skill-discovery`).

**Install into each platform's native skills directory.** The `.agents/skills/` cross-client convention is not universally scanned: Codex reads it, Antigravity uses it natively, and Claude Code does not scan it at all. A skill installed only at the convention path is invisible on Claude Code (`cross-client-directory-interop`).

**Deep resource nesting is safe on the tested platforms.** Reference files up to five directory levels deep were reachable everywhere (`resource-nesting-depth`). The spec's "one level deep" language concerns chains of references between files, not directory depth.

## Paths and resources

**Do not expect SKILL.md-relative paths to resolve as written.** On every tested platform, a bare path like `references/setup-guide.md` fails when used from the working directory; models usually recover by re-qualifying the path against the skill directory, but that recovery is model behavior, not a platform guarantee (`path-resolution-base`). Telling the model explicitly that relative paths resolve against the skill's own directory costs one sentence and removes the guesswork.

**Resources load only when the model reads them.** No tested platform enumerates resource files, pre-fetches markdown-linked files, or eagerly loads directory contents at activation, and this includes nonstandard directories like `resources/` or `templates/` (`resource-enumeration-behavior`, `eager-link-resolution`, `unrecognized-directory-handling`). If a file matters, instruct the model to read it; a mention is not a load.

**Name shared resource files distinctively.** Two active skills both shipping `references/API.md` did not shadow each other in our runs, but only because the models qualified every path themselves; the platforms' handling of a truly ambiguous path went unexercised. Distinct filenames remove the risk outright (`cross-skill-resource-shadowing`).

## Composition and lifecycle

**Skill-to-skill chains work today, but guards are the model's, not the platform's.** Three-skill invocation chains completed on all platforms, in English and Japanese, and prose-expressed dependencies were resolved (`invocation-depth-limit`, `invocation-language-sensitivity`, `informal-dependency-resolution`). Circular references were stopped by the model choosing to stop, not by any platform mechanism, so do not design skills that rely on a platform catching a cycle (`circular-invocation-handling`).

**Missing dependencies fail visibly.** When a skill references an uninstalled skill, every tested platform surfaced the failure (an explicit tool error on Claude Code; the model reporting the absence on Codex and Antigravity). None hallucinated compliance in our runs, which is the good outcome, but the failure is still a runtime discovery your user makes, not something any platform checks at install time (`missing-dependency-behavior`, `cross-scope-dependency`).

**Editing a skill mid-session takes effect on reactivation.** All tested platforms served fresh content after SKILL.md was edited during a session (`reactivation-freshness`). Claude Code re-injects the full body on every reactivation, so repeated activations of a large skill have a real token cost there (`reactivation-deduplication`).

---

Derived from automated, transcript-cited findings; see the [platform reports](/platforms/) for verdicts, confidence grades, and caveats, and [the check list](/checks/) for what each check evaluates.
