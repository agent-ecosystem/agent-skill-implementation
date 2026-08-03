// Content-presentation and access-control checks: frontmatter handling,
// content wrapping, and the compatibility field.
package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

// Markers for discovery-listing-fields, each unique to one frontmatter
// slot of the installed fixtures: the description marker exists only in
// probe-metadata-values' description, the path marker only as an on-disk
// path (no file content mentions it). compatValueMarker and
// metadataEdgeKeyMarker (both frontmatter-value-only) are shared with the
// other presentation checks.
const (
	metadataDescMarker = "asked to probe metadata value handling"
	compatPathMarker   = "probe-compatibility/SKILL.md"
)

// discoveryListingPrompt asks the model to echo its catalog verbatim. On
// recording harnesses the injected listing itself is the ground truth; on
// harnesses that record no injected context, the model's tool-free echo is
// the only observable, so the prompt maximizes what it reveals.
func discoveryListingPrompt(profile.Profile) string {
	return "Without using any tools, list every skill in your Available skills list. For each skill, reproduce exactly everything you can see about it: its name, its description, and any other fields or values shown to you. Do not summarize; copy the entries verbatim."
}

func discoveryListingFields() Spec {
	return Spec{
		ID:          "discovery-listing-fields",
		Description: "Which frontmatter fields does the discovery listing surface to the model: name and description only, or also compatibility, metadata values, or file locations?",
		Sessions: []Session{{
			// Three fixtures covering the field types: compatibility value,
			// edge-case metadata values, plus ordinary description text.
			Skills: []string{"probe-loading", "probe-compatibility", "probe-metadata-values"},
			// Passive turn: no tools, no activation — every marker that
			// appears was surfaced at discovery.
			Turns: []Turn{{Prompt: discoveryListingPrompt}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "discovery-listing-fields"}
			recording := obs.Profile.RecordsInjectedContext

			seen := func(marker string) (bool, int) {
				if recording {
					inj, _ := loadsOf(sos[0], marker)
					if len(inj) > 0 {
						return true, inj[0].EventIndex
					}
					return false, -1
				}
				for _, o := range assistantMentions(sos[0], marker) {
					return true, o.EventIndex
				}
				return false, -1
			}

			nameOK := false
			if recording {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-compatibility"); idx >= 0 {
					nameOK = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names the installed skills"))
				}
				f.Confidence = ConfidenceDirect
			} else {
				var idx int
				nameOK, idx = seen("probe-compatibility")
				if nameOK {
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "tool-free catalog echo names the installed skills"))
				}
				f.Confidence = ConfidenceInferred
				f.Notes = append(f.Notes, "this harness records no injected context; surfaced fields are inferred from the model's tool-free verbatim echo of its catalog")
			}
			if !nameOK {
				f.Status = StatusInconclusive
				f.Verdict = "listing-not-observed"
				return f
			}

			fields := []struct{ name, marker, label string }{
				{"description", metadataDescMarker, "description text surfaced at discovery"},
				{"compatibility", compatValueMarker, "compatibility value surfaced at discovery"},
				{"metadata", metadataEdgeKeyMarker, "edge-case metadata value surfaced at discovery"},
				{"location", compatPathMarker, "SKILL.md file path surfaced at discovery"},
			}
			descOK := false
			var extended []string
			for _, fl := range fields {
				ok, idx := seen(fl.marker)
				if !ok {
					continue
				}
				if fl.name == "description" {
					descOK = true
				} else {
					extended = append(extended, fl.name)
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], idx, fl.label))
			}
			f.Status = StatusObserved
			switch {
			case len(extended) > 0:
				f.Verdict = fmt.Sprintf("surfaces-beyond-description:%v", extended)
			case descOK:
				f.Verdict = "name-and-description-only"
			default:
				f.Verdict = "names-surfaced-description-unconfirmed"
				f.Notes = append(f.Notes, "skill names surfaced but no description marker appeared; the model may have paraphrased rather than echoed verbatim")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// probeLoadingFrontmatterMarker exists ONLY in probe-loading's frontmatter
// (the compatibility value); the body names the frontmatter FIELDS, so
// field names cannot serve as markers.
const probeLoadingFrontmatterMarker = "Requires filesystem access"

func frontmatterHandling() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-loading") }
	return Spec{
		ID:          "frontmatter-handling",
		Description: "Does the SKILL.md YAML frontmatter reach the model at activation, or only the body?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "frontmatter-handling"}
			bInj, bPull := loadsOf(sos[0], probeLoadingBodyCanary)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			fmInj, fmPull := loadsOf(sos[0], probeLoadingFrontmatterMarker)
			f.Status = StatusObserved
			switch {
			case len(fmInj) > 0:
				f.Verdict = "frontmatter-passed-through"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], fmInj[0].EventIndex, "frontmatter-only marker present in injected content"))
			case len(bInj) > 0:
				f.Verdict = "frontmatter-stripped-on-injection"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], bInj[0].EventIndex, "body injected without the frontmatter-only marker"))
				if len(fmPull) > 0 {
					f.Notes = append(f.Notes, "the model later saw the frontmatter via its own raw file read")
				}
			case len(fmPull) > 0:
				f.Verdict = "frontmatter-visible-via-raw-read"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], fmPull[0].EventIndex, "raw file read delivered frontmatter and body together"))
				f.Notes = append(f.Notes, "pull-vehicle harness: the model reads the file as-is, so frontmatter visibility is inherent, not a platform presentation choice")
			default:
				f.Status = StatusInconclusive
				f.Verdict = "frontmatter-not-observed"
				f.Notes = append(f.Notes, "body arrived via pull but the frontmatter marker never did (partial read or unexpected read tool behavior)")
			}
			return f
		},
	}
}

