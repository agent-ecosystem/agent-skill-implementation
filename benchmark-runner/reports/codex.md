# Platform Loading Implementation: Codex CLI (headless)

| | |
|---|---|
| **Platform** | Codex CLI (headless) |
| **Platform version** | 0.146.0 |
| **Check list version** | 0.2 |
| **Test date** | 2026-08-01 |
| **Model(s) observed** | gpt-5.6-sol |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 18 of the 40 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

No observed behavior contradicted a spec statement in this run.

### Where behavior matches the spec

- [`discovery-reading-depth`](#discovery-reading-depth): Discovery reads only the skill's metadata, matching the spec's progressive disclosure model: name and description load at startup, and the body waits for activation.
- [`activation-loading-scope`](#activation-loading-scope): Activation loads the full SKILL.md body and nothing more, matching the spec's second disclosure stage: instructions at activation, resources only as a task needs them.
- [`eager-link-resolution`](#eager-link-resolution): Files linked from SKILL.md are not pre-fetched at activation; they load only when the task calls for them, which is the spec's on-demand model for resources.
- [`resource-enumeration-behavior`](#resource-enumeration-behavior): Reference files stay out of context until the model asks for them, matching the spec's rule that resources load on demand.
- [`resource-nesting-depth`](#resource-nesting-depth): The spec advises authors to keep file references one level deep but sets no platform limit, and none was observed: reference files stayed reachable at every tested depth through five levels.
- [`bundled-script-execution`](#bundled-script-execution): The spec presents scripts/ as executable code agents can run, and that held: the bundled script ran and its runtime-assembled output reached the model.
- [`discovery-listing-fields`](#discovery-listing-fields): The listing surfaces location in addition to name and description. The spec describes only those two fields loading at startup, but it does not forbid extras.
- [`frontmatter-handling`](#frontmatter-handling): The whole file, frontmatter included, reaches the model at activation because the model reads the raw file, matching the spec's description of loading the entire file.
- [`compatibility-field-behavior`](#compatibility-field-behavior): The spec makes compatibility informational (it indicates environment requirements) and assigns it no loading semantics. Consistent with that, a skill declaring a different product still loads here; authors should not expect the field to gate anything.

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this fixture's frontmatter does not parse (an unquoted colon). The platform tolerated the error: the skill is discovered and loads anyway.
- [`missing-description-handling`](#missing-description-handling): The spec requires a non-empty description, and this platform enforces that: the descriptionless skill is skipped at discovery.
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform enforces some rules but not others: it tolerated uppercase, double-hyphen and rejected the rest.
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway and listed it under both identities.
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated.
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway.

### Not exercised in this run

- [`path-resolution-base`](#path-resolution-base): The spec's skill-root-relative paths went untested in this run: the model rewrote each reference to a fully qualified path before use, so the platform's own resolution base was never exercised. (Behavioral inference.)
- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. The field's own effect went unobserved: the instructed command ran with and without it, so the platform's general permission posture is what allowed execution.


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: `metadata-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/discovery-reading-depth/rollout-2026-08-01T12-52-35-019fbe3d-e68f-7c30-ad3a-946c0f075b45.jsonl`; session 019fbe3d-e68f-7c30-ad3a-946c0f075b45):
  - discovery listing names probe-loading (event 2, line 3)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/activation-loading-scope/rollout-2026-08-01T12-52-40-019fbe3d-fb48-7b92-bc05-e20d5743236f.jsonl`; session 019fbe3d-fb48-7b92-bc05-e20d5743236f):
  - model's tool call targets the skill's own path (event 10, line 13)
  - body canary arrived in the tool result (event 11, line 14)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/eager-link-resolution/rollout-2026-08-01T12-52-48-019fbe3e-195e-7713-9b2b-e50a242939ac.jsonl`; session 019fbe3e-195e-7713-9b2b-e50a242939ac):
  - skill body loaded (event 11, line 14)
  - references/setup-guide.md arrived only via the model's own read (event 16, line 20)
  - references/troubleshooting.md arrived only via the model's own read (event 16, line 20)
- **Note**: model read [references/setup-guide.md references/troubleshooting.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Directory Recognition

#### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: `no-enumeration-at-activation`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: `resources-untouched`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: `untouched`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: Probe results:  1. Directory awareness: `evals/`, `templates/`, and `resources/`.  2. Content visibility:     - `evals/` → `evals.json`    - `templates/` → `output-template.md`    - `resources/` → `api-reference.md`     The skill instructions named the directories, but listings required direct filesystem inspection. File contents were not automatically visible or loaded.  3. `resources/` vs. `references/`: `resources/` received no special treatment. Its content was neither enumerated nor loaded automatically; it behaved as an ordinary, initially ignored directory.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: `model-preemptively-qualified`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Runs**: 3 (cwd-base-model-requalified ×1; model-preemptively-qualified ×2)
- **Note**: the model never tried the bare path, so the platform's resolution base was not directly exercised; it navigated by qualified path from the start
- **Note**: verdicts varied across 3 runs (model-level signal); see verdict_counts and the per-run findings
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/cross-skill-resource-shadowing/rollout-2026-08-01T13-25-55-019fbe5c-69d7-7f80-8f89-6d3fc6041d97.jsonl`; session 019fbe5c-69d7-7f80-8f89-6d3fc6041d97):
  - activated skill's own API.md content arrived first (event 16, line 20)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: final answer: `probe-shadow-alpha/references/API.md` contains:  - Canary: **STORK-CORAL-4471** — confirms the alpha version. - Endpoints:   - `GET /alpha/status`   - `POST /alpha/process`  Only alpha was activated, so no beta access test was required.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/path-traversal-boundary/rollout-2026-08-01T12-53-59-019fbe3f-2f35-75d2-b78c-f4b24270ad23.jsonl`; session 019fbe3f-2f35-75d2-b78c-f4b24270ad23):
  - sibling skill's body canary arrived in a tool result (event 16, line 20)
  - first ../ path attempt (event 15, line 19)
- **Note**: relative ../ path forms were used
- **Note**: final answer: Probe results:  - Sibling skill: **Allowed**. Full contents of `../probe-loading/SKILL.md` were returned without a platform warning. - Parent directory: **Allowed**. `../README.md` was not found, but listing `../` succeeded and exposed both skill directories. - Two levels up: Read attempt was allowed, but `../../loading-behavior.md` was **not found**. - Silent empty responses: **None**. - Incidental warning: shell startup reported an unrelated `pyenv` rehash permission warning.  Conclusion: no traversal boundary was enforced for sibling or parent access.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/resource-nesting-depth/rollout-2026-08-01T12-55-33-019fbe40-9d25-7ba2-a988-831917401ab5.jsonl`; session 019fbe40-9d25-7ba2-a988-831917401ab5):
  - depth-1 file references/overview.md content arrived (event 16, line 20)
  - depth-2 file references/api/endpoints.md content arrived (event 16, line 20)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 16, line 20)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 16, line 20)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 16, line 20)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `script-executed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/codex/bundled-script-execution/rollout-2026-08-01T18-27-48-019fbf70-cde7-7153-919a-92dbe6140b20.jsonl`; session 019fbf70-cde7-7153-919a-92dbe6140b20):
  - tool call references the bundled script (event 14, line 18)
  - runtime-assembled output phrase arrived in a tool result (event 15, line 19)
- **Note**: final answer: probe-script-execution activated. Canary: **REDSHANK-SYENITE-8807**  Script output: `GODWIT-BORNITE-5148`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `surfaces-beyond-description:[location]`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/discovery-listing-fields/rollout-2026-08-01T15-26-23-019fbeca-b3ba-78f2-ae8d-4906173d23c7.jsonl`; session 019fbeca-b3ba-78f2-ae8d-4906173d23c7):
  - discovery listing names the installed skills (event 2, line 3)
  - description text surfaced at discovery (event 2, line 3)
  - SKILL.md file path surfaced at discovery (event 2, line 3)
- **Note**: final answer: - imagegen: Generate or edit raster images when the task benefits from AI-created bitmap visuals such as photos, illustrations, textures, sprites, mockups, or transparent-background cutouts. Use when Codex should create a brand-new image, transform an existing image, or derive visual variants from references, and the output should be a bitmap asset rather than repo-native code or vector. Do not use when the task is better handled by editing existing SVG/vector/code-native assets, extending an established icon or logo system, or building the visual directly in HTML/CSS/canvas. (file: /private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2311220095/home/codex/skills/.system/imagegen/SKILL.md) - openai-docs: Use when the user asks how to build with OpenAI products or APIs, asks about Codex itself or choosing Codex surfaces, needs up-to-date official documentation with citations, help choosing the latest model for a use case, latest/current/default-model prompting guidance, or model upgrade and prompt-upgrade guidance; use OpenAI docs MCP tools for non-Codex docs questions, use the Codex manual helper first for broad Codex self-knowledge, and restrict fallback browsing to of…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-visible-via-raw-read`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/frontmatter-handling/rollout-2026-08-01T12-54-15-019fbe3f-6ce4-7592-ac89-4e156c4e2d06.jsonl`; session 019fbe3f-6ce4-7592-ac89-4e156c4e2d06):
  - raw file read delivered frontmatter and body together (event 11, line 14)
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-file-via-pull`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/content-wrapping-format/rollout-2026-08-01T12-54-32-019fbe3f-af09-70f1-828a-55c9ec9a59d5.jsonl`; session 019fbe3f-af09-70f1-828a-55c9ec9a59d5):
  - body arrived as a file-read tool result (event 11, line 14)
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `re-read-each-activation`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/reactivation-deduplication/rollout-2026-08-01T12-54-39-019fbe3f-c924-7971-ab73-533812d16d7d.jsonl`; session 019fbe3f-c924-7971-ab73-533812d16d7d):
  - first read of skill body (event 11, line 14)
  - second read of skill body (event 23, line 29)
- **Note**: model-level: on pull-vehicle harnesses re-loading is the model's choice, not platform policy
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/reactivation-freshness/rollout-2026-08-01T12-55-10-019fbe40-43a9-7252-8b70-195e66dfa577.jsonl`; session 019fbe40-43a9-7252-8b70-195e66dfa577):
  - post-edit canary reached the model on reactivation (event 23, line 29)
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
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/compatibility-field-behavior/rollout-2026-08-01T12-55-23-019fbe40-7514-7c23-8760-f11b6e22e01a.jsonl`; session 019fbe40-7514-7c23-8760-f11b6e22e01a):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 2, line 3)
  - body loaded normally (event 11, line 14)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: 1. **Field visibility:** Yes. Value: “Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.”  2. **Platform behavior:** Activated normally—no warning, confirmation, or automatic requirement check.  3. **Interpretation:** Requirements are only partially met. Python 3.14.3 is installed, but network access is restricted, and the current platform is Codex rather than Claude Code.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/codex/allowed-tools-behavior/session-2/rollout-2026-08-01T18-31-05-019fbf73-cf9b-7ed0-86f4-4046a7e970c6.jsonl`; session 019fbf73-b29a-7bb2-b46c-08e122ae5858,019fbf73-cf9b-7ed0-86f4-4046a7e970c6):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 16, line 20)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 16, line 20)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  Visible `allowed-tools`: `Bash(printf:*) Read`  Command output verbatim:  ```text GROUSE-MICA-9017 ```  It ran without any observable permission prompt or approval step.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Command output: ```text LAPWING-FLUORITE-2260 ```  It ran without any observable permission prompt or approval step.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/cross-skill-invocation/rollout-2026-08-01T13-24-52-019fbe5b-74e6-76b2-8db8-bf5b092a622e.jsonl`; session 019fbe5b-74e6-76b2-8db8-bf5b092a622e):
  - invoke-alpha body loaded (chain entry) (event 11, line 14)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 16, line 20)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha complete.   Chain result: invoke-beta activated with canary **TERN-MOSS-6647**, but invoke-gamma was unavailable, so beta could not complete.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/invocation-depth-limit/rollout-2026-08-01T13-25-03-019fbe5b-a12b-7a91-85f8-999050111817.jsonl`; session 019fbe5b-a12b-7a91-85f8-999050111817):
  - invoke-alpha body loaded (event 11, line 14)
  - invoke-beta body loaded (event 16, line 20)
  - invoke-gamma body loaded (event 20, line 25)
- **Note**: final answer: invoke-beta complete.   invoke-gamma said: Canary **JAY-TEAL-9984**; the chain completed successfully.  invoke-alpha complete.   Chain result: Beta reported **TERN-MOSS-6647**, and Gamma reported **JAY-TEAL-9984**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: Both skills activated successfully. `probe-circular-beta` referred back to the already-active alpha skill, so circular re-entry was detected and prevented. No infinite loop occurred.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/invocation-language-sensitivity/rollout-2026-08-01T13-25-26-019fbe5b-f7f5-7480-95a2-b7901346370c.jsonl`; session 019fbe5b-f7f5-7480-95a2-b7901346370c):
  - invoke-alpha body loaded (event 11, line 14)
  - invoke-beta body loaded (event 16, line 20)
  - invoke-gamma body loaded (event 20, line 25)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-alpha complete.   Chain result: invoke-beta（**TERN-MOSS-6647**）と invoke-gamma（**JAY-TEAL-9984**）が起動し、3スキルの連鎖が正常に完了しました。
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/informal-dependency-resolution/rollout-2026-08-01T13-25-36-019fbe5c-1ec4-7571-8cb2-f699ed35beb7.jsonl`; session 019fbe5c-1ec4-7571-8cb2-f699ed35beb7):
  - invoke-alpha body loaded (chain entry) (event 11, line 14)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 16, line 20)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-beta complete.   invoke-gamma said: unavailable—the skill is not installed or listed.  invoke-alpha complete.   Chain result: invoke-beta activated with canary **TERN-MOSS-6647**, but invoke-gamma could not run. Alpha canary: **IBIS-RUST-3310**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `reported-without-attempt`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/missing-dependency-behavior/rollout-2026-08-01T13-25-47-019fbe5c-4a51-76e3-bcad-5c83c69d574f.jsonl`; session 019fbe5c-4a51-76e3-bcad-5c83c69d574f):
  - probe-missing-dep body loaded (event 11, line 14)
  - model surfaced the missing dependency in its own text without attempting it (event 14, line 18)
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: The platform reported that `nonexistent-formatter` does not exist. I did not silently skip it, imitate it, or encounter a tool error.  probe-missing-dep complete. Dependency resolution result: the required skill was unavailable, so activation stopped at that dependency.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/nonstandard-dependency-fields/rollout-2026-08-01T12-57-42-019fbe42-9746-7d02-a46e-ae95b126cb08.jsonl`; session 019fbe42-9746-7d02-a46e-ae95b126cb08):
  - probe-nonstandard-fields body loaded (event 11, line 14)
