package blevestore

import (
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

//go:generate mockgen -package mockblevestore -destination mockblevestore/mockblevestore.go  github.com/blevesearch/bleve Index

type store struct {
	idx bleve.Index
}

func newStore(path string) (*store, error) {
	var idx bleve.Index
	var err error
	if path == "" {
		// If no path provided, use in-mem store.
		idx, err = bleve.NewMemOnly(buildMpping())
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the path does not exist, create the index.
		idx, err = bleve.New(path, buildMpping())
	} else {
		// If the path exists, use the existing data.
		idx, err = bleve.Open(path)
	}

	if err != nil {
		return nil, err
	}
	return &store{idx}, nil
}

func buildMpping() *mapping.IndexMappingImpl {
	root := bleve.NewDocumentMapping()

	textFieldNotIndexed := bleve.NewTextFieldMapping()
	textFieldNotIndexed.Index = false

	textFieldNotStored := bleve.NewTextFieldMapping()
	textFieldNotStored.Store = false

	numFieldNotIndexed := bleve.NewNumericFieldMapping()
	numFieldNotIndexed.Index = false

	numFieldNotStored := bleve.NewTextFieldMapping()
	numFieldNotStored.Store = false

	// META
	meta := bleve.NewDocumentStaticMapping()
	meta.AddFieldMappingsAt("action", textFieldNotStored)
	meta.AddFieldMappingsAt("mapId", textFieldNotStored)
	meta.AddFieldMappingsAt("step", textFieldNotStored)
	meta.AddFieldMappingsAt("tags", textFieldNotStored)
	meta.AddFieldMappingsAt("prevLinkHash", textFieldNotStored)
	meta.AddFieldMappingsAt("prevLinkHash", numFieldNotStored)
	meta.AddFieldMappingsAt("priority", numFieldNotStored)

	// META.PROCESS
	process := bleve.NewDocumentStaticMapping()
	process.AddFieldMappingsAt("name", textFieldNotStored)
	process.AddFieldMappingsAt("state", textFieldNotStored)
	meta.AddSubDocumentMapping("process", process)

	root.AddSubDocumentMapping("meta", meta)

	// METADATA
	metadata := bleve.NewDocumentStaticMapping()
	metadata.AddFieldMappingsAt("createdById", textFieldNotStored)
	metadata.AddFieldMappingsAt("formId", textFieldNotStored)
	metadata.AddFieldMappingsAt("groupId", textFieldNotStored)
	metadata.AddFieldMappingsAt("inputs", textFieldNotStored)
	metadata.AddFieldMappingsAt("ownerId", textFieldNotStored)

	root.AddSubDocumentMapping("metadata", metadata)

	// DATA
	data := bleve.NewDocumentMapping()
	root.AddSubDocumentMapping("data", data)

	// RAW
	root.AddFieldMappingsAt("raw", numFieldNotIndexed)
	root.AddFieldMappingsAt("raw", textFieldNotIndexed)

	root.Dynamic = false

	mapping := bleve.NewIndexMapping()
	mapping.TypeField = "type"

	mapping.AddDocumentMapping("root", root)

	return mapping
}
