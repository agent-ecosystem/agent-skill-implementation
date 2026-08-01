# Platform Behavior Checks

**Check list version: 0.2** ([changelog](#changelog))

How do agent platforms actually load skill content? The [Agent Skills specification](https://agentskills.io/specification) defines a file format and recommends a three-tier "progressive disclosure" model, but gives platforms wide latitude in implementation. The [client implementation guide](https://agentskills.io/client-implementation/adding-skills-support) provides more detailed guidance, but claims it was derived from analysis of 7 of 25+ adopting platforms and published months after most platforms had already shipped their implementations.

This raises a question: does platform behavior actually vary, and if so, how? Skill authors currently have no way to know what will happen when their skill is activated on a given platform. This page catalogs the behaviors that need empirical testing to find out. The checks began with loading behavior and have grown to cover validation strictness, script execution, and access control.

## Background

The spec's progressive disclosure section (present since launch, Dec 18, 2025) recommends a three-tier structure:

1. **Metadata** (~100 tokens): `name` and `description` loaded at startup
2. **Instructions** (< 5,000 tokens recommended): Full `SKILL.md` body loaded on activation
3. **Resources** (as needed): Files in `scripts/`, `references/`, `assets/` loaded when required

The client implementation guide (rewritten Mar 5, 2026 in [PR #200](https://github.com/agentskills/agentskills/pull/200)) elevated this from a structural recommendation to "the core principle" of implementation, stating "every skills-compatible agent follows the same three-tier loading strategy." That guide was developed from analysis of [seven implementations](https://github.com/agentskills/agentskills/pull/200): OpenCode, Pi, Gemini CLI, Codex, VS Code Copilot Chat, Goose, and OpenHands. At least 25 platforms had adopted Agent Skills before that guide was published.

Until these questions are answered empirically across platforms, we cannot assume uniform behavior.

A note on quotes: the spec and the client implementation guide are unversioned and change without a changelog. Quoted language in this document was verified against both as of **2026-08-01**. If a quote no longer appears, check whether it moved, was reworded, or was removed; each of those is a data point about the spec's stability.

## Check Structure

Each check has:

- **ID**: A short identifier (e.g., `discovery-reading-depth`).
- **Category**: The area of platform behavior it evaluates.
- **What it checks**: A description of what the check evaluates.
- **Why it matters**: The observed or anticipated agent behavior that motivates the check.

Because these checks require empirical testing on each platform, we must record results per-platform.

---

## Category 1: Loading Timing

These checks evaluate *when* skill content enters the agent's context, and how much is loaded at each stage.

### `discovery-reading-depth`

- **Category**: Loading Timing
- **What it checks**: Whether the platform reads only SKILL.md frontmatter at discovery time, or reads the entire file (or more).
- **Why it matters**: The progressive disclosure model assumes platforms read only `name` and `description` at startup. If a platform reads the full SKILL.md body for every installed skill at session start, "inactive" skills are already consuming context. For a user with 20 installed skills at ~5,000 tokens each, that could mean ~100,000 tokens of instructions loaded before any conversation begins, leaving less room for the actual task.

### `activation-loading-scope`

- **Category**: Loading Timing
- **What it checks**: Whether the platform loads only the SKILL.md body when a skill is activated, or also loads some or all files from supporting directories.
- **Why it matters**: The spec says resources are loaded "as needed" and the guide says platforms should "not eagerly read" bundled resources. But a platform that loads the SKILL.md body plus all files in `scripts/`, `references/`, and `assets/` at activation time could inject thousands of tokens of content the model doesn't need for the current task. A skill with a 20-file `references/` directory would behave very differently on a platform that loads everything vs. one that waits for the model to request specific files.

### `eager-link-resolution`

- **Category**: Loading Timing
- **What it checks**: Whether the platform parses the SKILL.md body for markdown links (e.g., `[API errors](references/api-errors.md)`) and pre-fetches the linked files at activation time.
- **Why it matters**: Pre-fetching linked files is a reasonable engineering decision ("the skill references this file, so the model will probably need it"), but it collapses tier 2 and tier 3 of progressive disclosure into a single load event. A skill author who carefully structured their skill for on-demand loading (keeping SKILL.md lean and putting detail in referenced files) gets no benefit from that structure on platforms that pre-fetch. The context cost at activation becomes unpredictable for the skill author.

---

## Category 2: Directory Recognition

These checks evaluate which directories a platform recognizes as part of a skill, and how it handles directories it doesn't recognize.

### `recognized-directory-set`

- **Category**: Directory Recognition
- **What it checks**: Whether the platform builds loading or enumeration behavior specifically around the three spec-defined optional directories (`scripts/`, `references/`, `assets/`), and whether it treats them as a closed set or an open set.
- **Why it matters**: The spec originally documented only three optional directories. PR #216 (Mar 10, 2026) added a `... # Any additional files or directories` entry, but platforms that implemented before that change may treat the directory set as closed. A skill that puts content in `templates/` or `data/` may find that content is invisible to the model on platforms that only enumerate the three spec-defined directories. Conversely, platforms that treat the set as open may load unexpected content from directories the skill author didn't intend to be part of the skill's active context.

### `directory-naming-divergence`

- **Category**: Directory Recognition
- **What it checks**: Whether the platform looks for directories by the exact names in the spec, or uses different names for the same purpose (e.g., `resources/` instead of `references/`).
- **Why it matters**: If a platform's skill loader looks for `resources/` and a skill uses the spec-defined `references/`, the skill's reference content may be invisible on that platform. The skill author followed the spec, the platform followed its own convention, and the result is a silent failure where the model can't find supporting content that exists. This is particularly insidious because neither the skill author nor the platform developer did anything obviously wrong.

### `unrecognized-directory-handling`

- **Category**: Directory Recognition
- **What it checks**: What happens when a skill contains a directory that the platform doesn't recognize (e.g., `evals/`, `ci/`, `testing/`).
- **Why it matters**: The spec now says additional directories are permitted, but gives no guidance on how platforms should handle them. Some platforms may ignore them entirely (content is invisible). Some may load their contents into context (unexpected token cost). Some may present them to the model as part of the skill's file listing (the model may explore them during codebase navigation). The lack of guidance means every platform made its own choice, and skill authors can't predict which behavior they'll get.

---

## Category 3: Resource Access Patterns

These checks evaluate how platforms make supporting files (scripts, references, assets) available to the model.

### `resource-enumeration-behavior`

- **Category**: Resource Access Patterns
- **What it checks**: When a skill has a `references/` directory containing multiple files, whether the platform enumerates all files to the model, loads all file contents, presents a listing without loading contents, or ignores them until the model explicitly reads one.
- **Why it matters**: The guide says platforms should enumerate but "not eagerly read" resources. But the spec's progressive disclosure section simply says resources are loaded "as needed," which an implementer could reasonably interpret as "load all resources when the skill needs them" (i.e., at activation). The difference matters because a skill with 15 reference files could add anywhere from zero to tens of thousands of tokens at activation depending on the platform's approach.

### `path-resolution-base`

- **Category**: Resource Access Patterns
- **What it checks**: When SKILL.md references a file like `scripts/deploy.sh`, what the platform resolves that path against: the skill directory, the current working directory, or the project root.
- **Why it matters**: Incorrect path resolution means the model fails to find a file that exists, or loads an unrelated file from elsewhere in the project. The spec says to "use relative paths from the skill root," but doesn't define what "skill root" means in the context of platform path resolution. If the platform resolves against the working directory instead of the skill directory, a skill that works when the user is in the project root may break when they're in a subdirectory.

### `cross-skill-resource-shadowing`

- **Category**: Resource Access Patterns
- **What it checks**: If two active skills both contain a file at the same relative path (e.g., both have `references/API.md`), which one the model receives when it requests that file.
- **Why it matters**: The spec and guide don't address this scenario. If resource paths are bare relative paths rather than qualified with the skill name, the platform must decide which skill's file to return. Ambiguous resolution means a skill may receive another skill's reference content, leading to incorrect behavior that's extremely difficult to diagnose. This becomes more likely as users install more skills.

### `path-traversal-boundary`

- **Category**: Resource Access Patterns
- **What it checks**: Whether the model can access files outside the skill directory via relative paths (e.g., `../../other-file.md` or `../other-skill/SKILL.md`).
- **Why it matters**: Path traversal from a skill directory is both a security concern and a correctness concern. A skill that can read files outside its own directory could access sensitive project files or other skills' content. Platforms that don't enforce a boundary at the skill directory root expose users to potential prompt injection from malicious skills. Neither the spec nor the guide addresses path traversal validation explicitly. The guide's security guidance focuses on trust gating for project-level skills and assumes the agent's file-read tool enforces its own boundaries.

### `resource-nesting-depth`

- **Category**: Resource Access Patterns
- **What it checks**: Whether the platform enumerates and allows access to deeply nested resource files vs. only top-level entries in each directory, and how deep access extends. The probe has rungs at one, two, three, and five directory levels (e.g., `references/api/v2/history/deprecated/removed-endpoints.md`), so findings can report an actual depth bound rather than a yes/no.
- **Why it matters**: The spec's File references section says "Keep file references one level deep from `SKILL.md`. Avoid deeply nested reference chains", wording that reads as guidance about chains of references between files. It provides no guidance for how deep a directory tree platforms must support. Separately, the client guide suggests discovery scans use "reasonable bounds (e.g., max depth of 4-6 levels)." That bound is written for skill *discovery*, but an implementer could plausibly reuse it for resource enumeration or access. Skills that organize reference content hierarchically, such as API documentation with versioned endpoints, may find that nested files are invisible on platforms that bound depth. The skill works on platforms with deep access and loses content on shallower ones, with no error to diagnose. And because the spec is unversioned, platforms implemented against different snapshots may have internalized different depth expectations.

### `bundled-script-execution`

- **Category**: Resource Access Patterns
- **What it checks**: Whether the agent can actually run a bundled `scripts/` file and receive its output. The probe's script assembles its output phrase at runtime, so reading the source never reveals the phrase; only genuine execution produces it.
- **Why it matters**: The spec presents `scripts/` as "executable code that agents can run," with supported languages left to the implementation. Whether a platform's permission posture lets a skill's script run determines whether script-shipping skills work there at all: a skill that leans on its scripts degrades to prose-only, with no error anywhere, on a platform that blocks execution. Headless and interactive sessions may also differ, since interactive users can approve a command that headless policy refuses.

---

## Category 4: Content Presentation

These checks evaluate what the model actually sees, at discovery and at activation, and how it's formatted.

### `discovery-listing-fields`

- **Category**: Content Presentation
- **What it checks**: Which frontmatter fields the platform's discovery listing surfaces to the model: `name` and `description` only, or also `compatibility`, `metadata` values, file locations, or other fields.
- **Why it matters**: The guide says the catalog holds name and description (~50-100 tokens per skill), but platforms decide what actually goes in it. This determines what a skill author can rely on the model knowing *before* activation. A `compatibility` warning that only exists in frontmatter is invisible at selection time on a platform that lists name and description alone, so the model may activate a skill it cannot actually run. Combined with `frontmatter-handling` (what survives activation), this check catalogs which frontmatter fields reach the model through platform channels. On a platform that surfaces only name/description at discovery and strips frontmatter at activation, every other field carries no meaning to the model unless the model reads the raw file. Listings that include file locations give the model a path to read more, on its own initiative.

### `frontmatter-handling`

- **Category**: Content Presentation
- **What it checks**: Whether the platform passes the full SKILL.md file to the model (including YAML frontmatter) or strips the frontmatter and passes only the markdown body.
- **Why it matters**: The model sees different content on different platforms for the same skill. Frontmatter fields like `allowed-tools` and `compatibility` may influence model behavior on platforms that pass them through, and be invisible on platforms that strip them. A skill author who puts important context in the `compatibility` field, such as "requires Python 3.14+ and network access," may find that information reaches the model on one platform and is dropped on another.

### `content-wrapping-format`

- **Category**: Content Presentation
- **What it checks**: Whether the platform wraps skill content in structured tags (e.g., `<skill_content name="...">`) or injects it as raw markdown into the conversation context.
- **Why it matters**: Wrapping affects the model's ability to distinguish skill instructions from conversation history and other context. It also affects whether the platform can identify and protect skill content during context compaction. A skill that relies on the model treating its instructions as authoritative may find that instructions are confused with or overridden by conversation content on platforms that don't wrap. This is particularly relevant when multiple skills are active simultaneously.

---

## Category 5: Lifecycle Management

These checks evaluate how platforms manage skill content over the course of a conversation.

### `reactivation-deduplication`

- **Category**: Lifecycle Management
- **What it checks**: If the model activates the same skill a second time in a conversation (e.g., because it forgot it already loaded it, or because context was compacted), whether the platform deduplicates or injects the content again.
- **Why it matters**: Without deduplication, the same skill instructions can appear multiple times in context, wasting tokens and potentially confusing the model with redundant instructions. The guide recommends deduplication but frames it as optional ("consider tracking"). On platforms without deduplication, long conversations that trigger context compaction may see the model repeatedly re-activate skills, progressively consuming more of the context window with duplicate content.

### `reactivation-freshness`

- **Category**: Lifecycle Management
- **What it checks**: On re-activation (whether deduplicated or not), whether the platform re-reads the SKILL.md from disk or serves a cached copy from session start.
- **Why it matters**: If a skill author edits SKILL.md during an active conversation (a common workflow during skill development), platforms that cache at discovery time won't pick up the changes until the next session. Platforms that re-read on activation will reflect changes immediately. The guide explicitly calls this out as a design trade-off ("storing it makes activation faster; reading it at activation time... picks up changes to skill files between activations") but doesn't recommend one approach over the other, so behavior may vary across platforms.

### `context-compaction-protection`

- **Category**: Lifecycle Management
- **What it checks**: Whether the platform protects skill content from being pruned or summarized when the context window fills up.
- **Why it matters**: The guide says to "exempt skill content from pruning" because losing instructions "degrades the agent's performance without any visible error." But the guide only recommends this; nothing requires platforms to comply. On platforms that don't protect skill content, a long conversation may lose the skill's instructions partway through. The model continues operating but without the specialized guidance, producing subtly worse output. This is one of the hardest failures for a skill author to diagnose because there's no error; the skill just stops working mid-conversation.

---

## Category 6: Access Control

These checks evaluate how platforms gate skill loading and how control-related frontmatter fields are handled.

### `trust-gating-behavior`

- **Category**: Access Control
- **What it checks**: Whether the platform requires explicit trust approval for project-level skills (those found in the repository being worked on, which may be untrusted).
- **Why it matters**: Project-level skills come from the repository, which could be a freshly cloned open-source project. Without trust gating, cloning a repository injects skill instructions into the agent's context. Platforms could handle this in various ways: requiring explicit trust approval, gating based on workspace trust settings, or loading all skills with no trust check. If platforms diverge here, the same skill in the same repository may load on one platform and be blocked on another, with no signal to the skill author about what happened.

### `compatibility-field-behavior`

- **Category**: Access Control
- **What it checks**: How the platform handles the `compatibility` frontmatter field. Specifically: whether it parses the field for structured requirements, whether it uses the field to gate loading, and whether it surfaces the field to the model or user.
- **Why it matters**: The `compatibility` field is free-text with no structured format. The spec's own example (`Designed for Claude Code (or similar products)`) demonstrates platform-specific targeting in a supposedly platform-neutral format. A non-Claude platform encountering this field has no guidance on what to do: skip the skill entirely, warn the user, load it and hope for the best, or pass the text to the model and let it decide. Each choice produces different behavior. Skill authors who use this field to signal real requirements ("Requires Python 3.14+ and network access") can't predict whether that information will be acted on, displayed, or ignored.

### `allowed-tools-behavior`

- **Category**: Access Control
- **What it checks**: Whether the experimental `allowed-tools` field pre-approves the listed tools, measured against an identical control skill that has no such field. If the instructed command runs in both sessions, the platform's general permission posture, not the field, allowed it; if it runs only for the field-bearing twin, the field did the pre-approving.
- **Why it matters**: The spec marks `allowed-tools` as "Experimental. Support for this field may vary between agent implementations." Authors who rely on it for frictionless tool use need to know whether it does anything on their platform: where it is ignored, a skill's commands hit the normal permission flow; where it works, a skill can pre-approve tools the user never individually reviewed, which makes this an access-control behavior as much as a convenience.

---

## Category 7: Skill-to-Skill Invocation

These checks evaluate whether and how a skill can instruct the model to activate another skill.

The spec does not address skill-to-skill invocation at all. The guide describes a "subagent delegation" pattern as an advanced option, but frames it as the harness delegating to a subagent, not one skill invoking another. Three open issues on the spec repo have requested clarification since January 2026 ([#95](https://github.com/agentskills/agentskills/issues/95), [#100](https://github.com/agentskills/agentskills/issues/100), [#137](https://github.com/agentskills/agentskills/issues/137)).

Some early signals suggest platform behavior may already diverge. Claude Code's system prompt includes the phrase "Do not invoke a skill that is already running," which implies some awareness of invocation but leaves open how cross-skill invocation is handled. Issue [#95](https://github.com/agentskills/agentskills/issues/95) reports that GitHub Copilot does not appear to restrict skill-to-skill invocation. These are only scattered individual observations; empirical testing is needed to characterize each platform's actual behavior.

### `cross-skill-invocation`

- **Category**: Skill-to-Skill Invocation
- **What it checks**: Whether a skill's instructions can direct the model to activate a different installed skill by name.
- **Why it matters**: Skill composition is a common need. A `/review-and-commit` skill that chains `/review` then `/commit` is a natural pattern. But if the platform prevents or doesn't support cross-skill invocation, this pattern doesn't hold; the model either ignores the instruction, says it can't do it, or hallucinates compliance without actually activating the second skill. Skill authors who build composite skills have no way to know which platforms will support the pattern.

### `invocation-depth-limit`

- **Category**: Skill-to-Skill Invocation
- **What it checks**: Whether there is a limit on how many levels deep skill-to-skill invocation can go (skill A activates skill B, which activates skill C).
- **Why it matters**: Without depth limits, a chain of skill invocations could consume the entire context window with layered instructions, or enter a non-terminating loop if two skills reference each other. Platforms need some bound on invocation depth to prevent runaway context consumption, but no guidance exists on what that bound should be or how to communicate it to skill authors.

### `circular-invocation-handling`

- **Category**: Skill-to-Skill Invocation
- **What it checks**: What happens when skill A's instructions direct the model to activate skill B, and skill B's instructions direct the model to activate skill A.
- **Why it matters**: Circular dependencies are the most basic failure mode for any dependency system, and the Agent Skills ecosystem has no mechanism to prevent them. Some platforms may have partial guards (e.g., Claude Code's system prompt says "Do not invoke a skill that is already running"), but it's unclear whether such guards would catch an A→B→A cycle where A has finished before B re-invokes it. Without robust platform-level detection, circular invocations could loop until the context window is exhausted or the model decides to stop.

### `invocation-language-sensitivity`

- **Category**: Skill-to-Skill Invocation
- **What it checks**: Whether skill-to-skill invocation reliability varies based on the language of the user's prompt or the skill's instructions.
- **Why it matters**: Issue [agentskills/agentskills#95](https://github.com/agentskills/agentskills/issues/95) reports that one user observed Claude Code's skill-to-skill invocation failing ~10% of the time with Japanese prompts, with the model saying "execution outside of the main conversation is not permitted," while English prompts worked reliably. If this observation holds up under broader testing, it would suggest the failure isn't a hard platform restriction but emergent model behavior influenced by prompt language. Skills authored in or tested with only one language may behave differently for users in other languages, and the skill author would have no way to diagnose this.

---

## Category 8: Skill Dependencies

These checks evaluate how platforms handle the concept of one skill depending on another, despite no formal dependency mechanism existing in the spec.

The spec defines six frontmatter fields (`name`, `description`, `license`, `compatibility`, `metadata`, `allowed-tools`). None support declaring dependencies. Nine issues filed between Dec 2025 and Mar 2026 requested dependency support ([#11](https://github.com/agentskills/agentskills/issues/11), [#90](https://github.com/agentskills/agentskills/issues/90), [#95](https://github.com/agentskills/agentskills/issues/95), [#100](https://github.com/agentskills/agentskills/issues/100), [#110](https://github.com/agentskills/agentskills/issues/110), [#133](https://github.com/agentskills/agentskills/issues/133), [#137](https://github.com/agentskills/agentskills/issues/137), [#226](https://github.com/agentskills/agentskills/issues/226), [#256](https://github.com/agentskills/agentskills/issues/256)). The maintainer position is that dependencies belong in a distribution-layer manifest rather than SKILL.md, but no distribution layer exists.

### `informal-dependency-resolution`

- **Category**: Skill Dependencies
- **What it checks**: If a skill's body text says something like "first activate the `code-review` skill" or "this skill requires the `linting` skill to be installed," whether the platform attempts to resolve and load the referenced skill.
- **Why it matters**: In the absence of a formal dependency mechanism, skill authors use prose instructions to express dependencies. Whether this works depends entirely on the model's willingness and the platform's support for skill-to-skill invocation (see Category 8). On platforms where cross-skill invocation works, the dependency is resolved at runtime by the model. On platforms where it doesn't, the skill's instructions are partially unfulfillable, and the model may skip the dependency or produce degraded output.

### `missing-dependency-behavior`

- **Category**: Skill Dependencies
- **What it checks**: What happens when a skill references another skill (by name in its body text) that is not installed on the platform.
- **Why it matters**: There is no mechanism for a skill to declare its dependencies, and no mechanism for a platform to check whether dependencies are satisfied before activation. When a skill says "activate the `code-review` skill" and that skill isn't installed, the model discovers this at runtime. The failure mode is platform-dependent: the model might say it can't find the skill (visible failure), skip the step (degraded output), or attempt to fulfill the instruction from its training data without the skill's specialized guidance (incorrect output that looks correct). None of these are good outcomes, and the skill author can't prevent any of them.

### `nonstandard-dependency-fields`

- **Category**: Skill Dependencies
- **What it checks**: Whether any platform recognizes dependency-related frontmatter fields that aren't in the spec (e.g., `requires`, `depends`, `prerequisites`).
- **Why it matters**: The spec doesn't define dependency fields, but that doesn't mean no platform has implemented them. If a platform added a `requires` field that other platforms ignore, skills using that field would have dependencies resolved on one platform and ignored on all others. Discovering nonstandard fields in active use is important for understanding the real ecosystem behavior and for informing whether the spec should standardize something.

### `cross-scope-dependency`

- **Category**: Skill Dependencies
- **What it checks**: How platforms handle the case where a skill at one scope (e.g., project-level) references a skill that only exists at a different scope (e.g., user-level), or vice versa.
- **Why it matters**: A project-level skill that depends on a user-level utility skill works for the author (who has both installed) but fails for a collaborator who only has the project-level skills. There's no mechanism to signal that a skill has cross-scope dependencies, no way for a platform to check for them, and no standard behavior for when they're missing. This is a portability trap: the skill works in the author's environment and degrades in everyone else's.

---

## Category 9: Discovery Scope

These checks evaluate where platforms look for skills: which directories are scanned, how deep the scan goes, and what happens when two discovered skills claim the same name. Every check in this category results in a skill that is present and listed on one platform but absent, or answering to a different identity, on another.

### `cross-client-directory-interop`

- **Category**: Discovery Scope
- **What it checks**: Whether the platform discovers skills installed at the cross-client `.agents/skills/` convention path when that is not its native skills directory.
- **Why it matters**: The guide recommends scanning both a client-native directory and the `.agents/skills/` convention, "so skills installed by other compliant clients are automatically visible to yours, and vice versa," and notes some clients also scan `.claude/skills/` pragmatically. The spec itself mandates nothing about where skills live. A platform that only scans its native directory breaks the interop story. A skill installed by one client is invisible to another, with no error anywhere. Users who maintain one shared skills directory need to know which platforms actually honor it.

### `recursive-root-discovery`

- **Category**: Discovery Scope
- **What it checks**: Whether the platform discovers skills only in direct children of its skills root (`<root>/<skill>/SKILL.md`), or scans the root recursively (finding e.g. `<root>/group/skill/SKILL.md`), and whether a spec-compliant SKILL.md placed entirely outside any recognized skills root is discovered.
- **Why it matters**: Teams with many skills naturally want to organize them in subfolders, such as by domain, by team, or by lifecycle stage. On a platform that scans recursively, that layout works. On a platform that only reads direct children, every grouped skill vanishes from the catalog with no error. The inverse risk also exists: a recursive scanner turns *every* SKILL.md under the root into an installable-looking, invocable skill (see `nested-skill-discovery`), which surprises authors who ship example or vendored skills as content. And if any platform scans the whole project tree rather than just its root, cloning a repository with documentation examples could register skills the user never installed, a trust-boundary concern the spec doesn't address.

### `nested-skill-discovery`

- **Category**: Discovery Scope
- **What it checks**: If a SKILL.md exists deeper inside another skill's directory tree (e.g., `my-skill/references/helper-skill/SKILL.md`), whether the platform discovers both the outer and inner skills.
- **Why it matters**: The guide says to scan for "subdirectories containing a file named exactly `SKILL.md`," which would match both the outer skill and any nested SKILL.md files. A skill author who stores another skill inside their `references/` directory (perhaps as documentation or a dependency) may unintentionally register that inner skill as a separate available skill on the platform. This could lead to unexpected skill activations, name collisions, or confusing listings.

### `name-collision-precedence`

- **Category**: Discovery Scope
- **What it checks**: When two installed skills share the same `name` at different scopes (project-level and user-level) with different content, which one activates.
- **Why it matters**: The guide states the universal convention is that "project-level skills override user-level skills." If a platform resolves the other way, or nondeterministically, the same activation loads *different instructions* depending on platform. In this portability failure scenario, everything appears to work. This also matters for security reasoning: project-wins means a cloned repository can shadow a user's trusted skill of the same name.

---

## Category 10: Validation Strictness

These checks feed platforms skills that break the spec's format rules (invalid YAML, missing or oversize fields, rule-breaking names, out-of-spec metadata values) and observe the response: rejected, repaired, or loaded anyway. The spec binds skill authors here but says nothing about what platforms should do with a file that breaks the rules, so every platform chose its own posture. A skill that loads on a lenient platform vanishes with no error on a strict one.

### `malformed-yaml-tolerance`

- **Category**: Validation Strictness
- **What it checks**: Whether a skill whose frontmatter is technically invalid YAML, like the common unquoted-colon description (`description: Use when: ...`), is still discovered and loadable.
- **Why it matters**: The guide acknowledges that "skill files authored for other clients may contain technically invalid YAML that their parsers happen to accept" and recommends a repair fallback. Strict parsers reject the file outright ("mapping values are not allowed here"), so the same skill loads on lenient platforms and vanishes on strict ones. Because the mistake is easy to make and many authors test on only one platform, this is one of the most likely real-world portability breaks.

### `missing-description-handling`

- **Category**: Validation Strictness
- **What it checks**: What happens to a skill with no `description` field: skipped (as the guide prescribes: "a description is essential for disclosure"), loaded with an empty or synthesized description, or handled some other way.
- **Why it matters**: The guide's lenient-validation table suggests warn-but-load for name violations, but skip entirely for a missing description. Platforms that load such skills anyway create catalogs where the model has a name with no guidance on when to use it. Platforms that skip them mean the skill isn't visible to agents.

### `invalid-name-tolerance`

- **Category**: Validation Strictness
- **What it checks**: Whether skills whose names break the spec's rules (uppercase letters, consecutive hyphens, more than 64 characters) are discovered, normalized, or rejected. Each fixture's frontmatter name matches its directory name exactly, so the character or length rule is the only violation in play.
- **Why it matters**: The spec states the name rules as hard constraints, but they only bind if platforms enforce them. A rule-breaking name that loads on a lenient platform vanishes without an error on a strict one, and a platform that repairs the name (say, by lowercasing) changes the identity that documentation, users, and other skills refer to. Authors who test on one platform never see the break coming.

### `name-directory-mismatch`

- **Category**: Validation Strictness
- **What it checks**: When a skill's frontmatter `name` differs from the name of the directory it lives in, which identity the platform uses: is the skill listed and invocable under the frontmatter name, under the directory name, or rejected outright?
- **Why it matters**: The spec requires the skill's directory name to match its frontmatter `name`, but platforms decide independently whether to enforce, ignore, or partially honor the rule. Renaming a skill in frontmatter without renaming its folder (or vice versa, e.g. during development or after a fork) produces a mismatch that platforms have to resolve somehow. If one platform indexes by frontmatter name and another by directory name, the same installed skill answers to different names on different platforms. Instructions in other skills or documentation that reference it by one name would fail on platforms that chose the other. Early probing shows at least one platform (Codex) indexes purely by frontmatter name, so the failure mode is real rather than hypothetical.

### `metadata-value-edge-cases`

- **Category**: Validation Strictness
- **What it checks**: Whether platforms successfully parse and load a skill whose `metadata` frontmatter contains edge-case YAML values: empty strings (`''`, `""`), explicit nulls (`null`, `~`, `None`), and tagged nulls (`!!null null`).
- **Why it matters**: The spec defines `metadata` as a string-to-string mapping, but YAML parsers interpret `null`, `~`, and `None` as null values rather than strings. A platform whose loader expects all metadata values to be strings may throw a type error, drop the affected keys, or fail to load the skill entirely. Since the spec doesn't explicitly prohibit null values and YAML makes them easy to produce accidentally (an author who writes `foo:` with no value gets null rather than an empty string), these edge cases will appear in real-world skills. The question is whether each platform handles them gracefully or breaks.

### `oversize-description-handling`

- **Category**: Validation Strictness
- **What it checks**: Whether a skill whose description exceeds the spec's 1024-character cap (this fixture's runs to 1116) is discovered, and whether the value survives intact or truncated. The description carries a head marker near its start and a tail marker as its final characters, and the body never repeats either, so what surfaces reveals exactly how much of the value survived.
- **Why it matters**: Every installed skill's description occupies context at startup, which gives platforms a real incentive to cap or truncate. Rejection makes the skill vanish on strict platforms; silent truncation quietly deletes the end of the description, which is often where the "use when" activation cues live. Either way, an author with a long description gets different discovery behavior per platform without any error.

### `oversize-compatibility-handling`

- **Category**: Validation Strictness
- **What it checks**: Whether a skill whose `compatibility` value exceeds the spec's 500-character cap (this fixture's runs to 570, with a tail marker) is discovered and loadable.
- **Why it matters**: `compatibility` is optional and informational, so this check reveals how strictly platforms validate a field most skills omit entirely. A platform that rejects the skill outright turns harmless verbosity into a portability break; one that accepts it demonstrates the cap is documentation rather than enforcement.

---
## Benchmark Skills

The [`benchmark-skills/`](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills) directory contains 33 benchmark fixtures (spec-compliant skills plus deliberate structural and validation edge cases) designed to exercise these checks. Each skill contains unique **canary phrases** (e.g., CARDINAL-ZEBRA-7742) embedded in specific files. By asking the model whether it knows a canary phrase, testers can determine exactly what the platform loaded and when, without relying on the model's self-reporting about its own context.

See [`benchmark-skills/README.md`](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/benchmark-skills/README.md) for:

- The full skill inventory and what each skill tests
- A **check-to-skill mapping** showing which skill(s) to use for each check and the recommended test procedure
- A **canary phrase index** listing every canary phrase and which file it lives in

## Contributing

We need empirical data from people testing on real platforms. If you can test any of these checks on a specific platform, please open a PR with your findings using the template at [`platform-findings/template.md`](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/platform-findings/template.md).

Even partial data is valuable. A single platform tested thoroughly is more useful than speculation about all of them.

## Changelog

Findings submissions record the check list version they were tested against (see the template's "Check list version" field), so readers can tell which checks existed when a platform was tested.

### 0.2 (2026-08-01)

- Added `name-directory-mismatch` and `recursive-root-discovery` (Category 7: Structural Edge Cases), prompted by observed platform divergence in nested-skill discovery: one platform registers and invokes any SKILL.md found recursively under its skills root, indexed by frontmatter name alone.
- Added Category 10: Discovery and Validation (`cross-client-directory-interop`, `malformed-yaml-tolerance`, `missing-description-handling`, and `name-collision-precedence`), derived from portability-sensitive behaviors the client implementation guide prescribes but platforms adopted independently.
- Added `discovery-listing-fields` (Category 4: Content Presentation), prompted by the observation that one platform's discovery listing is strictly `name: description` lines while another's includes file locations, which changes what frontmatter can ever reach the model.
- Reworded `resource-nesting-depth`: the spec's "one level deep" language (now in its File references section) reads as guidance about reference chains rather than directory depth. The probe gained a five-levels-deep rung so findings report a measured depth bound instead of a yes/no.
- Added a note that spec/guide quotes in this document were verified as of 2026-08-01, since both upstream documents are unversioned and change without a changelog.
- Added `bundled-script-execution` (Category 3), `allowed-tools-behavior` (Category 6), and `invalid-name-tolerance`, `oversize-description-handling`, and `oversize-compatibility-handling` (Category 10), closing the spec statements a spec-alignment audit of the automated reports found untested: script executability, the experimental `allowed-tools` field, and the name, description, and compatibility limits.
- Restructured the categories (check IDs unchanged): dissolved Structural Edge Cases (`resource-nesting-depth` joined Resource Access Patterns; `nested-skill-discovery` and `recursive-root-discovery` joined the new Discovery Scope; `name-directory-mismatch` joined the new Validation Strictness), split Discovery and Validation into Discovery Scope (where platforms look) and Validation Strictness (how strictly they judge what they find), and moved `metadata-value-edge-cases` from Content Presentation to Validation Strictness to sit with the other spec-invalid-input checks.

### 0.1 (2026-03-22)

- Initial check list: 28 checks across Categories 1-9 (loading timing, directory recognition, resource access patterns, content presentation, lifecycle management, access control, structural edge cases, skill-to-skill invocation, and skill dependencies).
