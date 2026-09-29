package auth

import "testing"

func TestLoginRateLimit(t *testing.T) {
	service := &Service{loginAttempts: make(map[string]loginAttempt)}
	for i := 0; i < 5; i++ {
		if !service.allowLoginAttempt("127.0.0.1:admin") {
			t.Fatalf("percobaan %d seharusnya masih diizinkan", i+1)
		}
		service.registerLoginFailure("127.0.0.1:admin")
	}
	if service.allowLoginAttempt("127.0.0.1:admin") {
		t.Fatal("percobaan keenam harus diblokir")
	}
	if !service.allowLoginAttempt("127.0.0.1:lain") {
		t.Fatal("key login lain harus tetap diizinkan")
	}
}

func TestLoginRateLimitResetOnSuccess(t *testing.T) {
	service := &Service{loginAttempts: make(map[string]loginAttempt)}
	for i := 0; i < 4; i++ {
		service.registerLoginFailure("127.0.0.1:admin")
	}
	service.resetLoginAttempt("127.0.0.1:admin")
	if !service.allowLoginAttempt("127.0.0.1:admin") {
		t.Fatal("percobaan harus diizinkan kembali setelah login berhasil")
	}
}

func TestNormalizeUsername(t *testing.T) {
	cases := map[string]string{
		"  PakRT  ": "pakrt",
		"Asep":      "asep",
		" NUR ":     "nur",
	}
	for input, expected := range cases {
		if got := normalizeUsername(input); got != expected {
			t.Fatalf("normalizeUsername(%q) = %q, ingin %q", input, got, expected)
		}
	}
}
