---
title: "Codex CLI (headless)"
description: "Automated skill loading findings for Codex CLI (headless) (check list 0.2)."
date: 2026-08-01
showTableOfContents: true
---

| | |
|---|---|
| **Platform** | Codex CLI (headless) |
| **Platform version** | 0.146.0 |
| **Check list version** | 0.2 |
| **Test date** | 2026-08-01 |
| **Model(s) observed** | gpt-5.6-sol |
| **Environment** | Headless invocation via benchmark-runner + skillxp |

> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors they are point observations, not rates. Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says "not exercised" — automation does not probe recovery, so absence of a fallback observation is not evidence that none exists.

## Loading Timing

### `discovery-reading-depth`

_Does the harness read only SKILL.md metadata at discovery, or the full body?_

- **Status**: observed
- **Verdict**: Metadata only (`metadata-only`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-loading
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `activation-loading-scope`

_On activation, does the harness load only the SKILL.md body, or also bundled resources — and by which vehicle?_

- **Status**: observed
- **Verdict**: Body only (`body-only`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - model's tool call targets the skill's own path
  - body canary arrived in the tool result
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `eager-link-resolution`

_Does activation pre-fetch files markdown-linked from the SKILL.md body — and does that extend to a file mentioned only as plain text?_

- **Status**: observed
- **Verdict**: No pre-fetching (`no-prefetch`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - skill body loaded
  - references/setup-guide.md arrived only via the model's own read
  - references/troubleshooting.md arrived only via the model's own read
- **Note**: model read [references/setup-guide.md references/troubleshooting.md] itself, corroborating it did not already have them
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Directory Recognition

### `recognized-directory-set`

_Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration-at-activation`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `directory-naming-divergence`

_Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`resources-untouched`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `unrecognized-directory-handling`

_What happens to directories the spec never named (evals/, templates/) — injected, readable on demand, or invisible?_

- **Status**: observed
- **Verdict**: Not surfaced; model never looked (`untouched`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses
- **Note**: final answer: Probe results:  1. Directory awareness: `evals/`, `templates/`, and `resources/`.  2. Content visibility:     - `evals/` → `evals.json`    - `templates/` → `output-template.md`    - `resources/` → `api-reference.md`     The skill instructions named the directories, but listings required direct filesystem inspection. File contents were not automatically visible or loaded.  3. `resources/` vs. `references/`: `resources/` received no special treatment. Its content was neither enumerated nor loaded automatically; it behaved as an ordinary, initially ignored directory.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Resource Access Patterns

### `resource-enumeration-behavior`

_At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?_

- **Status**: observed
- **Verdict**: Nothing enumerated (`no-enumeration`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: the unlinked file's name never reached the model; only body-linked files are discoverable without exploration
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `path-resolution-base`

_When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against — and does the bare path work as written?_

- **Status**: observed
- **Verdict**: Model used full paths (base untested) (`model-preemptively-qualified`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Runs**: 3 (cwd-base-model-requalified ×1; model-preemptively-qualified ×2)
- **Note**: the model never tried the bare path, so the platform's resolution base was not directly exercised; it navigated by qualified path from the start
- **Note**: verdicts varied across 3 runs (model-level signal); see verdict_counts and the per-run findings
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `cross-skill-resource-shadowing`

_With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?_

- **Status**: observed
- **Verdict**: Got its own file (`own-resource-first`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - activated skill's own API.md content arrived first
- **Note**: every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised — the outcome reflects model path discipline, not platform disambiguation
- **Note**: final answer: `probe-shadow-alpha/references/API.md` contains:  - Canary: **STORK-CORAL-4471** — confirms the alpha version. - Endpoints:   - `GET /alpha/status`   - `POST /alpha/process`  Only alpha was activated, so no beta access test was required.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `path-traversal-boundary`

_Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?_

- **Status**: observed
- **Verdict**: Reads outside the skill allowed (`outside-skill-read-allowed`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - sibling skill's body canary arrived in a tool result
  - first ../ path attempt
- **Note**: relative ../ path forms were used
- **Note**: final answer: Probe results:  - Sibling skill: **Allowed**. Full contents of `../probe-loading/SKILL.md` were returned without a platform warning. - Parent directory: **Allowed**. `../README.md` was not found, but listing `../` succeeded and exposed both skill directories. - Two levels up: Read attempt was allowed, but `../../loading-behavior.md` was **not found**. - Silent empty responses: **None**. - Incidental warning: shell startup reported an unrelated `pyenv` rehash permission warning.  Conclusion: no traversal boundary was enforced for sibling or parent access.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Content Presentation

### `discovery-listing-fields`

_Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?_

- **Status**: observed
- **Verdict**: Also surfaces location (`surfaces-beyond-description:[location]`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names the installed skills
  - description text surfaced at discovery
  - SKILL.md file path surfaced at discovery
- **Note**: final answer: - imagegen: Generate or edit raster images when the task benefits from AI-created bitmap visuals such as photos, illustrations, textures, sprites, mockups, or transparent-background cutouts. Use when Codex should create a brand-new image, transform an existing image, or derive visual variants from references, and the output should be a bitmap asset rather than repo-native code or vector. Do not use when the task is better handled by editing existing SVG/vector/code-native assets, extending an established icon or logo system, or building the visual directly in HTML/CSS/canvas. (file: /private/var/folders/39/30d_262j423_n546m9g8_hc80000gn/T/skillxp-2311220095/home/codex/skills/.system/imagegen/SKILL.md) - openai-docs: Use when the user asks how to build with OpenAI products or APIs, asks about Codex itself or choosing Codex surfaces, needs up-to-date official documentation with citations, help choosing the latest model for a use case, latest/current/default-model prompting guidance, or model upgrade and prompt-upgrade guidance; use OpenAI docs MCP tools for non-Codex docs questions, use the Codex manual helper first for broad Codex self-knowledge, and restrict fallback browsing to of…
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `frontmatter-handling`

_Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?_

- **Status**: observed
- **Verdict**: Visible (model reads the raw file) (`frontmatter-visible-via-raw-read`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - raw file read delivered frontmatter and body together
- **Note**: pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `metadata-value-edge-cases`

_Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded — and do those keys reach the model?_

- **Status**: observed
- **Verdict**: Loaded fine (`loaded-despite-edge-case-metadata`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-metadata-values despite edge-case metadata
  - body canary loaded
  - frontmatter-only marker '!!null' in a tool result
- **Note**: edge-case metadata values reached the model only via its own raw file read
- **Note**: final answer: Canary: **THRUSH-FLINT-8294**  - Load success: Yes. - Metadata visibility: Yes, via the raw frontmatter. - Keys preserved: All seven are present. - Observed values:   - `normal`: `"baseline"`   - `empty-double`: `""`   - `empty-single`: `''`   - `explicit-null`: `null`   - `tilde-null`: `~`   - `none-null`: `None`   - `tagged-null`: `!!null null` - Platform behavior: The skill loaded without observable warnings, errors, rejected values, or stripped keys. The exposed skill catalog omitted the metadata details, so no platform-level coercion could be observed.
- **Fallback behavior**: Observed in-run: edge-case metadata values reached the model only via its own raw file read

### `content-wrapping-format`

_Is injected skill content wrapped in structured tags, or delivered as raw markdown — and what does the model see on pull harnesses?_

- **Status**: observed
- **Verdict**: Raw file via model read (`raw-file-via-pull`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - body arrived as a file-read tool result
- **Note**: pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Lifecycle Management

### `reactivation-deduplication`

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

### `reactivation-freshness`

_After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?_

- **Status**: observed
- **Verdict**: Edits picked up immediately (`fresh-content-served`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - post-edit canary reached the model on reactivation
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `context-compaction-protection`

_Is skill content protected when the context window fills up?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

## Access Control

### `trust-gating-behavior`

_Do project-level skills require trust approval before loading?_

- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)

### `compatibility-field-behavior`

_Does a compatibility field naming another platform gate loading, get surfaced to the model, or get ignored?_

- **Status**: observed
- **Verdict**: Loads normally, no gating (`activated-no-gating`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-compatibility despite the foreign-platform compatibility value
  - body loaded normally
- **Note**: compatibility value visible only via the model's own raw file read
- **Note**: final answer: 1. **Field visibility:** Yes. Value: “Designed for Claude Code (or similar products). Requires Python 3.14+ and network access.”  2. **Platform behavior:** Activated normally—no warning, confirmation, or automatic requirement check.  3. **Interpretation:** Requirements are only partially met. Python 3.14.3 is installed, but network access is restricted, and the current platform is Codex rather than Claude Code.
- **Fallback behavior**: Observed in-run: compatibility value visible only via the model's own raw file read

## Structural Edge Cases

### `nested-skill-discovery`

_Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?_

- **Status**: observed
- **Verdict**: Discovered as a separate skill (`nested-skill-discovered`)
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names nested-skill (found inside probe-deep-nesting/references/)
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `resource-nesting-depth`

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

### `name-directory-mismatch`

_When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?_

- **Status**: observed
- **Verdict**: Listed under both names (`listed-under-both`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing carries the FRONTMATTER name probe-name-mismatch
  - discovery listing carries the DIRECTORY name probe-mismatch-dir
- **Note**: activation loads: by frontmatter name=true, by directory name=false (a load proves reachability, not catalog identity — the model may map either prompt to the installed skill or read the file directly)
- **Note**: caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal
- **Note**: final answer: `probe-mismatch-dir` is not an available skill; the catalog lists only `probe-name-mismatch`. I can’t activate it under the directory name.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `recursive-root-discovery`

_Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?_

- **Status**: observed
- **Verdict**: Scans the root recursively; stray file ignored (`recursive-scan; stray:not-discovered`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-grouped from one level below the root
  - grouped skill's body loaded on activation
- **Note**: final answer: probe-grouped activated. Canary: **CROW-AGATE-6105**  **Discovery**: Confirmed. This skill appears in the available skills catalog despite sitting one level below the skills root.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Skill-to-Skill Invocation

### `cross-skill-invocation`

_Can one skill's instructions get a second installed skill activated by name?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-alpha complete.   Chain result: invoke-beta activated with canary **TERN-MOSS-6647**, but invoke-gamma was unavailable, so beta could not complete.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `invocation-depth-limit`

_How deep does a skill→skill→skill activation chain run before something stops it?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: final answer: invoke-beta complete.   invoke-gamma said: Canary **JAY-TEAL-9984**; the chain completed successfully.  invoke-alpha complete.   Chain result: Beta reported **TERN-MOSS-6647**, and Gamma reported **JAY-TEAL-9984**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `circular-invocation-handling`

_When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?_

- **Status**: observed
- **Verdict**: Model stopped the loop itself (`cycle-stopped-model-choice`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Note**: loads: alpha=1 beta=1; tool references: alpha=1 beta=1
- **Note**: beta activated but the model never re-attempted alpha (model-level stop)
- **Note**: final answer: Both skills activated successfully. `probe-circular-beta` referred back to the already-active alpha skill, so circular re-entry was detected and prevented. No infinite loop occurred.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `invocation-language-sensitivity`

_Does the invoke chain still complete when the activation prompt is Japanese?_

- **Status**: observed
- **Verdict**: Full three-skill chain completed (`chain-completed-depth-3`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded
  - invoke-beta body loaded
  - invoke-gamma body loaded
- **Note**: single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates — agentskills#95 reports ~10% Japanese-prompt failures on Claude Code
- **Note**: final answer: invoke-alpha complete.   Chain result: invoke-beta（**TERN-MOSS-6647**）と invoke-gamma（**JAY-TEAL-9984**）が起動し、3スキルの連鎖が正常に完了しました。
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Skill Dependencies

### `informal-dependency-resolution`

_Is a dependency expressed only in prose ("now activate the invoke-beta skill") resolved at runtime?_

- **Status**: observed
- **Verdict**: Second skill activated (`second-skill-loaded`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - invoke-alpha body loaded (chain entry)
  - invoke-beta's body canary loaded after invoke-alpha's instruction
- **Note**: invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here
- **Note**: final answer: invoke-beta complete.   invoke-gamma said: unavailable—the skill is not installed or listed.  invoke-alpha complete.   Chain result: invoke-beta activated with canary **TERN-MOSS-6647**, but invoke-gamma could not run. Alpha canary: **IBIS-RUST-3310**.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `missing-dependency-behavior`

_When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?_

- **Status**: observed
- **Verdict**: Reported missing without attempting (`reported-without-attempt`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-missing-dep body loaded
  - model surfaced the missing dependency in its own text without attempting it
- **Note**: which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs
- **Note**: final answer: The platform reported that `nonexistent-formatter` does not exist. I did not silently skip it, imitate it, or encounter a tool error.  probe-missing-dep complete. Dependency resolution result: the required skill was unavailable, so activation stopped at that dependency.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `nonstandard-dependency-fields`

_Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?_

- **Status**: observed
- **Verdict**: Ignored (`fields-ignored`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - probe-nonstandard-fields body loaded
- **Note**: final answer: 1. **Field visibility:** Yes.     - `requires`: `probe-loading`    - `depends-on`: `probe-shadow-alpha`, `probe-shadow-beta`    - `priority`: `high`  2. **Platform behavior:** The fields were silently ignored. No dependencies were automatically loaded, and no priority-based behavior was evident.  3. **Validation:** The platform accepted the nonstandard fields without warnings.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `cross-scope-dependency`

_Can a project-level skill invoke a dependency that exists only at user level — and what is the failure mode when it is absent?_

- **Status**: observed
- **Verdict**: Resolved across scopes; missing dependency reported, not attempted (`resolved-across-scopes; missing:reported-without-attempt`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - user-scope dependency's body canary pulled during project-skill session
  - model surfaced the missing dependency in its own text without attempting it
- **Note**: with-dependency final answer: ### probe-loading report  - **Frontmatter visibility:** Visible fields: `name`, `description`, `allowed-tools`, `compatibility`, and `metadata`. - **Content wrapping:** None; presented as raw Markdown with YAML frontmatter. - **Resource awa…
- **Note**: which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs
- **Note**: without-dependency final answer: probe-cross-scope activated. Canary: **CRANE-STEEL-1163**  Installation scope: **project-level**, as shown by its project `.codex/skills` location.  `probe-loading` could not be activated because it was absent from the platform’s available-skills catalog. Its scope therefore could not be determined.  probe-cross-scope complete. Cross-scope resolution result: **The project-level skill loaded, but its expected user-level dependency was not found.**
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

## Discovery and Validation

### `cross-client-directory-interop`

_Is a skill installed only at the cross-client .agents/skills convention path discovered?_

- **Status**: observed
- **Verdict**: Convention path scanned (`convention-scanned`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-interop
  - body canary loaded on activation
- **Note**: final answer: probe-interop activated. Canary: **SNIPE-OCHRE-2217**  Discovery: Yes, it appeared in the available skills catalog at `.agents/skills/probe-interop/SKILL.md`.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `malformed-yaml-tolerance`

_Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?_

- **Status**: observed
- **Verdict**: Tolerated and loaded (`tolerated-and-loaded`)
- **Vehicle**: model-pull
- **Confidence**: transcript-direct
- **Evidence**:
  - discovery listing names probe-malformed-yaml
  - body canary loaded on activation
- **Note**: final answer: probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**  Description visibility: `Use when: asked to probe malformed yaml parsing behavior` — intact, not truncated at the colon.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `missing-description-handling`

_Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?_

- **Status**: observed
- **Verdict**: Skipped (as the guide prescribes) (`skipped-as-guide-prescribes`)
- **Confidence**: transcript-direct
- **Note**: final answer: The `probe-no-description` skill isn’t available in this session, so I can’t activate it.
- **Fallback behavior**: Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent.

### `name-collision-precedence`

_With the same skill name installed at project and user scope, which variant's content activates?_

- **Status**: observed
- **Verdict**: Lists both; model chose the project variant (`catalog-lists-both-project-overrides-user`)
- **Vehicle**: model-pull
- **Confidence**: behavioral-inference
- **Evidence**:
  - project variant's canary loaded; user variant's never appeared
- **Note**: the discovery listing exposed BOTH variants (both descriptions present); the model's file choice determined the winner — model-level selection, not platform precedence
- **Note**: final answer: probe-collision activated. Canary: **RAVEN-CITRINE-6634** (project variant)  **Precedence:** Project scope won.
- **Fallback behavior**: the shadowed variant stays reachable: the catalog exposes both entries, so a user (or the model) can address either by path.

---

Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version 0.2) for what each check evaluates.
