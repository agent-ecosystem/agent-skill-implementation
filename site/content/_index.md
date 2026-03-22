---
title: "Agent Skill Implementation"
description: "Empirical research into how agent platforms implement Agent Skill loading, management, and presentation."
---

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
through empirical testing rather than assumptions.

## Loading Behavior Checks

The current focus is on **skill loading behavior**: how platforms load skill
content into the agent's context, what they load, and when. We've cataloged
**[23 checks across 9 categories](/checks/)** that need empirical testing:

| Category | Checks | What it evaluates |
|----------|--------|-------------------|
| Loading Timing | 3 | When skill content enters the agent's context, and how much is loaded at each stage |
| Directory Recognition | 3 | Which directories a platform recognizes as part of a skill |
| Resource Access Patterns | 4 | How platforms make supporting files available to the model |
| Content Presentation | 3 | What the model actually sees when a skill is activated |
| Lifecycle Management | 3 | How platforms manage skill content over the course of a conversation |
| Access Control | 2 | How platforms gate skill loading and handle control-related fields |
| Structural Edge Cases | 2 | Platform behavior with skill structures that push beyond the spec's core examples |
| Skill-to-Skill Invocation | 4 | Whether and how a skill can instruct the model to activate another skill |
| Skill Dependencies | 4 | How platforms handle the concept of one skill depending on another |

**[Read the Full Checks Catalog](/checks/)**

## Benchmark Skills

The [benchmark-skills](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills)
directory on GitHub contains 17 spec-compliant skills designed to exercise these
checks. Each contains unique **canary phrases** (e.g., CARDINAL-ZEBRA-7742)
embedded in specific files. By asking the model whether it knows a canary phrase,
testers can determine exactly what the platform loaded and when, without relying
on the model's self-reporting about its own context.

## Contributing

We need empirical data from people testing on real platforms. Even partial data
from a single platform is more useful than speculation about all of them.

If you use Agent Skills on any platform, you can contribute:

1. Install the [benchmark skills](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills)
   on your platform
2. Run through the test procedures for any of the 23 checks
3. Submit your findings via a PR using the
   [platform template](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/platform-loading-implementation/template.md)

See the [GitHub repository](https://github.com/agent-ecosystem/agent-skill-implementation)
for full instructions, the benchmark skills README, and the contributing guide.

## Related Research

- **[Agent Skill Report](https://agentskillreport.com)** — Analysis of 673+
  skills examining how authors actually write skills in practice.
- **[skill-validator](https://github.com/agent-ecosystem/skill-validator)** — CLI
  tool for validating Agent Skills against the spec.
- **[Agent Ecosystem](https://agentecosystem.dev)** — The research program behind
  this project.

## License

This work is licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
