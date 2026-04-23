package generator

import "testing"

func TestBase62GeneratorSequence(t *testing.T) {
	gen := NewBase62Generator(0)

	cases := []string{
		"1",
		"2",
		"3",
	}

	for i, want := range cases {
		got := gen.NextKey()
		if got != want {
			t.Fatalf("case %d: got %q, want %q", i, got, want)
		}
	}
}

func TestEncodeBase62(t *testing.T) {
	cases := []struct {
		n    uint64
		want string
	}{
		{0, "0"},
		{61, "Z"},
		{62, "10"},
		{63, "11"},
		{3843, "ZZ"},
	}

	for _, tc := range cases {
		got := encodeBase62(tc.n)
		if got != tc.want {
			t.Fatalf("n=%d: got %q, want %q", tc.n, got, tc.want)
		}
	}
}
