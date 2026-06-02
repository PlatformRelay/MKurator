package mqrest

import (
	"testing"
)

func TestCredentialsFromSecret(t *testing.T) {
	t.Parallel()
	user, pass, err := credentialsFromSecret(map[string][]byte{
		"username":        []byte("mquser"),
		"mqAdminPassword": []byte("secret"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if user != "mquser" || pass != "secret" {
		t.Fatalf("user=%q pass=%q", user, pass)
	}

	_, _, err = credentialsFromSecret(map[string][]byte{"username": []byte("u")})
	if err == nil {
		t.Fatal("expected error when password missing")
	}

	user, pass, err = credentialsFromSecret(map[string][]byte{"password": []byte("p")})
	if err != nil {
		t.Fatal(err)
	}
	if user != "admin" || pass != "p" {
		t.Fatalf("defaults user=%q pass=%q", user, pass)
	}
}

func TestCaPoolFromSecret(t *testing.T) {
	t.Parallel()
	pem := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpL1x5jTMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv
Y2FsaG9zdDAeFw0yNDAxMDEwMDAwMDBaFw0yNTAxMDEwMDAwMDBaMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABG1234567890
-----END CERTIFICATE-----`)
	_, err := caPoolFromSecret(map[string][]byte{"ca.crt": pem})
	if err == nil {
		t.Fatal("expected parse error for invalid PEM")
	}
}