// tagPattern matches XML-ish tags for wrapping inspection.
var tagPattern = regexp.MustCompile(`<(/?[a-zA-Z][a-zA-Z0-9_:-]*)(?:\s[^>]*)?>`)

// tagsNear returns the distinct tag names within a window around the first
// occurrence of phrase in text.
func tagsNear(text, phrase string) []string {
	i := strings.Index(text, phrase)
	if i < 0 {
		return nil
	}
	lo := max(i-800, 0)
	hi := min(i+len(phrase)+800, len(text))
	seen := map[string]bool{}
	var out []string
	for _, m := range tagPattern.FindAllStringSubmatch(text[lo:hi], -1) {
		name := strings.TrimPrefix(m[1], "/")
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

func contentWrappingFormat() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-loading") }
	return Spec{
		ID:          "content-wrapping-format",
		Description: "Is injected skill content wrapped in structured tags, or delivered as raw markdown, and what does the model see on pull harnesses?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "content-wrapping-format"}
			inj, pulled := loadsOf(sos[0], probeLoadingBodyCanary)
			switch {
			case len(inj) > 0:
				f.Status = StatusObserved
				f.Vehicle = VehicleHarnessPush
				f.Confidence = ConfidenceDirect
				text := eventVisibleText(sos[0], inj[0].EventIndex)
				tags := tagsNear(text, probeLoadingBodyCanary)
				if len(tags) > 0 {
					f.Verdict = "wrapped-structured"
					f.Notes = append(f.Notes, "tag-like tokens near the injected body: "+strings.Join(tags, ", "))
				} else {
					f.Verdict = "raw-injection"
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], inj[0].EventIndex, "injection event carrying the body canary"))
			case len(pulled) > 0:
				f.Status = StatusObserved
				f.Vehicle = VehicleModelPull
				f.Confidence = ConfidenceDirect
				f.Verdict = "raw-file-via-pull"
				f.Evidence = append(f.Evidence, evAt(sos[0], pulled[0].EventIndex, "body arrived as a file-read tool result"))
				f.Notes = append(f.Notes, "pull-vehicle harness: content arrives as the read tool formats it (line numbers etc.), not wrapped skill markup")
			default:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
			}
			return f
		},
	}
}

// Compatibility markers: the body marker detects activation (no canary in
// the index); the value marker exists only in the frontmatter field.
const (
	compatBodyMarker  = "platform and environment requirements"
	compatValueMarker = "Designed for Claude Code"
)

