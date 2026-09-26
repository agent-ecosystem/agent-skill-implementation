# Platform Loading Implementation: Codex CLI (headless)

| | |
|---|---|
| **Platform** | Codex CLI (headless) |
| **Platform version** | 0.157.0 |
| **Check list version** | 0.4 |
| **Test date** | 2026-09-26 |
| **Model(s) observed** | gpt-6-astra |
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
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/discovery-reading-depth/rollout-2026-09-25T22-54-23-01a0dba2-a19e-79b1-817a-9cfe68005dcb.jsonl`; session 01a0dba2-a19e-79b1-817a-9cfe68005dcb):
  - discovery listing names probe-loading (event 3, line 3)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/activation-loading-scope/rollout-2026-09-25T22-54-28-01a0dba2-b62b-7a73-b5f8-08d0fb070a00.jsonl`; session 01a0dba2-b62b-7a73-b5f8-08d0fb070a00):
  - model's tool call targets the skill's own path (event 13, line 13)
  - body canary arrived in the tool result (event 16, line 16)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/eager-link-resolution/rollout-2026-09-25T22-54-41-01a0dba2-e74a-7020-840e-46f61a9639fa.jsonl`; session 01a0dba2-e74a-7020-840e-46f61a9639fa):
  - skill body loaded (event 16, line 16)
  - references/setup-guide.md arrived only via the model's own read (event 24, line 24)
  - references/troubleshooting.md arrived only via the model's own read (event 24, line 24)
  - references/unlinked-data.md arrived only via the model's own read (event 24, line 24)
- **Note**: model read [references/setup-guide.md references/troubleshooting.md references/unlinked-data.md] itself, corroborating it did not already have them
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
- **Note**: final answer: Activated the `probe-nonstandard-dirs` skill and inspected its directory listing.  | Directory | Visible file | Contents automatically visible? | |---|---|---| | `evals/` | `evals.json` | No | | `templates/` | `output-template.md` | No | | `resources/` | `api-reference.md` | No |  These are all observed subdirectories. The skill instructions named them; I explicitly used filesystem tools to obtain their file listings. I read only `SKILL.md`, not the listed files.  `resources/` is accessible through filesystem tools, but its contents were neither automatically enumerated nor loaded. There is no `references/` directory here, so this probe cannot establish whether the platform treats the two names identically.
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
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/cross-skill-resource-shadowing/rollout-2026-09-25T22-55-49-01a0dba3-f006-7632-b5b3-88a7ec53e5b1.jsonl`; session 01a0dba3-f006-7632-b5b3-88a7ec53e5b1):
  - activated skill's own API.md content arrived first (event 21, line 21)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: final answer: Contents of alpha’s `references/API.md`:  ```markdown # API Reference (Alpha)  Shadow-alpha canary phrase: **STORK-CORAL-4471**  This file belongs to the `probe-shadow-alpha` skill. If you see this content when reading from the `probe-shadow-beta` skill's context, the platform is resolving resource paths ambiguously across skills.  ## Alpha Endpoints  - `GET /alpha/status` - Alpha health check - `POST /alpha/process` - Alpha processing ```  The canary **STORK-CORAL-4471** confirms the alpha version. Only alpha is active, so the conditional cross-skill check does not apply.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/path-traversal-boundary/rollout-2026-09-25T22-55-55-01a0dba4-0a30-7303-8fde-2ac456f4c819.jsonl`; session 01a0dba4-0a30-7303-8fde-2ac456f4c819):
  - sibling skill's body canary arrived in a tool result (event 23, line 23)
  - first ../ path attempt (event 18, line 18)
