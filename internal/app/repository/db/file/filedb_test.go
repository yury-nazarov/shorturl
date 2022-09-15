package filedb

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkFileDB(b *testing.B) {
	db := NewFileDB("TestFDB.txt")
	token := "123456789qwertyXYZ"
	ctx := context.Background()
	b.Run("inmemory_Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Add(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), fmt.Sprintf("%s_%d", "https://www.youtube.com/watch?v=09nmlZjxRFs", i), fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Get(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), token)
		}
	})
	b.Run("inmemory_GetToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetToken(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_GetUserURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetUserURL(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_OriginURLExists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.OriginURLExists(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i))
		}
	})
}
