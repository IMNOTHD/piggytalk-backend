package ent

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestExample(t *testing.T) {
	client, err := Open(dialect.MySQL, "root:123456@tcp(127.0.0.1:3306)/piggytalk?parseTime=True")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	//_, err = CreateUser(ctx, client)
	//if err != nil {
	//	fmt.Printf("failed creating user: %v", err)
	//}

}

func CreateUser(ctx context.Context, client *Client) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), 0)

	u, err := client.User.
		Create().
		SetUsername("piggy").
		SetPassword(string(hash)).
		SetUUID(uuid.New()).
		SetGmtCreate(time.Now()).
		SetGmtModified(time.Now()).
		Save(ctx)
	fmt.Println(bcrypt.CompareHashAndPassword(hash, []byte("123456")))
	if err != nil {
		return nil, fmt.Errorf("failed creating user: %v", err)
	}
	log.Println("user was created: ", u)
	return u, nil
}
