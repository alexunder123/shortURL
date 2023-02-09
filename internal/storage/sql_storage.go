package storage

import (
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
)

type SQLStorage struct {
	DB *sql.DB
}

func NewSQLStorager(cfg *config.Config) Storager {
	db, err := sql.Ope("pgx", P.SQL)
	if err != nil {
		log.Fatal().Err(err).Msg("OpenDB open sql error")
	}
	createDB(db)
	return db
	return &SQLStorage{
		DB: openDB(P),
	}
}

func (s *SQLStorage) SetShortURL(fURL, userID string, Params *config.Param) (string, error) {
	key := hashStr(fURL)

	result, err := s.DB.Exec("INSERT INTO Short_URLs(key, user_id, value, deleted) VALUES($1, $2, $3, false) ON CONFLICT ON CONSTRAINT unique_query DO NOTHING", key, userID, fURL)
	if err != nil {
		return "", err
	}
	changes, _ := result.RowsAffected()
	if changes == 0 {
		var oldkey string
		row, err := s.DB.Query("SELECT key FROM Short_URLs WHERE user_id = $1 AND value = $2", userID, fURL)
		if err != nil {
			return "", err
		}
		if err := row.Err(); err != nil {
			return "", err
		}
		defer row.Close()
		for row.Next() {
			err = row.Scan(&oldkey)
			if err != nil {
				return "", err
			}
			if oldkey != "" {
				return oldkey, ErrConflict
			}
		}
	}
	return key, nil
}

func (s *SQLStorage) RetFullURL(key string) (string, error) {
	var value string
	var deleted bool
	row := s.DB.QueryRow("SELECT value, deleted FROM Short_URLs WHERE key = $1", key)
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return "", ErrNoContent
	}
	if row.Err() != nil {
		return "", row.Err()
	}
	err := row.Scan(&value, &deleted)
	if err != nil {
		return "", err
	}
	if deleted {
		return "", ErrGone
	}

	return value, nil
}

func (s *SQLStorage) ReturnAllURLs(userID string, P *config.Param) ([]byte, error) {

	var allURLs = make([]urls, 0)
	rows, err := s.DB.Query("SELECT key, value FROM Short_URLs WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var nextURL urls
		var sURL string

		err = rows.Scan(&sURL, &nextURL.OriginalURL)
		if err != nil {
			return nil, err
		}
		nextURL.ShortURL = string(P.URL + "/" + sURL)
		allURLs = append(allURLs, nextURL)
	}
	if len(allURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(allURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *SQLStorage) CheckPing(P *config.Param) error {
	return s.DB.Ping()
}

func (s *SQLStorage) WriteMultiURL(m []MultiURL, userID string, P *config.Param) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO Short_URLs(key, user_id, value, deleted) VALUES($1, $2, $3, false)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for i, v := range m {
		Key := hashStr(v.OriginURL)
		if _, err = stmt.Exec(Key, userID, v.OriginURL); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatal().Msgf("update drivers: unable to rollback: %v", err)
			}
			return nil, err
		}
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + Key)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal().Msgf("update drivers: unable to commit: %v", err)
		return nil, err
	}
	return r, nil
}

func createDB(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS Short_URLs(key text, user_id text, value text, deleted boolean, CONSTRAINT unique_query UNIQUE (user_id, value));")
	if err != nil {
		log.Fatal().Err(err).Msg("CreateDB create table error")
	}
}

func (s *SQLStorage) CloseDB() {
	err := s.DB.Close()
	if err != nil {
		log.Error().Err(err).Msg("CloseDB DB closing err")
	}
	log.Info().Msg("db closed")
}

func (s *SQLStorage) MarkDeleted(keys []string, id string) {
	stmt, err := s.DB.Prepare("UPDATE Short_URLs SET deleted=true WHERE key=$1 AND user_id=$2")
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer stmt.Close()
	for _, key := range keys {
		if _, err = stmt.Exec(key, id); err != nil {
			log.Error().Err(err).Msg("MarkDeleted DB update err")
			return
		}
	}
}