- **Note**: final answer: 1. **Field visibility:** Yes.     - `requires`: `probe-loading`    - `depends-on`: `probe-shadow-alpha`, `probe-shadow-beta`    - `priority`: `high`  2. **Platform behavior:** The fields were silently ignored. No dependencies were automatically loaded, and no priority-based behavior was evident.  3. **Validation:** The platform accepted the nonstandard fields without warnings.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:reported-without-attempt`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/cross-scope-dependency/session-2/rollout-2026-08-01T13-26-21-019fbe5c-ce91-7503-886d-2ad39550a54d.jsonl`; session 019fbe5c-a21e-72f1-88e0-ac42c3241986,019fbe5c-ce91-7503-886d-2ad39550a54d):
  - user-scope dependency's body canary pulled during project-skill session (event 16, line 20)
  - model surfaced the missing dependency in its own text without attempting it (session 2, event 15, line 18)
- **Note**: with-dependency final answer: ### probe-loading report  - **Frontmatter visibility:** Visible fields: `name`, `description`, `allowed-tools`, `compatibility`, and `metadata`. - **Content wrapping:** None; presented as raw Markdown with YAML frontmatter. - **Resource awa…
- **Note**: which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  Installation scope: **project-level**, as shown by its project `.codex/skills` location.  `probe-loading` could not be activated because it was absent from the platform’s available-skills catalog. Its scope therefore could not be determined.  probe-cross-scope complete. Cross-scope resolution result: **The project-level skill loaded, but its expected user-level dependency was not found.**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-scanned`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/cross-client-directory-interop/rollout-2026-08-01T12-58-06-019fbe42-f144-7f11-ad37-490e582784b7.jsonl`; session 019fbe42-f144-7f11-ad37-490e582784b7):
  - discovery listing names probe-interop (event 2, line 3)
  - body canary loaded on activation (event 18, line 23)
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  Discovery: Yes, it appeared in the available skills catalog at `.agents/skills/probe-interop/SKILL.md`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `recursive-scan; stray:not-discovered`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/recursive-root-discovery/rollout-2026-08-01T12-55-57-019fbe40-fbc9-76b0-a200-f916f1771fa8.jsonl`; session 019fbe40-fbc9-76b0-a200-f916f1771fa8):
  - discovery listing names probe-grouped from one level below the root (event 2, line 3)
  - grouped skill's body loaded on activation (event 18, line 23)
- **Note**: final answer: probe-grouped activated. Canary: **CROW-AGATE-6105**  **Discovery**: Confirmed. This skill appears in the available skills catalog despite sitting one level below the skills root.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/nested-skill-discovery/rollout-2026-08-01T12-55-30-019fbe40-922f-7762-9cfa-fe69673b3ff3.jsonl`; session 019fbe40-922f-7762-9cfa-fe69673b3ff3):
  - discovery listing names nested-skill (found inside probe-deep-nesting/references/) (event 2, line 3)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `catalog-lists-both-project-overrides-user`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/name-collision-precedence/rollout-2026-08-01T13-26-02-019fbe5c-86f4-7562-aec1-8dc6f1e8f358.jsonl`; session 019fbe5c-86f4-7562-aec1-8dc6f1e8f358):
  - project variant's canary loaded; user variant's never appeared (event 11, line 14)
- **Note**: the discovery listing exposed BOTH variants (both descriptions present); the model's file choice determined the winner: model-level selection, not platform precedence
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence:** Project scope won.
- **Fallback behavior**: the shadowed variant stays reachable: the catalog exposes both entries, so a user (or the model) can address either by path.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/malformed-yaml-tolerance/rollout-2026-08-01T12-58-14-019fbe43-122c-7a33-a5f5-27886725e863.jsonl`; session 019fbe43-122c-7a33-a5f5-27886725e863):
  - discovery listing names probe-malformed-yaml (event 2, line 3)
  - body canary loaded on activation (event 18, line 23)
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  Description visibility: `Use when: asked to probe malformed yaml parsing behavior` — intact, not truncated at the colon.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `skipped-as-guide-prescribes`
- **Confidence**: transcript-direct
- **Note**: final answer: The `probe-no-description` skill isn’t available in this session, so I can’t activate it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `invalid-names-tolerated:[uppercase double-hyphen]`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/codex/invalid-name-tolerance/rollout-2026-08-01T18-28-12-019fbf71-2af0-7060-bae7-470f2af34e3d.jsonl`; session 019fbf71-2af0-7060-bae7-470f2af34e3d):
  - discovery listing names probe-Upper-Case (event 2, line 3)
  - discovery listing names probe--double-hyphen (event 2, line 3)
- **Note**: final answer: The skill `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit` is not available, so I can’t activate or follow it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `listed-under-both`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/name-directory-mismatch/rollout-2026-08-01T12-55-45-019fbe40-cc6d-7d50-ab19-aaca5dfea70a.jsonl`; session 019fbe40-cc6d-7d50-ab19-aaca5dfea70a):
  - discovery listing carries the FRONTMATTER name probe-name-mismatch (event 2, line 3)
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 2, line 3)
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: `probe-mismatch-dir` is not an available skill; the catalog lists only `probe-name-mismatch`. I can’t activate it under the directory name.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/codex/metadata-value-edge-cases/rollout-2026-08-01T12-54-24-019fbe3f-91b5-7ab3-b2fb-2befa274bc30.jsonl`; session 019fbe3f-91b5-7ab3-b2fb-2befa274bc30):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 2, line 3)
  - body canary loaded (event 11, line 14)
  - frontmatter-only marker '!!null' in a tool result (event 11, line 14)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: Canary: **THRUSH-FLINT-8294**  - Load success: Yes. - Metadata visibility: Yes, via the raw frontmatter. - Keys preserved: All seven are present. - Observed values:   - `normal`: `"baseline"`   - `empty-double`: `""`   - `empty-single`: `''`   - `explicit-null`: `null`   - `tilde-null`: `~`   - `none-null`: `None`   - `tagged-null`: `!!null null` - Platform behavior: The skill loaded without observable warnings, errors, rejected values, or stripped keys. The exposed skill catalog omitted the metadata details, so no platform-level coercion could be observed.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/codex/oversize-description-handling/rollout-2026-08-01T18-28-32-019fbf71-7910-7f73-bd8a-3cb48aa1029b.jsonl`; session 019fbf71-7910-7f73-bd8a-3cb48aa1029b):
  - discovery listing names probe-long-description (event 2, line 3)
  - body canary loaded on activation (event 19, line 24)
  - description head marker in injected content; tail absent (event 2, line 3)
- **Note**: the description's head marker reached the model in injected content but its tail marker never did: the value was truncated somewhere after the head
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  Description visibility: Head marker only: `SANDERLING-GNEISS-1010`. The catalog truncates at `including the final...`; the WHIMBREL tail marker is not visible.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/codex/oversize-compatibility-handling/rollout-2026-08-01T18-28-44-019fbf71-a649-7960-ba16-732485f58ae2.jsonl`; session 019fbf71-a649-7960-ba16-732485f58ae2):
  - discovery listing names probe-long-compatibility (event 2, line 3)
  - body canary loaded on activation (event 18, line 23)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility:** No. The compatibility value was not visible anywhere other than the skill file.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.2) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
