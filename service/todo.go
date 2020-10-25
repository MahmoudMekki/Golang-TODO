package service

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/TODO/m/view"
)

func Index(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(res, "index.gohtml", nil)
}

func AddTask(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodPost {
		task := view.Task{}
		claims := &view.UserClaims{}
		c, _ := req.Cookie("TodoSession")
		token := c.Value
		ok, err := checkMaxTodoPerDay(token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "Reached max todos per today", http.StatusNotAcceptable)
			return
		}
		claims, err = parseToken(token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusForbidden)
			return
		}
		task.Assigner = claims.UserName
		task.Assignee = req.FormValue("assignee")
		task.Content = req.FormValue("content")
		task.IssueDate = time.Now().Format("01-02-2006")
		task.DueDate = req.FormValue("dueDate")
		task.State = req.FormValue("status")

		stmt, err := db.Prepare("INSERT INTO Task (assigner,content,state,assignee,issue_date,due_date) VALUES (?,?,?,?,?,?);")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(task.Assigner, task.Content, task.State, task.Assignee, task.IssueDate, task.DueDate)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		token, err = updateTokenOnAdd(token)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Value = token
		http.SetCookie(res, c)
	}
	tpl.ExecuteTemplate(res, "addtask.gohtml", nil)

}

func ShowTasks(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	un := getUser(req)
	tasks := []view.Task{}
	rows, err := db.Query("SELECT * FROM Task WHERE assigner =?;", un)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	tpl.ExecuteTemplate(res, "showtasks.gohtml", tasks)
}

func ShowTask(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	taskID := req.FormValue("taskid")
	task := view.Task{}
	row := db.QueryRow("SELECT * FROM Task WHERE task_id =?;", taskID)
	err := row.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(res, req)
		return
	case err != nil:
		http.Error(res, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(res, "showtask.gohtml", task)
}

func UpdateTask(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodGet || req.Method == http.MethodPut {
		taskid := req.FormValue("taskid")
		task := view.Task{}
		row := db.QueryRow("SELECT * FROM Task WHERE task_id =?;", taskid)
		err := row.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		switch {
		case err == sql.ErrNoRows:
			http.NotFound(res, req)
			return
		case err != nil:
			http.Error(res, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		tpl.ExecuteTemplate(res, "update.gohtml", task)

	} else if req.Method == http.MethodPost {
		task := view.Task{}
		task.TaskID = req.FormValue("taskid")
		task.Assigner = req.FormValue("assigner")
		task.Assignee = req.FormValue("assignee")
		task.Content = req.FormValue("content")
		task.IssueDate = req.FormValue("issueDate")
		task.DueDate = req.FormValue("dueDate")
		task.State = req.FormValue("status")

		stmt, err := db.Prepare("UPDATE Task SET assigner=?,assignee=?,content=?,issue_date=?,due_date=?,state=? WHERE task_id=?;")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(&task.Assigner, &task.Assignee, &task.Content, &task.IssueDate, &task.DueDate, &task.State, &task.TaskID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tpl.ExecuteTemplate(res, "updated.gohtml", task)

	}

}

func DeleteTask(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	taskid := req.FormValue("taskid")

	_, err := db.Exec("DELETE FROM Task WHERE task_id=?;", taskid)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(res, req, "/task", http.StatusSeeOther)

}

func ShowCompleted(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	un := getUser(req)
	rows, err := db.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", un, "completed")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	tasks := []view.Task{}
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	tpl.ExecuteTemplate(res, "complete.gohtml", tasks)
}

func OverDue(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	un := getUser(req)
	rows, err := db.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", un, "overdue")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	tasks := []view.Task{}
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	tpl.ExecuteTemplate(res, "overdue.gohtml", tasks)
}

func Pending(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	un := getUser(req)
	rows, err := db.Query("SELECT * FROM Task WHERE assigner=? AND NOT state = ? OR NOT state=?;", un, "overdue", "completed")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	tasks := []view.Task{}
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	tpl.ExecuteTemplate(res, "pending.gohtml", tasks)
}
func TopAssigners(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	toppers := []view.Toppers{}
	rows, err := db.Query("SELECT assigner,COUNT(assigner) AS amount FROM Task GROUP BY assigner ORDER BY amount DESC LIMIT 2;")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		topper := view.Toppers{}
		err := rows.Scan(&topper.Name, &topper.Amount)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		toppers = append(toppers, topper)
	}
	tpl.ExecuteTemplate(res, "topassigners.gohtml", toppers)
}

func TopAssignees(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	toppers := []view.Toppers{}
	rows, err := db.Query("SELECT assignee,COUNT(assignee) AS amount FROM Task GROUP BY assignee ORDER BY amount DESC LIMIT 2;")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		topper := view.Toppers{}
		err := rows.Scan(&topper.Name, &topper.Amount)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		toppers = append(toppers, topper)
	}
	tpl.ExecuteTemplate(res, "topassignees.gohtml", toppers)
}

func TopResolvers(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedin(req) {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(res, "Bad Request Method", http.StatusBadRequest)
		return
	}
	toppers := []view.Toppers{}
	rows, err := db.Query("SELECT assigner,assignee,COUNT(state) AS amount FROM Task WHERE state = ? GROUP BY assigner,assignee ORDER BY amount DESC LIMIT 2 ;", "completed")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		topper := view.Toppers{}
		err := rows.Scan(&topper.Name, &topper.Name2, &topper.Amount)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		toppers = append(toppers, topper)
	}
	tpl.ExecuteTemplate(res, "topresolvers.gohtml", toppers)
}
