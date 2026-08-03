// Package checks defines the automated platform-behavior checks and their
// verdict logic. Check IDs match ../checks.md (check list
// version 0.2) and canary phrases match ../benchmark-skills/README.md's
// canary index. The invocation and observation machinery lives in
// skillxp; this package owns only what makes these observations a
// benchmark: which skills, which prompts, and how facts map to verdicts.
package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agent-ecosystem/agentminutes/session"
	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
	"github.com/agent-ecosystem/skillxp/trace"
)

// ChecklistVersion is the checks.md check list version these
// specs implement; reports stamp it so readers know which checks existed
// when a platform was tested.
const ChecklistVersion = "0.2"

// Statuses, matching the results template vocabulary.
const (
	StatusObserved     = "observed"
	StatusInconclusive = "inconclusive"
	StatusError        = "error"
)

// Loading vehicles: how skill content reached the model. "Loaded" is not a
// boolean; it has a mechanism, and the mechanism drives author-facing
// differences (frontmatter visibility, wrapping).
const (
	VehicleHarnessPush = "harness-push" // harness injected the content
	VehicleModelPull   = "model-pull"   // model fetched it with a tool
	VehicleNone        = "none"
)

// Confidence grades how directly the evidence supports the verdict.
const (
	ConfidenceDirect   = "transcript-direct"
	ConfidenceInferred = "behavioral-inference"
)

// Evidence is one citable observation supporting a finding.
type Evidence struct {
	Note       string `json:"note"`
	EventIndex int    `json:"event_index,omitempty"`
	Line       int    `json:"line,omitempty"`
	// Session is the 1-based session index for multi-session checks.
	Session int `json:"session,omitempty"`
}

// Finding is the outcome of one check run on one harness.
type Finding struct {
	CheckID        string     `json:"check_id"`
	Harness        string     `json:"harness"`
	HarnessVersion string     `json:"harness_version,omitempty"`
	Model          string     `json:"model,omitempty"`
	SessionID      string     `json:"session_id,omitempty"`
	TranscriptPath string     `json:"transcript_path,omitempty"`
	PromptUsed     string     `json:"prompt_used,omitempty"`
	Status         string     `json:"status"`
	Verdict        string     `json:"verdict"`
	Vehicle        string     `json:"vehicle,omitempty"`
	Confidence     string     `json:"confidence,omitempty"`
	Evidence       []Evidence `json:"evidence,omitempty"`
	Notes          []string   `json:"notes,omitempty"`
	Started        time.Time  `json:"started,omitzero"`
	Ended          time.Time  `json:"ended,omitzero"`

	// Runs and VerdictCounts appear on repeated checks (-runs > 1):
	// Verdict then holds the modal verdict, VerdictCounts the full
	// distribution, and RunErrors any repetitions that failed to execute.
	// Verdict variation across runs proves model-level behavior;
	// consistency only suggests platform-level behavior.
	Runs          int            `json:"runs,omitempty"`
	VerdictCounts map[string]int `json:"verdict_counts,omitempty"`
	RunErrors     []string       `json:"run_errors,omitempty"`
}

// Turn is one prompt in a check session.
type Turn struct {
	// Prompt builds the turn's prompt. Prompts must never contain a live
	// canary phrase: a literal-minded model truthfully answers "it's in
	// your message" (observed on antigravity), and the phrase contaminates
	// every echo location in the transcript.
	Prompt func(p profile.Profile) string

	// Activation marks the turn as a skill-activation turn.
	Activation bool

	// Before, when set, runs just before the turn's invocation — the hook
	// for editing skill files between activations.
	Before func(p profile.Profile, projectDir string) error
}