- **Note**: relative ../ path forms were used
- **Note**: final answer: Activated `probe-traversal` and completed all three attempts.  | Attempt | Result | Contents returned | |---|---|---| | `../probe-loading/SKILL.md` | Success; read allowed | Full file contents | | List `../` | Success; listing allowed | `probe-loading` and `probe-traversal` directories | | `../../checks.md` | File not found | No file contents; explicit error |  All commands emitted a pyenv warning that its shims directory wasn’t writable. No attempt silently returned nothing or reported a boundary block.  The shell permitted access outside the skill directory; the missing file does not establish whether that read would otherwise be allowed.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/resource-nesting-depth/rollout-2026-09-25T22-56-09-01a0dba4-4098-7553-929f-fae396b55b58.jsonl`; session 01a0dba4-4098-7553-929f-fae396b55b58):
  - depth-1 file references/overview.md content arrived (event 25, line 25)
  - depth-2 file references/api/endpoints.md content arrived (event 25, line 25)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 25, line 25)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 25, line 25)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 25, line 25)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `script-executed`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/bundled-script-execution/rollout-2026-09-25T22-56-30-01a0dba4-8f5e-7e02-981f-d98de178df21.jsonl`; session 01a0dba4-8f5e-7e02-981f-d98de178df21):
  - tool call references the bundled script (event 20, line 20)
  - runtime-assembled output phrase arrived in a tool result (event 23, line 23)
- **Note**: final answer: probe-script-execution activated. Canary: **REDSHANK-SYENITE-8807**  The script executed successfully and printed: `GODWIT-BORNITE-5148`
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: `no-enumeration`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: final answer: probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  **File awareness:** I was told about `SKILL.md` only. No bundled supporting files were listed.  **Count:** 1 file.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `surfaces-beyond-description:[location]`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/discovery-listing-fields/rollout-2026-09-25T22-56-41-01a0dba4-bd26-70f3-8c1f-554a8c209264.jsonl`; session 01a0dba4-bd26-70f3-8c1f-554a8c209264):
  - discovery listing names the installed skills (event 3, line 3)
  - description text surfaced at discovery (event 3, line 3)
  - SKILL.md file path surfaced at discovery (event 3, line 3)
- **Note**: final answer: ### Skill roots - `r0` = `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1350165750/project/.codex/skills` - `r1` = `/Users/dachary/.codex/skills/.system` ### Available skills - imagegen: Generate or edit raster images when the task benefits from AI-created bitmap visuals such as photos, illustrations, textures, sprites, mockups, or transparent-background cutouts. Use when Codex should create a brand-new image, transform an existing image, or derive visual variants from references, and the output should be a bitmap asset rather than repo-native code or vector. Do not use when the task is better handled by editing existing SVG/vector/code-native assets, extending an established icon or logo system, or building the visual directly in HTML/CSS/canvas. (file: r1/imagegen/SKILL.md) - openai-docs: Use for Codex models/pricing, scheduled tasks, skills, settings, setup, troubleshooting, customization, automations, and self-knowledge—including 'you,' 'your,' 'this app,' or 'this coding agent' when they refer to Codex—and for OpenAI APIs/products and ChatGPT Work. Also use for model choice/migration, prompting, SDKs, Responses, Realtime, agents, evals, and Chat/Work/Codex c…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-visible-via-raw-read`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/frontmatter-handling/rollout-2026-09-25T22-56-51-01a0dba4-e336-7c41-bc89-adeb19161580.jsonl`; session 01a0dba4-e336-7c41-bc89-adeb19161580):
  - raw file read delivered frontmatter and body together (event 16, line 16)
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-file-via-pull`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/content-wrapping-format/rollout-2026-09-25T22-56-57-01a0dba4-f9bd-7f43-af4f-6d76a3c0fe67.jsonl`; session 01a0dba4-f9bd-7f43-af4f-6d76a3c0fe67):
  - body arrived as a file-read tool result (event 16, line 16)
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: `path-only-in-discovery-listing`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/activation-location-disclosure/rollout-2026-09-25T22-57-07-01a0dba5-22f6-7e11-aba3-ff91dbf8e9d5.jsonl`; session 01a0dba5-22f6-7e11-aba3-ff91dbf8e9d5):
  - skill path appears only in the discovery listing (event 3, line 3)
