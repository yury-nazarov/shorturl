package service

import (
	"fmt"
	"log"
	"math/big"

	"github.com/itchyny/base58-go"
)

type LinkCompressor struct {
	urlLength int
}

func NewLinkCompressor(urlLength int) LinkCompressor{
	lc := LinkCompressor{
		urlLength: urlLength,
	}
	return lc
}

// ShortPath алгоритм сокращения URL на основе base58 - для получения
//					 набора символов которые человеком могут читатся однозначно.
func (l *LinkCompressor) ShortPath(originalLink string) string {
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