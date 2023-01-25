package storage

import (
	"database/sql"
	"encoding/json"
	"log"
	"shortURL/internal/config"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	DB *sql.DB
	StorageStruct
}

func NewSQLStorager(P *config.Param) Storager {
	return &SQLStorage{
		DB: OpenDB(P),
		StorageStruct: StorageStruct{
			UserID: "",
			Key:    "",
			Value:  "",
		},
	}
}

func (s *SQLStorage) SetShortURL(fURL, UserID string, Params *config.Param) (string, error) {
	key := HashStr(fURL)

	result, err := s.DB.Exec("INSERT INTO GO12Alex(key, user_id, value, deleted) VALUES($1, $2, $3, false) ON CONFLICT ON CONSTRAINT unique_query DO NOTHING", key, UserID, fURL)
	if err != nil {
		log.Fatal(err)
	}
	changes, _ := result.RowsAffected()
	if changes == 0 {
		var oldkey string
		row, err := s.DB.Query("SELECT key FROM GO12Alex WHERE user_id = $1 AND value = $2", UserID, fURL)
		if err != nil {
			log.Fatal(err)
		}
		if err := row.Err(); err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		for row.Next() {
			err = row.Scan(&oldkey)
			if err != nil {
				log.Fatal(err)
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
	row, err := s.DB.Query("SELECT value, deleted FROM GO12Alex WHERE key = $1", key)
	if err != nil {
		return "", err
	}
	if err := row.Err(); err != nil {
		return "", err
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&value, &deleted)
		if err != nil {
			return "", err
		}
	}
	if deleted {
		return "", ErrGone
	}
	return value, nil
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
	stmt, err := tx.Prepare("INSERT INTO GO12Alex(key, user_id, value, deleted) VALUES($1, $2, $3, false)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for i, v := range *m {
		Key := HashStr(v.OriginURL)
		if _, err = stmt.Exec(Key, UserID, v.OriginURL); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return nil, err
		}
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + Key)
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
	_, err := db.Exec("DROP TABLE IF EXISTS GO12Alex;")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS GO12Alex(key text, user_id text, value text, deleted boolean, CONSTRAINT unique_query UNIQUE (user_id, value));")
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SQLStorage) CloseDB() {
	err := s.DB.Close()
	if err != nil {
		log.Println(err)
	}
	log.Println("db closed")
}

func (s *SQLStorage) MarkDeleted(DeleteURLs *[]string, UserID string, P *config.Param) {
	inputCh := make(chan string)
	go func() {
		for _, del := range *DeleteURLs {
			inputCh <- del
		}
		close(inputCh)
	}()

	chQ := 5 //Количество каналов для работы
	//fan out
	fanOutChs := fanOut(inputCh, chQ)
	workerChs := make([]chan string, 0, chQ)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan string)
		s.newWorker(fanOutCh, workerCh, UserID)
		workerChs = append(workerChs, workerCh)
	}

	//fan in
	outCh := fanIn(workerChs)

	//update
	stmt, err := s.DB.Prepare("UPDATE GO12Alex SET deleted=true WHERE key=$1 AND user_id=$2")
	if err != nil {
		log.Println(err)
		return
	}
	defer stmt.Close()
	for key := range outCh {
		if _, err = stmt.Exec(key, UserID); err != nil {
			log.Println(err)
			return
		}
	}
}

func (s *SQLStorage) newWorker(in, out chan string, UserID string) {
	go func() {
		stmt, err := s.DB.Prepare("SELECT value FROM GO12Alex WHERE key = $1 AND user_id = $2")
		if err != nil {
			return
		}
		defer stmt.Close()
		for myURL := range in {
			key := strings.Trim(myURL, "\"")
			log.Println("key to delete:", key)
			row := stmt.QueryRow(key, UserID)
			if err := row.Err(); err != nil {
				log.Println(err)
				return
			}
			out <- key
		}
		close(out)
	}()
}
