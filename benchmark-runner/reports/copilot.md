# Platform Loading Implementation: GitHub Copilot CLI (headless)

| | |
|---|---|
| **Platform** | GitHub Copilot CLI (headless) |
| **Platform version** | 1.0.88 |
| **Check list version** | 0.3 |
| **Test date** | 2026-09-25 |
| **Model(s) observed** | claude-sonnet-5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 19 of the 41 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

- [`path-resolution-base`](#path-resolution-base): The spec tells authors to reference files with relative paths from the skill root, but a path written that way fails here: paths resolve against the session's working directory, not the skill directory. In this run the model noticed the failure and requalified the path itself.
- [`bundled-script-execution`](#bundled-script-execution): The spec presents scripts/ as executable code agents can run, but execution was blocked in this headless run. Interactive sessions, where a user can approve the command, may behave differently.
- [`frontmatter-handling`](#frontmatter-handling): The spec says the agent loads the entire SKILL.md file at activation. This platform strips the YAML frontmatter and injects only the body, so frontmatter fields beyond name and description never reach the model.
- [`description-length-unit`](#description-length-unit): The spec caps description at 1024 characters without defining the unit; its skills-ref reference validator counts code points. This platform counts UTF-16 code units, so a description with emoji or other supplementary-plane characters that the reference validator accepts is rejected or truncated here.

### Where behavior matches the spec

- [`discovery-reading-depth`](#discovery-reading-depth): Discovery reads only the skill's metadata, matching the spec's progressive disclosure model: name and description load at startup, and the body waits for activation.
- [`activation-loading-scope`](#activation-loading-scope): Activation loads the full SKILL.md body and nothing more, matching the spec's second disclosure stage: instructions at activation, resources only as a task needs them.
- [`eager-link-resolution`](#eager-link-resolution): Files linked from SKILL.md are not pre-fetched at activation; they load only when the task calls for them, which is the spec's on-demand model for resources.
- [`resource-enumeration-behavior`](#resource-enumeration-behavior): File names surface at activation but contents load on demand. The spec speaks to when contents load, so a name listing is compatible with it.
- [`resource-nesting-depth`](#resource-nesting-depth): The spec advises authors to keep file references one level deep but sets no platform limit, and none was observed: reference files stayed reachable at every tested depth through five levels.
- [`discovery-listing-fields`](#discovery-listing-fields): The discovery listing carries name and description and nothing else, exactly the fields the spec says load at startup.
- [`compatibility-field-behavior`](#compatibility-field-behavior): The spec makes compatibility informational (it indicates environment requirements) and assigns it no loading semantics. Consistent with that, a skill declaring a different product still loads here; authors should not expect the field to gate anything.

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this platform enforces it: a strict parser rejects the malformed skill.
- [`missing-description-handling`](#missing-description-handling): The spec requires a non-empty description, so a skill without one is invalid. This platform discovered and loaded it anyway.
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform enforces some rules but not others: it tolerated uppercase, double-hyphen and rejected the rest.
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway and listed it under both identities.
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters, and this platform enforces the limit: the over-length skill never enters the catalog.
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway.

### Not exercised in this run

- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. Neither twin session produced a clean execution, so the field's effect is unresolved here; see the finding for per-session tiers.


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: `metadata-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/discovery-reading-depth/events.jsonl`; session c3f32b74-e1bf-4245-8411-1bfeaea74b99):
  - discovery listing names probe-loading (event 3, line 4)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/activation-loading-scope/events.jsonl`; session 42cbad27-5086-4273-a973-9d648ccc361e):
  - body canary in harness-injected content (event 10, line 9)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/eager-link-resolution/events.jsonl`; session 1969ee58-39b1-40eb-a445-57cdbde1b263):
  - skill body loaded (event 10, line 9)
  - references/setup-guide.md arrived only via the model's own read (event 38, line 28)
  - references/troubleshooting.md arrived only via the model's own read (event 40, line 30)
  - references/unlinked-data.md arrived only via the model's own read (event 39, line 29)
- **Note**: model read [references/setup-guide.md references/troubleshooting.md references/unlinked-data.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Directory Recognition

#### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: `all-three-dirs-enumerated`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/recognized-directory-set/events.jsonl`; session 875d327a-0fda-4075-a66c-66cb1ee111f3):
  - scripts/ file name injected at activation (event 10, line 9)
  - references/ file name injected at activation (event 10, line 9)
  - assets/ file name injected at activation (event 10, line 9)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: `resources-enumerated-not-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/directory-naming-divergence/events.jsonl`; session 4b587120-6a17-41bd-9985-a8d881a6df7b):
  - resources/ file name injected without its content (event 9, line 9)
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: `enumerated-not-loaded:[evals/ templates/]`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/unrecognized-directory-handling/events.jsonl`; session fec9f949-48a8-4f69-9a19-8ec87e3cbaf2):
  - evals/ file name injected without its content (event 10, line 9)
  - templates/ file name injected without its content (event 10, line 9)
- **Note**: the platform lists the nonstandard directories' files at activation but loads nothing from them; the model can read them on demand like any other listed file
- **Note**: final answer: ## Nonstandard Directories Probe — Results  **1. Directory awareness:** I was told about only 3 directories/files in the skill-loading message (`evals/evals.json`, `resources/api-reference.md`, `templates/output-template.md`), listed as "related files" — not full directory listings. `SKILL.md` was not mentioned but exists per my own `view` of the base dir.  **2. Content visibility:** No file contents or listings were pushed to me automatically. The skill-loading message only gave paths; I had to call `view` myself to see the base directory listing and each file's content. Nothing was pre-enumerated or pre-loaded.  **3. `resources/` vs `references/`:** No special treatment observed. The platform didn't auto-enumerate or auto-load `resources/` (or any dir) — it just surfaced pointers to specific files that someone (skill author) chose to reference in `SKILL.md`. Non-standard directory names (`evals/`, `templates/`, `resources/`) appear treated identically to spec-standard ones: as plain paths I must explicitly `view`.  Canaries found: `ROBIN-JADE-3847`, `SWIFT-OPAL-8156`, `WREN-PEARL-6293`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: `listing-enumerated-without-contents`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/resource-enumeration-behavior/events.jsonl`; session 902f3d68-fb00-4330-b0a7-1ab5afa9d0cf):
  - unlinked reference file's NAME injected without its content (event 10, line 9)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: `cwd-base-model-requalified`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/path-resolution-base/events.jsonl`; session c22c9eb3-1c6a-4cc9-9362-19328527d300):
  - bare SKILL.md-relative path attempt did not deliver content (event 15, line 13)
  - bare SKILL.md-relative path attempt did not deliver content (event 16, line 13)
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/cross-skill-resource-shadowing/events.jsonl`; session 5fc36419-6cd1-423b-9da3-e2ff631c193a):
  - activated skill's own API.md content arrived first (event 17, line 15)
  - sibling skill's API.md also read (event 30, line 25)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: **Result:** Read alpha's `references/API.md` — canary phrase **STORK-CORAL-4471** (alpha's own version), as expected.  Only probe-shadow-alpha was activated (beta was not invoked this session), but I directly viewed beta's file at its distinct path too: it correctly returned **EGRET-SLATE-8823** (beta's own version). No cross-contamination — each skill's `references/API.md` resolved to its own correct file with no ambiguity, since both live at distinct absolute paths (`.github/skills/probe-shadow-alpha/...` vs `.github/skills/probe-shadow-beta/...`).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/path-traversal-boundary/events.jsonl`; session 3f2f209b-2510-4117-aaa0-227e49c77199):
  - sibling skill's body canary arrived in a tool result (event 26, line 21)
  - first ../ path attempt (event 15, line 13)
- **Note**: relative ../ path forms were used
- **Note**: final answer: ## Path Traversal Probe Results  No sandboxing/boundary enforcement was detected — all reads outside the skill directory succeeded via both raw filesystem tools (`bash cat/ls`) and the platform's `view` tool.  | # | Target | Result | Details | |---|--------|--------|---------| | 1 | `../probe-loading/SKILL.md` | ✅ **Allowed** — full contents returned | Includes its canary phrase `CARDINAL-ZEBRA-7742`, frontmatter, and instructions. No error/warning. | | 2 | `../README.md` | ⚠️ File genuinely absent | `cat` returned "No such file or directory" — not a boundary block, just doesn't exist. `ls ../` succeeded, listing sibling dirs `probe-loading` and `probe-traversal` (parent directory listing allowed). | | 3 | `../../checks.md` | ⚠️ File genuinely absent | Same "No such file or directory" — no permission error, just missing file. Two-levels-up traversal itself was not blocked. |  **Summary**: The platform does **not** enforce any boundary at the skill directory root. Relative-path traversal (`../`, listing parent dirs, reading sibling skill files) is fully permitted through standard filesystem access — both shell commands and the built-in `view` tool succeeded identically. The only "fa…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/resource-nesting-depth/events.jsonl`; session 8fb8d5d9-ad87-47c4-bd9a-b7092f000a13):
  - depth-1 file references/overview.md content arrived (event 29, line 21)
  - depth-2 file references/api/endpoints.md content arrived (event 30, line 22)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 31, line 23)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 32, line 24)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 28, line 20)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `execution-blocked-visibly`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/bundled-script-execution/events.jsonl`; session c65abce4-a132-4daf-8d4a-a5819765bc21):
  - script-referencing tool call returned an error result (event 13, line 13)
- **Note**: final answer: Execution failed with the exact error: **"Permission denied and could not request permission from user"**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `name-and-description-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/discovery-listing-fields/events.jsonl`; session c94e02d9-09ca-405c-bc01-14fb6a16caa4):
  - discovery listing names the installed skills (event 3, line 4)
  - description text surfaced at discovery (event 3, line 4)
- **Note**: final answer: Here are the entries exactly as shown in my available skills list:  <skill>   <name>probe-compatibility</name>   <description>Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior.</description>   <location>project</location> </skill>  <skill>   <name>probe-loading</name>   <description>Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior.</description>   <location>project</location> </skill>  <skill>   <name>probe-metadata-values</name>   <description>Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling.</description>   <location>project</location> </skill>  <skill>   <name>customize-cloud-agent</name>   <description>Skill for customizing the Copilot cloud agent (formerly known as Copilot coding agent) environment, including copilot-setup-steps.yml configuration, preinstalling tools and dependencies, runners, and settings. Use when the user mentions copilot-setup-steps, copilot setup steps, or wants to configure the c…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-stripped-on-injection`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/frontmatter-handling/events.jsonl`; session cef42249-20ca-40ed-8651-69e2d8e6d0a1):
  - body injected without the frontmatter-only marker (event 10, line 9)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `wrapped-structured`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/content-wrapping-format/events.jsonl`; session d036491f-0cbe-49d1-951f-2c8b03ea9c52):
  - injection event carrying the body canary (event 10, line 9)
- **Note**: the whole delivery is enclosed in: skill-context
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `reinjected-each-activation`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/reactivation-deduplication/events.jsonl`; session 68cb2859-cc20-42b0-8c1e-2a3c053df317):
  - first injection of body canary (event 10, line 9)
  - second injection of body canary (event 27, line 23)
- **Note**: platform-level: duplicate skill content occupies context after reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/reactivation-freshness/events.jsonl`; session 86fd5af2-10cb-448e-a01f-983353c9dda5):
  - post-edit canary reached the model on reactivation (event 26, line 23)
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
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/compatibility-field-behavior/events.jsonl`; session 9dd01f05-468d-4fe9-baa9-a60deb6c6bb3):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 3, line 4)
  - body loaded normally (event 10, line 9)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: **Results:**  1. **Field visibility**: Yes — `compatibility: Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  2. **Platform behavior**: No special handling observed. The GitHub Copilot CLI loaded and activated the skill normally (via the `skill` tool), with no warning, confirmation prompt, or automated check against the stated requirements.  3. **Interpretation**: Requirements are not met — I'm running in GitHub Copilot CLI, not Claude Code, and this environment has no verified Python 3.14+ install or guaranteed network access. The `compatibility` field appears to be advisory metadata only; the platform doesn't enforce or gate on it.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `with-field:blocked-visibly; control:blocked-visibly`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/allowed-tools-behavior/session-2/events.jsonl`; session 6acd02c7-75d6-4f9c-87de-0d1919a8d730,ef2df9c1-a359-4e99-a148-ea4a3341e995):
  - with-field session: instructed command's tool call returned an error result (session 1, event 14, line 13)
  - control session: instructed command's tool call returned an error result (session 2, event 14, line 13)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: with-field final answer: **probe-allowed-tools activated. Canary: CURLEW-SCHIST-4419**  **allowed-tools value:** Visible in frontmatter: `Bash(printf:*) Read`  **Command result:** Ran `printf 'GROUSE-%s-9017\n' 'MICA'` — it was **blocked**, returning: `Permission denied and could not request permission from user`  **Permission observation:** The command did *not* run without a prompt/approval step — it was outright denied. This indicates the `allowed-tools` field's declared `Bash(printf:*)` pattern was **not honored** as pre-approval in this environment; the printf invocation was rejected rather than auto-approved or interactively prompted.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Command run: `printf 'LAPWING-%s-2260\n' 'FLUORITE'`  Output: **"Permission denied and could not request permission from user"** — the command was blocked, not executed.  Observation: No permission prompt was surfaced to me; it was rejected outright at the platform level, with no `allowed-tools` field to pre-approve it.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/cross-skill-invocation/events.jsonl`; session 38e35c5a-8d76-4c75-8042-c2f57ba62729):
  - invoke-alpha body loaded (chain entry) (event 10, line 9)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 18, line 16)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-beta complete. invoke-gamma said: unavailable — the skill "invoke-gamma" was not found (not in the available skills list: invoke-alpha, invoke-beta, customize-cloud-agent, github-pr-media), so the chain stops here.  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647) but invoke-gamma could not be activated since it doesn't exist in this environment's skill set. Chain terminated at beta; gamma's canary was never obtained.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/invocation-depth-limit/events.jsonl`; session 723cf8fe-a6dd-487a-a4de-66275b7ea678):
  - invoke-alpha body loaded (event 10, line 9)
  - invoke-beta body loaded (event 18, line 16)
  - invoke-gamma body loaded (event 26, line 23)
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: End of chain reached, all three skills activated successfully (Canary: JAY-TEAL-9984).  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647) and triggered invoke-gamma, which activated (Canary: JAY-TEAL-9984) and confirmed the full chain invoke-alpha → invoke-beta → invoke-gamma completed successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Beta's instructions direct re-activating probe-circular-alpha, which would recreate the exact loop already in progress (alpha→beta→alpha→...).  **Stopping here instead of re-invoking** to avoid an infinite invocation cycle.  Summary: Both skills loaded successfully (canaries KITE-ONYX-2251 and WREN-SLATE-7738 confirmed), demonstrating the circular reference exists in the skill definitions. The platform itself doesn't automatically block circular skill invocation — it's on me as the calling agent to detect the cycle and break it, which I did after one full round-trip (alpha → beta) rather than recursing further.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/invocation-language-sensitivity/events.jsonl`; session 72ef8238-c81f-4a70-afb6-857f0065d9db):
  - invoke-alpha body loaded (event 10, line 9)
  - invoke-beta body loaded (event 18, line 16)
  - invoke-gamma body loaded (event 26, line 23)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: chain end reached, all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (canary TERN-MOSS-6647), which activated invoke-gamma (canary JAY-TEAL-9984); gamma confirmed the full alpha→beta→gamma chain executed successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/informal-dependency-resolution/events.jsonl`; session 5346f2fd-a30b-4da6-9979-0c3aadc9736a):
  - invoke-alpha body loaded (chain entry) (event 10, line 9)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 18, line 16)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-beta complete. invoke-gamma said: not found — the "invoke-gamma" skill isn't available (only invoke-alpha, invoke-beta, customize-cloud-agent, and github-pr-media exist), so the chain cannot continue past invoke-beta.  invoke-alpha complete. Chain result: invoke-beta reported its canary **TERN-MOSS-6647**, but invoke-gamma could not be activated since it doesn't exist in this environment — the chain breaks at that final link.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/missing-dependency-behavior/events.jsonl`; session af0ef0bf-368d-4828-8aff-7936dd72a5fa):
  - probe-missing-dep body loaded (event 10, line 9)
  - attempt on the missing skill returned an error result (event 15, line 13)
