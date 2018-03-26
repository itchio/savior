package seeksource_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/itchio/savior"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Uninitialized(t *testing.T) {
	{
		_, err := seeksource.FromBytes(nil).Read([]byte{})
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)
	}
	{
		_, err := seeksource.FromBytes(nil).ReadByte()
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)
	}
}

func Test_FromBytes(t *testing.T) {
	reference := semirandom.Bytes(16 * 1024)

	ss := seeksource.FromBytes(reference)
	assert.EqualValues(t, len(reference), ss.Size())

	_, err := ss.Resume(nil)
	must(t, err)

	out, err := ioutil.ReadAll(ss)
	must(t, err)

	assert.EqualValues(t, reference, out)
}

func Test_ReadByte(t *testing.T) {
	reference := semirandom.Bytes(256)

	ss := seeksource.FromBytes(reference)
	assert.EqualValues(t, len(reference), ss.Size())

	_, err := ss.Resume(nil)
	must(t, err)

	for i := 0; i < len(reference); i++ {
		b, err := ss.ReadByte()
		must(t, err)
		assert.EqualValues(t, reference[i], b)
	}
}

func Test_FromReadSeeker(t *testing.T) {
	reference := semirandom.Bytes(16 * 1024)

	ss := seeksource.NewWithSize(bytes.NewReader(reference), int64(len(reference)))
	assert.EqualValues(t, len(reference), ss.Size())

	_, err := ss.Resume(nil)
	must(t, err)

	out, err := ioutil.ReadAll(ss)
	must(t, err)

	assert.EqualValues(t, reference, out)
}

func Test_FromFile(t *testing.T) {
	f, err := ioutil.TempFile("", "seeksource-file")
	defer os.RemoveAll(f.Name())
	defer f.Close()

	reference := semirandom.Bytes(16 * 1024)
	_, err = f.Write(reference)
	must(t, err)

	ss := seeksource.FromFile(f)
	assert.EqualValues(t, len(reference), ss.Size())

	_, err = ss.Resume(nil)
	must(t, err)

	out, err := ioutil.ReadAll(ss)
	must(t, err)

	assert.EqualValues(t, reference, out)
}

func Test_SaveResume(t *testing.T) {
	reference := semirandom.Bytes(2 * 1024 * 1024)

	ss := seeksource.FromBytes(reference)
	_, err := ss.Resume(nil)
	must(t, err)

	var copied int
	buf := make([]byte, 128*1204)

	var checkpoint *savior.SourceCheckpoint
	ss.SetSourceSaveConsumer(&savior.CallbackSourceSaveConsumer{
		OnSave: func(c *savior.SourceCheckpoint) error {
			checkpoint = c
			return nil
		},
	})

	askedForSave := false
	for {
		n, err := ss.Read(buf)
		if err == io.EOF {
			break
		}
		must(t, err)

		copied += n
		if !askedForSave && copied > 1*1024*1024 {
			askedForSave = true
			ss.WantSave()
		}
	}
	assert.NotNil(t, checkpoint)

	// try resuming from another seeksource
	ss2 := seeksource.FromBytes(reference)
	ss2Offset, err := ss2.Resume(checkpoint)
	assert.True(t, ss2Offset > 0)
	assert.True(t, ss2Offset < int64(len(reference)))
	assert.EqualValues(t, ss2Offset, ss2.Tell())

	assert.True(t, ss2.Progress() > 0.4)
	assert.True(t, ss2.Progress() < 0.6)

	{
		copied, err := io.Copy(ioutil.Discard, ss2)
		must(t, err)
		assert.EqualValues(t, int64(len(reference))-ss2Offset, copied)
	}

	// try resuming from the same seeksource
	ssOffset, err := ss.Resume(checkpoint)
	assert.EqualValues(t, ss2Offset, ssOffset)
	assert.EqualValues(t, ssOffset, ss.Tell())

	{
		copied, err := io.Copy(ioutil.Discard, ss)
		must(t, err)
		assert.EqualValues(t, int64(len(reference))-ssOffset, copied)
	}
}

func Test_Section(t *testing.T) {
	reference := semirandom.Bytes(1024)

	ss := seeksource.FromBytes(reference)

	var err error

	// valid calls

	_, err = ss.Section(0, 512)
	assert.NoError(t, err)

	_, err = ss.Section(512, 512)
	assert.NoError(t, err)

	// invalid calls

	_, err = ss.Section(-1, 512)
	assert.Error(t, err)

	_, err = ss.Section(0, -29)
	assert.Error(t, err)

	_, err = ss.Section(0, 1025)
	assert.Error(t, err)

	_, err = ss.Section(512, 513)
	assert.Error(t, err)

	{
		ss2, err := ss.Section(0, 512)
		must(t, err)

		_, err = ss2.Resume(nil)
		must(t, err)

		out, err := ioutil.ReadAll(ss2)
		must(t, err)

		assert.EqualValues(t, reference[:512], out)
	}

	{
		ss2, err := ss.Section(256, 512)
		must(t, err)

		_, err = ss2.Resume(nil)
		must(t, err)

		out, err := ioutil.ReadAll(ss2)
		must(t, err)

		assert.EqualValues(t, reference[256:256+512], out)
	}

	{
		var sectionStart int64 = 171
		var sectionSize int64 = 700

		ss2, err := ss.Section(sectionStart, sectionSize)
		must(t, err)

		_, err = ss2.Resume(nil)
		must(t, err)

		numCheckpoints := 0

		ss2.SetSourceSaveConsumer(&savior.CallbackSourceSaveConsumer{
			OnSave: func(c *savior.SourceCheckpoint) error {
				numCheckpoints++
				_, err = ss2.Resume(c)
				return err
			},
		})

		buf := new(bytes.Buffer)

		for {
			_, err = io.CopyN(buf, ss2, 39)
			if err == io.EOF {
				return
			}
			must(t, err)

			ss2.WantSave()
		}

		refslice := reference[sectionStart : sectionStart+sectionSize]
		assert.EqualValues(t, refslice, buf.Bytes())
		assert.True(t, numCheckpoints > 0)
	}
}

func Test_InvalidCheckpoint(t *testing.T) {
	reference := semirandom.Bytes(1024)
	ss := seeksource.FromBytes(reference)

	{
		_, err := ss.Resume(&savior.SourceCheckpoint{
			Offset: -1,
		})
		assert.Error(t, err)
	}

	{
		_, err := ss.Resume(&savior.SourceCheckpoint{
			Offset: 1025,
		})
		assert.NoError(t, err)

		_, err = ss.ReadByte()
		assert.Error(t, err)
	}

	{
		ss2, err := ss.Section(512, 512)
		must(t, err)

		_, err = ss2.Resume(&savior.SourceCheckpoint{
			Offset: 512,
		})
		assert.NoError(t, err)

		_, err = ss2.Resume(&savior.SourceCheckpoint{
			Offset: -1,
		})
		assert.Error(t, err)

		_, err = ss2.Resume(&savior.SourceCheckpoint{
			Offset: 513,
		})
		assert.NoError(t, err)

		_, err = ss.ReadByte()
		assert.Error(t, err)
	}
}

func must(t *testing.T, err error) {
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
}
