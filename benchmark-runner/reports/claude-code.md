# Platform Loading Implementation: Claude Code (headless)

| | |
|---|---|
| **Platform** | Claude Code (headless) |
| **Platform version** | 2.1.212 |
| **Check list version** | 0.2 |
| **Test date** | 2026-08-01 |
| **Model(s) observed** | claude-fable-5, claude-sonnet-5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Evidence line numbers cite the archived transcripts in the results directories. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 18 of the 40 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

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


## All checks

The full finding for every check in the list, grouped by category.

### Loading Timing

#### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: `metadata-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/discovery-reading-depth/c3c41d11-96d7-45b8-938d-c5eadce4dde2.jsonl`; session c3c41d11-96d7-45b8-938d-c5eadce4dde2):
  - discovery listing names probe-loading (event 4, line 6)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?_

- **Status**: observed
- **Verdict**: `body-only`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/activation-loading-scope/da80528f-0456-4d35-8eb3-c4ddf661adbc.jsonl`; session da80528f-0456-4d35-8eb3-c4ddf661adbc):
  - body canary in harness-injected content (event 7, line 10)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: `no-prefetch`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/eager-link-resolution/58b5b9e7-50b5-40a4-a20c-794287b86333.jsonl`; session 58b5b9e7-50b5-40a4-a20c-794287b86333):
  - skill body loaded (event 8, line 11)
  - references/setup-guide.md arrived only via the model's own read (event 21, line 24)
  - references/troubleshooting.md arrived only via the model's own read (event 23, line 26)
  - references/unlinked-data.md arrived only via the model's own read (event 25, line 28)
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
- **Verdict**: `untouched`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: ## Report: Nonstandard Directories Probe  **1. Directory awareness**  When the skill activated, I was only given the `SKILL.md` body text (injected as the tool result) plus a "Base directory" path. No directory listing or file enumeration was provided automatically — I was not told what was inside `evals/`, `templates/`, or `resources/` until I ran `find` myself via Bash. After that manual step, I became aware of:  - `evals/evals.json` - `resources/api-reference.md` - `templates/output-template.md` - `SKILL.md` (given directly)  **2. Content visibility**  No file contents were pre-loaded or auto-surfaced for any of these three directories. I can see filenames (once I explicitly list them), but not contents — I would need to explicitly `Read` each file to see what's inside `evals.json`, `api-reference.md`, or `output-template.md`. Nothing was silently loaded into context the way `SKILL.md` itself was.  **3. Resources vs. references**  I observed no special handling of `resources/` at all — it was not enumerated, previewed, or treated differently from `templates/` or `evals/`. All three nonstandard directories were treated identically: invisible until I searched for them manually, an…
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
- **Runs**: 3 (cwd-base-model-requalified ×3)
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/path-resolution-base/run-01/fd8e9690-ce0c-4580-92c9-1c0575e50e4c.jsonl`; session fd8e9690-ce0c-4580-92c9-1c0575e50e4c):
  - bare SKILL.md-relative path attempt did not deliver content (event 11, line 14)
  - bare SKILL.md-relative path attempt did not deliver content (event 13, line 16)
- **Note**: the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory
- **Note**: verdict consistent across 3 runs
- **Fallback behavior**: agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.

#### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: `own-resource-first`
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/cross-skill-resource-shadowing/791c48c2-eda3-47f1-9f3c-95f09198755c.jsonl`; session 791c48c2-eda3-47f1-9f3c-95f09198755c):
  - activated skill's own API.md content arrived first (event 11, line 13)
  - sibling skill's API.md also read (event 15, line 17)
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation
- **Note**: the sibling skill's version was also accessible later in the session (no isolation between skills' resources)
- **Note**: final answer: ## Results  1. **Alpha skill's own file** (`.claude/skills/probe-shadow-alpha/references/API.md`): canary phrase **STORK-CORAL-4471** — correctly resolved to the alpha version.  2. **Beta skill's file** (`.claude/skills/probe-shadow-beta/references/API.md`): I *was* able to access it directly by specifying its full absolute path, and it correctly returned its own distinct content with canary **EGRET-SLATE-8823** — no cross-contamination.  **Conclusion**: no ambiguous resolution occurred. Each skill's `references/API.md` resolves to its own distinct absolute path (`.claude/skills/probe-shadow-alpha/references/API.md` vs. `.claude/skills/probe-shadow-beta/references/API.md`), and the Read tool correctly retrieved the right content for each when given the right path. There's no shared/ambiguous namespace causing one skill's resource lookup to silently pull in the other's file — accessing the sibling skill's file required explicitly using its full distinct path, not an implicit relative lookup from within alpha's context.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: `outside-skill-read-allowed`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/path-traversal-boundary/c2a4d60d-7580-4740-a882-4aedec90e0a6.jsonl`; session c2a4d60d-7580-4740-a882-4aedec90e0a6):
  - sibling skill's body canary arrived in a tool result (event 12, line 15)
  - first ../ path attempt (event 11, line 14)
