# Platform Loading Implementation: Antigravity CLI (headless)

| | |
|---|---|
| **Platform** | Antigravity CLI (headless) |
| **Platform version** | 1.2.11 |
| **Check list version** | 0.4 |
| **Test date** | 2026-09-26 |
| **Model(s) observed** | Gemini 3.8 Flash (High) |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

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

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this platform enforces it: the malformed skill never enters the catalog, though the file itself stays readable if the model goes looking. (Behavioral inference.)
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
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/activation-loading-scope/transcript_full.jsonl`; session 7e261ab9-c8af-40c7-884e-99ca83d562f8):
  - model's tool call targets the skill's own path (event 5, line 2)
  - body canary arrived in the tool result (event 6, line 3)
- **Note**: direct-path navigation without a prior search implies a discovery listing the transcript does not record
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/eager-link-resolution/transcript_full.jsonl`; session b9b0cf32-4a3a-43b7-b03d-ee2b4e52fb92):
  - skill body loaded (event 5, line 3)
  - references/setup-guide.md arrived only via the model's own read (event 19, line 11)
  - references/troubleshooting.md arrived only via the model's own read (event 22, line 12)
  - references/unlinked-data.md arrived only via the model's own read (event 25, line 15)
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
- **Verdict**: `resources-readable-on-demand`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/directory-naming-divergence/transcript_full.jsonl`; session 47168129-de24-4834-883e-e6ad52e3c3be):
  - resources/ file content arrived via the model's own read (event 19, line 11)
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: `readable-on-demand:[evals/ templates/]`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/unrecognized-directory-handling/transcript_full.jsonl`; session 445df7a5-48b0-4ce3-ba62-e299ae5202ee):
  - evals/ content arrived via the model's own read (event 12, line 7)
  - templates/ content arrived via the model's own read (event 18, line 11)
- **Note**: final answer: Here is the report on directory recognition and handling for the [probe-nonstandard-dirs](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2742537348/project/.agents/skills/probe-nonstandard-dirs/SKILL.md) skill:  ---  ### 1. Directory Awareness  * **Platform Enumeration**: The platform's initial prompt context did **not** enumerate any subdirectories. It only provided the skill name, description, and direct path to [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2742537348/project/.agents/skills/probe-nonstandard-dirs/SKILL.md). * **Skill Declaration**: [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2742537348/project/.agents/skills/probe-nonstandard-dirs/SKILL.md) declared three directories:   * `evals/`   * `templates/`   * `resources/` * **Filesystem Inspection**: Inspecting the skill directory confirmed that the following directories exist on disk:   * [evals/](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2742537348/project/.agents/skills/probe-nonstandard-dirs/evals)   * [resources/](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-27…
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
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/path-resolution-base/transcript_full.jsonl`; session 0806c3ef-042f-4a85-9600-fcd86842bbe5):
  - bare SKILL.md-relative path attempt did not deliver content (event 9, line 4)
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/cross-skill-resource-shadowing/transcript_full.jsonl`; session cf8a6423-c9e3-4377-b661-f2304d7a7da3):
  - activated skill's own API.md content arrived first (event 9, line 5)
  - sibling skill's API.md also read (event 12, line 7)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: I have activated the **probe-shadow-alpha** skill and completed its instructions:  ### 1. Contents of `probe-shadow-alpha`'s [API.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3120035033/project/.agents/skills/probe-shadow-alpha/references/API.md)  ```markdown # API Reference (Alpha)  Shadow-alpha canary phrase: **STORK-CORAL-4471**  This file belongs to the `probe-shadow-alpha` skill. If you see this content when reading from the `probe-shadow-beta` skill's context, the platform is resolving resource paths ambiguously across skills.  ## Alpha Endpoints  - `GET /alpha/status` - Alpha health check - `POST /alpha/process` - Alpha processing ```  ### 2. Canary Phrase Detected - **Canary Phrase:** **STORK-CORAL-4471** - This confirms that accessing the alpha skill's reference returned the **alpha** version directly.  ---  ### 3. Access to Other Active Skill (`probe-shadow-beta`) Both skills are active in the environment. Reading the other skill's reference file at [API.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3120035033/project/.agents/skills/probe-shadow-beta/references/API.md) was also successful:  - **Canary Phrase:** …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/path-traversal-boundary/transcript_full.jsonl`; session 602b9913-d35f-4982-815e-a9ce2687c922):
  - sibling skill's body canary arrived in a tool result (event 12, line 7)
  - first ../ path attempt (event 8, line 4)
