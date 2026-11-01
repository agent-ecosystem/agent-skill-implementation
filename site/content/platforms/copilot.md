---
title: "GitHub Copilot CLI (headless)"
description: "Automated skill loading findings for GitHub Copilot CLI (headless) (check list 0.4)."
date: 2026-11-01
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | GitHub Copilot CLI (headless) |
| **Platform version** | 1.0.105 |
| **Check list version** | 0.4 |
| **Test date** | 2026-11-01 |
| **Model(s) observed** | claude-sonnet-5.5 |
| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists. These findings describe the release in the header on the test date; platform behavior changes between releases, and the [changelog](/changelog/) records what moved from one run of the checks to the next.

> **Changed since the previous observation**: 1 verdict differs from the previous run of the checks on this platform: [`path-resolution-base`](#path-resolution-base). Each finding below states the earlier verdict and release.

## Spec alignment

Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. 21 of the 46 checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.

### Where behavior contradicts the spec

- [`bundled-script-execution`](#bundled-script-execution): The spec presents scripts/ as executable code agents can run, but execution was blocked in this headless run. Interactive sessions, where a user can approve the command, may behave differently.
- [`frontmatter-handling`](#frontmatter-handling): The spec says the agent loads the entire SKILL.md file at activation. This platform strips the YAML frontmatter and injects only the body, so frontmatter fields beyond name and description never reach the model.
- [`description-length-unit`](#description-length-unit): The spec caps description at 1024 characters without defining the unit; its skills-ref reference validator counts code points. This platform counts UTF-16 code units, so a description with emoji or other supplementary-plane characters that the reference validator accepts is rejected or truncated here.
- [`name-length-unit`](#name-length-unit): The spec caps name at 64 characters without defining the unit and allows unicode lowercase alphanumeric characters with an ASCII parenthetical, which reads two ways; its skills-ref reference validator accepts any Unicode alphanumeric and counts code points. This platform rejects non-ASCII names on the character rule before any counting, so names the reference validator accepts never enter its catalog and its counting unit cannot be observed.

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

- [`path-resolution-base`](#path-resolution-base): The spec's skill-root-relative paths went untested in this run: the model rewrote each reference to a fully qualified path before use, so the platform's own resolution base was never exercised. (Behavioral inference.)
- [`allowed-tools-behavior`](#allowed-tools-behavior): The spec marks allowed-tools experimental, with varying support. Neither twin session produced a clean execution, so the field's effect is unresolved here; see the finding for per-session tiers.
- [`allowed-tools-name-matching`](#allowed-tools-name-matching): The spec marks allowed-tools experimental, leaves tool names to each platform, and says support may vary, so no outcome contradicts it. No spelling unblocked the command, so either the field is ignored or none of the three spellings names a tool this platform pre-approves; the matching rule is unresolved here.


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
- **Verdict**: All three listed, contents not loaded (`all-three-dirs-enumerated`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - scripts/ file name injected at activation
  - references/ file name injected at activation
  - assets/ file name injected at activation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: Names listed, contents not loaded (`resources-enumerated-not-loaded`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - resources/ file name injected without its content
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: Names listed, contents not loaded (`enumerated-not-loaded:[evals/ templates/]`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - evals/ file name injected without its content
  - templates/ file name injected without its content
- **Note**: the platform lists the nonstandard directories' files at activation but loads nothing from them; the model can read them on demand like any other listed file
- **Note**: final answer: **1. Directory awareness:** The skill loader listed three files under the skill's base directory: - `evals/evals.json` - `resources/api-reference.md` - `templates/output-template.md`  So I know of `evals/`, `resources/` and `templates/`. The skill text names the same three directories. There is no `scripts/`, `references/` or `assets/` directory.  **2. Content visibility:** I can see the file listing for each directory, as paths under "Related files". I cannot see any file contents. Those are only available if I read them with the `view` tool, and I haven't. Nothing was preloaded.  **3. Resources vs. references:** I can't confirm how the platform treats `references/`, because this skill doesn't have one. The loader enumerated the `resources/` file path alongside `evals/` and `templates/` and treated all three the same way. Each was listed but its content was not loaded. From this one run, nothing sets `resources/` apart from the other nonstandard directories. The loader seems to list every file in the skill, whatever the directory name.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Resource Access Patterns

#### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: File names listed, contents not loaded (`listing-enumerated-without-contents`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - unlinked reference file's NAME injected without its content
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?_

- **Status**: observed
- **Verdict**: Model used full paths (base untested) (`model-preemptively-qualified`)
- **Changed since previous observation**: was Bare path fails; model recovers (`cwd-base-model-requalified`) on 1.0.88 (2026-09-25); the [changelog](/changelog/) says what moved it
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
- **Note**: final answer: **Alpha `references/API.md`** (canary **STORK-CORAL-4471**): the alpha skill's own version. - `GET /alpha/status`: health check - `POST /alpha/process`: processing  **Beta `references/API.md`** (canary **EGRET-SLATE-8823**): I could read it by its full path, `.github/skills/probe-shadow-beta/references/API.md`. - `GET /beta/status` - `POST /beta/process`  No shadowing occurred. Each skill's full path resolved to its own file. Only the alpha skill was loaded. I read the beta file directly by path, not through a loaded beta skill.
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
- **Note**: final answer: Path traversal probe results. I resolved each relative path from the probe-traversal skill directory (`.github/skills/probe-traversal`).  1. **Sibling skill (`../probe-loading/SKILL.md`)**: Allowed. I got the full file contents with no error or warning. The platform enforced no boundary. 2. **Parent directory**:    - `../README.md`: Error, "path does not exist". The file isn't there, so nothing was blocked.    - `../` listing: Allowed. It returned `probe-loading` and `probe-traversal`. 3. **Two levels up (`../../checks.md`)**: Error, "path does not exist". The file isn't there, so this isn't evidence of blocking either.  I saw no sandboxing, warnings, or silent empty results. Reads outside the skill root succeed whenever the file exists. The two failures were simply missing files, and `../../` resolves to `.github/`.
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
- **Note**: final answer: The script did not run. The bash call returned this error:  > Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission.  No phrase was printed, so I have nothing to report from the script. I won't guess it. The only canary I have is the one from the skill body, **REDSHANK-SYENITE-8807**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `bundled-file-enumeration-scale`

_When a skill ships forty numbered references, a hidden dotfile, a binary asset, and a vendored code tree, does the platform's activation listing carry all of them, stop at a cap, or leave some kinds out?_

- **Status**: observed
- **Verdict**: Omitted: vendored (`filtered:[vendored]`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - first numbered file name injected at activation
- **Note**: listed 40 of 40 numbered files; kinds listed: [dotfile binary]; kinds omitted: [vendored]
- **Note**: final answer: 1. probe-bulk-files activated. Canary: **SKUA-DIORITE-2917**  2. **File awareness**: These are the files I was told about at activation. The paths below are relative to the skill's base directory.    - `.hidden-dotfile-marker.md`    - `assets/binary-pixel-marker.png`    - `references/bulk-file-01.md` through `references/bulk-file-40.md`. That is 40 consecutively numbered files (01, 02, 03, … 40).     I was told about no vendored code tree files, although the skill text mentions one.  3. **Count**: 42 files (1 dotfile, 1 PNG, 40 numbered references).
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
- **Note**: final answer: Here are the skills in my available skills list, copied verbatim. Each entry shows only name, description, and location.  1. **probe-compatibility**    - description: Benchmark skill for testing how platforms handle the compatibility field. Use when asked to probe compatibility behavior.    - location: project  2. **probe-loading**    - description: Benchmark skill for testing loading timing, resource enumeration, content presentation, and lifecycle behavior. Use when asked to probe loading behavior.    - location: project  3. **probe-metadata-values**    - description: Benchmark skill for testing how platforms handle edge-case YAML values in the metadata frontmatter field. Use when asked to probe metadata value handling.    - location: project  4. **customize-cloud-agent**    - description: Skill for customizing the Copilot cloud agent (formerly known as Copilot coding agent) environment, including copilot-setup-steps.yml configuration, preinstalling tools and dependencies, runners, and settings. Use when the user mentions copilot-setup-steps, copilot setup steps, or wants to configure the cloud agent environment.    - location: builtin  5. **github-pr-media**    - description: Up…
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
- **Verdict**: Wrapped in structured tags (`wrapped-structured`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - injection event carrying the body canary
- **Note**: the whole delivery is enclosed in: skill-context
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `activation-location-disclosure`

_Does the content injected at activation tell the model where the skill lives on disk (its directory path), apart from any discovery listing?_

- **Status**: observed
- **Verdict**: Skill path stated at activation (`path-stated-at-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - skill path in harness-injected activation content
- **Note**: final answer: **Probe report**  1. **Frontmatter visibility**: No. The YAML frontmatter was stripped. I only see the body. The name and description appear in my available-skills list, but not as frontmatter. I can't see allowed-tools, compatibility, or metadata.  2. **Content wrapping**: Yes. The content is wrapped in a `<skill-context name="probe-loading">` XML-style tag. It begins with a "Base directory for this skill:" line and a "Related files" list.  3. **Resource awareness**: Yes. The activation text listed these files, with absolute paths under `.github/skills/probe-loading/`:    - `assets/config-template.yaml`    - `references/api-overview.md`    - `references/error-codes.md`    - `references/unreferenced-detail.md`    - `scripts/check-status.sh`     The list includes `unreferenced-detail.md`, even though the body never mentions it. So the platform enumerated the directory itself.  4. **Resource content**: No. I have only the file paths. None of the contents of the `references/`, `scripts/` or `assets/` files are in my context.  **Canary**: CARDINAL-ZEBRA-7742. I only know it because the skill was activated.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Lifecycle Management

#### `reactivation-deduplication`

_When the same skill is activated twice in one session, is its content loaded again or deduplicated?_

- **Status**: observed
- **Verdict**: Full content re-injected every time (`reinjected-each-activation`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Runs**: 3 (reinjected-each-activation ×3)
- **Evidence**:
  - first injection of body canary
  - second injection of body canary
- **Note**: platform-level: duplicate skill content occupies context after reactivation
- **Note**: verdict consistent across 3 runs
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
- **Note**: final answer: **1. Field visibility:** The skill context I was given contained only the body, not the frontmatter. I read the SKILL.md file directly to see the field. Its value is: "Designed for Claude Code (or similar products). Requires Python 3.14+ and network access."  **2. Platform behavior:** The skill activated normally. There was no warning, no confirmation prompt, and no requirements check before activation.  **3. My interpretation:** I can't confirm the requirements are met. - **Platform:** I'm the GitHub Copilot CLI, not Claude Code. It's arguably a "similar product", so this is ambiguous. - **Python 3.14+:** Unverified. My attempt to run `python3 --version` was denied for lack of an interactive approval. - **Network access:** Also unverified, because the same blocked command included a `curl` check.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

#### `allowed-tools-behavior`

_Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?_

- **Status**: observed
- **Verdict**: Blocked with and without the field (`with-field:blocked-visibly; control:blocked-visibly`)
- **Confidence**: transcript-direct
- **Evidence**:
  - with-field session: instructed command's tool call returned an error result
  - control session: instructed command's tool call returned an error result
- **Note**: allowed-tools value never reached the model
- **Note**: with-field final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. I can't see an `allowed-tools` value for this skill. The loaded context has only the skill body and base directory, with no frontmatter field.  3. The command was blocked. The exact error was: "Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission."  4. It did not run without an approval step. It needed approval, and none was available in this non-interactive session. So in this environment, `allowed-tools` (if it was declared) did not pre-approve the printf command.
- **Note**: control final answer: probe-allowed-tools-control activated. Canary: **STINT-MARBLE-9912**  Command output: the command was blocked. The exact error was:  `Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission.`  Permission prompt: the command did not run silently. It needed approval, and none could be given because this session is non-interactive. This skill has no `allowed-tools` field, so nothing pre-approved the command.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `allowed-tools-name-matching`

_Does the effect of allowed-tools depend on spelling the tool the platform's way? Three twins declare the same intent as Bash(printf:*), bash, and shell._

- **Status**: observed
- **Verdict**: Blocked under every spelling (`blocked-regardless-of-spelling`)
- **Confidence**: transcript-direct
- **Evidence**:
  - spec-style twin: instructed command's tool call returned an error result
  - lowercase twin: instructed command's tool call returned an error result
  - shell twin: instructed command's tool call returned an error result
- **Note**: spec-style twin final answer: 1. probe-allowed-tools activated. Canary: **CURLEW-SCHIST-4419**  2. I can't see an `allowed-tools` value for this skill. The loaded context has only the skill body and base directory, with no frontmatter field.  3. The command was blocked. The exact error was: "Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission."  4. It did not run without an approval step. The runtime required permission approval and denied it because this session is non-interactive. So if the skill declares `allowed-tools`, it did not pre-approve the printf command here.
- **Note**: lowercase twin final answer: 1. probe-allowed-tools-lowercase activated. Canary: **STILT-SCORIA-5526**  2. I can't see an `allowed-tools` value for this skill. The loaded context has only the skill body and base directory, with no frontmatter field.  3. The command was blocked. The exact error was: "Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission."  4. No. The command needed approval, and none could be given in this non-interactive session. Any `allowed-tools` pre-approval the skill declares did not apply to my bash tool.
- **Note**: shell twin final answer: 1. probe-allowed-tools-shell activated. Canary: **AUKLET-CHERT-3364**  2. I can't see an `allowed-tools` value for this skill. The loaded content has only the instructions body, with no frontmatter field.  3. The command was blocked. The exact error was: "Permission denied because no interactive user response was available. Retry in an interactive session so the user can approve it, or try an alternative that does not require this permission."  4. No, it didn't run without an approval step. It needed a permission approval, and none could be given in this non-interactive session. Any `allowed-tools` declaration did not pre-approve the shell here.
- **Note**: no spelling of the field unblocked the command; either the field is ignored or none of the spellings matched a tool this platform pre-approves
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
- **Note**: final answer: invoke-gamma isn't available, so the chain stopped there.  invoke-beta complete. invoke-gamma said nothing, because the skill wasn't found. The available skills are invoke-alpha, invoke-beta, customize-cloud-agent and github-pr-media.  invoke-alpha complete. Chain result: invoke-alpha reported canary IBIS-RUST-3310 and invoke-beta reported canary TERN-MOSS-6647. invoke-gamma could not be activated, so the chain stopped at the second skill.
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
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: it activated with canary JAY-TEAL-9984 and that the end of the chain was reached, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated with canary TERN-MOSS-6647. invoke-gamma then activated with canary JAY-TEAL-9984 and confirmed the end of the chain.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: probe-circular-beta activated. Canary: **WREN-SLATE-7738**  Activating probe-circular-alpha as directed.  Alpha is already active in this chain, so activating it again would start an infinite loop. I'm stopping here instead of re-invoking it.  **What happened:** - Alpha activated (canary KITE-ONYX-2251) and activated beta. - Beta activated (canary WREN-SLATE-7738) and was directed back to alpha. - I did not make that second call to alpha, so the cycle was cut after one round trip. - Nothing in the platform detected or blocked the circular reference. The loop was stopped only because I declined to follow it.
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
- **Note**: final answer: invoke-gamma activated. Canary: **JAY-TEAL-9984**  End of chain reached. All three skills in the invoke-alpha -> invoke-beta -> invoke-gamma chain activated successfully.  invoke-beta complete. invoke-gamma said: it activated with canary JAY-TEAL-9984 and the chain reached its end, with all three skills activated successfully.  invoke-alpha complete. Chain result: invoke-beta activated (canary TERN-MOSS-6647) and invoke-gamma activated (canary JAY-TEAL-9984). The chain ended successfully.
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
- **Note**: final answer: invoke-gamma isn't available, so the chain stopped there.  invoke-beta complete. invoke-gamma said: nothing. The skill was not found, and the available skills are invoke-alpha, invoke-beta, customize-cloud-agent and github-pr-media.  invoke-alpha complete. Chain result: invoke-alpha (canary IBIS-RUST-3310) and invoke-beta (canary TERN-MOSS-6647) both activated. invoke-gamma could not be activated because it doesn't exist in this project, so the third link did not run.
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
- **Note**: final answer: - The platform reported that the skill doesn't exist: `Skill "nonexistent-formatter" not found. Available skills: probe-missing-dep, customize-cloud-agent, github-pr-media`. - I didn't skip the step silently. I made the call and got an explicit error. - I didn't try to do the formatting from general knowledge. No formatting task was given.  probe-missing-dep complete. Dependency resolution result: activating the missing skill failed with a clear "not found" error that listed the available skills. No fallback was used.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: 1. **Field visibility:** No. The skill content I received had only the body instructions. The frontmatter wasn't included, so I can't see `requires`, `depends-on` or `priority` or their values. I only saw the skill's name and description, from the skill list.  2. **Platform behavior:** The fields appear to have been ignored.    - No `probe-loading` skill was loaded because of `requires`.    - No `probe-shadow-alpha` or `probe-shadow-beta` skills were loaded because of `depends-on`.    - I saw nothing that ordered skills or changed behavior based on `priority`.    - The only thing that happened was the skill I asked for being loaded.  3. **Validation:** I saw no warnings or errors about unrecognized frontmatter fields. The skill was accepted silently.  I didn't open the SKILL.md file itself, so I can't say what the fields contain.
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
- **Note**: with-dependency final answer: **probe-loading report**  1. **Frontmatter visibility:** No. The YAML frontmatter was stripped. I only see the base directory, the file list and the markdown body. The name and description reached me only through the skill list. 2. **Content wrapping:** Yes. The content comes in a `<skill-context name="probe-loading">` XML-style tag. Inside it are a "Base directory for this skill" line, a "Related files" list and then the markdown body. 3. **Resource awareness:** Yes, from the "Related files" list. I know of five files:    - `assets/config-template.yaml`    - `references/api-overview.md`    - `references/error-codes.md`    - `references/unreferenced-detail.md`    - `scripts/check-status.sh` 4. **Resource content:** No. I have the paths only, not the contents of any file in `references/`, `scripts/` or `assets/`.  The body also contains the canary **CARDINAL-ZEBRA-7742**.  **probe-cross-scope complete.** Cross-scope resolution result: the platform found `probe-loading` at user scope (`home/copilot/skills/probe-loading`). The project-level skill activated it without any failure. It loaded on activation, not at discovery time.
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  **Scope:** I'm a project-level skill. The skill list shows `<location>project</location>`, and the skill was loaded from `.github/skills/probe-cross-scope`.  **Activating `probe-loading`:** The platform did not find it. The call returned `Skill "probe-loading" not found. Available skills: probe-cross-scope, customize-cloud-agent, github-pr-media`. It failed cleanly with an error and no crash. Because it wasn't found, I can't say which scope it would have resolved from.  probe-cross-scope complete. Cross-scope resolution result: `probe-loading` was not resolved. It either isn't installed at user level or user-level skills aren't visible to this session. Only the project and builtin skills are available.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Discovery Scope

#### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path scanned (`convention-scanned`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-interop
  - body canary loaded on activation
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  **Discovery:** Yes, this skill appeared in my available skills catalog, listed with location "project". It was loaded from `/private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-3973730261/project/.agents/skills/probe-interop`. So the platform scans the cross-client `.agents/skills` convention path.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Scans the root recursively; stray file ignored (`recursive-scan; stray:not-discovered`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-grouped from one level below the root
  - grouped skill's body loaded on activation
- **Note**: final answer: probe-grouped activated. Canary: **CROW-AGATE-6105**  Discovery: Confirmed. probe-grouped appears in my available skills catalog, even though it sits one level below the skills root, at `probe-group/probe-grouped/`.
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
- **Verdict**: Project scope wins (`project-overrides-user`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - project variant's canary loaded; user variant's never appeared
- **Note**: platform-resolved: the harness's activation mechanism injected the winning variant
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence:** The project scope won. The loaded skill is the project-level variant, located at `.github/skills/probe-collision`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `multi-root-collision-precedence`

_With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?_

- **Status**: observed
- **Verdict**: Native root wins (`native-root-wins`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-multiroot-beacon-agents
  - discovery listing names probe-multiroot-beacon-claude
  - discovery listing carries the native variant's description
  - native variant's body canary loaded
- **Note**: foreign roots installed alongside the native copy: .agents/skills, .claude/skills
- **Note**: platform-level: the beacon skills from .agents/skills, .claude/skills were listed, so those roots were scanned and the colliding variant was dropped by name in favor of the native copy
- **Note**: listed variants: native
- **Note**: roots proven scanned by their beacon: .agents/skills, .claude/skills
- **Note**: final answer: probe-multiroot activated. Canary: **GREBE-AZURITE-7301** (native-root variant)  **Precedence:** The platform loaded the copy from the native root, `.github/skills/probe-multiroot`. That is the base directory the skill reported, and the canary matches the native-root variant.  **Catalog:** `probe-multiroot` appeared once in my available-skills listing, with location "project". The two beacon skills had distinct names and didn't collide.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### Validation Strictness

#### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Skipped by a strict parser (`skipped-strict-parser`)
- **Confidence**: transcript-direct
- **Note**: final answer: I couldn't activate `probe-malformed-yaml` because no skill with that name exists. The only skills available are `customize-cloud-agent` and `github-pr-media`. If you meant one of those, tell me which and what you want done.
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
- **Note**: final answer: probe-no-description activated. Canary: **VIREO-PUMICE-3049**  **Catalog entry:** Yes, the skill appears in my catalog (location: project). Its description text is not empty or a placeholder. The catalog shows text taken from the skill's own body: "This skill has no description field. The client implementation guide says platforms should skip such skills ("a description is essential for disclosure") and log the error. This skill tests whether platforms…". It is cut off mid-sentence, which suggests the client synthesized the description from the first lines of the body.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `invalid-name-tolerance`

_Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Tolerated only: uppercase, double-hyphen (`invalid-names-tolerated:[uppercase double-hyphen]`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-Upper-Case
  - discovery listing names probe--double-hyphen
- **Note**: final answer: I couldn't activate `probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit` because it isn't in my catalog. The skill tool returned "not found". The available skills are `probe--double-hyphen`, `probe-Upper-Case`, `customize-cloud-agent` and `github-pr-media`.  The overlong name was probably rejected or dropped when skills were loaded. The spec caps names at 64 characters, and this one is longer.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Listed under both names (`listed-under-both`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the FRONTMATTER name probe-name-mismatch
  - discovery listing carries the DIRECTORY name probe-mismatch-dir
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: The skill `probe-mismatch-dir` wasn't found. The platform registers the skill under its frontmatter name, `probe-name-mismatch`, and doesn't recognize the directory name as an identifier.  I already activated it under `probe-name-mismatch` in the previous turn. Its canary phrase is **SWAN-BERYL-3324**.
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
- **Note**: final answer: **Canary phrase:** THRUSH-FLINT-8294  1. **Load success:** Yes. The skill loaded and I can read its instructions.  2. **Metadata visibility:** The loaded skill context did not include the `metadata` field. I could only see it by reading the raw SKILL.md file from disk. The raw frontmatter says:    - `normal`: `"baseline"`    - `empty-double`: `""`    - `empty-single`: `''`    - `explicit-null`: `null`    - `tilde-null`: `~`    - `none-null`: `None`. A YAML parser reads this as the string "None", not null.    - `tagged-null`: `!!null null`     I can't report how the platform parsed these, because the parsed values weren't exposed to me.  3. **Key preservation:** All seven keys are present in the file. I can't confirm whether the platform kept all seven after parsing.  4. **Platform behavior:** I saw no warnings, errors, or rejections. The platform did not refuse the skill. The metadata was not passed into the skill context I received, so it may have been stripped or just not shown.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

#### `oversize-description-handling`

_Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?_

- **Status**: observed
- **Verdict**: Skipped (`skipped-oversize-description`)
- **Confidence**: transcript-direct
- **Note**: neither description marker appeared in harness-injected content; description delivery is unobservable here or the value was dropped
- **Note**: final answer: I couldn't activate `probe-long-description` because no skill with that name exists. The only available skills are `customize-cloud-agent` and `github-pr-media`. Tell me if you meant one of those, or give me a different task.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `description-length-unit`

_When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?_

- **Status**: observed
- **Verdict**: Counts UTF-16 code units (`counts-utf16-units`)
- **Vehicle**: harness-push
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-multibyte-description
  - body canary loaded on activation
  - description tail marker in injected content
- **Note**: description fates: ascii:rejected, multibyte:intact, astral:rejected
- **Note**: ascii final answer: I couldn't activate "probe-long-description" because no skill with that name exists. The available skills are `customize-cloud-agent` and `github-pr-media`. Tell me if you want one of those instead, or give me the instructions directly.
- **Note**: multibyte final answer: probe-multibyte-description activated. Canary: **PUFFIN-BASALT-4471**  **Description visibility:** I can see both markers in my catalog entry, so the description was not truncated.  - Head marker: "The head marker GANNET-PYRITE-1130 sits near the start of this description" - Tail marker: "The tail marker is SHRIKE-TALC-2210". It is the final text of the entry.  The full Japanese filler paragraph is also present. This platform accepted the 1822-byte description in full, so it is probably counting characters rather than UTF-8 bytes, or not validating length at all.
- **Note**: astral final answer: I couldn't activate `probe-astral-description` because no skill with that name exists. The only skills available are `customize-cloud-agent` and `github-pr-media`. If you meant one of those, tell me which and I'll use it.
- **Note**: the ASCII and astral overruns were enforced while the multibyte fixture survived: the platform counts UTF-16 code units (JavaScript's .length)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

#### `name-length-unit`

_When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?_

- **Status**: observed
- **Verdict**: Rejects non-ASCII names (unit unobservable) (`rejects-non-ascii-names`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd
- **Note**: name fates: ascii64:listed, ascii72:skipped, greek16:skipped, greek60:skipped, math40:skipped
- **Note**: the platform enforces the cap (72 ASCII rejected) but also dropped the 16-code-point Greek name, which is under 64 in every unit: non-ASCII names are rejected on the character rule, so the counting unit cannot be observed
- **Note**: final answer: probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd customize-cloud-agent github-pr-media
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
- **Note**: final answer: probe-long-compatibility activated. Canary: **KESTREL-BAUXITE-6690**  **Field visibility**: No. I can't see a compatibility value for this skill anywhere outside this file. The available skills list only showed the name, description, and location. The skill context I received contains the body, not the frontmatter. I have no compatibility text to quote, and I can't tell whether it ends with the TURNSTONE tail marker.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.4) for what each check evaluates.