- **Note**: relative ../ path forms were used
- **Note**: final answer: Now I have enough information to compile the report.  ## Path Traversal Probe Results  **1. Sibling skill (`../probe-loading/SKILL.md`)** - **Allowed — succeeded.** The Read tool resolved the `..` traversal and returned the full contents of the sibling skill's `SKILL.md` (including its canary phrase `CARDINAL-ZEBRA-7742` and full instructions). No warning or error was shown. This means file-level tools like Read do **not** enforce a boundary at the skill directory root — they'll follow `..` to sibling directories.  **2. Parent directory (`../README.md` and listing `../`)** - **Read**: returned "File does not exist" — but this is because no `README.md` actually exists there (confirmed by listing `.claude/skills/` directly, which shows only `probe-loading` and `probe-traversal`), not because of a boundary block. - **Listing via `ls ../`**: **Blocked**, with an explicit security error: *"Path contains '..' traversal after a directory segment, which may follow a symlink outside the working directory."* This is the Bash tool actively refusing the traversal syntax itself, regardless of whether the target exists.  **3. Two levels up (`../../loading-behavior.md`)** - **Read**: returned "Fi…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `resource-nesting-depth`

_How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels._

- **Status**: observed
- **Verdict**: `all-depths-accessible-through-5`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/resource-nesting-depth/bf0e1a46-8b60-471c-92a2-8ca327eaa4c9.jsonl`; session bf0e1a46-8b60-471c-92a2-8ca327eaa4c9):
  - depth-1 file references/overview.md content arrived (event 12, line 14)
  - depth-2 file references/api/endpoints.md content arrived (event 14, line 16)
  - depth-3 file references/api/v2/migration-guide.md content arrived (event 16, line 18)
  - depth-3 file references/guides/advanced/performance-tuning.md content arrived (event 18, line 20)
  - depth-5 file references/api/v2/history/deprecated/removed-endpoints.md content arrived (event 20, line 22)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-script-execution`

_Can the agent run a bundled scripts/ file and receive its output?_

- **Status**: observed
- **Verdict**: `execution-blocked-visibly`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/claude-code/bundled-script-execution/66d895e0-8a07-4c2f-834d-489e0eb3f73e.jsonl`; session 66d895e0-8a07-4c2f-834d-489e0eb3f73e):
  - script-referencing tool call returned an error result (event 11, line 13)
- **Note**: final answer: Script execution was blocked. Result of the probe:  - **Skill activation:** succeeded. Canary: **REDSHANK-SYENITE-8807** - **Script execution:** failed. Both attempts to run `scripts/emit-canary.sh` (via `bash <path>` and via direct `./emit-canary.sh`) were denied by the permission system. The exact errors were:   - `This command requires approval`   - `This Bash command contains multiple operations. The following part requires approval: ./emit-canary.sh`  Because this session is non-interactive, there is no way to grant the approval, so the script never ran and I cannot report its assembled phrase (the skill forbids guessing it). The platform in this configuration does **not** let the agent execute bundled skill scripts without explicit approval.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Content Presentation

#### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: `name-and-description-only`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/discovery-listing-fields/1e08a3fe-f6a0-4223-9281-1e8354975fb5.jsonl`; session 1e08a3fe-f6a0-4223-9281-1e8354975fb5):
  - discovery listing names the installed skills (event 4, line 6)
  - description text surfaced at discovery (event 4, line 6)