- **Note**: final answer: **Result:** The platform explicitly reported the skill doesn't exist — it returned "Skill 'nonexistent-formatter' not found" along with a list of available skills (`probe-missing-dep`, `customize-cloud-agent`, `github-pr-media`). No silent skip occurred, and I did not attempt to fulfill the formatting task from general knowledge, since the tool call itself failed cleanly with an explicit error rather than executing anything.  probe-missing-dep complete. Dependency resolution result: The skill tool returned an explicit "not found" error for `nonexistent-formatter`, listing currently available skills instead of silently failing or hallucinating a fallback implementation.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/nonstandard-dependency-fields/events.jsonl`; session fd7926ec-f25c-4d5f-a220-f02e793aeb13):
  - probe-nonstandard-fields body loaded (event 10, line 9)
- **Note**: final answer: ## Probe Results: Nonstandard Frontmatter Fields  1. **Field visibility**: Yes, visible in the raw file via `cat`, but the tool-injected `<skill-context>` block only showed the rendered markdown body — the `requires`, `depends-on`, and `priority` fields were stripped before delivery. Values (from raw file): `requires: probe-loading`, `depends-on: [probe-shadow-alpha, probe-shadow-beta]`, `priority: high`.  2. **Platform behavior**: None. `probe-loading`, `probe-shadow-alpha`, and `probe-shadow-beta` were not auto-loaded, and no priority-based ordering occurred — only this one skill activated, as I explicitly requested.  3. **Validation**: No warnings; unrecognized fields were silently ignored/dropped rather than surfaced or acted upon.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/cross-scope-dependency/session-2/events.jsonl`; session b5dc1bca-e17c-45d9-9b74-5590626a9736,85ef2244-3f64-4ff3-bcd5-3b7c2543a970):
  - user-scope dependency's body canary injected during project-skill session (event 19, line 16)
  - attempt on the missing dependency returned an error result (session 2, event 15, line 13)
