// Skill-to-skill invocation and dependency checks: the invoke chain, the
// circular pair, the missing-dependency probe, and nonstandard dependency
// frontmatter.
package checks

import (
	"fmt"

	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

// Body canaries, from the benchmark's canary index.
const (
	invokeAlphaBodyCanary     = "IBIS-RUST-3310"
	invokeBetaBodyCanary      = "TERN-MOSS-6647"
	invokeGammaBodyCanary     = "JAY-TEAL-9984"
	circularAlphaBodyCanary   = "KITE-ONYX-2251"
	circularBetaBodyCanary    = "WREN-SLATE-7738"
	probeMissingDepBodyCanary = "GULL-IRON-4492"
)

// alphaBetaSpec is shared by cross-skill-invocation and
// informal-dependency-resolution: the benchmark defines both as the same
// invoke-alpha → invoke-beta procedure, asked under different questions
// (can a skill activate another at all vs. does a prose-expressed
// dependency get resolved at runtime).
func alphaBetaSpec(id, description string) Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("invoke-alpha") }
	return Spec{
		ID:          id,
		Description: description,
		Sessions: []Session{{
			// invoke-gamma is deliberately absent, per the benchmark's
			// procedure: only the alpha→beta link is graded here; the full
			// chain belongs to invocation-depth-limit.
			Skills: []string{"invoke-alpha", "invoke-beta"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: id}
			aInj, aPull := loadsOf(sos[0], invokeAlphaBodyCanary)
			if len(aInj)+len(aPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "invoke-alpha itself never loaded")
				return f
			}
			f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(aInj, aPull)[0].EventIndex, "invoke-alpha body loaded (chain entry)"))

			bInj, bPull := loadsOf(sos[0], invokeBetaBodyCanary)
			attempts := trace.ToolReadsOf(sos[0].Final().Session, "invoke-beta")
			switch {
			case len(bInj)+len(bPull) > 0:
				f.Status = StatusObserved
				f.Verdict = "second-skill-loaded"
				f.Vehicle = vehicleOf(bInj, bPull)
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(bInj, bPull)[0].EventIndex, "invoke-beta's body canary loaded after invoke-alpha's instruction"))
				noteFileReadWorkaround(&f, len(aInj) > 0, bInj, bPull, "invoke-beta")
			case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
				f.Status = StatusObserved
				f.Verdict = "attempted-visible-failure"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on invoke-beta returned an error result"))
			case len(attempts) > 0:
				f.Status = StatusObserved
				f.Verdict = "attempted-not-loaded"
				f.Confidence = negConfidence(sos[0])
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on invoke-beta; its body never arrived"))
			default:
				f.Status = StatusObserved
				f.Verdict = "not-attempted"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "model never attempted invoke-beta (model-level refusal or omission)")
			}
			f.Notes = append(f.Notes,
				"invoke-gamma deliberately not installed; the chain tail beyond beta is out of scope here",
				"final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func crossSkillInvocation() Spec {
	return alphaBetaSpec("cross-skill-invocation",
		"Can one skill's instructions get a second installed skill activated by name?")
}

func informalDependencyResolution() Spec {
	return alphaBetaSpec("informal-dependency-resolution",
		"Is a dependency expressed only in prose (\"now activate the invoke-beta skill\") resolved at runtime?")
}

