// Directory-recognition and structural edge-case checks: the three spec
// directories, nonstandard directory names, nested skills, and deep
// resource nesting.
package checks

import (
	"fmt"

	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

func recognizedDirectorySet() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-loading") }
	// Filename-only markers, one per spec directory. The two body-linked
	// reference files are useless here (their names appear in the injected
	// body); only names the body never mentions can prove enumeration.
	dirs := []struct{ dir, marker string }{
		{"scripts/", "check-status.sh"},
		{"references/", "unreferenced-detail"},
		{"assets/", "config-template.yaml"},
	}
	return Spec{
		ID:          "recognized-directory-set",
		Description: "Are the three spec directories (scripts/, references/, assets/) enumerated to the model at activation?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "recognized-directory-set"}
			bInj, bPull := loadsOf(sos[0], probeLoadingBodyCanary)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			var enumerated, explored []string
			for _, d := range dirs {
				inj, pull := loadsOf(sos[0], d.marker)
				switch {
				case len(inj) > 0:
					enumerated = append(enumerated, d.dir)
					f.Evidence = append(f.Evidence, evAt(sos[0], inj[0].EventIndex, d.dir+" file name injected at activation"))
				case len(pull) > 0:
					explored = append(explored, d.dir)
					f.Evidence = append(f.Evidence, evAt(sos[0], pull[0].EventIndex, d.dir+" file name arrived only via the model's own exploration"))
				}
			}
			f.Status = StatusObserved
			switch {
			case len(enumerated) == len(dirs):
				f.Verdict = "all-three-dirs-enumerated"
				f.Confidence = ConfidenceDirect
			case len(enumerated) > 0:
				f.Verdict = fmt.Sprintf("partial-enumeration:%v", enumerated)
				f.Confidence = ConfidenceDirect
			default:
				f.Verdict = "no-enumeration-at-activation"
				f.Confidence = negConfidence(sos[0])
				if len(explored) > 0 {
					f.Notes = append(f.Notes, fmt.Sprintf("model explored %v itself; the platform enumerated nothing", explored))
				}
			}
			return f
		},
	}
}

// nonstandardDirsBodyMarker detects probe-nonstandard-dirs activation; the
// skill has no body canary.
const nonstandardDirsBodyMarker = "plausible alternative name"

func directoryNamingDivergence() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-nonstandard-dirs") }
	return Spec{
		ID:          "directory-naming-divergence",
		Description: "Is a resources/ directory (alternative to spec's references/) loaded, enumerated, readable, or invisible?",
		Sessions: []Session{{
			Skills: []string{"probe-nonstandard-dirs"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "directory-naming-divergence"}
			bInj, bPull := loadsOf(sos[0], nonstandardDirsBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			cInj, cPull := loadsOf(sos[0], "SWIFT-OPAL-8156")
			nInj, _ := loadsOf(sos[0], "api-reference.md")
			f.Status = StatusObserved
			switch {
			case len(cInj) > 0:
				f.Verdict = "resources-content-injected"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], cInj[0].EventIndex, "resources/ file content injected at activation"))
			case len(nInj) > 0:
				f.Verdict = "resources-enumerated-not-loaded"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], nInj[0].EventIndex, "resources/ file name injected without its content"))
			case len(cPull) > 0:
				f.Verdict = "resources-readable-on-demand"
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], cPull[0].EventIndex, "resources/ file content arrived via the model's own read"))
			default:
				f.Verdict = "resources-untouched"
				f.Confidence = negConfidence(sos[0])
			}
			f.Notes = append(f.Notes, "read alongside resource-enumeration-behavior: equal treatment of resources/ and references/ (both enumerated, or both untouched) means no naming divergence on this platform")
			return f
		},
	}
}

