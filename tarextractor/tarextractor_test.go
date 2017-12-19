package tarextractor_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/semirandom"

	"github.com/itchio/arkive/tar"
	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/tarextractor"
)

func must(t *testing.T, err error) {
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func TestTar(t *testing.T) {
	sink := checker.NewSink()
	rng := rand.New(rand.NewSource(0xf617a899))
	for i := 0; i < 100; i++ {
		if rng.Intn(100) < 10 {
			// ok, make a symlink
			name := fmt.Sprintf("symlink-%d", i)
			sink.Items[name] = &checker.Item{
				Entry: &savior.Entry{
					CanonicalPath: name,
					Kind:          savior.EntryKindDir,
				},
				Linkname: fmt.Sprintf("target-%d", i*2),
			}
		} else if rng.Intn(100) < 20 {
			// ok, make a dir
			name := fmt.Sprintf("dir-%d", i)
			sink.Items[name] = &checker.Item{
				Entry: &savior.Entry{
					CanonicalPath: name,
					Kind:          savior.EntryKindDir,
				},
			}
		} else {
			// ok, make a file
			size := rng.Int63n(512 * 1024)
			name := fmt.Sprintf("file-%d", i)
			sink.Items[name] = &checker.Item{
				Entry: &savior.Entry{
					CanonicalPath:    name,
					Kind:             savior.EntryKindFile,
					UncompressedSize: size,
				},
				Data: semirandom.Bytes(size),
			}
		}
	}

	tarBytes := makeTar(t, sink)
	source := seeksource.FromBytes(tarBytes)

	ex := tarextractor.New()
	err := ex.Configure(&savior.ExtractorParams{
		LastCheckpoint: nil,
		OnProgress:     nil,
		Source:         source,
		Sink:           sink,
	})
	must(t, err)

	err = ex.Work()
	must(t, err)
}

func makeTar(t *testing.T, sink *checker.Sink) []byte {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, item := range sink.Items {
		switch item.Entry.Kind {
		case savior.EntryKindDir:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}))
		case savior.EntryKindFile:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeReg,
				Size:     int64(len(item.Data)),
				Mode:     0644,
			}))

			_, err := tw.Write(item.Data)
			must(t, err)
		case savior.EntryKindSymlink:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeSymlink,
				Mode:     0644,
				Linkname: item.Linkname,
			}))
		}
	}

	err := tw.Close()
	must(t, err)

	return buf.Bytes()
}
