---
title: "Cross-Platform Authoring Guidance"
description: "What skill authors should do differently based on observed platform divergence, derived from platform reports."
date: 2026-08-01
showTableOfContents: true
---

This guidance is derived from the automated findings in the [platform comparison](/platforms/) (check list 0.3, tested against Antigravity CLI 1.1.9, Claude Code 2.1.212, Codex CLI 0.146.0, and GitHub Copilot CLI 1.0.88 in headless mode). Untested platforms may behave differently, so this guidance is likely to change as we study more harness skill implementations. The terms used here (push harness, pull harness, canary phrase, and more) are defined in the [glossary](/guidance/glossary/).

## Naming and identity

**Make the frontmatter `name` match the directory name exactly.** The spec requires it. In platforms under test, we've observed they don't enforce this convention, nor resolve it the same way. Codex, Antigravity, and Copilot CLI index the skill by its frontmatter name, while Claude Code lists it under the directory name only. A mismatched skill answers to different names on different platforms. Anything that references it by name, such as documentation, other skills, or user habit, breaks on whichever platform chose the other identity (`name-directory-mismatch`).

**Follow all of the name rules, not just the ones your platform enforces.** Enforcement varies by rule: every tested platform tolerated uppercase letters and consecutive hyphens in a skill name, but a name over the spec's 64-character limit is rejected by Codex and Copilot CLI while Claude Code and Antigravity load it anyway. A rule-breaking name works where you tested it and vanishes elsewhere, with no error (`invalid-name-tolerance`).

**Never install two skills with the same name at different scopes.** The Agent Skill implementation guide calls project-over-user precedence "a universal convention." In practice, it's not. Antigravity and Copilot CLI follow the convention. Claude Code resolves a naming collision user-over-project. Codex does not resolve it at all; its catalog lists both entries, and the model picks one, so the winner may vary between conversations. If a name collides, which content loads depends on the platform and sometimes on the run (`name-collision-precedence`).

## Frontmatter

**Always include a description, and quote any value containing a colon.** Codex doesn't register skills with no description. Claude Code, Antigravity, and Copilot CLI load them anyway, but with no description, these harnesses don't know when to invoke the skill. So a missing description means your skill does not exist for some users (`missing-description-handling`). In YAML formatting tests, an unquoted colon in a description is invalid YAML that Codex and Claude Code tolerate, but the Antigravity and Copilot CLI catalogs reject (`malformed-yaml-tolerance`).

**Do not put load-bearing information only in frontmatter.** Claude Code and Copilot CLI strip frontmatter before injecting skill content, and their skill listings carry only names and descriptions, so `compatibility` notes and `metadata` values never reach the model through any platform channel. The model can still stumble on the raw file; skills live at conventional paths it can guess, and project skills sit inside the working tree; but nothing guarantees it will. On pull harnesses (Antigravity, Codex CLI), the model sees frontmatter only because it reads the file itself. If the model must know something (requirements, warnings, version constraints), state it in the body (`frontmatter-handling`, `compatibility-field-behavior`, `discovery-listing-fields`).

**Keep the description under 1024 characters, and put the activation cues early.** Enforcement of the spec's limit ranges from none to total: Claude Code injects the full over-limit value into its listing, Antigravity loads the skill too, Codex truncates the description partway through, and Copilot CLI drops the skill from its catalog entirely. Anything past Codex's truncation point, including trailing "use when" sentences, never reaches the model, and on Copilot CLI the whole skill is gone (`oversize-description-handling`). The two enforcing platforms even measure the limit differently: Codex counts Unicode code points and Copilot CLI counts UTF-16 code units, so a description heavy in emoji or CJK text can pass on one and fail on the other (`description-length-unit`). An over-limit `compatibility` value also loaded everywhere (`oversize-compatibility-handling`).

**Do not rely on `allowed-tools` to pre-approve anything.** The spec marks the field experimental, and no tested platform demonstrably acted on it: the instructed command ran with and without the field on three platforms, and was blocked with and without it on Copilot CLI, each under the platform's normal headless permission posture. On Claude Code the value never reaches the model at all (`allowed-tools-behavior`).

**Expect nonstandard fields to be ignored.** No tested platform acted on `requires`, `depends-on`, or `priority`, even with the named dependencies installed (`nonstandard-dependency-fields`). Since Claude Code and Copilot CLI don't see metadata fields at all, and Antigravity and Codex CLI only see frontmatter when the model reads the file, tested harnesses do not benefit from any values in nonstandard fields.

