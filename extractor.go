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
	ShouldSave() bool
	Save(checkpoint *ExtractorCheckpoint) AfterSaveAction
}

type Extractor interface {
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

func (nsc *nopSaveConsumer) ShouldSave() bool {
	return false
}

func (nsc *nopSaveConsumer) Save(checkpoint *ExtractorCheckpoint) AfterSaveAction {
	return AfterSaveContinue
}