// Session is one independent harness session in a check. Most checks need
// exactly one; cross-scope-dependency runs a with-dependency and a
// without-dependency session.
type Session struct {
	// Skills are benchmark skill directory names (under benchmark-skills/)
	// to install at project scope.
	Skills []string

	// UserSkills are benchmark skill directory names to install at USER
	// scope. Requires the sandbox.
	UserSkills []string

	// ProjectDirs are benchmark fixture directory names copied into the
	// fixture's project ROOT rather than the skills directory — for
	// non-skill files like a stray SKILL.md outside any skills root.
	ProjectDirs []string

	// OverlayDirs are benchmark fixture directory names whose CONTENTS
	// are copied onto the project root, preserving internal layout — for
	// fixtures that must land at exact paths (e.g. .agents/skills/...).
	OverlayDirs []string

	// Turns run in order against this session.
	Turns []Turn
}

// Spec is one automatable check.
type Spec struct {
	// ID matches the check ID in checks.md.
	ID          string
	Description string

	// Sessions run in order, each independent (own fixture, own harness
	// session).
	Sessions []Session

	// RequiresSandbox marks checks that install at user scope or would
	// otherwise touch real user state; the runner refuses them
	// unsandboxed.
	RequiresSandbox bool

	// Evaluate maps the completed sessions (index-aligned with Sessions)
	// to a finding. The runner fills identity fields afterward.
	Evaluate func(sos []*observe.SessionObservation) Finding
}

// Registry returns the implemented checks, in benchmark order.
func Registry() []Spec {
	return []Spec{
		discoveryReadingDepth(),
		activationLoadingScope(),
		eagerLinkResolution(),
		recognizedDirectorySet(),
		directoryNamingDivergence(),
		unrecognizedDirectoryHandling(),
		resourceEnumerationBehavior(),
		pathResolutionBase(),
		crossSkillResourceShadowing(),
		pathTraversalBoundary(),
		resourceNestingDepth(),
		bundledScriptExecution(),
		discoveryListingFields(),
		frontmatterHandling(),
		contentWrappingFormat(),
		reactivationDeduplication(),
		reactivationFreshness(),
		compatibilityFieldBehavior(),
		allowedToolsBehavior(),
		crossSkillInvocation(),
		invocationDepthLimit(),
		circularInvocationHandling(),
		invocationLanguageSensitivity(),
		informalDependencyResolution(),
		missingDependencyBehavior(),
		nonstandardDependencyFields(),
		crossScopeDependency(),
		crossClientDirectoryInterop(),
		recursiveRootDiscovery(),
		nestedSkillDiscovery(),
		nameCollisionPrecedence(),
		malformedYamlTolerance(),
		missingDescriptionHandling(),
		invalidNameTolerance(),
		nameDirectoryMismatch(),
		metadataValueEdgeCases(),
		oversizeDescriptionHandling(),
		oversizeCompatibilityHandling(),
	}
}

// For returns the spec with the given ID.
func For(id string) (Spec, error) {
	for _, s := range Registry() {
		if s.ID == id {
			return s, nil
		}
	}
	return Spec{}, fmt.Errorf("checks: unknown or unimplemented check %q", id)
}

// Canary phrases, from the benchmark's canary index.
const (
	probeLoadingBodyCanary    = "CARDINAL-ZEBRA-7742"
	probeCrossScopeBodyCanary = "CRANE-STEEL-1163"
)

// editedBodyCanary replaces the body canary for the freshness check. It is
// deliberately absent from the benchmark's canary index: it exists only in
// a fixture edited mid-session.
const editedBodyCanary = "LOON-BASALT-4242"

var probeLoadingResourceCanaries = map[string]string{
	"references/api-overview.md":        "PELICAN-MANGO-3391",
	"references/error-codes.md":         "FALCON-QUARTZ-8819",
	"references/unreferenced-detail.md": "OSPREY-COBALT-5567",
	"scripts/check-status.sh":           "HERON-AMBER-2204",
	"assets/config-template.yaml":       "CRANE-TOPAZ-6638",
}

func activatePrompt(p profile.Profile) string {
	return p.ActivationPrompt("probe-loading")
}

func reactivatePrompt(profile.Profile) string {
	return "Activate the probe-loading skill again and follow its instructions a second time."
}