## Layout and installation

**Keep each skill a direct child of the skills directory.** Grouping skills in subfolders (`skills/data-tools/my-skill/`) works on Codex and Copilot CLI, which scan recursively, but Claude Code and Antigravity only see direct children. Grouped skills vanish from their catalogs (`recursive-root-discovery`).

**Do not ship a SKILL.md inside another skill's directories.** Codex registers any SKILL.md it finds under the skills root as a separate, invocable skill, including one inside your `references/` folder. An example or vendored skill shipped as documentation becomes a live skill on Codex and stays invisible on the other platforms (`nested-skill-discovery`).

**Install into each platform's native skills directory.** The `.agents/skills/` cross-client convention is not universally scanned. Codex and Copilot CLI read it, Antigravity uses it natively, and Claude Code does not scan it at all. A skill installed only at the convention path is invisible on Claude Code (`cross-client-directory-interop`).

**Deep resource nesting is safe on the tested platforms.** Reference files up to five directory levels deep were reachable everywhere (`resource-nesting-depth`). The spec's "one level deep" language concerns chains of references between files, not directory depth.

## Paths and resources

**Do not expect SKILL.md-relative paths to resolve as written.** The spec instructs skill authors "when referencing other files in your skill, use relative paths from the skill root." But on every tested platform, a bare path like `references/setup-guide.md` fails when used from the working directory. Models usually recover, at the cost of a failed read and re-qualifying the path against the skill directory, but that recovery is model behavior, not a platform guarantee (`path-resolution-base`). Telling the model explicitly that relative paths resolve against the skill's own directory costs one sentence and removes the guesswork.

**Resources load only when the model reads them.** No tested platform pre-fetches markdown-linked files or loads directory contents at activation, and only one enumerates them: Copilot CLI wraps the injected body in a header that lists the full path of every file under the skill directory, spec directories and nonstandard ones like `resources/` or `templates/` alike, but loads none of them. Antigravity, Claude Code, and Codex list nothing, so on those platforms the agent doesn't know what extra files ship with your skill (`resource-enumeration-behavior`, `eager-link-resolution`, `unrecognized-directory-handling`). If a file matters, instruct the model to read it. A mention is not a load, and a listing is not a load either.

**Treat scripts as an enhancement with a prose fallback.** The spec presents `scripts/` as code agents can run, but whether that holds depends on the platform's permission posture, not on your skill. Our bundled script ran and returned output on Codex and Antigravity; on Claude Code and Copilot CLI, the permission system blocked it in a headless session and the model could only report the denial (`bundled-script-execution`). If a step exists only inside a script, describe in the body what the script does and what to do when it cannot run. Interactive users may be able to approve the command where headless sessions cannot.

**Name shared resource files distinctively.** Two active skills both shipping `references/API.md` did not shadow each other in our runs, but only because the models qualified every path themselves. The platforms' handling of a truly ambiguous path went unexercised. Distinct filenames remove the risk outright (`cross-skill-resource-shadowing`).

## Composition and lifecycle

**Skill-to-skill chains work today, but guards are the model's, not the platform's.** Three-skill invocation chains completed on all platforms, in English and Japanese, and prose-expressed dependencies were resolved (`invocation-depth-limit`, `invocation-language-sensitivity`, `informal-dependency-resolution`). Circular references were stopped by the model choosing to stop, not by any platform mechanism, so do not design skills that rely on a platform catching a cycle (`circular-invocation-handling`).

**Missing dependencies fail visibly.** When a skill references an uninstalled skill, every tested platform surfaced the failure. This manifested as an explicit tool error on Claude Code and Copilot CLI, and the model reporting the absence on Codex and Antigravity. None hallucinated compliance in our runs, which is the good outcome. But the failure is still a runtime discovery your user makes, not something any platform checks at install time (`missing-dependency-behavior`, `cross-scope-dependency`).

**Editing a skill mid-session takes effect on the next skill activation.** All tested platforms served fresh content after SKILL.md was edited during a session (`reactivation-freshness`). New skills do not become discoverable until a new session, because catalog enumeration occurs at runtime. But editing existing skills becomes actionable within the session upon re-activating the skill. However, repeated activations of a large skill carry a token cost across all tested harnesses; Claude Code and Copilot CLI re-inject the full body on every reactivation as platform behavior, and Codex and Antigravity models re-read the full file by choice in our runs rather than relying on the copy already in context (`reactivation-deduplication`).
