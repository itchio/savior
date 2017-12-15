package savior

import "encoding/gob"

type ExtractorCheckpoint struct {
	SourceCheckpoint *SourceCheckpoint
	EntryIndex       int64
	Entry            *Entry
}

type ExtractorSaver interface {
	Save(c *ExtractorCheckpoint)
}

type ProgressFunc func(progress float64)

type ExtractorParams struct {
	Source         Source
	LastCheckpoint *ExtractorCheckpoint
	OnProgress     ProgressFunc
	Sink           *Sink
}

type Extractor interface {
	Configure(params *ExtractorParams) error
	Work() error
}

func init() {
	gob.Register(&ExtractorCheckpoint{})
}
