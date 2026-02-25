package store

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type SQliteDB struct {
	db   *sql.DB
	path string
}

func NewSQLiteDB(path string) (*SQliteDB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return &SQliteDB{
		db:   db,
		path: path,
	}, nil
}

func (d *SQliteDB) Close() error {
	if d == nil || d.db == nil {
		return errors.New("db nil")
	}

	err := d.db.Close()
	d.db = nil
	return err
}

func (d *SQliteDB) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *SQliteDB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *SQliteDB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}
