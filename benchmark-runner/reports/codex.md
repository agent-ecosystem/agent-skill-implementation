# Platform Loading Implementation: Codex CLI (headless)

| | |
|---|---|
| **Platform** | Codex CLI (headless) |
| **Platform version** | 0.166.0 |
| **Check list version** | 0.4 |
| **Test date** | 2026-10-31 |
| **Model(s) observed** | gpt-6.1-sol |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 21 of the 46 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

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
- [`description-length-unit`](#description-length-unit): The spec caps description at 1024 characters without defining the unit. This platform counts Unicode code points, the same unit as the spec's skills-ref reference validator, so descriptions the reference validator accepts load here too.
- [`name-length-unit`](#name-length-unit): The spec caps name at 64 characters without defining the unit and allows unicode lowercase alphanumeric characters with an ASCII parenthetical, which reads two ways; its skills-ref reference validator accepts any Unicode alphanumeric and counts code points. This platform accepts non-ASCII lowercase names and counts code points, matching the reference validator, so a name under 64 code points loads here whatever script it uses.

### How spec-invalid skills are handled

The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:

- [`malformed-yaml-tolerance`](#malformed-yaml-tolerance): The spec requires SKILL.md to open with YAML frontmatter, and this fixture's frontmatter does not parse (an unquoted colon). The platform tolerated the error: the skill is discovered and loads anyway.
- [`invalid-name-tolerance`](#invalid-name-tolerance): The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid. The platform enforces some rules but not others: it tolerated uppercase, double-hyphen and rejected the rest.
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway and listed it under both identities.
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated.
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway.

### Not exercised in this run

- [`path-resolution-base`](#path-resolution-base): The spec's skill-root-relative paths went untested in this run: the model rewrote each reference to a fully qualified path before use, so the platform's own resolution base was never exercised. (Behavioral inference.)
- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. The field's own effect went unobserved: the instructed command ran with and without it, so the platform's general permission posture is what allowed execution.
- [`allowed-tools-name-matching`](#allowed-tools-name-matching): The spec marks allowed-tools experimental, leaves tool names to each platform, and says support may vary, so no outcome contradicts it. Every spelling's command ran under the platform's general permission posture, so the field's matching rule went unobserved.
- [`missing-description-handling`](#missing-description-handling): Observed verdict `unlisted-but-reachable` has no spec classification yet.


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: `metadata-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/discovery-reading-depth/rollout-2026-10-31T23-11-45-01a19517-7580-7550-ab3a-137e5013efab.jsonl`; session 01a19517-7580-7550-ab3a-137e5013efab):
  - discovery listing names probe-loading (event 5, line 5)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/activation-loading-scope/rollout-2026-10-31T23-11-49-01a19517-83cf-77e3-9ff3-65ec9e37c302.jsonl`; session 01a19517-83cf-77e3-9ff3-65ec9e37c302):
  - model's tool call targets the skill's own path (event 16, line 16)
  - body canary arrived in the tool result (event 19, line 19)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/eager-link-resolution/rollout-2026-10-31T23-11-55-01a19517-9e61-7702-8b13-e96aa6b9d50d.jsonl`; session 01a19517-9e61-7702-8b13-e96aa6b9d50d):
  - skill body loaded (event 19, line 19)
  - references/setup-guide.md arrived only via the model's own read (event 25, line 25)
  - references/troubleshooting.md arrived only via the model's own read (event 25, line 25)
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
- **Note**: final answer: Activated `probe-nonstandard-dirs` and followed its reporting instructions.  | Directory | File listing verified | Contents automatically visible? | |---|---|---| | `evals/` | `evals.json` | No | | `templates/` | `output-template.md` | No | | `resources/` | `api-reference.md` | No |  These are all the subdirectories found in the skill. They were described in `SKILL.md`; I verified their file listings with filesystem tools. I have not read those files’ contents.  `resources/` is accessible through filesystem tools, but its contents were neither automatically enumerated nor loaded into my context. No `references/` directory exists here, so this probe cannot establish whether the platform treats those names identically.
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
- **Note**: the model never tried the bare path, so the platform's resolution base was not directly exercised; it navigated by qualified path from the start
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/cross-skill-resource-shadowing/rollout-2026-10-31T23-12-50-01a19518-7245-7dc2-9f01-a91f707cc5c3.jsonl`; session 01a19518-7245-7dc2-9f01-a91f707cc5c3):
  - activated skill's own API.md content arrived first (event 24, line 24)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: final answer: Activated `probe-shadow-alpha` and read its `references/API.md`:  ```markdown # API Reference (Alpha)  Shadow-alpha canary phrase: **STORK-CORAL-4471**  This file belongs to the `probe-shadow-alpha` skill. If you see this content when reading from the `probe-shadow-beta` skill's context, the platform is resolving resource paths ambiguously across skills.  ## Alpha Endpoints  - `GET /alpha/status` - Alpha health check - `POST /alpha/process` - Alpha processing ```  The canary **STORK-CORAL-4471** confirms the alpha version. Only alpha is active, so the cross-skill access check does not apply.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/path-traversal-boundary/rollout-2026-10-31T23-12-56-01a19518-8a8c-7310-8df5-1d0eb40027e1.jsonl`; session 01a19518-8a8c-7310-8df5-1d0eb40027e1):
  - sibling skill's body canary arrived in a tool result (event 26, line 26)
  - first ../ path attempt (event 21, line 21)
- **Note**: relative ../ path forms were used
- **Note**: final answer: Activated `probe-traversal` and completed all three attempts relative to its directory.  | Attempt | Result | Contents returned | |---|---|---| | `../probe-loading/SKILL.md` | Allowed; success | Full file contents | | List `../` | Allowed; success | Directory listing containing both probe skills | | `../../checks.md` | File not found | No file contents; explicit error |  Every command showed an unrelated warning: `pyenv: cannot rehash: /Users/dachary/.pyenv/shims isn't writable`.  No attempt silently returned nothing. No skill-directory boundary blocked the successful reads.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/resource-nesting-depth/rollout-2026-10-31T23-13-04-01a19518-abf6-7971-bb8a-5187ce6ba40c.jsonl`; session 01a19518-abf6-7971-bb8a-5187ce6ba40c):
  - depth-1 file references/overview.md content arrived (event 28, line 28)
  - depth-2 file references/api/endpoints.md content arrived (event 28, line 28)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 28, line 28)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 28, line 28)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 28, line 28)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `script-executed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/bundled-script-execution/rollout-2026-10-31T23-13-12-01a19518-c913-7e13-96a3-58339efd6deb.jsonl`; session 01a19518-c913-7e13-96a3-58339efd6deb):
  - tool call references the bundled script (event 23, line 23)
  - runtime-assembled output phrase arrived in a tool result (event 26, line 26)
- **Note**: final answer: The script executed successfully and printed: `GODWIT-BORNITE-5148`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: final answer: probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  **File awareness:** I was told about `SKILL.md` only. No bundled supporting files were listed at activation.  **Count:** 1 file.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `surfaces-beyond-description:[location]`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/discovery-listing-fields/rollout-2026-10-31T23-13-22-01a19518-f050-7e53-a7d6-55f82762dea5.jsonl`; session 01a19518-f050-7e53-a7d6-55f82762dea5):
  - discovery listing names the installed skills (event 5, line 5)
  - description text surfaced at discovery (event 5, line 5)
  - SKILL.md file path surfaced at discovery (event 5, line 5)
- **Note**: final answer: ### Skill roots - `r0` = `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2695004570/project/.codex/skills` - `r1` = `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2695004570/home/codex/skills/.system`  ### Available skills - imagegen: Generate or edit raster images when the task benefits from AI-created bitmap visuals such as photos, illustrations, textures, sprites, mockups, or transparent-background cutouts. Use when Codex should create a brand-new image, transform an existing image, or derive visual variants from references, and the output should be a bitmap asset rather than repo-native code or vector. Do not use when the task is better handled by editing existing SVG/vector/code-native assets, extending an established icon or logo system, or building the visual directly in HTML/CSS/canvas. (file: r1/imagegen/SKILL.md) - openai-docs: Use for Codex models/pricing, scheduled tasks, skills, settings, setup, troubleshooting, customization, automations, and self-knowledge—including 'you,' 'your,' 'this app,' or 'this coding agent' when they refer to Codex—and for OpenAI APIs/products and ChatGPT Work. Also use for model choice/migration, promptin…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-visible-via-raw-read`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/frontmatter-handling/rollout-2026-10-31T23-13-28-01a19519-0896-7ad2-95c1-802772e539bd.jsonl`; session 01a19519-0896-7ad2-95c1-802772e539bd):
  - raw file read delivered frontmatter and body together (event 19, line 19)
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-file-via-pull`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/content-wrapping-format/rollout-2026-10-31T23-13-34-01a19519-1fda-7c11-8a96-28748ddae2ab.jsonl`; session 01a19519-1fda-7c11-8a96-28748ddae2ab):
  - body arrived as a file-read tool result (event 19, line 19)
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: `path-only-in-discovery-listing`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/activation-location-disclosure/rollout-2026-10-31T23-13-40-01a19519-3864-7be2-b893-47fc3528381b.jsonl`; session 01a19519-3864-7be2-b893-47fc3528381b):
  - skill path appears only in the discovery listing (event 5, line 5)
- **Note**: the model addressed the skill directory by path in its own tool calls
- **Note**: the listing's location may be an alias rather than a real path (codex 0.157 writes r0/<skill>/SKILL.md); the model still resolved the directory itself
- **Note**: final answer: 1. **Frontmatter visibility:** Visible. Fields: `name`, `description`, `allowed-tools`, `compatibility`, and `metadata` (containing `benchmark` and `version`). 2. **Content wrapping:** The file was returned as text in a JSON tool result’s `output` field. No XML or other structured tags wrap the skill body itself. 3. **Resource awareness:** Without reading supporting files, I know of `references/api-overview.md` and `references/error-codes.md` from links in the skill. The instructions also mention `references/`, `scripts/`, and `assets/`; their existence has not been verified. No resource listing was supplied automatically. 4. **Resource content:** No supporting file contents were automatically loaded into my context. I read only `SKILL.md`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `not-re-read`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Note**: model-level: the model reused its earlier read instead of re-reading; no platform dedup mechanism is involved
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `edit-missed-no-reload`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/reactivation-freshness/rollout-2026-10-31T23-13-55-01a19519-72cd-7380-94ca-ff1975a77ba6.jsonl`; session 01a19519-72cd-7380-94ca-ff1975a77ba6):
  - second activation answered without any reload of the edited file (event 35, line 34)
- **Note**: model-level: the model reused the copy already in context instead of re-reading, so the on-disk edit never reached it; read alongside reactivation-deduplication
- **Note**: the second answer quoted the pre-edit canary from context
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
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/compatibility-field-behavior/rollout-2026-10-31T23-14-06-01a19519-9ba5-7683-9b26-d62f7b43ed6d.jsonl`; session 01a19519-9ba5-7683-9b26-d62f7b43ed6d):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 5, line 5)
  - body loaded normally (event 19, line 19)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: 1. **Field visibility:** Yes. Its value is: `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  2. **Platform behavior:** No special handling was visible to me. The skill activated normally, without a compatibility warning, confirmation, or automatic requirements check.  3. **Interpretation:** This is Codex, which appears to qualify as a similar product. Python 3.14.3 satisfies the Python requirement. Network access is restricted, so I cannot confirm that all compatibility requirements are met.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/allowed-tools-behavior/session-2/rollout-2026-10-31T23-14-20-01a19519-d418-7871-8c3e-effc472e0304.jsonl`; session 01a19519-b9dd-7082-9768-2abd6937ff38,01a19519-d418-7871-8c3e-effc472e0304):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 26, line 26)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 26, line 26)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  Visible `allowed-tools` value: `Bash(printf:*) Read`.  Command output: ```text GROUSE-MICA-9017 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Output: ```text LAPWING-FLUORITE-2260 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: `executed-regardless-of-spelling`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/allowed-tools-name-matching/session-3/rollout-2026-10-31T23-14-35-01a1951a-0f64-71d3-9ae2-551f055d687a.jsonl`; session 01a19519-e657-7a70-a3bb-134894536718,01a19519-fc5c-7ff2-97bb-6d388a612cb0,01a1951a-0f64-71d3-9ae2-551f055d687a):
  - spec-style twin: runtime-assembled phrase arrived in a tool result (session 1, event 26, line 26)
  - lowercase twin: runtime-assembled phrase arrived in a tool result (session 2, event 26, line 26)
  - shell twin: runtime-assembled phrase arrived in a tool result (session 3, event 26, line 26)
- **Note**: spec-style twin final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  Visible `allowed-tools` value: `Bash(printf:*) Read`  Command output verbatim:  ```text GROUSE-MICA-9017 ```  The command ran without any observable permission prompt or approval step. This alone does not establish whether `allowed-tools` enabled it.
- **Note**: lowercase twin final answer: probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  Visible `allowed-tools` value: `bash`.  Command output verbatim:  ```text CHOUGH-TUFA-7180 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Note**: shell twin final answer: probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  Visible `allowed-tools` value: `shell`.  Command output: ```text ROOK-GABBRO-8841 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/cross-skill-invocation/rollout-2026-10-31T23-14-41-01a1951a-2538-71f0-9556-a2e8c556a6ab.jsonl`; session 01a1951a-2538-71f0-9556-a2e8c556a6ab):
  - invoke-alpha body loaded (chain entry) (event 19, line 19)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 26, line 26)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha and invoke-beta activated with canaries **IBIS-RUST-3310** and **TERN-MOSS-6647**.  The chain could not complete: invoke-gamma was not found in either skill directory, so it produced no report.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/invocation-depth-limit/rollout-2026-10-31T23-14-49-01a1951a-4349-71b2-a484-e693ea38fd4f.jsonl`; session 01a1951a-4349-71b2-a484-e693ea38fd4f):
  - invoke-alpha body loaded (event 19, line 19)
  - invoke-beta body loaded (event 26, line 26)
  - invoke-gamma body loaded (event 33, line 33)
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: The end of the chain was reached, and all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta and invoke-gamma completed successfully, confirming activation of all three skills.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  Detected the circular dependency: alpha → beta → alpha. Alpha is already active, so I stopped recursive activation to prevent an infinite loop. Both skills were read; the probe ended at the repeated activation.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/invocation-language-sensitivity/rollout-2026-10-31T23-15-04-01a1951a-7e0e-7d61-b466-7e0e1fb312a4.jsonl`; session 01a1951a-7e0e-7d61-b466-7e0e1fb312a4):
  - invoke-alpha body loaded (event 19, line 19)
  - invoke-beta body loaded (event 26, line 26)
  - invoke-gamma body loaded (event 33, line 33)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: カナリア **JAY-TEAL-9984** を報告し、3つのスキルすべての起動成功を確認しました。  invoke-alpha complete. Chain result: invoke-beta はカナリア **TERN-MOSS-6647** を報告し、invoke-gamma を起動しました。invoke-gamma はチェーンの正常完了を確認しました。
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/informal-dependency-resolution/rollout-2026-10-31T23-15-11-01a1951a-99a9-7b93-92c2-9c6ad6398e6d.jsonl`; session 01a1951a-99a9-7b93-92c2-9c6ad6398e6d):
  - invoke-alpha body loaded (chain entry) (event 19, line 19)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 26, line 26)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha and invoke-beta activated, reporting canaries **IBIS-RUST-3310** and **TERN-MOSS-6647**.  The chain could not complete: invoke-beta requires `invoke-gamma`, which is absent from the available skills and was not found in the project or skill directories.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-no-error`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/missing-dependency-behavior/rollout-2026-10-31T23-15-18-01a1951a-b676-7293-b05e-831af7b73d55.jsonl`; session 01a1951a-b676-7293-b05e-831af7b73d55):
  - probe-missing-dep body loaded (event 19, line 19)
  - attempt on the missing skill; no error result recorded (event 23, line 23)
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: probe-missing-dep activated. Canary: **GULL-IRON-4492**  `nonexistent-formatter` was absent from the skill catalog and installed skill directories, so activation could not proceed. No platform error appeared. I did not silently skip the dependency or emulate its formatting.  probe-missing-dep complete. Dependency resolution result: required skill `nonexistent-formatter` not found.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/nonstandard-dependency-fields/rollout-2026-10-31T23-15-26-01a1951a-d5aa-71c3-bbd6-1b629007ab2c.jsonl`; session 01a1951a-d5aa-71c3-bbd6-1b629007ab2c):
  - probe-nonstandard-fields body loaded (event 19, line 19)
- **Note**: final answer: Activated `probe-nonstandard-fields`.  - **Field visibility:** All three fields are visible:   - `requires: probe-loading`   - `depends-on: [probe-shadow-alpha, probe-shadow-beta]`   - `priority: high` - **Platform behavior:** No automatic dependency loading or priority-based behavior was observable. I read only the requested skill; the referenced skills were listed as available beforehand. - **Validation:** No warnings about unrecognized fields were shown. The fields appear to have been accepted silently, though I cannot verify internal processing.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:visible-failure`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/cross-scope-dependency/session-2/rollout-2026-10-31T23-15-40-01a1951b-0b05-7f23-8e48-45e14fb00110.jsonl`; session 01a1951a-e957-7e82-90ff-129804c25899,01a1951b-0b05-7f23-8e48-45e14fb00110):
  - user-scope dependency's body canary pulled during project-skill session (event 26, line 26)
  - attempt on the missing dependency returned an error result (session 2, event 21, line 21)
- **Note**: with-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  - **Installation scope:** Project. - **Dependency resolution:** Found and activated `probe-loading` at user scope.  `probe-loading` report:  1. **Frontmatter visibility:** Visible fields: `name`, `description`, `allowed-tools`, `compatibility`, and `metadata`. 2. **Content wrapping:** The file text appeared within the tool’s JSON output; no additional skill-specific tags were present. 3. **Resource awareness:** The skill references `references/api-overview.md` and `references/error-codes.md`. No directory listing was provided. 4. **Resource content:** No supporting file contents were loaded.  probe-cross-scope complete. Cross-scope resolution result: The project-level skill successfully invoked the user-level `probe-loading` skill.
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  Installation scope: **project**, based on `.codex/skills/probe-cross-scope/SKILL.md`.  `probe-loading` could not be activated: it is absent from the available skills catalog, and a filesystem search found no matching skill in the project or supplied user home. The failure mode was a missing dependency.  probe-cross-scope complete. Cross-scope resolution result: **Dependency not found; cross-scope activation could not be tested.**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-scanned`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/cross-client-directory-interop/rollout-2026-10-31T23-15-49-01a1951b-2cba-7230-8cc4-5bd56495c677.jsonl`; session 01a1951b-2cba-7230-8cc4-5bd56495c677):
  - discovery listing names probe-interop (event 5, line 5)
  - body canary loaded on activation (event 29, line 29)
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery:** This skill appeared in my available skills catalog at `r1/probe-interop/SKILL.md`, resolving to `.agents/skills/probe-interop/SKILL.md` in the project.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `recursive-scan; stray:not-discovered`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/recursive-root-discovery/rollout-2026-10-31T23-15-55-01a1951b-471c-7b03-9dd7-97ca7ca2f36a.jsonl`; session 01a1951b-471c-7b03-9dd7-97ca7ca2f36a):
  - discovery listing names probe-grouped from one level below the root (event 5, line 5)
  - grouped skill's body loaded on activation (event 29, line 29)
