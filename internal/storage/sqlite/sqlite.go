package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	Db *sql.DB
}

// todo - сделать миграции
func New(path string) (*Storage, error) {
	// имя для логирования
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("%s, open:, %s", op, err.Error())
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url (
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS ids_alias ON url(alias);
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s, prepare:, %s", op, err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s, exec:, %s", op, err.Error())
	}

	return &Storage{Db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	smtp, err := s.Db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s, prepare: %s", op, err.Error())
	}

	res, err := smtp.Exec(urlToSave, alias)
	if err != nil {
		// проверка на ошибку ConstraintUnique
		// если проверяем алиас с тем url который был ранее сохранен то возвращаем общую ошибку
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s %s", op, storage.ErrUrlExists)
		}
		return 0, fmt.Errorf("%s, exec: %s", op, err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	smtp, err := s.Db.Prepare(`SELECT url FROM url WHERE alias = ?`)
	if err != nil {
		return "", fmt.Errorf("%s, prepare: %s", op, err.Error())
	}

	var resUrl string

	if err = smtp.QueryRow(alias).Scan(resUrl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUrlNotFound
		}
		return "", fmt.Errorf("%s, query row: %s", op, err.Error())
	}

	return resUrl, nil
}

// TODO: implement method
func (s *Storage) DeleteURL(alias string) (string, error) {
	return "", nil
}
