package badge

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tsg-ut/tsgctf-checker/checker"
)

type Badger struct {
	db *sqlx.DB
}

func NewBadger(db *sqlx.DB) *Badger {
	return &Badger{db: db}
}

func (bd *Badger) GetBadge(chall_name string) (string, error) {
	results, err := checker.FetchResult(bd.db, chall_name, 1)
	if err != nil {
		return "", err
	}

	if len(results) != 1 {
		return "", fmt.Errorf("Status for %s not found.", chall_name)
	}
	result := results[0]

	return GetBadge(result.Name, result.Result, result.Timestamp)
}
