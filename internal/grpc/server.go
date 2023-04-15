package grpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
	pb "shortURL/internal/grpc/proto"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

// ShortURLsServer поддерживает все необходимые методы сервера.
type ShortURLsServer struct {
	pb.ShortURLsServerServer
	cfg       *config.Config
	strg      storage.Storager
	workerDel *worker.Worker
}

// NewShortURLsServer генерирует структуру для gRPC сервера.
func NewShortURLsServer(cfg *config.Config, strg storage.Storager, wrkr *worker.Worker) *ShortURLsServer {
	return &ShortURLsServer{
		cfg:       cfg,
		strg:      strg,
		workerDel: wrkr,
	}
}

// AddShortURL метод принимает от пользователя и возвращает адрес на сокращение.
func (s *ShortURLsServer) AddShortURL(ctx context.Context, in *pb.NewURLEntry) (*pb.NewURLResponce, error) {
	var response pb.NewURLResponce
	if in.UserID == "" {
		log.Error().Msgf("AddShortURL userID empty")
		response.RequestStatus = http.StatusUnauthorized
		return &response, storage.ErrUnauthorized
	}
	_, err := url.Parse(in.Entry)
	if err != nil {
		log.Error().Err(err).Msg("AddShortURL url.Parse err")
		response.RequestStatus = http.StatusBadRequest
		return &response, storage.ErrBadRequest
	}
	newAddr, err := s.strg.SetShortURL(in.Entry, in.UserID, s.cfg)
	if errors.Is(err, storage.ErrConflict) {
		response.Responce = newAddr
		response.RequestStatus = http.StatusConflict
		return &response, storage.ErrConflict
	}
	if err != nil {
		log.Error().Err(err).Msg("AddShortURL storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.Responce = newAddr
	response.RequestStatus = http.StatusCreated
	return &response, nil
}

// AddJSONShortURL метод принимает от пользователя и возвращает в JSON адрес на сокращение.
func (s *ShortURLsServer) AddJSONShortURL(ctx context.Context, in *pb.NewJSONEntry) (*pb.NewJSONResponce, error) {
	var response pb.NewJSONResponce
	if in.UserID == "" {
		log.Error().Msgf("AddJSONShortURL userID empty")
		response.RequestStatus = http.StatusUnauthorized
		return &response, storage.ErrUnauthorized
	}
	newAddrBZ, err := s.strg.SetShortURLjs(in.Entry, in.UserID, s.cfg)
	if errors.Is(err, storage.ErrBadRequest) {
		log.Error().Err(err).Msg("AddJSONShortURL url.Parse err")
		response.RequestStatus = http.StatusBadRequest
		return &response, storage.ErrBadRequest
	}
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Err(err).Msg("AddJSONShortURL json error")
		response.RequestStatus = http.StatusUnsupportedMediaType
		return &response, storage.ErrUnsupported
	}
	if errors.Is(err, storage.ErrConflict) {
		response.Responce = newAddrBZ
		response.RequestStatus = http.StatusConflict
		return &response, storage.ErrConflict
	}
	if err != nil {
		log.Error().Err(err).Msg("AddJSONShortURL storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.Responce = newAddrBZ
	response.RequestStatus = http.StatusCreated
	return &response, nil
}

// AddBatchShortURL метод принимает от пользователя и возвращает в JSON список адресов на сокращение.
func (s *ShortURLsServer) AddBatchShortURL(ctx context.Context, in *pb.NewJSONEntry) (*pb.NewJSONResponce, error) {
	var response pb.NewJSONResponce
	if in.UserID == "" {
		log.Error().Msgf("AddBatchShortURL userID empty")
		response.RequestStatus = http.StatusUnauthorized
		return &response, storage.ErrUnauthorized
	}
	batchURLsBZ, err := s.strg.WriteMultiURL(in.Entry, in.UserID, s.cfg)
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Err(err).Msg("AddBatchShortURL json error")
		response.RequestStatus = http.StatusUnsupportedMediaType
		return &response, storage.ErrUnsupported
	}
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("AddBatchShortURL json no content")
		response.RequestStatus = http.StatusNoContent
		return &response, storage.ErrNoContent
	}
	if err != nil {
		log.Error().Err(err).Msg("AddBatchShortURL storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.Responce = batchURLsBZ
	response.RequestStatus = http.StatusCreated
	return &response, nil
}

// ReturnURL метод возвращает пользователю исходный адрес.
func (s *ShortURLsServer) ReturnURL(ctx context.Context, in *pb.ShortURL) (*pb.FullURL, error) {
	var response pb.FullURL
	address, err := s.strg.RetFullURL(in.ShortURL)
	if errors.Is(err, storage.ErrGone) {
		log.Error().Err(err).Msg("ReturnURL address deleted")
		response.RequestStatus = http.StatusGone
		return &response, storage.ErrGone
	}
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("ReturnURL address not found")
		response.RequestStatus = http.StatusBadRequest
		return &response, storage.ErrNoContent
	}
	if err != nil {
		log.Error().Err(err).Msg("ReturnURL storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.FullURL = address
	response.RequestStatus = http.StatusTemporaryRedirect
	return &response, nil
}

// ReturnURL метод возвращает пользователю список сокращенных им адресов.
func (s *ShortURLsServer) ReturnUserURLs(ctx context.Context, in *pb.UserID) (*pb.AllUserURLs, error) {
	var response pb.AllUserURLs
	if in.UserID == "" {
		log.Error().Msgf("AddBatchShortURL userID empty")
		response.RequestStatus = http.StatusUnauthorized
		return &response, storage.ErrUnauthorized
	}
	urlsBZ, err := s.strg.ReturnAllURLs(in.UserID, s.cfg)
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("ReturnURL address not found")
		response.RequestStatus = http.StatusBadRequest
		return &response, storage.ErrNoContent
	}
	if err != nil {
		log.Error().Err(err).Msg("ReturnURL storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.AllURLs = urlsBZ
	response.RequestStatus = http.StatusOK
	return &response, nil
}

// ReturnStats метод возвращает количество сокращенных URL и пользователей в сервисе.
func (s *ShortURLsServer) ReturnStats(ctx context.Context, in *pb.Statsrequest) (*pb.StatsResponce, error) {
	var response pb.StatsResponce
	if s.cfg.TrustedSubnet == "" {
		log.Error().Msgf("ReturnStats TrustedSubnet isn't determined")
		response.RequestStatus = http.StatusForbidden
		return &response, storage.ErrForbidden
	}
	userIP := net.ParseIP(in.UserIP)
	if userIP == nil {
		log.Error().Msgf("ReturnStats User IP-address not resolved")
		response.RequestStatus = http.StatusBadRequest
		return &response, storage.ErrBadRequest
	}
	if !s.cfg.Subnet.Contains(userIP) {
		log.Error().Msgf("ReturnStats User IP-address isn't CIDR subnet")
		response.RequestStatus = http.StatusForbidden
		return &response, storage.ErrForbidden
	}
	statsBZ, err := s.strg.ReturnStats()
	if err != nil {
		log.Error().Err(err).Msg("ReturnStats storage err")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.Stats = statsBZ
	response.RequestStatus = http.StatusOK
	return &response, nil
}

// PingDB метод возвращает статус наличия соединения с базой данных.
func (s *ShortURLsServer) PingDB(ctx context.Context, in *pb.Ping) (*pb.Status, error) {
	var response pb.Status
	err := s.strg.CheckPing(s.cfg)
	if err != nil {
		log.Error().Err(err).Msg("PingDB DB error")
		response.RequestStatus = http.StatusInternalServerError
		return &response, storage.ErrInternalError
	}
	response.RequestStatus = http.StatusOK
	return &response, nil
}

// PingDB метод возвращает статус наличия соединения с базой данных.
func (s *ShortURLsServer) MarkToDelete(ctx context.Context, in *pb.DeleteURLs) (*pb.Status, error) {
	var response pb.Status
	if in.UserID == "" {
		log.Error().Msgf("MarkToDelete userID empty")
		response.RequestStatus = http.StatusUnauthorized
		return &response, storage.ErrUnauthorized
	}
	err := s.workerDel.Add(in.ToDelete, in.UserID)
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Msgf("MarkToDelete json error")
		response.RequestStatus = http.StatusUnsupportedMediaType
		return &response, storage.ErrUnsupported
	}
	if errors.Is(err, storage.ErrUnavailable) {
		log.Error().Msgf("MarkToDelete the server is in the process of stopping")
		response.RequestStatus = http.StatusServiceUnavailable
		return &response, storage.ErrUnavailable
	}
	response.RequestStatus = http.StatusOK
	return &response, nil
}