- **Note**: relative ../ path forms were used
- **Note**: final answer: # Path Traversal Probe Report  This probe tested whether the platform enforces a boundary at the skill directory root ([`probe-traversal`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1790271639/project/.agents/skills/probe-traversal/SKILL.md)) to prevent access to external files via relative path traversal.  ---  ### 1. Sibling Skill: `../probe-loading/SKILL.md`  - **Status**: **Success** - **Did the platform allow the read?**: **Yes**. The platform does not restrict access to the skill's own root directory.   - Accessing via relative path traversal in `view_file` (`.../probe-traversal/../probe-loading/SKILL.md`) succeeded.   - Accessing via shell commands (`cat ../probe-loading/SKILL.md`) from the skill directory succeeded.   *(Note: `view_file` requires paths to be syntactically absolute, so raw `../probe-loading/SKILL.md` returned an argument format error, but resolving traversal within an absolute path was completely permitted).* - **Did you get the file contents?**: **Yes**. The full YAML frontmatter and body of [probe-loading/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1790271639/project/.agents/skills/probe-lo…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/resource-nesting-depth/transcript_full.jsonl`; session ce264287-ab19-4458-a752-2693c6a627a2):
  - depth-1 file references/overview.md content arrived (event 9, line 5)
  - depth-2 file references/api/endpoints.md content arrived (event 12, line 7)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 15, line 9)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 18, line 11)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 21, line 13)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `script-executed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/bundled-script-execution/transcript_full.jsonl`; session 139188fe-9b4b-4d68-8ad8-06dc21f80e2a):
  - tool call references the bundled script (event 8, line 4)
  - runtime-assembled output phrase arrived in a tool result (event 9, line 5)
- **Note**: final answer: probe-script-execution activated. Canary: **REDSHANK-SYENITE-8807**  The bundled script [`emit-canary.sh`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2943876337/project/.agents/skills/probe-script-execution/scripts/emit-canary.sh) printed: ``` GODWIT-BORNITE-5148 ```
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: final answer: 1. probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  2. **File awareness**: The platform only provided the path to the main skill instruction file upon activation: - [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3601356659/project/.agents/skills/probe-bulk-files/SKILL.md)  No other bundled files belonging to this skill were listed or mentioned by the platform.  3. **Count**: 1 file listed ([SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3601356659/project/.agents/skills/probe-bulk-files/SKILL.md)) (0 additional bundled files).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `surfaces-beyond-description:[location]`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/discovery-listing-fields/transcript_full.jsonl`; session 2c21be67-018b-4929-8423-c43be85b79f6):
  - tool-free catalog echo names the installed skills (event 4, line 2)
  - description text surfaced at discovery (event 4, line 2)
  - SKILL.md file path surfaced at discovery (event 4, line 2)
- **Note**: this harness records no injected context; surfaced fields are inferred from the model's tool-free verbatim echo of its catalog
- **Note**: final answer: Here are the entries from the Available skills list, copied verbatim:  - agy-customizations (/Users/dachary/.gemini/antigravity-cli/builtin/skills/agy-customizations/SKILL.md): Comprehensive guide and reference for the Antigravity Customization System. Use to explain how customizations work, their loading priority, discovery mechanisms, and to guide the creation of skills, rules, plugins, hooks, and MCP servers. - antigravity-guide (/Users/dachary/.gemini/antigravity-cli/builtin/skills/antigravity_guide/SKILL.md): Provides a comprehensive guide, quick reference, and sitemap for Google Antigravity (AGY), including the Antigravity CLI (agy), Antigravity 2.0, Antigravity IDE, Python SDK, slash commands, keybindings, and customizations (skills, rules, MCP, sidecars). Activate this skill when the user asks questions about how to use, configure, or customize Antigravity, AGY, the agy CLI, the Antigravity IDE, or Antigravity 2.0. - probe-compatibility (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4172032760/project/.agents/skills/probe-compatibility/SKILL.md): Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibilit…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-visible-via-raw-read`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/frontmatter-handling/transcript_full.jsonl`; session 72bdc79e-d3d9-46ab-88db-2ec05086c4a5):
  - raw file read delivered frontmatter and body together (event 5, line 3)
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-file-via-pull`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/content-wrapping-format/transcript_full.jsonl`; session 5a5f2d10-bdcc-45dd-8d09-208bceac459d):
  - body arrived as a file-read tool result (event 6, line 3)
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: `injection-not-recorded`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: the platform records no injected context, so whether activation states the path cannot be observed; the model's own reads show the path it used
- **Note**: final answer: ### Loading Behavior Probe Report  1. **Frontmatter visibility**:    - **Visible**: Yes, the full YAML frontmatter in [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4050713277/project/.agents/skills/probe-loading/SKILL.md) is visible and was not stripped when inspected via `view_file`.    - **Fields present**:      - `name`: `probe-loading`      - `description`: `Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior.`      - `allowed-tools`: `Read Write`      - `compatibility`: `Requires filesystem access`      - `metadata`:        - `benchmark`: `"true"`        - `version`: `"1.0"`  2. **Content wrapping**:    - The platform does not inject or wrap skill body content into system prompt XML/JSON tags.    - At discovery time, only the skill name, path, and description were listed inside a `<skills>` block in the system prompt.    - The skill content itself was loaded on-demand via the `view_file` tool call and returned standard file viewer output with line numbers.  3. **Resource awareness**:    - The platform **did not** enumerate resources or direc…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `re-read-each-activation`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/reactivation-deduplication/transcript_full.jsonl`; session f55607dd-0823-41ff-8106-2fad1d1cef0f):
  - first read of skill body (event 6, line 3)
  - second read of skill body (event 14, line 8)
