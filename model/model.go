package model

import (
	"encoding/base64"
	"github.com/google/uuid"
	"path/filepath"
	"time"
)

// ContextKey provides a typed key for context values.
type ContextKey string

// CtxDB is the context key for the datasource.
const CtxDB ContextKey = "db"

// Cursorable is the common interface for a record that may have a cursor that
// references its canonical position in the DB for the sake of "after" type
// queries.
type Cursorable interface {
	Cursor() string
}

func NewPhoto(date *time.Time, path string) *Photo {
	return &Photo{
		UUID:  uuid.New().String(),
		Path:  path,
		Name:  filepath.Base(path),
		Year:  date.Year(),
		Month: int(date.Month()),
		Date:  *date,
	}
}

// Photo tracks essential fields and adds helpers around photo records.
type Photo struct {
	UUID       string
	Path       string
	Name       string
	CursorData string `db:"cursor"`
	Year       int
	Month      int
	Date       time.Time
}

// Cursor returns the opaque cursor id for the record.
func (p *Photo) Cursor() string {
	return base64.StdEncoding.EncodeToString([]byte(p.CursorData))
}

type YearMonthBucket struct {
	Date       string
	Year       int
	Month      int
	TotalCount int `db:"total_count"`
	PhotoUuids []string
}
