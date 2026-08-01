---
title: "Glossary"
description: "Terms used throughout the checks, platform reports, and authoring guidance."
date: 2026-08-01
weight: 2
---

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
- **Pull harness**: A platform where the model fetches skill content itself with
  its file-read tools; activation *is* a read. The model sees the raw file
  (frontmatter included), and behaviors like re-reading on reactivation or
  resolving a dependency are largely model choices rather than platform policy.
  Automated findings record this as the `model-pull` vehicle. Codex CLI and
  Antigravity behave this way in our findings.
- **Push harness**: A platform whose harness injects skill content into the
  model's context at activation (e.g., via a dedicated skill tool). The platform
  controls what the model sees (it may strip frontmatter or wrap content), and
  loading behaviors like deduplication are enforceable platform-side. Automated
  findings record this as the `harness-push` vehicle. Claude Code behaves this
  way in our findings. A single platform can mix vehicles: a push harness still
  relies on model pulls for bundled resources.
