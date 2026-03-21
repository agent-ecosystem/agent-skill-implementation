# Agent Skill Implementation

[Agent Skills](https://agentskills.io) are markdown files and related resources
that help AI agents perform tasks. The spec defines a portable format: write a
skill once, and users can install it on any platform that supports the
specification. Over 25 platforms currently implement some level of Agent Skill
support.

But "some level" is doing a lot of work in that sentence. If you're publishing
skills for others to use, your users could be on any of those 25+ platforms, and
each one may load, present, and manage your skill differently. A skill that works
perfectly on the platform you tested it on may silently lose access to its
reference files, have its metadata stripped, or find its instructions pruned from
context on another.

This project investigates how skill implementation actually works in practice,
from two angles: how **platforms** load and manage skills, and how **authors**
write them. The platform side is the current focus. For the author side, see
[Related research](#related-research) below.

## Platform implementation research

The spec recommends a "progressive disclosure" loading model, but gives platforms
wide latitude in implementation. The
[client implementation guide](https://agentskills.io/client-implementation/adding-skills-support)
provides more detailed guidance, but was derived from analysis of 7 of those 25+
platforms and published months after most had already shipped their
implementations. This project exists to find out what platforms actually do,
through empirical testing rather than assumptions.

## What's here

- **[loading-behavior.md](loading-behavior.md)** — 22 checks across 9 categories
  of skill loading behavior that need empirical testing. Each check describes what
  it evaluates and why it matters for skill authors.

- **[benchmark-skills/](benchmark-skills/)** — 16 spec-compliant skills designed
  to exercise those checks. Each contains unique canary phrases that reveal what a
  platform loaded and when, without relying on model self-reporting. See the
  [benchmark skills README](benchmark-skills/README.md) for the full inventory,
  check-to-skill mapping, and test procedures.

- **[platform-loading-implementation/](platform-loading-implementation/)** — Per-platform
  results. The [template](platform-loading-implementation/template.md) captures
  platform details, methodology, and findings for each check, including whether
  observed behavior is platform-level or model-level and whether fallback
  workarounds exist.

## Why this matters

Skill authors currently have no way to know what will happen when their skill is
activated on a given platform. A skill that works perfectly on one platform may
silently lose access to its reference files, have its frontmatter stripped, or find
its instructions pruned from context on another. Without empirical data about how
platforms actually behave, skill authors are writing for an idealized loading model
that may not match reality anywhere.

## Future areas of investigation

Skill loading behavior is the starting point, but not the only area where platform
behavior is unspecified and likely diverges. We plan to investigate these areas next:

- **Tool provisioning** — The spec defines an `allowed-tools` frontmatter field but
  says nothing about how platforms should act on it. Does the platform restrict the
  model to declared tools, provision additional tools the skill requests, or ignore
  the field entirely?
- **Activation mechanisms** — How does a user activate a skill? Slash command,
  natural language, automatic activation based on context? A skill designed for one
  activation style may never get discovered on a platform that only supports another.
- **Multi-skill context management** — When multiple skills are active, how does the
  platform manage the context budget? Which skill gets pruned first when context is
  tight?
- **Instruction authority and conflict resolution** — If two active skills give
  contradictory instructions, or a skill's instructions conflict with the platform's
  system prompt, what wins?
- **Security boundaries** — Beyond path traversal, can a skill's instructions cause
  the agent to run shell commands, make network requests, or modify files outside
  the project? How do platforms sandbox skill-directed actions?
- **Prompt injection resistance** — Can content in a skill's reference files inject
  instructions that override the SKILL.md body or the platform's system prompt?
- **Skill persistence and session behavior** — Does an activated skill stay active
  for the entire session? Can a user deactivate mid-conversation? Are activations
  remembered across sessions?
- **Internationalization** — Do platforms handle non-English skill content
  differently? Is a SKILL.md written in Japanese discovered and presented the same
  way as one in English?
- **Output formatting influence** — When a skill specifies output format, how
  reliably does the model comply across platforms, and how much does platform-level
  framing affect compliance?

If you're interested in helping design checks or benchmark skills for any of these
areas, open an issue.

## Related research

This project investigates the platform side of skill implementation: how platforms
load, manage, and present skills. There is a separate and complementary line of
research investigating the author side: how people actually write skills in
practice, and what effect skill quality has on agent output.

- **[Agent Skill Report](https://agentskillreport.com)** — An analysis of 673
  skills examining real-world authoring patterns, structural choices, and common
  issues. Published findings are available now.
- **Broad skill implementation research** (in progress) — A larger-scale study
  cataloging 80,000+ skills from 11,000+ repositories to investigate skill quality
  patterns and their effects on agent behavior. Findings will be published as the
  research progresses.

- **[skill-validator](https://github.com/agent-ecosystem/skill-validator)** — A
  CLI tool for validating Agent Skills against the spec. Developed alongside the
  initial research and refined through community feedback and production use. If
  you're distributing skills that need to work across platforms, the validator can
  catch structural issues before your users encounter them.

The platform research here and the skill authoring research inform each other.
How platforms load skills determines which authoring patterns work; how authors
write skills determines which platform behaviors cause real problems.

## Contributing

We need people testing on real platforms. Even partial data from a single platform
is more useful than speculation about all of them.

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to submit platform results, propose
new checks, or improve the benchmark skills.

## Glossary

Terms used throughout this project:

- **Canary phrase**: A unique string (e.g., CARDINAL-ZEBRA-7742) embedded in a
  benchmark skill file. If the model knows a canary phrase without having
  explicitly read the file containing it, the platform loaded that file
  automatically. More reliable than asking the model to self-report about its
  context.
- **Context compaction**: When a platform truncates or summarizes older messages
  to free space in the context window during a long conversation. Also called
  context pruning or summarization.
- **Context window**: The total amount of text (measured in tokens) that a model
  can consider at once. Skill content, conversation history, and system prompts
  all compete for this space.
- **Fallback behavior**: What happens when a platform's default behavior doesn't
  surface content to the model. Can the agent self-recover, does the user need to
  intervene, or is the content inaccessible?
- **Harness**: The platform's infrastructure that wraps around the model. The
  harness handles skill discovery, file loading, tool provisioning, and context
  management. Harness behavior is deterministic; model behavior is probabilistic.
- **Model-level behavior**: Behavior determined by the model's interpretation of
  instructions. May vary by model, prompt language, or across runs. Example: the
  model deciding whether to follow a markdown link and read the referenced file.
- **Platform-level behavior**: Behavior enforced by the harness. Deterministic and
  consistent across runs. Example: the platform stripping YAML frontmatter before
  passing skill content to the model.
- **Progressive disclosure**: The spec's recommended three-tier loading model:
  metadata at startup, instructions on activation, resources on demand. Whether
  platforms actually follow this model is one of the core questions this project
  investigates.

## License

This work is licensed under [Creative Commons Attribution 4.0 International](LICENSE).

## AI Usage Disclosure

I (dacharyc) - the project creator - yeeted this into existence with the aid of
Claude Code across a few hours on a Saturday morning. All content was informed by
my ideas and questions I want to answer for the reasons enumerated here, but Claude
helped with landscape research, drafting the benchmark skills, and creating some of
the content. I have attempted to validate the accuracy of all assertions here (and
you'll find very few assertions here on purpose because of the speculative nature
of these questions), but if you spot any inaccuracies or logical gaps, please let
me know.

For anyone contributing to this project, I expect you to use AI, but please do it
carefully and responsibly. Validate outputs and assertions with your delightful
human brains to make sure everything you include is sensible and stands up to rigor
and scrutiny, and I will attempt to do the same.
