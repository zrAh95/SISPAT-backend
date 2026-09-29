package absensi

import (
	"os"
	"testing"
)

func TestValidatePhotoRejectsNonImage(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "evidence-*.php")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString("<?php echo 'sebaiknya ditolak'; ?>"); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, _, err := validatePhoto(file); err == nil {
		t.Fatal("file non-gambar harus ditolak")
	}
}

func TestValidatePhotoRejectsEmptyFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "evidence-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, _, err := validatePhoto(file); err == nil {
		t.Fatal("file kosong harus ditolak")
	}
}
