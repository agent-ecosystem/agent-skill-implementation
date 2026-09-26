# Benchmark Skills

These skills are designed to exercise the platform behaviors described in
[checks.md](../checks.md). Each skill contains canary
phrases (unique strings like **CARDINAL-ZEBRA-7742**) that let testers
determine what content the platform loaded and when.

## Skill Inventory

| Skill | Purpose | Files |
|-------|---------|-------|
| `probe-loading` | Core loading behavior: timing, resource enumeration, content presentation, lifecycle | SKILL.md + 3 references + 1 script + 1 asset |
| `probe-linked-resources` | Eager link resolution and path resolution | SKILL.md + 3 references (2 linked, 1 unlinked) |
| `probe-nonstandard-dirs` | Directory recognition and naming divergence | SKILL.md + evals/ + templates/ + resources/ |
| `probe-bulk-files` | Bundled file enumeration at scale: 40 numbered references, a hidden dotfile, a binary PNG, a vendored code tree | SKILL.md, references/ (40 files), .hidden-dotfile-marker.md, assets/, vendor/ |
| `probe-deep-nesting` | Deep resource nesting and nested skill discovery | SKILL.md + references nested 1-5 levels + nested SKILL.md |
| `probe-shadow-alpha` | Cross-skill resource shadowing (pair with beta) | SKILL.md + references/API.md |
| `probe-shadow-beta` | Cross-skill resource shadowing (pair with alpha) | SKILL.md + references/API.md |
| `probe-traversal` | Path traversal boundary enforcement | SKILL.md only (references siblings and parents) |
| `probe-compatibility` | Compatibility field handling | SKILL.md with compatibility field |
| `probe-nonstandard-fields` | Nonstandard frontmatter field handling | SKILL.md with requires, depends-on, priority |
| `probe-metadata-values` | Metadata value edge cases (nulls, empty strings) | SKILL.md with edge-case metadata values |
| `invoke-alpha` | Invocation chain entry point (3-skill chain) | SKILL.md only |
| `invoke-beta` | Invocation chain middle link | SKILL.md only |
| `invoke-gamma` | Invocation chain terminal link | SKILL.md only |
| `probe-circular-alpha` | Circular invocation (pair with beta) | SKILL.md only |
| `probe-circular-beta` | Circular invocation (pair with alpha) | SKILL.md only |
| `probe-missing-dep` | Missing dependency behavior | SKILL.md only (references nonexistent skill) |
| `probe-cross-scope` | Cross-scope dependency resolution | SKILL.md only (references skill at different scope) |
| `probe-mismatch-dir` | Name vs. directory mismatch (frontmatter name: `probe-name-mismatch`) | SKILL.md only |
| `probe-group` | Recursive root discovery: a plain grouping directory (not a skill) containing `probe-grouped/SKILL.md` one level down | probe-grouped/SKILL.md |
| `probe-stray` | Recursive root discovery: a valid skill installed OUTSIDE any skills root | SKILL.md only |
| `probe-malformed-yaml` | Malformed YAML tolerance (unquoted colon in description) | SKILL.md only |
| `probe-no-description` | Missing description handling | SKILL.md only |
| `probe-collision` | Name-collision precedence: PROJECT-scope variant | SKILL.md only |
| `probe-collision-user` | Wrapper holding the USER-scope `probe-collision` variant (install the inner directory at user scope) | probe-collision/SKILL.md |
| `overlay-agents-convention` | Wrapper holding `.agents/skills/probe-interop` for the cross-client interop check (copy its contents onto the project root) | .agents/skills/probe-interop/SKILL.md |
| `probe-multiroot` | Multi-root collision: NATIVE-root variant | SKILL.md only |
| `overlay-multiroot-agents` | Wrapper holding the `.agents/skills/probe-multiroot` variant plus a non-colliding beacon skill that proves the root was scanned (copy its contents onto the project root) | .agents/skills/probe-multiroot/SKILL.md, .agents/skills/probe-multiroot-beacon-agents/SKILL.md |
| `overlay-multiroot-claude` | Wrapper holding the `.claude/skills/probe-multiroot` variant plus a non-colliding beacon skill that proves the root was scanned (copy its contents onto the project root) | .claude/skills/probe-multiroot/SKILL.md, .claude/skills/probe-multiroot-beacon-claude/SKILL.md |
| `probe-script-execution` | Bundled script execution (script assembles its output phrase at runtime) | SKILL.md + scripts/emit-canary.sh |
| `probe-allowed-tools` | Experimental allowed-tools field (pair with control) | SKILL.md with allowed-tools field |
| `probe-allowed-tools-control` | Control twin with no allowed-tools field | SKILL.md only |
| `probe-allowed-tools-lowercase` | Naming twin: allowed-tools value is a bare lowercase `bash` | SKILL.md with allowed-tools field |
| `probe-allowed-tools-shell` | Naming twin: allowed-tools value is a bare `shell` | SKILL.md with allowed-tools field |
| `probe-Upper-Case` | Invalid name: uppercase letters (name matches directory) | SKILL.md only |
| `probe--double-hyphen` | Invalid name: consecutive hyphens (name matches directory) | SKILL.md only |
| `probe-overlong-name-…-limit` | Invalid name: 72 characters, past the 64-char limit (name matches directory) | SKILL.md only |
| `probe-long-description` | Oversize description (1116 chars, head + tail markers) | SKILL.md only |
| `probe-multibyte-description` | Description under 1024 code points but over 1024 UTF-8 bytes (Japanese prose; 848 chars, 1822 bytes; head + tail markers) | SKILL.md only |
| `probe-astral-description` | Description under 1024 code points but over 1024 UTF-16 units and bytes (emoji; 869 chars, 1319 UTF-16 units, 2219 bytes; head + tail markers) | SKILL.md only |
| `probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd` | Name length unit: ASCII name of exactly 64 characters (boundary control) | SKILL.md only |
| `probe-αβγδεζηθικ` | Name length unit: 16-code-point Greek name (16 UTF-16 units, 26 bytes), under the cap in every unit | SKILL.md only |
| `probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ` | Name length unit: 60-code-point Greek name (60 UTF-16 units, 114 bytes), over the cap only in bytes | SKILL.md only |
| `probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡` | Name length unit: 40-code-point astral-plane name (74 UTF-16 units, 142 bytes), over the cap in UTF-16 units and bytes | SKILL.md only |
| `probe-long-compatibility` | Oversize compatibility value (570 chars, tail marker) | SKILL.md only |
| `probe-αβγδεζηθικ`, `probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ`, `probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡` | **Error**: name must be lowercase alphanumeric with hyphens (the validator reads the spec's character rule as ASCII, where the spec's skills-ref accepts any Unicode alphanumeric). On skill-validator 1.6.1 and earlier the two longer names also reported byte lengths (114 and 142) as over the cap; 1.6.2 counts code points and reports no length error for them | Tests name length units and non-ASCII acceptance; the non-ASCII names are the fixture |

