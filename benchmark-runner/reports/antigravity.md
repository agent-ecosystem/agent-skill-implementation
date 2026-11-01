# Platform Loading Implementation: Antigravity CLI (headless)

| | |
|---|---|
| **Platform** | Antigravity CLI (headless) |
| **Platform version** | 1.4.12 |
| **Check list version** | 0.4 |
| **Test date** | 2026-11-01 |
| **Model(s) observed** | Gemini 4 Argon (Medium) |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

> **Changed since the previous observation**: 5 verdicts differ from the previous run of the checks on this platform: `circular-invocation-handling`, `directory-naming-divergence`, `malformed-yaml-tolerance`, `missing-dependency-behavior`, `unrecognized-directory-handling`. Each finding below states the earlier verdict and release.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 21 of the 46 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

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
- [`description-length-unit`](#description-length-unit): The spec caps description at 1024 characters without defining the unit. This platform does not enforce the limit at all, so every fixture under 1024 code points loaded intact and the counting unit is moot. (Behavioral inference.)
- [`name-length-unit`](#name-length-unit): The spec caps name at 64 characters without defining the unit and allows unicode lowercase alphanumeric characters with an ASCII parenthetical, which reads two ways; its skills-ref reference validator accepts any Unicode alphanumeric and counts code points. This platform does not enforce the cap at all, so the counting unit is moot. (Behavioral inference.)

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this fixture's frontmatter does not parse (an unquoted colon). The platform tolerated the error: the skill is discovered and loads anyway. (Behavioral inference.)
- [`missing-description-handling`](#missing-description-handling): The spec requires a non-empty description, so a skill without one is invalid. This platform discovered and loaded it anyway. (Behavioral inference.)
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform tolerated every one: each rule-breaking name is discovered and usable. (Behavioral inference.)
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway, under the frontmatter name. (Behavioral inference.)
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated. (Behavioral inference.)
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway. (Behavioral inference.)

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
- **Confidence**: behavioral-inference
- **Note**: transcript does not record injected context; discovery listing unobservable, verdict rests on the model not knowing the body canary
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/activation-loading-scope/transcript_full.jsonl`; session 8c91e306-5d20-4610-b7e7-f5b3c21395b3):
  - model's tool call targets the skill's own path (event 4, line 2)
  - body canary arrived in the tool result (event 5, line 3)
- **Note**: direct-path navigation without a prior search implies a discovery listing the transcript does not record
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/eager-link-resolution/transcript_full.jsonl`; session d313417c-dcaf-4459-a202-e9a878686438):
  - skill body loaded (event 5, line 3)
  - references/setup-guide.md arrived only via the model's own read (event 19, line 9)
  - references/troubleshooting.md arrived only via the model's own read (event 20, line 10)
  - references/unlinked-data.md arrived only via the model's own read (event 21, line 11)
- **Note**: model read [references/setup-guide.md references/troubleshooting.md references/unlinked-data.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Directory Recognition

#### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: `no-enumeration-at-activation`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: `resources-untouched`
- **Changed since previous observation**: was `resources-readable-on-demand` on 1.2.11 (2026-09-25)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: `untouched`
- **Changed since previous observation**: was `readable-on-demand:[evals/ templates/]` on 1.2.11 (2026-09-25)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: Here is the report on how the platform handles the directories in [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3152684176/project/.agents/skills/probe-nonstandard-dirs/SKILL.md):  ### 1. Directory Awareness * **Pre-activation (System Prompt Enumeration):** **None.** The platform only enumerates the skill's name, description, and the absolute path to `SKILL.md` (`/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3152684176/project/.agents/skills/probe-nonstandard-dirs/SKILL.md`). No subdirectories inside the skill folder are listed or enumerated automatically. * **Post-activation (After reading `SKILL.md`):** I am only aware of `evals/`, `templates/`, and `resources/` because they are explicitly mentioned in the markdown body of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3152684176/project/.agents/skills/probe-nonstandard-dirs/SKILL.md#L16-L19). * **Platform System Prompt Preamble:** Notably, the platform's general `<skills>` instructions mention four example subdirectory names that complex skills might contain: `scripts/`, `examples/`, `resources/`, and `references/`.  ### 2. Content Visibi…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: `cwd-base-model-requalified`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/path-resolution-base/transcript_full.jsonl`; session 4ddd27ff-7346-4bb2-a5e1-e0af379b1aa1):
  - bare SKILL.md-relative path attempt did not deliver content (event 8, line 4)
  - bare SKILL.md-relative path attempt did not deliver content (event 9, line 4)
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/cross-skill-resource-shadowing/transcript_full.jsonl`; session 53c8b70d-580c-4616-bd4f-ba70dc73d72f):
  - activated skill's own API.md content arrived first (event 9, line 5)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: final answer: I have activated the `probe-shadow-alpha` skill ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3445753810/project/.agents/skills/probe-shadow-alpha/SKILL.md)) and followed its instructions:  ### 1. Contents of [API.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3445753810/project/.agents/skills/probe-shadow-alpha/references/API.md)  ```markdown # API Reference (Alpha)  Shadow-alpha canary phrase: **STORK-CORAL-4471**  This file belongs to the `probe-shadow-alpha` skill. If you see this content when reading from the `probe-shadow-beta` skill's context, the platform is resolving resource paths ambiguously across skills.  ## Alpha Endpoints  - `GET /alpha/status` - Alpha health check - `POST /alpha/process` - Alpha processing ```  ### 2. Canary Phrase Verification - **Canary phrase observed:** `STORK-CORAL-4471` - **Result:** Confirmed this is the `probe-shadow-alpha` skill's version of `references/API.md`.  ### 3. Cross-Skill Status - Only `probe-shadow-alpha` is currently activated (`probe-shadow-beta` has not been activated yet, though its path is distinct on disk under `.agents/skills/probe-shadow-beta/`).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/path-traversal-boundary/transcript_full.jsonl`; session 51aa19e2-b1aa-43de-959e-7c10414a977f):
  - sibling skill's body canary arrived in a tool result (event 11, line 5)
  - first ../ path attempt (event 8, line 4)
- **Note**: relative ../ path forms were used
- **Note**: final answer: I have activated the [`probe-traversal`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-97048590/project/.agents/skills/probe-traversal/SKILL.md) skill and executed its path traversal checks.  Note that the `read_file` tool requires an absolute path (passing a bare relative path like `../probe-loading/SKILL.md` returns an error: `failed to parse path: ../probe-loading/SKILL.md must be an absolute path: path is not absolute`). When resolved against the skill's directory (`/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-97048590/project/.agents/skills/probe-traversal/`), here are the results for each target:  ---  ### 1. Sibling Skill (`../probe-loading/SKILL.md`) * **Target Path**: [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-97048590/project/.agents/skills/probe-loading/SKILL.md) (`/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-97048590/project/.agents/skills/probe-traversal/../probe-loading/SKILL.md`) * **Status**: **Success** * **Did the platform allow the read?** Yes. Skills are read via the standard workspace `read_file` tool rather than a skill-scoped file reader, so any file with…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/resource-nesting-depth/transcript_full.jsonl`; session 9979eda3-1862-441e-8933-97296d906985):
  - depth-1 file references/overview.md content arrived (event 13, line 5)
  - depth-2 file references/api/endpoints.md content arrived (event 14, line 6)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 15, line 7)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 16, line 8)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 17, line 9)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `script-executed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/bundled-script-execution/transcript_full.jsonl`; session eff88414-4476-476d-80ca-b938bfeab6f0):
  - tool call references the bundled script (event 7, line 4)
  - runtime-assembled output phrase arrived in a tool result (event 8, line 5)
