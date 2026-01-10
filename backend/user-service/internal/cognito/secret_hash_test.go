package cognito

import "testing"

func TestSecretHash_Deterministic(t *testing.T) {
	a := SecretHash("user@example.com", "client", "secret")
	b := SecretHash("user@example.com", "client", "secret")
	if a != b {
		t.Fatalf("expected deterministic secret hash")
	}
	if a == "" {
		t.Fatalf("expected non-empty secret hash")
	}
}
