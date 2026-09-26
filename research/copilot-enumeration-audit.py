#!/usr/bin/env python3
"""Estimate what Copilot CLI's activation file list costs per skill.

Copilot CLI (validated 1.0.88) wraps an activated skill in a <skill-context>
block whose header states the skill's base directory and lists the absolute
path of every file under the skill directory (see the
bundled-file-enumeration-scale check). The list is re-sent on every
activation. This script reconstructs that header for every SKILL.md in one
or more checked-out repositories and reports its token cost, split by where
the files live (spec directories, evals/, everything else).

Usage:
    python3 copilot-enumeration-audit.py <events.jsonl> <dir>=<label> [...]

<events.jsonl> is an archived Copilot transcript from a run of the
bundled-file-enumeration-scale check (results/<run>/copilot/
bundled-file-enumeration-scale/events.jsonl); the script reconstructs its
wrapper first and refuses to continue unless the reconstruction matches the
real one byte for byte. Requires tiktoken (pip install tiktoken); tokens are
counted with o200k_base as an approximation, since the tokenizer of the
model Copilot ran is not public.

Assumptions: skills are installed at /Users/dev/project/.github/skills/<name>
(the install path is roughly half of every line, so a deeper path costs
more); a top-level vendor/ tree is skipped, because the 1.0.88 run omitted
one; node_modules/ was not tested and is listed if present.
"""
import json, os, re, statistics, sys

try:
    import tiktoken
except ImportError:
    sys.exit("tiktoken is required: pip install tiktoken")

ENC = tiktoken.get_encoding("o200k_base")
SPEC = {"scripts", "references", "assets"}
INSTALL = "/Users/dev/project/.github/skills"
EXCLUDED_TOP = {"vendor"}


def toks(s):
    return len(ENC.encode(s))


def files_of(skilldir):
    out = []
    for root, dirs, fs in os.walk(skilldir):
        dirs[:] = [d for d in dirs if d != ".git"]
        if root == skilldir:
            dirs[:] = [d for d in dirs if d not in EXCLUDED_TOP]
        for f in fs:
            rel = os.path.relpath(os.path.join(root, f), skilldir)
            if rel != "SKILL.md":
                out.append(rel)
    return out


def wrapper(name, base, rels):
    s = f'<skill-context name="{name}">\nBase directory for this skill: {base}\n\n'
    if rels:
        s += "Related files (use view tool to read):\n" + "".join(f"  - {base}/{r}\n" for r in rels) + "\n"
    return s


def group(rel):
    top = rel.split("/")[0]
    if "/" not in rel:
        return "root-files"
    if top in SPEC:
        return top
    if top in ("evals", "eval"):
        return "evals"
    return "other-dirs"


def body_tokens(skillmd):
    t = open(skillmd, encoding="utf-8", errors="replace").read()
    m = re.match(r"^---\n.*?\n---\n", t, re.S)
    return toks(t[m.end():] if m else t)


def calibrate(events_path, fixture_dir):
    real = None
    for line in open(events_path):
        e = json.loads(line)
        if e["type"] == "skill.context_delivered_ref":
            real = e["data"]["prefix"]
            break
    if real is None:
        sys.exit("no skill.context_delivered_ref record in the transcript")
    base = real.split("\n")[1].split(": ", 1)[1]
    rels = files_of(fixture_dir)
    mine = wrapper(os.path.basename(fixture_dir), base, rels)
    listed = sorted(x.split(base + "/", 1)[1] for x in real.split("\n") if x.startswith("  - "))
    if len(mine) != len(real) or listed != sorted(rels):
        sys.exit(f"calibration failed: reconstructed {len(mine)} chars vs real {len(real)}; file sets equal: {listed == sorted(rels)}")
    print(f"calibration ok: {len(real)} chars, {len(rels)} files")


def audit(repo_dir, label):
    rows = []
    for root, dirs, fs in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in (".git", "node_modules")]
        if "SKILL.md" in fs:
            name = os.path.basename(root)
            rels = files_of(root)
            base = f"{INSTALL}/{name}"
            per = {}
            for r in rels:
                per[group(r)] = per.get(group(r), 0) + toks(f"  - {base}/{r}\n")
            rows.append(dict(name=name, files=len(rels), pre=toks(wrapper(name, base, rels)),
                             body=body_tokens(os.path.join(root, "SKILL.md")), per=per,
                             evals=sum(1 for r in rels if group(r) == "evals")))
    if not rows:
        print(f"\n## {label}: no SKILL.md found")
        return
    n = len(rows)
    ft = [r["files"] for r in rows]
    pt = [r["pre"] for r in rows]
    bt = [r["body"] for r in rows]
    tot = sum(pt)
    evt = sum(r["per"].get("evals", 0) for r in rows)
    oth = sum(r["per"].get("other-dirs", 0) + r["per"].get("root-files", 0) for r in rows)
    spec = sum(sum(v for k, v in r["per"].items() if k in SPEC) for r in rows)

    def q(v, p):
        v = sorted(v)
        return v[min(len(v) - 1, int(p * len(v)))]

    print(f"\n## {label}: {n} skills, {sum(1 for r in rows if r['evals'])} ship evals/")
    print(f"files per skill: median {statistics.median(ft):.0f}, p90 {q(ft, .9)}, max {max(ft)}")
    print(f"wrapper tokens per activation: median {statistics.median(pt):.0f}, mean {statistics.mean(pt):.0f}, "
          f"p90 {q(pt, .9)}, max {max(pt)} (SKILL.md body median {statistics.median(bt):.0f})")
    print(f"share of wrapper tokens: evals/ {100 * evt / tot:.0f}%, other non-spec {100 * oth / tot:.0f}%, "
          f"spec dirs {100 * spec / tot:.0f}%, fixed lines {100 * (tot - evt - oth - spec) / tot:.0f}%")
    print(f"skills whose wrapper costs more than their body: {sum(1 for r in rows if r['pre'] > r['body'])}/{n}")
    for r in sorted(rows, key=lambda r: -r["pre"])[:4]:
        print(f"  {r['name'][:40]:40} files={r['files']:4} wrapper={r['pre']:5} "
              f"(evals {r['per'].get('evals', 0)}, other {r['per'].get('other-dirs', 0) + r['per'].get('root-files', 0)}) body={r['body']}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    here = os.path.dirname(os.path.abspath(__file__))
    calibrate(sys.argv[1], os.path.join(here, "..", "benchmark-skills", "probe-bulk-files"))
    for arg in sys.argv[2:]:
        d, _, label = arg.partition("=")
        audit(d, label or d)
