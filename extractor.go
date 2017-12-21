package savior

import "encoding/gob"

type ExtractorCheckpoint struct {
	SourceCheckpoint *SourceCheckpoint
	EntryIndex       int64
	Entry            *Entry
	Progress         float64
	Data             interface{}
}

type ExtractorResult struct {
	Entries []*Entry
}

type ResumeSupport int

const (
	// While the extractor exposes Save/Resume, in practice, resuming
	// will probably waste I/O and processing redoing a lot of work
	// that was already done, so it's not recommended to run it against
	// a networked resource
	ResumeSupportNone ResumeSupport = 0
	// The extractor can save/resume between each entry, but not in the middle of an entry
	ResumeSupportEntry ResumeSupport = 1
	// The extractor can save/resume within an entry, on a deflate/bzip2 block boundary for example
	ResumeSupportBlock ResumeSupport = 2
)

func (rs ResumeSupport) String() string {
	switch rs {
	case ResumeSupportNone:
		return "none"
	case ResumeSupportEntry:
		return "entry"
	case ResumeSupportBlock:
		return "block"
	default:
		return "unknown resume support"
	}
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

type ProgressListener func(progress float64)

func NopProgressListener() ProgressListener {
	return func(progress float64) {
		// muffin
	}
}

type Extractor interface {
	SetSaveConsumer(saveConsumer SaveConsumer)
	SetProgressListener(progressListener ProgressListener)
	Resume(checkpoint *ExtractorCheckpoint) (*ExtractorResult, error)
	ResumeSupport() ResumeSupport
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