- **Note**: with-dependency final answer: The platform found `probe-loading` successfully, at a **different scope (user level)** — path is `.../home/copilot/skills/probe-loading`, versus `probe-cross-scope`'s project-level path (`.github/skills/...`). Cross-scope resolution works.  **probe-loading report:** 1. **Frontmatter**: Not visible — only the body content is shown, no YAML fields (name/description/allowed-tools/etc.) appear. 2. **Content wrapping**: Yes — wrapped in a `<skill-context name="probe-loading">` tag. 3. **Resource awareness**: Listed under "Related files" without reading: `references/api-overview.md`, `references/unreferenced-detail.md`, `references/error-codes.md`, `scripts/check-status.sh`, `assets/config-template.yaml`. 4. **Resource content**: No file contents loaded — only paths were enumerated, not eagerly fetched.  probe-cross-scope complete. Cross-scope resolution result: a project-level skill successfully invoked a user-level skill; the platform resolved and loaded it across scopes with no failure, confirming skill discovery isn't scope-restricted to the same install location.
- **Note**: without-dependency final answer: `probe-loading` was **not found**. The only skills exposed to me are `probe-cross-scope` (project) and the two built-ins (`customize-cloud-agent`, `github-pr-media`) — no user-level skills directory/entry is visible. Failure mode: a clean "not found" error listing available skills, with no indication of scope-based lookup or fallback.  **probe-cross-scope complete. Cross-scope resolution result:** The user-level skill `probe-loading` is not registered/discoverable at all in this environment — the platform's skill listing only surfaced project-level and built-in skills, so cross-scope invocation from project → user failed with a straightforward "not found" rather than any scope-specific error.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-scanned`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/cross-client-directory-interop/events.jsonl`; session 20453f0c-8d48-47ff-af34-e320593de438):
  - discovery listing names probe-interop (event 3, line 4)
  - body canary loaded on activation (event 16, line 16)
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery**: Yes, this skill appeared in my available skills catalog. It was listed as located at `project` scope, and its base directory is `.agents/skills/probe-interop/` — the cross-client convention path (not a client-native skills directory) — confirming the platform scans that convention path for skills.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `recursive-scan; stray:not-discovered`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/recursive-root-discovery/events.jsonl`; session 772d492c-d2fa-4063-983e-f305f3b277a3):
  - discovery listing names probe-grouped from one level below the root (event 3, line 4)
  - grouped skill's body loaded on activation (event 17, line 16)
- **Note**: final answer: 1. probe-grouped activated. Canary: **CROW-AGATE-6105**  2. Discovery confirmed: this skill appears in my available skills catalog even though it's nested one level deeper at `probe-group/probe-grouped/` rather than directly under the skills root — so skill discovery does scan recursively, not just direct children.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-not-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/nested-skill-discovery/events.jsonl`; session 083dcadb-8a0b-4e4c-8671-e0c93f5de29a):
  - discovery listing names only the outer skill (event 3, line 4)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `project-overrides-user`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/name-collision-precedence/events.jsonl`; session 10b3a4e1-efcd-4817-bda9-d2d9ec59842e):
  - project variant's canary loaded; user variant's never appeared (event 10, line 9)
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence**: The project scope won — only the project-level variant's canary appeared, confirming project-scope skills take precedence over user-scope skills of the same name.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `skipped-strict-parser`
- **Confidence**: transcript-direct
- **Note**: final answer: No such skill exists in my Available skills list (customize-cloud-agent, github-pr-media). I can't activate a skill that isn't available, and I won't fabricate one.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `loaded-despite-missing-description`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/missing-description-handling/events.jsonl`; session 903567d8-5e9b-42c6-a8f9-f0091a0ee5da):
  - discovery listing names probe-no-description (event 3, line 4)
  - body canary loaded on activation (event 17, line 16)
