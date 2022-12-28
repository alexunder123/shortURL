package handlers

import (
	"context"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	baseID        = make([]string, 0)
	name   nameID = "UserID"
)

type nameID string

func Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nc bool = false
		var ID string
		var c http.Cookie
		var err error
		for _, cookie := range r.Cookies() {
			if cookie.Name == "shortener" {
				c = *cookie
				break
			}
		}
		if err != nil {
			ID, err = NewCookie(&c)
			if err != nil {
				log.Println(err)
			} else {
				nc = true
			}
		} else {
			err = CheckCookie(&c)
			if err != nil {
				log.Println(err)
				ID, err = NewCookie(&c)
				if err != nil {
					log.Println(err)
				} else {
					nc = true
				}
			} else {
				ID, err = ReturnID(&c)
				if err != nil {
					log.Println(err)
				}
			}

		}
		if nc {
			http.SetCookie(w, &c)
		}
		ctx := context.WithValue(r.Context(), name, ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewCookie(c *http.Cookie) (string, error) {
	key := []byte("myShortenerURL00")
	ID := RandomID(16)
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	dst := make([]byte, aes.BlockSize)
	aesblock.Encrypt(dst, ID)

	c.Name = "shortener"
	c.Value = hex.EncodeToString(dst)
	c.Path = "/"
	c.MaxAge = 300
	SaveID(string(ID))
	return string(ID), nil
}

func CheckCookie(c *http.Cookie) error {
	err := c.Valid()
	if err != nil {
		return err
	}
	ID, err := ReturnID(c)
	if err != nil {
		return err
	}
	err = FindID(ID)
	return err
}

func ReturnID(c *http.Cookie) (string, error) {
	key := []byte("myShortenerURL00")
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	ID := make([]byte, 16)
	val, err := hex.DecodeString(c.Value)
	if err != nil {
		return "", errors.New("cannot decode cookie")
	}
	log.Println(val)
	aesblock.Decrypt(ID, val)
	return string(ID), nil
}

func SaveID(ID string) {
	baseID = append(baseID, ID)
}

func FindID(ID string) error {
	for _, value := range baseID {
		if ID == value {
			return nil
		}
	}
	return errors.New("invalid ID")
}

func RandomID(n int) []byte {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bts := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		bts[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return bts
}