- **Note**: the model addressed the skill directory by path in its own tool calls
- **Note**: the listing's location may be an alias rather than a real path (codex 0.157 writes r0/<skill>/SKILL.md); the model still resolved the directory itself
- **Note**: final answer: 1. **Frontmatter visibility:** Visible. Fields: `name`, `description`, `allowed-tools`, `compatibility`, and `metadata` (containing `benchmark` and `version`).  2. **Content wrapping:** The skill is Markdown with YAML frontmatter. It arrived as a string in a JSON tool result, without additional XML-style wrapping.  3. **Resource awareness:** From the skill text alone, I know of `references/`, `scripts/`, and `assets/`, plus two referenced files:    - `references/api-overview.md`    - `references/error-codes.md`     No directory listing was supplied; I have not verified these resources.  4. **Resource content:** No supporting file contents were automatically loaded into my context. I read only `SKILL.md`.
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
- **Runs**: 3 (edit-missed-no-reload ×3)
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/reactivation-freshness/run-01/rollout-2026-09-25T23-14-06-01a0dbb4-ae51-7010-ae84-91ce8911dba3.jsonl`; session 01a0dbb4-ae51-7010-ae84-91ce8911dba3):
  - second activation answered without any reload of the edited file (event 31, line 30)
- **Note**: model-level: the model reused the copy already in context instead of re-reading, so the on-disk edit never reached it; read alongside reactivation-deduplication
- **Note**: the second answer quoted the pre-edit canary from context
- **Note**: verdict consistent across 3 runs
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
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/compatibility-field-behavior/rollout-2026-09-25T22-57-46-01a0dba5-b9e5-7783-9f6e-bd70cd5eb44f.jsonl`; session 01a0dba5-b9e5-7783-9f6e-bd70cd5eb44f):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 3, line 3)
  - body loaded normally (event 16, line 16)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: 1. **Field visibility:** Yes. Its value is: `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  2. **Platform behavior:** No special handling was visible before activation: no warning, confirmation request, or requirements check. The skill activated normally.  3. **Requirements:** This is Codex, reasonably interpreted as a “similar product.” Python 3.14.3 meets the Python requirement. Network access is restricted, so full compatibility is not confirmed.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/allowed-tools-behavior/session-2/rollout-2026-09-25T22-57-59-01a0dba5-edf9-7bc3-a573-df5f5471c4c1.jsonl`; session 01a0dba5-d66e-7ca3-b18c-faad5e7d4964,01a0dba5-edf9-7bc3-a573-df5f5471c4c1):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 23, line 23)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 23, line 23)
- **Note**: allowed-tools value visible only via the model's own raw file read
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  Visible `allowed-tools` value: `Bash(printf:*) Read`  Command output: ```text GROUSE-MICA-9017 ```  The command ran without any observable permission prompt or approval step. This alone does not establish whether `allowed-tools` enabled execution.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Command output: ```text LAPWING-FLUORITE-2260 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Fallback behavior**: Observed in-run: allowed-tools value visible only via the model's own raw file read

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: `executed-regardless-of-spelling`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/allowed-tools-name-matching/session-3/rollout-2026-09-25T22-58-21-01a0dba6-43cd-7de2-9b17-7aed5227a78a.jsonl`; session 01a0dba6-10c8-7bb1-89a6-8729fa8c8787,01a0dba6-27f0-7d63-aaf0-b51dbeefb830,01a0dba6-43cd-7de2-9b17-7aed5227a78a):
  - spec-style twin: runtime-assembled phrase arrived in a tool result (session 1, event 23, line 23)
  - lowercase twin: runtime-assembled phrase arrived in a tool result (session 2, event 23, line 23)
  - shell twin: runtime-assembled phrase arrived in a tool result (session 3, event 23, line 23)