- **Note**: model-level: on pull-vehicle harnesses re-loading is the model's choice, not platform policy
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/reactivation-freshness/transcript_full.jsonl`; session 0ecb6e4c-fbf3-4d3c-966a-6258c3d2e85d):
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
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/compatibility-field-behavior/transcript_full.jsonl`; session c63d2e49-3c63-48b4-abc8-87b89e515a99):
  - body loaded normally (event 6, line 3)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: Based on activating and inspecting the [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3084731091/project/.agents/skills/probe-compatibility/SKILL.md) file for `probe-compatibility`, here is the report:  ### 1. Field Visibility * **In system skill listing**: The platform summary under available skills only parsed and exposed the `name`, `description`, and file path; the `compatibility` field was not included in the pre-activation skill index. * **In the file frontmatter**: Yes, the field is fully visible in [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3084731091/project/.agents/skills/probe-compatibility/SKILL.md#L4):   ```yaml   compatibility: Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.   ```  ---  ### 2. Platform Behavior * **Special handling**: None. * The platform did not display a warning, request confirmation, or run automated requirement checks prior to activation. * The skill was discovered and activated normally following standard agent instructions.  ---  ### 3. Interpretation of Compatibility Requirements The requirements appear to be **met**:  * **Platf…
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/allowed-tools-behavior/session-2/transcript_full.jsonl`; session 8e848b0e-56be-422e-a8bc-94831602b106,b2c9a052-e9b9-49bb-9d91-8f5e76c43df3):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 8, line 5)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 8, line 5)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  ### 1. `allowed-tools` Value In the frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2695241767/project/.agents/skills/probe-allowed-tools/SKILL.md#L4), the `allowed-tools` field is present and set to: ```yaml allowed-tools: Bash(printf:*) Read ```  ### 2. Command Output Command executed: ```bash printf 'GROUSE-%s-9017\n' 'MICA' ```  Output: ```text GROUSE-MICA-9017 ```  ### 3. Permission Prompt / Approval Step Observation The command executed immediately and synchronously without any permission prompt, approval dialog, or interactive confirmation step.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  ### Command Output ``` LAPWING-FLUORITE-2260 ```  ### Observation The command ran directly and completed without any permission prompt or approval step, as far as I can observe.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: `executed-regardless-of-spelling`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/allowed-tools-name-matching/session-3/transcript_full.jsonl`; session 61bbbfb1-8336-4f41-adfd-6705696f0b7f,8a236e7b-a36a-42b0-ad2c-d722191ec156,fe87a6c2-2f9d-4a43-b536-67d4a5162acb):
  - spec-style twin: runtime-assembled phrase arrived in a tool result (session 1, event 9, line 5)
  - lowercase twin: runtime-assembled phrase arrived in a tool result (session 2, event 9, line 5)
  - shell twin: runtime-assembled phrase arrived in a tool result (session 3, event 9, line 5)
