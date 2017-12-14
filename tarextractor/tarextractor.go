package tarextractor

import (
	"errors"

	"github.com/itchio/savior"
)

type tarExtractor struct {
	source savior.Source
}

var _ savior.Extractor = (*tarExtractor)(nil)

func New() *tarExtractor {
	return &tarExtractor{}
}

func (te *tarExtractor) Configure(params *savior.ExtractorParams) error {
	te.source = params.Source
	return nil
}

func (te *tarExtractor) Work() error {
	return errors.New("stub")
}