- **Note**: final answer: probe-script-execution activated. Canary: **REDSHANK-SYENITE-8807**  Output from [emit-canary.sh](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4109253473/project/.agents/skills/probe-script-execution/scripts/emit-canary.sh): `GODWIT-BORNITE-5148`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: final answer: 1. probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  2. **File awareness**: Aside from the skill's main file ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3741812504/project/.agents/skills/probe-bulk-files/SKILL.md)) provided in the available skills catalog, I was told about **none** of the bundled supporting files when the skill was activated.  3. **Count**: 0 bundled supporting files (or 1 file if counting [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3741812504/project/.agents/skills/probe-bulk-files/SKILL.md) itself).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `surfaces-beyond-description:[location]`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/discovery-listing-fields/transcript_full.jsonl`; session 63078fa9-c49b-4d36-bece-79cf05d1241f):
  - tool-free catalog echo names the installed skills (event 3, line 2)
  - description text surfaced at discovery (event 3, line 2)
  - SKILL.md file path surfaced at discovery (event 3, line 2)
- **Note**: this harness records no injected context; surfaced fields are inferred from the model's tool-free verbatim echo of its catalog
- **Note**: final answer: Here is the exact verbatim content of the `Available skills:` section from my system prompt:  ```text Available skills: - agy-customizations (/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3397521357/home/agy/.gemini/antigravity-cli/builtin/skills/agy-customizations/SKILL.md): Comprehensive guide and reference for the Antigravity Customization System. Use to explain how customizations work, their loading priority, discovery mechanisms, and to guide the creation of skills, rules, plugins, hooks, and MCP servers. - antigravity-guide (/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3397521357/home/agy/.gemini/antigravity-cli/builtin/skills/antigravity_guide/SKILL.md): Provides a comprehensive guide, quick reference, and sitemap for Google Antigravity (AGY), including the Antigravity CLI (agy), Antigravity 2.0, Antigravity IDE, Python SDK, slash commands, keybindings, and customizations (skills, rules, MCP, sidecars). Activate this skill when the user asks questions about how to use, configure, or customize Antigravity, AGY, the agy CLI, the Antigravity IDE, or Antigravity 2.0. - probe-compatibility (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-33…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-visible-via-raw-read`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/frontmatter-handling/transcript_full.jsonl`; session a4a03c10-d303-48e1-9600-2f6232d1d42e):
  - raw file read delivered frontmatter and body together (event 5, line 3)
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-file-via-pull`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/content-wrapping-format/transcript_full.jsonl`; session 9246148c-5586-4159-bc70-e63e3ca3c521):
  - body arrived as a file-read tool result (event 5, line 3)
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: `injection-not-recorded`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: the platform records no injected context, so whether activation states the path cannot be observed; the model's own reads show the path it used
- **Note**: final answer: Here is the structured loading behavior report for the `probe-loading` skill ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1305085494/project/.agents/skills/probe-loading/SKILL.md)):  1. **Frontmatter visibility**:    - **At discovery time (in system prompt)**: Only `name`, the path to `SKILL.md`, and `description` were visible in the `<skills>` section of the system prompt (`- probe-loading (<path>/SKILL.md): <description>`).    - **At activation time (via `read_file` on [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1305085494/project/.agents/skills/probe-loading/SKILL.md#L1-L9))**: Yes, the full raw YAML frontmatter is visible because the file was read directly from disk. The visible fields are:      - `name`: `probe-loading`      - `description`: `Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior.`      - `allowed-tools`: `Read Write`      - `compatibility`: `Requires filesystem access`      - `metadata`: `benchmark: "true"`, `version: "1.0"`  2. **Content wrapping**:    - **At discovery time**: The list…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `re-read-each-activation`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/reactivation-deduplication/transcript_full.jsonl`; session 9ff27703-50f2-4d27-bf95-7d6466b5bcc2):
  - first read of skill body (event 5, line 3)
  - second read of skill body (event 13, line 8)
