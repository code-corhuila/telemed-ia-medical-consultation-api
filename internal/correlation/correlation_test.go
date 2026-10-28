package correlation

import (
	"context"
	"testing"
)

func TestParseTraceparent_Valid(t *testing.T) {
	// W3C example from the spec.
	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	spanID := "00f067aa0ba902b7"
	header := "00-" + traceID + "-" + spanID + "-01"

	gotTrace, gotSpan, ok := ParseTraceparent(header)
	if !ok {
		t.Fatalf("expected valid traceparent, got invalid")
	}
	if gotTrace != traceID {
		t.Errorf("traceID: got %q, want %q", gotTrace, traceID)
	}
	if gotSpan != spanID {
		t.Errorf("spanID: got %q, want %q", gotSpan, spanID)
	}
}

func TestParseTraceparent_RejectsInvalid(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no separator", "garbage"},
		{"too few parts", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7"},
		{"too many parts", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra"},
		{"unknown version", "01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
		{"traceID too short", "00-4bf92f3577b34da6a3ce929d0e0e47-00f067aa0ba902b7-01"},
		{"traceID too long", "00-4bf92f3577b34da6a3ce929d0e0e47366-00f067aa0ba902b7-01"},
		{"spanID too short", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b-01"},
		{"flags too short", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-0"},
		{"traceID all zeros", "00-00000000000000000000000000000000-00f067aa0ba902b7-01"},
		{"spanID all zeros", "00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01"},
		{"traceID not hex", "00-zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz-00f067aa0ba902b7-01"},
		{"spanID not hex", "00-4bf92f3577b34da6a3ce929d0e0e4736-zzzzzzzzzzzzzzzz-01"},
		{"flags not hex", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-zz"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, ok := ParseTraceparent(c.input)
			if ok {
				t.Fatalf("expected invalid, got valid: %q", c.input)
			}
		})
	}
}

func TestWithTraceparent_FromContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceparent(ctx, "aaa", "bbb")

	if got := TraceIDFromContext(ctx); got != "aaa" {
		t.Errorf("TraceID: got %q, want %q", got, "aaa")
	}
	if got := SpanIDFromContext(ctx); got != "bbb" {
		t.Errorf("SpanID: got %q, want %q", got, "bbb")
	}
}

func TestFromContext_WithoutValues(t *testing.T) {
	ctx := context.Background()
	if got := FromContext(ctx); got != "" {
		t.Errorf("FromContext: got %q, want empty", got)
	}
	if got := TraceIDFromContext(ctx); got != "" {
		t.Errorf("TraceIDFromContext: got %q, want empty", got)
	}
	if got := SpanIDFromContext(ctx); got != "" {
		t.Errorf("SpanIDFromContext: got %q, want empty", got)
	}
}
