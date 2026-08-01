---
title: "Antigravity CLI (headless)"
description: "Automated skill loading findings for Antigravity CLI (headless) (check list 0.2)."
date: 2026-08-01
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | Antigravity CLI (headless) |
| **Platform version** | 1.1.9 |
| **Check list version** | 0.2 |
| **Test date** | 2026-08-01 |
| **Model(s) observed** | Gemini 3.6 Flash (High) |
| **Environment** | Headless invocation via benchmark-runner + skillxp |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 18 of the 40 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

- [`path-resolution-base`](#path-resolution-base): The spec tells authors to reference files with relative paths from the skill root, but a path written that way fails here: paths resolve against the session's working directory, not the skill directory. In this run the model noticed the failure and requalified the path itself.

### Where behavior matches the spec

- [`discovery-reading-depth`](#discovery-reading-depth): Discovery reads only the skill's metadata, matching the spec's progressive disclosure model: name and description load at startup, and the body waits for activation. (Behavioral inference.)
- [`activation-loading-scope`](#activation-loading-scope): Activation loads the full SKILL.md body and nothing more, matching the spec's second disclosure stage: instructions at activation, resources only as a task needs them.
- [`eager-link-resolution`](#eager-link-resolution): Files linked from SKILL.md are not pre-fetched at activation; they load only when the task calls for them, which is the spec's on-demand model for resources. (Behavioral inference.)
- [`resource-enumeration-behavior`](#resource-enumeration-behavior): Reference files stay out of context until the model asks for them, matching the spec's rule that resources load on demand. (Behavioral inference.)
- [`resource-nesting-depth`](#resource-nesting-depth): The spec advises authors to keep file references one level deep but sets no platform limit, and none was observed: reference files stayed reachable at every tested depth through five levels.
- [`bundled-script-execution`](#bundled-script-execution): The spec presents scripts/ as executable code agents can run, and that held: the bundled script ran and its runtime-assembled output reached the model.
- [`discovery-listing-fields`](#discovery-listing-fields): The listing surfaces location in addition to name and description. The spec describes only those two fields loading at startup, but it does not forbid extras. (Behavioral inference.)
- [`frontmatter-handling`](#frontmatter-handling): The whole file, frontmatter included, reaches the model at activation because the model reads the raw file, matching the spec's description of loading the entire file.
- [`compatibility-field-behavior`](#compatibility-field-behavior): The spec makes compatibility informational (it indicates environment requirements) and assigns it no loading semantics. Consistent with that, a skill declaring a different product still loads here; authors should not expect the field to gate anything.

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this platform enforces it: the malformed skill never enters the catalog, though the file itself stays readable if the model goes looking. (Behavioral inference.)
- [`missing-description-handling`](#missing-description-handling): The spec requires a non-empty description, so a skill without one is invalid. This platform discovered and loaded it anyway. (Behavioral inference.)
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform tolerated every one: each rule-breaking name is discovered and usable. (Behavioral inference.)
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway, under the frontmatter name. (Behavioral inference.)
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated. (Behavioral inference.)
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway. (Behavioral inference.)

### Not exercised in this run

- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. The field's own effect went unobserved: the instructed command ran with and without it, so the platform's general permission posture is what allowed execution.


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: Metadata only (`metadata-only`)
- **Confidence**: behavioral-inference
- **Note**: transcript does not record injected context; discovery listing unobservable, verdict rests on the model not knowing the body canary
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: Body only (`body-only`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - model's tool call targets the skill's own path
  - body canary arrived in the tool result
- **Note**: direct-path navigation without a prior search implies a discovery listing the transcript does not record
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: No pre-fetching (`no-prefetch`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
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
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: Readable when the model looks (`resources-readable-on-demand`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - resources/ file content arrived via the model's own read
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`untouched`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: Based on the instructions in `SKILL.md`, here is the report on directory handling for `probe-nonstandard-dirs`:  ---  ### 1. Directory Awareness The platform's initial skill metadata prompt provided only the path to `SKILL.md`. Reading `SKILL.md` described the presence of three directories, which were confirmed via direct filesystem listing:  * `evals/` * `resources/` * `templates/`  ---  ### 2. Content Visibility * **Initial Context (Pre-Tool Execution)**: Subdirectory file listings and file contents are **not** pre-loaded, pre-injected, or automatically visible in the initial context window upon skill activation. * **Via File Tools**: With standard filesystem tools (`list_dir` / `view_file`), all subdirectories and their contents are fully accessible:   * `evals/` contains `evals.json`   * […
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: Bare path fails; model recovers (`cwd-base-model-requalified`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Runs**: 3 (cwd-base-model-requalified ×3)
- **Evidence**:
  - bare SKILL.md-relative path attempt did not deliver content
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Note**: verdict consistent across 3 runs
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: Got its own file (`own-resource-first`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - activated skill's own API.md content arrived first
  - sibling skill's API.md also read
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: I have activated the **probe-shadow-alpha** skill and completed its instructions:  ### 1. Contents of `probe-shadow-alpha` API Reference  File: `references/API.md`  ```markdown # API Reference (Alpha)  Shadow-alpha canary phrase: **STORK-CORAL-4471**  This file belongs to the `probe-shadow-alpha` skill. If you see this content when reading from the `probe-shadow-beta` skill's context, the platform is resolving resource paths ambiguously across skills.  ## Alpha Endpoints  - `GET /alpha/status` - Alpha health check - `POST /alpha/process` - Alpha processing ```  ---  ### 2. Canary Phrase Verification  * **Canary Phrase Identified**: **STORK-CORAL-4471** * **Result**: Confirmed as the expected canary phrase for the `probe-shadow-alpha` skill.  ---  ### 3. Cross-Skill Access Test (`probe-shadow-beta`)  Both skills are available in the workspace. Reading the corresponding reference file from the `probe-shadow-beta` skill was successful:  File: [`references/API.md`]
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: Reads outside the skill allowed (`outside-skill-read-allowed`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - sibling skill's body canary arrived in a tool result
  - first ../ path attempt
- **Note**: relative ../ path forms were used
- **Note**: final answer: ### Path Traversal Probe Results  I have executed the path traversal probe tests as specified in probe-traversal SKILL.md. Here are the detailed results for each attempt:  ---  #### 1. Sibling Skill (`../probe-loading/SKILL.md`) * **Path Attempted:** `/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-781866117/project/.agents/skills/probe-traversal/../probe-loading/SKILL.md` * **Did the platform allow the read?** Yes. * **Did you get the file contents?** Yes, successfully retrieved the complete content of probe-loading/SKILL.md (52 lines). * **Did the platform show an error or warning?** No. * **Did the platform silently return nothing?** No.  ---  #### 2. Parent Directory (`../README.md` & listing `../`) * **Path Attempted:** `/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-781866117/project/.agents/skills/probe-traversal/../` * **Did the platform allow the read?** Yes. * **Did you get the file contents / listing?**    * …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: All depths reachable (tested to 5) (`all-depths-accessible-through-5`)
- **Vehicle**: model-pull
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
- **Verdict**: Script ran; output returned (`script-executed`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - tool call references the bundled script
  - runtime-assembled output phrase arrived in a tool result
- **Note**: final answer: probe-script-execution activated. Canary: **REDSHANK-SYENITE-8807**  The bundled script `emit-canary.sh` printed the following output:  `GODWIT-BORNITE-5148`  ### Summary of Work - Activated the `probe-script-execution` skill by reading `SKILL.md`. - Executed the bundled script `scripts/emit-canary.sh` using bash. - Captured and reported the exact runtime output.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: Also surfaces location (`surfaces-beyond-description:[location]`)
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free catalog echo names the installed skills
  - description text surfaced at discovery
  - SKILL.md file path surfaced at discovery
- **Note**: this harness records no injected context; surfaced fields are inferred from the model's tool-free verbatim echo of its catalog
- **Note**: final answer: Here are the available skills listed in the environment:  - **antigravity-guide** (`/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-420419029/home/agy/.gemini/antigravity-cli/builtin/skills/antigravity_guide/SKILL.md`): Provides a comprehensive guide, quick reference, and sitemap for Google Antigravity (AGY), including the Antigravity CLI (agy), Antigravity 2.0, Antigravity IDE, Python SDK, slash commands, keybindings, and customizations (skills, rules, MCP, sidecars). Activate this skill when the user asks questions about how to use, configure, or customize Antigravity, AGY, the agy CLI, the Antigravity IDE, or Antigravity 2.0. - **probe-compatibility** (`/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-420419029/project/.agents/skills/probe-compatibility/SKILL.md`): Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior. - **probe-loading** (`/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-420419029/project/.agents/skills/probe-loading/SKILL.md`): Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: Visible (model reads the raw file) (`frontmatter-visible-via-raw-read`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - raw file read delivered frontmatter and body together
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: Raw file via model read (`raw-file-via-pull`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - body arrived as a file-read tool result
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: Model re-reads each time (`re-read-each-activation`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - first read of skill body
  - second read of skill body
- **Note**: model-level: on pull-vehicle harnesses re-loading is the model's choice, not platform policy
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: Edits picked up immediately (`fresh-content-served`)
- **Vehicle**: model-pull
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
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - body loaded normally
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: Here is the compatibility report based on the probe-compatibility SKILL.md:  ### 1. Field Visibility Yes, the `compatibility` field is visible in the YAML frontmatter of the skill definition file. * **Value:** `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  ### 2. Platform Behavior * **Handling:** The skill was activated normally without any special platform-level handling. * **Checks/Warnings:** No warnings were presented, no confirmation was requested, and no precondition checks were enforced prior to invoking the skill.  ### 3. Interpretation * **Platform:** Running Antigravity (a similar agentic AI coding assistant environment). * **Python Requirement:** Python `3.14.3` is installed in the system environment. * **Network Access:** Enabled/available. * **Conclusion:** The specified compatibility requirements are met in the current environment.
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
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  ### 1. `allowed-tools` Visibility Yes, the frontmatter of SKILL.md specifies: ```yaml allowed-tools: Bash(printf:*) Read ```  ### 2. Command Execution & Output Ran command: `printf 'GROUSE-%s-9017\n' 'MICA'`  **Output:** ``` GROUSE-MICA-9017 ```  ### 3. Permission Prompt / Approval Step Observation The command executed immediately and returned exit code `0` without any permission prompts, errors, or manual approval steps required.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  ### Command Output ``` LAPWING-FLUORITE-2260 ```  ### Observation The command `printf 'LAPWING-%s-2260\n' 'FLUORITE'` executed successfully with exit code 0 without requiring any user permission prompt or approval step.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  invoke-beta complete. invoke-gamma said: `invoke-gamma` is not available (skill path not found).  invoke-alpha complete. Chain result: - `invoke-alpha` executed successfully with canary **IBIS-RUST-3310**. - `invoke-beta` executed successfully with canary **TERN-MOSS-6647**. - `invoke-gamma` could not be activated because the skill is missing from `.agents/skills/`.  ---  ### Summary 1. Read instructions for `invoke-alpha/SKILL.md` and reported canary **IBIS-RUST-3310**. 2. Activated `invoke-beta/SKILL.md` and reported canary **TERN-MOSS-6647**. 3. Checked for `invoke-gamma` and confirmed it is not p…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  ---  invoke-beta activated. Canary: **TERN-MOSS-6647**  ---  invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  ---  invoke-beta complete. invoke-gamma said: End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully with canary **JAY-TEAL-9984**.  ---  invoke-alpha complete. Chain result: All three skills in the `invoke-alpha` -> `invoke-beta` -> `invoke-gamma` chain were successfully executed in sequence, reporting the following canary values: - `invoke-alpha`: **IBIS-RUST-3310** - `invoke-beta`: **TERN-MOSS-6647** - `invoke-gamma`: **JAY-TEAL-9984**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  ---  ### Execution & Circular Dependency Summary  1. **probe-circular-alpha** was initialised (Canary: `KITE-ONYX-2251`). 2. **probe-circular-alpha** triggered **probe-circular-beta** (Canary: `WREN-SLATE-7738`). 3. **probe-circular-beta** requested re-activation of **probe-circular-alpha**, completing the loop (`probe-circular-alpha` $\rightarrow$ `probe-circular-beta` $\rightarrow$ `probe-circular-alpha`). 4. **Circular dependency detected**: Re-entering `probe-circular-alpha` was halted to prevent infinite recursion.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  ---  ### Step 1: Executing `invoke-alpha`  - **Status**: `invoke-alpha` activated. - **Canary**: **IBIS-RUST-3310** - **Action**: Activating `invoke-beta` skill...  ---  ### Step 2: Executing `invoke-beta`  - **Status**: `invoke-beta` activated. - **Canary**: **TERN-MOSS-6647** - **Action**: Activating `invoke-gamma` skill...  ---  ### Step 3: Executing `invoke-gamma`  - **Status**: `invoke-gamma` activated. - **Canary**: **JAY-TEAL-9984** - **Report**: End of chain reached. All three skills in the `invoke-alpha` -> `invoke-beta` -> `invoke-gamma` chain activated successfully.  ---  ### Step 4: `invoke-beta` Completion  - **Report**: invoke-beta complete.   - **invoke-gamma said**: `invoke-gamma` activated with Canary **JAY-TEAL-9984** and reported reaching the end of the chain, confirming that all three skills in the chain activated successfully.  ---  ### Step 5: `invoke-alpha` Completion…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  ---  ### Step 2: Activating `invoke-beta`  invoke-beta activated. Canary: **TERN-MOSS-6647**  Attempting to activate `invoke-gamma`: `invoke-gamma` skill was not found in the available skills.  invoke-beta complete.   invoke-gamma said: Skill `invoke-gamma` is not available.  ---  ### Step 3: Completing `invoke-alpha`  invoke-alpha complete.   Chain result: `invoke-beta` was activated with Canary **TERN-MOSS-6647**, but the chain terminated at `invoke-gamma` because the `invoke-gamma` skill does not exist in the workspace.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: Reported missing without attempting (`reported-without-attempt`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-missing-dep body loaded
  - model surfaced the missing dependency in its own text without attempting it
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: probe-missing-dep activated. Canary: **GULL-IRON-4492**  ### Dependency Resolution Status Attempted to activate the required `nonexistent-formatter` skill. However, `nonexistent-formatter` is not listed among the available skills in the platform environment (only `antigravity-guide` and `probe-missing-dep` are available).  - **Platform status**: The skill `nonexistent-formatter` does not exist in the available skills registry. - **Action taken**: Identified that the skill is missing from available capabilities and proceeded with reporting the missing dependency status.  probe-missing-dep complete. Dependency resolution result: The referenced skill `nonexistent-formatter` is not installed or available in the environment.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: I have activated the **probe-nonstandard-fields** skill by viewing its `SKILL.md` file and reviewing the nonstandard frontmatter metadata. Here is the probe report:  ---  ### 1. Field Visibility Yes, the nonstandard frontmatter fields are fully visible when reading `SKILL.md`. Their exact parsed values are:  - **`requires`**: `probe-loading` - **`depends-on`**:   - `probe-shadow-alpha`   - `probe-shadow-beta` - **`priority`**: `high`  ---  ### 2. Platform Behavior The platform **silently ignored** these nonstandard fields: - **No Dependency Auto-Loading**: The platform did not attempt to automatically load or inject `probe-loading`, `probe-shadow-alpha`, or `probe-shadow-beta` upon activating `probe-nonstandard-fields`. - **No Priority Re-ordering**: The platform did not modify skill ordering or alter execution logic based on `priority: high`.  ---  ### 3. Validation The platform **accepted the nonstandard fiel…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: Resolved across scopes; missing dependency reported, not attempted (`resolved-across-scopes; missing:reported-without-attempt`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - user-scope dependency's body canary pulled during project-skill session
  - model surfaced the missing dependency in its own text without attempting it
- **Note**: with-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  ### 1. Installation Scope `probe-cross-scope` is a **project-level skill** located at SKILL.md.  ### 2. Cross-Scope Resolution of `probe-loading` - **Found**: Yes, the platform successfully located `probe-loading`. - **Scope**: `probe-loading` was resolved at a **different scope (user-level)** at SKILL.md. - **Activation Status**: Successfully loaded and read the `probe-loading` instructions.  probe-cross-scope complete. Cross-scope resolution result: Project-level skill successfully identified and activated the user-level skill `probe-loading`.
- **Note**: which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  ### Installation Scope Report * **Scope**: Project-level skill * **Location**: /var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2290681380/project/.agents/skills/probe-cross-scope/SKILL.md  ### Activation of `probe-loading` Skill * **Found**: No * **Failure Mode**: The skill `probe-loading` is not present in the available skills list in the environment context, nor does it exist on the local file system at user or project levels. Consequently, it could not be resolved or loaded.  probe-cross-scope complete. Cross-scope resolution result: The cross-scope dependency `probe-loading` could not be resolved because it is missing/not installed in the user or project environment.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path is the native directory (`convention-is-native-dir`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-interop
  - body canary loaded on activation
- **Note**: this platform's native project skills directory IS .agents/skills, so the check cannot separate convention support from native scanning
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery**: Yes, the `probe-interop` skill appeared in my available skills catalog at the path `/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4246582053/project/.agents/skills/probe-interop/SKILL.md`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Direct children only; stray file ignored (`direct-children-only; stray:not-discovered`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - body loaded via file access, not catalog activation
- **Note**: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it
- **Note**: final answer: 1. probe-grouped activated. Canary: **CROW-AGATE-6105**  2. **Discovery**: The `probe-grouped` skill was located inside `.agents/skills/probe-group/probe-grouped/SKILL.md`, nested one directory level below the skills root inside the `probe-group` directory. Note that while only `antigravity-guide` was listed in the system prompt's initial `Available skills` list, the skill file was successfully discovered and loaded via directory scanning.
- **Fallback behavior**: Observed in-run: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: Not discovered (`nested-skill-not-discovered`)
- **Confidence**: behavioral-inference
- **Evidence**:
  - model's tool-free listing answer names only the outer skill
- **Note**: final answer: antigravity-guide probe-deep-nesting
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: Project scope wins (`project-overrides-user`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - project variant's canary loaded; user variant's never appeared
- **Note**: catalog visibility unrecorded on this harness; whether the platform or the model resolved the collision is not directly observable
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence**: Project scope won over user scope.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Not cataloged; file still readable (`unlisted-but-reachable`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - body canary loaded on activation
- **Note**: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  ### Description Visibility  - **Catalog Visibility**: The `probe-malformed-yaml` skill was **not present** in the system prompt's `Available skills` catalog (only `antigravity-guide` was listed). The strict YAML parser likely failed or skipped loading this skill frontmatter due to the unquoted colon in `description: Use when: asked to probe...`. - **On-Disk Inspection**: When inspecting `SKILL.md` directly, the description line is unquoted and intact: `description: Use when: asked to probe malformed yaml parsing behavior`.
- **Fallback behavior**: content remained reachable by direct file read even though the catalog omits the skill. Observed in-run: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-missing-description`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-no-description
  - body canary loaded on activation
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, this skill appears in the Available skills catalog list. The description text associated with it is completely empty (`- probe-no-description (.../SKILL.md): `).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: All three invalid names tolerated (`all-invalid-names-tolerated`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-Upper-Case
  - tool-free listing answer names probe--double-hyphen
  - tool-free listing answer names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  **Catalog identity**: In the system prompt's Available skills catalog, this skill appears under the full untruncated 72-character name: `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Frontmatter name wins (`frontmatter-name-identity`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-name-mismatch
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: The skill is listed under the frontmatter name `probe-name-mismatch` in the available skills catalog (not `probe-mismatch-dir` nor both). 3. **Validation**: The platform accepted the name/directory mismatch silently without emitting any warning to the model context.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: Loaded fine (`loaded-despite-edge-case-metadata`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - body canary loaded
  - frontmatter-only marker '!!null' in a tool result
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: I have activated and evaluated the `probe-metadata-values` skill.  **Canary Phrase**: **THRUSH-FLINT-8294**  ---  ### Report  1. **Load Success**:     - **Yes**, the skill loaded successfully. The file `SKILL.md` was accessed and read without rejection or parsing errors.  2. **Metadata Visibility**:    - The `metadata` field in the frontmatter contains the following key-value pairs:      - `normal`: `"baseline"` (Control string)      - `empty-double`: `""` (Empty string)      - `empty-single`: `''` (Empty string)      - `explicit-null`: `null` (Literal YAML null)      - `tilde-null`: `~` (YAML tilde shorthand for null)      - `none-null`: `None` (Unquoted literal `None`)      - `tagged-null`: `!!null null` (Explicit YAML tag `!!null` with value `null`)  3. **Key Preservation**:    - **All 7 keys are present**. None of the keys were stripped or omitted during indexing or file access.  4. **Platform Behavior**:    - The platfo…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read Observed in-run: final answer: I have activated and evaluated the `probe-metadata-values` skill.  **Canary Phrase**: **THRUSH-FLINT-8294**  ---  ### Report  1. **Load Success**:     - **Yes**, the skill loaded successfully. The file `SKILL.md` was accessed and read without rejection or parsing errors.  2. **Metadata Visibility**:    - The `metadata` field in the frontmatter contains the following key-value pairs:      - `normal`: `"baseline"` (Control string)      - `empty-double`: `""` (Empty string)      - `empty-single`: `''` (Empty string)      - `explicit-null`: `null` (Literal YAML null)      - `tilde-null`: `~` (YAML tilde shorthand for null)      - `none-null`: `None` (Unquoted literal `None`)      - `tagged-null`: `!!null null` (Explicit YAML tag `!!null` with value `null`)  3. **Key Preservation**:    - **All 7 keys are present**. None of the keys were stripped or omitted during indexing or file access.  4. **Platform Behavior**:    - The platfo…

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-oversize-description`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-long-description
  - body canary loaded on activation
- **Note**: the tail marker surfaced in the model's own text (nothing harness-injected records it); on this harness delivery is only inferable
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  ### Description visibility  Looking at the catalog entry in the system prompt's **Available skills** list, **both** the head marker (`SANDERLING-GNEISS-1010`) and the tail marker (`WHIMBREL-DOLOMITE-2020`) are visible intact.  **Exact catalog description:** > Benchmark skill whose description deliberately runs past the spec's 1024-character limit, to test oversize field handling. Use when asked to probe oversize description handling. The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end, so testers can tell whether the platform accepted the field intact, rejected the skill, or silently truncated the value somewhere between the markers. Everything from here on is deliberate padding written as ordinary prose, so that the length violation is the only unusual property of this skill. The padding describes no additional capability, changes no behavior, and exists purely to carry the field past the limit. It is phrased as complete sentences because some parsers and listings treat description text as display copy, and display copy that looks like natur…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: Loaded anyway (`loaded-despite-oversize-compatibility`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - tool-free listing answer names probe-long-compatibility
  - body canary loaded on activation
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No, the `compatibility` field value for this skill is not visible anywhere in the initial prompt or context outside of reading SKILL.md directly. The system prompt's `Available skills` listing only surfaces the skill's name, file path, and `description` field.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.2) for what each check evaluates.
