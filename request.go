package redes_writer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olivere/elastic/v7"
)

type (
	Request struct {
		Type   string `json:"type"`
		Index  Index  `json:"index"`
		Update Update `json:"update"`
		Delete Delete `json:"delete"`
	}

	Index struct {
		Index           string      `json:"index"`
		Type            string      `json:"type"`
		Id              string      `json:"id"`
		Parent          string      `json:"parent"`
		Routing         string      `json:"routing"`
		Version         *int64      `json:"version,omitEmpty"` // default is MATCH_ANY
		VersionType     *string     `json:"version_type"`      // default is "internal"
		Doc             interface{} `json:"doc"`
		Pipeline        string      `json:"pipeline"`
		RetryOnConflict int         `json:"retry_on_conflict"`
	}

	Update struct {
		Index           string          `json:"index"`
		Type            string          `json:"type"`
		Id              string          `json:"id"`
		Parent          string          `json:"parent"`
		Routing         string          `json:"routing"`
		Version         *int64          `json:"version,omitEmpty"` // default is MATCH_ANY
		VersionType     *string         `json:"version_type"`      // default is "internal"
		DetectNoop      *bool           `json:"detect_noop"`
		Doc             interface{}     `json:"doc"`
		DocAsUpsert     *bool           `json:"doc_as_upsert"`
		Upsert          interface{}     `json:"upsert"`
		Script          *elastic.Script `json:"script"`
		RetryOnConflict *int            `json:"retry_on_conflict"`
		ScriptedUpsert  bool            `json:"scripted_upsert"`
	}

	Delete struct {
		Index       string  `json:"index"`
		Type        string  `json:"type"`
		Id          string  `json:"id"`
		Parent      string  `json:"parent"`
		Routing     string  `json:"routing"`
		Version     *int64  `json:"version,omitEmpty"` // default is MATCH_ANY
		VersionType *string `json:"version_type"`      // default is "internal"
	}
)

func (r Request) String() string {
	lines, err := r.Source()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return strings.Join(lines, "\n")
}

func (r Request) Source() ([]string, error) {
	switch r.Type {
	case "index":
		return toBulkIndex(r).Source()

	case "update":
		return toBulkUpdate(r).Source()

	case "delete":
		return toBulkDelete(r).Source()
	}

	return nil, fmt.Errorf("invalid request type")
}

func fromBytes(raw string) (*Request, error) {
	req := &Request{}
	err := json.Unmarshal([]byte(raw), &req)
	if nil != err {
		return nil, err
	}

	return req, nil
}

func toBulkIndex(w Request) *elastic.BulkIndexRequest {
	req := &w.Index

	b := elastic.NewBulkIndexRequest()
	b.Index(req.Index)
	b.Type(req.Type)
	b.Id(req.Id)
	b.Parent(req.Parent)
	b.Routing(req.Routing)

	if nil != req.Version {
		b.Version(*req.Version)
		b.VersionType(*req.VersionType)
	}

	b.OpType("index")
	b.Doc(req.Doc)
	b.Pipeline(req.Pipeline)
	b.RetryOnConflict(req.RetryOnConflict)

	return b
}

func toBulkUpdate(w Request) *elastic.BulkUpdateRequest {
	req := &w.Update

	b := elastic.NewBulkUpdateRequest()
	b.Index(req.Index)
	b.Type(req.Type)
	b.Id(req.Id)
	b.Parent(req.Parent)
	b.Routing(req.Routing)

	if nil != req.Version {
		b.Version(*req.Version)
		b.VersionType(*req.VersionType)
	}

	if nil != req.DetectNoop {
		b.DetectNoop(*req.DetectNoop)
	}

	b.Doc(req.Doc)
	if req.DocAsUpsert != nil {
		b.DocAsUpsert(*req.DocAsUpsert)
	}

	b.Upsert(req.Upsert)

	if req.Script != nil {
		b.Script(req.Script)
		b.ScriptedUpsert(req.ScriptedUpsert)
	}

	if req.RetryOnConflict != nil {
		b.RetryOnConflict(*req.RetryOnConflict)
	}

	return b
}

func toBulkDelete(w Request) *elastic.BulkDeleteRequest {
	req := &w.Delete

	b := elastic.NewBulkDeleteRequest()
	b.Index(req.Index)
	b.Type(req.Type)
	b.Id(req.Id)
	b.Parent(req.Parent)
	b.Routing(req.Routing)

	if nil != req.Version {
		b.Version(*req.Version)
		b.VersionType(*req.VersionType)
	}

	return b
}