- **Note**: final answer: Here is everything shown to me in the Available skills list, reproduced verbatim:  - **probe-compatibility**: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior. - **probe-loading**: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior. - **probe-metadata-values**: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling. - **deep-research**: Deep research harness — fan-out web searches, fetch sources, adversarially verify claims, synthesize a cited report. - When the user wants a deep, multi-source, fact-checked research report on any topic. BEFORE invoking, check if the question is specific enough to research directly — if underspecified (e.g., "what car to buy" without budget/use-case/region), ask 2-3 clarifying questions to narrow scope. Then pass the refined question as args, weaving the answers in. - **dataviz**: Use this skill whenever you are about to create ANY chart, graph, plot, dashboard, or data visualizati…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: `frontmatter-stripped-on-injection`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/frontmatter-handling/db3486a9-eb3a-493a-8e82-3b56866c96b2.jsonl`; session db3486a9-eb3a-493a-8e82-3b56866c96b2):
  - body injected without the frontmatter-only marker (event 7, line 10)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: `raw-injection`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/content-wrapping-format/84a0b01a-da77-4f1b-80f7-b90fddc6130d.jsonl`; session 84a0b01a-da77-4f1b-80f7-b90fddc6130d):
  - injection event carrying the body canary (event 7, line 10)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: `reinjected-each-activation`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/reactivation-deduplication/b4d5a4a7-93d6-46ab-acbf-86dd5a3e705e.jsonl`; session b4d5a4a7-93d6-46ab-acbf-86dd5a3e705e):
  - first injection of body canary (event 7, line 10)
  - second injection of body canary (event 16, line 21)
- **Note**: platform-level: duplicate skill content occupies context after reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: `fresh-content-served`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/reactivation-freshness/99992428-c5f7-4a08-8700-500476df33ac.jsonl`; session 99992428-c5f7-4a08-8700-500476df33ac):
  - post-edit canary reached the model on reactivation (event 15, line 20)
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
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/compatibility-field-behavior/8140dc4e-af36-4c7b-af88-b9d559d0d9a2.jsonl`; session 8140dc4e-af36-4c7b-af88-b9d559d0d9a2):
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value (event 4, line 6)
  - body loaded normally (event 8, line 11)
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: ## Probe Results  **1. Field visibility:** Yes — I can see it directly in the frontmatter by reading the file. It was not, however, surfaced to me automatically in the skill-invocation content the harness injected (that only contained the body text, not the frontmatter fields). I had to `cat` the file myself to find: ``` compatibility: Designed for Claude Code (or similar products). Requires Python 3.14+ and network access. ```  **2. Platform behavior:** No special handling observed. The skill activated normally — no warning, no confirmation prompt, and no automated check of whether Python 3.14+ or network access is actually available. The `compatibility` field appears to be inert metadata as far as the harness is concerned; nothing gated activation on it.  **3. My interpretation:** I can't verify this from inside the sandbox with confidence: - "Claude Code (or similar products)" — plausibly satisfied, since this session is running in a Claude-Code-like harness. - "Python 3.14+" — unverified and not obviously relevant to this skill (the skill's own instructions involve no Python execution). I could check `python3 --version` on the host, but that wouldn't confirm what environment th…
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: `executed-regardless-of-field`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/claude-code/allowed-tools-behavior/session-2/ab356434-5430-408e-acc7-c92746edca5b.jsonl`; session 2f2b6b0d-38a5-42f8-bacd-24e4c67a91c9,ab356434-5430-408e-acc7-c92746edca5b):
  - with-field session: runtime-assembled phrase arrived in a tool result (session 1, event 13, line 15)
  - control session: runtime-assembled phrase arrived in a tool result (session 2, event 13, line 16)
