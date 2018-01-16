package countingsource_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/itchio/savior/countingsource"
	"github.com/itchio/savior/seeksource"
	"github.com/stretchr/testify/assert"
)

func Test_Callback(t *testing.T) {
	buf := make([]byte, 4*1024*1204)

	var numCalls int64
	var lastCount int64

	ss := seeksource.FromBytes(buf)
	cs := countingsource.New(ss, func(count int64) {
		assert.True(t, count >= lastCount, "count must always increase")
		lastCount = count
		numCalls++
	})

	_, err := cs.Resume(nil)
	assert.NoError(t, err)

	_, err = io.Copy(ioutil.Discard, cs)
	assert.NoError(t, err)

	assert.True(t, numCalls > 0, "progress must be called at least once")
}