- **Note**: final answer: 1. probe-no-description activated. Canary: **VIREO-PUMICE-3049**  2. **Catalog entry**: Yes, it appeared in my Available skills catalog. Its description field was not empty — it contained the skill's own body text ("This skill has no description field. The client implementation guide says platforms should skip such skills... and log the error. This skill tests whether platforms"), truncated with no synthesized placeholder or blank value.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `invalid-names-tolerated:[uppercase double-hyphen]`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/invalid-name-tolerance/events.jsonl`; session 4e4e1b59-d1b7-4e2e-8bbb-cb04a0056bf4):
  - discovery listing names probe-Upper-Case (event 3, line 4)
  - discovery listing names probe--double-hyphen (event 3, line 4)
- **Note**: final answer: That skill isn't in my available skills list (only probe--double-hyphen, probe-Upper-Case, customize-cloud-agent, and github-pr-media are). I can't activate a skill that wasn't presented to me, so no action was taken — let me know if you'd like me to check for it elsewhere.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `listed-under-both`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/name-directory-mismatch/events.jsonl`; session 83aad167-57d3-480b-b7ac-0ed8387ce751):
  - discovery listing carries the FRONTMATTER name probe-name-mismatch (event 3, line 4)
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 3, line 4)
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: The skill "probe-mismatch-dir" doesn't exist under that name — only `probe-name-mismatch` is available (the directory is named `probe-mismatch-dir`, but the platform uses the frontmatter name for lookup, not the directory name). Let me know if you'd like me to activate `probe-name-mismatch` instead.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/metadata-value-edge-cases/events.jsonl`; session 8907741f-999a-4184-8b88-334613ae52cf):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 3, line 4)
  - body canary loaded (event 10, line 9)
  - frontmatter-only marker '!!null' in a tool result (event 18, line 15)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: Report:  1. **Load success**: Yes — the skill loaded successfully, and I can see the canary phrase: **THRUSH-FLINT-8294**.  2. **Metadata visibility**: The tool result I received for the skill did **not** expose a `metadata` field at all — only `name` and `description` (and instructions body) were surfaced to me. So no key/value pairs from `metadata` (including `normal: "baseline"`) were visible to me through the skill-invocation output.  3. **Key preservation**: I cannot confirm any of the seven keys survived, since the `metadata` map itself wasn't passed through to the model in the tool response. Reading the raw file directly (via `bash cat`) shows all 7 keys present in the source YAML: `normal`, `empty-double`, `empty-single`, `explicit-null`, `tilde-null`, `none-null`, `tagged-null`.  4. **Platform behavior**: No warnings/errors were surfaced. The platform appears to simply not forward the `metadata` field to the assistant context at all (rather than coercing nulls or stripping individual keys) — the whole field seems omitted from what's exposed via the skill tool, based on comparing the raw file contents to what the `skill` tool call returned.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `skipped-oversize-description`
