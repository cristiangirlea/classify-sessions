package contract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func decodeAny(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// TestExamplesDecodeStrictly decodes each contract example into its Go struct with
// unknown fields disallowed, then re-marshals it and requires the same key set, so a
// field missing from the struct or from the example fails in either direction.
func TestExamplesDecodeStrictly(t *testing.T) {
	cases := []struct {
		file string
		dst  any
	}{
		{"summary.json", &SummaryResponse{}},
		{"models.json", &ModelsResponse{}},
		{"error.json", &ErrorResponse{}},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "..", "contract", "examples", c.file))
			if err != nil {
				t.Fatal(err)
			}
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.DisallowUnknownFields()
			if err := dec.Decode(c.dst); err != nil {
				t.Fatalf("strict decode: %v", err)
			}
			again, err := json.Marshal(c.dst)
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Join(KeySet(decodeAny(t, raw)), "\n")
			got := strings.Join(KeySet(decodeAny(t, again)), "\n")
			if want != got {
				t.Fatalf("key sets differ\nexample:\n%s\n\nstruct:\n%s", want, got)
			}
		})
	}
}
