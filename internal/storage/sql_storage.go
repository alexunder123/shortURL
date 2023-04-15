package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
)

// SQLStorage структура для создания хранилища базы данных.
type SQLStorage struct {
	DB *sql.DB
}

// NewSQLStorager метод генерирует хранилище данных.
func NewSQLStorager(cfg *config.Config) *SQLStorage {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("OpenDB open sql error")
	}
	err = createDB(db)
	if err != nil {
		log.Fatal().Err(err).Msg("CreateDB create table error")
	}
	return &SQLStorage{
		DB: db,
	}
}

// SetShortURL метод генерирует ключ для короткой ссылки, проверяет его наличие и сохраняет данные.
// Данные передаются и возвращаются текстом в теле запроса.
func (s *SQLStorage) SetShortURL(fURL, userID string, cfg *config.Config) (string, error) {
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
				return cfg.BaseURL + "/" + oldkey, ErrConflict
			}
		}
	}
	return cfg.BaseURL + "/" + key, nil
}

// SetShortURL метод генерирует ключ для короткой ссылки, проверяет его наличие и сохраняет данные.
// Данные передаются и возвращаются текстом в теле запроса.
func (s *SQLStorage) SetShortURLjs(bytes []byte, userID string, cfg *config.Config) ([]byte, error) {
	var addr postURL
	if err := json.Unmarshal(bytes, &addr); err != nil {
		return nil, ErrUnsupported
	}
	_, err := url.Parse(addr.GetURL)
	if err != nil {
		return nil, ErrBadRequest
	}
	key := hashStr(addr.GetURL)

	result, err := s.DB.Exec("INSERT INTO Short_URLs(key, user_id, value, deleted) VALUES($1, $2, $3, false) ON CONFLICT ON CONSTRAINT unique_query DO NOTHING",
		key, userID, addr.GetURL)
	if err != nil {
		return nil, err
	}
	changes, _ := result.RowsAffected()
	if changes == 0 {
		var oldkey string
		row, err := s.DB.Query("SELECT key FROM Short_URLs WHERE user_id = $1 AND value = $2", userID, addr.GetURL)
		if err != nil {
			return nil, err
		}
		if err := row.Err(); err != nil {
			return nil, err
		}
		defer row.Close()
		for row.Next() {
			err = row.Scan(&oldkey)
			if err != nil {
				return nil, err
			}
			if oldkey != "" {
				newAddr := postURL{SetURL: cfg.BaseURL + "/" + oldkey}
				newAddrBZ, err := json.Marshal(newAddr)
				if err != nil {
					return nil, err
				}
				return newAddrBZ, ErrConflict
			}
		}
	}
	newAddr := postURL{SetURL: cfg.BaseURL + "/" + key}
	newAddrBZ, err := json.Marshal(newAddr)
	if err != nil {
		return nil, err
	}
	return newAddrBZ, nil
}

// RetFullURL метод возвращает полный адрес по ключу от короткой ссылки.
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

// ReturnAllURLs метод возвращает список сокращенных адресов по ID пользователя.
func (s *SQLStorage) ReturnAllURLs(userID string, cfg *config.Config) ([]byte, error) {

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
		nextURL.ShortURL = string(cfg.BaseURL + "/" + sURL)
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

// CheckPing метод возвращает статус подключения к базе данных.
func (s *SQLStorage) CheckPing(cfg *config.Config) error {
	return s.DB.Ping()
}

// Метод обрабатывает, сохраняет и возвращает batch список сокращенных адресов.
func (s *SQLStorage) WriteMultiURL(bytes []byte, userID string, cfg *config.Config) ([]byte, error) {
	var m = make([]MultiURL, 0)
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, ErrUnsupported
	}
	if len(m) == 0 {
		return nil, ErrNoContent
	}
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
		r[i].ShortURL = string(cfg.BaseURL + "/" + Key)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal().Msgf("update drivers: unable to commit: %v", err)
		return nil, err
	}
	rBZ, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return rBZ, nil
}

// CloseDB метод закрывает соединение с хранилищем данных.
func (s *SQLStorage) CloseDB() {
	err := s.DB.Close()
	if err != nil {
		log.Error().Err(err).Msg("CloseDB DB closing err")
	}
	log.Info().Msg("db closed")
}

// MarkDeleted метод помечает на удаление адреса пользователя в хранилище.
func (s *SQLStorage) MarkDeleted(keys []string, ids []string) {
	stmt, err := s.DB.Prepare("UPDATE Short_URLs SET deleted=true WHERE key=$1 AND user_id=$2")
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer stmt.Close()
	for i, key := range keys {
		if _, err = stmt.Exec(key, ids[i]); err != nil {
			log.Error().Err(err).Msg("MarkDeleted DB update err")
			return
		}
	}
}

// ReturnStats метод возвращает статистику по количеству сохраненных сокращенных URL и пользователей.
func (s *SQLStorage) ReturnStats() ([]byte, error) {
	var urls, users int

	row := s.DB.QueryRow("SELECT count(*) FROM Short_URLs")
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return nil, ErrNoContent
	}
	if row.Err() != nil {
		return nil, row.Err()
	}
	err := row.Scan(&urls)
	if err != nil {
		return nil, err
	}

	row = s.DB.QueryRow("SELECT count(*) FROM Short_URLs GROUP BY user_id")
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return nil, ErrNoContent
	}
	if row.Err() != nil {
		return nil, row.Err()
	}
	err = row.Scan(&users)
	if err != nil {
		return nil, err
	}

	stats := stats{
		URLs:  urls,
		Users: users,
	}
	sb, err := json.Marshal(stats)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func createDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS Short_URLs(key text, user_id text, value text, deleted boolean, CONSTRAINT unique_query UNIQUE (user_id, value));")
	if err != nil {
		return err
	}
	return nil
}
