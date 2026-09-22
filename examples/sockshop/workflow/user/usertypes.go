package user

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type User struct {
	FirstName string    `bson:"FirstName"`
	LastName  string    `bson:"LastName"`
	Email     string    `bson:"Email"`
	Username  string    `bson:"Username"`
	Password  string    `bson:"Password,omitempty"`
	Addresses []Address `bson:"Addresses"`
	Cards     []Card    `bson:"Cards"`
	UserID    string    `bson:"UserID"`
	Salt      string    `bson:"Salt"`
}
type Address struct {
	Street   string `bson:"Street"`
	Number   string `bson:"Number"`
	Country  string `bson:"Country"`
	City     string `bson:"City"`
	PostCode string `bson:"PostCode"`
	ID       string `bson:"ID"`
}
type Card struct {
	LongNum string `bson:"LongNum"`
	Expires string `bson:"Expires"`
	CCV     string `bson:"CCV"`
	ID      string `bson:"ID"`
}

func newUser() User {
	u := User{Addresses: make([]Address, 0), Cards: make([]Card, 0)}
	u.newSalt()
	return u
}

func (u *User) newSalt() {
	h := sha1.New()
	io.WriteString(h, strconv.Itoa(int(time.Now().UnixNano())))
	u.Salt = fmt.Sprintf("%x", h.Sum(nil))
}

func (u *User) cardIDs() []string {
	ids := []string{}
	for _, card := range u.Cards {
		ids = append(ids, card.ID)
	}
	return ids
}

func (u *User) addressIDs() []string {
	ids := []string{}
	for _, address := range u.Addresses {
		ids = append(ids, address.ID)
	}
	return ids
}

// Replace all CC numbers with asterisks, for returning to the user for display
func (u *User) maskCCs() {
	for i := range u.Cards {
		u.Cards[i].maskCC()
	}
}

// Replaces the CC number with asterisks for returning to the user for display
func (c *Card) maskCC() {
	l := len(c.LongNum) - 4
	c.LongNum = fmt.Sprintf("%v%v", strings.Repeat("*", l), c.LongNum[l:])
}
