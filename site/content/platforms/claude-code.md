---
title: "Claude Code (headless)"
description: "Automated skill loading findings for Claude Code (headless) (check list 0.2)."
date: 2026-08-01
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | Claude Code (headless) |
| **Platform version** | 2.1.212 |
| **Check list version** | 0.2 |
| **Test date** | 2026-08-01 |
| **Model(s) observed** | claude-sonnet-5 |
| **Environment** | Headless invocation via benchmark-runner + skillxp |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors they are point observations, not rates. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised" — automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Loading Timing

### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: Metadata only (`metadata-only`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-loading
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources — and by which vehicle?_

- **Status**: observed
- **Verdict**: Body only (`body-only`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - body canary in harness-injected content
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body — and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: No pre-fetching (`no-prefetch`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - skill body loaded
  - references/setup-guide.md arrived only via the model's own read
  - references/troubleshooting.md arrived only via the model's own read
  - references/unlinked-data.md arrived only via the model's own read
- **Note**: model read [references/setup-guide.md references/troubleshooting.md references/unlinked-data.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Directory Recognition

### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration-at-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`resources-untouched`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/) — injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`untouched`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: ## Report: Nonstandard Directories Probe  **1. Directory awareness**  When the skill activated, I was only given the `SKILL.md` body text (injected as the tool result) plus a "Base directory" path. No directory listing or file enumeration was provided automatically — I was not told what was inside `evals/`, `templates/`, or `resources/` until I ran `find` myself via Bash. After that manual step, I became aware of:  - `evals/evals.json` - `resources/api-reference.md` - `templates/output-template.md` - `SKILL.md` (given directly)  **2. Content visibility**  No file contents were pre-loaded or auto-surfaced for any of these three directories. I can see filenames (once I explicitly list them), but not contents — I would need to explicitly `Read` each file to see what's inside `evals.json`, `api-reference.md`, or `output-template.md`. Nothing was silently loaded into context the way `SKILL.md` itself was.  **3. Resources vs. references**  I observed no special handling of `resources/` at all — it was not enumerated, previewed, or treated differently from `templates/` or `evals/`. All three nonstandard directories were treated identically: invisible until I searched for them manually, an…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Resource Access Patterns

### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against — and does the bare path work as written?_

- **Status**: observed
- **Verdict**: Bare path fails; model recovers (`cwd-base-model-requalified`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Runs**: 3 (cwd-base-model-requalified ×3)
- **Evidence**:
  - bare SKILL.md-relative path attempt did not deliver content
  - bare SKILL.md-relative path attempt did not deliver content
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Note**: verdict consistent across 3 runs
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: Got its own file (`own-resource-first`)
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Evidence**:
  - activated skill's own API.md content arrived first
  - sibling skill's API.md also read
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised — the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: ## Results  1. **Alpha skill's own file** (`.claude/skills/probe-shadow-alpha/references/API.md`): canary phrase **STORK-CORAL-4471** — correctly resolved to the alpha version.  2. **Beta skill's file** (`.claude/skills/probe-shadow-beta/references/API.md`): I *was* able to access it directly by specifying its full absolute path, and it correctly returned its own distinct content with canary **EGRET-SLATE-8823** — no cross-contamination.  **Conclusion**: no ambiguous resolution occurred. Each skill's `references/API.md` resolves to its own distinct absolute path (`.claude/skills/probe-shadow-alpha/references/API.md` vs. `.claude/skills/probe-shadow-beta/references/API.md`), and the Read tool correctly retrieved the right content for each when given the right path. There's no shared/ambiguous namespace causing one skill's resource lookup to silently pull in the other's file — accessing the sibling skill's file required explicitly using its full distinct path, not an implicit relative lookup from within alpha's context.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: Reads outside the skill allowed (`outside-skill-read-allowed`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - sibling skill's body canary arrived in a tool result
  - first ../ path attempt
- **Note**: relative ../ path forms were used
- **Note**: final answer: Now I have enough information to compile the report.  ## Path Traversal Probe Results  **1. Sibling skill (`../probe-loading/SKILL.md`)** - **Allowed — succeeded.** The Read tool resolved the `..` traversal and returned the full contents of the sibling skill's `SKILL.md` (including its canary phrase `CARDINAL-ZEBRA-7742` and full instructions). No warning or error was shown. This means file-level tools like Read do **not** enforce a boundary at the skill directory root — they'll follow `..` to sibling directories.  **2. Parent directory (`../README.md` and listing `../`)** - **Read**: returned "File does not exist" — but this is because no `README.md` actually exists there (confirmed by listing `.claude/skills/` directly, which shows only `probe-loading` and `probe-traversal`), not because of a boundary block. - **Listing via `ls ../`**: **Blocked**, with an explicit security error: *"Path contains '..' traversal after a directory segment, which may follow a symlink outside the working directory."* This is the Bash tool actively refusing the traversal syntax itself, regardless of whether the target exists.  **3. Two levels up (`../../loading-behavior.md`)** - **Read**: returned "Fi…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Content Presentation

### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: Name and description only (`name-and-description-only`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names the installed skills
  - description text surfaced at discovery
- **Note**: final answer: Here is everything shown to me in the Available skills list, reproduced verbatim:  - **probe-compatibility**: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior. - **probe-loading**: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior. - **probe-metadata-values**: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling. - **deep-research**: Deep research harness — fan-out web searches, fetch sources, adversarially verify claims, synthesize a cited report. - When the user wants a deep, multi-source, fact-checked research report on any topic. BEFORE invoking, check if the question is specific enough to research directly — if underspecified (e.g., "what car to buy" without budget/use-case/region), ask 2-3 clarifying questions to narrow scope. Then pass the refined question as args, weaving the answers in. - **dataviz**: Use this skill whenever you are about to create ANY chart, graph, plot, dashboard, or data visualizati…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: Stripped before injection (`frontmatter-stripped-on-injection`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - body injected without the frontmatter-only marker
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded — and do those keys reach the model?_

- **Status**: observed
- **Verdict**: Loaded fine (`loaded-despite-edge-case-metadata`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-metadata-values despite edge-case metadata
  - body canary loaded
  - frontmatter-only marker '!!null' in a tool result
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: ## Probe Results  **1. Load success:** Yes. The skill loaded and its instructions rendered fine, including the canary phrase **THRUSH-FLINT-8294**. The edge-case `metadata` values did not cause a rejection.  **2. Metadata visibility:** The `metadata` field was **not surfaced to me anywhere in the Skill tool output**. When I invoked the skill, I only received the body content (everything from `# Metadata Value Edge Cases Probe` onward) — no frontmatter, and no `metadata` block. The earlier skill-listing system-reminder also only exposed `name` and a one-line description, not `metadata`.  The only way I could see the raw values at all was by directly reading the file with the `Read` tool, bypassing the skill-loading mechanism entirely. Via `Read`, the raw YAML shows:  | Key | Raw YAML |  |---|---| | `normal` | `"baseline"` | | `empty-double` | `""` | | `empty-single` | `''` | | `explicit-null` | `null` | | `tilde-null` | `~` | | `none-null` | `None` (this is just the literal string `"None"` in YAML — not a null token) | | `tagged-null` | `!!null null` |  **3. Key preservation:** All seven keys are present in the source file. But since the platform's Skill-invocation path never expose…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown — and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: Raw markdown, no wrapper tags (`raw-injection`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - injection event carrying the body canary
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Lifecycle Management

### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: Full content re-injected every time (`reinjected-each-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - first injection of body canary
  - second injection of body canary
- **Note**: platform-level: duplicate skill content occupies context after reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: Edits picked up immediately (`fresh-content-served`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - post-edit canary reached the model on reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `context-compaction-protection`

_Is skill content protected when the context window fills up?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

## Access Control

### `trust-gating-behavior`

_Do project-level skills require trust approval before loading?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

### `compatibility-field-behavior`

_Does a compatibility field naming another platform gate loading, get surfaced to the model, or get ignored?_

- **Status**: observed
- **Verdict**: Loads normally, no gating (`activated-no-gating`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value
  - body loaded normally
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: ## Probe Results  **1. Field visibility:** Yes — I can see it directly in the frontmatter by reading the file. It was not, however, surfaced to me automatically in the skill-invocation content the harness injected (that only contained the body text, not the frontmatter fields). I had to `cat` the file myself to find: ``` compatibility: Designed for Claude Code (or similar products). Requires Python 3.14+ and network access. ```  **2. Platform behavior:** No special handling observed. The skill activated normally — no warning, no confirmation prompt, and no automated check of whether Python 3.14+ or network access is actually available. The `compatibility` field appears to be inert metadata as far as the harness is concerned; nothing gated activation on it.  **3. My interpretation:** I can't verify this from inside the sandbox with confidence: - "Claude Code (or similar products)" — plausibly satisfied, since this session is running in a Claude-Code-like harness. - "Python 3.14+" — unverified and not obviously relevant to this skill (the skill's own instructions involve no Python execution). I could check `python3 --version` on the host, but that wouldn't confirm what environment th…
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

## Structural Edge Cases

### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: Not discovered (`nested-skill-not-discovered`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names only the outer skill
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: All depths reachable (tested to 5) (`all-depths-accessible-through-5`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - depth-1 file references/overview.md content arrived
  - depth-2 file references/api/endpoints.md content arrived
  - depth-3 file references/api/v2/migration-guide.md content arrived
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Directory name wins (`directory-name-identity`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the DIRECTORY name probe-mismatch-dir
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity — the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  1. **Identity**: Same result as before — the catalog lists it as `probe-mismatch-dir` (directory name), even though the frontmatter declares `probe-name-mismatch`. Both invocation names (`probe-name-mismatch` and `probe-mismatch-dir`) successfully resolved to this same skill.  2. **Validation**: Again, accepted silently — no warning or error about the mismatch, and both the frontmatter name and directory name work as valid identifiers to activate it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Direct children only; stray file ignored (`direct-children-only; stray:not-discovered`)
- **Confidence**: transcript-direct
- **Note**: final answer: I don't have a skill called "probe-grouped" in my available skills list — I can't invoke skills that aren't listed. Here's what's actually available to me:  deep-research, dataviz, update-config, keybindings-help, verify, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, run, init, review, security-review  If you meant one of these, let me know. Otherwise, if "probe-grouped" is something you expect to exist (e.g., a custom project skill), it may not be registered or discoverable in this session.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Skill-to-Skill Invocation

### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: The chain breaks here: `invoke-gamma` is not an available skill (only `invoke-alpha` and `invoke-beta` were listed to me), so it cannot be activated.  invoke-beta complete. invoke-gamma said: N/A — invocation failed with "Unknown skill: invoke-gamma."  invoke-alpha complete. Chain result: invoke-alpha (canary IBIS-RUST-3310) successfully activated invoke-beta (canary TERN-MOSS-6647), but invoke-beta's attempt to activate invoke-gamma failed because that skill isn't registered/available in this environment. The chain terminated at the beta→gamma link.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "invoke-gamma activated. Canary: JAY-TEAL-9984" and confirmed the end of chain was reached, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), then activated invoke-gamma, which activated (Canary: JAY-TEAL-9984) and confirmed the full chain — invoke-alpha → invoke-beta → invoke-gamma — completed successfully. All three canaries were reported in sequence: IBIS-RUST-3310, TERN-MOSS-6647, JAY-TEAL-9984.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Beta's instructions direct me to activate alpha again, which would recreate the exact same cycle indefinitely (alpha → beta → alpha → beta → ...). Continuing to follow this literally would never terminate, so I'm breaking the loop here rather than re-invoking alpha a second time.  **Summary of the probe:** - probe-circular-alpha activated successfully, canary `KITE-ONYX-2251` emitted. - It invoked probe-circular-beta, which activated successfully, canary `WREN-SLATE-7738` emitted. - Beta's instructions call for re-activating alpha, closing the loop. - No platform-level guard fired to block the second invocation (the Skill tool doesn't appear to detect cycles automatically) — the loop was only stopped because I judged that mechanically re-entering it would run forever with no new information gained.  So: circular invocation is *possible* (nothing stopped step 3 in either skill from executing), and detection/prevention is currently a matter of agent judgment rather than a platform safeguard.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates — agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "invoke-gamma activated. Canary: JAY-TEAL-9984" and confirmed the end of chain was reached, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), which activated invoke-gamma (Canary: JAY-TEAL-9984). Gamma confirmed the end of the chain was reached and all three skills activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Skill Dependencies

### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-gamma isn't registered as an available skill, so it can't be activated — the call returned "Unknown skill: invoke-gamma." The chain stops here.  invoke-beta complete. invoke-gamma could not be activated (unknown/unregistered skill), so no result to report from it.  invoke-alpha complete. Chain result: invoke-beta activated successfully (canary TERN-MOSS-6647), but invoke-gamma is not available in this environment's skill registry, so the chain broke at the second link — invoke-gamma never ran.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: Failed with a visible error (`attempted-visible-failure`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-missing-dep body loaded
  - attempt on the missing skill returned an error result
- **Note**: final answer: The platform explicitly reported that the skill doesn't exist — it returned a tool error (`Unknown skill: nonexistent-formatter`) rather than silently skipping or letting me improvise the formatting from general knowledge. I did not attempt to fulfill the formatting task myself.  probe-missing-dep complete. Dependency resolution result: the harness rejected the activation attempt at the tool-call level with an explicit "Unknown skill" error, since `nonexistent-formatter` isn't in the available skills list — no silent failure, no fallback to general knowledge.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: ## Report: Nonstandard Frontmatter Fields Probe  **1. Field visibility** — Yes, by reading the SKILL.md file directly I can see all three nonstandard fields and their values: - `requires: probe-loading` - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` - `priority: high`  However, these fields were **not** surfaced to me through the normal skill-invocation channel. The `<system-reminder>` listing of available skills only exposed the standard `name` and `description` fields (as a one-line summary per skill). I only saw `requires`, `depends-on`, and `priority` because I explicitly opened the file with the Read tool — the skill-loading mechanism itself didn't pass them into context.  **2. Platform behavior** — None of the nonstandard fields triggered any automatic action: - `requires: probe-loading` did **not** cause `probe-loading` to be auto-loaded or activated. - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` did **not** cause either shadow skill to be loaded. - `priority: high` had no observable effect on ordering or how this skill was presented relative to others in the skill list.  All three fields were silently ignored by the platform's skill-loading logic.  **3. V…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level — and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: Resolved across scopes; missing dependency fails visibly (`resolved-across-scopes; missing:visible-failure`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - user-scope dependency's body canary injected during project-skill session
  - attempt on the missing dependency returned an error result
- **Note**: with-dependency final answer: `probe-loading` was found and activated successfully — resolved to a **user-level** location: `/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-671745375/home/claude/skills/probe-loading`, distinct from the project-level path (`..…
- **Note**: without-dependency final answer: `probe-loading` was not found. The Skill tool rejected it outright with `Unknown skill: probe-loading` — it isn't present in the available-skills listing at all (neither project nor user scope), so the platform didn't even attempt a scope-specific lookup; it just failed fast as an unrecognized skill name.  probe-cross-scope complete. Cross-scope resolution result: the project-level skill's reference to a user-level skill did **not** resolve — `probe-loading` is absent from the resolved skill set entirely, and invoking it errors immediately with "Unknown skill" rather than finding it at a different scope or giving a scope-specific not-installed message. This suggests either no user-level skill of that name exists in this environment, or the platform does not merge user-level skills into the resolvable set alongside project-level ones (at least not visibly to the model).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Discovery and Validation

### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path not scanned (`convention-not-scanned`)
- **Confidence**: transcript-direct
- **Note**: a skill installed by another client at .agents/skills is invisible here
- **Note**: final answer: I don't have a skill named "probe-interop" in my available skills list — it's not one of the skills shown to me (deep-research, dataviz, update-config, keybindings-help, verify, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, run, init, review, security-review). I can't invoke skills that aren't in that list, and I won't guess at or fabricate one. Let me know if you meant one of the above, or if this should be added to the project first.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Tolerated and loaded (`tolerated-and-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-malformed-yaml
  - body canary loaded on activation
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility**: In my available-skills catalog, this skill was listed as:  `probe-malformed-yaml: Use when: asked to probe malformed yaml parsing behavior`  The description appeared intact and readable — not truncated at the colon, and I have no way to tell from my view whether the underlying YAML was silently repaired (e.g., auto-quoted) before reaching me, since I only see the final rendered text, not the raw frontmatter source.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-missing-description`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-no-description
  - body canary loaded on activation
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, it appears in my available-skills list, with an empty/placeholder description. The entry showed only the title text "Missing Description Probe" — the same string used as the skill's markdown heading in this file — with no actual description content synthesized or written alongside it, unlike every other skill in the list which has a substantive one-line description explaining when to use it.  For reference, this behavior contradicts what the skill's own text claims the client implementation guide recommends (skipping skills without a description and logging an error) — I loaded and activated it anyway rather than skipping it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: User scope wins (`user-overrides-project`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - user variant's canary loaded; project variant's never appeared
- **Note**: contradicts the guide's 'universal convention' that project-level overrides user-level
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **PIPIT-SHALE-1147** (user variant)  **Precedence**: The **user**-scope variant won over the project-scope variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.2) for what each check evaluates.