// chainSpec is shared by invocation-depth-limit and
// invocation-language-sensitivity: the full three-skill chain, differing
// only in the activation prompt.
func chainSpec(id, description string, prompt func(profile.Profile) string, extraNotes ...string) Spec {
	return Spec{
		ID:          id,
		Description: description,
		Sessions: []Session{{
			Skills: []string{"invoke-alpha", "invoke-beta", "invoke-gamma"},
			Turns:  []Turn{{Prompt: prompt, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: id}
			links := []struct{ name, canary string }{
				{"invoke-alpha", invokeAlphaBodyCanary},
				{"invoke-beta", invokeBetaBodyCanary},
				{"invoke-gamma", invokeGammaBodyCanary},
			}
			depth := 0
			entryInjected := false
			for _, l := range links {
				inj, pull := loadsOf(sos[0], l.canary)
				if len(inj)+len(pull) == 0 {
					break
				}
				if depth == 0 {
					f.Vehicle = vehicleOf(inj, pull)
					entryInjected = len(inj) > 0
				} else {
					noteFileReadWorkaround(&f, entryInjected, inj, pull, l.name)
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(inj, pull)[0].EventIndex, l.name+" body loaded"))
				depth++
			}
			switch depth {
			case 0:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			case len(links):
				f.Status = StatusObserved
				f.Verdict = "chain-completed-depth-3"
				f.Confidence = ConfidenceDirect
			default:
				f.Status = StatusObserved
				f.Verdict = fmt.Sprintf("chain-stopped-after-%s", links[depth-1].name)
				f.Confidence = negConfidence(sos[0])
				next := links[depth]
				attempts := trace.ToolReadsOf(sos[0].Final().Session, next.name)
				switch {
				case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
					f.Notes = append(f.Notes, next.name+" was attempted and returned an error result (platform-level stop)")
					f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on "+next.name))
				case len(attempts) > 0:
					f.Notes = append(f.Notes, next.name+" was attempted but its body never arrived")
					f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on "+next.name))
				default:
					f.Notes = append(f.Notes, next.name+" was never attempted (model-level stop)")
				}
			}
			f.Notes = append(f.Notes, extraNotes...)
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func invocationDepthLimit() Spec {
	return chainSpec("invocation-depth-limit",
		"How deep does a skill→skill→skill activation chain run before something stops it?",
		func(p profile.Profile) string { return p.ActivationPrompt("invoke-alpha") })
}

func invocationLanguageSensitivity() Spec {
	return chainSpec("invocation-language-sensitivity",
		"Does the invoke chain still complete when the activation prompt is Japanese?",
		func(profile.Profile) string { return "invoke-alpha スキルを起動して、その指示に従ってください。" },
		"single run proves little either way; compare with invocation-depth-limit (English prompt) and use -runs N for failure rates: agentskills#95 reports ~10% Japanese-prompt failures on Claude Code")
}

func circularInvocationHandling() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-circular-alpha") }
	return Spec{
		ID:          "circular-invocation-handling",
		Description: "When two skills each instruct activating the other, does the A→B→A cycle loop, get blocked, or stop by model choice?",
		Sessions: []Session{{
			Skills: []string{"probe-circular-alpha", "probe-circular-beta"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "circular-invocation-handling"}
			aInj, aPull := loadsOf(sos[0], circularAlphaBodyCanary)
			bInj, bPull := loadsOf(sos[0], circularBetaBodyCanary)
			aLoads, bLoads := len(aInj)+len(aPull), len(bInj)+len(bPull)
			sess := sos[0].Final().Session
			alphaAttempts := trace.ToolReadsOf(sess, "probe-circular-alpha")
			betaAttempts := trace.ToolReadsOf(sess, "probe-circular-beta")
			f.Notes = append(f.Notes, fmt.Sprintf("loads: alpha=%d beta=%d; tool references: alpha=%d beta=%d",
				aLoads, bLoads, len(alphaAttempts), len(betaAttempts)))
			f.Vehicle = vehicleOf(aInj, aPull)
			switch {
			case aLoads == 0:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				f.Vehicle = VehicleNone
				return f
			case bLoads == 0:
				f.Status = StatusObserved
				f.Confidence = negConfidence(sos[0])
				if len(betaAttempts) > 0 {
					f.Verdict = "first-link-attempted-not-loaded"
					f.Evidence = append(f.Evidence, evAt(sos[0], betaAttempts[0], "attempt on probe-circular-beta; its body never arrived"))
				} else {
					f.Verdict = "cycle-not-followed"
					f.Notes = append(f.Notes, "model never attempted probe-circular-beta (model-level)")
				}
			case aLoads >= 2:
				all := loadsInOrder(aInj, aPull)
				f.Status = StatusObserved
				f.Verdict = fmt.Sprintf("cycle-re-entered:alpha=%d,beta=%d", aLoads, bLoads)
				f.Confidence = ConfidenceDirect
				noteFileReadWorkaround(&f, len(aInj) > 0, bInj, bPull, "probe-circular-beta")
				if len(aInj) == 1 && len(aPull) > 0 {
					f.Notes = append(f.Notes, "the alpha re-entry arrived via file read, not a second injection (likely a model workaround around the platform's activation path)")
				}
				f.Evidence = append(f.Evidence,
					evAt(sos[0], all[0].EventIndex, "first load of probe-circular-alpha"),
					evAt(sos[0], all[1].EventIndex, "alpha loaded again after beta; no platform guard stopped the cycle"))
				f.Notes = append(f.Notes, "loop terminated within the session (model choice or turn end), not by a visible platform block")
			case len(alphaAttempts) >= 2:
				f.Status = StatusObserved
				f.Confidence = ConfidenceDirect
				if errorResultsFor(sos[0], alphaAttempts[1:]) {
					f.Verdict = "reinvocation-blocked"
					f.Evidence = append(f.Evidence, evAt(sos[0], alphaAttempts[1], "re-attempt on probe-circular-alpha returned an error result (platform guard)"))
				} else {
					f.Verdict = "reinvocation-no-reload"
					f.Evidence = append(f.Evidence, evAt(sos[0], alphaAttempts[1], "re-attempt on probe-circular-alpha produced no second body load"))
				}
			default:
				f.Status = StatusObserved
				f.Verdict = "cycle-stopped-model-choice"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "beta activated but the model never re-attempted alpha (model-level stop)")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func missingDependencyBehavior() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-missing-dep") }
	return Spec{
		ID:          "missing-dependency-behavior",
		Description: "When a skill instructs activating a skill that is not installed, is the failure visible, reported, or silently skipped?",
		Sessions: []Session{{
			Skills: []string{"probe-missing-dep"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "missing-dependency-behavior"}
			inj, pull := loadsOf(sos[0], probeMissingDepBodyCanary)
			if len(inj)+len(pull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(inj, pull)
			f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(inj, pull)[0].EventIndex, "probe-missing-dep body loaded"))

			attempts := trace.ToolReadsOf(sos[0].Final().Session, "nonexistent-formatter")
			mentions := assistantMentions(sos[0], "nonexistent-formatter")
			f.Status = StatusObserved
			switch {
			case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
				f.Verdict = "attempted-visible-failure"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on the missing skill returned an error result"))
			case len(attempts) > 0:
				f.Verdict = "attempted-no-error"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "attempt on the missing skill; no error result recorded"))
			case len(mentions) > 0:
				// The tier the cross-scope run exposed: the model consulted
				// its listing and surfaced the absence without attempting a
				// call — distinct from, and better than, a silent skip.
				f.Verdict = "reported-without-attempt"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], mentions[0].EventIndex, "model surfaced the missing dependency in its own text without attempting it"))
			default:
				f.Verdict = "silent-skip"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "no attempt and no assistant-text acknowledgment of the missing skill")
			}
			if f.Vehicle == VehicleModelPull {
				f.Notes = append(f.Notes, "which failure tier appears (attempted vs reported vs silent) is the model's choice on pull harnesses and can vary between runs")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// nonstandardFieldsBodyMarker detects probe-nonstandard-fields activation;
// the skill has no body canary in the canary index.
const nonstandardFieldsBodyMarker = "recognizes and acts on these fields"

func nonstandardDependencyFields() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-nonstandard-fields") }
	return Spec{
		ID:          "nonstandard-dependency-fields",
		Description: "Does the platform act on nonstandard dependency frontmatter (requires, depends-on, priority)?",
		Sessions: []Session{{
			// The named dependencies are installed so a platform that DOES
			// act on the fields has something to resolve.
			Skills: []string{"probe-nonstandard-fields", "probe-loading", "probe-shadow-alpha", "probe-shadow-beta"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "nonstandard-dependency-fields"}
			bInj, bPull := loadsOf(sos[0], nonstandardFieldsBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(bInj, bPull)[0].EventIndex, "probe-nonstandard-fields body loaded"))

			// A dependency counts as platform-resolved only if its content
			// arrived BEFORE the model's first tool reference to it; after
			// that it is the model exploring on its own initiative (the
			// skill's instructions invite exactly that).
			sess := sos[0].Final().Session
			deps := []struct{ name, marker string }{
				{"probe-loading", probeLoadingBodyCanary},
				{"probe-shadow-alpha", shadowAlphaBodyMarker},
				{"probe-shadow-beta", shadowBetaBodyMarker},
			}
			var platformActed, modelExplored []string
			for _, d := range deps {
				inj, pull := loadsOf(sos[0], d.marker)
				arrivals := loadsInOrder(inj, pull)
				if len(arrivals) == 0 {
					continue
				}
				attempts := trace.ToolReadsOf(sess, d.name)
				if len(attempts) == 0 || arrivals[0].EventIndex < attempts[0] {
					platformActed = append(platformActed, d.name)
					f.Evidence = append(f.Evidence, evAt(sos[0], arrivals[0].EventIndex, d.name+" content arrived before any model attempt (platform-initiated)"))
				} else {
					modelExplored = append(modelExplored, d.name)
				}
			}
			f.Status = StatusObserved
			if len(platformActed) > 0 {
				f.Verdict = fmt.Sprintf("fields-acted-on:%v", platformActed)
				f.Confidence = ConfidenceDirect
			} else {
				f.Verdict = "fields-ignored"
				f.Confidence = negConfidence(sos[0])
				if len(modelExplored) > 0 {
					f.Notes = append(f.Notes, fmt.Sprintf("model explored %v itself after reading the field names (model-level curiosity, not platform dependency resolution)", modelExplored))
				}
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}
