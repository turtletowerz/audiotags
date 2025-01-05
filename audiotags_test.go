package audiotags_test

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/fs"
	"os"
	"testing"

	"github.com/sentriz/audiotags"
)

type testdata struct {
	openErr  error
	hasMedia bool
	noTags   bool
}

var defaultData = testdata{nil, true, false}

var fileTestdata = map[string]testdata{
	"testdata/metadata":          {audiotags.ErrBadFile, false, false},
	"testdata/sample.ape":        defaultData,
	"testdata/sample.flac":       defaultData,
	"testdata/sample.id3v11.mp3": {nil, true, true}, // TODO: This appears to be a bug as there are actually tags on the file
	"testdata/sample.id3v22.mp3": defaultData,
	"testdata/sample.id3v23.mp3": defaultData,
	"testdata/sample.id3v24.mp3": defaultData,
	"testdata/sample.m4a":        defaultData,
	"testdata/sample.mp4":        defaultData,
	"testdata/sample.ogg":        defaultData,
	"testdata/sample.wv":         defaultData,
}

func withOpen(t *testing.T, mem bool, fn func(*audiotags.File, testdata) error) {
	err := fs.WalkDir(os.DirFS("."), "testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		var audio *audiotags.File
		if mem {
			file, ferr := os.Open(path)
			if ferr != nil {
				return fmt.Errorf("could not open file %q: %s", path, ferr)
			}
			defer file.Close()

			audio, err = audiotags.OpenReader(file)
		} else {
			audio, err = audiotags.Open(path)
		}

		data, ok := fileTestdata[path]
		if !ok {
			return fmt.Errorf("could not find testdata info for file %q", path)
		}

		if !errors.Is(err, data.openErr) {
			return fmt.Errorf("expected error %q but got %q for file %q", data.openErr, err, path)
		}

		if err != nil {
			return nil
		}

		defer audio.Close()

		if fn != nil {
			if err = fn(audio, data); err != nil {
				return fmt.Errorf("test failed for file %q: %s", path, err)
			}
		}
		return nil
	})

	if err != nil {
		t.Fatal("Error: " + err.Error())
	}
}

func TestOpenMem(t *testing.T) {
	withOpen(t, true, nil)
}

func TestOpen(t *testing.T) {
	withOpen(t, false, nil)
}

func TestHasMedia(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, data testdata) error {
		hasMedia := f.HasMedia()
		if hasMedia != data.hasMedia {
			return fmt.Errorf("expected hasMedia %t but got %t", data.hasMedia, hasMedia)
		}
		return nil
	})
}

func TestReadTags(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, data testdata) error {
		tags := f.ReadTags()
		if len(tags) == 0 && !data.noTags {
			return fmt.Errorf("no tags")
		}
		return nil
	})
}

func TestWriteTags(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, data testdata) error {
		tags := f.ReadTags()
		if len(tags) == 0 && !data.noTags {
			return fmt.Errorf("no tags")
		}

		if !f.WriteTags(tags) {
			return fmt.Errorf("could not write tags")
		}
		return nil
	})
}

func TestReadImage(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, data testdata) error {
		img, err := f.ReadImage()
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
			return fmt.Errorf("empty embedded art")
		}
		return nil
	})
}

func TestReadImageRaw(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, data testdata) error {
		b := f.ReadImageRaw()
		img, typ, err := image.Decode(b)
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if typ != "jpeg" {
			return fmt.Errorf("bad art type %q", typ)
		}

		if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
			return fmt.Errorf("empty embedded art")
		}
		return nil
	})
}

func TestWriteImage(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, _ testdata) error {
		img, err := f.ReadImage()
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if err = f.WriteImage(img, audiotags.JPEG); err != nil {
			return fmt.Errorf("writing image to file: %w", err)
		}
		return nil
	})
}

func TestWriteImageRaw(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, _ testdata) error {
		b := f.ReadImageRaw()
		img, typ, err := image.Decode(b)
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if typ != "jpeg" {
			return fmt.Errorf("bad art type %q", typ)
		}

		if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
			return fmt.Errorf("empty embedded art")
		}

		var out bytes.Buffer
		if err = jpeg.Encode(&out, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
			return fmt.Errorf("encoding jpeg image: %w", err)
		}

		size := img.Bounds().Size()
		if !f.WriteImageRaw(out.Bytes(), "image/jpeg", size.X, size.Y) {
			return fmt.Errorf("couldn't write raw image")
		}
		return nil
	})
}

func TestWriteImagePNG(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, _ testdata) error {
		img, err := f.ReadImage()
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		var out bytes.Buffer
		if err = png.Encode(&out, img); err != nil {
			return fmt.Errorf("encoding png image: %w", err)
		}

		img2, err := png.Decode(&out)
		if err != nil {
			return fmt.Errorf("decoding png image: %w", err)
		}

		if err = f.WriteImage(img2, audiotags.PNG); err != nil {
			return fmt.Errorf("writing png image to file: %w", err)
		}
		return nil
	})
}

func TestBadImageWrite(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, _ testdata) error {
		img, err := f.ReadImage()
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if err = f.WriteImage(img, audiotags.PNG+1); err == nil {
			return fmt.Errorf("expected error, got nil")
		}

		// img.Bounds().Intersect(image.Rect(0, 0, 0, 0))
		// if err = f.WriteImage(, audiotags.PNG); err == nil {
		// 	return fmt.Errorf("expected error, got nil")
		// }
		return nil
	})
}

func TestRemoveImages(t *testing.T) {
	withOpen(t, false, func(f *audiotags.File, _ testdata) error {
		if !f.RemoveImages() {
			return fmt.Errorf("could not remove embedded art")
		}
		return nil
	})
}