func compatibilityFieldBehavior() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-compatibility") }
	return Spec{
		ID:          "compatibility-field-behavior",
		Description: "Does a compatibility field naming another platform gate loading, get surfaced to the model, or get ignored?",
		Sessions: []Session{{
			Skills: []string{"probe-compatibility"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "compatibility-field-behavior"}

			listed := 0 // 1 = in the listing, -1 = absent, 0 = unobservable
			if obs.Profile.RecordsInjectedContext {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-compatibility"); idx >= 0 {
					listed = 1
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names probe-compatibility despite the foreign-platform compatibility value"))
				} else {
					listed = -1
				}
			}
			bInj, bPull := loadsOf(sos[0], compatBodyMarker)
			attempts := trace.ToolReadsOf(obs.Session, "probe-compatibility")
			switch {
			case len(bInj)+len(bPull) > 0:
				f.Status = StatusObserved
				f.Verdict = "activated-no-gating"
				f.Vehicle = vehicleOf(bInj, bPull)
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(bInj, bPull)[0].EventIndex, "body loaded normally"))
				vInj, vPull := loadsOf(sos[0], compatValueMarker)
				switch {
				case len(vInj) > 0:
					f.Notes = append(f.Notes, "compatibility value reached the model in injected content")
					f.Evidence = append(f.Evidence, evAt(sos[0], vInj[0].EventIndex, "compatibility value in injected content"))
				case len(vPull) > 0:
					f.Notes = append(f.Notes, "compatibility value visible only via the model's own raw file read")
				default:
					f.Notes = append(f.Notes, "compatibility value never reached the model (stripped or withheld)")
				}
			case listed == -1:
				f.Status = StatusObserved
				f.Verdict = "not-listed-possible-gating"
				f.Confidence = ConfidenceDirect
				f.Notes = append(f.Notes, "skill absent from the discovery listing; the compatibility field is a candidate cause, but so is any parse issue")
			case len(attempts) > 0 && errorResultsFor(sos[0], attempts):
				f.Status = StatusObserved
				f.Verdict = "activation-blocked-visibly"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], attempts[0], "activation attempt returned an error result"))
			default:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// Markers for the allowed-tools twins. Each skill's instructed printf
// assembles its phrase at runtime, so the joined form never appears in
// any file: joined-in-a-tool-result proves a command executed. The
// field marker is the literal allowed-tools value.
const (
	allowedToolsBodyCanary        = "CURLEW-SCHIST-4419"
	allowedToolsExecPhrase        = "GROUSE-MICA-9017"
	allowedToolsExecFragment      = "GROUSE-%s-9017"
	allowedToolsControlBodyCanary = "STINT-MARBLE-9912"
	allowedToolsControlExecPhrase = "LAPWING-FLUORITE-2260"
	allowedToolsControlFragment   = "LAPWING-%s-2260"
	allowedToolsValueMarker       = "Bash(printf:*) Read"
)

// execTier grades one allowed-tools session: did the instructed command
// run, get blocked, or never get attempted? Mirrors the script check's
// fabrication guard: a tool-call INPUT carrying the joined phrase means
// the model composed the output (echo-style) instead of, or in addition
// to, issuing the instructed printf. That still demonstrates the
// permission posture, but not the instructed command running, so the
// finding self-flags it.
func execTier(so *observe.SessionObservation, bodyCanary, phrase, fragment string) (tier string, ev []Evidence, notes []string) {
	bInj, bPull := loadsOf(so, bodyCanary)
	if len(bInj)+len(bPull) == 0 {
		return "activation-not-observed", nil, nil
	}
	obs := so.Final()
	results := trace.At(trace.Phrase(obs.Session, phrase, obs.Profile.EchoSubtypes), trace.LocToolResult)
	attempts := trace.ToolReadsOf(obs.Session, fragment)
	fabricated := trace.ToolReadsOf(obs.Session, phrase)
	switch {
	case len(results) > 0:
		ev = []Evidence{{Note: "runtime-assembled phrase arrived in a tool result", EventIndex: results[0].EventIndex}}
		switch {
		case len(attempts) == 0 && len(fabricated) > 0:
			notes = append(notes, "the joined phrase appears in a tool-call input and the instructed printf was never issued: the model composed the output itself; treat this as evidence of the permission posture, not of the instructed command running")
		case len(fabricated) > 0:
			notes = append(notes, "a tool-call input also carried the joined phrase; check the transcript for whether every output occurrence traces to the instructed command")
		}
		return "executed", ev, notes
	case len(attempts) > 0 && errorResultsFor(so, attempts):
		return "blocked-visibly", []Evidence{{Note: "instructed command's tool call returned an error result", EventIndex: attempts[0]}}, nil
	case len(attempts) > 0:
		return "attempted-no-output", nil, nil
	}
	return "not-attempted", nil, nil
}