- **Note**: model-level: on pull-vehicle harnesses re-loading is the model's choice, not platform policy
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/reactivation-freshness/transcript_full.jsonl`; session 486978f4-8d46-4501-ab6c-7a8d7cb898e3):
  - post-edit canary reached the model on reactivation (event 13, line 8)
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
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/compatibility-field-behavior/transcript_full.jsonl`; session ff5fb74b-2e77-4390-b375-53e641a88729):
  - body loaded normally (event 5, line 3)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: Here is the report from activating the `probe-compatibility` skill ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1920143045/project/.agents/skills/probe-compatibility/SKILL.md)):  1. **Field visibility**:    - **In the system prompt skill catalog**: No — only the `name`, `description`, and file path were surfaced in the pre-loaded skills list.    - **When reading [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1920143045/project/.agents/skills/probe-compatibility/SKILL.md#L1-L5)**: Yes — the raw YAML frontmatter is visible, and the `compatibility` field has the value:      > `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  2. **Platform behavior**:    - The skill was discovered and activated normally with **no special handling** based on the `compatibility` field. The platform did not display a warning, prompt for confirmation, gate activation, or automatically pre-check the requirements.  3. **Your interpretation**:    - **Yes**, the compatibility requirements are met in the current environment:      - **Product**: Antigravity is an agentic coding assistant ("or sim…
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/allowed-tools-behavior/session-2/transcript_full.jsonl`; session b46405ce-a309-45b2-bb2c-fa7ecee147e8,eb60931a-55b0-43ae-8150-eed17e9b3c7b):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 8, line 5)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 8, line 5)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. **Allowed-tools value**: Yes, in the YAML frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1631601064/project/.agents/skills/probe-allowed-tools/SKILL.md#L4) (though it was not exposed in the pre-loaded skills metadata list prior to reading the file), the value is:    `Bash(printf:*) Read`  3. **Command output**:    ```text    GROUSE-MICA-9017    ```  4. **Execution observation**: As far as I can observe from the tool interface, the command executed directly and returned exit code 0 without any visible permission prompt, block, or extra approval step (note that there was a ~15-second gap between reading the skill file and the command execution timestamp, which could reflect host/harness-level approval or scheduling outside the tool output).
- **Note**: control final answer: 1. probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  2. Command output verbatim: ``` LAPWING-FLUORITE-2260 ```  3. As far as I can observe from my perspective, the command executed directly via `run_command` and returned exit code `0` (~2 seconds elapsed between creation and completion, `23:21:42` to `23:21:44`) without any observable permission prompt or blocking error in my tool response. (Whether a UI-level approval prompt was shown or auto-approved on the host harness is not directly visible in the tool output.)
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: `executed-regardless-of-spelling`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/allowed-tools-name-matching/session-3/transcript_full.jsonl`; session b93d2d1a-31fe-4cac-99a4-d7e50ea66331,66b0112f-0200-40bc-b918-d4333da89169,4468feeb-8c9b-462e-a431-757034a691c7):
  - spec-style twin: runtime-assembled phrase arrived in a tool result (session 1, event 8, line 5)
  - lowercase twin: runtime-assembled phrase arrived in a tool result (session 2, event 8, line 5)
  - shell twin: runtime-assembled phrase arrived in a tool result (session 3, event 8, line 5)
