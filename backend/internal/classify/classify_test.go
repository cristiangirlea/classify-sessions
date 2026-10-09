package classify

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

// TestClassifyRulesTableDriven runs invented prompts, written for this test and not taken
// from the corpus, through rules-v1. Subtest names and failure messages carry the case index
// and rule ids only, never the prompt.
func TestClassifyRulesTableDriven(t *testing.T) {
	cases := []struct {
		prompt   string
		category string
		rule     string
	}{
		// One or more cases per rule.
		{"Plan the migration of the photo archive to object storage", "plan", "plan.lead"},
		{"Before making any changes, sketch how the billing worker should retry webhooks", "plan", "plan.before-code"},
		{"Write a design for the notification centre", "plan", "plan.write-plan"},
		{"Is the new rate card importer ready to merge?", "review", "review.ship-ready"},
		{"Go through the session store for anything that might leak memory", "review", "review.find-breakage"},
		{"Scan the queue consumer: anything that could break if two workers start at once?", "review", "review.find-breakage"},
		{"Code review my branch before Friday", "review", "review.review"},
		{"Please review this patch to the tax tables", "review", "review.review"},
		{"Why is the nightly import skipping half the rows?", "debug", "debug.why-does"},
		{"KeyError: 'region' shows up whenever the exporter starts", "debug", "debug.error-message"},
		{"How does the retry queue work when the broker restarts?", "research", "research.how-does"},
		{"Walk me through the lifecycle of a print job", "research", "research.explain"},
		{"Compare SQLite and DuckDB for the reporting cache, no code yet", "research", "research.compare"},
		{"Refactoring pass on the invoice formatter, keep outputs identical", "refactor", "refactor.refactor"},
		{"Merge the two near-identical payroll exporters without changing behaviour", "refactor", "refactor.behaviour-preserving"},
		{"Rename the misleading flags in the cart service", "refactor", "refactor.restructure"},
		{"Set up a pipeline that lints every merge request", "ops", "ops.ci-container"},
		{"Cut a release candidate for the mobile app", "ops", "ops.deploy-release"},
		{"Bump the image library to its latest minor version", "ops", "ops.dependencies"},
		{"Raise coverage of the tariff calculator above ninety percent", "test", "test.coverage"},
		{"Write unit tests for the coupon validator", "test", "test.write-tests"},
		{"Deflake the checkout end-to-end test", "test", "test.repair"},
		{"Fix the off-by-one in the calendar grid", "debug", "debug.fix"},
		{"The upload worker stalls overnight, please investigate", "debug", "debug.fix"},
		{"The export button is broken on narrow screens", "debug", "debug.symptom"},
		{"Add a favourites tab to the profile screen", "implement", "implement.add"},
		{"Implement single sign-on for the admin console", "implement", "implement.build"},

		// Precedence: plan beats implement.
		{"Draft a proposal, then add a search box to the header", "plan", "plan.write-plan"},
		// Precedence: test nouns beat a generic add or write.
		{"Add property-based tests for the slug generator", "test", "test.write-tests"},
		// A test file path alone is not a test noun.
		{"Build a tooltip helper inside widgets/chart_test.go", "implement", "implement.build"},
		// Ops before test: CI work that mentions tests stays ops.
		{"Make the CI job run the integration tests nightly", "ops", "ops.ci-container"},
		// Repairing a test is test, ahead of the debug fix verbs.
		{"Fix the flaky test in the cart module", "test", "test.repair"},
		// Precedence: symptom questions are debug, even when they mention working.
		{"Why does the search page work only for admins?", "debug", "debug.why-does"},
		{"How come the dashboard counts double after midnight?", "debug", "debug.why-does"},
		// Precedence: "how does X work/handle" is research.
		{"How does the scheduler handle leap seconds?", "research", "research.how-does"},
		// Precedence: review phrasing beats debug words.
		{"Look over the refund flow for bugs before the audit", "review", "review.find-breakage"},
		{"Is the CSV importer good enough to ship? Be blunt.", "review", "review.ship-ready"},

		// No rule.
		{"Thanks, that was helpful", "unknown", ""},
		{"ok", "unknown", ""},
		{"", "unknown", ""},
		{"   ", "unknown", ""},
	}

	fired := map[string]bool{}
	for i, c := range cases {
		label := c.rule
		if label == "" {
			label = "none"
		}
		t.Run(fmt.Sprintf("%02d-%s", i, label), func(t *testing.T) {
			got := Classify(c.prompt)
			if got.Category != c.category || got.Rule != c.rule {
				t.Errorf("case %d: got %s/%q, want %s/%q", i, got.Category, got.Rule, c.category, c.rule)
			}
			if (got.RulePtr() == nil) != (c.rule == "") {
				t.Errorf("case %d: RulePtr nil-ness wrong", i)
			}
			if upper := Classify(strings.ToUpper(c.prompt)); upper != got {
				t.Errorf("case %d: classification depends on case", i)
			}
		})
		fired[c.rule] = true
	}

	idShape := regexp.MustCompile(`^[a-z]+\.[a-z0-9-]+$`)
	seen := map[string]bool{}
	for _, r := range Rules() {
		if seen[r.ID] {
			t.Errorf("duplicate rule id %s", r.ID)
		}
		seen[r.ID] = true
		if !idShape.MatchString(r.ID) || !strings.HasPrefix(r.ID, r.Category+".") {
			t.Errorf("rule id %s does not have the form %s.slug", r.ID, r.Category)
		}
		if !InTaxonomy(r.Category) || r.Category == Unknown {
			t.Errorf("rule %s has category %s", r.ID, r.Category)
		}
		if !fired[r.ID] {
			t.Errorf("rule %s has no test case", r.ID)
		}
	}
	covered := map[string]bool{}
	for _, c := range cases {
		covered[c.category] = true
	}
	for _, cat := range Taxonomy {
		if !covered[cat] {
			t.Errorf("category %s has no test case", cat)
		}
	}

	t.Run("session", func(t *testing.T) {
		if r := Session(&model.Session{}); r.Category != Unknown || r.Rule != "" {
			t.Errorf("session without a prompt: %+v", r)
		}
		s := &model.Session{HasFirstPrompt: true, FirstPrompt: "Add a favourites tab to the profile screen"}
		if r := Session(s); r.Rule != "implement.add" {
			t.Errorf("session with a prompt: rule %q", r.Rule)
		}
	})
}
