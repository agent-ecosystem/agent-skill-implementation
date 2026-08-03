// Resource-access and content-presentation checks: eager link resolution
// and metadata value edge cases.
package checks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/agent-ecosystem/agentminutes/session"
	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

// linkedResourcesBodyMarker stands in for a body canary: the canary index
// gives probe-linked-resources none, so activation is detected by a
// distinctive body sentence instead.
const linkedResourcesBodyMarker = "mentioned here only as a plain text path"

// linkedResourceCanaries maps probe-linked-resources files to canaries,
// split by whether the SKILL.md body markdown-links the file. The unlinked
// file separates link-driven pre-fetching from bulk directory loading.
var linkedResourceCanaries = []struct {
	file   string
	canary string
	linked bool
}{
	{"references/setup-guide.md", "PARROT-SILVER-4412", true},
	{"references/troubleshooting.md", "TOUCAN-BRONZE-9931", true},
	{"references/unlinked-data.md", "EAGLE-COPPER-1178", false},
}

func eagerLinkResolution() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-linked-resources") }
	return Spec{
		ID:          "eager-link-resolution",
		Description: "Does activation pre-fetch files markdown-linked from the SKILL.md body, and does that extend to a file mentioned only as plain text?",
		Sessions: []Session{{
			Skills: []string{"probe-linked-resources"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "eager-link-resolution"}
			bodyInj, bodyPull := loadsOf(sos[0], linkedResourcesBodyMarker)
			if len(bodyInj)+len(bodyPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "the probe-linked-resources body marker never arrived via injection or tool result")
				return f
			}
			f.Vehicle = vehicleOf(bodyInj, bodyPull)
			f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(bodyInj, bodyPull)[0].EventIndex, "skill body loaded"))

			// Injection is the pre-fetch signal. A pull is the opposite: the
			// model reading a file itself is exactly the on-demand behavior
			// the skill's own instructions call for.
			var eagerLinked, eagerUnlinked, onDemand []string
			for _, r := range linkedResourceCanaries {
				inj, pull := loadsOf(sos[0], r.canary)
				switch {
				case len(inj) > 0:
					if r.linked {
						eagerLinked = append(eagerLinked, r.file)
					} else {
						eagerUnlinked = append(eagerUnlinked, r.file)
					}
					f.Evidence = append(f.Evidence, evAt(sos[0], inj[0].EventIndex, r.file+" injected by the harness"))
				case len(pull) > 0:
					onDemand = append(onDemand, r.file)
					f.Evidence = append(f.Evidence, evAt(sos[0], pull[0].EventIndex, r.file+" arrived only via the model's own read"))
				}
			}
			f.Status = StatusObserved
			switch {
			case len(eagerUnlinked) > 0:
				f.Verdict = "bulk-directory-load"
				f.Confidence = ConfidenceDirect
				f.Notes = append(f.Notes, "even the unlinked file was injected; pre-fetching is not link-driven")
			case len(eagerLinked) > 0:
				f.Verdict = "eager-link-prefetch"
				f.Confidence = ConfidenceDirect
			default:
				f.Verdict = "no-prefetch"
				f.Confidence = negConfidence(sos[0])
				if len(onDemand) > 0 {
					f.Notes = append(f.Notes, fmt.Sprintf("model read %v itself, corroborating it did not already have them", onDemand))
				}
			}
			return f
		},
	}
}

// unlinkedReferenceName is a filename-only marker: the file
// references/unreferenced-detail.md exists on disk but its NAME appears in
// no skill body, so the text can only reach the model via a directory
// listing or platform enumeration.
const unlinkedReferenceName = "unreferenced-detail"

