package handler

import "testing"

func TestParseTodoLimit(t *testing.T) {
	cases := map[string]int{
		"":    defaultTodoLimit,
		"0":   defaultTodoLimit,
		"-1":  defaultTodoLimit,
		"7":   7,
		"500": maxTodoLimit,
	}

	for raw, want := range cases {
		if got := parseTodoLimit(raw); got != want {
			t.Fatalf("parseTodoLimit(%q) = %d, want %d", raw, got, want)
		}
	}
}