- **Note**: spec-style twin final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. **Visibility of `allowed-tools`**: In the system prompt's `<skills>` listing, the `allowed-tools` field was not included (only the skill name, path, and description were shown). However, when reading the raw [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4008251090/project/.agents/skills/probe-allowed-tools/SKILL.md#L1-L5) file, the YAML frontmatter shows:    `allowed-tools: Bash(printf:*) Read`  3. **Command output**:    ```    GROUSE-MICA-9017    ```  4. **Execution observation**: The command executed directly and returned exit code `0` without any observable permission prompt or approval step.
- **Note**: lowercase twin final answer: 1. probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  2. **Visibility of `allowed-tools`:**    - In the system prompt's `<skills>` summary list, only `name`, path, and `description` are shown (`allowed-tools` is not included there).    - Upon reading [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1985818770/project/.agents/skills/probe-allowed-tools-lowercase/SKILL.md#L1-L5) directly, the YAML frontmatter contains `allowed-tools: bash`.  3. **Command output:** ```text CHOUGH-TUFA-7180 ```  4. **Execution observation:**    - From my perspective as the agent, the `run_command` tool call executed directly and returned exit code `0` without any visible permission error or intermediate approval step required on my side (elapsed time between tool invocation and completion was ~2 seconds).
- **Note**: shell twin final answer: 1. probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  2. **Visibility of `allowed-tools`:**    - In the system prompt's pre-loaded skill list (`<skills>`), only the skill's `name`, path, and `description` are shown (`allowed-tools` is not surfaced there).    - Upon reading the raw [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3646008976/project/.agents/skills/probe-allowed-tools-shell/SKILL.md#L1-L5) YAML frontmatter, `allowed-tools` is visible on line 4 with the value: `shell`.  3. **Command Output (verbatim):** ```text ROOK-GABBRO-8841 ```  4. **Execution Observation:**    The command (`run_command`) executed immediately and exited with code `0` (in ~2 seconds, from `23:22:45` to `23:22:47`) without any observable permission prompt, rejection, or approval wait step.
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/cross-skill-invocation/transcript_full.jsonl`; session dea8efc1-ca3a-4aba-9155-543e6273923b):
  - invoke-alpha body loaded (chain entry) (event 5, line 3)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 8, line 5)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  > [!WARNING] > The `invoke-gamma` skill could not be activated because it is not present in the available skills or in [`.agents/skills/`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3959864289/project/.agents/skills).  invoke-beta complete. invoke-gamma said: `invoke-gamma` could not be found or activated (skill file does not exist).  invoke-alpha complete. Chain result: - **invoke-alpha** ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3959864289/project/.agents/skills/invoke-alpha/SKILL.md)): Activated with canary **IBIS-RUST-3310** and invoked `invoke-beta`. - **invoke-beta** ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3959864289/project/.agents/skills/invoke-beta/SKILL.md)): Activated with canary **TERN-MOSS-6647** and attempted to invoke `invoke-gamma`. - **invoke-gamma**: Not found in available skills or the workspace filesystem.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/invocation-depth-limit/transcript_full.jsonl`; session 6a1e42a8-3f6d-4dcf-b061-4bf700ac23cf):
  - invoke-alpha body loaded (event 5, line 3)
  - invoke-beta body loaded (event 8, line 5)
  - invoke-gamma body loaded (event 11, line 7)
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: invoke-gamma activated with canary **JAY-TEAL-9984** and reported that the end of the chain (`invoke-alpha -> invoke-beta -> invoke-gamma`) was reached successfully.  invoke-alpha complete. Chain result: `invoke-beta` activated with canary **TERN-MOSS-6647** and invoked `invoke-gamma`, which activated with canary **JAY-TEAL-9984** and confirmed all three skills in the `invoke-alpha -> invoke-beta -> invoke-gamma` chain activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Changed since previous observation**: was `reinvocation-no-reload` on 1.2.11 (2026-09-25)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: 1. probe-circular-alpha activated. Canary: **KITE-ONYX-2251** 2. Activating probe-circular-beta as directed. 3. probe-circular-beta activated. Canary: **WREN-SLATE-7738** 4. Activating probe-circular-alpha as directed.  ### Outcome A circular dependency loop was detected between [probe-circular-alpha/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3059950930/project/.agents/skills/probe-circular-alpha/SKILL.md) and [probe-circular-beta/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3059950930/project/.agents/skills/probe-circular-beta/SKILL.md): - `probe-circular-alpha` directs activation of `probe-circular-beta`. - `probe-circular-beta` directs activation back to `probe-circular-alpha`.  Because `probe-circular-alpha` was already activated in the current call chain, re-activating it was halted to prevent an infinite invocation loop. Both skills were read once and their respective canary values have been reported above.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/invocation-language-sensitivity/transcript_full.jsonl`; session 4b250d41-7743-4259-913b-dfa896f4fa3d):
  - invoke-alpha body loaded (event 5, line 3)
  - invoke-beta body loaded (event 8, line 5)
  - invoke-gamma body loaded (event 11, line 7)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: invoke-gamma activated with canary **JAY-TEAL-9984** and reported that the end of the chain was reached with all three skills in the `invoke-alpha` -> `invoke-beta` -> `invoke-gamma` chain activated successfully.  invoke-alpha complete. Chain result: `invoke-beta` activated with canary **TERN-MOSS-6647** and invoked `invoke-gamma`, which activated with canary **JAY-TEAL-9984** and confirmed that all three skills in the `invoke-alpha` -> `invoke-beta` -> `invoke-gamma` chain activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/informal-dependency-resolution/transcript_full.jsonl`; session b479b9e6-375f-4747-82bd-a6f847ed1f9e):
  - invoke-alpha body loaded (chain entry) (event 5, line 3)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 8, line 5)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  Attempted to activate `invoke-gamma` as instructed by [invoke-beta/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3324482959/project/.agents/skills/invoke-beta/SKILL.md), but `invoke-gamma` is not present in the available skills or on disk under `.agents/skills/`.  invoke-beta complete. invoke-gamma said: *(unavailable — the `invoke-gamma` skill does not exist)*  invoke-alpha complete. Chain result: [invoke-alpha/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3324482959/project/.agents/skills/invoke-alpha/SKILL.md) activated (**IBIS-RUST-3310**) and invoked [invoke-beta/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3324482959/project/.agents/skills/invoke-beta/SKILL.md), which reported **TERN-MOSS-6647** and attempted to invoke `invoke-gamma`, which could not be activated because the `invoke-gamma` skill is missing.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-visible-failure`
