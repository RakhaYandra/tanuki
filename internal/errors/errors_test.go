package errors

import (
	stderrors "errors"
	"testing"
)

func TestCodeOf(t *testing.T) {
	if got := CodeOf(nil); got != CodeOK {
		t.Fatalf("CodeOf(nil) = %d, want %d", got, CodeOK)
	}
	if got := CodeOf(New(CodeConfig, "bad")); got != CodeConfig {
		t.Fatalf("CodeOf(config) = %d, want %d", got, CodeConfig)
	}
	if got := CodeOf(stderrors.New("x")); got != CodeGeneral {
		t.Fatalf("CodeOf(plain) = %d, want %d", got, CodeGeneral)
	}
}

func TestErrorMessage(t *testing.T) {
	e := Wrap(CodeConflict, "port in use", stderrors.New("busy"))
	if e.Code != CodeConflict {
		t.Fatalf("code = %d", e.Code)
	}
	if got := e.Error(); got != "port in use: busy" {
		t.Fatalf("msg = %q", got)
	}
	plain := New(CodeConfig, "bad yaml")
	if plain.Error() != "bad yaml" {
		t.Fatalf("msg = %q", plain.Error())
	}
}