// loadsOf reports how a phrase reached the model in one session: harness
// injections and tool-result pulls.
func loadsOf(so *observe.SessionObservation, phrase string) (injected, pulled []trace.Occurrence) {
	sess := so.Final().Session
	occs := trace.Phrase(sess, phrase, so.Final().Profile.EchoSubtypes)
	return trace.At(occs, trace.LocHarnessInjected), trace.At(occs, trace.LocToolResult)
}

// bodyLoads reports how the probe-loading body reached the model across a
// session, plus the model's skill invocations/reads referencing it.
func bodyLoads(so *observe.SessionObservation) (injected, pulled []trace.Occurrence, attempts []int) {
	injected, pulled = loadsOf(so, probeLoadingBodyCanary)
	return injected, pulled, trace.ToolReadsOf(so.Final().Session, "probe-loading")
}

// finalAnswer returns the last assistant message's text, flattened to one
// line for use in notes. The cap exists only to bound pathological
// ramblers; typical structured answers fit well inside it — 240 proved
// too small and truncated real content.
func finalAnswer(so *observe.SessionObservation) string {
	events := so.Final().Session.Events
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Kind == session.KindAssistantMessage {
			return flattenAnswer(events[i].AssistantMessage.Text())
		}
	}
	return ""
}

// flattenAnswer applies finalAnswer's formatting to raw answer text,
// truncating on rune boundaries (answers can be non-ASCII).
func flattenAnswer(text string) string {
	if r := []rune(text); len(r) > 1200 {
		text = string(r[:1200]) + "…"
	}
	return strings.ReplaceAll(text, "\n", " ")
}

// errorResultsFor reports whether any tool result correlated with the
// given tool-call event indices is an error.
func errorResultsFor(so *observe.SessionObservation, callIdxs []int) bool {
	events := so.Final().Session.Events
	ids := map[string]bool{}
	for _, i := range callIdxs {
		ids[events[i].ToolCall.ToolCallID] = true
	}
	for i := range events {
		if events[i].Kind == session.KindToolResult && ids[events[i].ToolResult.ToolCallID] && events[i].ToolResult.IsError {
			return true
		}
	}
	return false
}

// vehicleOf labels how content arrived given its injected and pulled
// occurrences.
func vehicleOf(injected, pulled []trace.Occurrence) string {
	switch {
	case len(injected) > 0:
		return VehicleHarnessPush
	case len(pulled) > 0:
		return VehicleModelPull
	}
	return VehicleNone
}

// loadsInOrder merges injected and pulled occurrences in event order, for
// citing "the first load" or "the second load" regardless of vehicle.
func loadsInOrder(injected, pulled []trace.Occurrence) []trace.Occurrence {
	out := append(append([]trace.Occurrence{}, injected...), pulled...)
	sort.Slice(out, func(i, j int) bool { return out[i].EventIndex < out[j].EventIndex })
	return out
}

// negConfidence grades a verdict that rests on the ABSENCE of injected
// content: transcript-direct where the harness records injections,
// behavioral inference where it does not (antigravity).
func negConfidence(so *observe.SessionObservation) string {
	if so.Final().Profile.RecordsInjectedContext {
		return ConfidenceDirect
	}
	return ConfidenceInferred
}

// noteFileReadWorkaround flags a chained skill whose content arrived only
// by file read on a harness whose activation mechanism injects: the model
// may be working around a platform invocation limit, so the load proves
// content reachability, not platform activation support.
func noteFileReadWorkaround(f *Finding, entryInjected bool, inj, pull []trace.Occurrence, name string) {
	if entryInjected && len(inj) == 0 && len(pull) > 0 {
		f.Notes = append(f.Notes, name+" arrived only via the model's own file read even though this harness injects on activation (likely a model workaround); treat as content reachability rather than platform activation support")
	}
}

// turnOf returns the 0-based turn index whose invocation produced event
// idx: turn observations accumulate, so the first turn whose cumulative
// event count exceeds idx owns it.
func turnOf(so *observe.SessionObservation, idx int) int {
	for i, t := range so.Turns {
		if idx < len(t.Session.Events) {
			return i
		}
	}
	return len(so.Turns) - 1
}

