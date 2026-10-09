// Package classify assigns each session a purpose category from its first genuine prompt.
//
// rules-v1 is an ordered list of (id, category, regexp) applied to the lowercased first
// prompt; the first match wins and no match is "unknown" with no rule. The result carries a
// category and a rule id only, never prompt text.
//
// Order is the precedence: plan, review, debug questions and error text, research, refactor,
// ops, test, debug fix verbs and symptoms, implement. Consequences the order encodes:
//   - plan beats implement ("plan for adding" is plan);
//   - test nouns beat a generic "add" or "write";
//   - a symptom question ("why does ...") is debug, "how does X work/handle" is research;
//   - "look at ... for anything that would break" and "good enough to ship" are review;
//   - ops runs before test, so CI work that mentions running tests stays ops;
//   - test rules need a verb and a test noun phrase, so a test file path alone never makes
//     a session "test".
package classify

import (
	"regexp"
	"strings"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

// Method is reported as classifier.method.
const Method = "rules-v1"

// Unknown is the category of a session no rule matches.
const Unknown = "unknown"

// Taxonomy is the contract's Category enum, in schema order.
var Taxonomy = []string{"debug", "implement", "refactor", "test", "review", "research", "plan", "ops", Unknown}

// InTaxonomy reports whether c is one of the contract's categories.
func InTaxonomy(c string) bool {
	for _, t := range Taxonomy {
		if t == c {
			return true
		}
	}
	return false
}

// Rule is one entry of the ordered rule list.
type Rule struct {
	ID       string
	Category string
	Pattern  string
	re       *regexp.Regexp
}

// Result is a classification: a category and the id of the rule that fired ("" when none did).
type Result struct {
	Category string
	Rule     string
}

func rule(id, category, pattern string) Rule {
	return Rule{ID: id, Category: category, Pattern: pattern, re: regexp.MustCompile(pattern)}
}

var rules = []Rule{
	// plan: design or plan work before doing it. Ahead of implement.
	rule("plan.before-code", "plan", `\bbefore (you )?(writing|write|making|changing|touching|implementing) (any )?(code|changes)\b`),
	rule("plan.write-plan", "plan", `\b(give me|write|draft|propose|outline|sketch|prepare|come up with) (me )?(a |an )?([a-z-]+ )?(plan|design|spec|specification|proposal|rfc|roadmap)\b`),
	rule("plan.lead", "plan", `^(plan|design)\b`),

	// review: assess existing code or a change for problems or readiness.
	rule("review.ship-ready", "review", `\b(good enough|ready|safe) to (ship|merge|release|deploy)\b`),
	rule("review.find-breakage", "review", `\b(look at|look over|go over|go through|check|audit|inspect)\b.*\bfor (anything|problems|issues|bugs|risks)\b|\banything that (would|could|might) break\b`),
	rule("review.review", "review", `^(please )?(code[- ]?)?review\b|\breview (the|this|my|our) |\bwhat would you block on\b`),

	// debug, part one: symptom questions and pasted error text.
	rule("debug.why-does", "debug", `^(why (does|do|is|are|did|would|won't|doesn't|isn't|can't)|how come)\b`),
	rule("debug.error-message", "debug", `\b[a-z]*(error|exception)\s*:|\btraceback\b|\bpanic:|\bstack ?trace\b|\bsegfault\b|\bsegmentation fault\b`),

	// research: understand how something works or compare options, no change asked for.
	rule("research.how-does", "research", `\bhow does\b[^?]*\b(work|works|handle|handles|implement|implements|manage|manages|resolve|resolves|decide|decides|store|stores|load|loads)\b`),
	rule("research.explain", "research", `^(explain|describe|walk me through|tell me about|what is|what are)\b|\bexplain how\b|\bpoint me (at|to)\b`),
	rule("research.compare", "research", `\bcompare\b|\bpros and cons\b|\btrade-?offs?\b|\bdifference between\b|\bshould (i|we) use\b|\b(versus|vs)\b`),

	// refactor: restructure existing code without changing its behaviour.
	rule("refactor.refactor", "refactor", `\brefactor`),
	rule("refactor.behaviour-preserving", "refactor", `\bno behaviou?r(al)? changes?\b|\bwithout changing (the |its |any )?behaviou?r\b|\bremove (the )?duplication\b|\bdeduplicate\b`),
	rule("refactor.restructure", "refactor", `^(extract|rename|restructure|reorgani[sz]e|split|inline|simplify|clean up|tidy up|consolidate|decouple|modulari[sz]e)\b`),

	// ops: build, deploy, CI, infrastructure, dependencies, environment or release work.
	rule("ops.ci-container", "ops", `\b(dockerfile|docker|docker-compose|github actions|gitlab ci|ci|pipeline|kubernetes|k8s|helm|terraform|ansible|nginx)\b`),
	rule("ops.deploy-release", "ops", `\b(deploy|deploys|deployment|release|rollback|roll back|provision)\b`),
	rule("ops.dependencies", "ops", `\b(upgrade|bump)\b|\bdependenc(y|ies)\b|\blockfile\b|\b(npm|pip|go mod|poetry|yarn|pnpm) (install|update|tidy|audit)\b|\bmakefile\b|\bbuild script\b`),

	// test: write, extend or repair tests or coverage. Ahead of the generic add and write.
	rule("test.coverage", "test", `\b(increase|improve|raise|extend|add|boost|expand)\b[^.;]*\bcoverage\b|\bcoverage (on|for|of)\b`),
	rule("test.write-tests", "test", `\b(write|add|create|extend|expand|generate)\b (a |an |some |more |the |new )?([a-z-]+ ){0,2}(tests?|test cases)\b`),
	rule("test.repair", "test", `\b(repair|update|stabili[sz]e|deflake|unskip)\b (the )?([a-z-]+ ){0,2}tests?\b|\bflaky tests?\b`),

	// debug, part two: fix verbs and symptom words, after test so test work wins over them.
	rule("debug.fix", "debug", `^(please )?(fix|debug)\b|\b(investigate|diagnose|troubleshoot)\b`),
	rule("debug.symptom", "debug", `\b(is|are|keeps) (broken|failing|crashing|hanging|flaky)\b|\b(crash|crashes|crashed|hangs|regressed|returns the wrong)\b|\b(doesn't|does not|isn't|is not) work(ing)?\b|\bnot working\b|\bstopped working\b`),

	// implement: build new functionality or add a feature. Last: the most generic verbs.
	rule("implement.add", "implement", `^(please )?add\b|\badd (a|an|the|support for|new)\b`),
	rule("implement.build", "implement", `^(please )?(implement|create|build|write|introduce|support|make|generate|scaffold|wire up|set up|integrate)\b`),
}

// Rules returns a copy of the ordered rule list.
func Rules() []Rule {
	return append([]Rule(nil), rules...)
}

// Classify applies the rules to a prompt. The prompt is lowercased; nothing of it is returned.
func Classify(prompt string) Result {
	p := strings.ToLower(prompt)
	for i := range rules {
		if rules[i].re.MatchString(p) {
			return Result{Category: rules[i].Category, Rule: rules[i].ID}
		}
	}
	return Result{Category: Unknown}
}

// Session classifies a session by its first genuine prompt; no prompt means unknown.
func Session(s *model.Session) Result {
	if !s.HasFirstPrompt {
		return Result{Category: Unknown}
	}
	return Classify(s.FirstPrompt)
}

// RulePtr returns the rule id as the contract's nullable string.
func (r Result) RulePtr() *string {
	if r.Rule == "" {
		return nil
	}
	id := r.Rule
	return &id
}