- **Note**: allowed-tools value never reached the model
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: Here are the probe results:  1. **Activation confirmed.** probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. **allowed-tools visibility:** I cannot see any `allowed-tools` value for this skill. Only the skill body was loaded into my context; no frontmatter fields were surfaced. So while the body claims the frontmatter declares a printf-scoped Bash pattern plus Read, that literal value never appeared anywhere in what I received.  3. **Command output, verbatim:**    ```    GROUSE-MICA-9017    ```  4. **Permission observation:** The command ran immediately and returned its output with no permission prompt, approval step, or error visible to me. From my side I can't distinguish whether that's because an `allowed-tools` pre-approval took effect or because the session's general permission posture allowed the command; comparing against `probe-allowed-tools-control` (which lacks the field) would be needed to tell those apart.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Command output, verbatim:  ``` LAPWING-FLUORITE-2260 ```  The command ran without any permission prompt or approval step that I could observe: the Bash tool call returned the output immediately, with no denial, no hook interception, and no indication of a user approval gate. Since this control skill declares no `allowed-tools` field, that means the platform's general permission posture (session permission mode/allowlist) permitted the command on its own.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill-to-Skill Invocation

#### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/cross-skill-invocation/faf458a5-bb16-4b08-a270-8263e82dcf52.jsonl`; session faf458a5-bb16-4b08-a270-8263e82dcf52):
  - invoke-alpha body loaded (chain entry) (event 8, line 11)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 13, line 16)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: The chain breaks here: `invoke-gamma` is not an available skill (only `invoke-alpha` and `invoke-beta` were listed to me), so it cannot be activated.  invoke-beta complete. invoke-gamma said: N/A — invocation failed with "Unknown skill: invoke-gamma."  invoke-alpha complete. Chain result: invoke-alpha (canary IBIS-RUST-3310) successfully activated invoke-beta (canary TERN-MOSS-6647), but invoke-beta's attempt to activate invoke-gamma failed because that skill isn't registered/available in this environment. The chain terminated at the beta→gamma link.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/invocation-depth-limit/f1f1607f-4fc2-4de2-8669-a040c6413090.jsonl`; session f1f1607f-4fc2-4de2-8669-a040c6413090):
  - invoke-alpha body loaded (event 8, line 11)
  - invoke-beta body loaded (event 13, line 16)
  - invoke-gamma body loaded (event 18, line 21)
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "invoke-gamma activated. Canary: JAY-TEAL-9984" and confirmed the end of chain was reached, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), then activated invoke-gamma, which activated (Canary: JAY-TEAL-9984) and confirmed the full chain — invoke-alpha → invoke-beta → invoke-gamma — completed successfully. All three canaries were reported in sequence: IBIS-RUST-3310, TERN-MOSS-6647, JAY-TEAL-9984.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: `cycle-stopped-model-choice`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Beta's instructions direct me to activate alpha again, which would recreate the exact same cycle indefinitely (alpha → beta → alpha → beta → ...). Continuing to follow this literally would never terminate, so I'm breaking the loop here rather than re-invoking alpha a second time.  **Summary of the probe:** - probe-circular-alpha activated successfully, canary `KITE-ONYX-2251` emitted. - It invoked probe-circular-beta, which activated successfully, canary `WREN-SLATE-7738` emitted. - Beta's instructions call for re-activating alpha, closing the loop. - No platform-level guard fired to block the second invocation (the Skill tool doesn't appear to detect cycles automatically) — the loop was only stopped because I judged that mechanically re-entering it would run forever with no new information gained.  So: circular invocation is *possible* (nothing stopped step 3 in either skill from executing), and detection/prevention is currently a matter of agent judgment rather than a platform safeguard.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: `chain-completed-depth-3`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/invocation-language-sensitivity/a5079d7f-61c2-442b-95f9-67deedbcdc76.jsonl`; session a5079d7f-61c2-442b-95f9-67deedbcdc76):
  - invoke-alpha body loaded (event 8, line 11)
  - invoke-beta body loaded (event 13, line 16)
  - invoke-gamma body loaded (event 18, line 21)
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: "invoke-gamma activated. Canary: JAY-TEAL-9984" and confirmed the end of chain was reached, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (Canary: TERN-MOSS-6647), which activated invoke-gamma (Canary: JAY-TEAL-9984). Gamma confirmed the end of the chain was reached and all three skills activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Skill Dependencies

#### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: `second-skill-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/informal-dependency-resolution/81d2e849-11d3-4eb9-a534-bd8c04919481.jsonl`; session 81d2e849-11d3-4eb9-a534-bd8c04919481):
  - invoke-alpha body loaded (chain entry) (event 8, line 11)
  - invoke-beta's body canary loaded after invoke-alpha's instruction (event 13, line 16)
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-gamma isn't registered as an available skill, so it can't be activated — the call returned "Unknown skill: invoke-gamma." The chain stops here.  invoke-beta complete. invoke-gamma could not be activated (unknown/unregistered skill), so no result to report from it.  invoke-alpha complete. Chain result: invoke-beta activated successfully (canary TERN-MOSS-6647), but invoke-gamma is not available in this environment's skill registry, so the chain broke at the second link — invoke-gamma never ran.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: `attempted-visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/missing-dependency-behavior/3df44faf-c337-49eb-a2f3-881409fc4665.jsonl`; session 3df44faf-c337-49eb-a2f3-881409fc4665):
  - probe-missing-dep body loaded (event 8, line 11)
  - attempt on the missing skill returned an error result (event 12, line 15)
- **Note**: final answer: The platform explicitly reported that the skill doesn't exist — it returned a tool error (`Unknown skill: nonexistent-formatter`) rather than silently skipping or letting me improvise the formatting from general knowledge. I did not attempt to fulfill the formatting task myself.  probe-missing-dep complete. Dependency resolution result: the harness rejected the activation attempt at the tool-call level with an explicit "Unknown skill" error, since `nonexistent-formatter` isn't in the available skills list — no silent failure, no fallback to general knowledge.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: `fields-ignored`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/nonstandard-dependency-fields/b7f8d41a-692e-455d-8b9d-0604bcc3a493.jsonl`; session b7f8d41a-692e-455d-8b9d-0604bcc3a493):
  - probe-nonstandard-fields body loaded (event 7, line 10)
- **Note**: final answer: ## Report: Nonstandard Frontmatter Fields Probe  **1. Field visibility** — Yes, by reading the SKILL.md file directly I can see all three nonstandard fields and their values: - `requires: probe-loading` - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` - `priority: high`  However, these fields were **not** surfaced to me through the normal skill-invocation channel. The `<system-reminder>` listing of available skills only exposed the standard `name` and `description` fields (as a one-line summary per skill). I only saw `requires`, `depends-on`, and `priority` because I explicitly opened the file with the Read tool — the skill-loading mechanism itself didn't pass them into context.  **2. Platform behavior** — None of the nonstandard fields triggered any automatic action: - `requires: probe-loading` did **not** cause `probe-loading` to be auto-loaded or activated. - `depends-on: [probe-shadow-alpha, probe-shadow-beta]` did **not** cause either shadow skill to be loaded. - `priority: high` had no observable effect on ordering or how this skill was presented relative to others in the skill list.  All three fields were silently ignored by the platform's skill-loading logic.  **3. V…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: `resolved-across-scopes; missing:visible-failure`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/cross-scope-dependency/session-2/e037f8ae-247d-4674-9a5c-ec18ee1b0579.jsonl`; session fcc5c291-9d96-4d70-908c-f8c9bc21b994,e037f8ae-247d-4674-9a5c-ec18ee1b0579):
  - user-scope dependency's body canary injected during project-skill session (event 13, line 16)
  - attempt on the missing dependency returned an error result (session 2, event 12, line 15)
- **Note**: with-dependency final answer: `probe-loading` was found and activated successfully — resolved to a **user-level** location: `/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-671745375/home/claude/skills/probe-loading`, distinct from the project-level path (`..…
- **Note**: without-dependency final answer: `probe-loading` was not found. The Skill tool rejected it outright with `Unknown skill: probe-loading` — it isn't present in the available-skills listing at all (neither project nor user scope), so the platform didn't even attempt a scope-specific lookup; it just failed fast as an unrecognized skill name.  probe-cross-scope complete. Cross-scope resolution result: the project-level skill's reference to a user-level skill did **not** resolve — `probe-loading` is absent from the resolved skill set entirely, and invoking it errors immediately with "Unknown skill" rather than finding it at a different scope or giving a scope-specific not-installed message. This suggests either no user-level skill of that name exists in this environment, or the platform does not merge user-level skills into the resolvable set alongside project-level ones (at least not visibly to the model).
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: `convention-not-scanned`
- **Confidence**: transcript-direct
- **Note**: a skill installed by another client at .agents/skills is invisible here
- **Note**: final answer: I don't have a skill named "probe-interop" in my available skills list — it's not one of the skills shown to me (deep-research, dataviz, update-config, keybindings-help, verify, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, run, init, review, security-review). I can't invoke skills that aren't in that list, and I won't guess at or fabricate one. Let me know if you meant one of the above, or if this should be added to the project first.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: `direct-children-only; stray:not-discovered`
- **Confidence**: transcript-direct
- **Note**: final answer: I don't have a skill called "probe-grouped" in my available skills list — I can't invoke skills that aren't listed. Here's what's actually available to me:  deep-research, dataviz, update-config, keybindings-help, verify, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, run, init, review, security-review  If you meant one of these, let me know. Otherwise, if "probe-grouped" is something you expect to exist (e.g., a custom project skill), it may not be registered or discoverable in this session.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: `nested-skill-not-discovered`
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/nested-skill-discovery/180eb052-18f9-48e4-a298-731a15ff735d.jsonl`; session 180eb052-18f9-48e4-a298-731a15ff735d):
  - discovery listing names only the outer skill (event 4, line 6)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: `user-overrides-project`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/name-collision-precedence/30a04ffa-62f5-4226-b019-c56e056de93a.jsonl`; session 30a04ffa-62f5-4226-b019-c56e056de93a):
  - user variant's canary loaded; project variant's never appeared (event 8, line 11)