## Check-to-Skill Mapping

Each check from checks.md maps to one or more benchmark skills. The
**primary skill** is the one designed specifically for that check. **Secondary
skills** provide additional signal or are needed as part of the test setup.

### Category 1: Loading Timing

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `discovery-reading-depth` | `probe-loading` | Install the skill, start a new session, and ask the model "Do you know the phrase CARDINAL-ZEBRA-7742?" WITHOUT activating the skill. If the model knows it, the platform loaded the full body at discovery time. |
| `activation-loading-scope` | `probe-loading` | Activate the skill and check step 3-4 of its instructions. If the model already has contents of references/, scripts/, or assets/ files without reading them, the platform loaded them at activation. Look for canary phrases PELICAN-MANGO-3391, FALCON-QUARTZ-8819, OSPREY-COBALT-5567, HERON-AMBER-2204, CRANE-TOPAZ-6638. |
| `eager-link-resolution` | `probe-linked-resources` | Activate the skill and check whether the model already has contents of the linked files (PARROT-SILVER-4412, TOUCAN-BRONZE-9931) without reading them. Also check whether the unlinked file (EAGLE-COPPER-1178) was loaded, which distinguishes link-based pre-fetching from bulk directory loading. |

### Category 2: Directory Recognition

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `recognized-directory-set` | `probe-loading` | Activate the skill and check step 3. The skill has all three spec directories. Does the platform enumerate all of them? |
| `directory-naming-divergence` | `probe-nonstandard-dirs` | Activate the skill and check whether `resources/` is treated the same as `references/` would be. Is SWIFT-OPAL-8156 visible or enumerated? |
| `unrecognized-directory-handling` | `probe-nonstandard-dirs` | Activate the skill and check which of the nonstandard directories (evals/, templates/, resources/) the model is aware of. Look for canary phrases ROBIN-JADE-3847, WREN-PEARL-6293, SWIFT-OPAL-8156. |

