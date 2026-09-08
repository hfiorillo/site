// Command gallery builds static WebP previews and a gallery manifest locally.
// Run after adding photos, then commit the generated files with the post.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hfiorillo/site/internal/markdown"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
)

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	encoder, err := exec.LookPath("cwebp")
	if err != nil {
		return fmt.Errorf("install the WebP tools first (macOS: brew install webp): %w", err)
	}
	posts, err := markdown.LoadMarkdownPosts(context.Background())
	if err != nil {
		return err
	}
	outputDir := strings.TrimPrefix(paths.GalleryAssets, "/")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	groups := []models.GalleryYear{}
	retained := map[string]bool{}
	seen := map[string]bool{}
	var originals, previews int64
	count := 0
	for _, post := range posts {
		entries, err := os.ReadDir(filepath.Join(paths.PostImagesDirectory, post.Filename))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		for _, entry := range entries {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if entry.IsDir() || (ext != ".jpg" && ext != ".jpeg" && ext != ".png") {
				continue
			}
			source := filepath.Join(paths.PostImagesDirectory, post.Filename, entry.Name())
			raw, err := os.ReadFile(source)
			if err != nil {
				return err
			}
			// Include encoder settings in the key so changing quality/size invalidates caches.
			digest := sha256.Sum256(append([]byte("gallery-v2-oriented-q72-360-720:"), raw...))
			key := fmt.Sprintf("%x", digest[:16])
			if seen[key] {
				continue
			}
			seen[key] = true
			f, err := os.Open(source)
			if err != nil {
				return err
			}
			cfg, _, decodeErr := image.DecodeConfig(f)
			closeErr := f.Close()
			if decodeErr != nil {
				return fmt.Errorf("decoding %s: %w", source, decodeErr)
			}
			if closeErr != nil {
				return closeErr
			}
			orientation := jpegOrientation(raw)
			if orientation >= 5 {
				cfg.Width, cfg.Height = cfg.Height, cfg.Width
			}
			width := min(720, cfg.Width)
			filename := fmt.Sprintf("%s-%d.webp", key, width)
			output := filepath.Join(outputDir, filename)
			if info, err := os.Stat(output); err != nil || info.Size() == 0 {
				if err := encodePreview(encoder, source, raw, orientation, width, output); err != nil {
					return err
				}
			}
			retained[filename] = true
			info, err := os.Stat(output)
			if err != nil {
				return err
			}
			previews += info.Size()
			photo := models.GalleryPhoto{Original: paths.PostImages + post.Filename + "/" + entry.Name(), Preview: paths.GalleryAssets + filename, Title: post.Title, PostURL: paths.Blog + "/" + post.Filename, Width: cfg.Width, Height: cfg.Height}

			year := post.Date.Year()
			if len(groups) == 0 || groups[len(groups)-1].Year != year {
				groups = append(groups, models.GalleryYear{Year: year})
			}
			groups[len(groups)-1].Photos = append(groups[len(groups)-1].Photos, photo)
			originals += int64(len(raw))
			count++
		}
	}
	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(paths.GalleryManifest, append(data, '\n'), 0644); err != nil {
		return err
	}
	if err := prunePreviews(outputDir, retained); err != nil {
		return err
	}
	fmt.Printf("Gallery: %d photos, originals %.1f MB, previews %.1f MB (%.0f%% smaller).\n", count, float64(originals)/1e6, float64(previews)/1e6, 100*(1-float64(previews)/float64(max(originals, 1))))
	return nil
}

// Encode only one preview per photo, normalising orientation without touching the original.
func encodePreview(encoder, source string, raw []byte, orientation, width int, output string) error {
	input := source
	if orientation != 1 {
		var err error
		input, err = orientedPNG(raw, orientation)
		if err != nil {
			return err
		}
		defer os.Remove(input) // Best-effort cleanup of temporary pixels on encoder failure.
	}
	tmp, err := os.CreateTemp(filepath.Dir(output), ".preview-*.webp")
	if err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // Clean up any incomplete encoding.
	command := exec.Command(encoder, "-quiet", "-q", "72", "-m", "6", "-resize", fmt.Sprint(width), "0", input, "-o", tmp.Name())
	if result, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("encoding %s: %w: %s", source, err, result)
	}
	return os.Rename(tmp.Name(), output)
}

// Remove only files owned by this generator, after the new manifest is written.
func prunePreviews(dir string, retained map[string]bool) error {
	generated := regexp.MustCompile(`^[a-f0-9]{32}-[0-9]+\.webp$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && generated.MatchString(entry.Name()) && !retained[entry.Name()] {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
