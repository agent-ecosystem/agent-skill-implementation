// Discovery-and-validation checks (Category 10): scan locations,
// frontmatter validation strictness, and name-collision precedence.
package checks

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

// Body canaries, from the benchmark's canary index.
const (
	interopBodyCanary       = "SNIPE-OCHRE-2217"
	malformedYamlBodyCanary = "QUAIL-FELDSPAR-7448"
	noDescriptionBodyCanary = "VIREO-PUMICE-3049"
	collisionProjectCanary  = "RAVEN-CITRINE-6634"
	collisionUserCanary     = "PIPIT-SHALE-1147"
)

// catalogPresence reports whether skillName is in the platform's catalog:
// via the recorded discovery listing where available, else via the model's
// tool-free listing answer in turn 0. The bool result is only as direct as
// the second return value says.
func catalogPresence(so *observe.SessionObservation, skillName string) (listed bool, confidence string, evidence []Evidence) {
	obs := so.Final()
	if obs.Profile.RecordsInjectedContext {
		if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, skillName); idx >= 0 {
			return true, ConfidenceDirect, []Evidence{evAt(so, idx, "discovery listing names "+skillName)}
		}
		return false, ConfidenceDirect, nil
	}
	for _, o := range assistantMentions(so, skillName) {
		if turnOf(so, o.EventIndex) == 0 {
			return true, ConfidenceInferred, []Evidence{evAt(so, o.EventIndex, "tool-free listing answer names "+skillName)}
		}
	}
	return false, ConfidenceInferred, nil
}

