package handler

import (
	"fmt"
	"log"
	"math/big"

	"github.com/itchyny/base58-go"
)

// ShortPath алгоритм сокращения URL на основе base58 - для получения
//					 набора символов которые человеком могут читатся однозначно.
func ShortPath(initialLink string, URLLength int) string {
	generatedNumber := new(big.Int).SetBytes([]byte(initialLink)).Uint64()
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:URLLength]
}

func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(encoded)
}