package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("password-123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if !VerifyPassword(hash, "password-123") {
		t.Fatal("password yang benar harus terverifikasi")
	}
	if VerifyPassword(hash, "password-124") {
		t.Fatal("password yang salah tidak boleh terverifikasi")
	}
}

func TestHashPasswordRejectsShortPassword(t *testing.T) {
	if _, err := HashPassword("pendek"); err == nil {
		t.Fatal("password pendek harus ditolak")
	}
}
