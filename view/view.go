package view

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	jwt.StandardClaims
	UserName string
	Password string
	MaxTODO  int
	Created  int
	Date     string
}
type Toppers struct {
	Name   string
	Name2  string
	Amount int
}
type User struct {
	Userid   string
	Password string
	Max      int
	Created  int
}
type Task struct {
	TaskID    string
	Assigner  string
	Content   string
	IssueDate string
	DueDate   string
	State     string
	Assignee  string
}

func (u *UserClaims) Valid() error {
	if !u.VerifyExpiresAt(time.Now().Unix(), false) {
		return fmt.Errorf("Token is timed out")
	}

	return nil
}
