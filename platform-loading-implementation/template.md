# Platform Loading Behavior: [Platform Name]

<!--
  Copy this template to a new file named after the platform (e.g., claude-code.md,
  github-copilot.md, cursor.md) and fill in your findings.

  You don't need to test every check in one session. Partial results are valuable.
  Leave sections you haven't tested as "Not tested."
-->

## Platform Details

| Field | Value |
|-------|-------|
| **Platform** | <!-- e.g., Claude Code, GitHub Copilot, Cursor, Roo Code --> |
| **Platform version** | <!-- e.g., 1.0.20, VS Code 1.107 + Copilot Chat 0.24 --> |
| **Date tested** | <!-- YYYY-MM-DD. Implementation details change; this is a snapshot. --> |
| **Model used** | <!-- e.g., Claude Sonnet 4.6, GPT-4o, Gemini 2.5 Pro. Note the specific model, not just the family. --> |
| **Tester** | <!-- Your name or GitHub handle --> |

### A note on models

Some behaviors may vary by model *within* the same platform harness. For example,
a platform that lets the user choose between Claude Sonnet and Claude Opus may
produce different skill-loading behavior depending on which model is active,
because the model itself makes activation decisions. If you test with multiple
models, either create separate files for each model or note model-specific
differences inline.

### Platform-level vs. model-level behavior

When recording results, try to distinguish between:

- **Platform-level behavior**: Enforced by the harness (deterministic). Example:
  the platform strips frontmatter before passing content to the model. This won't
  vary by model or across runs.
- **Model-level behavior**: Determined by the model's interpretation of instructions
  (probabilistic). Example: the model decides whether to follow a markdown link and
  read the referenced file. This may vary by model, prompt language, or even across
  runs with the same model.

This distinction matters because platform-level behaviors are stable and predictable,
while model-level behaviors may need multiple test runs to characterize and may
change when the user switches models.

## Methodology

<!--
  Describe how you tested. This section helps others reproduce your findings
  and judge whether observed behaviors are platform-level or model-level.

  The benchmark skills in benchmark-skills/ are designed for these tests. See
  benchmark-skills/README.md for the check-to-skill mapping and test procedures.

  Include:

  - Whether you ran each check in a separate session (recommended) or multiple
    checks per session. Prior checks leave context that can contaminate later
    results — e.g., the agent may "know" a canary phrase from an earlier check,
    not because the platform loaded it for the current check.
  - Which benchmark skills you used (or link to custom skills if you used your own)
  - How skills were installed (project-level, user-level, or both)
  - The prompts you used to trigger skill activation
  - How you observed what was loaded (e.g., inspecting context via debug tools,
    reading platform logs, observing model behavior, checking canary phrases)
  - How many times you repeated each test (for model-dependent behaviors,
    single observations may not be representative)
  - Any platform settings or configuration that may have influenced results
    (e.g., trust settings, permission modes, skill-loading toggles)
-->

## Results

For each check, record what you observed. Use one of these statuses:

- **Observed**: You tested this and have a finding.
- **Inconclusive**: You tested this but the result was ambiguous or inconsistent
  across runs.
- **Not tested**: You haven't tested this check yet.

### About fallback behavior

Each check includes a **Fallback behavior** field. This captures what happens
when the platform's default behavior doesn't surface content to the model. I
hypothesize there may be three patterns:

- **Agent self-recovers**: The agent independently realizes it needs the content
  and uses a file-read tool or other mechanism to access it without user
  intervention. This may be hard to distinguish from platform-level loading; note
  whether you observed the agent making an explicit tool call to read the file.
- **User prompt required**: The user must explicitly instruct the agent (e.g.,
  "read the file at references/api-overview.md") to access the content. The
  content is accessible but only with manual intervention.
- **No fallback**: The content is truly inaccessible through any mechanism. The
  platform blocks access or the agent has no tool capable of reaching it.

This matters for skill authors writing portable skills. If a skill's references
are invisible on a platform due to a closed directory set, but the user can work
around it with an explicit prompt, the skill author can include a note like
"If your agent doesn't automatically load the reference files, ask it to read
references/api-overview.md." If there's no fallback, the skill simply doesn't
work on that platform.

---

### Category 1: Loading Timing

#### `discovery-reading-depth`

