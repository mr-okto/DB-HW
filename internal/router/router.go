package router

import (
	h "db-hw/internal/handlers"
	"github.com/gorilla/mux"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/post/{id}/details", h.GetPost).Methods("get")
	r.HandleFunc("/api/post/{id}/details", h.UpdatePost).Methods("post")

	r.HandleFunc("/api/thread/{slug_or_id}/posts", h.GetPosts).Methods("get")
	r.HandleFunc("/api/thread/{slug_or_id}/details", h.GetThread).Methods("get")
	r.HandleFunc("/api/thread/{slug_or_id}/create", h.CreatePost).Methods("post")
	r.HandleFunc("/api/thread/{slug_or_id}/details", h.UpdateThread).Methods("post")
	r.HandleFunc("/api/thread/{slug_or_id}/vote", h.AddVote).Methods("post")

	r.HandleFunc("/api/user/{nickname}/profile", h.GetUser).Methods("get")
	r.HandleFunc("/api/user/{nickname}/profile", h.UpdateUser).Methods("post")
	r.HandleFunc("/api/user/{nickname}/create", h.CreateUser).Methods("post")

	r.HandleFunc("/api/forum/{slug}/details", h.GetForum).Methods("get")
	r.HandleFunc("/api/forum/{slug}/users", h.GetUsers).Methods("get")
	r.HandleFunc("/api/forum/{slug}/threads", h.GetThreads).Methods("get")
	r.HandleFunc("/api/forum/create", h.CreateForum).Methods("post")
	r.HandleFunc("/api/forum/{slug}/create", h.CreateThread).Methods("post")

	r.HandleFunc("/api/service/status", h.GetInfo).Methods("get")
	r.HandleFunc("/api/service/clear", h.ClearData).Methods("post")

	return r
}
