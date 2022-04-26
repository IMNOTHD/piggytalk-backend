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
	u := uuid.MustParse("8f52b66f-f1fd-4f05-ba0a-7488ebdb8b21")
	table := "single_message_" + u.String()

	ua := uuid.MustParse("38ec8e53-7c68-4410-b81e-04c2b5eb4f0c")
	ub := uuid.MustParse("c0cfb721-5bd0-4e5e-951f-089149fa5d9d")
	uc := uuid.MustParse("7e8b7435-6522-a557-1fa8-9c5b23d86ff8")
	ud := uuid.MustParse("bae23b0b-5a9a-a5a9-57c7-57e0d6a7e47a")
	ue := uuid.MustParse("1a9a82a3-8c38-8279-d509-2de536c1b996")

	var us []uuid.UUID = []uuid.UUID{ua, ub, uc, ud, ue}
	for i := 0; i < 50; i++ {
		k := make([]*SingleMessage, 0)

		for j := 0; j < 1000; j++ {
			uid := us[rand.Intn(5)]
			k = append(k, &SingleMessage{
				MessageId:   int64(i*1000 + j),
				SenderUuid:  uid,
				Talk:        uid,
				Message:     "",
				MessageUuid: uuid.New(),
				AlreadyRead: rand.Intn(2) != 0,
			})
		}
		db.Table(table).Create(k)
	}
}