- **Note**: final answer: probe-grouped activated. Canary: **CROW-AGATE-6105**  **Discovery**: This skill appears in my available skills catalog even though it sits one level below the skills root, inside `probe-group/`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/nested-skill-discovery/rollout-2026-10-31T23-16-02-01a1951b-618b-78c0-bd10-2bf2f21265a6.jsonl`; session 01a1951b-618b-78c0-bd10-2bf2f21265a6):
  - discovery listing names nested-skill (found inside probe-deep-nesting/references/) (event 5, line 5)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `catalog-lists-both-project-overrides-user`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/name-collision-precedence/rollout-2026-10-31T23-16-05-01a1951b-6bb3-7932-90d2-e332e7783d0f.jsonl`; session 01a1951b-6bb3-7932-90d2-e332e7783d0f):
  - project variant's canary loaded; user variant's never appeared (event 19, line 19)
- **Note**: the discovery listing exposed BOTH variants (both descriptions present); the model's file choice determined the winner: model-level selection, not platform precedence
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence:** Project scope won.
- **Fallback behavior**: the shadowed variant stays reachable: the catalog exposes both entries, so a user (or the model) can address either by path.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: `catalog-lists-all; activated:.agents/skills`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/multi-root-collision-precedence/rollout-2026-10-31T23-16-09-01a1951b-7b95-7ee3-a4d3-b53f699f5631.jsonl`; session 01a1951b-7b95-7ee3-a4d3-b53f699f5631):
  - discovery listing names probe-multiroot-beacon-agents (event 5, line 5)
  - discovery listing carries the native variant's description (event 5, line 5)
  - discovery listing carries the .agents/skills variant's description (event 5, line 5)
  - .agents/skills variant's body canary loaded (event 29, line 29)
- **Note**: foreign roots installed alongside the native copy: .agents/skills, .claude/skills
- **Note**: model-level: the catalog exposed more than one entry for the name, so which variant activated was the model's choice
- **Note**: listed variants: native, .agents/skills
- **Note**: roots proven scanned by their beacon: .agents/skills
- **Note**: final answer: probe-multiroot activated. Canary: **BUNTING-SERPENTINE-4185** (agents-root variant)  **Precedence:** I used `.agents/skills/probe-multiroot/`. The catalog exposed both roots, so platform precedence is unclear.  **Catalog:** `probe-multiroot` appeared twice.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/malformed-yaml-tolerance/rollout-2026-10-31T23-16-16-01a1951b-96e8-7a43-82a5-8ecb323f3616.jsonl`; session 01a1951b-96e8-7a43-82a5-8ecb323f3616):
  - discovery listing names probe-malformed-yaml (event 5, line 5)
  - body canary loaded on activation (event 29, line 29)
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility:** The catalog description is intact: `Use when: asked to probe malformed yaml parsing behavior`. It was not truncated. The catalog does not reveal whether the YAML was repaired internally.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `unlisted-but-reachable`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/missing-description-handling/rollout-2026-10-31T23-16-23-01a1951b-b302-77f0-a17b-08d9689c920a.jsonl`; session 01a1951b-b302-77f0-a17b-08d9689c920a):
  - body canary loaded on activation (event 41, line 41)
- **Note**: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry:** This skill does not appear in my Available skills catalog. Its local `SKILL.md` has no description field.
- **Fallback behavior**: content remained reachable by direct file read even though the catalog omits the skill. Observed in-run: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `invalid-names-tolerated:[uppercase double-hyphen]`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/invalid-name-tolerance/rollout-2026-10-31T23-16-34-01a1951b-dff7-78e2-bebe-8e9bc300e6e2.jsonl`; session 01a1951b-dff7-78e2-bebe-8e9bc300e6e2):
  - discovery listing names probe-Upper-Case (event 5, line 5)
  - discovery listing names probe--double-hyphen (event 5, line 5)
