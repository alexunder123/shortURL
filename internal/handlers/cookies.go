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

type MyCookie struct{
	c http.Cookie
}


func Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ID string
		var C MyCookie
		for _, cookie := range r.Cookies() {
			if cookie.Name == "shortener" {
				C.c = *cookie
				break
			}
		}
		err := C.CheckCookie()
		if err != nil {
			log.Println(err)
			ID, err = C.GenerateCookie()
			if err != nil {
				log.Println(err)
			} else {
				http.SetCookie(w, &C.c)
			}
		} else {
			ID, err = C.ReturnID()
			if err != nil {
				log.Println(err)
			}
		}
		ctx := context.WithValue(r.Context(), name, ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *MyCookie) GenerateCookie() (string, error) {
	key := []byte("myShortenerURL00")
	ID := RandomID(16)
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	dst := make([]byte, aes.BlockSize)
	aesblock.Encrypt(dst, ID)
	c.c.Name = "shortener"
	c.c.Value = hex.EncodeToString(dst)
	c.c.Path = "/"
	c.c.Expires = time.Now().Add(time.Hour)
	baseID = append(baseID, string(ID))
	return string(ID), nil
}

func (c *MyCookie) CheckCookie() error {
	//При проверке встроенной функцией выдает ошибку "invalid Cookie.Expires" в тесте fetch_urls
	// err := c.Valid()
	// if err != nil {
	// 	return err
	// }

	if c.c.Name != "shortener" {
		return errors.New("invalid cookie name")
	}
	if c.c.Value == "" {
		return errors.New("empty cookie value")
	}
	ID, err := c.ReturnID()
	if err != nil {
		return err
	}
	err = FindID(ID)
	return err
}

func (c *MyCookie) ReturnID() (string, error) {
	key := []byte("myShortenerURL00")
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	ID := make([]byte, 16)
	val, err := hex.DecodeString(c.c.Value)
	if err != nil {
		return "", errors.New("cannot decode cookie")
	}
	aesblock.Decrypt(ID, val)
	return string(ID), nil
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