- **Note**: spec-style twin final answer: probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  Visible `allowed-tools` value: `Bash(printf:*) Read`.  Command output verbatim:  ```text GROUSE-MICA-9017 ```  The command ran without any observable permission prompt or approval step. This alone does not establish whether the field enabled execution.
- **Note**: lowercase twin final answer: probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  Visible `allowed-tools` value: `bash`.  Command output: ```text CHOUGH-TUFA-7180 ```  The command ran without any permission prompt or approval step, as far as I could observe.
- **Note**: shell twin final answer: probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  Visible `allowed-tools` value: `shell`.  Command output: ```text ROOK-GABBRO-8841 ```  The command ran without any observable permission prompt or approval step.
- **Note**: the platform's permission posture allowed every twin's command; spelling had no observable effect
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/cross-skill-invocation/rollout-2026-09-25T22-58-28-01a0dba6-5e70-7a10-80b4-9754d53fc32a.jsonl`; session 01a0dba6-5e70-7a10-80b4-9754d53fc32a):
  - invoke-alpha body loaded (chain entry) (event 16, line 16)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 23, line 23)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha activated. Canary: **IBIS-RUST-3310**  invoke-beta activated. Canary: **TERN-MOSS-6647**  The chain could not complete: `invoke-gamma` is unavailable in the project and system skill directories. Neither skill’s completion step was reached.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/invocation-depth-limit/rollout-2026-09-25T22-58-39-01a0dba6-8917-75d3-b6e8-a2d4c4e9a420.jsonl`; session 01a0dba6-8917-75d3-b6e8-a2d4c4e9a420):
  - invoke-alpha body loaded (event 16, line 16)
  - invoke-beta body loaded (event 23, line 23)
  - invoke-gamma body loaded (event 30, line 30)
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: Canary **JAY-TEAL-9984**; end of chain reached successfully.  invoke-alpha complete. Chain result: invoke-beta reported canary **TERN-MOSS-6647** and completed. invoke-gamma reported canary **JAY-TEAL-9984** and confirmed all three skills activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  Detected the circular dependency: alpha → beta → alpha. Alpha is already active, so I stopped recursive activation to prevent an infinite loop. Both skills’ instructions were read; this was an assistant-level guard, not evidence of platform-level detection.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/invocation-language-sensitivity/rollout-2026-09-25T22-58-59-01a0dba6-d891-79e3-b469-b3d82da6e71c.jsonl`; session 01a0dba6-d891-79e3-b469-b3d82da6e71c):
  - invoke-alpha body loaded (event 16, line 16)
  - invoke-beta body loaded (event 23, line 23)
  - invoke-gamma body loaded (event 30, line 30)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: 最終段階に到達し、3つのスキルすべてが正常に起動しました。  invoke-alpha complete. Chain result: invoke-beta（**TERN-MOSS-6647**）が invoke-gamma（**JAY-TEAL-9984**）を起動し、チェーン全体の正常完了を報告しました。
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/informal-dependency-resolution/rollout-2026-09-25T22-59-12-01a0dba7-08bf-7c41-8622-d5abc4fea118.jsonl`; session 01a0dba7-08bf-7c41-8622-d5abc4fea118):
  - invoke-alpha body loaded (chain entry) (event 16, line 16)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 23, line 23)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha and invoke-beta activated, reporting canaries **IBIS-RUST-3310** and **TERN-MOSS-6647**.  The chain cannot complete because invoke-gamma was not found in either skill directory. [invoke-beta’s instructions](/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1094994251/project/.codex/skills/invoke-beta/SKILL.md) require: “Now activate the `invoke-gamma` skill to continue the chain.”
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-no-error`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/missing-dependency-behavior/rollout-2026-09-25T22-59-21-01a0dba7-2ebd-78e2-a423-a39ea98e8147.jsonl`; session 01a0dba7-2ebd-78e2-a423-a39ea98e8147):
  - probe-missing-dep body loaded (event 16, line 16)
  - attempt on the missing skill; no error result recorded (event 20, line 20)
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: probe-missing-dep activated. Canary: **GULL-IRON-4492**  probe-missing-dep complete. Dependency resolution result: `nonexistent-formatter` was absent from the skill catalog and searched directories, so activation could not proceed. No platform error appeared. I did not silently skip the dependency or substitute formatting from general knowledge.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/nonstandard-dependency-fields/rollout-2026-09-25T22-59-32-01a0dba7-57c9-7f82-9fa0-7e01f9fd78c2.jsonl`; session 01a0dba7-57c9-7f82-9fa0-7e01f9fd78c2):
  - probe-nonstandard-fields body loaded (event 16, line 16)
- **Note**: final answer: Activated `probe-nonstandard-fields`.  1. **Field visibility:** All three fields are visible:    - `requires`: `probe-loading`    - `depends-on`: `probe-shadow-alpha`, `probe-shadow-beta`    - `priority`: `high`  2. **Platform behavior:** No automatic dependency loading or priority-based behavior was observable. I read only this skill. The referenced skills were already listed as available; that does not demonstrate dependency handling.  3. **Validation:** No warnings appeared. The fields were presented silently, though I cannot confirm whether the platform internally validated or ignored them.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:visible-failure`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/cross-scope-dependency/session-2/rollout-2026-09-25T23-02-09-01a0dba9-bdbd-7292-914b-4b1c7f202e6d.jsonl`; session 01a0dba9-992a-7972-9c89-87986a533488,01a0dba9-bdbd-7292-914b-4b1c7f202e6d):
  - user-scope dependency's body canary pulled during project-skill session (event 23, line 23)
  - attempt on the missing dependency returned an error result (session 2, event 18, line 18)
- **Note**: with-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  Installation scope: **project**. The dependency `probe-loading` was found and read successfully at **user** scope.  `probe-loading` report: 1. **Frontmatter visibility:** Visible fields: `name`, `description`, `allowed-tools`, `compatibility`, `metadata` (including `benchmark` and `version`). 2. **Content wrapping:** The file was returned as text inside a JSON tool result; no skill-specific XML wrapping appeared. 3. **Resource awareness:** The body mentions `references/`, `scripts/`, and `assets/`, and links `references/api-overview.md` and `references/error-codes.md`. No directory listing was supplied. 4. **Resource content:** No supporting file contents were loaded or read.  Dependency canary: **CARDINAL-ZEBRA-7742**  probe-cross-scope complete. Cross-scope resolution result: **The project-level skill successfully resolved and activated the user-level dependency through the skill catalog and filesystem read.**
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  - Installation scope: **project-level**, inferred from `.codex/skills/probe-cross-scope/SKILL.md`. - `probe-loading` was absent from the available skill catalog and not found in the project or supplied user-home directories. Activation could not proceed because the dependency was missing; its scope could not be determined.  probe-cross-scope complete. Cross-scope resolution result: **dependency not found**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-scanned`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/cross-client-directory-interop/rollout-2026-09-25T22-59-37-01a0dba7-6c99-7472-a4d7-678d5d9eeaa6.jsonl`; session 01a0dba7-6c99-7472-a4d7-678d5d9eeaa6):
  - discovery listing names probe-interop (event 3, line 3)
  - body canary loaded on activation (event 27, line 27)
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery:** This skill appeared in my available skills catalog at `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-286000029/project/.agents/skills/probe-interop/SKILL.md`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `recursive-scan; stray:not-discovered`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/recursive-root-discovery/rollout-2026-09-25T22-59-44-01a0dba7-8543-7902-91b9-2f0b8e27cdec.jsonl`; session 01a0dba7-8543-7902-91b9-2f0b8e27cdec):
  - discovery listing names probe-grouped from one level below the root (event 3, line 3)
  - grouped skill's body loaded on activation (event 27, line 27)