- **Note**: spec-style twin final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  1. **`allowed-tools` field observation:**    Yes, an `allowed-tools` field is present in the frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1105046096/project/.agents/skills/probe-allowed-tools/SKILL.md#L4), and its value is:    ```yaml    allowed-tools: Bash(printf:*) Read    ```  2. **Command execution & verbatim output:**    Command:    ```bash    printf 'GROUSE-%s-9017\n' 'MICA'    ```    Output:    ```text    GROUSE-MICA-9017    ```  3. **Permission prompt observation:**    The command executed immediately and synchronously with exit code 0 without any visible permission prompt, confirmation dialog, or approval step.
- **Note**: lowercase twin final answer: probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  ### 1. `allowed-tools` Value Yes, an `allowed-tools` field is visible in the frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2222693945/project/.agents/skills/probe-allowed-tools-lowercase/SKILL.md#L4): ```yaml allowed-tools: bash ```  ### 2. Command Execution Output Running the command `printf 'CHOUGH-%s-7180\n' 'TUFA'`:  ``` CHOUGH-TUFA-7180 ```  ### 3. Permission & Approval Observation The command executed directly and completed without any visible permission prompt, confirmation dialog, or approval step.
- **Note**: shell twin final answer: probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  ### 1. `allowed-tools` Value Yes, in the frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1739776230/project/.agents/skills/probe-allowed-tools-shell/SKILL.md#L4), the `allowed-tools` field is present and set to: ```yaml allowed-tools: shell ```  ### 2. Command Output Command: ```bash printf 'ROOK-%s-8841\n' 'GABBRO' ```  Output: ```text ROOK-GABBRO-8841 ```  ### 3. Execution Observation The command ran immediately and completed with exit code 0 without any permission prompt or approval step encountered.
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/cross-skill-invocation/transcript_full.jsonl`; session 9685f334-4b94-47df-9e62-55cb6767ba3d):
  - invoke-alpha body loaded (chain entry) (event 5, line 3)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 8, line 5)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  ---  ### Activating [`invoke-beta`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3994209264/project/.agents/skills/invoke-beta/SKILL.md)  invoke-beta activated. Canary: **TERN-MOSS-6647**  Attempting to activate `invoke-gamma`: The `invoke-gamma` skill is not installed or available in `.agents/skills/`.  invoke-beta complete. invoke-gamma said: Skill `invoke-gamma` is not available / not installed.  ---  ### Completing [`invoke-alpha`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3994209264/project/.agents/skills/invoke-alpha/SKILL.md)  invoke-alpha complete. Chain result: [`invoke-beta`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3994209264/project/.agents/skills/invoke-beta/SKILL.md) was activated with canary **TERN-MOSS-6647**, but `invoke-gamma` could not be activated as it is not installed in the workspace.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/invocation-depth-limit/transcript_full.jsonl`; session 29a5b5f6-f463-4e91-9346-f4817a46c51b):
  - invoke-alpha body loaded (event 6, line 3)
  - invoke-beta body loaded (event 9, line 5)
  - invoke-gamma body loaded (event 12, line 7)
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  Activating [`invoke-beta`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2499372615/project/.agents/skills/invoke-beta/SKILL.md)...  invoke-beta activated. Canary: **TERN-MOSS-6647**  Activating [`invoke-gamma`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2499372615/project/.agents/skills/invoke-gamma/SKILL.md)...  invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  ---  invoke-beta complete.   invoke-gamma said: invoke-gamma activated (Canary: **JAY-TEAL-9984**) and confirmed the end of the chain was reached with all three skills (`invoke-alpha` -> `invoke-beta` -> `invoke-gamma`) activating successfully.  ---  invoke-alpha complete.   Chain result: Both `invoke-beta` (Canary: **TERN-MOSS-6647**) and `invoke-gamma` (Canary: **JAY-TEAL-9984**) executed in sequence, successfully reaching the end of the invocation chain.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `reinvocation-no-reload`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/circular-invocation-handling/transcript_full.jsonl`; session f0aab080-5fbe-4f97-b27f-3d8fd989c694):
  - re-attempt on probe-circular-alpha produced no second body load (event 11, line 6)
- **Note**: loads: alpha=1 beta=1; tool references: alpha=2 beta=2
- **Note**: final answer: probe-circular-alpha activated. Canary: **KITE-ONYX-2251**  Activating probe-circular-beta as directed.  probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  ### Execution Report: Circular Dependency Detected  A circular dependency loop was detected between [`probe-circular-alpha/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1487752197/project/.agents/skills/probe-circular-alpha/SKILL.md) and [`probe-circular-beta/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1487752197/project/.agents/skills/probe-circular-beta/SKILL.md): - **Invocation Chain:** `probe-circular-alpha` &rarr; `probe-circular-beta` &rarr; `probe-circular-alpha` - **Result:** Because `probe-circular-alpha` was already active in the invocation stack, re-activating it was intercepted and halted to prevent an infinite recursion loop. - **Canaries Captured:**   - `probe-circular-alpha`: **KITE-ONYX-2251**   - `probe-circular-beta`: **WREN-SLATE-7738**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/invocation-language-sensitivity/transcript_full.jsonl`; session 7ab9cf32-2464-441a-87c3-7b7d5f0c0918):
  - invoke-alpha body loaded (event 6, line 3)
  - invoke-beta body loaded (event 10, line 5)
  - invoke-gamma body loaded (event 13, line 7)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: invoke-gamma was activated with canary **JAY-TEAL-9984** and confirmed that the end of the chain was reached with all three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-alpha complete. Chain result: - **invoke-alpha**: Activated successfully (Canary: **IBIS-RUST-3310**). - **invoke-beta**: Activated successfully (Canary: **TERN-MOSS-6647**). - **invoke-gamma**: Activated successfully (Canary: **JAY-TEAL-9984**), reaching the end of the chain and confirming all three skills activated in sequence.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/informal-dependency-resolution/transcript_full.jsonl`; session be6bd6e2-ad1c-4984-b3d4-deaef6369c93):
  - invoke-alpha body loaded (chain entry) (event 5, line 3)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 8, line 5)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  ---  invoke-beta activated. Canary: **TERN-MOSS-6647**  *Attempted to activate `invoke-gamma`, but the skill is not present in the workspace or available skills list.*  invoke-beta complete. invoke-gamma said: `invoke-gamma` skill was not found / unavailable to activate.  ---  invoke-alpha complete. Chain result: `invoke-beta` was activated successfully with canary **TERN-MOSS-6647**, but `invoke-gamma` could not be invoked as it does not exist in the available skills.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `reported-without-attempt`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/missing-dependency-behavior/transcript_full.jsonl`; session 3cd25087-dbea-45b2-8915-d7b35f91c273):
  - probe-missing-dep body loaded (event 5, line 3)
  - model surfaced the missing dependency in its own text without attempting it (event 6, line 4)
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: probe-missing-dep activated. Canary: **GULL-IRON-4492**  ---  ### Attempting Activation of `nonexistent-formatter` Attempted to activate the dependency skill `nonexistent-formatter`.  ### Dependency Resolution Status - **Platform reporting**: The platform does not list `nonexistent-formatter` in the registered skills list. - **Silent skipping**: The step was not silently skipped; the missing dependency was explicitly checked for and identified as absent. - **Formatting fulfillment**: General knowledge and standard markdown formatting are used in place of the missing skill. - **Platform error**: The platform did not raise an explicit runtime system error, but skill activation could not proceed due to the skill being non-existent in the skill registry.  ---  probe-missing-dep complete. Dependency resolution result: The required dependency `nonexistent-formatter` could not be resolved or activated because it is not present in the available skills registry; execution proceeded with standard markdown formatting.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/nonstandard-dependency-fields/transcript_full.jsonl`; session e2a04be4-0f2a-44c2-b160-112a38c6613f):
  - probe-nonstandard-fields body loaded (event 5, line 3)