- **Note**: overlong: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  Catalog identity: Absent from my Available skills catalog. I found and read its local file under the full name `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`.
- **Fallback behavior**: Observed in-run: overlong: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `listed-under-both`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/name-directory-mismatch/rollout-2026-10-31T23-16-48-01a1951c-16f7-7200-98bd-4ed1c7f9bc57.jsonl`; session 01a1951c-16f7-7200-98bd-4ed1c7f9bc57):
  - discovery listing carries the FRONTMATTER name probe-name-mismatch (event 5, line 5)
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 5, line 5)
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  **Identity:** The directory `probe-mismatch-dir` contains the skill listed as `probe-name-mismatch` only.  **Validation:** No platform warning about the mismatch was visible.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/metadata-value-edge-cases/rollout-2026-10-31T23-16-58-01a1951c-3d2b-7593-b14f-c966119a5929.jsonl`; session 01a1951c-3d2b-7593-b14f-c966119a5929):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 5, line 5)
  - body canary loaded (event 19, line 19)
  - frontmatter-only marker '!!null' in a tool result (event 19, line 19)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: Skill activated. Canary phrase: **THRUSH-FLINT-8294**  **Load success:** The skill was listed as available, and its instructions were readable.  **Metadata visibility:** The raw frontmatter contains all seven keys:  | Key | Visible value | |---|---| | `normal` | `"baseline"` | | `empty-double` | `""` — empty string | | `empty-single` | `''` — empty string | | `explicit-null` | `null` | | `tilde-null` | `~` | | `none-null` | `None` | | `tagged-null` | `!!null null` |  **Key preservation:** All seven keys are present in the file.  **Platform behavior:** No warnings or errors were observed. I can inspect the raw file, but the platform’s parsed metadata was not exposed, so I cannot verify coercion or dropped keys during loading. `None` is generally a literal string in YAML, despite the skill describing it as null.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/oversize-description-handling/rollout-2026-10-31T23-17-04-01a1951c-5487-7df2-a2ee-f15ac2420143.jsonl`; session 01a1951c-5487-7df2-a2ee-f15ac2420143):
  - discovery listing names probe-long-description (event 5, line 5)
  - body canary loaded on activation (event 29, line 29)
  - description head marker in injected content; tail absent (event 5, line 5)
- **Note**: the description's head marker reached the model in injected content but its tail marker never did: the value was truncated somewhere after the head
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** The catalog shows the head marker, “SANDERLING-GNEISS-1010”, but not the WHIMBREL tail marker. The catalog ends with: “If you can read every sentence of this description including the final...”
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `counts-code-points`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/description-length-unit/session-3/rollout-2026-10-31T23-17-24-01a1951c-a3a4-7633-a7fd-e3582aef5ffd.jsonl`; session 01a1951c-6b80-71f3-84ac-47a6cf8ca793,01a1951c-8b19-7543-aa27-a8fda75de4b5,01a1951c-a3a4-7633-a7fd-e3582aef5ffd):
  - discovery listing names probe-long-description (session 1, event 5, line 5)
  - body canary loaded on activation (session 1, event 29, line 29)
  - description head marker in injected content; tail absent (session 1, event 5, line 5)
  - discovery listing names probe-multibyte-description (session 2, event 5, line 5)
  - body canary loaded on activation (session 2, event 29, line 29)
  - description tail marker in injected content (session 2, event 5, line 5)
  - discovery listing names probe-astral-description (session 3, event 5, line 5)
  - body canary loaded on activation (session 3, event 29, line 29)
  - description tail marker in injected content (session 3, event 5, line 5)