// eventVisibleText returns the model-visible text of one event, mirroring
// trace's extraction, for structural inspection around a traced phrase.
func eventVisibleText(so *observe.SessionObservation, idx int) string {
	ev := so.Final().Session.Events[idx]
	switch ev.Kind {
	case session.KindUserMessage:
		return ev.UserMessage.Text()
	case session.KindSystem:
		return ev.System.Text
	case session.KindToolResult:
		return ev.ToolResult.Text()
	case session.KindAssistantMessage:
		return ev.AssistantMessage.Text()
	}
	return ""
}

// assistantMentions returns occurrences of phrase in assistant-authored
// prose (messages and thinking), excluding tool-call inputs — the signal
// that the model surfaced something rather than silently skipping it.
func assistantMentions(so *observe.SessionObservation, phrase string) []trace.Occurrence {
	occs := trace.Phrase(so.Final().Session, phrase, so.Final().Profile.EchoSubtypes)
	var out []trace.Occurrence
	for _, o := range trace.At(occs, trace.LocModelOutput) {
		if o.Kind == session.KindAssistantMessage || o.Kind == session.KindThinking {
			out = append(out, o)
		}
	}
	return out
}

func discoveryReadingDepth() Spec {
	return Spec{
		ID:          "discovery-reading-depth",
		Description: "Does the harness read only SKILL.md metadata at discovery, or the full body?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns: []Turn{{
				Prompt: func(profile.Profile) string {
					// Canary-free by design; the pattern description lets
					// the model recite any canary it can actually see.
					return "Without using any tools, list every phrase of the form BIRD-MINERAL-NUMBER " +
						"(three hyphen-separated parts in capitals, e.g. SPARROW-GRANITE-0000) that is " +
						"visible anywhere in your context right now. Reply with one phrase per line, or NONE."
				},
			}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "discovery-reading-depth"}
			occs := trace.Phrase(obs.Session, probeLoadingBodyCanary, obs.Profile.EchoSubtypes)
			injected := trace.At(occs, trace.LocHarnessInjected)
			emitted := trace.At(occs, trace.LocModelOutput)

			if obs.Profile.RecordsInjectedContext {
				f.Confidence = ConfidenceDirect
				idx := trace.SkillListing(obs.Session, obs.Profile.SkillListingSubtypes, "probe-loading")
				if idx < 0 {
					f.Status = StatusObserved
					f.Verdict = "skill-not-discovered"
					f.Notes = append(f.Notes, "no skill listing mentioning probe-loading; install path or discovery gating needs investigation")
					return f
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], idx, "discovery listing names probe-loading"))
			} else {
				f.Confidence = ConfidenceInferred
				f.Notes = append(f.Notes, "transcript does not record injected context; discovery listing unobservable, verdict rests on the model not knowing the body canary")
			}

			switch {
			case len(injected) > 0:
				f.Status = StatusObserved
				f.Verdict = "full-body-at-discovery"
				f.Evidence = append(f.Evidence, evAt(sos[0], injected[0].EventIndex, "body canary in harness-injected content before any activation"))
			case len(emitted) > 0:
				f.Status = StatusInconclusive
				f.Verdict = "model-knows-canary-without-visible-source"
				f.Evidence = append(f.Evidence, evAt(sos[0], emitted[0].EventIndex, "model emitted the body canary but no visible source records it"))
			default:
				f.Status = StatusObserved
				f.Verdict = "metadata-only"
			}
			return f
		},
	}
}

