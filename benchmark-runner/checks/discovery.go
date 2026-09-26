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
	multirootNativeCanary   = "GREBE-AZURITE-7301"
	multirootAgentsCanary   = "BUNTING-SERPENTINE-4185"
	multirootClaudeCanary   = "NIGHTJAR-KYANITE-6072"
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

// Fixtures and markers for the description-length-unit check. Read
// alongside probe-long-description (ASCII, over 1024 in every unit), the
// two fixtures below separate what an enforcing platform counts: the
// multibyte one is under 1024 code points and UTF-16 units but over in
// UTF-8 bytes; the astral one is under 1024 code points but over in both
// UTF-16 units and bytes.
const (
	multibyteDescBodyCanary = "PUFFIN-BASALT-4471"
	multibyteDescHeadMarker = "GANNET-PYRITE-1130"
	multibyteDescTailMarker = "SHRIKE-TALC-2210"
	astralDescBodyCanary    = "ORIOLE-GRANITE-5583"
	astralDescHeadMarker    = "MAGPIE-OBSIDIAN-1240"
	astralDescTailMarker    = "LINNET-MALACHITE-2420"
)

// descriptionFate is what became of one fixture's description value on
// its way to the model.
type descriptionFate string

const (
	descIntact     descriptionFate = "intact"     // tail marker reached the model
	descTruncated  descriptionFate = "truncated"  // head marker reached the model, tail did not
	descRejected   descriptionFate = "rejected"   // skill absent from the catalog
	descUnobserved descriptionFate = "unobserved" // listed, but marker delivery is not recorded
)

// descriptionFateOf grades one fixture session. Marker delivery in
// harness-injected content is the direct signal; where the harness does
// not record what it injected, the model's own quoting of the markers in
// its listing answer stands in.
func descriptionFateOf(so *observe.SessionObservation, sessionNum int, skillName, bodyCanary, head, tail string) (descriptionFate, []Evidence) {
	obs := so.Final()
	listed, _, ev := catalogPresence(so, skillName)
	for i := range ev {
		ev[i].Session = sessionNum
	}
	if inj, pull := loadsOf(so, bodyCanary); len(inj)+len(pull) > 0 {
		ev = append(ev, evSession(so, sessionNum, loadsInOrder(inj, pull)[0].EventIndex, "body canary loaded on activation"))
	}
	headInj := trace.At(trace.Phrase(obs.Session, head, obs.Profile.EchoSubtypes), trace.LocHarnessInjected)
	tailInj := trace.At(trace.Phrase(obs.Session, tail, obs.Profile.EchoSubtypes), trace.LocHarnessInjected)
	switch {
	case len(tailInj) > 0:
		return descIntact, append(ev, evSession(so, sessionNum, tailInj[0].EventIndex, "description tail marker in injected content"))
	case len(headInj) > 0:
		return descTruncated, append(ev, evSession(so, sessionNum, headInj[0].EventIndex, "description head marker in injected content; tail absent"))
	case !listed:
		return descRejected, ev
	}
	if m := assistantMentions(so, tail); len(m) > 0 {
		return descIntact, append(ev, evSession(so, sessionNum, m[0].EventIndex, "model quoted the tail marker from its catalog entry (inferred)"))
	}
	if m := assistantMentions(so, head); len(m) > 0 {
		return descTruncated, append(ev, evSession(so, sessionNum, m[0].EventIndex, "model quoted the head marker but never the tail (inferred)"))
	}
	return descUnobserved, ev
}