func unrecognizedDirectoryHandling() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-nonstandard-dirs") }
	dirs := []struct{ dir, canary string }{
		{"evals/", "ROBIN-JADE-3847"},
		{"templates/", "WREN-PEARL-6293"},
	}
	return Spec{
		ID:          "unrecognized-directory-handling",
		Description: "What happens to directories the spec never named (evals/, templates/): injected, readable on demand, or invisible?",
		Sessions: []Session{{
			Skills: []string{"probe-nonstandard-dirs"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "unrecognized-directory-handling"}
			bInj, bPull := loadsOf(sos[0], nonstandardDirsBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			var injected, readable []string
			for _, d := range dirs {
				inj, pull := loadsOf(sos[0], d.canary)
				switch {
				case len(inj) > 0:
					injected = append(injected, d.dir)
					f.Evidence = append(f.Evidence, evAt(sos[0], inj[0].EventIndex, d.dir+" content injected at activation"))
				case len(pull) > 0:
					readable = append(readable, d.dir)
					f.Evidence = append(f.Evidence, evAt(sos[0], pull[0].EventIndex, d.dir+" content arrived via the model's own read"))
				}
			}
			f.Status = StatusObserved
			switch {
			case len(injected) > 0:
				f.Verdict = fmt.Sprintf("injected-at-activation:%v", injected)
				f.Confidence = ConfidenceDirect
			case len(readable) > 0:
				f.Verdict = fmt.Sprintf("readable-on-demand:%v", readable)
				f.Confidence = ConfidenceDirect
			default:
				f.Verdict = "untouched"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "no nonstandard directory's content ever reached the model; whether that is 'ignored by platform' or 'model chose not to look' is model-level on pull harnesses")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

func nestedSkillDiscovery() Spec {
	return Spec{
		ID:          "nested-skill-discovery",
		Description: "Is a SKILL.md nested inside another skill's references/ tree discovered as a separate skill?",
		Sessions: []Session{{
			Skills: []string{"probe-deep-nesting"},
			// Passive listing turn: the only way "nested-skill"
			// (hyphenated name) enters the transcript is the platform's
			// own discovery listing.
			Turns: []Turn{{Prompt: listingPrompt}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "nested-skill-discovery"}
			if obs.Profile.RecordsInjectedContext {
				outer := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-deep-nesting")
				nested := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "nested-skill")
				switch {
				case outer < 0:
					f.Status = StatusInconclusive
					f.Verdict = "outer-skill-not-discovered"
				case nested >= 0:
					f.Status = StatusObserved
					f.Verdict = "nested-skill-discovered"
					f.Confidence = ConfidenceDirect
					f.Evidence = append(f.Evidence, evAt(sos[0], nested, "discovery listing names nested-skill (found inside probe-deep-nesting/references/)"))
				default:
					f.Status = StatusObserved
					f.Verdict = "nested-skill-not-discovered"
					f.Confidence = ConfidenceDirect
					f.Evidence = append(f.Evidence, evAt(sos[0], outer, "discovery listing names only the outer skill"))
				}
				return f
			}
			outerM := assistantMentions(sos[0], "probe-deep-nesting")
			nestedM := assistantMentions(sos[0], "nested-skill")
			switch {
			case len(outerM) == 0:
				f.Status = StatusInconclusive
				f.Verdict = "outer-skill-not-discovered"
				f.Notes = append(f.Notes, "model's listing answer never named the outer skill either; discovery unobservable")
			case len(nestedM) > 0:
				f.Status = StatusObserved
				f.Verdict = "nested-skill-discovered"
				f.Confidence = ConfidenceInferred
				f.Evidence = append(f.Evidence, evAt(sos[0], nestedM[0].EventIndex, "model's tool-free listing answer names nested-skill"))
			default:
				f.Status = StatusObserved
				f.Verdict = "nested-skill-not-discovered"
				f.Confidence = ConfidenceInferred
				f.Evidence = append(f.Evidence, evAt(sos[0], outerM[0].EventIndex, "model's tool-free listing answer names only the outer skill"))
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// listingPrompt is the shared passive first turn: tools are off on
// non-activation turns, so the only way a skill name can enter the
// transcript is the platform's own discovery listing (or, on harnesses
// that record no injected context, the model's tool-free answer).
func listingPrompt(profile.Profile) string {
	return "Without using any tools, list the name of every skill in your Available skills list, one per line. If the list is empty, reply NONE."
}

const mismatchBodyCanary = "SWAN-BERYL-3324"

func nameDirectoryMismatch() Spec {
	return Spec{
		ID:          "name-directory-mismatch",
		Description: "When directory name and frontmatter name disagree, which identity is the skill listed and invocable under?",
		Sessions: []Session{{
			Skills: []string{"probe-mismatch-dir"},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt("probe-name-mismatch") }, Activation: true},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt("probe-mismatch-dir") }, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "name-directory-mismatch"}

			// Catalog identity is the check's question; which activation
			// prompts happen to work is secondary (a model can satisfy
			// either prompt by inference or file access).
			nameListed, dirListed := false, false
			if obs.Profile.RecordsInjectedContext {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-name-mismatch"); idx >= 0 {
					nameListed = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing carries the FRONTMATTER name probe-name-mismatch"))
				}
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-mismatch-dir"); idx >= 0 {
					dirListed = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing carries the DIRECTORY name probe-mismatch-dir"))
				}
				f.Confidence = ConfidenceDirect
			} else {
				// Tool-free listing turn: the model can only echo its
				// system-prompt catalog.
				for _, o := range assistantMentions(sos[0], "probe-name-mismatch") {
					if turnOf(sos[0], o.EventIndex) == 0 {
						nameListed = true
						f.Evidence = append(f.Evidence, evAt(sos[0], o.EventIndex, "tool-free listing answer names probe-name-mismatch"))
						break
					}
				}
				for _, o := range assistantMentions(sos[0], "probe-mismatch-dir") {
					if turnOf(sos[0], o.EventIndex) == 0 {
						dirListed = true
						f.Evidence = append(f.Evidence, evAt(sos[0], o.EventIndex, "tool-free listing answer names probe-mismatch-dir"))
						break
					}
				}
				f.Confidence = ConfidenceInferred
			}

			// Invocability, as supporting notes: canary loads in the
			// frontmatter-name turn (2) and directory-name turn (3).
			inj, pull := loadsOf(sos[0], mismatchBodyCanary)
			var fmWorked, dirWorked bool
			for _, o := range loadsInOrder(inj, pull) {
				switch turnOf(sos[0], o.EventIndex) {
				case 1:
					fmWorked = true
				case 2:
					dirWorked = true
				}
			}
			f.Notes = append(f.Notes, fmt.Sprintf("activation loads: by frontmatter name=%v, by directory name=%v (a load proves reachability, not catalog identity; the model may map either prompt to the installed skill or read the file directly)", fmWorked, dirWorked))

			f.Status = StatusObserved
			f.Vehicle = vehicleOf(inj, pull)
			switch {
			case nameListed && dirListed:
				f.Verdict = "listed-under-both"
				if obs.Profile.RecordsInjectedContext {
					f.Notes = append(f.Notes, "caveat: listings that carry file paths always contain the directory name; frontmatter-name presence is the load-bearing signal")
				}
			case nameListed:
				f.Verdict = "frontmatter-name-identity"
			case dirListed:
				f.Verdict = "directory-name-identity"
			case fmWorked || dirWorked:
				f.Verdict = "unlisted-but-reachable"
				f.Notes = append(f.Notes, "no catalog identity observed, yet the body loaded, likely via model file access or an unrecorded listing")
			default:
				f.Verdict = "not-discovered"
				f.Confidence = negConfidence(sos[0])
				f.Notes = append(f.Notes, "the platform may have rejected the skill over the name/directory mismatch, a validator-level error some platforms enforce")
			}
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

const (
	groupedBodyCanary = "CROW-AGATE-6105"
	strayBodyCanary   = "MERLIN-GYPSUM-8852"
)

func recursiveRootDiscovery() Spec {
	return Spec{
		ID:          "recursive-root-discovery",
		Description: "Does the skills root get scanned recursively (a skill under a grouping directory), and is a SKILL.md outside any root discovered?",
		Sessions: []Session{{
			Skills:      []string{"probe-group"},
			ProjectDirs: []string{"probe-stray"},
			Turns: []Turn{
				{Prompt: listingPrompt},
				{Prompt: func(p profile.Profile) string { return p.ActivationPrompt("probe-grouped") }, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "recursive-root-discovery"}

			// Catalog discovery is the check's question. A canary load on
			// the activation turn proves only reachability: a pull-vehicle
			// model asked to "activate probe-grouped" can find the file by
			// exploring the tree even when the catalog never listed it
			// (observed on antigravity: its answer said the catalog held
			// only builtins, then it went looking).
			grouped := false
			if obs.Profile.RecordsInjectedContext {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-grouped"); idx >= 0 {
					grouped = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names probe-grouped from one level below the root"))
				}
			} else {
				for _, o := range assistantMentions(sos[0], "probe-grouped") {
					if turnOf(sos[0], o.EventIndex) == 0 {
						grouped = true
						f.Evidence = append(f.Evidence, evAt(sos[0], o.EventIndex, "tool-free listing answer names probe-grouped"))
						break
					}
				}
			}
			gInj, gPull := loadsOf(sos[0], groupedBodyCanary)
			if len(gInj)+len(gPull) > 0 {
				first := loadsInOrder(gInj, gPull)[0]
				f.Vehicle = vehicleOf(gInj, gPull)
				if grouped {
					f.Evidence = append(f.Evidence, evAt(sos[0], first.EventIndex, "grouped skill's body loaded on activation"))
				} else {
					f.Notes = append(f.Notes, "the grouped skill's body still loaded on the activation turn; the model reached it by file access despite the catalog not listing it")
					f.Evidence = append(f.Evidence, evAt(sos[0], first.EventIndex, "body loaded via file access, not catalog activation"))
				}
			}

			// The stray skill is graded ONLY on discovery signals from the
			// tool-free listing turn: any canary arrival in later turns
			// would be the model reading a project file, which is file
			// access, not discovery.
			stray := false
			if obs.Profile.RecordsInjectedContext {
				if idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-stray"); idx >= 0 {
					stray = true
					f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names probe-stray despite it living outside the skills root"))
				}
			} else {
				for _, o := range assistantMentions(sos[0], "probe-stray") {
					if turnOf(sos[0], o.EventIndex) == 0 {
						stray = true
						f.Evidence = append(f.Evidence, evAt(sos[0], o.EventIndex, "model's tool-free listing answer names probe-stray"))
						break
					}
				}
			}

			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			if !obs.Profile.RecordsInjectedContext {
				// Both discovery signals rest on the model's tool-free
				// echo of its catalog.
				f.Confidence = ConfidenceInferred
			}
			g := "direct-children-only"
			if grouped {
				g = "recursive-scan"
			}
			s := "; stray:not-discovered"
			if stray {
				s = "; stray:discovered"
			}
			f.Verdict = g + s
			f.Notes = append(f.Notes, "final answer: "+finalAnswer(sos[0]))
			return f
		},
	}
}

// deepNestingBodyMarker detects probe-deep-nesting activation; the skill
// has no body canary.
const deepNestingBodyMarker = "resources nested more than one level"

func resourceNestingDepth() Spec {
	activate := func(p profile.Profile) string { return p.ActivationPrompt("probe-deep-nesting") }
	files := []struct {
		depth  int
		file   string
		canary string
	}{
		{1, "references/overview.md", "DOVE-GARNET-1029"},
		{2, "references/api/endpoints.md", "LARK-RUBY-4483"},
		{3, "references/api/v2/migration-guide.md", "OWL-EMERALD-7756"},
		{3, "references/guides/advanced/performance-tuning.md", "FINCH-SAPPHIRE-2098"},
		{5, "references/api/v2/history/deprecated/removed-endpoints.md", "PLOVER-JASPER-5590"},
	}
	const maxDepth = 5
	return Spec{
		ID:          "resource-nesting-depth",
		Description: "How deep in the directory tree do reference files stay reachable? Rungs at one, two, three, and five levels.",
		Sessions: []Session{{
			Skills: []string{"probe-deep-nesting"},
			Turns:  []Turn{{Prompt: activate, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "resource-nesting-depth"}
			bInj, bPull := loadsOf(sos[0], deepNestingBodyMarker)
			if len(bInj)+len(bPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				return f
			}
			f.Vehicle = vehicleOf(bInj, bPull)
			deepest := 0
			var missing []string
			for _, fl := range files {
				inj, pull := loadsOf(sos[0], fl.canary)
				arr := loadsInOrder(inj, pull)
				if len(arr) == 0 {
					missing = append(missing, fl.file)
					continue
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], arr[0].EventIndex, fmt.Sprintf("depth-%d file %s content arrived", fl.depth, fl.file)))
				if fl.depth > deepest {
					deepest = fl.depth
				}
			}
			f.Status = StatusObserved
			f.Confidence = ConfidenceDirect
			switch {
			case deepest == maxDepth && len(missing) == 0:
				f.Verdict = "all-depths-accessible-through-5"
			case deepest == 0:
				f.Status = StatusInconclusive
				f.Verdict = "no-resources-accessed"
				f.Confidence = negConfidence(sos[0])
			default:
				f.Verdict = fmt.Sprintf("deepest-accessed-depth-%d", deepest)
				f.Notes = append(f.Notes, fmt.Sprintf("never arrived: %v", missing))
			}
			return f
		},
	}
}