- **Note**: final answer: probe-grouped activated. Canary: **CROW-AGATE-6105**  **Discovery**: This skill appears in my available skills catalog even though it sits one level below the skills root, at `probe-group/probe-grouped/`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/nested-skill-discovery/rollout-2026-09-25T22-59-51-01a0dba7-a137-7b31-b4a6-0e4a6f45c64a.jsonl`; session 01a0dba7-a137-7b31-b4a6-0e4a6f45c64a):
  - discovery listing names nested-skill (found inside probe-deep-nesting/references/) (event 3, line 3)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `catalog-lists-both-project-overrides-user`
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/name-collision-precedence/rollout-2026-09-25T23-02-20-01a0dba9-e8fe-7b61-854c-09751b22a2f2.jsonl`; session 01a0dba9-e8fe-7b61-854c-09751b22a2f2):
  - project variant's canary loaded; user variant's never appeared (event 16, line 16)
- **Note**: the discovery listing exposed BOTH variants (both descriptions present); the model's file choice determined the winner: model-level selection, not platform precedence
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence:** Project scope won.
- **Fallback behavior**: the shadowed variant stays reachable: the catalog exposes both entries, so a user (or the model) can address either by path.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: `catalog-lists-all; activated:.agents/skills`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/multi-root-collision-precedence/rollout-2026-09-25T22-59-57-01a0dba7-b7cf-7873-8595-30680199ab92.jsonl`; session 01a0dba7-b7cf-7873-8595-30680199ab92):
  - discovery listing names probe-multiroot-beacon-agents (event 3, line 3)
  - discovery listing carries the native variant's description (event 3, line 3)
  - discovery listing carries the .agents/skills variant's description (event 3, line 3)
  - .agents/skills variant's body canary loaded (event 27, line 27)
- **Note**: foreign roots installed alongside the native copy: .agents/skills, .claude/skills
- **Note**: model-level: the catalog exposed more than one entry for the name, so which variant activated was the model's choice
- **Note**: listed variants: native, .agents/skills
- **Note**: roots proven scanned by their beacon: .agents/skills
- **Note**: final answer: probe-multiroot activated. Canary: **BUNTING-SERPENTINE-4185** (agents-root variant)  **Precedence:** I selected the first catalog entry, from `.agents/skills/`. Both variants were listed, so platform precedence is unclear.  **Catalog:** `probe-multiroot` appeared twice.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/malformed-yaml-tolerance/rollout-2026-09-25T23-00-05-01a0dba7-d8a9-79c2-96ac-c0d8fdbe5d2b.jsonl`; session 01a0dba7-d8a9-79c2-96ac-c0d8fdbe5d2b):
  - discovery listing names probe-malformed-yaml (event 3, line 3)
  - body canary loaded on activation (event 27, line 27)
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility:** The catalog shows `Use when: asked to probe malformed yaml parsing behavior` intact, with no truncation or visible added quotes. The catalog alone doesn’t reveal whether the underlying YAML was repaired.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `unlisted-but-reachable`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/missing-description-handling/rollout-2026-09-25T23-00-12-01a0dba7-f3f3-7c80-b1ec-20b2eea005e5.jsonl`; session 01a0dba7-f3f3-7c80-b1ec-20b2eea005e5):
  - body canary loaded on activation (event 32, line 32)
- **Note**: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry:** This skill does not appear in my Available skills catalog, so no description text is provided. I found and read its local `SKILL.md`.
- **Fallback behavior**: content remained reachable by direct file read even though the catalog omits the skill. Observed in-run: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `invalid-names-tolerated:[uppercase double-hyphen]`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/invalid-name-tolerance/rollout-2026-09-25T23-00-24-01a0dba8-2413-7e40-81a8-499619cd467a.jsonl`; session 01a0dba8-2413-7e40-81a8-499619cd467a):
  - discovery listing names probe-Upper-Case (event 3, line 3)
  - discovery listing names probe--double-hyphen (event 3, line 3)