- **Note**: contradicts the guide's 'universal convention' that project-level overrides user-level
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **PIPIT-SHALE-1147** (user variant)  **Precedence**: The **user**-scope variant won over the project-scope variant.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `tolerated-and-loaded`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/malformed-yaml-tolerance/abb82fae-9635-4726-9cbb-86b39e1cbeaf.jsonl`; session abb82fae-9635-4726-9cbb-86b39e1cbeaf):
  - discovery listing names probe-malformed-yaml (event 4, line 6)
  - body canary loaded on activation (event 10, line 16)
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility**: In my available-skills catalog, this skill was listed as:  `probe-malformed-yaml: Use when: asked to probe malformed yaml parsing behavior`  The description appeared intact and readable — not truncated at the colon, and I have no way to tell from my view whether the underlying YAML was silently repaired (e.g., auto-quoted) before reaching me, since I only see the final rendered text, not the raw frontmatter source.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: `loaded-despite-missing-description`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/missing-description-handling/cf54321c-0a3c-4754-b74e-d2bb767ec395.jsonl`; session cf54321c-0a3c-4754-b74e-d2bb767ec395):
  - discovery listing names probe-no-description (event 4, line 6)
  - body canary loaded on activation (event 9, line 15)
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry**: Yes, it appears in my available-skills list, with an empty/placeholder description. The entry showed only the title text "Missing Description Probe" — the same string used as the skill's markdown heading in this file — with no actual description content synthesized or written alongside it, unlike every other skill in the list which has a substantive one-line description explaining when to use it.  For reference, this behavior contradicts what the skill's own text claims the client implementation guide recommends (skipping skills without a description and logging an error) — I loaded and activated it anyway rather than skipping it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: `all-invalid-names-tolerated`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/claude-code/invalid-name-tolerance/e868bb32-f88d-43b6-baf3-5ee0af433a4e.jsonl`; session e868bb32-f88d-43b6-baf3-5ee0af433a4e):
  - discovery listing names probe-Upper-Case (event 4, line 6)
  - discovery listing names probe--double-hyphen (event 4, line 6)
  - discovery listing names probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit (event 4, line 6)
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  **Catalog identity**: The skill appears in my catalog under its full 72-character name: `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`. It was not truncated or rejected for exceeding the 64-character limit, and invoking it by the complete name succeeded.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: `directory-name-identity`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/name-directory-mismatch/4e898c10-e554-4ba6-a5f6-93c6723588d2.jsonl`; session 4e898c10-e554-4ba6-a5f6-93c6723588d2):
  - discovery listing carries the DIRECTORY name probe-mismatch-dir (event 4, line 6)
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: final answer: probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  1. **Identity**: Same result as before — the catalog lists it as `probe-mismatch-dir` (directory name), even though the frontmatter declares `probe-name-mismatch`. Both invocation names (`probe-name-mismatch` and `probe-mismatch-dir`) successfully resolved to this same skill.  2. **Validation**: Again, accepted silently — no warning or error about the mismatch, and both the frontmatter name and directory name work as valid identifiers to activate it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?_