### Category 3: Resource Access Patterns

| Check | Primary Skill | Secondary | Test Procedure |
|-------|--------------|-----------|----------------|
| `resource-enumeration-behavior` | `probe-loading` |  | Activate the skill. The references/ directory has 3 files (2 linked, 1 unreferenced). Check whether all 3 are enumerated, only the linked ones, or none. |
| `path-resolution-base` | `probe-linked-resources` |  | Activate the skill and have the model try to read files using the relative paths in the SKILL.md. Note what directory the paths resolve against. |
| `cross-skill-resource-shadowing` | `probe-shadow-alpha` | `probe-shadow-beta` | Activate both skills. Have each one read `references/API.md`. Check which canary phrase appears: STORK-CORAL-4471 (alpha) or EGRET-SLATE-8823 (beta). |
| `path-traversal-boundary` | `probe-traversal` |  | Activate the skill and follow its instructions to attempt reads outside the skill directory. |
| `resource-nesting-depth` | `probe-deep-nesting` |  | Activate the skill and follow its instructions to read files at 1, 2, 3, and 5 levels of nesting. Note the deepest level that succeeds. |
| `bundled-script-execution` | `probe-script-execution` |  | Activate the skill and let it run `scripts/emit-canary.sh`. GODWIT-BORNITE-5148 in a tool result proves execution (the script assembles it at runtime; the source never contains the joined phrase). The literal format string `GODWIT-%s-5148` in a tool result means the source was read instead. |
| `bundled-file-enumeration-scale` | `probe-bulk-files` | Activate and check the injected content for file names the body never mentions: bulk-file-01.md through bulk-file-40.md, .hidden-dotfile-marker.md, binary-pixel-marker.png, vendored-lib-marker.js. All present = complete listing; a missing tail = cap; a missing kind = filter; none = no enumeration. |

### Category 4: Content Presentation

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `discovery-listing-fields` | `probe-loading` + `probe-compatibility` + `probe-metadata-values` | Install all three, then WITHOUT activating anything, ask the model to reproduce its available-skills catalog verbatim. Which frontmatter reached it: names and descriptions only, or also the compatibility value ("Designed for Claude Code…"), metadata values (`!!null`), or file paths? |
| `frontmatter-handling` | `probe-loading` | Activate the skill and check step 1. The skill has `allowed-tools`, `compatibility`, and `metadata` fields. If the model can see them, frontmatter was passed through. Also test with `probe-compatibility` for a skill where the compatibility field contains meaningful requirements. |
| `content-wrapping-format` | `probe-loading` | Activate the skill and check step 2. Ask the model to describe how the skill content was presented to it. |
| `activation-location-disclosure` | `probe-loading` | Activate and check whether injected content (not the listing) carries the skill's path, e.g. `skills/probe-loading`. |

### Category 5: Lifecycle Management

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `reactivation-deduplication` | `probe-loading` | Activate the skill, have a conversation, then activate it again. Ask the model if it sees the skill instructions twice in its context. |
| `reactivation-freshness` | `probe-loading` | Activate the skill, edit the SKILL.md file to change the canary phrase, then activate it again in the same session. Ask for the canary phrase to see if the edit was picked up. |
| `context-compaction-protection` | `probe-loading` | Activate the skill, then have a long conversation (enough to trigger context compaction). Ask the model to recall the canary phrase CARDINAL-ZEBRA-7742 and the skill's specific instructions. If it can't, skill content was pruned. |

