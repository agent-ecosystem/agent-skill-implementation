---
title: "Agent Skill Implementation"
description: "Empirical research into how agent platforms implement Agent Skill loading, management, and presentation."
---

<style>
.home-hero {display: grid; grid-template-columns: 1.1fr 0.9fr; gap: 2.2rem; align-items: center; margin: 0.6rem 0 2.2rem;}
@media (max-width: 820px) {.home-hero {grid-template-columns: 1fr;}}
.home-headline {font-size: 2.5rem; line-height: 1.15; font-weight: 800; margin: 0 0 1.1rem;}
@media (max-width: 820px) {.home-headline {font-size: 1.9rem;}}
.home-tagline {font-size: 1.15rem; line-height: 1.6; margin: 0;}
.home-cta {display: flex; flex-wrap: wrap; gap: 0.7rem; margin-top: 1.4rem;}
.home-btn {display: inline-block; padding: 0.55rem 1.25rem; border-radius: 9999px; font-weight: 600; font-size: 0.95rem; text-decoration: none;}
.home-btn-primary {background: rgba(var(--color-primary-600), 1); color: #fff;}
.home-btn-primary:hover {background: rgba(var(--color-primary-700), 1);}
.home-btn-secondary {background: rgba(128, 128, 128, 0.16); color: inherit;}
.home-btn-secondary:hover {background: rgba(128, 128, 128, 0.28);}
.home-term {background: #15171c; border-radius: 0.9rem; padding: 1.1rem 1.3rem; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 0.78rem; line-height: 1.75; color: #d4d4d8; overflow-x: auto; white-space: pre;}
.home-term .t-head {color: #93c5fd;}
.home-term .t-dim {color: #8b8b93;}
.home-term .t-ok {color: #4ade80;}
.home-term .t-bad {color: #f87171;}
.home-cards {display: grid; grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr)); gap: 1rem; margin: 1.5rem 0 2.2rem;}
.home-card {background: rgba(128, 128, 128, 0.09); border-radius: 0.9rem; padding: 1.15rem 1.3rem; display: flex; flex-direction: column; gap: 0.5rem;}
.home-card h3 {margin: 0; font-size: 1rem; font-weight: 700;}
.home-card p {margin: 0; font-size: 0.9rem; line-height: 1.55;}
.home-card .home-card-link {margin-top: auto; padding-top: 0.4rem; font-size: 0.85rem; font-weight: 600; text-decoration: none; color: rgba(var(--color-primary-500), 1);}
</style>

<div class="not-prose home-hero">
  <div>
    <h1 class="home-headline">The same skill behaves differently on every platform</h1>
    <p class="home-tagline">Agent Skills promise write-once portability across 25+ platforms. This site measures where that promise holds and where it breaks: the same probe skills, run on real platforms, every finding cited to a transcript.</p>
    <div class="home-cta">
      <a class="home-btn home-btn-primary" href="/platforms/">See What Platforms Do</a>
      <a class="home-btn home-btn-secondary" href="/checks/">Browse the Checks</a>
      <a class="home-btn home-btn-secondary" href="/guidance/">Write Portable Skills</a>
    </div>
  </div>
  <div class="home-term"><span class="t-head">platform comparison · bundled-script-execution</span>
<span class="t-dim">──────────────────────────────────────────────</span>
Can the agent run a bundled scripts/ file?
<span> </span>
  Antigravity CLI  <span class="t-ok">✓ script ran; output returned</span>
  Claude Code      <span class="t-bad">✗ blocked with an error</span>
  Codex CLI        <span class="t-ok">✓ script ran; output returned</span>
<span> </span>
<span class="t-dim">40 checks · 3 platforms · every claim transcript-cited</span></div>
</div>

<div class="not-prose home-cards">
  <div class="home-card">
    <h3>The Checks</h3>
    <p>40 checks across 10 categories, from loading timing to validation strictness. Each asks one testable question about platform behavior and explains why the answer matters to skill authors.</p>
    <a class="home-card-link" href="/checks/">Browse the catalog &rarr;</a>
  </div>
  <div class="home-card">
    <h3>Platform Reports</h3>
    <p>Automated, transcript-cited findings for every tested platform, a comparison table of where they agree and diverge, and a summary of where each platform contradicts the Agent Skills specification.</p>
    <a class="home-card-link" href="/platforms/">Read the reports &rarr;</a>
  </div>
  <div class="home-card">
    <h3>Authoring Guidance</h3>
    <p>The findings turned into practice: rules for writing skills that survive platform differences, each backed by the checks that motivated it, plus a glossary of the terms used throughout.</p>
    <a class="home-card-link" href="/guidance/">Get the guidance &rarr;</a>
  </div>
</div>

## One skill, different outcomes

[Agent Skills](https://agentskills.io) define a portable format, but the spec
leaves most loading, validation, and permission behavior open, and platforms
filled the gaps differently. The same probe skills, run against the platforms
with headless modes, come back with different answers:

<div class="not-prose home-cards">
  <div class="home-card">
    <h3>Frontmatter can vanish</h3>
    <p>Claude Code strips YAML frontmatter before injecting a skill. On Codex CLI and Antigravity, the model sees it only if it reads the raw file. Load-bearing information that lives only in frontmatter may never reach the model.</p>
    <a class="home-card-link" href="/platforms/claude-code/#frontmatter-handling">See the finding &rarr;</a>
  </div>
  <div class="home-card">
    <h3>Grouped skills disappear</h3>
    <p>Organize skills in subfolders and Codex CLI still finds them; Claude Code and Antigravity list direct children only. The grouped skills vanish from their catalogs with no error anywhere.</p>
    <a class="home-card-link" href="/platforms/#discovery-scope">See the comparison &rarr;</a>
  </div>
  <div class="home-card">
    <h3>Names resolve differently</h3>
    <p>Install the same skill name at project and user scope and Codex CLI and Antigravity load the project variant. Claude Code loads the user variant, against the implementation guide's "universal convention."</p>
    <a class="home-card-link" href="/platforms/#discovery-scope">See the comparison &rarr;</a>
  </div>
</div>

A skill that works perfectly where you wrote it can misbehave everywhere else,
with no error and no way to tell from the outside. Every claim above traces to
a transcript-cited finding.

## How the testing works

The [benchmark skills](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills)
are 33 fixtures (spec-compliant skills plus deliberate rule-breakers) seeded
with unique **canary phrases**. By asking the model whether it knows a canary
phrase, we can tell exactly what a platform loaded and when, without trusting
the model's self-reporting about its own context. An automated runner drives
the checks headlessly and cites every finding to an archived transcript.

## Contributing

We need empirical data from real platforms, and even partial data from a single
platform beats speculation about all of them. Install the
[benchmark skills](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-skills),
run any of the 40 checks, and submit findings with the
[platform template](https://github.com/agent-ecosystem/agent-skill-implementation/blob/main/platform-findings/template.md).
The [GitHub repository](https://github.com/agent-ecosystem/agent-skill-implementation)
has full instructions.

## Related Research

- **[Agent Skill Report](https://agentskillreport.com)**: Analysis of 673+
  skills examining how authors actually write skills in practice.
- **[skill-validator](https://github.com/agent-ecosystem/skill-validator)**: CLI
  tool for validating Agent Skills against the spec.
- **[Agent Ecosystem](https://agentecosystem.dev)**: The research program behind
  this project.

## License

This work is licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