- **Status**: observed
- **Verdict**: `loaded-despite-edge-case-metadata`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/sandboxed-2026-08-01/claude-code/metadata-value-edge-cases/8b469579-2654-4953-a484-83ee2150b28f.jsonl`; session 8b469579-2654-4953-a484-83ee2150b28f):
  - discovery listing names probe-metadata-values despite edge-case metadata (event 4, line 6)
  - body canary loaded (event 8, line 11)
  - frontmatter-only marker '!!null' in a tool result (event 13, line 15)
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: ## Probe Results  **1. Load success:** Yes. The skill loaded and its instructions rendered fine, including the canary phrase **THRUSH-FLINT-8294**. The edge-case `metadata` values did not cause a rejection.  **2. Metadata visibility:** The `metadata` field was **not surfaced to me anywhere in the Skill tool output**. When I invoked the skill, I only received the body content (everything from `# Metadata Value Edge Cases Probe` onward) — no frontmatter, and no `metadata` block. The earlier skill-listing system-reminder also only exposed `name` and a one-line description, not `metadata`.  The only way I could see the raw values at all was by directly reading the file with the `Read` tool, bypassing the skill-loading mechanism entirely. Via `Read`, the raw YAML shows:  | Key | Raw YAML |  |---|---| | `normal` | `"baseline"` | | `empty-double` | `""` | | `empty-single` | `''` | | `explicit-null` | `null` | | `tilde-null` | `~` | | `none-null` | `None` (this is just the literal string `"None"` in YAML — not a null token) | | `tagged-null` | `!!null null` |  **3. Key preservation:** All seven keys are present in the source file. But since the platform's Skill-invocation path never expose…
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-description`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/claude-code/oversize-description-handling/503eb78a-af2e-4e53-91fd-4e655f70c7bb.jsonl`; session 503eb78a-af2e-4e53-91fd-4e655f70c7bb):
  - discovery listing names probe-long-description (event 4, line 6)
  - body canary loaded on activation (event 11, line 17)
  - description tail marker in injected content (event 4, line 6)
- **Note**: the description's tail marker reached the model in harness-injected content: the oversize value survived past 1024 characters untruncated
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility**: Both markers are visible in my catalog entry, meaning the over-limit description survived intact end to end (no rejection, no truncation).  - Head marker, quoted from the catalog: "The head marker SANDERLING-GNEISS-1010 sits near the start of this description and a matching tail marker sits at the very end..." - Tail marker, quoted from the catalog (the entry's final characters): "...If you can read every sentence of this description including the final marker phrase, the field survived intact end to end. The tail marker is WHIMBREL-DOLOMITE-2020"  The full padding prose between the markers is also present in the catalog entry, so the platform accepted the ~1100+ character description without enforcing the spec's 1024-character limit.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `oversize-compatibility-handling`

_Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?_

- **Status**: observed
- **Verdict**: `loaded-despite-oversize-compatibility`
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence** (transcript: `results/batch4-2026-08-01/claude-code/oversize-compatibility-handling/d01c79b0-a513-4a13-ad00-ef4b83983e40.jsonl`; session d01c79b0-a513-4a13-ad00-ef4b83983e40):
  - discovery listing names probe-long-compatibility (event 4, line 6)
  - body canary loaded on activation (event 11, line 17)
- **Note**: the compatibility value's tail marker never reached the model
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No. The only place this skill surfaces in my context outside its SKILL.md file is the available-skills listing in my system prompt, and that entry contains only the name and description. It reads exactly:  > `- probe-long-compatibility: Benchmark skill whose compatibility field runs past the spec's 500-character limit. Use when asked to probe oversize compatibility handling.`  No compatibility value is exposed to me there or anywhere else, so I cannot see the TURNSTONE tail marker (or any portion of the compatibility field) at all. The field appears to be either stripped or simply not surfaced by this harness, rather than truncated: truncation would show a partial value, but I see none.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

---

Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list 0.2) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.