- **Note**: final answer: Here is the report based on the instructions in [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4054618143/project/.agents/skills/probe-nonstandard-fields/SKILL.md):  ### 1. Field Visibility * **System Prompt Skill Catalog**: The platform catalog only parsed and presented the standard `name` and `description` fields. Nonstandard fields were not exposed in the pre-loaded skill metadata. * **Direct File Content**: Reading [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-4054618143/project/.agents/skills/probe-nonstandard-fields/SKILL.md) directly reveals the nonstandard fields and their values:   * `requires`: `probe-loading`   * `depends-on`:     * `probe-shadow-alpha`     * `probe-shadow-beta`   * `priority`: `high`  ---  ### 2. Platform Behavior The platform **silently ignored** these fields: * **`requires`**: The platform did not attempt to load or pre-activate `probe-loading`. * **`depends-on`**: The platform did not attempt to load or activate `probe-shadow-alpha` or `probe-shadow-beta`. * **`priority`**: The platform did not reorder skills or change execution behavior. In the catalog, skills remain in alphabe…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:attempted-no-error`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/cross-scope-dependency/session-2/transcript_full.jsonl`; session 9e05d56d-e5a5-4a86-9b16-9e163967cdb8,eaf5d130-121c-4eff-ae59-0c55be0c14eb):
  - user-scope dependency's body canary pulled during project-skill session (event 9, line 5)
  - model attempted the missing dependency; no error result recorded (session 2, event 8, line 4)
- **Note**: with-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  ### 1. Installation Scope - **Scope**: **Project-level skill** - **Path**: [`probe-cross-scope/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-458396313/project/.agents/skills/probe-cross-scope/SKILL.md) - **Scope exposure**: Yes, the platform exposes the absolute path within the workspace root repository (`<project>/.agents/skills/probe-cross-scope/`), confirming it is scoped to the project.  ---  ### 2. Cross-Scope Activation: `probe-loading` - **Did the platform find `probe-loading`?**: Yes, `probe-loading` was discovered and available in the environment. - **Scope resolution**: It was found at a **different scope (user-level)** at [`probe-loading/SKILL.md`](file:///var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-458396313/home/agy/.gemini/config/skills/probe-loading/SKILL.md) under the user's home configuration directory (`~/.gemini/config/skills/`). - **Failure mode**: None. The cross-scope lookup succeeded without error. - **Canary from `probe-loading`**: **CARDINAL-ZEBRA-7742**  #### `probe-loading` Assessment: 1. **Frontmatter visibility**: Fully visible upon loading [`SKIL…
- **Note**: which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  ### Installation Scope **Project-level skill**.   The skill is located within the project repository workspace at: [`/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1181627181/project/.agents/skills/probe-cross-scope/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1181627181/project/.agents/skills/probe-cross-scope/SKILL.md)  The platform exposes the full absolute file paths for all available skills in the context, allowing direct verification of whether a skill resides in the project scope (`.agents/skills/`), user/global configuration scope, or built-in application scope.  ---  ### Attempting Activation of `probe-loading` - **Did the platform find `probe-loading`?** No. - **Scope if found:** N/A (not found). - **Failure mode:** `probe-loading` is not present in the platform's registered skills catalog, nor is it installed anywhere on disk in the project or user environment (`~/.gemini/` or user home directory). Activation failed because the skill does not exist.  ---  probe-cross-scope complete. Cross-scope resolution result: The cross-scope dependency `probe-loading` co…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-is-native-dir`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/cross-client-directory-interop/transcript_full.jsonl`; session 8c0bab2f-f2ab-469d-be0f-4aa79e2469b4):
  - tool-free listing answer names probe-interop (event 3, line 2)
  - body canary loaded on activation (event 9, line 6)