func descriptionLengthUnit() Spec {
	fixtures := []struct{ label, dir, body, head, tail string }{
		{"ascii", "probe-long-description", longDescBodyCanary, longDescHeadMarker, longDescTailMarker},
		{"multibyte", "probe-multibyte-description", multibyteDescBodyCanary, multibyteDescHeadMarker, multibyteDescTailMarker},
		{"astral", "probe-astral-description", astralDescBodyCanary, astralDescHeadMarker, astralDescTailMarker},
	}
	sessions := make([]Session, 0, len(fixtures))
	for _, fx := range fixtures {
		dir := fx.dir
		sessions = append(sessions, Session{
			Skills: []string{dir},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt(dir) }, Activation: true},
			},
		})
	}
	enforced := func(d descriptionFate) bool { return d == descTruncated || d == descRejected }
	return Spec{
		ID:          "description-length-unit",
		Description: "When a platform enforces the 1024-character description limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes?",
		Sessions:    sessions,
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "description-length-unit", Status: StatusObserved, Confidence: ConfidenceDirect}
			fates := make([]descriptionFate, len(fixtures))
			summary := make([]string, 0, len(fixtures))
			for i, fx := range fixtures {
				fate, ev := descriptionFateOf(sos[i], i+1, fx.dir, fx.body, fx.head, fx.tail)
				fates[i] = fate
				f.Evidence = append(f.Evidence, ev...)
				summary = append(summary, fx.label+":"+string(fate))
				if !sos[i].Final().Profile.RecordsInjectedContext {
					f.Confidence = ConfidenceInferred
				}
				f.Notes = append(f.Notes, fx.label+" final answer: "+finalAnswer(sos[i]))
				if inj, pull := loadsOf(sos[i], fx.body); f.Vehicle == "" && len(inj)+len(pull) > 0 {
					f.Vehicle = vehicleOf(inj, pull)
				}
			}
			f.Notes = append([]string{"description fates: " + strings.Join(summary, ", ")}, f.Notes...)
			ascii, multibyte, astral := fates[0], fates[1], fates[2]
			switch {
			case ascii == descUnobserved || multibyte == descUnobserved || astral == descUnobserved:
				f.Status = StatusInconclusive
				f.Verdict = "length-unit-unobservable"
				f.Notes = append(f.Notes, "at least one fixture's description delivery could not be observed, so the counting unit cannot be inferred")
			case ascii == descIntact && multibyte == descIntact && astral == descIntact:
				f.Verdict = "no-length-enforcement"
				f.Notes = append(f.Notes, "all three descriptions reached the model intact, including the ASCII one that exceeds 1024 in every unit: the platform does not enforce the limit, so its counting unit is moot")
			case enforced(ascii) && multibyte == descIntact && astral == descIntact:
				f.Verdict = "counts-code-points"
				f.Notes = append(f.Notes, "the ASCII overrun was enforced while both fixtures under 1024 code points survived intact: the platform counts code points, the reference validator's unit")
			case enforced(ascii) && multibyte == descIntact && enforced(astral):
				f.Verdict = "counts-utf16-units"
				f.Notes = append(f.Notes, "the ASCII and astral overruns were enforced while the multibyte fixture survived: the platform counts UTF-16 code units (JavaScript's .length)")
			case enforced(ascii) && enforced(multibyte) && enforced(astral):
				f.Verdict = "counts-bytes"
				f.Notes = append(f.Notes, "all three fixtures were enforced, including the two under 1024 code points: the platform counts UTF-8 bytes")
			default:
				f.Verdict = "inconsistent-length-unit"
				f.Notes = append(f.Notes, "the three outcomes fit no single counting unit; see the per-fixture fates")
			}
			return f
		},
	}
}

// multirootVariant is one copy of probe-multiroot: where it lands, the
// canary that proves its body loaded, and the description-only phrase that
// proves its catalog entry was listed (bodies say "-root variant" in
// lowercase, so the uppercase description marker is listing-only).
type multirootVariant struct {
	overlay, root, canary, descMarker, label string
	// beacon names a non-colliding control skill the same overlay installs
	// beside the variant; its presence in the catalog proves the root was
	// scanned, so a missing variant was dropped by name, not by the scan.
	beacon string
}

