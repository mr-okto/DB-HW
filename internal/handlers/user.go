package handlers

import (
	"db-hw/internal/db"
	"db-hw/internal/models"
	w "db-hw/internal/writers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetUser(res http.ResponseWriter, req *http.Request) {
	nickname, _ := mux.Vars(req)["nickname"]
	u, err := db.GetUser(nickname)
	if err != nil || u.Email == "" {
		w.WriteError(res, http.StatusNotFound, "GetUser error")
		return
	}
	w.WriteEasyJson(res, http.StatusOK, u)
}

func UpdateUser(res http.ResponseWriter, req *http.Request) {
	nickname, _ := mux.Vars(req)["nickname"]
	u := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = u.UnmarshalJSON(body)
	u.Nickname = nickname
	result, err, errCode := db.UpdateUser(u)
	if err != nil {
		w.WriteError(res, errCode, err.Error())
		return
	}
	w.WriteEasyJson(res, http.StatusOK, result)
}

func CreateUser(res http.ResponseWriter, req *http.Request) {
	nickname, _ := mux.Vars(req)["nickname"]
	u := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = u.UnmarshalJSON(body)
	u.Nickname = nickname
	result, err := db.CreateUser(u)
	if err != nil {
		exitingUsers, err := db.GetExistingUsers(u.Nickname, u.Email)
		if err != nil {
			w.WriteError(res, http.StatusConflict, err.Error())
			return
		}
		w.WriteEasyJson(res, http.StatusConflict, exitingUsers)
		return
	}
	w.WriteEasyJson(res, http.StatusCreated, result)
}

func GetUsers(res http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))
	slug, _ := mux.Vars(req)["slug"]
	forum, err := db.GetForum(slug)
	if err != nil {
		w.WriteError(res, http.StatusNotFound,
			errors.Wrap(err, "GetUsers error").Error())
		return
	}
	users, err, status := db.GetUsers(forum, limit, since, desc)
	if err != nil {
		if status == http.StatusNotFound {
			w.WriteError(res, status, err.Error())
			return
		}
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, http.StatusOK, users)
}
