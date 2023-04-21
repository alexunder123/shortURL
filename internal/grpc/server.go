package grpc

import (
	"context"
	"errors"
	"net"
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
	Subnet    net.IPNet
}

// NewShortURLsServer генерирует структуру для gRPC сервера.
func NewShortURLsServer(cfg *config.Config, strg storage.Storager, wrkr *worker.Worker) *ShortURLsServer {
	s := ShortURLsServer{
		cfg:       cfg,
		strg:      strg,
		workerDel: wrkr,
	}
	if cfg.TrustedSubnet != "" {
		_, subnet, _ := net.ParseCIDR(cfg.TrustedSubnet)
		s.Subnet = *subnet
	}
	return &s
}

// AddShortURL метод принимает от пользователя и возвращает адрес на сокращение.
func (s *ShortURLsServer) AddShortURL(ctx context.Context, in *pb.NewURLRequest) (*pb.NewURLResponce, error) {
	if in.UserID == "" {
		log.Error().Msgf("AddShortURL userID empty")
		return nil, storage.ErrUnauthorized
	}
	_, err := url.Parse(in.Entry)
	if err != nil {
		log.Error().Err(err).Msg("AddShortURL url.Parse err")
		return nil, storage.ErrBadRequest
	}
	newAddr, err := s.strg.SetShortURL(in.Entry, in.UserID, s.cfg)
	var response pb.NewURLResponce
	if errors.Is(err, storage.ErrConflict) {
		response.Responce = newAddr
		return &response, storage.ErrConflict
	}
	if err != nil {
		log.Error().Err(err).Msg("AddShortURL storage err")
		return nil, storage.ErrInternalError
	}
	response.Responce = newAddr
	return &response, nil
}

// AddBatchShortURL метод принимает от пользователя и возвращает в JSON список адресов на сокращение.
func (s *ShortURLsServer) AddBatchShortURL(ctx context.Context, in *pb.NewBatchRequest) (*pb.NewBatchResponce, error) {
	if in.UserID == "" {
		log.Error().Msgf("AddBatchShortURL userID empty")
		return nil, storage.ErrUnauthorized
	}
	if len(in.Request) == 0 {
		log.Error().Msgf("AddBatchShortURL incoming no content")
		return nil, storage.ErrNoContent
	}
	var batchURLs = make([]storage.MultiURL, len(in.Request), 0)
	for _, v := range in.Request {
		batchURLs = append(batchURLs, storage.MultiURL{CorrID: v.CorrID, OriginURL: v.OriginURL})
	}
	shortURLs, err := s.strg.WriteMultiURL(batchURLs, in.UserID, s.cfg)
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Err(err).Msg("AddBatchShortURL json error")
		return nil, storage.ErrUnsupported
	}
	if err != nil {
		log.Error().Err(err).Msg("AddBatchShortURL storage err")
		return nil, storage.ErrInternalError
	}
	var response pb.NewBatchResponce
	for _, v := range shortURLs {
		response.Responce = append(response.Responce, &pb.NewBatchResponce_Responce{CorrID: v.CorrID, ShortURL: v.ShortURL})
	}
	return &response, nil
}

// ReturnURL метод возвращает пользователю исходный адрес.
func (s *ShortURLsServer) ReturnURL(ctx context.Context, in *pb.ShortURLRequest) (*pb.FullURLResponce, error) {
	address, err := s.strg.RetFullURL(in.ShortURL)
	if errors.Is(err, storage.ErrGone) {
		log.Error().Err(err).Msg("ReturnURL address deleted")
		return nil, storage.ErrGone
	}
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("ReturnURL address not found")
		return nil, storage.ErrNoContent
	}
	if err != nil {
		log.Error().Err(err).Msg("ReturnURL storage err")
		return nil, storage.ErrInternalError
	}
	var response pb.FullURLResponce
	response.FullURL = address
	return &response, nil
}

// ReturnURL метод возвращает пользователю список сокращенных им адресов.
func (s *ShortURLsServer) ReturnUserURLs(ctx context.Context, in *pb.UserIDRequest) (*pb.AllUserURLsResponce, error) {
	if in.UserID == "" {
		log.Error().Msgf("AddBatchShortURL userID empty")
		return nil, storage.ErrUnauthorized
	}
	urls, err := s.strg.ReturnAllURLs(in.UserID, s.cfg)
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("ReturnURL address not found")
		return nil, storage.ErrNoContent
	}
	if err != nil {
		log.Error().Err(err).Msg("ReturnURL storage err")
		return nil, storage.ErrInternalError
	}
	var response pb.AllUserURLsResponce
	for _, v := range urls {
		response.Responce = append(response.Responce, &pb.AllUserURLsResponce_Responce{ShortURL: v.ShortURL, OriginalURL: v.OriginalURL})
	}
	return &response, nil
}

// ReturnStats метод возвращает количество сокращенных URL и пользователей в сервисе.
func (s *ShortURLsServer) ReturnStats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsResponce, error) {
	if s.cfg.TrustedSubnet == "" {
		log.Error().Msgf("ReturnStats TrustedSubnet isn't determined")
		return nil, storage.ErrForbidden
	}
	userIP := net.ParseIP(in.UserIP)
	if userIP == nil {
		log.Error().Msgf("ReturnStats User IP-address not resolved")
		return nil, storage.ErrBadRequest
	}
	if !s.Subnet.Contains(userIP) {
		log.Error().Msgf("ReturnStats User IP-address isn't CIDR subnet")
		return nil, storage.ErrForbidden
	}
	stats, err := s.strg.ReturnStats()
	if err != nil {
		log.Error().Err(err).Msg("ReturnStats storage err")
		return nil, storage.ErrInternalError
	}
	response := pb.StatsResponce{URLs: int32(stats.URLs), Users: int32(stats.Users)}
	return &response, nil
}

// PingDB метод возвращает статус наличия соединения с базой данных.
func (s *ShortURLsServer) PingDB(ctx context.Context, in *pb.PingRequest) (*pb.StatusResponce, error) {
	var response pb.StatusResponce
	err := s.strg.CheckPing(s.cfg)
	if err != nil {
		log.Error().Err(err).Msg("PingDB DB error")
		return nil, storage.ErrInternalError
	}
	response.RequestStatus = "StatusOK"
	return &response, nil
}

// PingDB метод возвращает статус наличия соединения с базой данных.
func (s *ShortURLsServer) MarkToDelete(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.StatusResponce, error) {
	if in.UserID == "" {
		log.Error().Msgf("MarkToDelete userID empty")
		return nil, storage.ErrUnauthorized
	}
	err := s.workerDel.Add(in.ToDelete, in.UserID)
	if errors.Is(err, storage.ErrUnavailable) {
		log.Error().Msgf("MarkToDelete the server is in the process of stopping")
		return nil, storage.ErrUnavailable
	}
	var response pb.StatusResponce
	response.RequestStatus = "StatusAccepted"
	return &response, nil
}
