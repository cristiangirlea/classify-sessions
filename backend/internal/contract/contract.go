// Package contract mirrors contract/*.schema.json. The schemas are the source of truth;
// these structs are checked against contract/examples by TestExamplesDecodeStrictly.
// Nullable schema fields are pointers so they marshal as null rather than disappearing.
package contract

type SummaryResponse struct {
	GeneratedAt    string            `json:"generated_at"`
	Source         Source            `json:"source"`
	Records        Records           `json:"records"`
	Sessions       Sessions          `json:"sessions"`
	Categories     []CategorySummary `json:"categories"`
	Spend          Spend             `json:"spend"`
	Pricing        Pricing           `json:"pricing"`
	Classifier     Classifier        `json:"classifier"`
	SessionList    []SessionRow      `json:"session_list"`
	ExampleRequest string            `json:"example_request"`
	ModelChoice    ModelsResponse    `json:"model_choice"`
}

type Source struct {
	Files     int    `json:"files"`
	InputHint string `json:"input_hint"`
}

type Records struct {
	Total            int         `json:"total"`
	Parsed           int         `json:"parsed"`
	Malformed        int         `json:"malformed"`
	Unknown          int         `json:"unknown"`
	Duplicate        int         `json:"duplicate"`
	UnparsedFraction float64     `json:"unparsed_fraction"`
	Diagnostics      Diagnostics `json:"diagnostics"`
}

type Diagnostics struct {
	MissingUsage       int `json:"missing_usage"`
	MissingTimestamp   int `json:"missing_timestamp"`
	MissingCacheFields int `json:"missing_cache_fields"`
	Sidechain          int `json:"sidechain"`
}

type Sessions struct {
	Total                        int     `json:"total"`
	Classified                   int     `json:"classified"`
	Unclassified                 int     `json:"unclassified"`
	UnclassifiedFractionSessions float64 `json:"unclassified_fraction_sessions"`
	UnclassifiedFractionRecords  float64 `json:"unclassified_fraction_records"`
}

type Components struct {
	InputUSD      float64 `json:"input_usd"`
	OutputUSD     float64 `json:"output_usd"`
	CacheReadUSD  float64 `json:"cache_read_usd"`
	CacheWriteUSD float64 `json:"cache_write_usd"`
}

type CategorySummary struct {
	Category               string     `json:"category"`
	Sessions               int        `json:"sessions"`
	Records                int        `json:"records"`
	CostUSD                float64    `json:"cost_usd"`
	CostIncompleteSessions int        `json:"cost_incomplete_sessions"`
	Components             Components `json:"components"`
}

type Spend struct {
	TotalUSD           float64      `json:"total_usd"`
	IncompleteSessions int          `json:"incomplete_sessions"`
	Components         Components   `json:"components"`
	ByModel            []ModelSpend `json:"by_model"`
}

type ModelSpend struct {
	Model    string   `json:"model"`
	PriceKey *string  `json:"price_key"`
	Requests int      `json:"requests"`
	CostUSD  float64  `json:"cost_usd"`
	Flags    []string `json:"flags"`
}

type Pricing struct {
	SourceURL            string   `json:"source_url"`
	AsOf                 string   `json:"as_of"`
	CacheReadMultiplier  float64  `json:"cache_read_multiplier"`
	CacheWriteMultiplier float64  `json:"cache_write_multiplier"`
	Assumptions          []string `json:"assumptions"`
}

type Classifier struct {
	Method   string   `json:"method"`
	Taxonomy []string `json:"taxonomy"`
	Heldout  Heldout  `json:"heldout"`
}

type Heldout struct {
	Labelled          int            `json:"labelled"`
	Agreements        int            `json:"agreements"`
	AgreementFraction *float64       `json:"agreement_fraction"`
	Disagreements     []Disagreement `json:"disagreements"`
}

type Disagreement struct {
	SessionID string  `json:"session_id"`
	Labelled  string  `json:"labelled"`
	Predicted string  `json:"predicted"`
	Rule      *string `json:"rule"`
}

type SessionRow struct {
	ID           string   `json:"id"`
	StartedAt    *string  `json:"started_at"`
	EndedAt      *string  `json:"ended_at"`
	Records      int      `json:"records"`
	Category     string   `json:"category"`
	Rule         *string  `json:"rule"`
	CostUSD      float64  `json:"cost_usd"`
	CostComplete bool     `json:"cost_complete"`
	Requests     int      `json:"requests"`
	Models       []string `json:"models"`
	ToolCalls    int      `json:"tool_calls"`
	ToolErrors   int      `json:"tool_errors"`
}

type ModelsResponse struct {
	Models           []string         `json:"models"`
	Categories       []ModelCategory  `json:"categories"`
	ModelTotals      []ModelCell      `json:"model_totals"`
	ExcludedSessions []string         `json:"excluded_sessions"`
	Advice           ModelAdvice      `json:"advice"`
	Definitions      ModelDefinitions `json:"definitions"`
	PricingSource    string           `json:"pricing_source"`
	PricingDate      string           `json:"pricing_date"`
}

type ModelCategory struct {
	Category   string      `json:"category"`
	Comparable bool        `json:"comparable"`
	Supported  bool        `json:"supported"`
	Reason     string      `json:"reason"`
	Cells      []ModelCell `json:"cells"`
}

type ModelCell struct {
	Model                 string   `json:"model"`
	Category              string   `json:"category"`
	Sessions              int      `json:"sessions"`
	MixedModelSessions    int      `json:"mixed_model_sessions"`
	LongContextSessions   int      `json:"long_context_sessions"`
	CompleteCostSessions  int      `json:"complete_cost_sessions"`
	MeanKnownUSD          float64  `json:"mean_known_usd"`
	MedianKnownUSD        float64  `json:"median_known_usd"`
	TotalKnownUSD         float64  `json:"total_known_usd"`
	UserMessages          int      `json:"user_messages"`
	CorrectionMessages    int      `json:"correction_messages"`
	CorrectionRate        *float64 `json:"correction_rate"`
	CorrectedSessions     int      `json:"corrected_sessions"`
	ToolResultsKnown      int      `json:"tool_results_known"`
	ToolFailures          int      `json:"tool_failures"`
	ToolFailureRate       *float64 `json:"tool_failure_rate"`
	UnknownOutcomeResults int      `json:"unknown_outcome_results"`
	SessionIDs            []string `json:"session_ids"`
}

type ModelAdvice struct {
	ExpensiveModel   *string  `json:"expensive_model"`
	ComparatorModel  *string  `json:"comparator_model"`
	SharedCategories []string `json:"shared_categories"`
	CostRatio        *float64 `json:"cost_ratio"`
	Verdict          string   `json:"verdict"`
	Statement        string   `json:"statement"`
	Caveats          []string `json:"caveats"`
}

type ModelDefinitions struct {
	DominantModel      string `json:"dominant_model"`
	ComparableCategory string `json:"comparable_category"`
	SupportedCategory  string `json:"supported_category"`
	CostPerSession     string `json:"cost_per_session"`
	CorrectionMessage  string `json:"correction_message"`
	CorrectionRate     string `json:"correction_rate"`
	ToolFailureRate    string `json:"tool_failure_rate"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