func resourceEnumerationBehavior() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-loading") }
	return Spec{
		ID:          "resource-enumeration-behavior",
		Description: "At activation, are a skill's reference files enumerated to the model (names), loaded outright (contents), or invisible until explored?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "resource-enumeration-behavior"}
			bInj, bPull := loadsOf(sos[0], probeLoadingBodyCanary)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			contentInj, _ := loadsOf(sos[0], probeLoadingResourceCanaries["references/unreferenced-detail.md"])
			nameInj, namePull := loadsOf(sos[0], unlinkedReferenceName)
			f.Status = StatusObserved
			switch {
			case len(contentInj) > 0:
				f.Verdict = "contents-loaded-at-activation"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], contentInj[0].EventIndex, "unlinked reference file's CONTENT injected at activation"))
			case len(nameInj) > 0:
				f.Verdict = "listing-enumerated-without-contents"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], nameInj[0].EventIndex, "unlinked reference file's NAME injected without its content"))
			case len(namePull) > 0:
				f.Verdict = "model-enumerated-on-demand"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], namePull[0].EventIndex, "unlinked file's name arrived only via the model's own directory exploration"))
			default:
				f.Verdict = "no-enumeration"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "the unlinked file's name never reached the model; only body-linked files are discoverable without exploration")
			}
			return f
		},
	}
}

func pathResolutionBase() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-linked-resources") }
	return Spec{
		ID:          "path-resolution-base",
		Description: "When the model follows a SKILL.md relative path like references/setup-guide.md, what does it resolve against, and does the bare path work as written?",
		Sessions: []Session{{
			Skills: []string{"probe-linked-resources"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "path-resolution-base"}
			bodyInj, bodyPull := loadsOf(sos[0], linkedResourcesBodyMarker)
			if len(bodyInj)+len(bodyPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bodyInj, bodyPull)

			// Classify every read call targeting the linked files: a call
			// naming the skill directory is qualified; a call using only
			// the SKILL.md-relative form is bare. A bare call can only
			// succeed if something resolves it against the skill dir (the
			// harness cwd is the project root, where the bare path does
			// not exist).
			sess := sos[0].Final().Session
			calls := append(trace.ToolReadsOf(sess, "setup-guide.md"), trace.ToolReadsOf(sess, "troubleshooting.md")...)
			sort.Ints(calls)
			// "Succeeded" means the linked file's canary came back in THAT
			// call's result. A completed tool call is not enough: a shell
			// or script wrapper reports per-path failures in ordinary
			// output with no tool-level error (observed on codex, whose
			// model looped over the bare paths in one exec script).
			resultHasCanary := func(callIdx int) bool {
				id := sess.Events[callIdx].ToolCall.ToolCallID
				for i := range sess.Events {
					ev := &sess.Events[i]
					if ev.Kind == session.KindToolResult && ev.ToolResult.ToolCallID == id {
						text := ev.ToolResult.Text()
						return strings.Contains(text, "PARROT-SILVER-4412") || strings.Contains(text, "TOUCAN-BRONZE-9931")
					}
				}
				return false
			}
			var bareOK, bareFail, qualOK bool
			for _, idx := range calls {
				input := string(sess.Events[idx].ToolCall.Input)
				qualified := strings.Contains(input, "probe-linked-resources")
				delivered := !errorResultsFor(sos[0], []int{idx}) && resultHasCanary(idx)
				switch {
				case qualified && delivered:
					qualOK = true
				case !qualified && delivered:
					bareOK = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "bare SKILL.md-relative path delivered the file's canary"))
				case !qualified:
					bareFail = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "bare SKILL.md-relative path attempt did not deliver content"))
				}
			}
			arrived := func(canary string) bool {
				inj, pull := loadsOf(sos[0], canary)
				return len(inj)+len(pull) > 0
			}
			gotFiles := arrived("PARROT-SILVER-4412") || arrived("TOUCAN-BRONZE-9931")

			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			switch {
			case bareOK:
				f.Verdict = "bare-relative-resolves-to-skill-dir"
			case bareFail && qualOK:
				f.Verdict = "cwd-base-model-requalified"
				f.Notes = append(f.Notes, "the path as written in SKILL.md does not resolve; the model recovered by qualifying it with the skill directory")
			case bareFail:
				f.Verdict = "bare-relative-fails-no-recovery"
			case qualOK && gotFiles:
				f.Verdict = "model-preemptively-qualified"
				f.Confidence = ConfidenceInferred
				f.Notes = append(f.Notes, "the model never tried the bare path, so the platform's resolution base was not directly exercised; it navigated by qualified path from the start")
			default:
				f.Status = StatusInconclusive
				f.Verdict = "files-not-read"
			}
			return f
		},
	}
}

