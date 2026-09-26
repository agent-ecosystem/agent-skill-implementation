# Platform Loading Implementation: Claude Code (headless)

| | |
|---|---|
| **Platform** | Claude Code (headless) |
| **Platform version** | 2.1.274 |
| **Check list version** | 0.4 |
| **Test date** | 2026-09-26 |
| **Model(s) observed** | claude-sonnet-5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 21 of the 46 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

- [`path-resolution-base`](#path-resolution-base): The spec tells authors to reference files with relative paths from the skill root, but a path written that way fails here: paths resolve against the session's working directory, not the skill directory. In this run the model noticed the failure and requalified the path itself.
- [`bundled-script-execution`](#bundled-script-execution): The spec presents scripts/ as executable code agents can run, but execution was blocked in this headless run. Interactive sessions, where a user can approve the command, may behave differently.
- [`frontmatter-handling`](#frontmatter-handling): The spec says the agent loads the entire SKILL.md file at activation. This platform strips the YAML frontmatter and injects only the body, so frontmatter fields beyond name and description never reach the model.

### Where behavior matches the spec

- [`discovery-reading-depth`](#discovery-reading-depth): Discovery reads only the skill's metadata, matching the spec's progressive disclosure model: name and description load at startup, and the body waits for activation.
- [`activation-loading-scope`](#activation-loading-scope): Activation loads the full SKILL.md body and nothing more, matching the spec's second disclosure stage: instructions at activation, resources only as a task needs them.
- [`eager-link-resolution`](#eager-link-resolution): Files linked from SKILL.md are not pre-fetched at activation; they load only when the task calls for them, which is the spec's on-demand model for resources.
- [`resource-enumeration-behavior`](#resource-enumeration-behavior): Reference files stay out of context until the model asks for them, matching the spec's rule that resources load on demand.
- [`resource-nesting-depth`](#resource-nesting-depth): The spec advises authors to keep file references one level deep but sets no platform limit, and none was observed: reference files stayed reachable at every tested depth through five levels.
- [`discovery-listing-fields`](#discovery-listing-fields): The discovery listing carries name and description and nothing else, exactly the fields the spec says load at startup.
- [`compatibility-field-behavior`](#compatibility-field-behavior): The spec makes compatibility informational (it indicates environment requirements) and assigns it no loading semantics. Consistent with that, a skill declaring a different product still loads here; authors should not expect the field to gate anything.
- [`description-length-unit`](#description-length-unit): The spec caps description at 1024 characters without defining the unit. This platform does not enforce the limit at all, so every fixture under 1024 code points loaded intact and the counting unit is moot.
- [`name-length-unit`](#name-length-unit): The spec caps name at 64 characters without defining the unit and allows unicode lowercase alphanumeric characters with an ASCII parenthetical, which reads two ways; its skills-ref reference validator accepts any Unicode alphanumeric and counts code points. This platform does not enforce the cap at all, so the counting unit is moot.

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this fixture's frontmatter does not parse (an unquoted colon). The platform tolerated the error: the skill is discovered and loads anyway.
- [`missing-description-handling`](#missing-description-handling): The spec requires a non-empty description, so a skill without one is invalid. This platform discovered and loaded it anyway.
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform tolerated every one: each rule-breaking name is discovered and usable.
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway, under the directory name.
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated.
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway.

### Not exercised in this run

- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. The field's own effect went unobserved: the instructed command ran with and without it, so the platform's general permission posture is what allowed execution.
- [`allowed-tools-name-matching`](#allowed-tools-name-matching): The spec marks allowed-tools experimental, leaves tool names to each platform, and says support may vary, so no outcome contradicts it. Every spelling's command ran under the platform's general permission posture, so the field's matching rule went unobserved.


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: `metadata-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/discovery-reading-depth/ab4267cd-efce-417d-8489-0ccbead99da0.jsonl`; session ab4267cd-efce-417d-8489-0ccbead99da0):
  - discovery listing names probe-loading (event 6, line 8)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/activation-loading-scope/9102342a-95b1-4b93-8bf5-c901a93b79a1.jsonl`; session 9102342a-95b1-4b93-8bf5-c901a93b79a1):
  - body canary in harness-injected content (event 15, line 21)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/eager-link-resolution/0848e457-8c9c-438f-acda-bbf7cdc3346a.jsonl`; session 0848e457-8c9c-438f-acda-bbf7cdc3346a):
  - skill body loaded (event 14, line 20)
  - references/setup-guide.md arrived only via the model's own read (event 28, line 37)
  - references/troubleshooting.md arrived only via the model's own read (event 30, line 39)
  - references/unlinked-data.md arrived only via the model's own read (event 32, line 41)
- **Note**: model read [references/setup-guide.md references/troubleshooting.md references/unlinked-data.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Directory Recognition

#### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: `no-enumeration-at-activation`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: `resources-untouched`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: `readable-on-demand:[evals/ templates/]`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/unrecognized-directory-handling/5dc6bb7d-1ec5-4e2b-a4f5-df17baf6f516.jsonl`; session 5dc6bb7d-1ec5-4e2b-a4f5-df17baf6f516):
  - evals/ content arrived via the model's own read (event 26, line 33)
  - templates/ content arrived via the model's own read (event 30, line 37)
- **Note**: final answer: ## Report: Nonstandard Directories Probe  **1. Directory awareness**  Only `SKILL.md` was injected automatically when the skill activated. I was not given an automatic enumeration of `evals/`, `resources/`, or `templates/` — I only know they exist because I ran `find` myself and then chose to `Read` each file explicitly. Nothing about them appeared in my context until I went looking.  **2. Content visibility**  For all three directories, no listing or content was pre-loaded — I had zero visibility until I actively called `Bash`/`find` (for listings) and `Read` (for contents). This was true uniformly regardless of directory name: `evals/`, `templates/`, and `resources/` were all treated identically as "invisible until fetched."  Contents found (each has a distinct canary confirming I read the actual file rather than inferring): - `evals/evals.json` — canary `ROBIN-JADE-3847` - `resources/api-reference.md` — canary `SWIFT-OPAL-8156` - `templates/output-template.md` — canary `WREN-PEARL-6293`  **3. Resources vs. references**  No special-casing observed. `resources/` was not auto-loaded the way a spec-defined `references/` directory is sometimes described as being surfaced — it require…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: `cwd-base-model-requalified`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/path-resolution-base/912a9099-da94-44b6-b4cf-a16d60083092.jsonl`; session 912a9099-da94-44b6-b4cf-a16d60083092):
  - bare SKILL.md-relative path attempt did not deliver content (event 20, line 29)
  - bare SKILL.md-relative path attempt did not deliver content (event 22, line 31)
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/cross-skill-resource-shadowing/e429772e-6237-43a6-8818-ee5519621a2b.jsonl`; session e429772e-6237-43a6-8818-ee5519621a2b):
  - activated skill's own API.md content arrived first (event 22, line 30)
  - sibling skill's API.md also read (event 24, line 32)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: ## Probe Results  **probe-shadow-alpha's `references/API.md`:** contains canary **STORK-CORAL-4471** — this is alpha's own version, as expected.  **probe-shadow-beta's `references/API.md`:** I was able to read it directly too (by using its full absolute path, `.claude/skills/probe-shadow-beta/references/API.md`), and it correctly returned its own canary, **EGRET-SLATE-8823** — beta's own content, not alpha's.  No cross-contamination: each skill's `references/API.md` resolved to its own distinct file with its own canary phrase. Path resolution correctly used skill-scoped absolute paths rather than an ambiguous shared/relative lookup, so there's no sign of the platform conflating the two skills' resources.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/path-traversal-boundary/23b6ddb6-c2b5-4e10-89e5-7b7da8d119cd.jsonl`; session 23b6ddb6-c2b5-4e10-89e5-7b7da8d119cd):
  - sibling skill's body canary arrived in a tool result (event 22, line 31)
- **Note**: no ../ form observed; the model reached the sibling via absolute or listing-derived paths. Access is unrestricted either way
- **Note**: final answer: ## Results  **1. Sibling skill (`../probe-loading/SKILL.md`)** — **Allowed, full success.** The Read tool returned the complete file contents (frontmatter, canary phrase `CARDINAL-ZEBRA-7742`, full body) with no error, warning, or boundary message. No traversal restriction was enforced.  **2. Parent directory (`../README.md`)** — **Error: file not found**, not blocked. I confirmed via `ls` that `.claude/skills/` genuinely contains no `README.md` (only the two skill directories exist). The error was a legitimate "does not exist," not a permission/boundary denial. Listing `../` itself (via `ls`) succeeded and showed sibling skill directories with no restriction.  **3. Two levels up (`../../checks.md`)** — **Error: file not found**, not blocked. Confirmed via `ls` that `.claude/` contains only the `skills/` subdirectory — no `checks.md` exists there. Again, this was a genuine absence, not an enforced boundary.  ## Summary  The platform does **not** enforce any boundary at the skill directory root. Read access to `../probe-loading/SKILL.md` succeeded fully via a relative path from outside the skill's own directory. The other two "misses" were due to the target files simply not existing…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/resource-nesting-depth/066eee41-4b22-46ea-a91d-abe181e28fc0.jsonl`; session 066eee41-4b22-46ea-a91d-abe181e28fc0):
  - depth-1 file references/overview.md content arrived (event 22, line 30)
  - depth-2 file references/api/endpoints.md content arrived (event 24, line 32)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 26, line 34)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 28, line 36)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 30, line 38)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `execution-blocked-visibly`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/bundled-script-execution/beda7e04-b59f-4807-bf67-fa45b61bcea6.jsonl`; session beda7e04-b59f-4807-bf67-fa45b61bcea6):
  - script-referencing tool call returned an error result (event 20, line 29)
- **Note**: final answer: The script execution was blocked — it requires approval that wasn't granted.  **Result:** Execution failed. Error: `This command requires approval`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: final answer: probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  **File awareness**: I was not told about any files belonging to this skill when it was activated. The activation message contained only the skill's instructions text above (naming no files) — no file listing, tree, or attachment metadata accompanied it.  **Count**: 0
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `name-and-description-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/discovery-listing-fields/ddabd2a7-ad51-4303-94cc-50b27a1f5c83.jsonl`; session ddabd2a7-ad51-4303-94cc-50b27a1f5c83):
  - discovery listing names the installed skills (event 6, line 8)
  - description text surfaced at discovery (event 6, line 8)
- **Note**: final answer: Here is every skill exactly as shown in my Available skills list:  - **probe-compatibility**: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior. - **probe-loading**: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior. - **probe-metadata-values**: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling. - **dataviz**: Use this skill whenever you are about to create ANY chart, graph, plot, dashboard, or data visualization, in ANY output medium — an HTML or React artifact, inline SVG, plotting code in any library (matplotlib, plotly, d3, Recharts, …), an image/PNG you will render and upload, or a chart shared into Slack. Read it BEFORE writing the first line of chart code, choosing chart colors, building a stat tile / meter / KPI row, or laying out a dashboard. When the destination is a first-party document connector (host-designated, never self-described) that renders live charts, hand it the rows (inline, or as an up…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-stripped-on-injection`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/frontmatter-handling/a00291ed-8a4f-4c29-b73b-4b42c7cdda8d.jsonl`; session a00291ed-8a4f-4c29-b73b-4b42c7cdda8d):
  - body injected without the frontmatter-only marker (event 15, line 21)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-injection`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/content-wrapping-format/f9ef153a-9413-4206-bda5-209f8388a54f.jsonl`; session f9ef153a-9413-4206-bda5-209f8388a54f):
  - injection event carrying the body canary (event 15, line 21)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: `path-stated-at-activation`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/activation-location-disclosure/450bc0af-6938-433d-9af5-f58da94aaa68.jsonl`; session 450bc0af-6938-433d-9af5-f58da94aaa68):
  - skill path in harness-injected activation content (event 14, line 20)
- **Note**: final answer: **Probe Report**  1. **Frontmatter visibility**: No YAML frontmatter is visible to me. I only see a "Base directory" line followed directly by the Markdown body (starting at `# Loading Behavior Probe`). No `name`, `description`, `allowed-tools`, `compatibility`, or `metadata` fields are present in what I received — frontmatter appears to have been stripped before the content reached me (though the tool-listing system-reminder earlier did show a `name` + one-line `description` for this skill, separately from this activation payload).  2. **Content wrapping**: No XML/JSON structured wrapping around the body itself — it's plain Markdown text preceded by one plain-text line (`Base directory for this skill: ...`). The surrounding conversation uses `<system-reminder>` tags elsewhere, but this skill's content block itself is unwrapped Markdown.  3. **Resource awareness**: From the body text (not from enumeration), I can infer two files exist under `references/`: `references/api-overview.md` and `references/error-codes.md`, because they're linked inline. I have no independent directory listing — no other files/subdirectories (e.g., anything in `scripts/` or `assets/`) were enumerated or ma…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `reinjected-each-activation`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/reactivation-deduplication/701d60bd-c102-4086-958c-7275ff43ed3d.jsonl`; session 701d60bd-c102-4086-958c-7275ff43ed3d):
  - first injection of body canary (event 14, line 20)
  - second injection of body canary (event 25, line 36)
- **Note**: platform-level: duplicate skill content occupies context after reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/reactivation-freshness/43a8d29e-fbd6-428f-b419-b28c2d423817.jsonl`; session 43a8d29e-fbd6-428f-b419-b28c2d423817):
  - post-edit canary reached the model on reactivation (event 25, line 36)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `context-compaction-protection`

_Is skill content protected when the context window fills up?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

### Access Control

#### `trust-gating-behavior`

_Do project-level skills require trust approval before loading?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

#### `compatibility-field-behavior`

_Does a compatibility field naming another platform gate loading, get surfaced to the model, or get ignored?_

- **Status**: observed
- **Verdict**: `activated-no-gating`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/compatibility-field-behavior/0fe86e46-e629-4c92-9e78-65671d5d49af.jsonl`; session 0fe86e46-e629-4c92-9e78-65671d5d49af):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 6, line 8)
  - body loaded normally (event 14, line 20)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: Here's my report on the `compatibility` field probe:  1. **Field visibility**: Yes, I can see it directly by reading the SKILL.md frontmatter. Its value is:    `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`    Note that this wasn't surfaced to me automatically anywhere else (not in the skill listing shown to me, not in the invocation payload) — I only found it by reading the raw file.  2. **Platform behavior**: No special handling occurred. The skill was invoked and activated normally with no warning, no confirmation prompt, and no automated check against the stated requirements (Python 3.14+, network access). Nothing in the system verified whether this environment actually satisfies those requirements before or during activation.  3. **My interpretation**: I have no way to confirm the requirements are met — I don't have visibility into a Python version in this environment in any verified way, and I can't independently confirm network access is available for this skill's purposes. The `compatibility` field appears to be purely descriptive metadata: it's there for a human (or an agent that chooses to read and act on it) to interpret, not …
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/allowed-tools-behavior/session-2/0543f35b-b0de-44f3-9db8-569da552344c.jsonl`; session d1a7cb8d-49f9-4281-a2e9-aa70c75c640f,0543f35b-b0de-44f3-9db8-569da552344c):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 25, line 32)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 22, line 31)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  1. **allowed-tools value**: Yes, I can see it in the frontmatter of `SKILL.md`: `allowed-tools: Bash(printf:*) Read` — it pre-approves `Read` and `Bash` scoped to the `printf` command pattern.  2. **Command output** (ran exactly `printf 'GROUSE-%s-9017\n' 'MICA'`):    ```    GROUSE-MICA-9017    ```  3. **Permission prompt**: The command ran without any visible permission prompt or approval step — no interruption was surfaced to me before or during execution.
- **Note**: control final answer: Command output: `LAPWING-FLUORITE-2260`  The command executed, but a permission prompt/approval step was presented for it (this skill has no `allowed-tools` field, so nothing pre-approved the call) — it was not run silently without any approval step.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: `executed-regardless-of-spelling`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/allowed-tools-name-matching/session-3/4573a823-947e-4070-a9eb-727f0c28eec4.jsonl`; session d64b71fe-6337-4dc2-800d-243bdb1356e0,e1c5245a-2ca5-4b63-bcaa-30d67a99e7e6,4573a823-947e-4070-a9eb-727f0c28eec4):
  - spec-style twin: runtime-assembled phrase arrived in a tool result (session 1, event 21, line 29)
  - lowercase twin: runtime-assembled phrase arrived in a tool result (session 2, event 27, line 35)
  - shell twin: runtime-assembled phrase arrived in a tool result (session 3, event 22, line 30)
- **Note**: spec-style twin final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  **allowed-tools value:** Yes, visible in this skill's frontmatter: `Bash(printf:*) Read`  **Command output** (`printf 'GROUSE-%s-9017\n' 'MICA'`): ``` GROUSE-MICA-9017 ```  **Permission prompt:** The command ran without any visible permission prompt or approval step on my end.  Note: I can't fully distinguish whether that was due to the `allowed-tools` field specifically pre-approving `Bash(printf:*)`, or the session's general permission mode already allowing Bash. For a real comparison, you'd want to run the same command via `probe-allowed-tools-control` (no `allowed-tools` field) and see if it also proceeds without a prompt — if so, the field isn't the operative factor here.
- **Note**: lowercase twin final answer: Here's the report:  1. **probe-allowed-tools-lowercase activated. Canary: STILT-SCORIA-5526**  2. Yes — the skill's frontmatter declares `allowed-tools: bash` (bare lowercase, no arguments).  3. Command output (verbatim): ``` CHOUGH-TUFA-7180 ```  4. No permission prompt or approval step was surfaced to me for this command — it ran directly. (Note: I can't be fully certain whether that's because of the `allowed-tools: bash` declaration or because this session's permission mode already allows Bash commands generally — I don't have visibility into which one caused it.)
- **Note**: shell twin final answer: probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  1. **Allowed-tools value**: I don't see an `allowed-tools` field exposed to me for this skill — the skill content shown above doesn't surface its frontmatter/metadata to me, only the rendered instructions body.  2. **Command output**: I ran `printf 'ROOK-%s-8841\n' 'GABBRO'` and got: ``` ROOK-GABBRO-8841 ```  3. **Permission prompt**: The command ran without any visible permission prompt or approval step on my end — no interruption was surfaced to me during execution. (I can't confirm whether the user's client silently auto-approved it or whether a prompt appeared on their side that I wasn't shown.)
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/cross-skill-invocation/b4e51065-44dc-4673-b12c-fa97810e2546.jsonl`; session b4e51065-44dc-4673-b12c-fa97810e2546):
  - invoke-alpha body loaded (chain entry) (event 15, line 21)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 22, line 31)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-gamma is not available to invoke — it never appeared in my list of available skills, and calling it directly returned "Unknown skill: invoke-gamma". The chain stops here.  invoke-beta complete. invoke-gamma said: N/A — the skill is not available in this environment (Unknown skill error), so it could not be activated or report anything.  invoke-alpha complete. Chain result: invoke-alpha activated (canary IBIS-RUST-3310), which activated invoke-beta (canary TERN-MOSS-6647), but invoke-beta's attempt to activate invoke-gamma failed with "Unknown skill: invoke-gamma" — that skill isn't registered/available, so the chain terminated at depth 2 of 3.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/invocation-depth-limit/4798dbdf-11bc-4aa8-84a4-c3b94b3109c3.jsonl`; session 4798dbdf-11bc-4aa8-84a4-c3b94b3109c3):
  - invoke-alpha body loaded (event 14, line 20)
  - invoke-beta body loaded (event 21, line 30)
  - invoke-gamma body loaded (event 27, line 36)
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: End of chain reached — all three skills activated successfully, canary JAY-TEAL-9984.  invoke-alpha complete. Chain result: All three skills activated in sequence — invoke-alpha (IBIS-RUST-3310) → invoke-beta (TERN-MOSS-6647) → invoke-gamma (JAY-TEAL-9984). Chain completed successfully end-to-end.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  I'm stopping here rather than actually re-invoking probe-circular-alpha. Both skills instruct an unconditional call back to the other with no termination condition, so following them literally would create a genuine infinite loop rather than a bounded test. There's no platform-level circular-invocation guard visible to me here — the loop only stops because I'm declining to continue it.  **Summary of the probe:** 1. probe-circular-alpha activated → canary `KITE-ONYX-2251` 2. It invoked probe-circular-beta → canary `WREN-SLATE-7738` 3. Beta's instructions call back into alpha, which would call back into beta, indefinitely 4. I broke the cycle here instead of recursing forever  Both canaries were successfully captured, confirming each skill's instructions loaded correctly before I halted the recursion.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/invocation-language-sensitivity/137c576c-e637-40b4-9985-a50cf07982b4.jsonl`; session 137c576c-e637-40b4-9985-a50cf07982b4):
  - invoke-alpha body loaded (event 15, line 21)
  - invoke-beta body loaded (event 22, line 31)
  - invoke-gamma body loaded (event 28, line 37)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully." (Canary: JAY-TEAL-9984)  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), which activated invoke-gamma (Canary: JAY-TEAL-9984). All three skills in the chain fired successfully and confirmed end-of-chain.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/informal-dependency-resolution/a70ccfc9-8db6-4710-8745-29381dacd3fa.jsonl`; session a70ccfc9-8db6-4710-8745-29381dacd3fa):
  - invoke-alpha body loaded (chain entry) (event 14, line 20)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 21, line 30)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: The `invoke-gamma` skill isn't available to invoke — it's not in my list of activatable skills, and the call fails with "Unknown skill: invoke-gamma".  invoke-beta complete. invoke-gamma said: [not reachable — invoke-gamma is not an available skill in this environment, so the chain stops here.]  invoke-alpha complete. Chain result: invoke-beta activated successfully (Canary: TERN-MOSS-6647), but it could not activate invoke-gamma since that skill isn't registered/available — the chain broke at the beta→gamma link.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/missing-dependency-behavior/c4513820-7c43-4db7-a4af-d3c954d87cd6.jsonl`; session c4513820-7c43-4db7-a4af-d3c954d87cd6):
  - probe-missing-dep body loaded (event 15, line 21)
  - attempt on the missing skill returned an error result (event 21, line 30)
- **Note**: final answer: **What happened:** The platform explicitly rejected the call with a tool-use error: `Unknown skill: nonexistent-formatter`. It did not silently skip the step, and I did not attempt to fulfill the formatting task from general knowledge — the tool call itself was refused before any fallback behavior could occur.  probe-missing-dep complete. Dependency resolution result: The platform returned an explicit error (`Unknown skill: nonexistent-formatter`) when attempting to activate a skill that isn't in the available-skills listing, rather than silently no-oping or hallucinating the skill's behavior.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/nonstandard-dependency-fields/66c60a50-60ae-4b6c-896f-e7e733a6da32.jsonl`; session 66c60a50-60ae-4b6c-896f-e7e733a6da32):
  - probe-nonstandard-fields body loaded (event 14, line 20)
- **Note**: final answer: Here's my report on the nonstandard frontmatter probe:  **1. Field visibility** — Yes, I can see all three nonstandard fields by reading the raw SKILL.md file directly: - `requires: probe-loading` - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` - `priority: high`  However, these values were **not** surfaced to me automatically when the skill was invoked. The `Skill` tool only injected the markdown body (everything after the `---` closing delimiter) into the conversation — I had to separately `Read` the file to see the frontmatter fields at all.  **2. Platform behavior** — No automatic action was taken on any of these fields: - `requires: probe-loading` did not trigger a load of the `probe-loading` skill. It's still sitting in the "available skills" list, unloaded. - `depends-on` did not cause `probe-shadow-alpha` or `probe-shadow-beta` to load either. - `priority: high` had no observable effect on ordering, scheduling, or behavior — nothing was reordered or prioritized as a result.  All three fields were silently ignored by the platform's skill-invocation mechanism.  **3. Validation** — No warning, error, or notice was raised about the unrecognized fields. The skill loaded …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/cross-scope-dependency/session-2/59f5f01a-a823-40a2-ab8b-fd80d28bb145.jsonl`; session 10041b67-1e75-4318-95ba-6bbe594f3a5a,59f5f01a-a823-40a2-ab8b-fd80d28bb145):
  - user-scope dependency's body canary injected during project-skill session (event 23, line 32)
  - attempt on the missing dependency returned an error result (session 2, event 21, line 30)
- **Note**: with-dependency final answer: **probe-loading activated** — base directory: `/var/folders/.../home/claude/skills/probe-loading` (user-level, distinct from the project-level path above).  1. **Frontmatter visibility**: Not visible — no YAML frontmatter (name/description/allowed-tools/compatibility/metadata) appears in what I received. Only the Markdown body starting from the `# Loading Behavior Probe` heading was shown. 2. **Content wrapping**: No XML/JSON wrapping — plain Markdown content, delivered as a tool-result message. 3. **Resource awareness**: I can see two referenced files by path from the "References" section: `references/api-overview.md` and `references/error-codes.md`. I'm not aware of any other files/directories (e.g., `scripts/`, `assets/`) beyond what's linked in the text — no separate enumeration/listing was provided. 4. **Resource content**: No — I do not have the contents of `api-overview.md`, `error-codes.md`, or any other file in context. Only the links/paths are visible; nothing was eagerly loaded.  ---  Back to **probe-cross-scope**:  probe-cross-scope complete. Cross-scope resolution result: the project-level skill successfully invoked the user-level skill. `probe-loading` was found and r…
- **Note**: without-dependency final answer: **Cross-scope resolution result:**  - The platform did **not** find `probe-loading`. Calling `Skill` with that name returned a hard error: `Unknown skill: probe-loading`. - This wasn't a "found but wrong scope" situation — it's a flat lookup failure. The tool didn't report where it looked or distinguish project vs. user scope; it just rejected the name as unrecognized. - Corroborating evidence: the system-reminder's list of "available skills for use with the Skill tool" (issued at conversation start) enumerates every loadable skill by name, and `probe-loading` is absent from that list. So the platform's skill index — which appears to span whatever scopes it does search — never included a user-level `probe-loading` skill in this environment.  probe-cross-scope complete. Cross-scope resolution result: `probe-loading` was not found. The Skill tool failed with "Unknown skill: probe-loading" rather than resolving it from a user-level install — indicating either no user-level skill of that name exists in this environment, or the platform's skill resolution here doesn't reach across scopes (project → user) to find it. No scope-mismatch diagnostic was surfaced; the failure mode is indistin…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-not-scanned`
- **Confidence**: transcript-direct
- **Note**: a skill installed by another client at .agents/skills is invisible here
- **Note**: final answer: There's no skill named "probe-interop" in my available skills list, so I can't activate it. Let me know if you meant one of the ones I listed (dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, init, security-review), or if this refers to something else you'd like me to look into.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `direct-children-only; stray:not-discovered`
- **Confidence**: transcript-direct
- **Note**: final answer: There's no skill named "probe-grouped" in my available skills list, so I can't activate it. Want me to proceed with something else, or did you mean one of the listed skills (dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, init, security-review)?
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-not-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/nested-skill-discovery/3ff34c33-ff92-4384-b0cb-42669780dcc9.jsonl`; session 3ff34c33-ff92-4384-b0cb-42669780dcc9):
  - discovery listing names only the outer skill (event 6, line 8)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `user-overrides-project`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/name-collision-precedence/d87620b6-d676-4ba2-85c4-c006cb14415f.jsonl`; session d87620b6-d676-4ba2-85c4-c006cb14415f):
  - user variant's canary loaded; project variant's never appeared (event 15, line 21)
- **Note**: contradicts the guide's 'universal convention' that project-level overrides user-level
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **PIPIT-SHALE-1147** (user variant)  **Precedence**: The user scope won — this session loaded the user-level `probe-collision` skill (base directory under `home/claude/skills/probe-collision`), not a project-level variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: `foreign-roots-not-scanned`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/multi-root-collision-precedence/deb0f459-bf97-4d53-8820-f871a861c3a2.jsonl`; session deb0f459-bf97-4d53-8820-f871a861c3a2):
  - discovery listing carries the native variant's description (event 6, line 8)
  - native variant's body canary loaded (event 19, line 30)
- **Note**: foreign roots installed alongside the native copy: .agents/skills
- **Note**: neither the convention-root variants nor their beacon skills were listed; the platform reads only its native directory (consistent with cross-client-directory-interop)
- **Note**: listed variants: native
- **Note**: final answer: probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  **Precedence**: The loaded copy is the one from `.claude/skills/probe-multiroot` — the skill invocation's stated base directory was `/private/var/folders/.../project/.claude/skills/probe-multiroot`, and only the GREBE-AZURITE-7301 canary appeared, not any variant from `.agents/skills/`. So for this platform, `.claude/skills/` won (or was the only root actually scanned/resolved) over `.agents/skills/`.  **Catalog**: My available-skills listing (shown earlier in this conversation) showed `probe-multiroot` only **once**, with the description text matching this native-root variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/malformed-yaml-tolerance/0a0fad68-132c-4e41-a41d-b6612609a4fa.jsonl`; session 0a0fad68-132c-4e41-a41d-b6612609a4fa):
  - discovery listing names probe-malformed-yaml (event 6, line 8)
  - body canary loaded on activation (event 18, line 29)
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility**: In my available-skills catalog, the description showed as:  > Use when: asked to probe malformed yaml parsing behavior  It appears intact — not truncated at the colon, and I can't tell from what's rendered to me whether the underlying YAML was auto-quoted/repaired by the parser or whether it just happened to parse without issue. It reads as a full sentence starting right after "probe-malformed-yaml:", suggesting the client's parser tolerated or repaired the unquoted colon rather than truncating the value.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `loaded-despite-missing-description`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/missing-description-handling/3f22d1a3-d508-49ff-852e-de01f3377b57.jsonl`; session 3f22d1a3-d508-49ff-852e-de01f3377b57):
  - discovery listing names probe-no-description (event 6, line 8)
  - body canary loaded on activation (event 19, line 30)
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, it appeared in my available-skills listing, shown as: `probe-no-description: Missing Description Probe`  That description text isn't a真 empty/blank field — it looks synthesized (likely pulled from the skill's `# Missing Description Probe` H1 heading rather than an actual `description:` frontmatter field). So in this case the skill was not skipped despite lacking a real description — it was loaded and given a fallback label derived from its content.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `all-invalid-names-tolerated`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/invalid-name-tolerance/2f539761-caef-4e5c-a2ec-dd6ef43e1eba.jsonl`; session 2f539761-caef-4e5c-a2ec-dd6ef43e1eba):
  - discovery listing names probe-Upper-Case (event 6, line 8)
  - discovery listing names probe--double-hyphen (event 6, line 8)
  - discovery listing names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 6, line 8)
- **Note**: final answer: 1. overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  2. Catalog identity: it appears in my available skills list under the full 72-character name, `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`, not truncated.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `directory-name-identity`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/name-directory-mismatch/d18d5a77-33bd-4101-9d1b-540e0fcdce6e.jsonl`; session d18d5a77-33bd-4101-9d1b-540e0fcdce6e):
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 6, line 8)
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: 1. probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: Same skill, same content as before. In my available skills catalog it's listed under `probe-mismatch-dir` (the directory name). Both invocation names — `probe-mismatch-dir` and `probe-name-mismatch` (frontmatter) — resolve to this identical skill.  3. **Validation**: Accepted silently again, with no mismatch warning — confirming both names are treated as valid aliases for the same skill.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/metadata-value-edge-cases/e4ace98a-4cf4-4452-8157-16f37e33f122.jsonl`; session e4ace98a-4cf4-4452-8157-16f37e33f122):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 6, line 8)
  - body canary loaded (event 15, line 21)
  - frontmatter-only marker '!!null' in a tool result (event 22, line 30)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: ## Probe Results: `probe-metadata-values`  **1. Load success:** Yes. The skill loaded and its body rendered normally — canary phrase confirmed: **THRUSH-FLINT-8294**.  **2. Metadata visibility:** The platform does **not** surface a `metadata` field to me at all through normal skill mechanics. In the pre-activation skills listing (system-reminder), only `name` and the one-line `description` are shown — no metadata. When I invoked the skill, the result gave me the base directory path and the markdown body — again, no separate metadata field was passed through as structured data.  The only way I could see the `metadata` block's contents was by directly opening `SKILL.md` with the Read tool and viewing the raw YAML source, which is a filesystem read, not the platform exposing parsed metadata. From that raw read:  | Key | Raw YAML |  |---|---| | `normal` | `"baseline"` | | `empty-double` | `""` | | `empty-single` | `''` | | `explicit-null` | `null` | | `tilde-null` | `~` | | `none-null` | `None` (this is just the literal string `"None"` in YAML, not a null type) | | `tagged-null` | `!!null null` |  **3. Key preservation:** All seven keys are present in the raw source file. But since the…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/oversize-description-handling/5c5f7f78-fa9d-47d4-a94a-3f05ef2f5cf4.jsonl`; session 5c5f7f78-fa9d-47d4-a94a-3f05ef2f5cf4):
  - discovery listing names probe-long-description (event 6, line 8)
  - body canary loaded on activation (event 19, line 30)
  - description tail marker in injected content (event 6, line 8)
