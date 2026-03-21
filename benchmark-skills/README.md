# Benchmark Skills

These skills are designed to exercise the loading behaviors described in
[loading-behavior.md](../loading-behavior.md). Each skill contains canary
phrases (unique strings like **CARDINAL-ZEBRA-7742**) that let testers
determine what content the platform loaded and when.

## Skill Inventory

| Skill | Purpose | Files |
|-------|---------|-------|
| `probe-loading` | Core loading behavior: timing, resource enumeration, content presentation, lifecycle | SKILL.md + 3 references + 1 script + 1 asset |
| `probe-linked-resources` | Eager link resolution and path resolution | SKILL.md + 3 references (2 linked, 1 unlinked) |
| `probe-nonstandard-dirs` | Directory recognition and naming divergence | SKILL.md + evals/ + templates/ + resources/ |
| `probe-deep-nesting` | Deep resource nesting and nested skill discovery | SKILL.md + references nested 1-3 levels + nested SKILL.md |
| `probe-shadow-alpha` | Cross-skill resource shadowing (pair with beta) | SKILL.md + references/API.md |
| `probe-shadow-beta` | Cross-skill resource shadowing (pair with alpha) | SKILL.md + references/API.md |
| `probe-traversal` | Path traversal boundary enforcement | SKILL.md only (references siblings and parents) |
| `probe-compatibility` | Compatibility field handling | SKILL.md with compatibility field |
| `probe-nonstandard-fields` | Nonstandard frontmatter field handling | SKILL.md with requires, depends-on, priority |
| `invoke-alpha` | Invocation chain entry point (3-skill chain) | SKILL.md only |
| `invoke-beta` | Invocation chain middle link | SKILL.md only |
| `invoke-gamma` | Invocation chain terminal link | SKILL.md only |
| `probe-circular-alpha` | Circular invocation (pair with beta) | SKILL.md only |
| `probe-circular-beta` | Circular invocation (pair with alpha) | SKILL.md only |
| `probe-missing-dep` | Missing dependency behavior | SKILL.md only (references nonexistent skill) |
| `probe-cross-scope` | Cross-scope dependency resolution | SKILL.md only (references skill at different scope) |

## Check-to-Skill Mapping

Each check from loading-behavior.md maps to one or more benchmark skills. The
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
| `resource-enumeration-behavior` | `probe-loading` | | Activate the skill. The references/ directory has 3 files (2 linked, 1 unreferenced). Check whether all 3 are enumerated, only the linked ones, or none. |
| `path-resolution-base` | `probe-linked-resources` | | Activate the skill and have the model try to read files using the relative paths in the SKILL.md. Note what directory the paths resolve against. |
| `cross-skill-resource-shadowing` | `probe-shadow-alpha` | `probe-shadow-beta` | Activate both skills. Have each one read `references/API.md`. Check which canary phrase appears: STORK-CORAL-4471 (alpha) or EGRET-SLATE-8823 (beta). |
| `path-traversal-boundary` | `probe-traversal` | | Activate the skill and follow its instructions to attempt reads outside the skill directory. |

### Category 4: Content Presentation

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `frontmatter-handling` | `probe-loading` | Activate the skill and check step 1. The skill has `allowed-tools`, `compatibility`, and `metadata` fields. If the model can see them, frontmatter was passed through. Also test with `probe-compatibility` for a skill where the compatibility field contains meaningful requirements. |
| `content-wrapping-format` | `probe-loading` | Activate the skill and check step 2. Ask the model to describe how the skill content was presented to it. |

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

### Category 7: Structural Edge Cases

| Check | Primary Skill | Test Procedure |
|-------|--------------|----------------|
| `nested-skill-discovery` | `probe-deep-nesting` | Install the skill and check the available skills list. Does `nested-skill` appear as a separate skill? Its SKILL.md is at `probe-deep-nesting/references/nested-skill/SKILL.md`. |
| `resource-nesting-depth` | `probe-deep-nesting` | Activate the skill and follow its instructions to read files at 1, 2, and 3 levels of nesting. Note which depths succeed. |

### Category 8: Skill-to-Skill Invocation

| Check | Primary Skill | Secondary | Test Procedure |
|-------|--------------|-----------|----------------|
| `cross-skill-invocation` | `invoke-alpha` | `invoke-beta` | Activate invoke-alpha. Does it successfully activate invoke-beta? Look for canary phrase TERN-MOSS-6647 in the output. |
| `invocation-depth-limit` | `invoke-alpha` | `invoke-beta`, `invoke-gamma` | Activate invoke-alpha and let the chain run. Does it reach invoke-gamma (JAY-TEAL-9984)? If the chain breaks, at which link? |
| `circular-invocation-handling` | `probe-circular-alpha` | `probe-circular-beta` | Activate probe-circular-alpha. Does the platform detect the circular reference and stop, or does it loop? Look for how many times each canary phrase (KITE-ONYX-2251, WREN-SLATE-7738) appears. |
| `invocation-language-sensitivity` | `invoke-alpha` | `invoke-beta`, `invoke-gamma` | Run the invocation chain test in English, then repeat in another language (e.g., Japanese: "呼び出しチェーンを開始してください"). Compare success rates. |

### Category 9: Skill Dependencies

| Check | Primary Skill | Secondary | Test Procedure |
|-------|--------------|-----------|----------------|
| `informal-dependency-resolution` | `invoke-alpha` | `invoke-beta` | Same as cross-skill-invocation. The invoke chain uses prose instructions to express dependencies between skills. |
| `missing-dependency-behavior` | `probe-missing-dep` | | Activate the skill. It references `nonexistent-formatter` which doesn't exist. Observe the failure mode. |
| `nonstandard-dependency-fields` | `probe-nonstandard-fields` | | Activate the skill. It has `requires` and `depends-on` frontmatter fields. Check whether the platform acted on them or ignored them. |
| `cross-scope-dependency` | `probe-cross-scope` | `probe-loading` | Install probe-cross-scope at project level and probe-loading at user level. Activate probe-cross-scope and see if it can invoke probe-loading across scopes. Then remove probe-loading from user level and test again. |

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
| PARROT-SILVER-4412 | references/setup-guide.md | probe-linked-resources |
| TOUCAN-BRONZE-9931 | references/troubleshooting.md | probe-linked-resources |
| EAGLE-COPPER-1178 | references/unlinked-data.md | probe-linked-resources |
| ROBIN-JADE-3847 | evals/evals.json | probe-nonstandard-dirs |
| WREN-PEARL-6293 | templates/output-template.md | probe-nonstandard-dirs |
| SWIFT-OPAL-8156 | resources/api-reference.md | probe-nonstandard-dirs |
| DOVE-GARNET-1029 | references/overview.md | probe-deep-nesting |
| LARK-RUBY-4483 | references/api/endpoints.md | probe-deep-nesting |
| OWL-EMERALD-7756 | references/api/v2/migration-guide.md | probe-deep-nesting |
| FINCH-SAPPHIRE-2098 | references/guides/advanced/performance-tuning.md | probe-deep-nesting |
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