- **Benchmark skill**: `probe-loading` — Install the skill, start a new session, and ask the model "Do you know the phrase CARDINAL-ZEBRA-7742?" WITHOUT activating the skill. If the model knows it, the platform loaded the full body at discovery time.
- **Status**: Not tested
- **Observation**: <!-- Does the platform read only frontmatter at discovery, or the entire SKILL.md? -->
- **Evidence**: <!-- How did you determine this? (e.g., debug logs, token counter, canary phrase test) -->
- **Platform-level or model-level?**: <!-- Platform-level (the harness decides what to read) -->
- **Fallback behavior**: <!-- If the body was loaded at discovery, this check is about context cost, not access. Note whether the platform provides a way to defer loading (e.g., a lazy-load setting). -->

#### `activation-loading-scope`

- **Benchmark skill**: `probe-loading` — Activate the skill and check steps 3-4 of its instructions. If the model already has contents of files in references/, scripts/, or assets/ without reading them, the platform loaded them at activation. Look for canary phrases: PELICAN-MANGO-3391, FALCON-QUARTZ-8819, OSPREY-COBALT-5567, HERON-AMBER-2204, CRANE-TOPAZ-6638.
- **Status**: Not tested
- **Observation**: <!-- When a skill activates, does the platform load only the SKILL.md body, or also supporting files? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Could be either. The harness may eagerly load files (platform-level), or the model may independently read them (model-level). Try to distinguish. -->
- **Fallback behavior**: <!-- If supporting files were NOT loaded at activation, can the agent read them on its own when it encounters a reference in the SKILL.md body? Or does the user need to explicitly say "read references/api-overview.md"? -->

#### `eager-link-resolution`

- **Benchmark skill**: `probe-linked-resources` — Activate the skill and check whether the model already has contents of the linked files (PARROT-SILVER-4412, TOUCAN-BRONZE-9931) without reading them. Also check whether the unlinked file (EAGLE-COPPER-1178) was loaded, which distinguishes link-based pre-fetching from bulk directory loading.
- **Status**: Not tested
- **Observation**: <!-- Does the platform pre-fetch files linked in the SKILL.md body at activation time? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- If the harness pre-fetches, it's platform-level. If the model chooses to follow links, it's model-level. The distinction matters: platform-level pre-fetching is deterministic; model-level link-following varies by model and run. -->
- **Fallback behavior**: <!-- If linked files were NOT pre-fetched, does the agent independently follow the markdown links when it reaches them in the instructions? Or does the user need to prompt it? -->

---

### Category 2: Directory Recognition

#### `recognized-directory-set`

- **Benchmark skill**: `probe-loading` — Activate the skill and check step 3. The skill has all three spec directories (scripts/, references/, assets/). Does the platform enumerate all of them?
- **Status**: Not tested
- **Observation**: <!-- Does the platform enumerate scripts/, references/, assets/ specifically? Does it treat them as a closed or open set? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the platform treats directories as a closed set, can the agent still access files in those directories via explicit file-read tool calls? Or are unenumerated directories completely invisible? -->

#### `directory-naming-divergence`

- **Benchmark skill**: `probe-nonstandard-dirs` — Activate the skill and check whether `resources/` is treated the same as `references/` would be. Is the canary phrase SWIFT-OPAL-8156 visible or enumerated?
- **Status**: Not tested
- **Observation**: <!-- Does the platform use the spec's directory names, or different names (e.g., resources/ instead of references/)? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the platform doesn't recognize resources/, can the user instruct the agent to read files from it by path? Does the agent self-recover by browsing the skill directory? -->

#### `unrecognized-directory-handling`

- **Benchmark skill**: `probe-nonstandard-dirs` — Activate the skill and check which of the nonstandard directories (evals/, templates/, resources/) the model is aware of. Look for canary phrases: ROBIN-JADE-3847, WREN-PEARL-6293, SWIFT-OPAL-8156.
- **Status**: Not tested
- **Observation**: <!-- What happens with a directory not in the spec (e.g., evals/, templates/)? Ignored? Loaded? Listed? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Could be either. The harness may ignore unknown dirs (platform-level), or the model may browse into them (model-level). -->
- **Fallback behavior**: <!-- If the platform ignores unrecognized directories, can the user instruct the agent to read files from them directly? Does the agent ever discover them on its own (e.g., via directory listing)? -->

---

### Category 3: Resource Access Patterns

#### `resource-enumeration-behavior`