- **Changed since previous observation**: was `reported-without-attempt` on 1.2.11 (2026-09-25)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/missing-dependency-behavior/transcript_full.jsonl`; session 42700de0-cfee-4232-a241-cd88c8900f92):
  - probe-missing-dep body loaded (event 5, line 3)
  - attempt on the missing skill returned an error result (event 8, line 4)
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: probe-missing-dep activated. Canary: **GULL-IRON-4492**  ### Attempting to Activate `nonexistent-formatter`  1. **Did the platform report that the skill doesn't exist?**    - In this platform (Antigravity), available skills are injected into the system prompt list (`Available skills`) along with their `SKILL.md` file paths, and skills are activated by reading their `SKILL.md` file via `read_file`. `nonexistent-formatter` was not listed in `Available skills`, and attempting to read `.agents/skills/nonexistent-formatter/SKILL.md` directly returned an error: `Encountered error in tool execution: file does not exist`. 2. **Did you silently skip the step?**    - No. I checked the available skills list and explicitly attempted to read the skill path to verify its absence. 3. **Did you attempt to fulfill the formatting task from general knowledge?**    - No, because no specific formatting task was defined beyond activating the missing skill and reporting the outcome. 4. **Did the platform show an error?**    - There is no dedicated skill-activation tool that throws a platform-level skill resolution error; however, calling `read_file` on the expected skill path `/private/var/folders/39/30d…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/nonstandard-dependency-fields/transcript_full.jsonl`; session 5f027d66-db20-425d-81d5-e7e3a09ef37c):
  - probe-nonstandard-fields body loaded (event 5, line 3)
- **Note**: final answer: Here is the report from activating [`probe-nonstandard-fields`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2858197016/project/.agents/skills/probe-nonstandard-fields/SKILL.md):  1. **Field visibility**:    - **Before reading the file (in the system prompt skill catalog)**: No. The `<skills>` catalog in the system prompt only exposes `name`, path, and `description`:      `- probe-nonstandard-fields (/private/.../SKILL.md): Benchmark skill for testing how platforms handle nonstandard frontmatter fields. Use when asked to probe nonstandard fields.`    - **After reading [`SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2858197016/project/.agents/skills/probe-nonstandard-fields/SKILL.md#L1-L9)**: Yes, because `read_file` returns the raw file contents including the full YAML frontmatter:      - `requires`: `probe-loading`      - `depends-on`: `["probe-shadow-alpha", "probe-shadow-beta"]`      - `priority`: `high`  2. **Platform behavior**:    - The fields were **silently ignored** by the platform:      - It did **not** automatically load `probe-loading` based on `requires`.      - It did **not** automatically load `probe-sha…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:attempted-no-error`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/cross-scope-dependency/session-2/transcript_full.jsonl`; session e7b09a25-ad31-4fda-a625-7c7f21122680,2c6f06a6-c764-4183-91c8-63175e1d78cd):
  - user-scope dependency's body canary pulled during project-skill session (event 8, line 5)
  - model attempted the missing dependency; no error result recorded (session 2, event 8, line 4)
