# Contributing

This project needs empirical data from people testing on real platforms. There are
several ways to contribute, ranging from a single test on one platform to proposing
entirely new checks.

## Submitting platform results

This is the most valuable contribution. If you can test skill loading behavior on
a platform you use, here's how:

### Setup

1. Fork and clone this repository.
2. Copy `platform-loading-implementation/template.md` to a new file named after
   the platform (e.g., `platform-loading-implementation/claude-code.md`).
3. Install the benchmark skills from `benchmark-skills/` on your platform. See
   `benchmark-skills/README.md` for the full inventory, check-to-skill mapping,
   and test procedures.

### Testing

4. Work through the checks in your platform file. You don't need to test every
   check in one session; partial results are valuable.
5. For each check you test, fill in:
   - **Status**: Observed, Inconclusive, or Not tested
   - **Observation**: What happened
   - **Evidence**: How you know (debug logs, canary phrase results, screenshots)
   - **Platform-level or model-level?**: Whether the behavior is enforced by the
     harness or determined by the model
   - **Fallback behavior**: Whether the user can work around any limitations
6. Fill in the Platform Details table and Methodology section. These help others
   reproduce your findings and judge how much weight to give them.

### Submitting

7. Open a PR with your platform file. The PR description should note:
   - Which platform and version you tested
   - Which model(s) you used
   - How many checks you covered
   - Anything surprising you found

### Tips for good results

- **Start a fresh session for each check.** Prior checks leave context in the
  conversation that can contaminate later results. If the agent saw a canary phrase
  during an earlier check, it might "know" that phrase for a later check not because
  the platform loaded it, but because it's in the conversation history. Some checks
  also have specific context requirements (e.g., `context-compaction-protection`
  needs a long conversation, while `discovery-reading-depth` needs a clean session
  with no prior activation). If you do run multiple checks in one session, note that
  in your methodology so others can weigh the results accordingly.
- **Use the canary phrases.** They're the most reliable signal for what was loaded.
  Asking the model "do you know the phrase CARDINAL-ZEBRA-7742?" is more
  trustworthy than asking "what files do you have access to?"
- **Distinguish platform from model behavior.** If you can test with two different
  models on the same platform, differences between them point to model-level
  behavior. If both models behave the same way, it's more likely platform-level.
- **Repeat probabilistic tests.** Model-level behaviors may vary across runs. If
  something works 3 out of 5 times, that's a different finding than 5 out of 5.
  Note the ratio.
- **Record negative results.** "The platform did nothing" is a finding. "The model
  ignored the instruction" is a finding. These are just as useful as positive
  results.
- **Note your methodology.** How you installed skills, what prompts you used, and
  what tools you used to observe behavior all matter for reproducibility.

## Updating existing platform results

Platform behavior changes across versions. If you test a platform that already has
a results file and find different behavior:

- If the platform version is the same, add your observations inline with a note
  about the discrepancy. Conflicting data from different testers is valuable.
- If the platform version is different, consider creating a new file
  (e.g., `claude-code-v1.1.md`) or updating the existing file with a note about
  which version each observation applies to.

## Proposing new checks

If you've observed a loading behavior that isn't covered by the existing checks in
`loading-behavior.md`, open an issue describing:

- **What you observed**: The specific platform behavior.
- **Why it matters**: How it affects skill authors or users.
- **How to test it**: A reproducible procedure, ideally with a benchmark skill
  design.

If the check is accepted, it will need:

1. An entry in `loading-behavior.md` with an ID, category, what it checks, and
   why it matters.
2. A corresponding entry in `platform-loading-implementation/template.md`.
3. A benchmark skill in `benchmark-skills/` (if existing skills don't cover it),
   with canary phrases and a test procedure.
4. An entry in the check-to-skill mapping in `benchmark-skills/README.md`.

## Improving benchmark skills

If you find that a benchmark skill doesn't reliably test what it's supposed to, or
if a test procedure is ambiguous, open an issue or PR describing the problem and
your suggested fix. Changes to benchmark skills need care because they may
invalidate existing platform results that reference specific canary phrases.

## Writing style

This project is deliberately careful about the distinction between observations and
assertions. When writing check descriptions, platform results, or any other prose:

- Frame untested behaviors as questions, not claims. "Does the platform deduplicate?"
  not "The platform deduplicates."
- Attribute observations to their source. "One tester observed..." or "Issue #95
  reports..." rather than stating others' findings as established fact.
- Use hedging language for behaviors you haven't personally verified: "may vary,"
  "could," "it's unclear whether." Save definitive language for things you've
  tested and can show evidence for.
- Record what you saw, not what you expected to see. If a result surprises you,
  note the surprise, but report the observation.

## AI usage

We expect contributors to use AI tools, and encourage it. But the value of this
project depends on the accuracy of its content. If you use AI to help draft
platform findings, check descriptions, or benchmark skills, please validate the
output before submitting. AI tools are good at generating plausible-sounding
descriptions of platform behavior; the whole point of this project is that
plausible-sounding descriptions aren't good enough. Verify claims against what
you actually observed, and make sure any assertions you include are backed by
evidence you can point to.