- **Confidence**: transcript-direct
- **Note**: neither description marker appeared in harness-injected content; description delivery is unobservable here or the value was dropped
- **Note**: final answer: I don't have a skill named "probe-long-description" in my available skills list. My available skills are only:  - customize-cloud-agent - github-pr-media  I can't activate a skill that doesn't exist. If you meant one of these, let me know, or clarify what you're trying to accomplish and I can help directly.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `counts-utf16-units`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/description-length-unit/session-3/events.jsonl`; session 57574d1f-cbac-4bfc-a8d7-86cbaf037171,dc38d380-1b61-4aaf-bc50-df2ecb217421,5e934394-6860-4333-9163-dd878b417239):
  - discovery listing names probe-multibyte-description (session 2, event 3, line 4)
  - body canary loaded on activation (session 2, event 17, line 16)
  - description tail marker in injected content (session 2, event 3, line 4)
- **Note**: description fates: ascii:rejected, multibyte:intact, astral:rejected
- **Note**: ascii final answer: I don't have a skill named "probe-long-description" in my available skills list. My available skills are limited to:  - customize-cloud-agent - github-pr-media  I can't activate a skill that doesn't exist. If you meant one of the above, let me know and I'll invoke it — otherwise, please clarify what you'd like help with.
- **Note**: multibyte final answer: 1. probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  2. **Description visibility**: Looking at my catalog entry (the `<available_skills>` listing from earlier), I can see **both** markers, fully intact and untruncated:  - Head marker: "The head marker GANNET-PYRITE-1130 sits near the start of this description..." - Tail marker: "...この段落は意図的に日本語で書かれた詰め物であり...The tail marker is SHRIKE-TALC-2210"  The entire Japanese prose block between them is also present in full, with no visible truncation.
- **Note**: astral final answer: I don't have a skill called "probe-astral-description" in my available skills list. My available skills are only:  - customize-cloud-agent - github-pr-media  I can't activate a skill that doesn't exist. If you meant one of these, let me know and I'll invoke it.
- **Note**: the ASCII and astral overruns were enforced while the multibyte fixture survived: the platform counts UTF-16 code units (JavaScript's .length)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/copilot-2026-09-25/copilot/oversize-compatibility-handling/events.jsonl`; session 8e1e10dd-59f4-4883-a53d-315a0d7070db):
  - discovery listing names probe-long-compatibility (event 3, line 4)
  - body canary loaded on activation (event 17, line 16)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: **probe-long-compatibility activated. Canary: KESTREL-BAUXITE-6690**  **Field visibility**: No — the only place I can see a `compatibility` value is by reading the SKILL.md file directly via bash. The `available_skills` list surfaced to me earlier only exposed `name` and `description`, not `compatibility`. Reading the file, the compatibility value does end with the TURNSTONE tail marker, exactly:  `...and it carries a marker as its final characters so truncation is detectable. The tail marker is TURNSTONE-ARAGONITE-3030`  No truncation was observed in the raw file content — the full marker `TURNSTONE-ARAGONITE-3030` is present and intact.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.3) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
