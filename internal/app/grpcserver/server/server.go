package server

import (
	"context"
	"fmt"
	pb "github.com/yury-nazarov/shorturl/internal/app/grpcserver/proto"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

type ShorURLService struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortURLServer
	//pb.UnsafeShortURLServer
	DB db.Repository
	LC  service.LinkCompressor

}

// AddURL добавляет новый URL
func (s *ShorURLService) AddURL(ctx context.Context, in *pb.AddURLRequest) (*pb.AddURLResponse, error) {
	var response pb.AddURLResponse

	shortURL := s.LC.SortURL(in.OriginURL)
	err := s.DB.Add(context.Background(), shortURL, in.OriginURL, in.UserToken)
	if err != nil {
		return  &response, status.Errorf(codes.Internal,"can't add url: %s", err)
	}
	response.ShortURL = shortURL
	return &response, status.Errorf(codes.OK, "")
}

// GetURL вернет оригинальный URL по сокращенному
func (s *ShorURLService) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error)  {
	var response pb.GetURLResponse

	originURL, err := s.DB.Get(ctx, in.ShortURL, in.UserToken)
	if err != nil {
		return &response, status.Errorf(codes.Internal, "can't get url: %s", err)
	}
	response.OriginURL = originURL
	return &response, status.Errorf(codes.OK, "success get url")
}

// GetAllUserURL вернет все URL пользователя
func (s *ShorURLService) GetAllUserURL(ctx context.Context, in *pb.GetAllUserURLRequest) (*pb.GetAllUserURLResponse, error) {
	var response pb.GetAllUserURLResponse

	// Получаем все записи из БД
	userURL, err := s.DB.GetUserURL(ctx, in.UserToken)
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

	// Получаем id записей которые нужно пометить удаленными
	urlsID := make(chan int, len(in.Urls))

	var wg sync.WaitGroup
	for _, identity := range in.Urls {
		wg.Add(1)
		go func(identity string) {
			id := s.DB.GetShortURLByIdentityPath(ctx, identity, in.UserToken)
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
	// TODO

	return &response, status.Errorf(codes.OK, "success")
}