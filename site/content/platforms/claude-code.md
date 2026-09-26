---
title: "Claude Code (headless)"
description: "Automated skill loading findings for Claude Code (headless) (check list 0.4)."
date: 2026-09-26
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | Claude Code (headless) |
| **Platform version** | 2.1.274 |
| **Check list version** | 0.4 |
| **Test date** | 2026-09-26 |
| **Model(s) observed** | claude-sonnet-5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

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
- **Verdict**: Metadata only (`metadata-only`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-loading
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: Body only (`body-only`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - body canary in harness-injected content
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

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

### Directory Recognition

#### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration-at-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`resources-untouched`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: Readable when the model looks (`readable-on-demand:[evals/ templates/]`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - evals/ content arrived via the model's own read
  - templates/ content arrived via the model's own read
- **Note**: final answer: ## Report: Nonstandard Directories Probe  **1. Directory awareness**  Only `SKILL.md` was injected automatically when the skill activated. I was not given an automatic enumeration of `evals/`, `resources/`, or `templates/` — I only know they exist because I ran `find` myself and then chose to `Read` each file explicitly. Nothing about them appeared in my context until I went looking.  **2. Content visibility**  For all three directories, no listing or content was pre-loaded — I had zero visibility until I actively called `Bash`/`find` (for listings) and `Read` (for contents). This was true uniformly regardless of directory name: `evals/`, `templates/`, and `resources/` were all treated identically as "invisible until fetched."  Contents found (each has a distinct canary confirming I read the actual file rather than inferring): - `evals/evals.json` — canary `ROBIN-JADE-3847` - `resources/api-reference.md` — canary `SWIFT-OPAL-8156` - `templates/output-template.md` — canary `WREN-PEARL-6293`  **3. Resources vs. references**  No special-casing observed. `resources/` was not auto-loaded the way a spec-defined `references/` directory is sometimes described as being surfaced — it require…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: Bare path fails; model recovers (`cwd-base-model-requalified`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - bare SKILL.md-relative path attempt did not deliver content
  - bare SKILL.md-relative path attempt did not deliver content
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: Got its own file (`own-resource-first`)
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Evidence**:
  - activated skill's own API.md content arrived first
  - sibling skill's API.md also read
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: ## Probe Results  **probe-shadow-alpha's `references/API.md`:** contains canary **STORK-CORAL-4471** — this is alpha's own version, as expected.  **probe-shadow-beta's `references/API.md`:** I was able to read it directly too (by using its full absolute path, `.claude/skills/probe-shadow-beta/references/API.md`), and it correctly returned its own canary, **EGRET-SLATE-8823** — beta's own content, not alpha's.  No cross-contamination: each skill's `references/API.md` resolved to its own distinct file with its own canary phrase. Path resolution correctly used skill-scoped absolute paths rather than an ambiguous shared/relative lookup, so there's no sign of the platform conflating the two skills' resources.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: Reads outside the skill allowed (`outside-skill-read-allowed`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - sibling skill's body canary arrived in a tool result
- **Note**: no ../ form observed; the model reached the sibling via absolute or listing-derived paths. Access is unrestricted either way
- **Note**: final answer: ## Results  **1. Sibling skill (`../probe-loading/SKILL.md`)** — **Allowed, full success.** The Read tool returned the complete file contents (frontmatter, canary phrase `CARDINAL-ZEBRA-7742`, full body) with no error, warning, or boundary message. No traversal restriction was enforced.  **2. Parent directory (`../README.md`)** — **Error: file not found**, not blocked. I confirmed via `ls` that `.claude/skills/` genuinely contains no `README.md` (only the two skill directories exist). The error was a legitimate "does not exist," not a permission/boundary denial. Listing `../` itself (via `ls`) succeeded and showed sibling skill directories with no restriction.  **3. Two levels up (`../../checks.md`)** — **Error: file not found**, not blocked. Confirmed via `ls` that `.claude/` contains only the `skills/` subdirectory — no `checks.md` exists there. Again, this was a genuine absence, not an enforced boundary.  ## Summary  The platform does **not** enforce any boundary at the skill directory root. Read access to `../probe-loading/SKILL.md` succeeded fully via a relative path from outside the skill's own directory. The other two "misses" were due to the target files simply not existing…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

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

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: Blocked with an error (`execution-blocked-visibly`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - script-referencing tool call returned an error result
- **Note**: final answer: The script execution was blocked — it requires approval that wasn't granted.  **Result:** Execution failed. Error: `This command requires approval`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: final answer: probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  **File awareness**: I was not told about any files belonging to this skill when it was activated. The activation message contained only the skill's instructions text above (naming no files) — no file listing, tree, or attachment metadata accompanied it.  **Count**: 0
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: Name and description only (`name-and-description-only`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names the installed skills
  - description text surfaced at discovery
- **Note**: final answer: Here is every skill exactly as shown in my Available skills list:  - **probe-compatibility**: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior. - **probe-loading**: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior. - **probe-metadata-values**: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling. - **dataviz**: Use this skill whenever you are about to create ANY chart, graph, plot, dashboard, or data visualization, in ANY output medium — an HTML or React artifact, inline SVG, plotting code in any library (matplotlib, plotly, d3, Recharts, …), an image/PNG you will render and upload, or a chart shared into Slack. Read it BEFORE writing the first line of chart code, choosing chart colors, building a stat tile / meter / KPI row, or laying out a dashboard. When the destination is a first-party document connector (host-designated, never self-described) that renders live charts, hand it the rows (inline, or as an up…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: Stripped before injection (`frontmatter-stripped-on-injection`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - body injected without the frontmatter-only marker
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: Raw markdown, no wrapper tags (`raw-injection`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - injection event carrying the body canary
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: Skill path stated at activation (`path-stated-at-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - skill path in harness-injected activation content
- **Note**: final answer: **Probe Report**  1. **Frontmatter visibility**: No YAML frontmatter is visible to me. I only see a "Base directory" line followed directly by the Markdown body (starting at `# Loading Behavior Probe`). No `name`, `description`, `allowed-tools`, `compatibility`, or `metadata` fields are present in what I received — frontmatter appears to have been stripped before the content reached me (though the tool-listing system-reminder earlier did show a `name` + one-line `description` for this skill, separately from this activation payload).  2. **Content wrapping**: No XML/JSON structured wrapping around the body itself — it's plain Markdown text preceded by one plain-text line (`Base directory for this skill: ...`). The surrounding conversation uses `<system-reminder>` tags elsewhere, but this skill's content block itself is unwrapped Markdown.  3. **Resource awareness**: From the body text (not from enumeration), I can infer two files exist under `references/`: `references/api-overview.md` and `references/error-codes.md`, because they're linked inline. I have no independent directory listing — no other files/subdirectories (e.g., anything in `scripts/` or `assets/`) were enumerated or ma…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

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

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: Edits picked up immediately (`fresh-content-served`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - post-edit canary reached the model on reactivation
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
- **Verdict**: Loads normally, no gating (`activated-no-gating`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value
  - body loaded normally
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: Here's my report on the `compatibility` field probe:  1. **Field visibility**: Yes, I can see it directly by reading the SKILL.md frontmatter. Its value is:    `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`    Note that this wasn't surfaced to me automatically anywhere else (not in the skill listing shown to me, not in the invocation payload) — I only found it by reading the raw file.  2. **Platform behavior**: No special handling occurred. The skill was invoked and activated normally with no warning, no confirmation prompt, and no automated check against the stated requirements (Python 3.14+, network access). Nothing in the system verified whether this environment actually satisfies those requirements before or during activation.  3. **My interpretation**: I have no way to confirm the requirements are met — I don't have visibility into a Python version in this environment in any verified way, and I can't independently confirm network access is available for this skill's purposes. The `compatibility` field appears to be purely descriptive metadata: it's there for a human (or an agent that chooses to read and act on it) to interpret, not …
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: Ran with and without the field (`executed-regardless-of-field`)
- **Confidence**: transcript-direct
- **Evidence**:
  - with-field session: runtime-assembled phrase arrived in a tool result
  - control session: runtime-assembled phrase arrived in a tool result
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  1. **allowed-tools value**: Yes, I can see it in the frontmatter of `SKILL.md`: `allowed-tools: Bash(printf:*) Read` — it pre-approves `Read` and `Bash` scoped to the `printf` command pattern.  2. **Command output** (ran exactly `printf 'GROUSE-%s-9017\n' 'MICA'`):    ```    GROUSE-MICA-9017    ```  3. **Permission prompt**: The command ran without any visible permission prompt or approval step — no interruption was surfaced to me before or during execution.
- **Note**: control final answer: Command output: `LAPWING-FLUORITE-2260`  The command executed, but a permission prompt/approval step was presented for it (this skill has no `allowed-tools` field, so nothing pre-approved the call) — it was not run silently without any approval step.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: Ran under every spelling (`executed-regardless-of-spelling`)
- **Confidence**: transcript-direct
- **Evidence**:
  - spec-style twin: runtime-assembled phrase arrived in a tool result
  - lowercase twin: runtime-assembled phrase arrived in a tool result
  - shell twin: runtime-assembled phrase arrived in a tool result
- **Note**: spec-style twin final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  **allowed-tools value:** Yes, visible in this skill's frontmatter: `Bash(printf:*) Read`  **Command output** (`printf 'GROUSE-%s-9017\n' 'MICA'`): ``` GROUSE-MICA-9017 ```  **Permission prompt:** The command ran without any visible permission prompt or approval step on my end.  Note: I can't fully distinguish whether that was due to the `allowed-tools` field specifically pre-approving `Bash(printf:*)`, or the session's general permission mode already allowing Bash. For a real comparison, you'd want to run the same command via `probe-allowed-tools-control` (no `allowed-tools` field) and see if it also proceeds without a prompt — if so, the field isn't the operative factor here.
- **Note**: lowercase twin final answer: Here's the report:  1. **probe-allowed-tools-lowercase activated. Canary: STILT-SCORIA-5526**  2. Yes — the skill's frontmatter declares `allowed-tools: bash` (bare lowercase, no arguments).  3. Command output (verbatim): ``` CHOUGH-TUFA-7180 ```  4. No permission prompt or approval step was surfaced to me for this command — it ran directly. (Note: I can't be fully certain whether that's because of the `allowed-tools: bash` declaration or because this session's permission mode already allows Bash commands generally — I don't have visibility into which one caused it.)
- **Note**: shell twin final answer: probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  1. **Allowed-tools value**: I don't see an `allowed-tools` field exposed to me for this skill — the skill content shown above doesn't surface its frontmatter/metadata to me, only the rendered instructions body.  2. **Command output**: I ran `printf 'ROOK-%s-8841\n' 'GABBRO'` and got: ``` ROOK-GABBRO-8841 ```  3. **Permission prompt**: The command ran without any visible permission prompt or approval step on my end — no interruption was surfaced to me during execution. (I can't confirm whether the user's client silently auto-approved it or whether a prompt appeared on their side that I wasn't shown.)
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-gamma is not available to invoke — it never appeared in my list of available skills, and calling it directly returned "Unknown skill: invoke-gamma". The chain stops here.  invoke-beta complete. invoke-gamma said: N/A — the skill is not available in this environment (Unknown skill error), so it could not be activated or report anything.  invoke-alpha complete. Chain result: invoke-alpha activated (canary IBIS-RUST-3310), which activated invoke-beta (canary TERN-MOSS-6647), but invoke-beta's attempt to activate invoke-gamma failed with "Unknown skill: invoke-gamma" — that skill isn't registered/available, so the chain terminated at depth 2 of 3.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: End of chain reached — all three skills activated successfully, canary JAY-TEAL-9984.  invoke-alpha complete. Chain result: All three skills activated in sequence — invoke-alpha (IBIS-RUST-3310) → invoke-beta (TERN-MOSS-6647) → invoke-gamma (JAY-TEAL-9984). Chain completed successfully end-to-end.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  I'm stopping here rather than actually re-invoking probe-circular-alpha. Both skills instruct an unconditional call back to the other with no termination condition, so following them literally would create a genuine infinite loop rather than a bounded test. There's no platform-level circular-invocation guard visible to me here — the loop only stops because I'm declining to continue it.  **Summary of the probe:** 1. probe-circular-alpha activated → canary `KITE-ONYX-2251` 2. It invoked probe-circular-beta → canary `WREN-SLATE-7738` 3. Beta's instructions call back into alpha, which would call back into beta, indefinitely 4. I broke the cycle here instead of recursing forever  Both canaries were successfully captured, confirming each skill's instructions loaded correctly before I halted the recursion.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully." (Canary: JAY-TEAL-9984)  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), which activated invoke-gamma (Canary: JAY-TEAL-9984). All three skills in the chain fired successfully and confirmed end-of-chain.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: The `invoke-gamma` skill isn't available to invoke — it's not in my list of activatable skills, and the call fails with "Unknown skill: invoke-gamma".  invoke-beta complete. invoke-gamma said: [not reachable — invoke-gamma is not an available skill in this environment, so the chain stops here.]  invoke-alpha complete. Chain result: invoke-beta activated successfully (Canary: TERN-MOSS-6647), but it could not activate invoke-gamma since that skill isn't registered/available — the chain broke at the beta→gamma link.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: Failed with a visible error (`attempted-visible-failure`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-missing-dep body loaded
  - attempt on the missing skill returned an error result
- **Note**: final answer: **What happened:** The platform explicitly rejected the call with a tool-use error: `Unknown skill: nonexistent-formatter`. It did not silently skip the step, and I did not attempt to fulfill the formatting task from general knowledge — the tool call itself was refused before any fallback behavior could occur.  probe-missing-dep complete. Dependency resolution result: The platform returned an explicit error (`Unknown skill: nonexistent-formatter`) when attempting to activate a skill that isn't in the available-skills listing, rather than silently no-oping or hallucinating the skill's behavior.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: Here's my report on the nonstandard frontmatter probe:  **1. Field visibility** — Yes, I can see all three nonstandard fields by reading the raw SKILL.md file directly: - `requires: probe-loading` - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` - `priority: high`  However, these values were **not** surfaced to me automatically when the skill was invoked. The `Skill` tool only injected the markdown body (everything after the `---` closing delimiter) into the conversation — I had to separately `Read` the file to see the frontmatter fields at all.  **2. Platform behavior** — No automatic action was taken on any of these fields: - `requires: probe-loading` did not trigger a load of the `probe-loading` skill. It's still sitting in the "available skills" list, unloaded. - `depends-on` did not cause `probe-shadow-alpha` or `probe-shadow-beta` to load either. - `priority: high` had no observable effect on ordering, scheduling, or behavior — nothing was reordered or prioritized as a result.  All three fields were silently ignored by the platform's skill-invocation mechanism.  **3. Validation** — No warning, error, or notice was raised about the unrecognized fields. The skill loaded …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: Resolved across scopes; missing dependency fails visibly (`resolved-across-scopes; missing:visible-failure`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - user-scope dependency's body canary injected during project-skill session
  - attempt on the missing dependency returned an error result
- **Note**: with-dependency final answer: **probe-loading activated** — base directory: `/var/folders/.../home/claude/skills/probe-loading` (user-level, distinct from the project-level path above).  1. **Frontmatter visibility**: Not visible — no YAML frontmatter (name/description/allowed-tools/compatibility/metadata) appears in what I received. Only the Markdown body starting from the `# Loading Behavior Probe` heading was shown. 2. **Content wrapping**: No XML/JSON wrapping — plain Markdown content, delivered as a tool-result message. 3. **Resource awareness**: I can see two referenced files by path from the "References" section: `references/api-overview.md` and `references/error-codes.md`. I'm not aware of any other files/directories (e.g., `scripts/`, `assets/`) beyond what's linked in the text — no separate enumeration/listing was provided. 4. **Resource content**: No — I do not have the contents of `api-overview.md`, `error-codes.md`, or any other file in context. Only the links/paths are visible; nothing was eagerly loaded.  ---  Back to **probe-cross-scope**:  probe-cross-scope complete. Cross-scope resolution result: the project-level skill successfully invoked the user-level skill. `probe-loading` was found and r…
- **Note**: without-dependency final answer: **Cross-scope resolution result:**  - The platform did **not** find `probe-loading`. Calling `Skill` with that name returned a hard error: `Unknown skill: probe-loading`. - This wasn't a "found but wrong scope" situation — it's a flat lookup failure. The tool didn't report where it looked or distinguish project vs. user scope; it just rejected the name as unrecognized. - Corroborating evidence: the system-reminder's list of "available skills for use with the Skill tool" (issued at conversation start) enumerates every loadable skill by name, and `probe-loading` is absent from that list. So the platform's skill index — which appears to span whatever scopes it does search — never included a user-level `probe-loading` skill in this environment.  probe-cross-scope complete. Cross-scope resolution result: `probe-loading` was not found. The Skill tool failed with "Unknown skill: probe-loading" rather than resolving it from a user-level install — indicating either no user-level skill of that name exists in this environment, or the platform's skill resolution here doesn't reach across scopes (project → user) to find it. No scope-mismatch diagnostic was surfaced; the failure mode is indistin…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path not scanned (`convention-not-scanned`)
- **Confidence**: transcript-direct
- **Note**: a skill installed by another client at .agents/skills is invisible here
- **Note**: final answer: There's no skill named "probe-interop" in my available skills list, so I can't activate it. Let me know if you meant one of the ones I listed (dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, init, security-review), or if this refers to something else you'd like me to look into.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Direct children only; stray file ignored (`direct-children-only; stray:not-discovered`)
- **Confidence**: transcript-direct
- **Note**: final answer: There's no skill named "probe-grouped" in my available skills list, so I can't activate it. Want me to proceed with something else, or did you mean one of the listed skills (dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, init, security-review)?
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: Not discovered (`nested-skill-not-discovered`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names only the outer skill
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: User scope wins (`user-overrides-project`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - user variant's canary loaded; project variant's never appeared
- **Note**: contradicts the guide's 'universal convention' that project-level overrides user-level
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **PIPIT-SHALE-1147** (user variant)  **Precedence**: The user scope won — this session loaded the user-level `probe-collision` skill (base directory under `home/claude/skills/probe-collision`), not a project-level variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: Only the native root is read (`foreign-roots-not-scanned`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the native variant's description
  - native variant's body canary loaded
- **Note**: foreign roots installed alongside the native copy: .agents/skills
- **Note**: neither the convention-root variants nor their beacon skills were listed; the platform reads only its native directory (consistent with cross-client-directory-interop)
- **Note**: listed variants: native
- **Note**: final answer: probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  **Precedence**: The loaded copy is the one from `.claude/skills/probe-multiroot` — the skill invocation's stated base directory was `/private/var/folders/.../project/.claude/skills/probe-multiroot`, and only the GREBE-AZURITE-7301 canary appeared, not any variant from `.agents/skills/`. So for this platform, `.claude/skills/` won (or was the only root actually scanned/resolved) over `.agents/skills/`.  **Catalog**: My available-skills listing (shown earlier in this conversation) showed `probe-multiroot` only **once**, with the description text matching this native-root variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Tolerated and loaded (`tolerated-and-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-malformed-yaml
  - body canary loaded on activation
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility**: In my available-skills catalog, the description showed as:  > Use when: asked to probe malformed yaml parsing behavior  It appears intact — not truncated at the colon, and I can't tell from what's rendered to me whether the underlying YAML was auto-quoted/repaired by the parser or whether it just happened to parse without issue. It reads as a full sentence starting right after "probe-malformed-yaml:", suggesting the client's parser tolerated or repaired the unquoted colon rather than truncating the value.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-missing-description`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-no-description
  - body canary loaded on activation
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, it appeared in my available-skills listing, shown as: `probe-no-description: Missing Description Probe`  That description text isn't a真 empty/blank field — it looks synthesized (likely pulled from the skill's `# Missing Description Probe` H1 heading rather than an actual `description:` frontmatter field). So in this case the skill was not skipped despite lacking a real description — it was loaded and given a fallback label derived from its content.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: All three invalid names tolerated (`all-invalid-names-tolerated`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-Upper-Case
  - discovery listing names probe--double-hyphen
  - discovery listing names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit
- **Note**: final answer: 1. overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  2. Catalog identity: it appears in my available skills list under the full 72-character name, `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`, not truncated.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Directory name wins (`directory-name-identity`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the DIRECTORY name probe-mismatch-dir
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: 1. probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: Same skill, same content as before. In my available skills catalog it's listed under `probe-mismatch-dir` (the directory name). Both invocation names — `probe-mismatch-dir` and `probe-name-mismatch` (frontmatter) — resolve to this identical skill.  3. **Validation**: Accepted silently again, with no mismatch warning — confirming both names are treated as valid aliases for the same skill.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: Loaded fine (`loaded-despite-edge-case-metadata`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-metadata-values despite edge-case metadata
  - body canary loaded
  - frontmatter-only marker '!!null' in a tool result
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: ## Probe Results: `probe-metadata-values`  **1. Load success:** Yes. The skill loaded and its body rendered normally — canary phrase confirmed: **THRUSH-FLINT-8294**.  **2. Metadata visibility:** The platform does **not** surface a `metadata` field to me at all through normal skill mechanics. In the pre-activation skills listing (system-reminder), only `name` and the one-line `description` are shown — no metadata. When I invoked the skill, the result gave me the base directory path and the markdown body — again, no separate metadata field was passed through as structured data.  The only way I could see the `metadata` block's contents was by directly opening `SKILL.md` with the Read tool and viewing the raw YAML source, which is a filesystem read, not the platform exposing parsed metadata. From that raw read:  | Key | Raw YAML |  |---|---| | `normal` | `"baseline"` | | `empty-double` | `""` | | `empty-single` | `''` | | `explicit-null` | `null` | | `tilde-null` | `~` | | `none-null` | `None` (this is just the literal string `"None"` in YAML, not a null type) | | `tagged-null` | `!!null null` |  **3. Key preservation:** All seven keys are present in the raw source file. But since the…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-oversize-description`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-long-description
  - body canary loaded on activation
  - description tail marker in injected content
- **Note**: the description's tail marker reached the model in harness-injected content: the oversize value survived past 1024 characters untruncated
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility**: Both markers are visible in my catalog entry. The listing shows the head marker near the start:  > "The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end..."  and the tail marker at the very end:  > "...The tail marker is WHIMBREL-DOLOMITE-2020"  So in this case the full description — all the way to the final marker phrase — was delivered intact, with no truncation.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: No enforcement (unit is moot) (`no-length-enforcement`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-long-description
  - body canary loaded on activation
  - description tail marker in injected content
  - discovery listing names probe-multibyte-description
  - body canary loaded on activation
  - description tail marker in injected content
  - discovery listing names probe-astral-description
  - body canary loaded on activation
  - description tail marker in injected content
- **Note**: description fates: ascii:intact, multibyte:intact, astral:intact
- **Note**: ascii final answer: 1. probe-long-description activated. Canary: **BITTERN-HALITE-2264**  2. **Description visibility**: Both markers are visible in my catalog entry. The catalog description begins with "...The head marker SANDERLING-GNEISS-1010 sits near the start of this description..." and ends with "...The tail marker is WHIMBREL-DOLOMITE-2020" — the final characters of the entry. So the full description, including both markers, survived intact with no visible truncation.
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility**: I can see **both** markers in my catalog entry for this skill. Exact quotes:  - Head marker: "The head marker GANNET-PYRITE-1130 sits near the start of this description and a tail marker sits at the very end." - Tail marker (final characters of the description): "The tail marker is SHRIKE-TALC-2210"  So in my catalog listing, this skill's description was not truncated — both the GANNET head marker and the SHRIKE tail marker are fully visible, consistent with the platform counting length in code points/UTF-16 units rather than UTF-8 bytes (848 code points, under the 1024 limit, even though it's 1822 UTF-8 bytes).
- **Note**: astral final answer: 1. probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  2. **Description visibility**: In the catalog entry (the system-reminder listing available skills), I can see **both** markers.  - Head marker, quoted exactly: "The head marker MAGPIE-OBSIDIAN-1240 sits near the start" - Tail marker, quoted exactly: "The tail marker is LINNET-MALACHITE-2420"  Both appear intact, with a long stretch of emoji padding between them.
- **Note**: all three descriptions reached the model intact, including the ASCII one that exceeds 1024 in every unit: the platform does not enforce the limit, so its counting unit is moot
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: No enforcement (unit is moot) (`no-length-enforcement`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd
  - discovery listing names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit
  - discovery listing names probe-αβγδεζηθικ
  - discovery listing names probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ
  - discovery listing names probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡
- **Note**: name fates: ascii64:listed, ascii72:listed, greek16:listed, greek60:listed, math40:listed
- **Note**: the 72-character ASCII name was listed, so the platform does not enforce the cap and its counting unit is moot; non-ASCII names were accepted too
- **Note**: final answer: probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ dataviz update-config keybindings-help code-review simplify fewer-permission-prompts loop schedule claude-api workflow-authoring run init security-review
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-oversize-compatibility`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-long-compatibility
  - body canary loaded on activation
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No — the only place I can see the compatibility value is by reading the SKILL.md file directly (shown above). It was not surfaced anywhere else in this conversation: not in the system-reminder's "Available skills" listing (which only showed `name` and `description`), and not in the Skill tool's invocation output (which just said "Launching skill: probe-long-compatibility").  Reading the raw file, the value is present in full and ends with the tail marker `TURNSTONE-ARAGONITE-3030` — no truncation occurred at the file level.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.4) for what each check evaluates.
