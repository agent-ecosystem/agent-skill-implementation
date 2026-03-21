# Skill Loading Behavior

How do agent platforms actually load skill content? The [Agent Skills specification](https://agentskills.io/specification) defines a file format and recommends a three-tier "progressive disclosure" model, but gives platforms wide latitude in implementation. The [client implementation guide](https://agentskills.io/client-implementation/adding-skills-support) provides more detailed guidance, but was derived from analysis of 7 of 25+ adopting platforms and published months after most platforms had already shipped their implementations.

This raises a question: does loading behavior actually vary across platforms, and if so, how? Skill authors currently have no way to know what will happen when their skill is activated on a given platform. This page catalogs the behaviors that need empirical testing to find out.

## Background

The spec's progressive disclosure section (present since launch, Dec 18, 2025) recommends a three-tier structure:

1. **Metadata** (~100 tokens): `name` and `description` loaded at startup
2. **Instructions** (< 5,000 tokens recommended): Full `SKILL.md` body loaded on activation
3. **Resources** (as needed): Files in `scripts/`, `references/`, `assets/` loaded when required

The client implementation guide (rewritten Mar 5, 2026 in [PR #200](https://github.com/agentskills/agentskills/pull/200)) elevated this from a structural recommendation to "the core principle" of implementation, stating "every skills-compatible agent follows the same three-tier loading strategy." That guide was developed from analysis of [seven implementations](https://github.com/agentskills/agentskills/pull/200): OpenCode, Pi, Gemini CLI, Codex, VS Code Copilot Chat, Goose, and OpenHands. At least 25 platforms had adopted Agent Skills before that guide was published.

Until these questions are answered empirically across platforms, we cannot assume uniform behavior.

## Check Structure

Each check has:

- **ID**: A short identifier (e.g., `discovery-reading-depth`).
- **Category**: The area of loading behavior it evaluates.
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
- **Why it matters**: The guide says platforms should enumerate but "not eagerly read" resources. But the spec's progressive disclosure section simply says resources are loaded "as needed," which an implementer could reasonably interpret as "load all resources when the skill needs them" (i.e., at activation). The difference matters: a skill with 15 reference files could add anywhere from zero to tens of thousands of tokens at activation depending on the platform's approach.

### `path-resolution-base`

- **Category**: Resource Access Patterns
- **What it checks**: When SKILL.md references a file like `scripts/deploy.sh`, what the platform resolves that path against: the skill directory, the current working directory, or the project root.
- **Why it matters**: Incorrect path resolution means the model silently fails to find a file that exists, or loads an unrelated file from elsewhere in the project. The spec says to "use relative paths from the skill root," but doesn't define what "skill root" means in the context of platform path resolution. If the platform resolves against the working directory instead of the skill directory, a skill that works when the user is in the project root may break when they're in a subdirectory.

### `cross-skill-resource-shadowing`

- **Category**: Resource Access Patterns
- **What it checks**: If two active skills both contain a file at the same relative path (e.g., both have `references/API.md`), which one the model receives when it requests that file.
- **Why it matters**: The spec and guide don't address this scenario. If resource paths are bare relative paths rather than qualified with the skill name, the platform must decide which skill's file to return. Ambiguous resolution means a skill may silently receive another skill's reference content, leading to incorrect behavior that's extremely difficult to diagnose. This becomes more likely as users install more skills.

### `path-traversal-boundary`

- **Category**: Resource Access Patterns
- **What it checks**: Whether the model can access files outside the skill directory via relative paths (e.g., `../../other-file.md` or `../other-skill/SKILL.md`).
- **Why it matters**: Path traversal from a skill directory is both a security concern and a correctness concern. A skill that can read files outside its own directory could access sensitive project files or other skills' content. Platforms that don't enforce a boundary at the skill directory root expose users to potential prompt injection from malicious skills. Neither the spec nor the guide addresses path traversal validation explicitly; the guide's security guidance focuses on trust gating for project-level skills and assumes the agent's file-read tool enforces its own boundaries.

---

## Category 4: Content Presentation

These checks evaluate what the model actually sees when a skill is activated, and how it's formatted.

### `frontmatter-handling`

- **Category**: Content Presentation
- **What it checks**: Whether the platform passes the full SKILL.md file to the model (including YAML frontmatter) or strips the frontmatter and passes only the markdown body.
- **Why it matters**: The model sees different content on different platforms for the same skill. Frontmatter fields like `allowed-tools` and `compatibility` may influence model behavior on platforms that pass them through, and be invisible on platforms that strip them. A skill author who puts important context in the `compatibility` field ("requires Python 3.14+ and network access") may find that information reaches the model on one platform and is silently dropped on another.

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
- **Why it matters**: The guide says to "exempt skill content from pruning" because losing instructions "silently degrades the agent's performance without any visible error." But this is a recommendation, not a requirement. On platforms that don't protect skill content, a long conversation may silently lose the skill's instructions partway through. The model continues operating but without the specialized guidance, producing subtly worse output. This is one of the hardest failures for a skill author to diagnose because there's no error; the skill just stops working mid-conversation.

---

## Category 6: Access Control

These checks evaluate how platforms gate skill loading and how control-related frontmatter fields are handled.

### `trust-gating-behavior`

- **Category**: Access Control
- **What it checks**: Whether the platform requires explicit trust approval for project-level skills (those found in the repository being worked on, which may be untrusted).
- **Why it matters**: Project-level skills come from the repository, which could be a freshly cloned open-source project. Without trust gating, cloning a repository silently injects skill instructions into the agent's context. Platforms could handle this in various ways: requiring explicit trust approval, gating based on workspace trust settings, or loading all skills with no trust check. If platforms diverge here, the same skill in the same repository may load on one platform and be blocked on another, with no signal to the skill author about what happened.

### `compatibility-field-behavior`

- **Category**: Access Control
- **What it checks**: How the platform handles the `compatibility` frontmatter field. Specifically: whether it parses the field for structured requirements, whether it uses the field to gate loading, and whether it surfaces the field to the model or user.
- **Why it matters**: The `compatibility` field is free-text with no structured format. The spec's own example (`Designed for Claude Code (or similar products)`) demonstrates platform-specific targeting in a supposedly platform-neutral format. A non-Claude platform encountering this field has no guidance on what to do: skip the skill entirely, warn the user, load it and hope for the best, or pass the text to the model and let it decide. Each choice produces different behavior. Skill authors who use this field to signal real requirements ("Requires Python 3.14+ and network access") can't predict whether that information will be acted on, displayed, or silently ignored.

---

## Category 7: Structural Edge Cases

These checks evaluate platform behavior with skill directory structures that push beyond the spec's core examples.

### `nested-skill-discovery`

- **Category**: Structural Edge Cases
- **What it checks**: If a SKILL.md exists deeper inside another skill's directory tree (e.g., `my-skill/references/helper-skill/SKILL.md`), whether the platform discovers both the outer and inner skills.
- **Why it matters**: The guide says to scan for "subdirectories containing a file named exactly `SKILL.md`," which would match both the outer skill and any nested SKILL.md files. A skill author who stores another skill inside their `references/` directory (perhaps as documentation or a dependency) may unintentionally register that inner skill as a separate available skill on the platform. This could lead to unexpected skill activations, name collisions, or confusing listings.

### `resource-nesting-depth`

- **Category**: Structural Edge Cases
- **What it checks**: Whether the platform enumerates and allows access to deeply nested resource files (e.g., `references/api/v2/endpoints.md`) vs. only top-level entries in each directory.
- **Why it matters**: The spec recommends "keep file references one level deep from `SKILL.md`" but this is a recommendation, not a constraint. Skills that organize reference content hierarchically (common for API documentation with versioned endpoints) may find that nested files are invisible on platforms that only enumerate the top level of each directory. The skill works on platforms with deep enumeration and silently loses content on platforms with shallow enumeration, with no error to diagnose.

---

## Category 8: Skill-to-Skill Invocation

These checks evaluate whether and how a skill can instruct the model to activate another skill.

The spec does not address skill-to-skill invocation at all. The guide describes a "subagent delegation" pattern as an advanced option, but frames it as the harness delegating to a subagent, not one skill invoking another. Three open issues on the spec repo have requested clarification since January 2026 ([#95](https://github.com/agentskills/agentskills/issues/95), [#100](https://github.com/agentskills/agentskills/issues/100), [#137](https://github.com/agentskills/agentskills/issues/137)).

Some early signals suggest platform behavior may already diverge. Claude Code's system prompt includes the phrase "Do not invoke a skill that is already running," which implies some awareness of invocation but leaves open how cross-skill invocation is handled. Issue [#95](https://github.com/agentskills/agentskills/issues/95) reports that GitHub Copilot does not appear to restrict skill-to-skill invocation. These are individual observations, not systematic findings; empirical testing is needed to characterize each platform's actual behavior.

### `cross-skill-invocation`

- **Category**: Skill-to-Skill Invocation
- **What it checks**: Whether a skill's instructions can direct the model to activate a different installed skill by name.
- **Why it matters**: Skill composition is a common need. A `/review-and-commit` skill that chains `/review` then `/commit` is a natural pattern. But if the platform prevents or doesn't support cross-skill invocation, this pattern silently fails: the model either ignores the instruction, says it can't do it, or hallucinates compliance without actually activating the second skill. Skill authors who build composite skills have no way to know which platforms will support the pattern.

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

## Category 9: Skill Dependencies

These checks evaluate how platforms handle the concept of one skill depending on another, despite no formal dependency mechanism existing in the spec.

The spec defines six frontmatter fields (`name`, `description`, `license`, `compatibility`, `metadata`, `allowed-tools`). None support declaring dependencies. Nine issues filed between Dec 2025 and Mar 2026 requested dependency support ([#11](https://github.com/agentskills/agentskills/issues/11), [#90](https://github.com/agentskills/agentskills/issues/90), [#95](https://github.com/agentskills/agentskills/issues/95), [#100](https://github.com/agentskills/agentskills/issues/100), [#110](https://github.com/agentskills/agentskills/issues/110), [#133](https://github.com/agentskills/agentskills/issues/133), [#137](https://github.com/agentskills/agentskills/issues/137), [#226](https://github.com/agentskills/agentskills/issues/226), [#256](https://github.com/agentskills/agentskills/issues/256)). The maintainer position is that dependencies belong in a distribution-layer manifest rather than SKILL.md, but no distribution layer exists.

### `informal-dependency-resolution`

- **Category**: Skill Dependencies
- **What it checks**: If a skill's body text says something like "first activate the `code-review` skill" or "this skill requires the `linting` skill to be installed," whether the platform attempts to resolve and load the referenced skill.
- **Why it matters**: In the absence of a formal dependency mechanism, skill authors use prose instructions to express dependencies. Whether this works depends entirely on the model's willingness and the platform's support for skill-to-skill invocation (see Category 8). On platforms where cross-skill invocation works, the dependency is resolved at runtime by the model. On platforms where it doesn't, the skill's instructions are partially unfulfillable, and the model may silently skip the dependency or produce degraded output.

### `missing-dependency-behavior`

- **Category**: Skill Dependencies
- **What it checks**: What happens when a skill references another skill (by name in its body text) that is not installed on the platform.
- **Why it matters**: There is no mechanism for a skill to declare its dependencies, and no mechanism for a platform to check whether dependencies are satisfied before activation. When a skill says "activate the `code-review` skill" and that skill isn't installed, the model discovers this at runtime. The failure mode is platform-dependent: the model might say it can't find the skill (visible failure), silently skip the step (degraded output), or attempt to fulfill the instruction from its training data without the skill's specialized guidance (incorrect output that looks correct). None of these are good outcomes, and the skill author can't prevent any of them.

### `nonstandard-dependency-fields`

- **Category**: Skill Dependencies
- **What it checks**: Whether any platform recognizes dependency-related frontmatter fields that aren't in the spec (e.g., `requires`, `depends`, `prerequisites`).
- **Why it matters**: The spec doesn't define dependency fields, but that doesn't mean no platform has implemented them. If a platform added a `requires` field that other platforms ignore, skills using that field would have dependencies resolved on one platform and silently ignored on all others. Discovering nonstandard fields in active use is important for understanding the real ecosystem behavior and for informing whether the spec should standardize something.

### `cross-scope-dependency`

- **Category**: Skill Dependencies
- **What it checks**: How platforms handle the case where a skill at one scope (e.g., project-level) references a skill that only exists at a different scope (e.g., user-level), or vice versa.
- **Why it matters**: A project-level skill that depends on a user-level utility skill works for the author (who has both installed) but fails for a collaborator who only has the project-level skills. There's no mechanism to signal that a skill has cross-scope dependencies, no way for a platform to check for them, and no standard behavior for when they're missing. This is a portability trap: the skill works in the author's environment and silently degrades in everyone else's.

---

## Benchmark Skills

The [`benchmark-skills/`](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills) directory contains 16 spec-compliant skills designed to exercise these checks. Each skill contains unique **canary phrases** (e.g., CARDINAL-ZEBRA-7742) embedded in specific files. By asking the model whether it knows a canary phrase, testers can determine exactly what the platform loaded and when, without relying on the model's self-reporting about its own context.

See [`benchmark-skills/README.md`](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/benchmark-skills/README.md) for:

- The full skill inventory and what each skill tests
- A **check-to-skill mapping** showing which skill(s) to use for each check and the recommended test procedure
- A **canary phrase index** listing every canary phrase and which file it lives in

## Contributing

We need empirical data from people testing on real platforms. If you can test any of these checks on a specific platform, please open a PR with your findings using the template at [`platform-loading-implementation/template.md`](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/platform-loading-implementation/template.md).

Even partial data is valuable. A single platform tested thoroughly is more useful than speculation about all of them.
