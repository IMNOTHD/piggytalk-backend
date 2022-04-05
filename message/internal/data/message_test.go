package data

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMessage(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/piggytalk?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"), &gorm.Config{})
	if err != nil {
		t.Error(err)
	}
	db.Migrator().CreateTable(&FriendAddMessage{})
	ua := uuid.MustParse("38ec8e53-7c68-4410-b81e-04c2b5eb4f0c")
	ub := uuid.MustParse("c0cfb721-5bd0-4e5e-951f-089149fa5d9d")
	for i := 0; i < 500; i++ {
		k := make([]*FriendAddMessage, 0)
		for j := 0; j < 3000; j++ {
			if j%1000 == 0 {
				k = append(k, &FriendAddMessage{
					EventId:   int64(i*3000+j)*1000 + rand.Int63n(1000),
					UserA:     ua,
					UserB:     ub,
					Type:      "WAITING",
					EventUuid: uuid.New(),
				})
			} else if j%1000 == 1 {
				k = append(k, &FriendAddMessage{
					EventId:   int64(i*3000+j)*1000 + rand.Int63n(1000),
					UserA:     ub,
					UserB:     ua,
					Type:      "WAITING",
					EventUuid: uuid.New(),
				})
			} else {
				k = append(k, &FriendAddMessage{
					EventId:   int64(i*3000+j)*1000 + rand.Int63n(1000),
					UserA:     uuid.New(),
					UserB:     uuid.New(),
					Type:      "WAITING",
					EventUuid: uuid.New(),
				})
			}
		}
		db.Create(k)
	}
}