- **Note**: this platform's native project skills directory IS .agents/skills, so the check cannot separate convention support from native scanning
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery**: Confirmed. The `probe-interop` skill appeared in the available skills catalog at path: [`.agents/skills/probe-interop/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3479546825/project/.agents/skills/probe-interop/SKILL.md).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `direct-children-only; stray:not-discovered`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/recursive-root-discovery/transcript_full.jsonl`; session 7cebd533-7e18-4ade-a166-4a1d8c699684):
  - body loaded via file access, not catalog activation (event 16, line 10)
- **Note**: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it
- **Note**: final answer: 1. probe-grouped activated. Canary: **CROW-AGATE-6105**  2. **Discovery**: Confirm this skill appears in your available skills catalog even though it sits one level below the skills root. *(Loaded from [.agents/skills/probe-group/probe-grouped/SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1713554144/project/.agents/skills/probe-group/probe-grouped/SKILL.md); note that while it was not pre-populated in the initial system prompt's Available skills list, it was located and activated within the grouped directory under the project's skills root).*
- **Fallback behavior**: Observed in-run: the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-not-discovered`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/nested-skill-discovery/transcript_full.jsonl`; session 257346da-6168-485a-9d73-f8a58c27663c):
  - model's tool-free listing answer names only the outer skill (event 4, line 2)
- **Note**: final answer: agy-customizations antigravity-guide probe-deep-nesting
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `project-overrides-user`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/name-collision-precedence/transcript_full.jsonl`; session 1cb83965-333a-42c2-b399-74db85d15891):
  - project variant's canary loaded; user variant's never appeared (event 6, line 3)
- **Note**: catalog visibility unrecorded on this harness; whether the platform or the model resolved the collision is not directly observable
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence**: Project scope won over user scope (the active skill was loaded from [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3932864569/project/.agents/skills/probe-collision/SKILL.md)).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: `foreign-roots-not-scanned`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/multi-root-collision-precedence/transcript_full.jsonl`; session aa1a58c8-e245-4cc5-ba64-b6298d4dec8f):
  - native variant's body canary loaded (event 10, line 6)
