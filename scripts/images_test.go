package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestConversionProtectsOriginals(t *testing.T) {
	script, err := filepath.Abs("images.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, encoder string
		success       bool
	}{
		{"encoder fails", "#!/bin/sh\nexit 1\n", false},
		{"encoder produces empty file", "#!/bin/sh\nexit 0\n", false},
		{"successful conversion", "#!/bin/sh\nif [ \"$1\" = -g ]; then echo 'format: jpeg'; exit 0; fi\nfor target; do :; done\nprintf 'converted jpeg' > \"$target\"\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "public", "images")
			bin := filepath.Join(root, "bin")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(bin, 0700); err != nil {
				t.Fatal(err)
			}
			original := filepath.Join(dir, "photo with spaces.heic")
			if err := os.WriteFile(original, []byte("precious original"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(bin, "sips"), []byte(tc.encoder), 0700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("/bin/sh", script)
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"))
			output, err := cmd.CombinedOutput()
			if (err == nil) != tc.success {
				t.Fatalf("unexpected conversion status %v: %s", err, output)
			}
			_, originalErr := os.Stat(original)
			_, jpegErr := os.Stat(filepath.Join(dir, "photo with spaces.jpg"))
			if tc.success {
				if !os.IsNotExist(originalErr) || jpegErr != nil {
					t.Fatal("successful conversion incomplete")
				}
			} else {
				if originalErr != nil || !os.IsNotExist(jpegErr) {
					t.Fatal("failed conversion damaged files")
				}
			}
		})
	}
}
