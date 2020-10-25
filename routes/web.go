package routes

import (
	"net/http"

	"github.com/TODO/m/service"
)

func InitRoutes() {

	http.HandleFunc("/", service.Index)
	http.HandleFunc("/signup", service.SignUp)
	http.HandleFunc("/login", service.Login)
	http.HandleFunc("/logout", service.Logout)
	http.HandleFunc("/task", service.ShowTasks)
	http.HandleFunc("/task/add", service.AddTask)
	http.HandleFunc("/task/show", service.ShowTask)
	http.HandleFunc("/task/update", service.UpdateTask)
	http.HandleFunc("/task/delete", service.DeleteTask)
	http.HandleFunc("/task/overdue", service.OverDue)
	http.HandleFunc("/task/complete", service.ShowCompleted)
	http.HandleFunc("/task/pending", service.Pending)
	http.HandleFunc("/task/topassigners", service.TopAssigners)
	http.HandleFunc("/task/topassignees", service.TopAssignees)
	http.HandleFunc("/task/topresolvers", service.TopResolvers)

	http.Handle("/favicon.ico", http.NotFoundHandler())
}