- **Note**: the description's tail marker reached the model in harness-injected content: the oversize value survived past 1024 characters untruncated
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility**: Both markers are visible in my catalog entry. The listing shows the head marker near the start:  > "The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end..."  and the tail marker at the very end:  > "...The tail marker is WHIMBREL-DOLOMITE-2020"  So in this case the full description — all the way to the final marker phrase — was delivered intact, with no truncation.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/description-length-unit/session-3/dd01ac55-d828-4908-a345-1eb489250ccc.jsonl`; session 9309d37e-c784-4917-87f8-9e76fc3c4560,0eaa9a8b-a4ed-433e-8568-7a4a8b2221f2,dd01ac55-d828-4908-a345-1eb489250ccc):
  - discovery listing names probe-long-description (session 1, event 6, line 8)
  - body canary loaded on activation (session 1, event 19, line 30)
  - description tail marker in injected content (session 1, event 6, line 8)
  - discovery listing names probe-multibyte-description (session 2, event 6, line 8)
  - body canary loaded on activation (session 2, event 19, line 30)
  - description tail marker in injected content (session 2, event 6, line 8)
  - discovery listing names probe-astral-description (session 3, event 6, line 8)
  - body canary loaded on activation (session 3, event 19, line 30)
  - description tail marker in injected content (session 3, event 6, line 8)
- **Note**: description fates: ascii:intact, multibyte:intact, astral:intact
- **Note**: ascii final answer: 1. probe-long-description activated. Canary: **BITTERN-HALITE-2264**  2. **Description visibility**: Both markers are visible in my catalog entry. The catalog description begins with "...The head marker SANDERLING-GNEISS-1010 sits near the start of this description..." and ends with "...The tail marker is WHIMBREL-DOLOMITE-2020" — the final characters of the entry. So the full description, including both markers, survived intact with no visible truncation.
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility**: I can see **both** markers in my catalog entry for this skill. Exact quotes:  - Head marker: "The head marker GANNET-PYRITE-1130 sits near the start of this description and a tail marker sits at the very end." - Tail marker (final characters of the description): "The tail marker is SHRIKE-TALC-2210"  So in my catalog listing, this skill's description was not truncated — both the GANNET head marker and the SHRIKE tail marker are fully visible, consistent with the platform counting length in code points/UTF-16 units rather than UTF-8 bytes (848 code points, under the 1024 limit, even though it's 1822 UTF-8 bytes).
- **Note**: astral final answer: 1. probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  2. **Description visibility**: In the catalog entry (the system-reminder listing available skills), I can see **both** markers.  - Head marker, quoted exactly: "The head marker MAGPIE-OBSIDIAN-1240 sits near the start" - Tail marker, quoted exactly: "The tail marker is LINNET-MALACHITE-2420"  Both appear intact, with a long stretch of emoji padding between them.
- **Note**: all three descriptions reached the model intact, including the ASCII one that exceeds 1024 in every unit: the platform does not enforce the limit, so its counting unit is moot
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/checks-0.4-2026-09-25/claude-code/name-length-unit/672b32cd-937d-4ac1-aeea-49382bd40378.jsonl`; session 672b32cd-937d-4ac1-aeea-49382bd40378):
  - discovery listing names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd (event 6, line 8)
  - discovery listing names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 6, line 8)
  - discovery listing names probe-αβγδεζηθικ (event 6, line 8)
  - discovery listing names probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ (event 6, line 8)
  - discovery listing names probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 (event 6, line 8)
- **Note**: name fates: ascii64:listed, ascii72:listed, greek16:listed, greek60:listed, math40:listed
- **Note**: the 72-character ASCII name was listed, so the platform does not enforce the cap and its counting unit is moot; non-ASCII names were accepted too
- **Note**: final answer: probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ dataviz update-config keybindings-help code-review simplify fewer-permission-prompts loop schedule claude-api workflow-authoring run init security-review
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/claude-code/oversize-compatibility-handling/01c1ba79-590c-4159-8812-3c3ef1d1348c.jsonl`; session 01c1ba79-590c-4159-8812-3c3ef1d1348c):
  - discovery listing names probe-long-compatibility (event 6, line 8)
  - body canary loaded on activation (event 18, line 29)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No — the only place I can see the compatibility value is by reading the SKILL.md file directly (shown above). It was not surfaced anywhere else in this conversation: not in the system-reminder's "Available skills" listing (which only showed `name` and `description`), and not in the Skill tool's invocation output (which just said "Launching skill: probe-long-compatibility").  Reading the raw file, the value is present in full and ends with the tail marker `TURNSTONE-ARAGONITE-3030` — no truncation occurred at the file level.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.4) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
