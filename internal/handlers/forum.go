package handlers

import (
	"db-hw/internal/db"
	"db-hw/internal/models"
	w "db-hw/internal/writers"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func CreateForum(res http.ResponseWriter, req *http.Request) {
	f := models.Forum{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = f.UnmarshalJSON(body)
	forum, err := db.CreateForum(f)
	if err != nil {
		_, err := db.GetUser(f.User)
		if err != nil {
			w.WriteError(res, http.StatusNotFound, "CreateForum error")
			return
		}
		conflictForum, err := db.GetForum(f.Slug)
		if err != nil {
			w.WriteError(res, http.StatusNotFound, err.Error())
			return
		}
		w.WriteEasyJson(res, http.StatusConflict, conflictForum)
		return
	}
	w.WriteEasyJson(res, http.StatusCreated, forum)
}

func GetForum(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug"]
	f, err := db.GetForum(slug)
	if err != nil || f.User == "" {
		w.WriteError(res, http.StatusNotFound, "GetForum error")
		return
	}
	w.WriteEasyJson(res, http.StatusOK, f)
}