// Shadow pair markers. The raw canaries (STORK-CORAL-4471 /
// EGRET-SLATE-8823) are USELESS for tracing here: both skills' BODIES name
// both canaries as instructions, so a body load alone delivers them. The
// "Shadow-<x> canary phrase" sentences exist only inside the respective
// references/API.md, making them the real read-detection markers.
const (
	shadowAlphaBodyMarker = "Shadow Alpha Probe"
	shadowBetaBodyMarker  = "Shadow Beta Probe"
	shadowAlphaAPIMarker  = "Shadow-alpha canary phrase"
	shadowBetaAPIMarker   = "Shadow-beta canary phrase"
)

func crossSkillResourceShadowing() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-shadow-alpha") }
	return Spec{
		ID:          "cross-skill-resource-shadowing",
		Description: "With two skills both owning references/API.md, does the activated skill's read get its own file or the sibling's?",
		Sessions: []Session{{
			Skills: []string{"probe-shadow-alpha", "probe-shadow-beta"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "cross-skill-resource-shadowing"}
			bInj, bPull := loadsOf(sos[0], shadowAlphaBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			aInj, aPull := loadsOf(sos[0], shadowAlphaAPIMarker)
			oInj, oPull := loadsOf(sos[0], shadowBetaAPIMarker)
			own := loadsInOrder(aInj, aPull)
			other := loadsInOrder(oInj, oPull)

			// The check's platform question — which file does an AMBIGUOUS
			// references/API.md resolve to — is only exercised if some
			// read used a bare, unqualified path. Models that qualify
			// every path sidestep the ambiguity themselves.
			sess := sos[0].Final().Session
			bareExercised := false
			for _, idx := range trace.ToolReadsOf(sess, "API.md") {
				in := string(sess.Events[idx].ToolCall.Input)
				if !strings.Contains(in, "probe-shadow-alpha") && !strings.Contains(in, "probe-shadow-beta") {
					bareExercised = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "bare references/API.md read (platform-side resolution exercised)"))
					break
				}
			}

			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			if !bareExercised && len(own)+len(other) > 0 {
				f.Confidence = ConfidenceInferred
				f.Notes = append(f.Notes, "every API.md read was skill-qualified; the platform's ambiguous-path resolution was never exercised, so the outcome reflects model path discipline, not platform disambiguation")
			}
			switch {
			case len(own) == 0 && len(other) == 0:
				f.Status = StatusInconclusive
				f.Verdict = "resource-not-read"
				f.Notes = append(f.Notes, "neither API.md version's canary ever arrived; the model skipped the read step")
			case len(own) > 0 && (len(other) == 0 || own[0].EventIndex <= other[0].EventIndex):
				f.Verdict = "own-resource-first"
				f.Evidence = append(f.Evidence, evAt(sos[0], own[0].EventIndex, "activated skill's own API.md content arrived first"))
				if len(other) > 0 {
					f.Notes = append(f.Notes, "the sibling skill's version was also accessible later in the session (no isolation between skills' resources)")
					f.Evidence = append(f.Evidence, evAt(sos[0], other[0].EventIndex, "sibling skill's API.md also read"))
				}
			default:
				f.Verdict = "shadowed-by-sibling"
				f.Evidence = append(f.Evidence, evAt(sos[0], other[0].EventIndex, "activated probe-shadow-alpha but probe-shadow-beta's API.md content arrived first"))
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// traversalBodyMarker detects probe-traversal activation; the skill has no
// body canary.
const traversalBodyMarker = "boundary at the skill"

func pathTraversalBoundary() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-traversal") }
	return Spec{
		ID:          "path-traversal-boundary",
		Description: "Can the model read outside the activated skill's directory (a sibling skill's file), and is anything visibly blocked?",
		Sessions: []Session{{
			// probe-loading is installed as the sibling the traversal
			// targets, so a failed read means a boundary, not a missing
			// file.
			Skills: []string{"probe-traversal", "probe-loading"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "path-traversal-boundary"}
			bInj, bPull := loadsOf(sos[0], traversalBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			attempts := trace.ToolReadsOf(sos[0].Final().Session, "../")
			sibInj, sibPull := loadsOf(sos[0], probeLoadingBodyCanary)
			f.Status = StatusObserved
			switch {
			case len(sibPull) > 0:
				f.Verdict = "outside-skill-read-allowed"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], sibPull[0].EventIndex, "sibling skill's body canary arrived in a tool result"))
				if len(attempts) > 0 {
					f.Notes = append(f.Notes, "relative ../ path forms were used")
					f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "first ../ path attempt"))
				} else {
					f.Notes = append(f.Notes, "no ../ form observed; the model reached the sibling via absolute or listing-derived paths. Access is unrestricted either way")
				}
			case len(sibInj) > 0:
				f.Verdict = "sibling-activated-not-read"
				f.Confidence = ConfidenceDirect
				f.Notes = append(f.Notes, "the model invoked the sibling as a skill instead of reading its file; boundary not exercised")
			case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
				f.Verdict = "traversal-blocked-visibly"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "../ attempt returned an error result"))
			case len(attempts) > 0:
				f.Status = StatusInconclusive
				f.Verdict = "attempted-no-content-no-error"
			default:
				f.Status = StatusInconclusive
				f.Verdict = "not-attempted"
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

const probeMetadataBodyCanary = "THRUSH-FLINT-8294"

// metadataEdgeKeyMarker shows edge-case metadata reached the model. It
// must be a string that exists ONLY in the frontmatter: the body prose
// names every metadata key, so a key name like "tagged-null" cannot
// distinguish frontmatter passthrough from an ordinary body load. The
// YAML tag "!!null" appears nowhere in the body.
const metadataEdgeKeyMarker = "!!null"

func metadataValueEdgeCases() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-metadata-values") }
	return Spec{
		ID:          "metadata-value-edge-cases",
		Description: "Is a skill whose metadata frontmatter holds nulls and empty strings still discovered and loaded, and do those keys reach the model?",
		Sessions: []Session{{
			Skills: []string{"probe-metadata-values"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "metadata-value-edge-cases"}

			// listed: 1 = in the discovery listing, -1 = absent from it,
			// 0 = unobservable (transcript omits injected context).
			listed := 0
			if obs.Profile.RecordsInjectedContext {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-metadata-values"); idx >= 0 {
					listed = 1
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names probe-metadata-values despite edge-case metadata"))
				} else {
					listed = -1
				}
			}

			inj, pull := loadsOf(sos[0], probeMetadataBodyCanary)
			attempts := trace.ToolReadsOf(obs.Session, "probe-metadata-values")
			switch {
			case len(inj)+len(pull) > 0:
				f.Status = StatusObserved
				f.Verdict = "loaded-despite-edge-case-metadata"
				f.Vehicle = vehicleOf(inj, pull)
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(inj, pull)[0].EventIndex, "body canary loaded"))
				kInj, kPull := loadsOf(sos[0], metadataEdgeKeyMarker)
				switch {
				case len(kInj) > 0:
					f.Notes = append(f.Notes, "edge-case metadata values visible in harness-injected content (frontmatter passed through)")
					f.Evidence = append(f.Evidence, evAt(sos[0], kInj[0].EventIndex, "frontmatter-only marker '!!null' in injected content"))
				case len(kPull) > 0:
					f.Notes = append(f.Notes, "edge-case metadata values reached the model only via its own raw file read")
					f.Evidence = append(f.Evidence, evAt(sos[0], kPull[0].EventIndex, "frontmatter-only marker '!!null' in a tool result"))
				default:
					f.Notes = append(f.Notes, "edge-case metadata values never reached the model (frontmatter stripped or withheld)")
				}
			case listed == -1:
				f.Status = StatusObserved
				f.Verdict = "not-discovered"
				f.Confidence = ConfidenceDirect
				f.Notes = append(f.Notes, "skill absent from the discovery listing; the edge-case metadata may have broken frontmatter parsing")
			case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
				f.Status = StatusObserved
				f.Verdict = "activation-failed-visibly"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "activation attempt returned an error result"))
			case len(attempts) > 0:
				f.Status = StatusInconclusive
				f.Verdict = "attempted-not-loaded"
			default:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// Canaries for the script-execution probe. The body canary lives in
