package savior

import "encoding/gob"

type ExtractorCheckpoint struct {
	SourceCheckpoint *SourceCheckpoint
	EntryIndex       int64
	Entry            *Entry
	Data             interface{}
}

type AfterSaveAction int

const (
	AfterSaveContinue AfterSaveAction = 1
	AfterSaveStop     AfterSaveAction = 2
)

type SaveConsumer interface {
	ShouldSave(copiedBytes int64) bool
	Save(checkpoint *ExtractorCheckpoint) (AfterSaveAction, error)
}

type Extractor interface {
	SetSaveConsumer(saveConsumer SaveConsumer)
	Resume(checkpoint *ExtractorCheckpoint) error
}

func init() {
	gob.Register(&ExtractorCheckpoint{})
}

type nopSaveConsumer struct{}

var _ SaveConsumer = (*nopSaveConsumer)(nil)

func NopSaveConsumer() SaveConsumer {
	return &nopSaveConsumer{}
}

func (nsc *nopSaveConsumer) ShouldSave(n int64) bool {
	return false
}

func (nsc *nopSaveConsumer) Save(checkpoint *ExtractorCheckpoint) (AfterSaveAction, error) {
	return AfterSaveContinue, nil
}
