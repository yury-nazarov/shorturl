package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	pb "github.com/yury-nazarov/shorturl/internal/app/grpcserver/proto"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/config"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ShorURLService struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortURLServer
	//pb.UnsafeShortURLServer
	DB db.Repository
	LC  service.LinkCompressor
	CFG config.Config

}

func checkContextValue(ctx context.Context, value string) (string, bool) {
	// Проверяем наличие метедаты
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		arr := md.Get(value)
		if len(arr) > 0 {
			return arr[0], true
		}
	}
	return "", false
}

// AddURL добавляет новый URL
func (s *ShorURLService) AddURL(ctx context.Context, in *pb.AddURLRequest) (*pb.AddURLResponse, error) {
	var response pb.AddURLResponse

	// Проверяем наличие токена
	token, ok := checkContextValue(ctx, "token")
	if !ok {
		return  &response, status.Errorf(codes.NotFound,"token not install")
	}

	shortURL := s.LC.SortURL(in.OriginURL)
	err := s.DB.Add(context.Background(), shortURL, in.OriginURL, token)
	if err != nil {
		return  &response, status.Errorf(codes.Internal,"can't add url: %s", err)
	}
	response.ShortURL = shortURL
	return &response, status.Errorf(codes.OK, "")
}

// GetURL вернет оригинальный URL по сокращенному
func (s *ShorURLService) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error)  {
	var response pb.GetURLResponse

	// Проверяем наличие токена
	token, ok := checkContextValue(ctx, "token")
	if !ok {
		return  &response, status.Errorf(codes.NotFound,"token not install")
	}

	originURL, err := s.DB.Get(ctx, in.ShortURL, token)
	if err != nil {
		return &response, status.Errorf(codes.Internal, "can't get url: %s", err)
	}
	response.OriginURL = originURL
	return &response, status.Errorf(codes.OK, "success get url")
}

// GetAllUserURL вернет все URL пользователя
func (s *ShorURLService) GetAllUserURL(ctx context.Context, in *pb.GetAllUserURLRequest) (*pb.GetAllUserURLResponse, error) {
	var response pb.GetAllUserURLResponse

	// Проверяем наличие токена
	token, ok := checkContextValue(ctx, "token")
	if !ok {
		return  &response, status.Errorf(codes.NotFound,"token not install")
	}

	// Получаем все записи из БД
	userURL, err := s.DB.GetUserURL(ctx, token)
	if err != nil {
		return &response, status.Errorf(codes.NotFound,"content not found: %s", err)
	}

	for _, url := range userURL {
		response.Url = append(response.Url, fmt.Sprintf("%s,%s", url.OriginURL, url.ShortURL))
	}
	return &response, status.Error(codes.OK, "success get all url")
}

// DeleteURL помечает URL удаленными
func (s *ShorURLService) DeleteURL(ctx context.Context, in *pb.DeleteURLRequest) (*pb.DeleteURLResponse, error) {
	var response pb.DeleteURLResponse

	// Проверяем наличие токена
	token, ok := checkContextValue(ctx, "token")
	if !ok {
		return  &response, status.Errorf(codes.NotFound,"token not install")
	}

	// Получаем id записей которые нужно пометить удаленными
	urlsID := make(chan int, len(in.Urls))

	var wg sync.WaitGroup
	for _, identity := range in.Urls {
		wg.Add(1)
		go func(identity string) {
			id := s.DB.GetShortURLByIdentityPath(ctx, identity, token)
			urlsID <- id
			wg.Done()
		}(identity)
	}
	// Закрываем канал когда он заполнился
	wg.Wait()
	close(urlsID)

	// Помечаем удаленными пачку записей
	if err := s.DB.URLBulkDelete(ctx, urlsID); err != nil {
		return &response, err
	}
	return &response, status.Error(codes.OK, "success delete")
}

// Stats - Вернет кол-во сокращенных URL и кол-во пользователей в сервис
func (s *ShorURLService) Stats(ctx context.Context, in *pb.StatsURLRequest) (*pb.StatsURLResponse, error) {
	var response pb.StatsURLResponse

	// Проверяем src ip
	ipAddr, ok := checkContextValue(ctx, "X-Real-IP")
	if !ok {
		return  &response, status.Errorf(codes.PermissionDenied,"assess forbidden")
	}
	clientIP := net.ParseIP(ipAddr)
	// Парсим довереную IP подсеть из конфига
	_, trustNet, err := net.ParseCIDR(s.CFG.TrustedSubnet)
	if err != nil {
		return  &response, status.Errorf(codes.PermissionDenied,"assess forbidden")
	}
	// Проверяем что clientIP удалось получить и адрес входит в довереную подсеть
	if clientIP == nil || !trustNet.Contains(clientIP) {
		return  &response, status.Errorf(codes.PermissionDenied,"assess forbidden")
	}


	// Выполняем запрос
	stats, err := s.DB.Stats(context.Background())
	if err != nil {
		return &response, status.Errorf(codes.Internal,"can't get stats from DB")
	}
	response.CountShortURL = int32(stats.URLs)
	response.CountUsers = int32(stats.Users)

	return &response, status.Errorf(codes.OK, "success")
}