func activationLoadingScope() Spec {
	return Spec{
		ID:          "activation-loading-scope",
		Description: "On activation, does the harness load only the SKILL.md body, or also bundled resources, and by which vehicle?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns:  []Turn{{Prompt: activatePrompt, Activation: true}},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "activation-loading-scope"}
			injected, pulled, reads := bodyLoads(sos[0])

			switch {
			case len(injected) > 0:
				f.Vehicle = VehicleHarnessPush
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence, evAt(sos[0], injected[0].EventIndex, "body canary in harness-injected content"))
			case len(pulled) > 0 && len(reads) > 0:
				f.Vehicle = VehicleModelPull
				f.Confidence = ConfidenceDirect
				f.Evidence = append(f.Evidence,
					evAt(sos[0], reads[0], "model's tool call targets the skill's own path"),
					evAt(sos[0], pulled[0].EventIndex, "body canary arrived in the tool result"))
				if !obs.Profile.RecordsInjectedContext {
					f.Notes = append(f.Notes, "direct-path navigation without a prior search implies a discovery listing the transcript does not record")
				}
			default:
				f.Status = StatusInconclusive
				f.Vehicle = VehicleNone
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "body canary never appeared via injection or tool result; the skill may not have activated")
				return f
			}

			var eager []string
			for file, canary := range probeLoadingResourceCanaries {
				rocc := trace.Phrase(obs.Session, canary, obs.Profile.EchoSubtypes)
				loaded := append(trace.At(rocc, trace.LocHarnessInjected), trace.At(rocc, trace.LocToolResult)...)
				if len(loaded) > 0 {
					eager = append(eager, file)
					f.Evidence = append(f.Evidence, evAt(sos[0], loaded[0].EventIndex, "resource canary for "+file+" present"))
				}
			}
			f.Status = StatusObserved
			if len(eager) == 0 {
				f.Verdict = "body-only"
			} else {
				f.Verdict = fmt.Sprintf("body-plus-resources:%v", eager)
			}
			return f
		},
	}
}

func reactivationDeduplication() Spec {
	return Spec{
		ID:          "reactivation-deduplication",
		Description: "When the same skill is activated twice in one session, is its content loaded again or deduplicated?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns: []Turn{
				{Prompt: activatePrompt, Activation: true},
				{Prompt: reactivatePrompt, Activation: true},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "reactivation-deduplication"}
			injected, pulled, reads := bodyLoads(sos[0])

			switch {
			case len(injected) >= 2:
				f.Status = StatusObserved
				f.Vehicle = VehicleHarnessPush
				f.Confidence = ConfidenceDirect
				f.Verdict = "reinjected-each-activation"
				f.Evidence = append(f.Evidence,
					evAt(sos[0], injected[0].EventIndex, "first injection of body canary"),
					evAt(sos[0], injected[1].EventIndex, "second injection of body canary"))
				f.Notes = append(f.Notes, "platform-level: duplicate skill content occupies context after reactivation")
			case len(injected) == 1 && len(reads) >= 2:
				f.Status = StatusObserved
				f.Vehicle = VehicleHarnessPush
				f.Confidence = ConfidenceDirect
				f.Verdict = "deduplicated"
				f.Evidence = append(f.Evidence,
					evAt(sos[0], injected[0].EventIndex, "single injection of body canary"),
					evAt(sos[0], reads[1], "second skill invocation produced no second injection"))
				f.Notes = append(f.Notes, "platform-level: second activation acknowledged without re-injecting content")
			case len(pulled) >= 2:
				f.Status = StatusObserved
				f.Vehicle = VehicleModelPull
				f.Confidence = ConfidenceDirect
				f.Verdict = "re-read-each-activation"
				f.Evidence = append(f.Evidence,
					evAt(sos[0], pulled[0].EventIndex, "first read of skill body"),
					evAt(sos[0], pulled[1].EventIndex, "second read of skill body"))
				f.Notes = append(f.Notes, "model-level: on pull-vehicle harnesses re-loading is the model's choice, not platform policy")
			case len(pulled) == 1:
				f.Status = StatusObserved
				f.Vehicle = VehicleModelPull
				f.Confidence = ConfidenceInferred
				f.Verdict = "not-re-read"
				f.Notes = append(f.Notes, "model-level: the model reused its earlier read instead of re-reading; no platform dedup mechanism is involved")
			case len(injected)+len(pulled) == 1:
				f.Status = StatusInconclusive
				f.Vehicle = VehicleNone
				f.Verdict = "second-activation-not-attempted"
				f.Notes = append(f.Notes, fmt.Sprintf("body loads: %d, skill invocations/reads: %d; the model likely answered turn 2 from conversation memory, so platform dedup was not exercised", len(injected)+len(pulled), len(reads)))
			default:
				f.Status = StatusInconclusive
				f.Vehicle = VehicleNone
				f.Verdict = "activation-not-observed"
			}
			return f
		},
	}
}