var multirootForeign = []multirootVariant{
	{"overlay-multiroot-agents", filepath.Join(".agents", "skills"), multirootAgentsCanary, "AGENTS-root variant", ".agents/skills", "probe-multiroot-beacon-agents"},
	{"overlay-multiroot-claude", filepath.Join(".claude", "skills"), multirootClaudeCanary, "CLAUDE-root variant", ".claude/skills", "probe-multiroot-beacon-claude"},
}

// multirootVariantsFor returns the native variant plus every foreign
// variant whose root is not the harness's own skills directory.
func multirootVariantsFor(p profile.Profile) []multirootVariant {
	out := []multirootVariant{{"", p.ProjectSkillDir, multirootNativeCanary, "NATIVE-root variant", "native", ""}}
	for _, v := range multirootForeign {
		if v.root != p.ProjectSkillDir {
			out = append(out, v)
		}
	}
	return out
}

func multiRootCollisionPrecedence() Spec {
	return Spec{
		ID:          "multi-root-collision-precedence",
		Description: "With the same skill name installed under two project roots the platform scans (its native directory plus .agents/skills or .claude/skills), which variant is listed and which activates?",
		Sessions: []Session{{
			Skills: []string{"probe-multiroot"},
			OverlayDirsFor: func(p profile.Profile) []string {
				var out []string
				for _, v := range multirootVariantsFor(p)[1:] {
					out = append(out, v.overlay)
				}
				return out
			},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt("probe-multiroot") }, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			so := sos[0]
			obs := so.Final()
			f := Finding{CheckID: "multi-root-collision-precedence"}
			variants := multirootVariantsFor(obs.Profile)
			var installed []string
			for _, v := range variants[1:] {
				installed = append(installed, v.label)
			}
			f.Notes = append(f.Notes, "foreign roots installed alongside the native copy: "+strings.Join(installed, ", "))

			// Scan control: each foreign overlay also installs a beacon
			// skill with a unique name. A listed beacon proves its root
			// was scanned.
			var scanned []string
			for _, v := range variants[1:] {
				if listed, conf, ev := catalogPresence(so, v.beacon); listed {
					scanned = append(scanned, v.label)
					f.Evidence = append(f.Evidence, ev...)
					_ = conf
				}
			}

			// Listing: which variants' description markers reached the
			// model on the listing turn (turn 0).
			var listed []string
			for _, v := range variants {
				if obs.Profile.RecordsInjectedContext {
					inj, _ := loadsOf(so, v.descMarker)
					for _, o := range inj {
						if turnOf(so, o.EventIndex) == 0 {
							listed = append(listed, v.label)
							f.Evidence = append(f.Evidence, evAt(so, o.EventIndex, "discovery listing carries the "+v.label+" variant's description"))
							break
						}
					}
					continue
				}
				for _, o := range assistantMentions(so, v.descMarker) {
					if turnOf(so, o.EventIndex) == 0 {
						listed = append(listed, v.label)
						f.Evidence = append(f.Evidence, evAt(so, o.EventIndex, "tool-free listing answer echoes the "+v.label+" variant's description"))
						break
					}
				}
			}

			// Activation: the first variant whose body loaded is the one
			// the activation resolved to; any later load is the model
			// exploring (on pull harnesses it may read every copy).
			var loaded []string
			var injAny, pullAny []trace.Occurrence
			firstAt := -1
			for _, v := range variants {
				inj, pull := loadsOf(so, v.canary)
				if len(inj)+len(pull) == 0 {
					continue
				}
				at := loadsInOrder(inj, pull)[0].EventIndex
				if firstAt < 0 || at < firstAt {
					firstAt = at
					loaded = append([]string{v.label}, loaded...)
				} else {
					loaded = append(loaded, v.label)
				}
				injAny = append(injAny, inj...)
				pullAny = append(pullAny, pull...)
				f.Evidence = append(f.Evidence, evAt(so, at, v.label+" variant's body canary loaded"))
			}
			if len(loaded) > 1 {
				f.Notes = append(f.Notes, "the model also read the "+strings.Join(loaded[1:], " and ")+" variant(s) after the first load; graded on the first load, the rest is exploration")
				loaded = loaded[:1]
			}
			if len(loaded) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "final answer: "+finalAnswer(so))
				return f
			}
			f.Status = StatusObserved
			f.Vehicle = vehicleOf(injAny, pullAny)
			if obs.Profile.RecordsInjectedContext {
				f.Confidence = ConfidenceDirect
			} else {
				f.Confidence = ConfidenceInferred
				f.Notes = append(f.Notes, "listing evidence rests on the model's catalog echo; the platform records no injected context")
			}
			foreignListed := 0
			for _, l := range listed {
				if l != "native" {
					foreignListed++
				}
			}
			switch {
			case len(listed) > 1:
				f.Verdict = "catalog-lists-all; activated:" + strings.Join(loaded, "+")
				f.Notes = append(f.Notes, "model-level: the catalog exposed more than one entry for the name, so which variant activated was the model's choice")
			case foreignListed == 0 && len(loaded) == 1 && loaded[0] == "native" && len(scanned) == 0:
				f.Verdict = "foreign-roots-not-scanned"
				f.Notes = append(f.Notes, "neither the convention-root variants nor their beacon skills were listed; the platform reads only its native directory (consistent with cross-client-directory-interop)")
			case len(loaded) == 1 && loaded[0] == "native":
				f.Verdict = "native-root-wins"
				if len(scanned) > 0 {
					f.Notes = append(f.Notes, "platform-level: the beacon skills from "+strings.Join(scanned, ", ")+" were listed, so those roots were scanned and the colliding variant was dropped by name in favor of the native copy")
				}
			case len(loaded) == 1:
				f.Verdict = "foreign-root-wins:" + loaded[0]
			default:
				f.Verdict = "foreign-root-wins:" + loaded[0]
			}
			if len(listed) > 0 {
				f.Notes = append(f.Notes, "listed variants: "+strings.Join(listed, ", "))
			}
			if len(scanned) > 0 {
				f.Notes = append(f.Notes, "roots proven scanned by their beacon: "+strings.Join(scanned, ", "))
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(so))
			return f
		},
	}
}

