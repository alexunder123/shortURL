package storage

import (
	"database/sql"
	"encoding/json"
	"shortURL/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type SQLStorage struct {
	DB *sql.DB
}

func NewSQLStorager(P *config.Param) Storager {
	return &SQLStorage{
		DB: OpenDB(P),
	}
}

func (s *SQLStorage) SetShortURL(fURL, userID string, Params *config.Param) (string, error) {
	key := HashStr(fURL)

	result, err := s.DB.Exec("INSERT INTO ShortURL_GO12Alex(key, user_id, value) VALUES($1, $2, $3) ON CONFLICT ON CONSTRAINT unique_query DO NOTHING", key, userID, fURL)
	if err != nil {
		return "", err
	}
	changes, _ := result.RowsAffected()
	if changes == 0 {
		var oldkey string
		row, err := s.DB.Query("SELECT key FROM ShortURL_GO12Alex WHERE user_id = $1 AND value = $2", userID, fURL)
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
	row, err := s.DB.Query("SELECT value FROM ShortURL_GO12Alex WHERE key = $1", key)
	if err != nil {
		return "", err
	}
	if err := row.Err(); err != nil {
		return "", err
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&value)
		if err != nil {
			return "", err
		}
	}
	return value, nil
}

func (s *SQLStorage) ReturnAllURLs(userID string, P *config.Param) ([]byte, error) {

	var allURLs = make([]URLs, 0)
	rows, err := s.DB.Query("SELECT key, value FROM ShortURL_GO12Alex WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var nextURL URLs
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
	err := s.DB.Ping()
	return err
}

func (s *SQLStorage) WriteMultiURL(m []MultiURL, userID string, P *config.Param) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO ShortURL_GO12Alex(key, user_id, value) VALUES($1, $2, $3)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for i, v := range m {
		Key := HashStr(v.OriginURL)
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

func OpenDB(P *config.Param) *sql.DB {
	db, err := sql.Open("pgx", P.SQL)
	if err != nil {
		log.Fatal().Err(err)
	}
	CreateDB(db)
	return db
}

func CreateDB(db *sql.DB) {

	// Использовать при необходимости изменения структуры таблицы
	// _, err := db.Exec("DROP TABLE IF EXISTS ShortURL_GO12Alex;")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS ShortURL_GO12Alex(key text, user_id text, value text, CONSTRAINT unique_query UNIQUE (user_id, value));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS GO12Alex(key text, user_id text, value text, CONSTRAINT unique_query UNIQUE (user_id, value));")
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SQLStorage) CloseDB() {
	err := s.DB.Close()
	if err != nil {
		log.Error().Err(err)
	}
	log.Info().Msg("db closed")
}
