package cryptoUtil

import (
	"bytes"
	"testing"
)

func TestGenerateRecoveryPassword_Deterministic(t *testing.T) {
	salt := []byte("test-salt-abc")
	a := GenerateRecoveryPassword([]string{"MyPet"}, salt)
	b := GenerateRecoveryPassword([]string{"MyPet"}, salt)
	if a != b {
		t.Fatalf("expected same result for same inputs, got %q vs %q", a, b)
	}
}

func TestGenerateRecoveryPassword_CaseAndSpaceNormalized(t *testing.T) {
	salt := []byte("test-salt-xyz")
	// Answers are normalized to lowercase + trimmed, so all three must produce the same key.
	a := GenerateRecoveryPassword([]string{"MyPet"}, salt)
	b := GenerateRecoveryPassword([]string{"mypet"}, salt)
	c := GenerateRecoveryPassword([]string{"  mypet  "}, salt)
	if a != b || b != c {
		t.Fatalf("normalization failed: %q %q %q", a, b, c)
	}
}

func TestGenerateRecoveryPassword_DifferentAnswersProduceDifferentKeys(t *testing.T) {
	salt := []byte("shared-salt")
	a := GenerateRecoveryPassword([]string{"correctAnswer"}, salt)
	b := GenerateRecoveryPassword([]string{"wrongAnswer"}, salt)
	if a == b {
		t.Fatal("different answers must produce different keys")
	}
}

func TestGenerateRecoveryPassword_DifferentSaltsProduceDifferentKeys(t *testing.T) {
	a := GenerateRecoveryPassword([]string{"sameAnswer"}, []byte("salt-one"))
	b := GenerateRecoveryPassword([]string{"sameAnswer"}, []byte("salt-two"))
	if a == b {
		t.Fatal("different salts must produce different keys, even for the same answer")
	}
}

func TestGenerateRecoveryPassword_OutputLength(t *testing.T) {
	// PBKDF2 with 32-byte key -> 64 hex chars
	result := GenerateRecoveryPassword([]string{"answer"}, []byte("salt"))
	if len(result) != 64 {
		t.Fatalf("expected 64 hex chars, got %d: %q", len(result), result)
	}
}

func TestGenerateRecoveryPassword_FallbackSaltBackwardsCompat(t *testing.T) {
	// Confirm old static-salt behaviour still produces a deterministic result
	// (allows recovery for keys generated before per-key salts were introduced).
	staticSalt := []byte("ecnotes-static-salt-v1")
	first := GenerateRecoveryPassword([]string{"mypet"}, staticSalt)
	second := GenerateRecoveryPassword([]string{"mypet"}, staticSalt)
	if first != second {
		t.Fatal("static-salt fallback must be deterministic")
	}
	if bytes.Equal([]byte(first), []byte(GenerateRecoveryPassword([]string{"mypet"}, []byte("different-salt")))) {
		t.Fatal("static salt and a different salt must diverge")
	}
}