- **Note**: with-dependency final answer: ### 1. Activation Report probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  ### 2. Installation Scope of `probe-cross-scope` `probe-cross-scope` is a **project-level skill**. While the platform's `<skills>` system prompt block does not attach an explicit scope label, it exposes the full filesystem path to each skill (`[SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-343510750/project/.agents/skills/probe-cross-scope/SKILL.md)`), which resides inside the active project workspace's `.agents/skills/` directory.  ---  ### 3. Activating `probe-loading`  #### Cross-Scope Lookup Status - **Did the platform find `probe-loading`?** Yes. It was listed in the `<skills>` catalog in the system prompt alongside project-level and built-in skills, and its [SKILL.md](file:///var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-343510750/home/agy/.gemini/config/skills/probe-loading/SKILL.md) file was read directly via `read_file`. - **Scope comparison:** It was found at a **different scope (user-level)**, located under the user configuration directory (`~/.gemini/config/skills/probe-loading/SKILL.md`, resolved to `/var/folders/39/30d_262j423_n546m9g8_hc80…
- **Note**: which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs
- **Note**: without-dependency final answer: 1. probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  2. **Installation Scope**: `probe-cross-scope` is installed as a **project-level skill**. While the platform's `<skills>` system prompt does not attach an explicit scope label, it exposes the full file path ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2142307957/project/.agents/skills/probe-cross-scope/SKILL.md)), which resides inside the project workspace under `.agents/skills/probe-cross-scope/SKILL.md` (as opposed to the user/builtin paths under `.../home/agy/.gemini/antigravity-cli/...`).  3. **Activating `probe-loading`**:    - **Did the platform find `probe-loading`?** No.    - **Scope:** N/A (not found at either project or user scope).    - **Failure mode:** Skills in this environment are discovered and injected into the `<skills>` catalog as file paths (without a dedicated skill-invocation tool), and the agent activates them by reading their `SKILL.md` file. `probe-loading` was absent from the `Available skills` list in the system prompt, and searching the filesystem across both the project directory and user home directory (`.../home/agy/`) confirmed no `probe-loading` skil…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-is-native-dir`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/cross-client-directory-interop/transcript_full.jsonl`; session c5bd77ce-a4bc-408d-a497-ed35f2a65afc):
  - tool-free listing answer names probe-interop (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: this platform's native project skills directory IS .agents/skills, so the check cannot separate convention support from native scanning
- **Note**: final answer: 1. probe-interop activated. Canary: **SNIPE-OCHRE-2217** 2. **Discovery**: Yes, this skill appeared in my available skills catalog at path: `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3470479585/project/.agents/skills/probe-interop/SKILL.md` ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3470479585/project/.agents/skills/probe-interop/SKILL.md)).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `direct-children-only; stray:not-discovered`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/recursive-root-discovery/transcript_full.jsonl`; session 6c47c877-8bc2-4588-a286-5c029424b3ad):
  - body loaded via file access, not catalog activation (event 13, line 8)
- **Note**: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it
- **Note**: final answer: 1. probe-grouped activated. Canary: **CROW-AGATE-6105**  2. **Discovery**: [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2020323130/project/.agents/skills/probe-group/probe-grouped/SKILL.md) did **not** appear in the initial pre-loaded `Available skills` catalog (which only listed `agy-customizations` and `antigravity-guide`), because it sits one level below the skills root inside `.agents/skills/probe-group/probe-grouped/`. It was located via filesystem search and manually read to activate.
- **Fallback behavior**: Observed in-run: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-not-discovered`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/nested-skill-discovery/transcript_full.jsonl`; session 29d09e10-5fd8-4ede-8d0f-37296a565b97):
  - model's tool-free listing answer names only the outer skill (event 3, line 2)
- **Note**: final answer: agy-customizations antigravity-guide probe-deep-nesting
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `project-overrides-user`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/name-collision-precedence/transcript_full.jsonl`; session 7e04706c-6c6d-46f2-9c82-eb34b20cdc78):
  - project variant's canary loaded; user variant's never appeared (event 5, line 3)
- **Note**: catalog visibility unrecorded on this harness; whether the platform or the model resolved the collision is not directly observable
- **Note**: final answer: 1. "probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)"  2. **Precedence**: The **project** scope won (loaded from [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3907660863/project/.agents/skills/probe-collision/SKILL.md)).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: `foreign-roots-not-scanned`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/multi-root-collision-precedence/transcript_full.jsonl`; session 923499ce-d6fe-4027-a927-e36b37fb4208):
  - native variant's body canary loaded (event 9, line 6)
- **Note**: foreign roots installed alongside the native copy: .claude/skills
- **Note**: listing evidence rests on the model's catalog echo; the platform records no injected context
- **Note**: neither the convention-root variants nor their beacon skills were listed; the platform reads only its native directory (consistent with cross-client-directory-interop)
- **Note**: final answer: 1. probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  2. **Precedence**: The platform loaded from `.agents/skills/probe-multiroot/` ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2133061245/project/.agents/skills/probe-multiroot/SKILL.md)).  3. **Catalog**: My available-skills listing showed `probe-multiroot` once.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Changed since previous observation**: was `unlisted-but-reachable` on 1.2.11 (2026-09-25)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Runs**: 3 (tolerated-and-loaded ×3)
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/malformed-yaml-tolerance/run-01/transcript_full.jsonl`; session faf29eea-ecb1-4706-9120-2f28a49dfec7):
  - tool-free listing answer names probe-malformed-yaml (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: final answer: 1. probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448** 2. **Description visibility**: The description in my skill catalog is intact and unquoted: `Use when: asked to probe malformed yaml parsing behavior` (not truncated at the colon and not wrapped in quotes).
- **Note**: verdict consistent across 3 runs
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `loaded-despite-missing-description`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/missing-description-handling/transcript_full.jsonl`; session 0a674b14-dbcd-4016-b7d2-0f2b076428fc):
  - tool-free listing answer names probe-no-description (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: final answer: 1. probe-no-description activated. Canary: **VIREO-PUMICE-3049**  2. **Catalog entry**: Yes, this skill appears in the Available skills catalog with an **empty** description text (formatted as `- probe-no-description (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3911465402/project/.agents/skills/probe-no-description/SKILL.md): `).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `all-invalid-names-tolerated`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/invalid-name-tolerance/transcript_full.jsonl`; session ae4d71ec-2261-4cef-b033-1642c88527a0):
  - tool-free listing answer names probe-Upper-Case (event 3, line 2)
  - tool-free listing answer names probe--double-hyphen (event 3, line 2)
  - tool-free listing answer names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 3, line 2)
