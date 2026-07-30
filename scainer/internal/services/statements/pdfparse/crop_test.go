package pdfparse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCropPages_example10(t *testing.T) {
	src := filepath.Join("..", "..", "..", "..", "examples", "10.pdf")
	if _, err := os.Stat(src); err != nil {
		t.Skip(src, err)
	}
	dir := t.TempDir()
	dst := filepath.Join(dir, "G.pdf")
	if err := CropPages(src, dst, PageRange{StartPage: 1, EndPage: 2}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() < 1000 {
		t.Fatalf("cropped pdf too small: %d", info.Size())
	}
}
