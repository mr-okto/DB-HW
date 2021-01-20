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

func CreateThread(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug"]
	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = t.UnmarshalJSON(body)
	t.Forum = slug
	thread, err, status := db.CreateThread(t)
	if err != nil {
		if status == http.StatusNotFound {
			w.WriteError(res, status, err.Error())
			return
		}
		if status == http.StatusConflict {
			errThread, _, _ := db.GetThread(-1, t.Slug)
			w.WriteEasyJson(res, status, errThread)
			return
		}
	}
	w.WriteEasyJson(res, http.StatusCreated, thread)
}

func GetThreads(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug"]
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))
	threads, err, status := db.GetThreads(slug, limit, since, desc)
	if err != nil {
		if status == http.StatusNotFound {
			w.WriteError(res, status, err.Error())
			return
		}
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, http.StatusOK, threads)
}

func UpdateThread(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug_or_id"]
	id, _ := strconv.ParseInt(slug, 10, 32)
	if id == 0 {
		id = -1
	}
	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	t.UnmarshalJSON(body)
	thread, err, status := db.GetThread(int(id), slug)
	if err != nil {
		w.WriteError(res, status, errors.Wrap(err, "UpdateThread error").Error())
		return
	}
	result, err, status := db.UpdateThread(thread, t)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, status, result)
}

func AddVote(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug_or_id"]
	id, _ := strconv.ParseInt(slug, 10, 32)
	if id == 0 {
		id = -1
	}
	thread, err, status := db.GetThread(int(id), slug)
	if err != nil {
		w.WriteError(res, status, errors.Wrap(err, "AddVote error").Error())
		return
	}
	vote := models.Vote{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	vote.UnmarshalJSON(body)

	user, err := db.GetUser(vote.Nickname)
	if err != nil {
		w.WriteError(res, http.StatusNotFound, errors.Wrap(err, "AddVote error").Error())
		return
	}
	vote.Thread = thread.ID
	vote.Nickname = user.Nickname
	result, err, status := db.CreateVote(vote)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, http.StatusOK, result)
}

func GetThread(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug_or_id"]
	id, _ := strconv.ParseInt(slug, 10, 32)
	if id == 0 {
		id = -1
	}
	thread, err, status := db.GetThread(int(id), slug)
	if err != nil {
		w.WriteError(res, status, errors.Wrap(err, "slug not found").Error())
		return
	}
	w.WriteEasyJson(res, http.StatusOK, thread)
}