### Category 6: Access Control

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `trust-gating-behavior` | Any skill | Install any benchmark skill at project level in a freshly cloned or untrusted repository. Start a new session and check whether the skill appears in the available skills list, or if the platform prompts for trust approval. |
| `compatibility-field-behavior` | `probe-compatibility` | Activate the skill and follow its instructions. Also test on a non-Claude platform to see how it handles the "Designed for Claude Code" text. |
| `allowed-tools-behavior` | `probe-allowed-tools` + `probe-allowed-tools-control` | Run each twin in its own session and compare. GROUSE-MICA-9017 in a tool result = the field-bearing skill's printf ran; LAPWING-FLUORITE-2260 = the control's ran. Execution in both sessions means the platform's permission posture, not the field, allowed it. |
| `allowed-tools-name-matching` | `probe-allowed-tools-lowercase` + `probe-allowed-tools-shell` (+ `probe-allowed-tools`) | One session per twin. CHOUGH-TUFA-7180 in a tool result = the lowercase twin's printf ran; ROOK-GABBRO-8841 = the shell twin's; GROUSE-MICA-9017 = the spec-style twin's. A spelling that runs where another is blocked means the field matches platform tool names. |

### Category 7: Skill-to-Skill Invocation

| Check | Primary Skill | Secondary | Test Procedure |
|-------|--------------|-----------|----------------|
| `cross-skill-invocation` | `invoke-alpha` | `invoke-beta` | Activate invoke-alpha. Does it successfully activate invoke-beta? Look for canary phrase TERN-MOSS-6647 in the output. |
| `invocation-depth-limit` | `invoke-alpha` | `invoke-beta`, `invoke-gamma` | Activate invoke-alpha and let the chain run. Does it reach invoke-gamma (JAY-TEAL-9984)? If the chain breaks, at which link? |
| `circular-invocation-handling` | `probe-circular-alpha` | `probe-circular-beta` | Activate probe-circular-alpha. Does the platform detect the circular reference and stop, or does it loop? Look for how many times each canary phrase (KITE-ONYX-2251, WREN-SLATE-7738) appears. |
| `invocation-language-sensitivity` | `invoke-alpha` | `invoke-beta`, `invoke-gamma` | Run the invocation chain test in English, then repeat in another language (e.g., Japanese: "呼び出しチェーンを開始してください"). Compare success rates. |

### Category 8: Skill Dependencies

| Check | Primary Skill | Secondary | Test Procedure |
|-------|--------------|-----------|----------------|
| `informal-dependency-resolution` | `invoke-alpha` | `invoke-beta` | Same as cross-skill-invocation. The invoke chain uses prose instructions to express dependencies between skills. |
| `missing-dependency-behavior` | `probe-missing-dep` |  | Activate the skill. It references `nonexistent-formatter` which doesn't exist. Observe the failure mode. |
| `nonstandard-dependency-fields` | `probe-nonstandard-fields` |  | Activate the skill. It has `requires` and `depends-on` frontmatter fields. Check whether the platform acted on them or ignored them. |
| `cross-scope-dependency` | `probe-cross-scope` | `probe-loading` | Install probe-cross-scope at project level and probe-loading at user level. Activate probe-cross-scope and see if it can invoke probe-loading across scopes. Then remove probe-loading from user level and test again. |