- **Benchmark skill**: `probe-loading` — Activate the skill. The references/ directory has 3 files (2 linked from SKILL.md, 1 unreferenced). Check whether all 3 are enumerated, only the linked ones, or none. The unreferenced file's canary phrase is OSPREY-COBALT-5567.
- **Status**: Not tested
- **Observation**: <!-- For a references/ directory with multiple files, does the platform enumerate all files, load all contents, present a listing, or ignore until the model reads one? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Enumeration is platform-level. Whether the model then reads listed files is model-level. -->
- **Fallback behavior**: <!-- If files are not enumerated, can the agent discover them by listing the directory? Or does the user need to tell the agent which files exist? -->

#### `path-resolution-base`

- **Benchmark skill**: `probe-linked-resources` — Activate the skill and have the model try to read files using the relative paths in the SKILL.md. Note what directory the paths resolve against.
- **Status**: Not tested
- **Observation**: <!-- When SKILL.md references a relative path, does the platform resolve against the skill directory, working directory, or project root? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If relative paths don't resolve correctly, can the user provide an absolute path as a workaround? Does the agent try alternative resolution bases on its own? -->

#### `cross-skill-resource-shadowing`

- **Benchmark skills**: `probe-shadow-alpha` + `probe-shadow-beta` — Activate both skills. Have each one read `references/API.md`. Check which canary phrase appears: STORK-CORAL-4471 (alpha) or EGRET-SLATE-8823 (beta).
- **Status**: Not tested
- **Observation**: <!-- If two active skills both have references/API.md, which one does the model get? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level (path resolution) but model-level factors may influence which skill's context the model is "thinking in" when it requests the file. -->
- **Fallback behavior**: <!-- Can the user disambiguate by providing the full path (e.g., "read probe-shadow-alpha/references/API.md")? Does the agent attempt disambiguation on its own? -->

#### `path-traversal-boundary`

- **Benchmark skill**: `probe-traversal` — Activate the skill and follow its instructions to attempt reads outside the skill directory (../probe-loading/SKILL.md, ../README.md, ../../loading-behavior.md).
- **Status**: Not tested
- **Observation**: <!-- Can the model access files outside the skill directory via relative paths? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level (a security boundary should be enforced by the harness). -->
- **Fallback behavior**: <!-- N/A for security checks. If traversal is blocked, that's the desired behavior. Note whether the agent reports the block or silently fails. -->

---

### Category 4: Content Presentation

#### `frontmatter-handling`

- **Benchmark skills**: `probe-loading`, `probe-compatibility` — Activate the skill and check whether the model can see frontmatter fields (allowed-tools, compatibility, metadata). probe-loading has multiple optional fields; probe-compatibility has a compatibility field with meaningful requirements text.
- **Status**: Not tested
- **Observation**: <!-- Does the model see the full SKILL.md including frontmatter, or only the body? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If frontmatter is stripped, can the user instruct the agent to re-read the raw SKILL.md file to see the frontmatter? Is there a platform setting to change stripping behavior? -->

#### `content-wrapping-format`

- **Benchmark skill**: `probe-loading` — Activate the skill and check step 2. Ask the model to describe how the skill content was presented to it (raw markdown, XML tags, JSON, etc.).
- **Status**: Not tested
- **Observation**: <!-- Is skill content wrapped in structured tags (XML, JSON, etc.) or injected as raw markdown? If wrapped, what format? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- N/A. Wrapping is informational; there is no "failure" to fall back from. Note whether the wrapping format helps or hinders the model's ability to follow skill instructions. -->

---

### Category 5: Lifecycle Management

#### `reactivation-deduplication`

- **Benchmark skill**: `probe-loading` — Activate the skill, have a conversation, then activate it again. Ask the model if it sees the skill instructions twice in its context.
- **Status**: Not tested
- **Observation**: <!-- If the model activates the same skill twice, does the platform deduplicate or inject again? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the platform does not deduplicate, is there a user action that avoids the double injection (e.g., telling the agent "you already have this skill loaded")? -->

#### `reactivation-freshness`

- **Benchmark skill**: `probe-loading` — Activate the skill, then edit the SKILL.md to change the canary phrase from CARDINAL-ZEBRA-7742 to something else. Activate the skill again in the same session and ask for the canary phrase.
- **Status**: Not tested
- **Observation**: <!-- If SKILL.md is edited mid-session and re-activated, does the platform pick up changes? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the platform caches and doesn't pick up changes, can the user force a refresh (e.g., restarting the session, using a reload command)? -->

