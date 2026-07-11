package branch

import "testing"

func TestParseOptionalActiveParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    *bool
		wantErr bool
	}{
		{name: "empty means nil", input: "", want: nil},
		{name: "true", input: "true", want: ptrBool(true)},
		{name: "false", input: "false", want: ptrBool(false)},
		{name: "one", input: "1", want: ptrBool(true)},
		{name: "zero", input: "0", want: ptrBool(false)},
		{name: "invalid", input: "yes", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseOptionalActiveParam(tc.input)
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
				t.Fatalf("expected %v, got %v", *tc.want, got)
			}
		})
	}
}

func ptrBool(v bool) *bool { return &v }