### Category 9: Discovery Scope

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `cross-client-directory-interop` | `overlay-agents-convention` | Copy the wrapper's contents onto the project root so the skill lands at `<project>/.agents/skills/probe-interop/`, and do NOT install it in the platform's native skills directory. Is `probe-interop` listed? Can it activate (SNIPE-OCHRE-2217)? On platforms whose native directory IS `.agents/skills/`, record that instead. |
| `recursive-root-discovery` | `probe-group` + `probe-stray` | Install `probe-group` (with its nested `probe-grouped` skill) into the skills directory, and copy `probe-stray` somewhere in the project OUTSIDE the skills directory. Check the listing for `probe-grouped` (CROW-AGATE-6105) and `probe-stray` (MERLIN-GYPSUM-8852), then activate whichever appeared. |
| `nested-skill-discovery` | `probe-deep-nesting` | Install the skill and check the available skills list. Does `nested-skill` appear as a separate skill? Its SKILL.md is at `probe-deep-nesting/references/nested-skill/SKILL.md`. |
| `name-collision-precedence` | `probe-collision` + `probe-collision-user` | Install `probe-collision` at project scope and `probe-collision-user/probe-collision` at USER scope. Activate `probe-collision`. RAVEN-CITRINE-6634 = project variant won; PIPIT-SHALE-1147 = user variant won. |
| `multi-root-collision-precedence` | `probe-multiroot` + `overlay-multiroot-agents` + `overlay-multiroot-claude` | Install `probe-multiroot` natively and overlay the two convention-root variants (skip the one that coincides with the native root). List, then activate. GREBE-AZURITE-7301 = native root won; BUNTING-SERPENTINE-4185 = .agents root; NIGHTJAR-KYANITE-6072 = .claude root. Descriptions carry `NATIVE-root variant` / `AGENTS-root variant` / `CLAUDE-root variant`. A listed beacon (`probe-multiroot-beacon-agents`, `probe-multiroot-beacon-claude`) proves its root was scanned, so a missing variant was dropped by name. |

### Category 10: Validation Strictness

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `malformed-yaml-tolerance` | `probe-malformed-yaml` | Install normally. Is the skill listed despite the invalid YAML? What description text survived? Activate and look for QUAIL-FELDSPAR-7448. |
| `missing-description-handling` | `probe-no-description` | Install normally. Is the skill listed with no description, a placeholder, or skipped entirely? Activate and look for VIREO-PUMICE-3049. |
| `invalid-name-tolerance` | `probe-Upper-Case` + `probe--double-hyphen` + `probe-overlong-name-…-limit` | Install all three. Check the listing for each (exact or normalized form), then activate each by name. Canaries: DUNLIN-OLIVINE-7821 (uppercase), PETREL-GALENA-3306 (double hyphen), AVOCET-ZIRCON-5573 (overlong). |
| `name-directory-mismatch` | `probe-mismatch-dir` | Install the directory as-is. Check the available skills list: does the skill appear as `probe-name-mismatch` (frontmatter), `probe-mismatch-dir` (directory), or not at all? Then activate it by whichever name appeared and look for SWAN-BERYL-3324. |
| `metadata-value-edge-cases` | `probe-metadata-values` | Activate the skill. If it loads successfully, the platform didn't reject the edge-case metadata. Check step 2-3 to see which values the model received and whether any keys were dropped. Look for canary phrase THRUSH-FLINT-8294 to confirm the body loaded. |
| `oversize-description-handling` | `probe-long-description` | Install normally. Is the skill listed despite the 1116-char description? Does the listing show the head marker SANDERLING-GNEISS-1010 but not the tail marker WHIMBREL-DOLOMITE-2020 (truncation)? Activate and look for BITTERN-HALITE-2264. |
| `description-length-unit` | `probe-long-description` + `probe-multibyte-description` + `probe-astral-description` | Install all three. For each, is it listed, and does the listing show its head marker, its tail marker, both, or neither? ASCII fixture rejected or truncated while both others survive intact = code points; multibyte survives but astral does not = UTF-16 units; none survive = bytes; all three survive = no enforcement. Markers: GANNET-PYRITE-1130 / SHRIKE-TALC-2210 (multibyte), MAGPIE-OBSIDIAN-1240 / LINNET-MALACHITE-2420 (astral). Body canaries: PUFFIN-BASALT-4471 (multibyte), ORIOLE-GRANITE-5583 (astral). |
| `name-length-unit` | `probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd` + `probe-αβγδεζηθικ` + `probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ` + `probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡` + `probe-overlong-name-…-limit` | Install all five and list skills. 72 ASCII listed = no enforcement; 16-code-point Greek skipped = non-ASCII names rejected; both long non-ASCII listed = code points; Greek listed and astral skipped = UTF-16 units; both long non-ASCII skipped = bytes. Description markers `Length-unit probe ascii-sixty-four` / `greek-sixteen` / `greek-sixty` / `math-forty` stand in for names a catalog echo may not reproduce. |
| `oversize-compatibility-handling` | `probe-long-compatibility` | Install normally. Is the skill listed despite the 570-char compatibility value? Activate and look for KESTREL-BAUXITE-6690; note whether the tail marker TURNSTONE-ARAGONITE-3030 surfaces anywhere. |
## Structural Validation