#### `context-compaction-protection`

- **Benchmark skill**: `probe-loading` — Activate the skill, then have a long conversation (enough to trigger context compaction). Ask the model to recall the canary phrase CARDINAL-ZEBRA-7742 and the skill's specific instructions.
- **Status**: Not tested
- **Observation**: <!-- In a long conversation, does the platform protect skill content from context pruning? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level (compaction is a harness decision). -->
- **Fallback behavior**: <!-- If skill content is pruned, can the user re-activate the skill to restore it? Does the agent recognize that it lost the skill instructions and re-activate on its own? -->

---

### Category 6: Access Control

#### `trust-gating-behavior`

- **Benchmark skill**: Any benchmark skill — Install at project level in a freshly cloned or untrusted repository. Start a new session and check whether the skill appears in the available skills list, or if the platform prompts for trust approval.
- **Status**: Not tested
- **Observation**: <!-- Does the platform require trust approval for project-level skills? What happens if trust is not granted? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If trust is not granted, is there a way for the user to grant it mid-session? Is there a platform setting to pre-trust certain directories? -->

#### `compatibility-field-behavior`

- **Benchmark skill**: `probe-compatibility` — Activate the skill and follow its instructions. The skill's compatibility field says "Designed for Claude Code (or similar products). Requires Python 3.14+ and network access." Test on a non-Claude platform to see how it handles the Claude-specific text.
- **Status**: Not tested
- **Observation**: <!-- How does the platform handle the compatibility field? Does it gate loading, surface it to the model, or ignore it? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level (whether the harness acts on it) and model-level (whether the model interprets it if passed through). -->
- **Fallback behavior**: <!-- If the platform blocks loading based on compatibility, can the user override? If the model self-restricts based on the compatibility text, can the user instruct it to proceed anyway? -->

---

### Category 7: Structural Edge Cases

#### `nested-skill-discovery`

- **Benchmark skill**: `probe-deep-nesting` — Install the skill and check the available skills list. Does `nested-skill` appear as a separate skill? Its SKILL.md is at `probe-deep-nesting/references/nested-skill/SKILL.md`. Canary phrase: HAWK-ONYX-5534.
- **Status**: Not tested
- **Observation**: <!-- If a SKILL.md exists inside another skill's directory tree, does the platform discover both? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the nested skill is not discovered, can the user manually activate it by path? If it IS discovered as a separate skill, can the user suppress it? -->

#### `resource-nesting-depth`

- **Benchmark skill**: `probe-deep-nesting` — Activate the skill and follow its instructions to read files at 1 level (DOVE-GARNET-1029), 2 levels (LARK-RUBY-4483), and 3 levels (OWL-EMERALD-7756, FINCH-SAPPHIRE-2098) of nesting.
- **Status**: Not tested
- **Observation**: <!-- Can the model access deeply nested resource files? At what depth does it fail? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level (whether the harness enumerates nested files) and model-level (whether the model attempts to traverse deeper). -->
- **Fallback behavior**: <!-- If deeply nested files aren't enumerated, can the agent access them via explicit path? Does the user need to provide the full path, or can the agent navigate the directory tree? -->

---

### Category 8: Skill-to-Skill Invocation

#### `cross-skill-invocation`

- **Benchmark skills**: `invoke-alpha` + `invoke-beta` — Activate invoke-alpha. Does it successfully activate invoke-beta? Look for canary phrase TERN-MOSS-6647 in the output.
- **Status**: Not tested
- **Observation**: <!-- Can a skill's instructions direct the model to activate a different installed skill? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Both. The harness must support the activation mechanism, but the model must choose to invoke it. Test with multiple models if possible. -->
- **Fallback behavior**: <!-- If the agent doesn't invoke the second skill on its own, can the user prompt it ("now activate invoke-beta")? Does that workaround consistently succeed? -->

#### `invocation-depth-limit`

- **Benchmark skills**: `invoke-alpha` + `invoke-beta` + `invoke-gamma` — Activate invoke-alpha and let the full chain run. Does it reach invoke-gamma? Look for canary phrase JAY-TEAL-9984. If the chain breaks, note which link failed.
- **Status**: Not tested
- **Observation**: <!-- Is there a limit on chained skill invocations (A -> B -> C)? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Could be either. A harness-enforced limit is platform-level. The model choosing to stop chaining is model-level. -->
- **Fallback behavior**: <!-- If the chain breaks at a certain depth, can the user manually continue it by activating the next skill? Is the limit configurable? -->