// nameLengthFixtures are the name-length-unit probes: names whose length in
// code points, UTF-16 code units, and UTF-8 bytes diverge. The 72-character
// ASCII fixture from invalid-name-tolerance proves enforcement exists; the
// exact-64 ASCII fixture is the boundary control; the short Greek name is
// under the cap in every unit, so its absence means non-ASCII names are
// rejected outright rather than counted. Descriptions carry a listing-only
// marker phrase because a model's catalog echo may not reproduce a long
// non-ASCII name verbatim.
var nameLengthFixtures = []struct{ label, dir, marker string }{
	{"ascii64", "probe-name-at-exactly-sixty-four-characters-to-mark-the-cap-abcd", "Length-unit probe ascii-sixty-four"},
	{"ascii72", "probe-overlong-name-padded-well-past-the-spec-sixty-four-character-limit", ""},
	{"greek16", "probe-αβγδεζηθικ", "Length-unit probe greek-sixteen"},
	{"greek60", "probe-αβγδεζηθικλμνξοπρστυφχψωαβγδεζηθικλμνξοπρστυφχψωαβγδεζ", "Length-unit probe greek-sixty"},
	{"math40", "probe-𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡", "Length-unit probe math-forty"},
}

// listedByNameOrMarker reports catalog presence via the skill name or,
// failing that, a description-only marker phrase.
func listedByNameOrMarker(so *observe.SessionObservation, name, marker string) (bool, []Evidence) {
	if listed, _, ev := catalogPresence(so, name); listed {
		return true, ev
	}
	if marker == "" {
		return false, nil
	}
	obs := so.Final()
	if obs.Profile.RecordsInjectedContext {
		inj, _ := loadsOf(so, marker)
		for _, o := range inj {
			if turnOf(so, o.EventIndex) == 0 {
				return true, []Evidence{evAt(so, o.EventIndex, "discovery listing carries the description marker for "+name)}
			}
		}
		return false, nil
	}
	for _, o := range assistantMentions(so, marker) {
		if turnOf(so, o.EventIndex) == 0 {
			return true, []Evidence{evAt(so, o.EventIndex, "tool-free listing answer echoes the description marker for "+name)}
		}
	}
	return false, nil
}

