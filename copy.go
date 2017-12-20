package savior

import (
	"io"

	"github.com/go-errors/errors"
)

type MakeCheckpointFunc func() (*ExtractorCheckpoint, error)

type CopyResult struct {
	Action AfterSaveAction
}

type CopyParams struct {
	Src   Source
	Dst   io.Writer
	Entry *Entry

	SaveConsumer SaveConsumer

	MakeCheckpoint MakeCheckpointFunc
}

func CopyWithSaver(params *CopyParams) (*CopyResult, error) {
	buf := make([]byte, 8*1024)

	for {
		n, readErr := params.Src.Read(buf)

		m, err := params.Dst.Write(buf[:n])
		params.Entry.WriteOffset += int64(m)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}

		if readErr != nil {
			if readErr == io.EOF {
				// cool, we're done!
				return &CopyResult{
					Action: AfterSaveContinue,
				}, nil
			}
			return nil, errors.Wrap(err, 0)
		}

		if params.SaveConsumer != nil && params.SaveConsumer.ShouldSave() {
			checkpoint, err := params.MakeCheckpoint()
			if err != nil {
				return nil, errors.Wrap(err, 0)
			}

			action := params.SaveConsumer.Save(checkpoint)
			if action != AfterSaveContinue {
				return &CopyResult{
					Action: AfterSaveStop,
				}, nil
			}
		}
	}
}