#### `circular-invocation-handling`

- **Benchmark skills**: `probe-circular-alpha` + `probe-circular-beta` — Activate probe-circular-alpha. Does the platform detect the circular reference and stop, or does it loop? Count how many times each canary phrase appears (KITE-ONYX-2251, WREN-SLATE-7738).
- **Status**: Not tested
- **Observation**: <!-- What happens with circular skill references (A invokes B, B invokes A)? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Both. The harness may enforce a guard, or the model may detect the loop and stop. -->
- **Fallback behavior**: <!-- N/A. Circular invocation should be stopped, not worked around. Note how the loop terminates: platform guard, model self-detection, context exhaustion, or user intervention. -->

#### `invocation-language-sensitivity`

- **Benchmark skills**: `invoke-alpha` + `invoke-beta` + `invoke-gamma` — Run the invocation chain test in English, then repeat in another language (e.g., Japanese: "呼び出しチェーンを開始してください"). Compare success rates across multiple runs.
- **Status**: Not tested
- **Observation**: <!-- Does skill-to-skill invocation reliability vary by prompt language? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Model-level. This is expected to be driven by the model's interpretation of system prompt framing. -->
- **Fallback behavior**: <!-- If invocation fails in a non-English language, does switching the user's prompt to English resolve it while keeping the skill instructions in the original language? Or must both be in English? -->

---

### Category 9: Skill Dependencies

#### `informal-dependency-resolution`

- **Benchmark skills**: `invoke-alpha` + `invoke-beta` — Same test as cross-skill-invocation. The invoke chain uses prose instructions to express dependencies between skills.
- **Status**: Not tested
- **Observation**: <!-- If a skill's body directs the model to activate another skill, does it attempt to do so? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Model-level (the model interprets the prose instruction). -->
- **Fallback behavior**: <!-- If the agent doesn't follow the prose dependency instruction, can the user manually activate the dependency skill first, then re-activate the original? Does pre-loading the dependency change the outcome? -->

#### `missing-dependency-behavior`

- **Benchmark skill**: `probe-missing-dep` — Activate the skill. It references `nonexistent-formatter` which doesn't exist. Observe the failure mode: does the model report the skill doesn't exist, silently skip the step, or attempt to fulfill the task from general knowledge?
- **Status**: Not tested
- **Observation**: <!-- When a skill references another skill that isn't installed, what happens? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Model-level (the model encounters the failure at runtime). -->
- **Fallback behavior**: <!-- Does the model tell the user about the missing dependency so they can install it? Or does it silently degrade? If the user asks "what skills does this skill need?", can the model answer from the skill's body text? -->

#### `nonstandard-dependency-fields`

- **Benchmark skill**: `probe-nonstandard-fields` — Activate the skill. It has `requires: probe-loading` and `depends-on: [probe-shadow-alpha, probe-shadow-beta]` in frontmatter. Check whether the platform acted on these fields or ignored them.
- **Status**: Not tested
- **Observation**: <!-- Does the platform recognize any dependency-related frontmatter fields not in the spec? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Platform-level -->
- **Fallback behavior**: <!-- If the platform ignores nonstandard fields but passes frontmatter to the model, does the model interpret and act on the fields on its own? -->

#### `cross-scope-dependency`

- **Benchmark skills**: `probe-cross-scope` + `probe-loading` — Install probe-cross-scope at project level and probe-loading at user level. Activate probe-cross-scope and see if it can invoke probe-loading across scopes. Then remove probe-loading from user level and test again.
- **Status**: Not tested
- **Observation**: <!-- How does the platform handle a project-level skill referencing a user-level skill? -->
- **Evidence**: <!--  -->
- **Platform-level or model-level?**: <!-- Both. Scope visibility is platform-level. Whether the model can invoke across scopes depends on what the harness exposes. -->
- **Fallback behavior**: <!-- If the agent can't see skills at a different scope, can the user manually activate the cross-scope skill? Does moving the dependency to the same scope resolve the issue? -->

---

## Additional Notes

<!--
  Anything else worth noting about this platform's skill loading behavior
  that doesn't fit neatly into the checks above. For example:

  - Platform-specific skill management commands or UI
  - Debug tools or logs that reveal loading behavior
  - Known bugs or quirks
  - Planned changes the platform has announced
-->
