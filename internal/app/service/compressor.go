package service

import (
	"fmt"
	"log"
	"math/big"

	"github.com/yury-nazarov/shorturl/internal/config"

	"github.com/itchyny/base58-go"
	"github.com/sirupsen/logrus"
)

type LinkCompressor struct {
	urlLength   int
	ServiceName string
	logger 		*logrus.Logger
}

// NewLinkCompressor - объект содержит в себе все необходимое для подготови короткого URL
//func NewLinkCompressor(urlLength int, urlAnswer string, logger *logrus.Logger) LinkCompressor {
func NewLinkCompressor(cfg config.Config, logger *logrus.Logger) LinkCompressor {
	lc := LinkCompressor{
		urlLength:   cfg.URLLength,
		ServiceName: cfg.BaseURL,
		logger: logger,
	}
	logger.Info("the link compressor success init")
	return lc
}


// SortURL - собирает сокращенный URL
// 			 TODO: Опечатка. И переименовать в отражающее суть: MakeShortURL
func (l *LinkCompressor) SortURL(originalLink string) string {
	path := l.shortPath(originalLink)
	url := fmt.Sprintf("%s/%s", l.ServiceName, path[:l.urlLength])
	return url
}

// ShortPath алгоритм сокращения URL на основе base58 - для получения
//					 набора символов которые человеком могут читатся однозначно.
func (l *LinkCompressor) shortPath(originalLink string) string {
	generatedNumber := new(big.Int).SetBytes([]byte(originalLink)).Uint64()
	finalString := l.base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:l.urlLength]
}

func (l *LinkCompressor) base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(encoded)
}