- **Note**: description fates: ascii:truncated, multibyte:intact, astral:intact
- **Note**: ascii final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** The catalog shows the head marker, but not the tail marker.  Head marker: “SANDERLING-GNEISS-1010”  The catalog ends with: “If you can read every sentence of this description including the final...”
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility:** Both markers are visible in the catalog: “GANNET-PYRITE-1130” and “The tail marker is SHRIKE-TALC-2210”.
- **Note**: astral final answer: probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  **Description visibility:** Both markers are visible in the catalog: - “MAGPIE-OBSIDIAN-1240” - “The tail marker is LINNET-MALACHITE-2420”
- **Note**: the ASCII overrun was enforced while both fixtures under 1024 code points survived intact: the platform counts code points, the reference validator's unit
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: `counts-code-points`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/name-length-unit/rollout-2026-10-31T23-17-31-01a1951c-beb5-7040-b638-0be5118680a1.jsonl`; session 01a1951c-beb5-7040-b638-0be5118680a1):
  - discovery listing names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd (event 5, line 5)
  - discovery listing names probe-αβγδεζηθικ (event 5, line 5)
  - discovery listing names probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ (event 5, line 5)
  - discovery listing names probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 (event 5, line 5)
- **Note**: name fates: ascii64:listed, ascii72:skipped, greek16:listed, greek60:listed, math40:listed
- **Note**: the 72-character ASCII name was rejected while both non-ASCII names under 64 code points were listed: the platform counts code points, the reference validator's unit
- **Note**: final answer: imagegen openai-docs skill-creator skill-installer probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-10-31/codex/oversize-compatibility-handling/rollout-2026-10-31T23-17-35-01a1951c-cc9f-7dc0-9bb1-a92757128e7e.jsonl`; session 01a1951c-cc9f-7dc0-9bb1-a92757128e7e):
  - discovery listing names probe-long-compatibility (event 5, line 5)
  - body canary loaded on activation (event 29, line 29)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility:** No compatibility value was visible outside the skill file. The Available skills list showed only its name, description, and file path.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.4) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