These skills have been validated with
[skill-validator](https://github.com/anthropics/skill-validator) `validate structure`.
All pass except `probe-mismatch-dir`, whose single error is the point of
the fixture (see below). `probe-group/` is a grouping wrapper, not a skill —
validate its inner `probe-group/probe-grouped/` instead. Several fixtures
produce expected warnings or errors because they intentionally use
nonstandard structures to test platform loading behavior:

| Skill | Warnings | Why they're expected |
|-------|----------|----------------------|
| `probe-deep-nesting` | Deep nesting in `references/` | Tests whether platforms follow nested resource paths |
| `probe-linked-resources` | Orphaned files in `references/` and `assets/` | Tests eager link resolution vs. bulk directory loading |
| `probe-nonstandard-dirs` | Unknown directories `evals/`, `resources/`, `templates/` | Tests how platforms handle non-spec directory names |
| `probe-nonstandard-fields` | Unrecognized frontmatter fields `requires`, `depends-on`, `priority` | Tests whether platforms act on or ignore extra fields |
| `probe-metadata-values` | Non-string values in `metadata` (`null`, `~`, `None`, `!!null null`) | Tests whether platforms handle null metadata values gracefully |
| `probe-mismatch-dir` | **Error**: name does not match directory name | Tests which identity platforms use when frontmatter name and directory name disagree — the mismatch is the fixture |
| `probe-malformed-yaml` | **Error**: frontmatter fails strict YAML parsing | Tests parser leniency — the unquoted colon is the fixture |
| `probe-no-description` | **Error**: missing required description | Tests skip-vs-load behavior — the omission is the fixture |
| `probe-Upper-Case` | **Error**: uppercase characters in name | Tests invalid-name tolerance — the case violation is the fixture |
| `probe--double-hyphen` | **Error**: consecutive hyphens in name | Tests invalid-name tolerance — the hyphen violation is the fixture |
| `probe-overlong-name-…-limit` | **Error**: name exceeds 64 characters | Tests invalid-name tolerance — the length violation is the fixture |
| `probe-long-description` | **Error**: description exceeds 1024 characters | Tests oversize-field handling — the overrun is the fixture |
| `probe-multibyte-description` | **Error** on skill-validator 1.6.1 and earlier only: description exceeds 1024 characters | Those versions counted UTF-8 bytes ([issue #94](https://github.com/agent-ecosystem/skill-validator/issues/94)); the description is 848 characters and passes once the validator counts characters |
| `probe-astral-description` | **Error** on skill-validator 1.6.1 and earlier only: description exceeds 1024 characters | Same byte-counting bug; the description is 869 characters |
| `probe-long-compatibility` | **Error**: compatibility exceeds 500 characters | Tests oversize-field handling — the overrun is the fixture |
| `probe-allowed-tools` | Experimental `allowed-tools` field | Tests whether the field pre-approves tools — pair with its control twin |
| `probe-allowed-tools-lowercase`, `probe-allowed-tools-shell` | Experimental `allowed-tools` field | Naming twins for allowed-tools-name-matching |
| `probe-bulk-files` | Unknown directory `vendor/`, orphaned files in `references/` and `assets/`, a dotfile at the skill root | Tests bundled-file enumeration at scale; every file is deliberately unreferenced |

If you run the validator yourself and see only these warnings, everything is
fine. Errors or warnings on other skills would indicate a problem.

## Canary Phrase Index

Each file contains a unique canary phrase. If a tester can identify which
canary phrases the model knows without having explicitly read those files,
it reveals what the platform loaded automatically.

| Canary Phrase | File | Skill |
|---------------|------|-------|
| CARDINAL-ZEBRA-7742 | SKILL.md body | probe-loading |
| PELICAN-MANGO-3391 | references/api-overview.md | probe-loading |
| FALCON-QUARTZ-8819 | references/error-codes.md | probe-loading |
| OSPREY-COBALT-5567 | references/unreferenced-detail.md | probe-loading |
| HERON-AMBER-2204 | scripts/check-status.sh | probe-loading |
| CRANE-TOPAZ-6638 | assets/config-template.yaml | probe-loading |
| THRUSH-FLINT-8294 | SKILL.md body | probe-metadata-values |
| PARROT-SILVER-4412 | references/setup-guide.md | probe-linked-resources |
| TOUCAN-BRONZE-9931 | references/troubleshooting.md | probe-linked-resources |
| EAGLE-COPPER-1178 | references/unlinked-data.md | probe-linked-resources |
| ROBIN-JADE-3847 | evals/evals.json | probe-nonstandard-dirs |
| WREN-PEARL-6293 | templates/output-template.md | probe-nonstandard-dirs |
| SWIFT-OPAL-8156 | resources/api-reference.md | probe-nonstandard-dirs |
| SKUA-DIORITE-2917 | SKILL.md body | probe-bulk-files |
| bulk-file-01.md … bulk-file-40.md, .hidden-dotfile-marker.md, binary-pixel-marker.png, vendored-lib-marker.js | file names only (the body never mentions them); presence in injected content proves enumeration | probe-bulk-files |
| DOVE-GARNET-1029 | references/overview.md | probe-deep-nesting |
| LARK-RUBY-4483 | references/api/endpoints.md | probe-deep-nesting |
| OWL-EMERALD-7756 | references/api/v2/migration-guide.md | probe-deep-nesting |
| FINCH-SAPPHIRE-2098 | references/guides/advanced/performance-tuning.md | probe-deep-nesting |
| PLOVER-JASPER-5590 | references/api/v2/history/deprecated/removed-endpoints.md | probe-deep-nesting |
| HAWK-ONYX-5534 | references/nested-skill/SKILL.md | probe-deep-nesting |
| STORK-CORAL-4471 | references/API.md | probe-shadow-alpha |
| EGRET-SLATE-8823 | references/API.md | probe-shadow-beta |
| IBIS-RUST-3310 | SKILL.md body | invoke-alpha |
| TERN-MOSS-6647 | SKILL.md body | invoke-beta |
| JAY-TEAL-9984 | SKILL.md body | invoke-gamma |
| KITE-ONYX-2251 | SKILL.md body | probe-circular-alpha |
| WREN-SLATE-7738 | SKILL.md body | probe-circular-beta |
| GULL-IRON-4492 | SKILL.md body | probe-missing-dep |
| CRANE-STEEL-1163 | SKILL.md body | probe-cross-scope |
| SWAN-BERYL-3324 | SKILL.md body | probe-mismatch-dir (name: probe-name-mismatch) |
| CROW-AGATE-6105 | probe-grouped/SKILL.md body | probe-group |
| MERLIN-GYPSUM-8852 | SKILL.md body | probe-stray |
| SNIPE-OCHRE-2217 | .agents/skills/probe-interop/SKILL.md body | overlay-agents-convention |
| QUAIL-FELDSPAR-7448 | SKILL.md body | probe-malformed-yaml |
| VIREO-PUMICE-3049 | SKILL.md body | probe-no-description |
| RAVEN-CITRINE-6634 | SKILL.md body (project variant) | probe-collision |
| PIPIT-SHALE-1147 | probe-collision/SKILL.md body (user variant) | probe-collision-user |
| GREBE-AZURITE-7301 | SKILL.md body (native-root variant) | probe-multiroot |
| BUNTING-SERPENTINE-4185 | .agents/skills/probe-multiroot/SKILL.md body (agents-root variant) | overlay-multiroot-agents |
| NIGHTJAR-KYANITE-6072 | .claude/skills/probe-multiroot/SKILL.md body (claude-root variant) | overlay-multiroot-claude |
| DIPPER-LIMONITE-9106 | beacon SKILL.md body (agents root; the beacon's listed name is the scan control) | overlay-multiroot-agents |
| WAGTAIL-SIDERITE-2473 | beacon SKILL.md body (claude root; the beacon's listed name is the scan control) | overlay-multiroot-claude |
| REDSHANK-SYENITE-8807 | SKILL.md body | probe-script-execution |
| GODWIT-BORNITE-5148 | derived: printed by scripts/emit-canary.sh at runtime; never present in any file | probe-script-execution |
| CURLEW-SCHIST-4419 | SKILL.md body | probe-allowed-tools |
| GROUSE-MICA-9017 | derived: printed by the instructed printf at runtime; never present joined in any file | probe-allowed-tools |
| Bash(printf:*) Read | frontmatter allowed-tools value ONLY (the body never spells it out) | probe-allowed-tools |
| STINT-MARBLE-9912 | SKILL.md body | probe-allowed-tools-control |
| LAPWING-FLUORITE-2260 | derived: printed by the instructed printf at runtime; never present joined in any file | probe-allowed-tools-control |
| STILT-SCORIA-5526 | SKILL.md body | probe-allowed-tools-lowercase |
| CHOUGH-TUFA-7180 | derived: printed by the instructed printf at runtime; never present joined in any file | probe-allowed-tools-lowercase |
| AUKLET-CHERT-3364 | SKILL.md body | probe-allowed-tools-shell |
| ROOK-GABBRO-8841 | derived: printed by the instructed printf at runtime; never present joined in any file | probe-allowed-tools-shell |
| DUNLIN-OLIVINE-7821 | SKILL.md body | probe-Upper-Case |
| PETREL-GALENA-3306 | SKILL.md body | probe--double-hyphen |
| AVOCET-ZIRCON-5573 | SKILL.md body | probe-overlong-name-…-limit |
| BITTERN-HALITE-2264 | SKILL.md body | probe-long-description |
| SANDERLING-GNEISS-1010 | SKILL.md description ONLY (head marker; the body never spells it out) | probe-long-description |
| WHIMBREL-DOLOMITE-2020 | SKILL.md description ONLY (tail marker; the body never spells it out) | probe-long-description |
| PUFFIN-BASALT-4471 | SKILL.md body | probe-multibyte-description |
| GANNET-PYRITE-1130 | SKILL.md description ONLY (head marker; the body never spells it out) | probe-multibyte-description |
| SHRIKE-TALC-2210 | SKILL.md description ONLY (tail marker; the body never spells it out) | probe-multibyte-description |
| ORIOLE-GRANITE-5583 | SKILL.md body | probe-astral-description |
| MAGPIE-OBSIDIAN-1240 | SKILL.md description ONLY (head marker; the body never spells it out) | probe-astral-description |
| LINNET-MALACHITE-2420 | SKILL.md description ONLY (tail marker; the body never spells it out) | probe-astral-description |
| Length-unit probe ascii-sixty-four | SKILL.md description ONLY (listing marker; the body never repeats it) | probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd |
| Length-unit probe greek-sixteen | SKILL.md description ONLY (listing marker) | probe-αβγδεζηθικ |
| Length-unit probe greek-sixty | SKILL.md description ONLY (listing marker) | probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ |
| Length-unit probe math-forty | SKILL.md description ONLY (listing marker) | probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡 |
| KESTREL-BAUXITE-6690 | SKILL.md body | probe-long-compatibility |
| TURNSTONE-ARAGONITE-3030 | SKILL.md compatibility ONLY (tail marker; the body never spells it out) | probe-long-compatibility |