func nameLengthUnit() Spec {
	var dirs []string
	for _, fx := range nameLengthFixtures {
		dirs = append(dirs, fx.dir)
	}
	return Spec{
		ID:          "name-length-unit",
		Description: "When a platform enforces the 64-character name limit, does it count Unicode code points, UTF-16 code units, or UTF-8 bytes, or does it reject non-ASCII names regardless of length?",
		Sessions: []Session{{
			Skills: dirs,
			Turns:  []Turn{{Prompt: listingPrompt}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			so := sos[0]
			f := Finding{CheckID: "name-length-unit", Status: StatusObserved, Confidence: ConfidenceDirect}
			if !so.Final().Profile.RecordsInjectedContext {
				f.Confidence = ConfidenceInferred
			}
			listed := map[string]bool{}
			var summary []string
			for _, fx := range nameLengthFixtures {
				ok, ev := listedByNameOrMarker(so, fx.dir, fx.marker)
				listed[fx.label] = ok
				f.Evidence = append(f.Evidence, ev...)
				fate := "skipped"
				if ok {
					fate = "listed"
				}
				summary = append(summary, fx.label+":"+fate)
			}
			f.Notes = append(f.Notes, "name fates: "+strings.Join(summary, ", "))
			if !listed["ascii64"] {
				f.Notes = append(f.Notes, "the exact-64 ASCII control was not listed: the cap sits below 64 or another rule fired; read the per-name fates with care")
			}
			switch {
			case listed["ascii72"] && listed["greek16"]:
				f.Verdict = "no-length-enforcement"
				f.Notes = append(f.Notes, "the 72-character ASCII name was listed, so the platform does not enforce the cap and its counting unit is moot; non-ASCII names were accepted too")
			case listed["ascii72"]:
				f.Verdict = "no-length-enforcement; rejects-non-ascii-names"
				f.Notes = append(f.Notes, "the 72-character ASCII name was listed, so the platform does not enforce the cap; it does reject non-ASCII names, since the 16-code-point Greek name was absent")
			case !listed["greek16"]:
				f.Verdict = "rejects-non-ascii-names"
				f.Notes = append(f.Notes, "the platform enforces the cap (72 ASCII rejected) but also dropped the 16-code-point Greek name, which is under 64 in every unit: non-ASCII names are rejected on the character rule, so the counting unit cannot be observed")
			case listed["greek60"] && listed["math40"]:
				f.Verdict = "counts-code-points"
				f.Notes = append(f.Notes, "the 72-character ASCII name was rejected while both non-ASCII names under 64 code points were listed: the platform counts code points, the reference validator's unit")
			case listed["greek60"] && !listed["math40"]:
				f.Verdict = "counts-utf16-units"
				f.Notes = append(f.Notes, "the Greek name (60 in every unit but bytes) was listed and the astral name (74 UTF-16 units) was not: the platform counts UTF-16 code units")
			case !listed["greek60"] && !listed["math40"]:
				f.Verdict = "counts-bytes"
				f.Notes = append(f.Notes, "the short Greek name was listed but both longer non-ASCII names (114 and 142 bytes) were not: the platform counts UTF-8 bytes")
			default:
				f.Verdict = "inconsistent-length-unit"
				f.Notes = append(f.Notes, "the outcomes fit no single counting unit; see the per-name fates")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(so))
			return f
		},
	}
}