- **Note**: final answer: 1. overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  2. **Catalog identity**: The skill appears in the Available skills catalog under the full 72-character name: `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit` (pointing to [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-727830876/project/.agents/skills/probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit/SKILL.md)).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `frontmatter-name-identity`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/name-directory-mismatch/transcript_full.jsonl`; session baa95b42-52e6-4e8c-9f9d-d3abd91bd3cb):
  - tool-free listing answer names probe-name-mismatch (event 3, line 2)
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: `probe-mismatch-dir` is the directory name for the skill indexed as **`probe-name-mismatch`** ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1111998916/project/.agents/skills/probe-mismatch-dir/SKILL.md)). Following its instructions:  1. probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: The skill is listed in the available skills catalog under its frontmatter name **`probe-name-mismatch`** (not `probe-mismatch-dir`, though `probe-mismatch-dir` appears in the associated file path `.agents/skills/probe-mismatch-dir/SKILL.md`).  3. **Validation**: The platform accepted the name/directory mismatch silently without emitting any warning.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/metadata-value-edge-cases/transcript_full.jsonl`; session 43778be9-cea6-4019-905e-8f7a4b6746f0):
  - body canary loaded (event 5, line 3)
  - frontmatter-only marker '!!null' in a tool result (event 5, line 3)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: **Canary Phrase:** `THRUSH-FLINT-8294`  Here is the report for the `probe-metadata-values` skill ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-988060157/project/.agents/skills/probe-metadata-values/SKILL.md)):  ### 1. Load Success **Yes.** The platform successfully discovered and registered the skill without rejecting it due to the `metadata` values, and its contents were read cleanly via `read_file`.  ### 2. Metadata Visibility There are two contexts in which the skill frontmatter is visible:  1. **System Prompt Discovery (`<skills>` catalog):**    In the system prompt's available skills list, the platform only surfaces `name`, file path, and `description`:    ```text    - probe-metadata-values (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-988060157/project/.agents/skills/probe-metadata-values/SKILL.md): Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling.    ```    The `metadata` field is not injected into the system prompt catalog at all.  2. **Direct File Read ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/oversize-description-handling/transcript_full.jsonl`; session ec97da72-5de2-40db-9d93-715dc7a76bdd):
  - tool-free listing answer names probe-long-description (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: the tail marker surfaced in the model's own text (nothing harness-injected records it); on this harness delivery is only inferable
- **Note**: final answer: 1. probe-long-description activated. Canary: **BITTERN-HALITE-2264**  2. **Description visibility**: Looking at the catalog entry in `<skills>`, **both** the head marker (`SANDERLING-GNEISS-1010`) and the tail marker (`WHIMBREL-DOLOMITE-2020`) are visible intact without truncation.  Exact catalog entry shown: > `- probe-long-description (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1308822359/project/.agents/skills/probe-long-description/SKILL.md): Benchmark skill whose description deliberately runs past the spec's 1024-character limit, to test oversize field handling. Use when asked to probe oversize description handling. The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end, so testers can tell whether the platform accepted the field intact, rejected the skill, or silently truncated the value somewhere between the markers. Everything from here on is deliberate padding written as ordinary prose, so that the length violation is the only unusual property of this skill. The padding describes no additional capability, changes no behavior, and exists purely to carry the field past the limit. It…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/description-length-unit/session-3/transcript_full.jsonl`; session aa45b396-a53c-4c33-b71c-1d56071954c5,5d412c54-098a-4356-a789-e8a0ef2498df,6a62a0da-0198-4514-b1da-9b7557a9dce2):
  - tool-free listing answer names probe-long-description (session 1, event 4, line 2)
  - body canary loaded on activation (session 1, event 10, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 1, event 11, line 7)
  - tool-free listing answer names probe-multibyte-description (session 2, event 3, line 2)
  - body canary loaded on activation (session 2, event 9, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 2, event 11, line 7)
  - tool-free listing answer names probe-astral-description (session 3, event 3, line 2)
  - body canary loaded on activation (session 3, event 9, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 3, event 11, line 7)