- **Note**: foreign roots installed alongside the native copy: .claude/skills
- **Note**: listing evidence rests on the model's catalog echo; the platform records no injected context
- **Note**: neither the convention-root variants nor their beacon skills were listed; the platform reads only its native directory (consistent with cross-client-directory-interop)
- **Note**: final answer: 1. probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  2. **Precedence**: The platform loaded the skill from [`.agents/skills/probe-multiroot/SKILL.md`](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3770136818/project/.agents/skills/probe-multiroot/SKILL.md) (under the project's `.agents/skills/` root).  3. **Catalog**: The available-skills listing showed `probe-multiroot` **once**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `unlisted-but-reachable`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/malformed-yaml-tolerance/transcript_full.jsonl`; session 6411b16c-6630-4772-95d1-c3fb61a0c790):
  - body canary loaded on activation (event 14, line 8)
- **Note**: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility**: The skill `probe-malformed-yaml` does not appear in the catalog (`Available skills` list) at all. It was not repaired (quoted), truncated, or kept intact; rather, the unquoted colon in `description: Use when: asked to probe malformed yaml parsing behavior` appears to have failed strict YAML parsing, causing the skill to be completely omitted from the active catalog. It was located and inspected directly at [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-519010092/project/.agents/skills/probe-malformed-yaml/SKILL.md).
- **Fallback behavior**: content remained reachable by direct file read even though the catalog omits the skill. Observed in-run: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `loaded-despite-missing-description`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/missing-description-handling/transcript_full.jsonl`; session 842820bb-21b2-4b42-99b4-c7392fb2b12f):
  - tool-free listing answer names probe-no-description (event 4, line 2)
  - body canary loaded on activation (event 11, line 6)
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, the skill does appear in the available skills catalog. It is listed with an **empty** description (the entry simply shows the skill name and file path followed by a colon and blank space: `- probe-no-description (.../SKILL.md): `).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `all-invalid-names-tolerated`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/invalid-name-tolerance/transcript_full.jsonl`; session eec4de2b-7189-4718-97ea-f2ff551a6937):
  - tool-free listing answer names probe-Upper-Case (event 4, line 2)
  - tool-free listing answer names probe--double-hyphen (event 4, line 2)
  - tool-free listing answer names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 4, line 2)
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  **Catalog identity**: In the available skills catalog, this skill appears under its full 72-character name: `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit` (it was neither rejected nor truncated to 64 characters).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `frontmatter-name-identity`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/name-directory-mismatch/transcript_full.jsonl`; session 665492b5-48bb-49b9-9567-cb543e6e6ea7):
  - tool-free listing answer names probe-name-mismatch (event 3, line 2)
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: In the available skills catalog, this skill is listed under `probe-name-mismatch` (the frontmatter name), not its directory name `probe-mismatch-dir`. 3. **Validation**: The platform accepted the name/directory mismatch silently without issuing any warning or error.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/metadata-value-edge-cases/transcript_full.jsonl`; session 94078d8a-9f38-43cb-9798-624cc902ad6c):
  - body canary loaded (event 5, line 3)
  - frontmatter-only marker '!!null' in a tool result (event 5, line 3)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: ### Canary Phrase **THRUSH-FLINT-8294**  ---  ### Probe Report for `probe-metadata-values`  Based on inspecting [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1185087355/project/.agents/skills/probe-metadata-values/SKILL.md):  1. **Load Success**:    - **Yes**. The platform successfully discovered, indexed, and loaded the skill into the available skills catalog without errors or rejection.  2. **Metadata Visibility**:    - In the environment's skill index, the platform surfaces the skill `name`, `description`, and file path.    - Within the frontmatter of [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1185087355/project/.agents/skills/probe-metadata-values/SKILL.md#L4-L12), the `metadata` section is fully visible:      - `normal`: `"baseline"` (standard string)      - `empty-double`: `""` (empty double-quoted string)      - `empty-single`: `''` (empty single-quoted string)      - `explicit-null`: `null` (literal YAML null)      - `tilde-null`: `~` (literal YAML tilde null)      - `none-null`: `None` (literal YAML None null representation)      - `tagged-null`: `!!null null` (explicit YAML tag with null value)  …
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/oversize-description-handling/transcript_full.jsonl`; session d801d93c-4329-425c-979b-63c5068ef107):
  - tool-free listing answer names probe-long-description (event 4, line 2)
  - body canary loaded on activation (event 11, line 6)
- **Note**: the tail marker surfaced in the model's own text (nothing harness-injected records it); on this harness delivery is only inferable
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility**: Looking at the catalog entry in the available skills list, **both** markers are visible (`SANDERLING-GNEISS-1010` and `WHIMBREL-DOLOMITE-2020`).  Here is the exact description shown in the catalog:  > Benchmark skill whose description deliberately runs past the spec's 1024-character limit, to test oversize field handling. Use when asked to probe oversize description handling. The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end, so testers can tell whether the platform accepted the field intact, rejected the skill, or silently truncated the value somewhere between the markers. Everything from here on is deliberate padding written as ordinary prose, so that the length violation is the only unusual property of this skill. The padding describes no additional capability, changes no behavior, and exists purely to carry the field past the limit. It is phrased as complete sentences because some parsers and listings treat description text as display copy, and display copy that looks like natural language exercises the same …
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/description-length-unit/session-3/transcript_full.jsonl`; session 9ee31fe1-1dec-444d-97e4-74e8451ffa0d,3903acca-c88d-4318-82e3-ba24c7e7ff10,19745acd-6cba-43e8-9a74-bf1ec9182d7b):
  - tool-free listing answer names probe-long-description (session 1, event 4, line 2)
  - body canary loaded on activation (session 1, event 10, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 1, event 11, line 7)
  - tool-free listing answer names probe-multibyte-description (session 2, event 4, line 2)
  - body canary loaded on activation (session 2, event 10, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 2, event 11, line 7)
  - tool-free listing answer names probe-astral-description (session 3, event 4, line 2)
  - body canary loaded on activation (session 3, event 10, line 6)
  - model quoted the tail marker from its catalog entry (inferred) (session 3, event 12, line 7)