// validationSpec builds the shared shape of the frontmatter-validation
// checks: listing turn, activation turn, catalog-first grading. Optional
// post hooks run on the assembled finding for check-specific notes (the
// oversize checks use one for marker-based truncation detection).
func validationSpec(id, description, skillDir, skillName, canary, loadedVerdict, skippedVerdict string, post ...func(*Finding, []*observe.SessionObservation)) Spec {
	return Spec{
		ID:          id,
		Description: description,
		Sessions: []Session{{
			Skills: []string{skillDir},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt(skillName) }, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: id}
			listed, conf, ev := catalogPresence(sos[0], skillName)
			f.Confidence = conf
			f.Evidence = append(f.Evidence, ev...)
			inj, pull := loadsOf(sos[0], canary)
			loaded := len(inj)+len(pull) > 0
			if loaded {
				f.Vehicle = vehicleOf(inj, pull)
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(inj, pull)[0].EventIndex, "body canary loaded on activation"))
			}
			f.Status = StatusObserved
			switch {
			case listed && loaded:
				f.Verdict = loadedVerdict
			case listed:
				f.Verdict = "listed-not-loaded"
			case loaded:
				f.Verdict = "unlisted-but-reachable"
				f.Notes = append(f.Notes, "not in the catalog, yet the body loaded: model file access, not platform acceptance")
			default:
				f.Verdict = skippedVerdict
			}
			for _, p := range post {
				p(&f, sos)
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func malformedYamlTolerance() Spec {
	return validationSpec(
		"malformed-yaml-tolerance",
		"Is a skill whose description holds an unquoted colon (invalid YAML) still discovered and loadable?",
		"probe-malformed-yaml", "probe-malformed-yaml", malformedYamlBodyCanary,
		"tolerated-and-loaded", "skipped-strict-parser")
}

func missingDescriptionHandling() Spec {
	return validationSpec(
		"missing-description-handling",
		"Is a skill with no description field skipped (as the guide prescribes), or loaded anyway?",
		"probe-no-description", "probe-no-description", noDescriptionBodyCanary,
		"loaded-despite-missing-description", "skipped-as-guide-prescribes")
}

func crossClientDirectoryInterop() Spec {
	return Spec{
		ID:          "cross-client-directory-interop",
		Description: "Is a skill installed only at the cross-client .agents/skills convention path discovered?",
		Sessions: []Session{{
			// The overlay lands the skill at <project>/.agents/skills/;
			// nothing is installed in the harness's native skills dir.
			OverlayDirs: []string{"overlay-agents-convention"},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt("probe-interop") }, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "cross-client-directory-interop"}
			native := obs.Profile.ProjectSkillDir == filepath.Join(".agents", "skills")
			listed, conf, ev := catalogPresence(sos[0], "probe-interop")
			f.Confidence = conf
			f.Evidence = append(f.Evidence, ev...)
			inj, pull := loadsOf(sos[0], interopBodyCanary)
			if len(inj)+len(pull) > 0 {
				f.Vehicle = vehicleOf(inj, pull)
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(inj, pull)[0].EventIndex, "body canary loaded on activation"))
			}
			f.Status = StatusObserved
			switch {
			case native && listed:
				f.Verdict = "convention-is-native-dir"
				f.Notes = append(f.Notes, "this platform's native project skills directory IS .agents/skills, so the check cannot separate convention support from native scanning")
			case listed:
				f.Verdict = "convention-scanned"
			case len(inj)+len(pull) > 0:
				f.Verdict = "unlisted-but-reachable"
				f.Notes = append(f.Notes, "not in the catalog, yet the body loaded: model file access, not convention scanning")
			default:
				f.Verdict = "convention-not-scanned"
				f.Notes = append(f.Notes, "a skill installed by another client at .agents/skills is invisible here")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func nameCollisionPrecedence() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-collision") }
	return Spec{
		ID:              "name-collision-precedence",
		Description:     "With the same skill name installed at project and user scope, which variant's content activates?",
		RequiresSandbox: true,
		Sessions: []Session{{
			Skills:     []string{"probe-collision"},
			UserSkills: []string{filepath.Join("probe-collision-user", "probe-collision")},
			Turns:      []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "name-collision-precedence"}
			pInj, pPull := loadsOf(sos[0], collisionProjectCanary)
			uInj, uPull := loadsOf(sos[0], collisionUserCanary)
			proj, user := len(pInj)+len(pPull) > 0, len(uInj)+len(uPull) > 0

			// Attribution: a push harness's activation tool resolves the
			// collision itself, but on pull harnesses the winner depends
			// on what the catalog showed. The variants' DESCRIPTIONS are
			// distinctive ("-scope variant"; bodies say "-level variant"),
			// so their presence in harness-injected text reveals whether
			// the catalog exposed one entry or both.
			winnerInjected := len(pInj) > 0 || len(uInj) > 0
			descProj, _ := loadsOf(sos[0], "PROJECT-scope variant")
			descUser, _ := loadsOf(sos[0], "USER-scope variant")
			bothListed := len(descProj) > 0 && len(descUser) > 0
			attribute := func() {
				obs := sos[0].Final()
				switch {
				case winnerInjected:
					f.Notes = append(f.Notes, "platform-resolved: the harness's activation mechanism injected the winning variant")
				case !obs.Profile.RecordsInjectedContext:
					f.Confidence = ConfidenceInferred
					f.Notes = append(f.Notes, "catalog visibility unrecorded on this harness; whether the platform or the model resolved the collision is not directly observable")
				case bothListed:
					f.Confidence = ConfidenceInferred
					f.Verdict = "catalog-lists-both-" + f.Verdict
					f.Notes = append(f.Notes, "the discovery listing exposed BOTH variants (both descriptions present); the model's file choice determined the winner: model-level selection, not platform precedence")
				default:
					f.Notes = append(f.Notes, "platform-resolved at discovery: the listing exposed a single variant")
				}
			}

			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			switch {
			case proj && !user:
				f.Verdict = "project-overrides-user"
				f.Vehicle = vehicleOf(pInj, pPull)
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(pInj, pPull)[0].EventIndex, "project variant's canary loaded; user variant's never appeared"))
				attribute()
			case user && !proj:
				f.Verdict = "user-overrides-project"
				f.Vehicle = vehicleOf(uInj, uPull)
				f.Evidence = append(f.Evidence, evAt(sos[0], loadsInOrder(uInj, uPull)[0].EventIndex, "user variant's canary loaded; project variant's never appeared"))
				f.Notes = append(f.Notes, "contradicts the guide's 'universal convention' that project-level overrides user-level")
				attribute()
			case proj && user:
				f.Verdict = "both-variants-loaded"
				f.Evidence = append(f.Evidence,
					evAt(sos[0], loadsInOrder(pInj, pPull)[0].EventIndex, "project variant's canary loaded"),
					evAt(sos[0], loadsInOrder(uInj, uPull)[0].EventIndex, "user variant's canary loaded"))
				f.Notes = append(f.Notes, "no shadowing at all: both scopes' content reached the model (check the transcript for whether the model or the platform did the merging)")
			default:
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// Fixtures and canaries for the spec-limit validation checks.
const (
	upperNameCanary    = "DUNLIN-OLIVINE-7821"
	doubleHyphenCanary = "PETREL-GALENA-3306"
	overlongNameCanary = "AVOCET-ZIRCON-5573"
	overlongNameDir    = "probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit"

	longDescBodyCanary   = "BITTERN-HALITE-2264"
	longDescHeadMarker   = "SANDERLING-GNEISS-1010"
	longDescTailMarker   = "WHIMBREL-DOLOMITE-2020"
	longCompatBodyCanary = "KESTREL-BAUXITE-6690"
	longCompatTailMarker = "TURNSTONE-ARAGONITE-3030"
)

func invalidNameTolerance() Spec {
	variants := []struct{ label, name, canary string }{
		{"uppercase", "probe-Upper-Case", upperNameCanary},
		{"double-hyphen", "probe--double-hyphen", doubleHyphenCanary},
		{"overlong", overlongNameDir, overlongNameCanary},
	}
	turns := []Turn{{Prompt: listingPrompt}}
	for _, v := range variants {
		name := v.name
		turns = append(turns, Turn{
			Prompt:     func(p profile.Profile) string { return p.ActivationPrompt(name) },
			Activation: true,
		})
	}
	dirs := make([]string, len(variants))
	for i, v := range variants {
		dirs[i] = v.name
	}
	return Spec{
		ID:          "invalid-name-tolerance",
		Description: "Are skills whose names break the spec's rules (uppercase, consecutive hyphens, over 64 characters) still discovered and loadable?",
		Sessions:    []Session{{Skills: dirs, Turns: turns}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "invalid-name-tolerance"}
			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			var tolerated, rejected []string
			for _, v := range variants {
				listed, conf, ev := catalogPresence(sos[0], v.name)
				if conf == ConfidenceInferred {
					f.Confidence = ConfidenceInferred
				}
				f.Evidence = append(f.Evidence, ev...)
				// A normalized listing (the uppercase name lowercased)
				// still counts as tolerated: the platform accepted the
				// skill, just under a repaired identity.
				normalized := false
				if !listed && v.name != strings.ToLower(v.name) {
					if lowListed, _, lowEv := catalogPresence(sos[0], strings.ToLower(v.name)); lowListed {
						listed, normalized = true, true
						f.Evidence = append(f.Evidence, lowEv...)
						f.Notes = append(f.Notes, v.label+": listed under a lowercased (normalized) name")
					}
				}
				inj, pull := loadsOf(sos[0], v.canary)
				loaded := len(inj)+len(pull) > 0
				if loaded && f.Vehicle == "" {
					f.Vehicle = vehicleOf(inj, pull)
				}
				switch {
				case listed:
					tolerated = append(tolerated, v.label)
					if !loaded && !normalized {
						f.Notes = append(f.Notes, v.label+": listed but its body never loaded on the activation turn")
					}
				case loaded:
					rejected = append(rejected, v.label)
					f.Notes = append(f.Notes, v.label+": not in the catalog, yet the body loaded: model file access, not platform acceptance")
				default:
					rejected = append(rejected, v.label)
				}
			}
			switch {
			case len(rejected) == 0:
				f.Verdict = "all-invalid-names-tolerated"
			case len(tolerated) == 0:
				f.Verdict = "all-invalid-names-rejected"
			default:
				f.Verdict = fmt.Sprintf("invalid-names-tolerated:%v", tolerated)
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func oversizeDescriptionHandling() Spec {
	return validationSpec(
		"oversize-description-handling",
		"Is a skill whose description exceeds the spec's 1024-character limit still discovered, and does the full value survive untruncated?",
		"probe-long-description", "probe-long-description", longDescBodyCanary,
		"loaded-despite-oversize-description", "skipped-oversize-description",
		func(f *Finding, sos []*observe.SessionObservation) {
			obs := sos[0].Final()
			// The body never spells out either marker, so a marker in
			// harness-injected content can only come from the platform's
			// own delivery of the description value.
			headInj := trace.At(trace.Phrase(obs.Session, longDescHeadMarker, obs.Profile.EchoSubtypes), trace.LocHarnessInjected)
			tailInj := trace.At(trace.Phrase(obs.Session, longDescTailMarker, obs.Profile.EchoSubtypes), trace.LocHarnessInjected)
			switch {
			case len(tailInj) > 0:
				f.Notes = append(f.Notes, "the description's tail marker reached the model in harness-injected content: the oversize value survived past 1024 characters untruncated")
				f.Evidence = append(f.Evidence, evAt(sos[0], tailInj[0].EventIndex, "description tail marker in injected content"))
			case len(headInj) > 0:
				f.Notes = append(f.Notes, "the description's head marker reached the model in injected content but its tail marker never did: the value was truncated somewhere after the head")
				f.Evidence = append(f.Evidence, evAt(sos[0], headInj[0].EventIndex, "description head marker in injected content; tail absent"))
			default:
				if len(assistantMentions(sos[0], longDescTailMarker)) > 0 {
					f.Notes = append(f.Notes, "the tail marker surfaced in the model's own text (nothing harness-injected records it); on this harness delivery is only inferable")
				} else {
					f.Notes = append(f.Notes, "neither description marker appeared in harness-injected content; description delivery is unobservable here or the value was dropped")
				}
			}
		})
}

func oversizeCompatibilityHandling() Spec {
	return validationSpec(
		"oversize-compatibility-handling",
		"Is a skill whose compatibility value exceeds the spec's 500-character limit still discovered and loadable?",
		"probe-long-compatibility", "probe-long-compatibility", longCompatBodyCanary,
		"loaded-despite-oversize-compatibility", "skipped-oversize-compatibility",
		func(f *Finding, sos []*observe.SessionObservation) {
			tInj, tPull := loadsOf(sos[0], longCompatTailMarker)
			switch {
			case len(tInj) > 0:
				f.Notes = append(f.Notes, "the compatibility value's tail marker reached the model in injected content: the oversize value survived past 500 characters")
			case len(tPull) > 0:
				f.Notes = append(f.Notes, "the compatibility value's tail marker is visible only via the model's own raw file read")
			default:
				f.Notes = append(f.Notes, "the compatibility value's tail marker never reached the model")
			}
		})
}
