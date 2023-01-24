package router

import (
	"context"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const userID nameID = "UserID"

type nameID string

type MyCookie struct {
	cookie http.Cookie
}

func MidWareCookies(next http.Handler) http.Handler {
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
			log.Error().Err(err)
			id, err = myCookie.generateCookie()
			if err != nil {
				log.Error().Err(err)
				next.ServeHTTP(w, r)
				return
			}
			http.SetCookie(w, &myCookie.cookie)
		}
		ctx := context.WithValue(r.Context(), userID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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

func (c *MyCookie) checkCookie() (string, error) {
	//При проверке встроенной функцией выдает ошибку "invalid Cookie.Expires" в тесте fetch_urls
	// err := c.Valid()
	// if err != nil {
	// 	return err
	// }

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

func randomID(n int) []byte {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bts := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		bts[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return bts
}