func reactivationFreshness() Spec {
	return Spec{
		ID:          "reactivation-freshness",
		Description: "After SKILL.md is edited mid-session, does reactivation serve the fresh content or a cached copy?",
		Sessions: []Session{{
			Skills: []string{"probe-loading"},
			Turns: []Turn{
				{Prompt: activatePrompt, Activation: true},
				{
					Prompt:     reactivatePrompt,
					Activation: true,
					Before: func(p profile.Profile, projectDir string) error {
						path := filepath.Join(projectDir, p.ProjectSkillDir, "probe-loading", "SKILL.md")
						data, err := os.ReadFile(path)
						if err != nil {
							return err
						}
						edited := strings.ReplaceAll(string(data), probeLoadingBodyCanary, editedBodyCanary)
						return os.WriteFile(path, []byte(edited), 0o644)
					},
				},
			},
		}},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			obs := sos[0].Final()
			f := Finding{CheckID: "reactivation-freshness"}
			injected, pulled, _ := bodyLoads(sos[0])
			editedOccs := trace.Phrase(obs.Session, editedBodyCanary, obs.Profile.EchoSubtypes)
			editedLoaded := append(trace.At(editedOccs, trace.LocHarnessInjected), trace.At(editedOccs, trace.LocToolResult)...)
			originalLoaded := len(injected) + len(pulled)

			switch {
			case len(editedLoaded) > 0:
				f.Status = StatusObserved
				f.Verdict = "fresh-content-served"
				f.Confidence = ConfidenceDirect
				if len(trace.At(editedOccs, trace.LocHarnessInjected)) > 0 {
					f.Vehicle = VehicleHarnessPush
				} else {
					f.Vehicle = VehicleModelPull
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], editedLoaded[0].EventIndex, "post-edit canary reached the model on reactivation"))
			case originalLoaded >= 2:
				f.Status = StatusObserved
				f.Verdict = "stale-content-served"
				f.Confidence = ConfidenceDirect
				last := pulled
				if len(injected) >= 2 {
					f.Vehicle = VehicleHarnessPush
					last = injected
				} else {
					f.Vehicle = VehicleModelPull
				}
				f.Evidence = append(f.Evidence, evAt(sos[0], last[len(last)-1].EventIndex, "reactivation delivered the pre-edit canary despite the on-disk edit"))
			default:
				f.Status = StatusInconclusive
				f.Verdict = "reload-not-observed"
				f.Notes = append(f.Notes, "the second activation never re-delivered skill content (deduplication or model memory); freshness is unobservable in this session, so read this alongside the reactivation-deduplication finding")
			}
			return f
		},
	}
}