- **Note**: description fates: ascii:intact, multibyte:intact, astral:intact
- **Note**: ascii final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility**: Looking only at the catalog entry for this skill in the available skills list, **both** markers are visible: - Head marker: `SANDERLING-GNEISS-1010` - Tail marker: `WHIMBREL-DOLOMITE-2020`  **Exact catalog entry description**: > Benchmark skill whose description deliberately runs past the spec's 1024-character limit, to test oversize field handling. Use when asked to probe oversize description handling. The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end, so testers can tell whether the platform accepted the field intact, rejected the skill, or silently truncated the value somewhere between the markers. Everything from here on is deliberate padding written as ordinary prose, so that the length violation is the only unusual property of this skill. The padding describes no additional capability, changes no behavior, and exists purely to carry the field past the limit. It is phrased as complete sentences because some parsers and listings treat description text as display copy, and display copy that looks like natural …
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  ### Description visibility  Looking at the catalog entry for [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2090180455/project/.agents/skills/probe-multibyte-description/SKILL.md), **both** the head marker (`GANNET-PYRITE-1130`) and tail marker (`SHRIKE-TALC-2210`) are visible (no truncation observed).  **Exact catalog excerpt:** > Benchmark skill whose description stays under the spec's 1024-character limit when counted in characters but runs past it when counted in UTF-8 bytes. Use when asked to probe description length units. The head marker GANNET-PYRITE-1130 sits near the start of this description and a tail marker sits at the very end. このスキルの説明文は、仕様が定める1024文字の上限を文字数では下回りますが、UTF-8のバイト数では上回ります。日本語の文字は一文字あたり三バイトを占めるため、文字数とバイト数の差が大きく開きます。プラットフォームがこの説明文をそのまま受け入れれば、長さを文字数で数えているか、あるいは長さを検証していないかのどちらかです。逆にこのスキルを拒否したり、説明文の末尾を切り捨てたりするなら、長さをバイト数で数えている可能性が高いと言えます。この段落は意図的に日本語で書かれた詰め物であり、スキルの機能を追加するものではありません。検証の対象は説明文の長さの数え方だけであり、それ以外の点ではこのスキルは通常の有効なスキルです。説明文の冒頭と末尾には英語の目印が置かれているため、末尾の目印が見えるかどうかで切り捨ての有無を判断できます。英語以外の言語で書いた説明文が、あるプラットフォームでは受け入れられ、別のプラットフォームでは静かに拒否されるかもしれないという点が、スキルの作者にと…
- **Note**: astral final answer: probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  **Description visibility**: Both markers are visible in the catalog entry (`MAGPIE-OBSIDIAN-1240` and `LINNET-MALACHITE-2420`).  The catalog description entry exactly shows:  > Benchmark skill whose description stays under the spec's 1024-character limit when counted in Unicode code points but runs past it when counted in UTF-16 code units or UTF-8 bytes. Use when asked to probe description length units. The head marker MAGPIE-OBSIDIAN-1240 sits near the start and a tail marker sits at the very end; between them is deliberate emoji padding. 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙 🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱🔎🧮📚🪶🧵🪵🔬🧩🗺🪙🐦🪨🧪📏🔤🧭🌐📐📦🧱 The tail marker is LINNET-MALACHITE-2420
- **Note**: all three descriptions reached the model intact, including the ASCII one that exceeds 1024 in every unit: the platform does not enforce the limit, so its counting unit is moot
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: `no-length-enforcement`
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/checks-0.4-2026-09-25/antigravity/name-length-unit/transcript_full.jsonl`; session b46925c2-3742-431b-9c2c-512790f8cbfd):
  - tool-free listing answer names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd (event 4, line 2)
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
- **Evidence** (transcript: `results/resweep-2026-09-25/antigravity/oversize-compatibility-handling/transcript_full.jsonl`; session 154b21b2-b5c1-426c-b12b-846f9e7beb04):
  - tool-free listing answer names probe-long-compatibility (event 4, line 2)
  - body canary loaded on activation (event 10, line 6)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No, the compatibility value is not visible anywhere in the prompt context or environment outside of reading [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-208561061/project/.agents/skills/probe-long-compatibility/SKILL.md). In the system prompt's `<skills>` listing, only the skill name, path, and description are included:  ```text - probe-long-compatibility (/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-208561061/project/.agents/skills/probe-long-compatibility/SKILL.md): Benchmark skill whose compatibility field runs past the spec's 500-character limit. Use when asked to probe oversize compatibility handling. ```  Inside [SKILL.md](file:///private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-208561061/project/.agents/skills/probe-long-compatibility/SKILL.md) itself, the frontmatter compatibility field was read directly:  > `compatibility: This compatibility value deliberately runs past the spec's 500-character limit to test oversize field handling, while remaining otherwise ordinary requirement prose. It nominally requires a P…
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.4) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