- **Note**: overlong: not in the catalog, yet the body loaded: model file access, not platform acceptance
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  **Catalog identity:** This skill does not appear in my Available skills catalog. I found and read it locally under the full name `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`.
- **Fallback behavior**: Observed in-run: overlong: not in the catalog, yet the body loaded: model file access, not platform acceptance

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `listed-under-both`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/name-directory-mismatch/rollout-2026-09-25T23-00-43-01a0dba8-6eea-70a1-bf85-229f84b8a032.jsonl`; session 01a0dba8-6eea-70a1-bf85-229f84b8a032):
  - discovery listing carries the FRONTMATTER name probe-name-mismatch (event 3, line 3)
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 3, line 3)
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  **Identity:** Listed as `probe-name-mismatch` (frontmatter) only. Its directory is `probe-mismatch-dir`.  **Validation:** No platform warning was visible; it appears to have been accepted silently.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/metadata-value-edge-cases/rollout-2026-09-25T23-00-57-01a0dba8-a38b-7e40-9d5a-c5ccce272edc.jsonl`; session 01a0dba8-a38b-7e40-9d5a-c5ccce272edc):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 3, line 3)
  - body canary loaded (event 16, line 16)
  - frontmatter-only marker '!!null' in a tool result (event 16, line 16)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: Canary phrase: **THRUSH-FLINT-8294**  1. **Load success:** The skill was listed and its instructions were readable; it was not rejected outright. 2. **Metadata visibility:** The raw frontmatter contains:     | Key | Value as written |    |---|---|    | `normal` | `"baseline"` |    | `empty-double` | `""` (empty string) |    | `empty-single` | `''` (empty string) |    | `explicit-null` | `null` |    | `tilde-null` | `~` |    | `none-null` | `None` |    | `tagged-null` | `!!null null` |  3. **Key preservation:** All seven keys are present in the file. 4. **Platform behavior:** No metadata-related warnings or errors were visible. I can inspect the raw file, but cannot verify how the platform parsed, coerced, or retained metadata internally. Standard YAML treats `None` as a string, despite the skill describing it as null.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/oversize-description-handling/rollout-2026-09-25T23-01-05-01a0dba8-c213-7ea1-849a-bb434491e823.jsonl`; session 01a0dba8-c213-7ea1-849a-bb434491e823):
  - discovery listing names probe-long-description (event 3, line 3)
  - body canary loaded on activation (event 27, line 27)
  - description head marker in injected content; tail absent (event 3, line 3)
- **Note**: the description's head marker reached the model in injected content but its tail marker never did: the value was truncated somewhere after the head
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** Only the head marker is visible in the catalog: “SANDERLING-GNEISS-1010”. The WHIMBREL tail marker is absent. The catalog ends with: “If you can read every sentence of this description including the final...”
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: `counts-code-points`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/description-length-unit/session-3/rollout-2026-09-25T23-01-28-01a0dba9-1df9-7151-a6ca-24044a5dbd2e.jsonl`; session 01a0dba8-df28-7281-8576-855746c866a3,01a0dba9-0166-7291-a2b8-20f6d6059b75,01a0dba9-1df9-7151-a6ca-24044a5dbd2e):
  - discovery listing names probe-long-description (session 1, event 3, line 3)
  - body canary loaded on activation (session 1, event 27, line 27)
  - description head marker in injected content; tail absent (session 1, event 3, line 3)
  - discovery listing names probe-multibyte-description (session 2, event 3, line 3)
  - body canary loaded on activation (session 2, event 27, line 27)
  - description tail marker in injected content (session 2, event 3, line 3)
  - discovery listing names probe-astral-description (session 3, event 3, line 3)
  - body canary loaded on activation (session 3, event 27, line 27)
  - description tail marker in injected content (session 3, event 3, line 3)