func crossScopeDependency() Spec {
	activateCross := func(p profile.Profile) string {
		return p.ActivationPrompt("probe-cross-scope")
	}
	return Spec{
		ID:              "cross-scope-dependency",
		Description:     "Can a project-level skill invoke a dependency that exists only at user level, and what is the failure mode when it is absent?",
		RequiresSandbox: true,
		Sessions: []Session{
			{
				Skills:     []string{"probe-cross-scope"},
				UserSkills: []string{"probe-loading"},
				Turns:      []Turn{{Prompt: activateCross, Activation: true}},
			},
			{
				Skills: []string{"probe-cross-scope"},
				Turns:  []Turn{{Prompt: activateCross, Activation: true}},
			},
		},
		Evaluate: func(sos []*observe.SessionObservation) Finding {
			f := Finding{CheckID: "cross-scope-dependency"}
			present, absent := sos[0], sos[1]

			// Session 1: dependency installed at user scope.
			crossInj, crossPull := loadsOf(present, probeCrossScopeBodyCanary)
			if len(crossInj)+len(crossPull) == 0 {
				f.Status = StatusInconclusive
				f.Verdict = "activation-not-observed"
				f.Notes = append(f.Notes, "probe-cross-scope itself never loaded in the with-dependency session")
				return f
			}
			depInj, depPull, _ := bodyLoads(present)
			attempts := trace.ToolReadsOf(present.Final().Session, "probe-loading")

			var resolved string
			switch {
			case len(depInj) > 0:
				resolved = "resolved-across-scopes"
				f.Vehicle = VehicleHarnessPush
				f.Evidence = append(f.Evidence, evAt(present, depInj[0].EventIndex, "user-scope dependency's body canary injected during project-skill session"))
			case len(depPull) > 0:
				resolved = "resolved-across-scopes"
				f.Vehicle = VehicleModelPull
				f.Evidence = append(f.Evidence, evAt(present, depPull[0].EventIndex, "user-scope dependency's body canary pulled during project-skill session"))
			case len(attempts) > 0:
				resolved = "attempted-not-resolved"
				f.Evidence = append(f.Evidence, evAt(present, attempts[0], "model attempted the dependency but its content never arrived"))
			default:
				resolved = "not-attempted"
				f.Notes = append(f.Notes, "model never invoked or read the user-scope dependency despite instructions (model-level)")
			}
			f.Notes = append(f.Notes, "with-dependency final answer: "+finalAnswer(present))

			// Session 2: dependency absent everywhere.
			absAttempts := trace.ToolReadsOf(absent.Final().Session, "probe-loading")
			absInj, absPull, _ := bodyLoads(absent)
			var missing string
			switch {
			case len(absInj)+len(absPull) > 0:
				// The canary is not installed anywhere; loading it would
				// mean contamination or a phantom source.
				missing = "phantom-content"
				f.Status = StatusInconclusive
				f.Notes = append(f.Notes, "dependency content appeared despite not being installed; investigate contamination")
			case len(absAttempts) > 0 && errorResultsFor(absent, absAttempts):
				missing = "visible-failure"
				f.Evidence = append(f.Evidence, evSession(absent, 2, absAttempts[0], "attempt on the missing dependency returned an error result"))
			case len(absAttempts) > 0:
				missing = "attempted-no-error"
				f.Evidence = append(f.Evidence, evSession(absent, 2, absAttempts[0], "model attempted the missing dependency; no error result recorded"))
			case len(assistantMentions(absent, "probe-loading")) > 0:
				// Better than a silent skip: the model consulted its listing
				// and surfaced the absence without attempting a call.
				missing = "reported-without-attempt"
				f.Evidence = append(f.Evidence, evSession(absent, 2, assistantMentions(absent, "probe-loading")[0].EventIndex, "model surfaced the missing dependency in its own text without attempting it"))
			default:
				missing = "silent-skip"
			}
			if missing == "reported-without-attempt" || missing == "attempted-no-error" {
				f.Notes = append(f.Notes, "which missing-dependency tier appears (attempted vs reported) is the model's choice on pull harnesses and can vary between runs")
			}
			f.Notes = append(f.Notes, "without-dependency final answer: "+finalAnswer(absent))

			f.Status = StatusObserved
			if missing == "phantom-content" {
				f.Status = StatusInconclusive
			}
			f.Confidence = ConfidenceDirect
			f.Verdict = resolved + "; missing:" + missing
			return f
		},
	}
}

// evAt builds an Evidence entry citing an event in a session's final
// transcript.
func evAt(so *observe.SessionObservation, idx int, note string) Evidence {
	return evSession(so, 0, idx, note)
}

// evSession is evAt with an explicit 1-based session number for
// multi-session checks (0 omits the field).
func evSession(so *observe.SessionObservation, sessionNum, idx int, note string) Evidence {
	e := Evidence{Note: note, EventIndex: idx, Session: sessionNum}
	if p := so.Final().Session.Events[idx].Provenance; p != nil {
		e.Line = p.Line
	}
	return e
}
