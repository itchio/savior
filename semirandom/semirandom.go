package semirandom

import (
	"bytes"
	"math/rand"
)

func Generate(length int) []byte {
	gen := rand.New(rand.NewSource(0xfaadbeef))
	inputBuf := new(bytes.Buffer)

	var oldSeqs [][]byte

	for inputBuf.Len() < length {
		var seq []byte

		if gen.Intn(100) >= 80 {
			// re-use old seq
			seq = oldSeqs[gen.Intn(len(oldSeqs))]
		} else {
			seqLength := gen.Intn(48 * 1024)
			seq := make([]byte, seqLength)
			for j := 0; j < seqLength; j++ {
				seq[j] = byte(gen.Intn(255))
			}
			oldSeqs = append(oldSeqs, seq)
		}

		numRepetitions := gen.Intn(24)
		for j := 0; j < numRepetitions; j++ {
			inputBuf.Write(seq)
		}
	}

	return inputBuf.Bytes()[:length]
}
