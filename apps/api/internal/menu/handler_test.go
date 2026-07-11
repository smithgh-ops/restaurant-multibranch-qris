package menu

import "testing"

func TestParseOptionalUint64Query(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    *uint64
		wantErr bool
	}{
		{name: "empty means nil", input: "", want: nil},
		{name: "valid number", input: "42", want: ptrUint64(42)},
		{name: "invalid number", input: "abc", wantErr: true},
		{name: "negative number", input: "-1", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseOptionalUint64Query(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", *got)
				}
				return
			}
			if got == nil || *got != *tc.want {
				t.Fatalf("expected %d, got %v", *tc.want, got)
			}
		})
	}
}

func ptrUint64(v uint64) *uint64 { return &v }
