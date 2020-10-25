package service

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	config "github.com/TODO/m/config_db"
	"github.com/TODO/m/view"
)

var db = config.Database()
var tpl = template.Must(template.ParseGlob("./templates/*gohtml"))

func SignUp(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		u := view.User{}
		claims := view.UserClaims{}
		un := req.FormValue("username")
		pwd := req.FormValue("password")
		mx, err := strconv.Atoi(req.FormValue("max"))
		if err != nil {
			http.Error(res, "Ivalied maxTodo entry!", http.StatusBadRequest)
			return
		}
		stmt := `SELECT user_id FROM Users WHERE user_id = ?;`
		row := db.QueryRow(stmt, un)
		err = row.Scan(&u)
		if err == nil {
			http.Error(res, "the user name is already taken", http.StatusBadRequest)
			return
		}
		hashpass, err := hashPassword(pwd)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		u = view.User{
			Userid:   un,
			Password: string(hashpass),
			Max:      mx,
			Created:  0,
		}
		claims = view.UserClaims{
			UserName: un,
			Password: string(hashpass),
			MaxTODO:  mx,
			Created:  0,
			Date:     time.Now().Format("01-02-2006"),
		}
		token, err := createToken(&claims)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		stm, err := db.Prepare(`INSERT INTO Users VALUES (?,?,?,?);`)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stm.Exec(u.Userid, u.Password, u.Max, u.Created)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		stm, err = db.Prepare("INSERT INTO LastActivity VALUES (?,?);")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stm.Exec(un, token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(res, req, "/login", http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(res, "signup.gohtml", nil)

}

func Login(res http.ResponseWriter, req *http.Request) {
	if alreadyLoggedin(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodPost {
		u := view.User{}
		un := req.FormValue("username")
		password := req.FormValue("password")

		row := db.QueryRow("SELECT * FROM Users WHERE userid =?;", un)
		err := row.Scan(&u.Userid, &u.Password, &u.Max, &u.Created)
		if err != nil {
			http.Error(res, "Wrond Username or PWD", http.StatusBadRequest)
			return
		}

		if !compare(password, []byte(u.Password)) {
			http.Error(res, "Wrond Username or PWD", http.StatusBadRequest)
			return
		}
		token := ""
		row = db.QueryRow("SELECT token FROM LastActivity WHERE user_id =?;", un)
		err = row.Scan(&token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err = updateToken(token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		c := &http.Cookie{
			Name:  "TodoSession",
			Value: token,
		}
		http.SetCookie(res, c)
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(res, "login.gohtml", nil)
}

func Logout(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	c, err := req.Cookie("TodoSession")
	if err != nil {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	token := c.Value
	err = updateTokenDB(token, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	c.MaxAge = -1
	http.SetCookie(res, c)
	http.Redirect(res, req, "/login", http.StatusSeeOther)
	return
}