// SKILL.md; the output phrase is assembled by scripts/emit-canary.sh at
// runtime and never appears joined in any file, so its presence in a
// tool result proves execution. The script's literal format string
// doubles as the source-was-read marker.
const (
	scriptExecBodyCanary   = "REDSHANK-SYENITE-8807"
	scriptExecOutputPhrase = "GODWIT-BORNITE-5148"
	scriptExecSourceMarker = "GODWIT-%s-5148"
)

func bundledScriptExecution() Spec {
	return Spec{
		ID:          "bundled-script-execution",
		Description: "Can the agent run a bundled scripts/ file and receive its output?",
		Sessions: []Session{{
			Skills: []string{"probe-script-execution"},
			Turns: []Turn{{
				Prompt:     func(p profile.Profile) string { return p.ActivationPrompt("probe-script-execution") },
				Activation: true,
			}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "bundled-script-execution"}
			bInj, bPull := loadsOf(sos[0], scriptExecBodyCanary)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Vehicle = VehicleNone
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect

			outOccs := trace.Phrase(obs.Session, scriptExecOutputPhrase, obs.Profile.EchoSubtypes)
			outResults := trace.At(outOccs, trace.LocToolResult)
			// A tool-call INPUT carrying the joined phrase means the model
			// composed the output itself (echo-style) instead of running
			// the script; genuine execution needs a call that references
			// the script plus the joined phrase arriving in a result.
			fabricated := trace.ToolReadsOf(obs.Session, scriptExecOutputPhrase)
			scriptCalls := trace.ToolReadsOf(obs.Session, "emit-canary.sh")
			srcOccs := trace.Phrase(obs.Session, scriptExecSourceMarker, obs.Profile.EchoSubtypes)
			srcRead := trace.At(srcOccs, trace.LocToolResult)

			switch {
			case len(outResults) > 0 && len(scriptCalls) > 0:
				f.Verdict = "script-executed"
				f.Evidence = append(f.Evidence,
					evAt(sos[0], scriptCalls[0], "tool call references the bundled script"),
					evAt(sos[0], outResults[0].EventIndex, "runtime-assembled output phrase arrived in a tool result"))
				if len(fabricated) > 0 {
					f.Notes = append(f.Notes, "at least one tool-call input also carried the joined phrase; check the transcript for whether every output occurrence traces to the script itself")
				}
			case len(outResults) > 0:
				f.Status = StatusInconclusive
				f.Verdict = "output-without-script-reference"
				f.Notes = append(f.Notes, "the output phrase appeared in a tool result, but no tool call referenced the script; likely reconstructed by the model rather than produced by the script")
			case len(scriptCalls) > 0 && errorResultsFor(sos[0], scriptCalls):
				f.Verdict = "execution-blocked-visibly"
				f.Evidence = append(f.Evidence, evAt(sos[0], scriptCalls[0], "script-referencing tool call returned an error result"))
			case len(srcRead) > 0:
				f.Verdict = "source-read-not-executed"
				f.Evidence = append(f.Evidence, evAt(sos[0], srcRead[0].EventIndex, "script source arrived in a tool result (literal format string); the assembled output never did"))
			case len(scriptCalls) > 0:
				f.Verdict = "attempted-no-output"
				f.Evidence = append(f.Evidence, evAt(sos[0], scriptCalls[0], "script-referencing tool call recorded; neither output nor an error result carried a traceable marker"))
			default:
				f.Verdict = "not-attempted"
				f.Confidence = negConfidence(sos[0])
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}
