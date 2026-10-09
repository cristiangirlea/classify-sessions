// Package evaluate compares classifier predictions with a labels file and returns the
// contract's classifier.heldout object. Only session ids, categories and rule ids are read
// or reported.
package evaluate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/cristiangirlea/classify-sessions/backend/internal/classify"
	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

// Label is one labelled session.
type Label struct {
	SessionID string `json:"session_id"`
	Category  string `json:"category"`
}

// Load reads a labels file (top-level "labels": [{session_id, category}]) and validates it:
// every category is in the taxonomy, no session id appears twice, and every id is a session
// of the input (known). Error messages carry no file path.
func Load(path string, known map[string]bool) ([]Label, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) {
			err = pe.Err
		}
		return nil, fmt.Errorf("read labels file: %w", err)
	}
	return Parse(raw, known)
}

// Parse validates the bytes of a labels file; see Load.
func Parse(raw []byte, known map[string]bool) ([]Label, error) {
	var doc struct {
		Labels []Label `json:"labels"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, errors.New("labels file is not valid JSON of the expected shape")
	}
	if doc.Labels == nil {
		return nil, errors.New("labels file has no labels array")
	}
	seen := map[string]bool{}
	for i, l := range doc.Labels {
		switch {
		case l.SessionID == "":
			return nil, fmt.Errorf("labels[%d] has no session_id", i)
		case l.Category == "":
			return nil, fmt.Errorf("labels[%d] has no category", i)
		case !classify.InTaxonomy(l.Category):
			return nil, fmt.Errorf("labels[%d] has category %q, not in the taxonomy", i, l.Category)
		case seen[l.SessionID]:
			return nil, fmt.Errorf("labels[%d] repeats session %s", i, l.SessionID)
		case !known[l.SessionID]:
			return nil, fmt.Errorf("labels[%d] names session %s, which is not in the input", i, l.SessionID)
		}
		seen[l.SessionID] = true
	}
	return doc.Labels, nil
}

// SessionIDs is the set of session ids of an input, for Load's presence check.
func SessionIDs(sessions []*model.Session) map[string]bool {
	ids := make(map[string]bool, len(sessions))
	for _, s := range sessions {
		ids[s.ID] = true
	}
	return ids
}

// Predict classifies every session.
func Predict(sessions []*model.Session) map[string]classify.Result {
	out := make(map[string]classify.Result, len(sessions))
	for _, s := range sessions {
		out[s.ID] = classify.Session(s)
	}
	return out
}

// Empty is the heldout object when no labels file is given.
func Empty() contract.Heldout {
	return contract.Heldout{Disagreements: []contract.Disagreement{}}
}

// Compare scores predictions against labels. A labelled session missing from pred counts as
// predicted unknown with no rule. Disagreements are sorted by session id; agreement_fraction
// is null when nothing is labelled.
func Compare(labels []Label, pred map[string]classify.Result) contract.Heldout {
	h := Empty()
	h.Labelled = len(labels)
	for _, l := range labels {
		p, ok := pred[l.SessionID]
		if !ok {
			p = classify.Result{Category: classify.Unknown}
		}
		if p.Category == l.Category {
			h.Agreements++
			continue
		}
		h.Disagreements = append(h.Disagreements, contract.Disagreement{
			SessionID: l.SessionID,
			Labelled:  l.Category,
			Predicted: p.Category,
			Rule:      p.RulePtr(),
		})
	}
	sort.Slice(h.Disagreements, func(i, j int) bool {
		return h.Disagreements[i].SessionID < h.Disagreements[j].SessionID
	})
	if h.Labelled > 0 {
		f := float64(h.Agreements) / float64(h.Labelled)
		h.AgreementFraction = &f
	}
	return h
}
