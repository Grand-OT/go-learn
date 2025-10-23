package path

import (
	"errors"
	"testing"
)

func TestMatch_LiteralMismatch(t *testing.T) {
	ok, parts, err := Match("/abc/def", "/abc/xyz")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ok {
		t.Fatalf("expected not matched")
	}
	if parts != nil {
		t.Fatalf("expected nil params, got %v", parts)
	}
}

func TestMatch_DifferentLengthMeansNoMatch(t *testing.T) {
	ok, parts, err := Match("/a/b/c", "/a/b")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ok {
		t.Fatalf("expected not matched")
	}
	if parts != nil {
		t.Fatalf("expected nil params")
	}
}

func TestMatch_RootPath(t *testing.T) {
	ok, _, err := Match("/", "/")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !ok {
		t.Fatal("expected match")
	}
}

func TestMatch_ParamCaptured(t *testing.T) {
	ok, parts, err := Match("/abc/:id", "/abc/123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ok {
		t.Fatalf("expected match")
	}
	if got := parts["id"]; got != "123" {
		t.Fatalf("id mismatch: got %q want %q", got, "123")
	}
}

func TestMatch_ParamNameUniqueness(t *testing.T) {
	ok, parts, err := Match("/x/:id/:id", "/x/1/2")
	if !errors.Is(err, ErrDuplicateParam) {
		t.Fatalf("expected ErrDuplicateParam, got %v", err)
	}
	if ok {
		t.Fatalf("expected not matched")
	}
	if parts != nil {
		t.Fatalf("expected nil params")
	}
}

func TestMatch_PathDecoding_PerSegment(t *testing.T) {
	ok, parts, err := Match("/abc/:id", "/abc/hello%20world")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ok {
		t.Fatalf("expected match")
	}
	if got := parts["id"]; got != "hello world" {
		t.Fatalf("decoded value mismatch: got %q", got)
	}
}

func TestMatch_BadEncoding_ReturnsError(t *testing.T) {
	ok, parts, err := Match("/abc/:id", "/abc/%ZZ")
	if !errors.Is(err, ErrBadEncoding) {
		t.Fatalf("expected ErrBadEncoding, got %v", err)
	}
	if ok {
		t.Fatalf("expected not matched on bad encoding")
	}
	if parts != nil {
		t.Fatalf("expected nil params")
	}
}

func TestMatch_EncodedSlash_ShouldNotSplitParam(t *testing.T) {
	ok, _, err := Match("/f/:p", "/f/foo%2Fbar")
	if !errors.Is(err, ErrEncodedSlash) {
		t.Fatalf("expected ErrEncodedSlash, got %v", err)
	}
	if ok {
		t.Fatalf("expected mismatched")
	}
}

func TestMatch_TrailingSlashInPath_InvalidPath(t *testing.T) {
	ok, parts, err := Match("/abc/:id", "/abc/123/")
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got %v", err)
	}
	if ok {
		t.Fatalf("expected not matched due to trailing slash")
	}
	if parts != nil {
		t.Fatalf("expected nil params")
	}
}

func TestMatch_EmptyParamValue_NoMatch(t *testing.T) {
	// По твоей реализации пустое значение параметра — это просто no match (404), а не ошибка.
	ok, parts, err := Match("/abc/:id", "/abc/")
	if err != nil {
		t.Fatalf("unexpected err (want nil), got %v", err)
	}
	if ok {
		t.Fatalf("expected not matched for empty param")
	}
	if parts != nil {
		t.Fatalf("expected nil params")
	}
}

func TestMatch_PatternValidationErrors(t *testing.T) {
	_, _, err := Match("", "/abc")
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("expected ErrInvalidPattern for empty pattern, got %v", err)
	}
	_, _, err = Match("abc", "/abc")
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("expected ErrInvalidPattern for missing leading slash, got %v", err)
	}
	_, _, err = Match("/abc/", "/abc")
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("expected ErrInvalidPattern for trailing slash, got %v", err)
	}
}

func TestMatch_MultipleParams(t *testing.T) {
	ok, parts, err := Match("/u/:uid/todos/:id", "/u/42/todos/7")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ok {
		t.Fatalf("expected match")
	}
	if parts["uid"] != "42" || parts["id"] != "7" {
		t.Fatalf("params mismatch: %v", parts)
	}
}
