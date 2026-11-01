---
title: "Claude Code (headless)"
description: "Automated skill loading findings for Claude Code (headless) (check list 0.4)."
date: 2026-11-01
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | Claude Code (headless) |
| **Platform version** | 2.1.309 |
| **Check list version** | 0.4 |
| **Test date** | 2026-11-01 |
| **Model(s) observed** | claude-sonnet-5-5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists. These findings describe the release in the header on the test date; platform behavior changes between releases, and the [changelog](/changelog/) records what moved from one run of the checks to the next.

> **Changed since the previous observation**: 3 verdicts differ from the previous run of the checks on this platform: [`name-directory-mismatch`](#name-directory-mismatch), [`path-resolution-base`](#path-resolution-base), [`unrecognized-directory-handling`](#unrecognized-directory-handling). Each finding below states the earlier verdict and release.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 21 of the 46 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

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
- [`name-directory-mismatch`](#name-directory-mismatch): The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity. The platform loaded it anyway and listed it under both identities.
- [`metadata-value-edge-cases`](#metadata-value-edge-cases): The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it.
- [`oversize-description-handling`](#oversize-description-handling): The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated.
- [`oversize-compatibility-handling`](#oversize-compatibility-handling): The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway.

### Not exercised in this run

- [`path-resolution-base`](#path-resolution-base): The spec's skill-root-relative paths went untested in this run: the model rewrote each reference to a fully qualified path before use, so the platform's own resolution base was never exercised. (Behavioral inference.)
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
- **Verdict**: Not surfaced; model never looked (`untouched`)
- **Changed since previous observation**: was Readable when the model looks (`readable-on-demand:[evals/ templates/]`) on 2.1.274 (2026-09-26); the [changelog](/changelog/) says what moved it
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: **1. Directory awareness**  The platform gave me only the skill's base directory path and the text of `SKILL.md`. It did not enumerate any subdirectories. I know about `evals/`, `templates/` and `resources/` only because `SKILL.md` names them. I then ran `ls -R` myself, which confirmed all three exist and that there are no others (`scripts/`, `references/` and `assets/` are absent).  **2. Content visibility**  - **Without running anything:** I could not see any file listing or file contents. The skill load injected none of them. - **After my own `ls -R`:** I could see these files:   - `evals/evals.json`   - `templates/output-template.md`   - `resources/api-reference.md` - **Contents:** I have not read any of the three files. The platform did not preload them.  **3. Resources vs. references**  `resources/` was treated the same as `evals/` and `templates/`. It was not enumerated, loaded or specially handled. The platform ignored it and gave me the base path, so the files are reachable only if I read them myself. I can't say how `references/` would be treated, because this skill has no such directory and I have nothing to compare against. In this run, no directory got special treatmen…
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
- **Verdict**: Model used full paths (base untested) (`model-preemptively-qualified`)
- **Changed since previous observation**: was Bare path fails; model recovers (`cwd-base-model-requalified`) on 2.1.274 (2026-09-26); the [changelog](/changelog/) says what moved it
- **Vehicle**: harness-push
- **Confidence**: behavioral-inference
- **Note**: the model never tried the bare path, so the platform's resolution base was not directly exercised; it navigated by qualified path from the start
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

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
- **Note**: final answer: **Alpha's `references/API.md`** (from `probe-shadow-alpha`):  - Title: API Reference (Alpha) - Canary phrase: **STORK-CORAL-4471**, so this is the alpha version. - Endpoints:   - `GET /alpha/status`: alpha health check   - `POST /alpha/process`: alpha processing  **Beta's version:** I could read it by its full path, `.claude/skills/probe-shadow-beta/references/API.md`. - Title: API Reference (Beta) - Canary phrase: **EGRET-SLATE-8823** - Endpoints:   - `GET /beta/status`   - `POST /beta/process`  **Result:** There was no shadowing. Each skill's `references/API.md` resolved to its own content when I read it by explicit path. I didn't test a bare relative path like `references/API.md`, so I can't say how that would resolve.
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
- **Note**: final answer: Results of the three traversal attempts:  1. **Sibling skill** (`../probe-loading/SKILL.md`): **Allowed.** The read succeeded and returned the full file, frontmatter and body included. There was no error, warning, or block.  2. **Parent directory** (`../README.md` and `../`):    - `../README.md` doesn't exist. The read returned "File does not exist", so this was a plain not-found error and not a block.    - Listing `../` with `ls` worked. It shows two entries, `probe-loading` and `probe-traversal`.  3. **Two levels up** (`../../checks.md`): This resolves to `.claude/checks.md`, which doesn't exist. The read returned "File does not exist", so again not-found and not a block.  **Summary:** I saw no boundary enforcement at the skill directory root. The platform didn't block, warn about, or silently empty any of these reads. The two failures happened only because the files aren't there. The sibling read and the directory listing both succeeded. The benchmark may expect `README.md` and `checks.md` to exist in a different layout, so I can't tell from this run whether a file that did exist outside the skill would have been readable. The sibling read suggests it would.
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
- **Note**: final answer: The script didn't run. The Bash call returned "This command requires approval", so the command was blocked and I have no output from it. I won't guess the phrase.  The call was `bash <skill dir>/scripts/emit-canary.sh`, using the skill's own directory. If you approve it, I can run it again and report exactly what it prints.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: final answer: 1. probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  2. **File awareness:** I was told about no files belonging to this skill. The activation gave me only the skill's base directory and its instructions. It had no file listing. I did not run any command or read the directory.  3. **Count:** 0
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
- **Note**: final answer: Each entry below shows a name and a description. No other fields or values are shown for any skill.  - **probe-compatibility**: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior.  - **probe-loading**: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior.  - **probe-metadata-values**: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling.  - **dataviz**: Use this skill whenever you are about to create ANY chart, graph, plot, dashboard, or data visualization, in ANY output medium — an HTML or React artifact, inline SVG, plotting code in any library (matplotlib, plotly, d3, Recharts, …), an image/PNG you will render and upload, or a chart shared into Slack. Read it BEFORE writing the first line of chart code, choosing chart colors, building a stat tile / meter / KPI row, or laying out a dashboard. When the destination is a first-party document connector (host-designated, never self-described) that renders live charts, …
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
- **Note**: final answer: **Probe report**  1. **Frontmatter visibility:** No. I can't see any YAML frontmatter. The skill content begins at the "Base directory for this skill" line and then the "# Loading Behavior Probe" heading. The frontmatter was stripped, so I can't see name, description, allowed-tools, compatibility or metadata. I know the skill's name and description only from the available-skills listing in the system prompt.  2. **Content wrapping:** None. The body came as plain markdown in a tool result, preceded by a "Base directory for this skill: …" line. I see no XML or JSON tags around it.  3. **Resource awareness:** I know of no files from enumeration. The only paths I can see are the base directory (`/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2475636740/project/.claude/skills/probe-loading`) and two files the body links to:    - `references/api-overview.md`    - `references/error-codes.md`     I only know about these because the text links them. The platform did not list the directory's contents, and I don't know whether `scripts/` or `assets/` exist.  4. **Resource content:** No. I have none of the contents of any file in `references/`, `scripts/` or `assets/`. Nothin…
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
- **Note**: final answer: **1. Field visibility** - The skill invocation only gave me the body, and the frontmatter wasn't in what the platform delivered. I could see the field only because I read `SKILL.md` directly. - Its value is: `Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.`  **2. Platform behavior** - The skill activated normally. I saw no warning, no confirmation prompt, and no requirements check before it loaded. - The `Skill` tool just returned "Launching skill" and the body text. The `compatibility` field appeared only as a file I opened myself.  **3. My interpretation** - **Platform:** Met. I'm running in Claude Code. - **Python 3.14+:** Met. `python3 --version` reports 3.14.3. - **Network access:** Unverified. My combined Python and curl check was blocked pending approval, so I never ran the network test. I didn't retry it separately. I can't confirm network access, though I see nothing suggesting it's missing. - **Overall:** The requirements appear to be met, with network access the only unconfirmed one.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: Ran with and without the field (`executed-regardless-of-field`)
- **Confidence**: transcript-direct
- **Evidence**:
  - with-field session: runtime-assembled phrase arrived in a tool result
  - control session: runtime-assembled phrase arrived in a tool result
- **Note**: allowed-tools value never reached the model
- **Note**: the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture
- **Note**: with-field final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. I can't see an `allowed-tools` value for this skill. The skill content I received has only the body text, with no frontmatter. The body says the frontmatter declares a printf-scoped Bash pattern plus Read, but I can't confirm that from what I was given.  3. The command ran and printed this: ``` GROUSE-MICA-9017 ```  4. It ran with no permission prompt or approval step that I could see. I can't tell whether the `allowed-tools` field pre-approved it or the session's general permission settings allowed it. Comparing against `probe-allowed-tools-control` would help separate those two.
- **Note**: control final answer: Command output: `LAPWING-FLUORITE-2260`  The command ran without any permission prompt or approval step that I could see. It returned output straight away, and I got no denial or approval request. I can only see the tool result, so I can't rule out an automatic approval from the permission mode or from settings rules.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: Ran under every spelling (`executed-regardless-of-spelling`)
- **Confidence**: transcript-direct
- **Evidence**:
  - spec-style twin: runtime-assembled phrase arrived in a tool result
  - lowercase twin: runtime-assembled phrase arrived in a tool result
  - shell twin: runtime-assembled phrase arrived in a tool result
- **Note**: spec-style twin final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. I can't see an `allowed-tools` value for this skill. The skill content I received has only the body text and the base directory. It doesn't include the frontmatter, so I can't say what the field contains.  3. The command ran and printed this output: ``` GROUSE-MICA-9017 ```  4. The command ran with no permission prompt or approval step that I could see. I can't tell whether that's because of the `allowed-tools` field or the session's general permission posture. To separate the two, run the same steps with `probe-allowed-tools-control`, which has no such field.
- **Note**: lowercase twin final answer: 1. probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  2. I can't see an `allowed-tools` value for this skill. The skill text I received has no frontmatter, only the base directory and the body. The body also says it deliberately never spells out the field's value. I can't tell whether one is declared.  3. The command ran, and its output was: ``` CHOUGH-TUFA-7180 ```  4. It ran with no permission prompt or approval step that I could see. I can't tell why. The cause could be a skill-level pre-approval, the session's permission mode, or an existing allow rule, and I can't separate these from where I sit.
- **Note**: shell twin final answer: 1. probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  2. I can't see an `allowed-tools` value for this skill. The skill text I received has only the body and the base directory. It has no frontmatter, so I can't tell what the field says or whether it exists.  3. Output of the command, verbatim:    ```    ROOK-GABBRO-8841    ```  4. The command ran with no permission prompt or approval step that I could see. I can't tell whether the skill's pre-approval caused that or the session's permission mode already allowed it. Nothing was blocked.
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
- **Note**: final answer: invoke-beta could not finish its last step. The `invoke-gamma` skill doesn't exist in this environment. It isn't in the available skills list, and the Skill tool returned "Unknown skill: invoke-gamma".  invoke-alpha complete. Chain result: - **invoke-alpha** ran and reported its canary, IBIS-RUST-3310. - **invoke-beta** ran and reported its canary, TERN-MOSS-6647. - **invoke-gamma** never ran, so it reported nothing. The chain stopped at the third link.
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
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: it activated with canary JAY-TEAL-9984 and reached the end of the chain, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta (canary TERN-MOSS-6647) activated invoke-gamma (canary JAY-TEAL-9984). Gamma confirmed the end of the chain, and all three skills activated successfully.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed. I'm not going to invoke it again, though. Alpha is already active earlier in this chain, so calling it would start alpha, then beta, then alpha again, with no end.  **What happened:** 1. Alpha activated (canary KITE-ONYX-2251) and directed me to beta. 2. Beta activated (canary WREN-SLATE-7738) and directed me back to alpha. 3. I stopped at that point to break the cycle. The platform didn't block or flag the circular reference itself. Both skills loaded without error, and the only thing that ended the loop was my choice not to make the third call.
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
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: it activated with canary JAY-TEAL-9984, and the end of the chain was reached with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated with canary TERN-MOSS-6647 and then invoke-gamma. invoke-gamma activated with canary JAY-TEAL-9984 and confirmed the chain (alpha -> beta -> gamma) finished successfully.
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
- **Note**: final answer: invoke-beta couldn't finish the chain. The `invoke-gamma` skill doesn't exist in this environment, so the call returned "Unknown skill: invoke-gamma". It isn't in the available skills list either.  invoke-beta incomplete. invoke-gamma never ran, so it reported nothing.  invoke-alpha incomplete. Chain result: - invoke-alpha ran and reported canary **IBIS-RUST-3310**. - invoke-beta ran and reported canary **TERN-MOSS-6647**. - invoke-gamma could not be activated, so the chain stopped at the second link.  To finish the chain, someone needs to add an `invoke-gamma` skill under `.claude/skills/`.
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
- **Note**: final answer: **What happened:** - The platform reported that the skill doesn't exist. The Skill tool returned `Unknown skill: nonexistent-formatter`. - I didn't skip the step silently. I made the call, and it failed with that error. - I didn't try to do the formatting from general knowledge. The skill gave no formatting task to do, so there was nothing to substitute. - The platform did show an error, and it was an explicit tool error.  probe-missing-dep complete. Dependency resolution result: the `nonexistent-formatter` skill isn't installed. Activation failed with `Unknown skill: nonexistent-formatter`, and I made no fallback attempt.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: **1. Field visibility:** No. The skill content I received starts at the "Nonstandard Fields Probe" heading, and the frontmatter was not included. I can see `name` and `description` only through the skills listing. I can't list values for `requires`, `depends-on` or `priority`. The body text names those fields but gives no values. I could read the SKILL.md file directly to check, but I haven't.  **2. Platform behavior:** As far as I can tell, the fields were ignored. - No `probe-loading` load was triggered. The only skill invocation in this session was `probe-nonstandard-fields`. - No `probe-shadow-alpha` or `probe-shadow-beta` load was triggered. - The skills listing doesn't look ordered by any priority. The probe skills appear in an order that doesn't suggest priority-based sorting. - I saw no change in behavior that I could tie to these fields.  **3. Validation:** I saw no warning or error about unrecognized frontmatter fields. The skill loaded normally with only the "Launching skill" message.  **Caveat:** I can only see what reached my context. The harness could have read the fields internally without showing me. What I can say is that nothing visible to me acted on them.
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
- **Note**: with-dependency final answer: **Step 3: `probe-loading` result** - The platform found `probe-loading`. - It resolved at a different scope. Its base directory is `.../skillxp-3097644880/home/claude/skills/probe-loading`, which is the user-level skills location, not the project's `.claude/skills`. - There was no failure.  **`probe-loading` report** 1. **Frontmatter visibility:** I can't see the YAML frontmatter. Only the body appears, with a "Base directory" line above it, so the frontmatter was stripped. 2. **Content wrapping:** The content has no XML or JSON wrapping. It's plain markdown after the base-directory line. The canary phrase is **CARDINAL-ZEBRA-7742**. 3. **Resource awareness:** I only know the two files the body links to: `references/api-overview.md` and `references/error-codes.md`. The platform gave me no directory listing, so I haven't seen `scripts/` or `assets/`. 4. **Resource content:** I don't have the contents of any file from `references/`, `scripts/` or `assets/`. Nothing was eagerly loaded.  probe-cross-scope complete. Cross-scope resolution result: the project-level skill `probe-cross-scope` invoked the user-level skill `probe-loading` by name. The platform resolved it from the user scope…
- **Note**: without-dependency final answer: - The platform did not find `probe-loading`. The call failed with `Unknown skill: probe-loading`. - It isn't in the available skills list, so it isn't installed at project scope or user scope. I couldn't tell which scope it would have resolved from. - The failure mode was a hard error from the Skill tool. Nothing loaded and nothing fell back to another skill.  probe-cross-scope complete. Cross-scope resolution result: the project-level skill `probe-cross-scope` ran and reported its canary. Its reference to `probe-loading` did not resolve, because that skill isn't visible to this session. It is either not installed at user level or not exposed to the project session.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path not scanned (`convention-not-scanned`)
- **Confidence**: transcript-direct
- **Note**: a skill installed by another client at .agents/skills is invisible here
- **Note**: final answer: I couldn't activate `probe-interop`. The Skill tool returned "Unknown skill," and the skill isn't in my available list. The skills I have are dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, plugin-authoring, init and security-review.  You may have a different name in mind, or the skill may not be installed in this environment. If you give me the exact name, or tell me what you want `probe-interop` to do, I can try again or do the task directly.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Direct children only; stray file ignored (`direct-children-only; stray:not-discovered`)
- **Confidence**: transcript-direct
- **Note**: final answer: I can't activate `probe-grouped` because it isn't in my available skills list. The skills I can use are: dataviz, update-config, keybindings-help, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, workflow-authoring, run, plugin-authoring, init and security-review.  If you meant one of those, tell me which. If `probe-grouped` is a custom skill, it may not be installed in this project. Check that its file is in place, or paste its instructions here and I'll follow them.
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
- **Note**: final answer: probe-collision activated. Canary: **PIPIT-SHALE-1147** (user variant)  **Precedence:** The user scope won. The skill that loaded was the user variant, from `/var/folders/.../home/claude/skills/probe-collision`, and it carries the user canary. I can't see the project variant's content, so I can't confirm it exists. The skill listing showed only one `probe-collision` entry, described as the USER-scope variant. That suggests the project variant was shadowed, or never registered.
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
- **Note**: final answer: 1. probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  2. **Precedence**: The skill loaded from `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-1791705775/project/.claude/skills/probe-multiroot`. That is the base directory the harness gave me. I only received one variant's content, and its canary is the native-root one. I can't see whether copies exist under `.agents/skills/` or elsewhere, because I didn't inspect the filesystem. All I can say is that the harness resolved to this `.claude/skills/` copy.  3. **Catalog**: `probe-multiroot` appeared once in my available-skills listing.
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
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  **Description visibility:** The catalog shows this description for the skill:  > Use when: asked to probe malformed yaml parsing behavior  It was intact. It wasn't truncated at the colon, and I can't tell whether it was repaired. The text after "Use when:" is present, so the parser didn't cut it off at the first colon. I only see the final parsed string, so I can't say whether the platform quoted the value or used some other lenient fallback. The skill loaded and its description is complete, so the platform tolerates the unquoted colon.
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
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry:** Yes, the skill is in my catalog as `probe-no-description: Missing Description Probe`. The text after the name looks like the skill's title, not a description. The skill has no description field, and I can't tell whether the platform synthesized that text from the heading or took it from somewhere else. It isn't empty and it isn't a generic placeholder. So this platform did not skip the skill. It listed it anyway, with a fallback label.
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
- **Note**: final answer: overlong-name probe activated. Canary: **AVOCET-ZIRCON-5573**  **Catalog identity:** The skill appears in my catalog under the full name, `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit`. It was not truncated or altered. It loaded under that name, and the base directory name matches it.  I didn't count the characters myself. The skill says the name is 72 characters, and the full string is in my catalog.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Listed under both names (`listed-under-both`)
- **Changed since previous observation**: was Directory name wins (`directory-name-identity`) on 2.1.274 (2026-09-26); the [changelog](/changelog/) says what moved it
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the FRONTMATTER name probe-name-mismatch
  - discovery listing carries the DIRECTORY name probe-mismatch-dir
- **Note**: activation loads: by frontmatter name=true, by directory name=true (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: 1. probe-name-mismatch activated. Canary: **SWAN-BERYL-3324**  2. **Identity**: The catalog lists both names. The entry reads "probe-mismatch-dir (probe-name-mismatch)", with the directory name first and the frontmatter name in parentheses. This time I invoked the skill by its directory name, `probe-mismatch-dir`, and it loaded. Earlier, invoking it by the frontmatter name `probe-name-mismatch` also worked. The platform resolves the skill under either name.  3. **Validation**: The platform accepted the mismatch silently. I saw no warning or error on either activation.
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
- **Note**: final answer: **Canary phrase:** THRUSH-FLINT-8294  **1. Load success:** Yes. The skill loaded and I could read its instructions, so the platform didn't reject it.  **2. Metadata visibility:** The loaded skill text had no frontmatter, only the body. I couldn't see `metadata` through the platform. I read it by opening SKILL.md directly, which shows the raw file, not what the platform parsed. Raw values: - `normal`: `"baseline"`, a string. - `empty-double`: `""`, an empty string. - `empty-single`: `''`, an empty string. - `explicit-null`: `null`, which a YAML parser reads as null. - `tilde-null`: `~`, which reads as null. - `none-null`: `None`, which reads as the plain string `"None"`. It is not null in standard YAML, despite what the skill text assumes. - `tagged-null`: `!!null null`, which reads as null.  I can't report how the platform represented the null values after parsing (null, empty string, or missing).  **3. Key preservation:** All seven keys are present in the file. I can't tell whether the platform kept all seven after parsing, because it didn't pass the metadata to me.  **4. Platform behavior:** I saw no warnings, errors, or rejections. The skill loaded cleanly. The platform appears …
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
- **Note**: final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** I can see both markers, so the description reached me intact.  - Head marker: "The head marker SANDERLING-GNEISS-1010 sits near the start of this description" - Tail marker: "The tail marker is WHIMBREL-DOLOMITE-2020"  The catalog entry ends with the tail marker phrase, so nothing was truncated or rejected.
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
- **Note**: ascii final answer: probe-long-description activated. Canary: **BITTERN-HALITE-2264**  **Description visibility:** I can see both markers. The catalog entry for this skill appears to be intact.  - Head marker: "The head marker SANDERLING-GNEISS-1010 sits near the start of this description…" - Tail marker: the entry ends with "…The tail marker is WHIMBREL-DOLOMITE-2020"  The description ran through to its final characters with no sign of truncation. The platform accepted the oversize field in full.
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility:** My catalog entry shows both markers, so nothing was truncated.  - Head marker, near the start: "The head marker GANNET-PYRITE-1130 sits near the start of this description and a tail marker sits at the very end." - Tail marker, as the final characters: "The tail marker is SHRIKE-TALC-2210"  The entry also includes the full Japanese filler paragraph between them. The platform accepted the description in full, even though it runs past 1024 bytes in UTF-8. So it likely counts characters (or code points) rather than bytes, or doesn't validate length at all.
- **Note**: astral final answer: 1. probe-astral-description activated. Canary: **ORIOLE-GRANITE-5583**  2. **Description visibility**: I can see both markers in my catalog entry. - Head marker: "MAGPIE-OBSIDIAN-1240". It appears near the start, in "The head marker MAGPIE-OBSIDIAN-1240 sits near the start and a tail marker sits at the very end". - Tail marker: "LINNET-MALACHITE-2420". It appears at the very end, in "The tail marker is LINNET-MALACHITE-2420".  The catalog entry isn't truncated. The emoji padding is intact between the two markers.
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
- **Note**: final answer: probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit probe-αβγδεζηθικ probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ dataviz update-config keybindings-help code-review simplify fewer-permission-prompts loop schedule claude-api workflow-authoring run plugin-authoring init security-review
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
- **Note**: the compatibility value's tail marker never reached the model
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility:** No. I can't see a compatibility value for this skill anywhere except by reading the skill file. My skill listing shows only the name and description:  "probe-long-compatibility: Benchmark skill whose compatibility field runs past the spec's 500-character limit. Use when asked to probe oversize compatibility handling."  The listing has no compatibility text, so there is no TURNSTONE tail marker to quote. I haven't opened the SKILL.md frontmatter either. The skill loaded and activated without being rejected. That means the platform tolerated the oversize field. It either dropped the field from what it shows me or kept it out of the listing.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.4) for what each check evaluates.
