package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/TODO/m/view"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("MahmoudMekk")

func createToken(c *view.UserClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, c)
	token, err := t.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("Error while generating token")
	}
	return token, nil
}

func parseToken(token string) (*view.UserClaims, error) {
	claims := &view.UserClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, fmt.Errorf("Invalid signing algorithm")
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error in parsing the token")
	}
	if !t.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	return t.Claims.(*view.UserClaims), nil

}

func hashPassword(password string) ([]byte, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error with converting to hash %w", err)
	}
	return bs, nil
}

func compare(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		return false
	}
	return true
}

func alreadyLoggedin(req *http.Request) bool {
	claims := &view.UserClaims{}
	u := view.User{}
	c, err := req.Cookie("TodoSession")
	if err != nil {
		return false
	}
	token := c.Value
	claims, err = parseToken(token)
	if err != nil {
		return false
	}
	row := db.QueryRow(`SELECT * FROM Users WHERE userid=? AND password=?;`, claims.UserName, claims.Password)
	err = row.Scan(&u.Userid, &u.Password, &u.Max, &u.Created)
	if err != nil {
		return false
	}
	return true
}

func getUser(req *http.Request) string {
	claims := &view.UserClaims{}
	c, _ := req.Cookie("TodoSession")
	token := c.Value
	claims, _ = parseToken(token)
	username := claims.UserName
	return username
}

func updateToken(token string) (string, error) {

	claims, _ := parseToken(token)
	if claims.Date != time.Now().Format("01-02-2006") {
		claims.Date = time.Now().Format("01-02-2006")
		claims.Created = 0
		un := claims.UserName
		stmt, err := db.Prepare("UPDATE Users SET created = ? WHERE userid =?;")
		if err != nil {
			return "", err
		}
		_, err = stmt.Exec(0, un)
		if err != nil {
			return "", err
		}
		token, err = createToken(claims)
		if err != nil {
			return "", err
		}
	}
	return token, nil
}

// in logout
func updateTokenDB(token string, req *http.Request) error {
	stmt, err := db.Prepare("UPDATE LastActivity SET token = ? WHERE user_id=?;")
	if err != nil {
		return err
	}
	un := getUser(req)
	_, err = stmt.Exec(token, un)
	if err != nil {
		return err
	}
	return nil
}

func updateTokenOnAdd(token string) (string, error) {
	token, err := updateToken(token)
	if err != nil {
		return "", err
	}
	claims := &view.UserClaims{}
	claims, err = parseToken(token)
	if err != nil {
		return "", err
	}
	claims.Created++
	stmt, err := db.Prepare("UPDATE Users SET created = ? WHERE userid=?;")
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(claims.Created, claims.UserName)
	if err != nil {
		return "", err
	}
	token, err = createToken(claims)
	if err != nil {
		return "", err
	}

	return token, nil

}

func checkMaxTodoPerDay(token string) (bool, error) {
	claims := &view.UserClaims{}
	claims, err := parseToken(token)
	if err != nil {
		return false, err
	}
	currentTime := time.Now()
	today := currentTime.Format("01-02-2006")
	if claims.Date == today && claims.MaxTODO-claims.Created == 0 {
		return false, nil
	}
	return true, nil
}