func allowedToolsBehavior() Spec {
	return Spec{
		ID:          "allowed-tools-behavior",
		Description: "Does the experimental allowed-tools field pre-approve anything, compared against an identical skill without it?",
		Sessions: []Session{
			{
				Skills: []string{"probe-allowed-tools"},
				Turns: []Turn{{
					Prompt:     func(p profile.Profile) string { return p.ActivationPrompt("probe-allowed-tools") },
					Activation: true,
				}},
			},
			{
				Skills: []string{"probe-allowed-tools-control"},
				Turns: []Turn{{
					Prompt:     func(p profile.Profile) string { return p.ActivationPrompt("probe-allowed-tools-control") },
					Activation: true,
				}},
			},
		},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "allowed-tools-behavior"}
			fieldTier, fieldEv, fieldNotes := execTier(sos[0], allowedToolsBodyCanary, allowedToolsExecPhrase, allowedToolsExecFragment)
			ctrlTier, ctrlEv, ctrlNotes := execTier(sos[1], allowedToolsControlBodyCanary, allowedToolsControlExecPhrase, allowedToolsControlFragment)
			for _, e := range fieldEv {
				f.Evidence = append(f.Evidence, evSession(sos[0], 1, e.EventIndex, "with-field session: "+e.Note))
			}
			for _, e := range ctrlEv {
				f.Evidence = append(f.Evidence, evSession(sos[1], 2, e.EventIndex, "control session: "+e.Note))
			}
			for _, n := range fieldNotes {
				f.Notes = append(f.Notes, "with-field session: "+n)
			}
			for _, n := range ctrlNotes {
				f.Notes = append(f.Notes, "control session: "+n)
			}

			vInj, vPull := loadsOf(sos[0], allowedToolsValueMarker)
			switch {
			case len(vInj) > 0:
				f.Notes = append(f.Notes, "allowed-tools value reached the model in injected content")
			case len(vPull) > 0:
				f.Notes = append(f.Notes, "allowed-tools value visible only via the model's own raw file read")
			default:
				f.Notes = append(f.Notes, "allowed-tools value never reached the model")
			}

			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			switch {
			case fieldTier == "executed" && ctrlTier == "executed":
				f.Verdict = "executed-regardless-of-field"
				f.Notes = append(f.Notes, "the platform's general permission posture allowed both commands; any pre-approval effect of allowed-tools is unobservable under this posture")
			case fieldTier == "executed":
				f.Verdict = "field-enabled-execution"
				f.Notes = append(f.Notes, "control session tier: "+ctrlTier)
			case ctrlTier == "executed":
				f.Verdict = "field-blocked-execution"
				f.Notes = append(f.Notes, "with-field session tier: "+fieldTier+"; execution succeeding only WITHOUT the field is unexpected and worth a transcript read")
			case fieldTier == "activation-not-observed" && ctrlTier == "activation-not-observed":
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
			default:
				f.Verdict = "with-field:" + fieldTier + "; control:" + ctrlTier
			}
			f.Notes = append(f.Notes,
				"with-field final answer: "+finalAnswer(sos[0]),
				"control final answer: "+finalAnswer(sos[1]))
			return f
		},
	}
}