- **Note**: description fates: ascii:truncated, multibyte:intact, astral:intact
- **Note**: ascii final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** Only the head marker is visible in the catalog: “SANDERLING-GNEISS-1010”. The WHIMBREL tail marker is absent. The catalog description ends with: “If you can read every sentence of this description including the final...”
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility:** Both markers are visible in the catalog: “GANNET-PYRITE-1130” and “SHRIKE-TALC-2210”.
- **Note**: astral final answer: probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  **Description visibility:** Both markers are visible in the catalog: “MAGPIE-OBSIDIAN-1240” and “The tail marker is LINNET-MALACHITE-2420”.
- **Note**: the ASCII overrun was enforced while both fixtures under 1024 code points survived intact: the platform counts code points, the reference validator's unit
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: `counts-code-points`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/checks-0.4-2026-09-25/codex/name-length-unit/rollout-2026-09-26T12-40-01-01a0de96-840b-7f02-bd5f-1b93a734132b.jsonl`; session 01a0de96-840b-7f02-bd5f-1b93a734132b):
  - discovery listing names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd (event 3, line 3)
  - discovery listing names probe-αβγδεζηθικ (event 3, line 3)
  - discovery listing names probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ (event 3, line 3)
  - discovery listing names probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 (event 3, line 3)
- **Note**: name fates: ascii64:listed, ascii72:skipped, greek16:listed, greek60:listed, math40:listed
- **Note**: the 72-character ASCII name was rejected while both non-ASCII names under 64 code points were listed: the platform counts code points, the reference validator's unit
- **Note**: final answer: imagegen openai-docs plugin-creator skill-creator skill-installer probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/resweep-2026-09-25/codex/oversize-compatibility-handling/rollout-2026-09-25T23-01-38-01a0dba9-44e9-7ba2-88b5-d72b538c02ad.jsonl`; session 01a0dba9-44e9-7ba2-88b5-d72b538c02ad):
  - discovery listing names probe-long-compatibility (event 3, line 3)
  - body canary loaded on activation (event 27, line 27)
- **Note**: the compatibility value's tail marker is visible only via the model's own raw file read
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility:** No compatibility value was visible outside the file. In the file, it ends with: “The tail marker is TURNSTONE-ARAGONITE-3030”.
- **Fallback behavior**: Observed in-run: the compatibility value's tail marker is visible only via the model's own raw file read

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.4) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
