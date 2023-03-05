package midware

import (
	"context"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type nameID string

// Константа для добавления контекста в запрос.
// Предназначена для последующего считывания ID пользователя.
const UserID nameID = "UserID"

// MyCookie - структура для создания куки, и передачи ее пользователю.
type MyCookie struct {
	cookie http.Cookie
}

// Cookies middleware функция проверяет наличие и добавляет куки пользователя.
func Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id string
		var myCookie MyCookie
		for _, cookie := range r.Cookies() {
			if cookie.Name == "shortener" {
				myCookie.cookie = *cookie
				break
			}
		}
		id, err := myCookie.checkCookie()
		if err != nil {
			log.Info().Err(err).Msg("MidWareCookies checkCookie err")
			id, err = myCookie.generateCookie()
			if err != nil {
				log.Error().Err(err).Msg("MidWareCookies generateCookie err")
				next.ServeHTTP(w, r)
				return
			}
			http.SetCookie(w, &myCookie.cookie)
		}
		ctx := context.WithValue(r.Context(), UserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateCookie метод генерирует новые куки.
func (c *MyCookie) generateCookie() (string, error) {
	key := []byte("myShortenerURL00")
	id := randomID(16)
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	dst := make([]byte, aes.BlockSize)
	aesblock.Encrypt(dst, id)
	c.cookie.Name = "shortener"
	c.cookie.Value = hex.EncodeToString(dst)
	c.cookie.Path = "/"
	c.cookie.Expires = time.Now().Add(time.Hour)
	return string(id), nil
}

// checkCookie метод проверяет куки на корректность.
func (c *MyCookie) checkCookie() (string, error) {
	if c.cookie.Name != "shortener" {
		return "", errors.New("invalid cookie name")
	}
	if c.cookie.Value == "" {
		return "", errors.New("empty cookie value")
	}
	id, err := c.returnID()
	if err != nil {
		return "", err
	}
	// err = findID(id)
	return id, nil
}

// returnID метод возвращает новый ID пользователя.
func (c *MyCookie) returnID() (string, error) {
	key := []byte("myShortenerURL00")
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	id := make([]byte, 16)
	val, err := hex.DecodeString(c.cookie.Value)
	if err != nil {
		return "", errors.New("cannot decode cookie")
	}
	aesblock.Decrypt(id, val)
	return string(id), nil
}

// randomID функция генерирует новый ID пользователя.
func randomID(n int) []byte {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bts := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		bts[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return bts
}
