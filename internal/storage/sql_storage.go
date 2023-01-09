package storage

import (
	"database/sql"
	"encoding/json"
	"log"
	"shortURL/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	DB *sql.DB
	StorageStruct
}

var ()

func NewSQLStorager(P *config.Param) Storager {
	DBs := OpenDB(P)
	defer DBs.Close()
	return &SQLStorage{
		DB: DBs,
		StorageStruct: StorageStruct{
			UserID: "",
			Key:    "",
			Value:  "",
		},
	}
}

func (s *SQLStorage) SetShortURL(fURL, UserID string, Params *config.Param) string {
	s.Key = HashStr(fURL)

	var isexist string
	row, err := s.DB.Query("SELECT key FROM GO12Alex WHERE key = $1", s.Key)
	if err != nil {
		log.Fatal(err)
	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&isexist)
		if err != nil {
			log.Fatal(err)
		}
		if isexist != "" {
			return s.Key
		}
	}
	_, err = s.DB.Exec("INSERT INTO GO12Alex(key, user_id, value) VALUES($1, $2, $3)", s.Key, UserID, fURL)
	if err != nil {
		log.Fatal(err)
	}
	return s.Key
}

func (s *SQLStorage) RetFullURL(key string) string {
	var value string
	row, err := s.DB.Query("SELECT value FROM GO12Alex WHERE key = $1", key)
	if err != nil {
		log.Fatal(err)
	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&value)
		if err != nil {
			log.Fatal(err)
		}
	}
	return value
}

func (s *SQLStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {

	var AllURLs = make([]URLs, 0)
	rows, err := s.DB.Query("SELECT key, value FROM GO12Alex WHERE user_id = $1", UserID)
	if err != nil {
		log.Fatal(err)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var nextURL URLs
		var sURL string

		err = rows.Scan(&sURL, &nextURL.OriginalURL)
		if err != nil {
			panic(err)
		}
		nextURL.ShortURL = string(P.URL + "/" + sURL)
		AllURLs = append(AllURLs, nextURL)
	}
	if len(AllURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(AllURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *SQLStorage) CheckPing(P *config.Param) error {
	err := s.DB.Ping()
	return err
}

func (s *SQLStorage) WriteMultiURL(m *[]MultiURL, UserID string, P *config.Param) (*[]MultiURL, error) {
	r := make([]MultiURL, len(*m))
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO GO12Alex(key, user_id, value) VALUES($1, $2, $3)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for _, v := range *m {
		Key := HashStr(v.OriginURL)
		if _, err = stmt.Exec(Key, UserID, v.OriginURL); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return nil, err
		}
		r = append(r, MultiURL{CorrID: v.CorrID, ShortURL: string(P.URL + "/" + Key)})
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
		return nil, err
	}
	return &r, nil
}

func OpenDB(P *config.Param) *sql.DB {
	db, err := sql.Open("pgx", P.SQL)
	if err != nil {
		log.Fatal(err)
	}
	CreateDB(db)
	return db
}

func CreateDB(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + `GO12Alex("key" text, "user_id" text, "value" text);`)
	if err != nil {
		log.Fatal(err)
	}
}
