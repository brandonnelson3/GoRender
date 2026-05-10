package rendertest_test

import (
	"flag"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// update controls whether actuals are promoted to golden files.
// Run with: go test ./rendertest/... -update
var update = flag.Bool("update", false, "overwrite golden files with the current actual output")

const (
	goldenDir = "testdata/golden"
	actualDir  = "testdata/actual"

	// maxChannelDelta is the per-channel (R/G/B) tolerance in [0,255].
	// Set to 2 to absorb GPU driver float rounding without masking real regressions.
	maxChannelDelta = 2
)

// TestGoldens compares every PNG in testdata/actual against its golden counterpart.
//
// Workflow:
//  1. Run `go run . -rendertest` to populate testdata/actual/.
//  2. Run `go test ./rendertest/... -update` to promote actuals → goldens (first time).
//  3. Commit testdata/golden/*.png.
//  4. In CI, run steps 1 then `go test ./rendertest/...` (no -update).
func TestGoldens(t *testing.T) {
	actuals, err := filepath.Glob(filepath.Join(actualDir, "*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(actuals) == 0 {
		t.Skip("no actual images found — run 'go run . -rendertest' first to generate them")
	}

	for _, actualPath := range actuals {
		name := filepath.Base(actualPath)
		t.Run(name, func(t *testing.T) {
			actual := mustLoadPNG(t, actualPath)
			goldenPath := filepath.Join(goldenDir, name)

			if *update {
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("creating golden dir: %v", err)
				}
				mustSavePNG(t, goldenPath, actual)
				t.Logf("updated golden: %s", goldenPath)
				return
			}

			if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
				t.Fatalf("golden file missing: %s\nRun 'go test ./rendertest/... -update' to create it.", goldenPath)
			}
			golden := mustLoadPNG(t, goldenPath)
			compareImages(t, golden, actual)
		})
	}
}

// compareImages performs per-pixel, per-channel comparison with a small tolerance.
func compareImages(t *testing.T, golden, actual image.Image) {
	t.Helper()
	gb := golden.Bounds()
	ab := actual.Bounds()
	if gb != ab {
		t.Fatalf("image size mismatch: golden=%v actual=%v", gb, ab)
	}

	var (
		maxDelta   int
		totalDelta int64
		diffPixels int
	)
	pixels := gb.Dx() * gb.Dy()

	for y := gb.Min.Y; y < gb.Max.Y; y++ {
		for x := gb.Min.X; x < gb.Max.X; x++ {
			gr, gg, gb_, _ := golden.At(x, y).RGBA() // 16-bit
			ar, ag, ab_, _ := actual.At(x, y).RGBA()

			dr := absDiff(gr>>8, ar>>8)
			dg := absDiff(gg>>8, ag>>8)
			db := absDiff(gb_>>8, ab_>>8)

			pixMax := maxOf(dr, dg, db)
			if pixMax > maxDelta {
				maxDelta = pixMax
			}
			totalDelta += int64(dr + dg + db)
			if pixMax > maxChannelDelta {
				diffPixels++
			}
		}
	}

	avgDelta := float64(totalDelta) / float64(int64(pixels)*3)
	diffPct := float64(diffPixels) / float64(pixels) * 100

	if maxDelta > maxChannelDelta {
		t.Errorf(
			"image diff exceeds threshold (max_channel_delta=%d, threshold=%d, avg=%.3f, pixels_over_threshold=%d/%.0f%%)",
			maxDelta, maxChannelDelta, avgDelta, diffPixels, math.Round(diffPct),
		)
	}
}

func absDiff(a, b uint32) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func maxOf(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func mustLoadPNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func mustSavePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}
