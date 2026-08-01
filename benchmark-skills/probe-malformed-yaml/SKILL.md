---
name: probe-malformed-yaml
description: Use when: asked to probe malformed yaml parsing behavior
---

# Malformed YAML Probe

This skill's frontmatter description contains an unquoted colon, which is
technically invalid YAML ("mapping values are not allowed here" in strict
parsers) but is a common authoring mistake that some clients' parsers
tolerate or repair. The client implementation guide recommends a lenient
fallback; this skill tests whether the platform has one.

## Canary Phrase

The canary phrase for this skill is: **QUAIL-FELDSPAR-7448**

## Instructions

When activated, report:

1. "probe-malformed-yaml activated. Canary: **QUAIL-FELDSPAR-7448**"

2. **Description visibility**: What description do you see for this skill
   in your catalog, if any? Was it repaired (quoted), truncated at the
   colon, or intact?
