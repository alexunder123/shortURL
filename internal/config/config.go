// Модуль считывает и хранит параметры конфигурации сервиса.
package config

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	mrand "math/rand"

	"github.com/caarlos0/env/v6"
)

// SaveMethod - определяем тип данных для выбора места хранения данных в зависимости от полученных параметров.
type SaveMethod int

// Определяем константы для выбора хранилища данных.
const (
	SaveMemory SaveMethod = iota
	SaveFile
	SaveSQL
)

// Config хранит основные параметры конфигурации сервиса.
type Config struct {
	ServerAddress         string        `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL               string        `env:"BASE_URL" json:"base_url"`
	FileStoragePath       string        `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN           string        `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS           bool          `env:"ENABLE_HTTPS" json:"enable_https"`
	Config                string        `env:"CONFIG" json:"-"`
	SavePlace             SaveMethod    `json:"-"`
	DeletingBufferSize    int           `json:"-"`
	DeletingBufferTimeout time.Duration `json:"-"`
}

// NewConfig считывает основные параметры и генерирует структуру Config.
func NewConfig() (*Config, error) {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "Адрес запускаемого сервера")
	}
	if config.BaseURL == "" {
		flag.StringVar(&config.BaseURL, "b", "http://127.0.0.1:8080", "Базовый адрес результирующего URL")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", "", "Файловое хранилище URL")
	}
	if config.DatabaseDSN == "" {
		flag.StringVar(&config.DatabaseDSN, "d", "", "База данных SQL")
	}
	if !config.EnableHTTPS {
		flag.BoolVar(&config.EnableHTTPS, "s", false, "Вариант запуска HTTPS сервера")
	}
	if config.Config == "" {
		flag.StringVar(&config.Config, "c,config", "", "Файл конфигурации")
	}
	flag.Parse()

	if config.Config != "" {
		err = ReadConfigFile(&config)
		if err != nil {
			return nil, err
		}
	}

	if config.DatabaseDSN != "" {
		config.SavePlace = SaveSQL
	} else if config.FileStoragePath != "" {
		config.SavePlace = SaveFile
	}

	config.DeletingBufferSize = 10
	config.DeletingBufferTimeout = 100 * time.Millisecond

	return &config, nil
}

// NewSertificate генерирует сертификат и приватный ключ для запуска HTTPS сервера.
func NewSertificate(cnfg *Config) (string, string, error) {
	certDir := "../../temp/cert.pem"
	pKeyDir := "../../temp/private_key.pem"
	mrand.Seed(1)
	sNO := int64(1000 + mrand.Intn(9000))
	strIP, _, _ := strings.Cut(cnfg.ServerAddress, ":")
	sliceIP := strings.Split(strIP, ".")
	bytesIP := make([]byte, 0, 4)
	for _, v := range sliceIP {
		i, err := strconv.Atoi(v)
		if err != nil {
			return "", "", err
		}
		bytesIP = append(bytesIP, byte(i))
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(sNO),
		Subject: pkix.Name{
			Organization: []string{"ShortURL"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(bytesIP[0], bytesIP[1], bytesIP[2], bytesIP[3]), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 1),
		SubjectKeyId: []byte{2, 4, 3, 4, 1},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	file, err := os.OpenFile(certDir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return "", "", err
	}
	_, err = file.WriteString(certPEM.String())
	if err != nil {
		return "", "", err
	}
	err = file.Close()
	if err != nil {
		return "", "", err
	}

	file, err = os.OpenFile(pKeyDir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return "", "", err
	}
	_, err = file.WriteString(privateKeyPEM.String())
	if err != nil {
		return "", "", err
	}
	err = file.Close()
	if err != nil {
		return "", "", err
	}

	return certDir, pKeyDir, nil
}

// ReadConfigFile считывает указанный файл конфигурации сервера.
// Значения из файла конфигурации применяются в последнюю очередь.
func ReadConfigFile(config *Config) error {
	file, err := os.Open(config.Config)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	bytes := make([]byte, stat.Size())
	r := bufio.NewReader(file)
	_, err = r.Read(bytes)
	if err != nil {
		return err
	}
	var fileConf Config
	err = json.Unmarshal(bytes, &fileConf)
	if err != nil {
		return err
	}
	if config.ServerAddress == "" {
		config.ServerAddress = fileConf.ServerAddress
	}
	if config.BaseURL == "" {
		config.BaseURL = fileConf.BaseURL
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = fileConf.FileStoragePath
	}
	if config.DatabaseDSN == "" {
		config.DatabaseDSN = fileConf.DatabaseDSN
	}
	if !config.EnableHTTPS {
		config.EnableHTTPS = fileConf.EnableHTTPS
	}
	return nil
}