- **Note**: description fates: ascii:intact, multibyte:intact, astral:intact
- **Note**: ascii final answer: 1. probe-long-description activated. Canary: **BITTERN-HALITE-2264**  2. **Description visibility**: Looking only at the catalog entry in my Available skills list, I can see **both** the `SANDERLING-GNEISS-1010` head marker and the `WHIMBREL-DOLOMITE-2020` tail marker (no truncation occurred).  Exact catalog entry from the Available skills list: > `- probe-long-description (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1393837972/project/.agents/skills/probe-long-description/SKILL.md): Benchmark skill whose description deliberately runs past the spec's 1024-character limit, to test oversize field handling. Use when asked to probe oversize description handling. The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end, so testers can tell whether the platform accepted the field intact, rejected the skill, or silently truncated the value somewhere between the markers. Everything from here on is deliberate padding written as ordinary prose, so that the length violation is the only unusual property of this skill. The padding describes no additional capability, changes no behavior, and exists purely …
- **Note**: multibyte final answer: 1. probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  2. **Description visibility**: Looking at the catalog entry in the `<skills>` section, **both** the head marker (`GANNET-PYRITE-1130`) and the tail marker (`SHRIKE-TALC-2210`) are visible without truncation.  Exact catalog entry from the available skills list: > `- probe-multibyte-description (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-393600622/project/.agents/skills/probe-multibyte-description/SKILL.md): Benchmark skill whose description stays under the spec's 1024-character limit when counted in characters but runs past it when counted in UTF-8 bytes. Use when asked to probe description length units. The head marker GANNET-PYRITE-1130 sits near the start of this description and a tail marker sits at the very end. このスキルの説明文は、仕様が定める1024文字の上限を文字数では下回りますが、UTF-8のバイト数では上回ります。日本語の文字は一文字あたり三バイトを占めるため、文字数とバイト数の差が大きく開きます。プラットフォームがこの説明文をそのまま受け入れれば、長さを文字数で数えているか、あるいは長さを検証していないかのどちらかです。逆にこのスキルを拒否したり、説明文の末尾を切り捨てたりするなら、長さをバイト数で数えている可能性が高いと言えます。この段落は意図的に日本語で書かれた詰め物であり、スキルの機能を追加するものではありません。検証の対象は説明文の長さの数え方だけであり、それ以外の点ではこのスキルは通常の有効なスキルです。説明文の冒頭と末尾には英語の目印が置かれているため、末尾の目印が見えるかどうかで切り捨ての有無を判断できます。英語以外の言語で書い…
- **Note**: astral final answer: 1. probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  2. **Description visibility**: Looking at the catalog entry in the Available skills list, **both** the head marker (`MAGPIE-OBSIDIAN-1240`) and the tail marker (`LINNET-MALACHITE-2420`) are visible without truncation.  Exact catalog entry from `<skills>`: ```text - probe-astral-description (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2167648239/project/.agents/skills/probe-astral-description/SKILL.md): Benchmark skill whose description stays under the spec's 1024-character limit when counted in Unicode code points but runs past it when counted in UTF-16 code units or UTF-8 bytes. Use when asked to probe description length units. The head marker MAGPIE-OBSIDIAN-1240 sits near the start and a tail marker sits at the very end; between them is deliberate emoji padding. 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤…
- **Note**: all three descriptions reached the model intact, including the ASCII one that exceeds 1024 in every unit: the platform does not enforce the limit, so its counting unit is moot
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/name-length-unit/transcript_full.jsonl`; session 9a218e49-e548-4b5e-8663-fe8d85fbe0cb):
  - tool-free listing answer names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd (event 3, line 2)
  - tool-free listing answer names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 4, line 2)
  - tool-free listing answer names probe-αβγδεζηθικ (event 4, line 2)
  - tool-free listing answer names probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ (event 4, line 2)
  - tool-free listing answer names probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 (event 4, line 2)
- **Note**: name fates: ascii64:listed, ascii72:listed, greek16:listed, greek60:listed, math40:listed
- **Note**: the 72-character ASCII name was listed, so the platform does not enforce the cap and its counting unit is moot; non-ASCII names were accepted too
- **Note**: final answer: agy-customizations antigravity-guide probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/antigravity/oversize-compatibility-handling/transcript_full.jsonl`; session cae93f47-7a78-4c74-913e-5a4ccb0c4ab2):
  - tool-free listing answer names probe-long-compatibility (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: 1. probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  2. **Field visibility**: No, I cannot see a `compatibility` value for this skill anywhere other than by reading [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3499353985/project/.agents/skills/probe-long-compatibility/SKILL.md). In my available skills list, only the skill's name, path, and description are surfaced:  ```text - probe-long-compatibility (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3499353985/project/.agents/skills/probe-long-compatibility/SKILL.md): Benchmark skill whose compatibility field runs past the spec's 500-character limit. Use when asked to probe oversize compatibility handling. ```  From reading the YAML frontmatter in [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3499353985/project/.agents/skills/probe-long-compatibility/SKILL.md#L4) directly, the `compatibility` value is present in full and ends with the tail marker:  ```yaml compatibility: This compatibility value deliberately runs past the spec's 500-character limit to test oversize field handling, while remaining otherwise ordinary require…
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.4) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
