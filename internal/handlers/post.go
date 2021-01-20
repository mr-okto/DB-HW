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
	"strings"
)

func GetPost(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["id"]
	id, _ := strconv.ParseInt(slug, 10, 64)
	if id == 0 {
		id = -1
	}
	query := req.URL.Query()
	related := strings.Split(query.Get("related"), ",")
	post, err, status := db.GetPost(id)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	details, err, status := db.GetPostDetails(post, related)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	details.Post = &post
	w.WriteEasyJson(res, status, details)
}

func GetPosts(res http.ResponseWriter, req *http.Request) {
	slugOrID, _ := mux.Vars(req)["slug_or_id"]
	slug := slugOrID
	id, err := strconv.ParseInt(slug, 10, 32)
	if id == 0 {
		id = -1
	}
	thread, err, status := db.GetThread(int(id), slug)
	if err != nil {
		w.WriteError(res, status, errors.Wrap(err, "GetPosts error").Error())
		return
	}
	q := req.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	since, _ := strconv.Atoi(q.Get("since"))
	sort := q.Get("sort")
	sortDesc, _ := strconv.ParseBool(q.Get("desc"))
	posts, err, status := db.GetPosts(thread, limit, since, sort, sortDesc)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, status, posts)
}

func CreatePost(res http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["slug_or_id"]
	id, _ := strconv.ParseInt(slug, 10, 32)
	if id == 0 {
		id = -1
	}
	thread, err, status := db.GetThread(int(id), slug)
	if err != nil {
		w.WriteError(res, status, errors.Wrap(err, "CreatePost error").Error())
		return
	}
	posts := models.Posts{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = posts.UnmarshalJSON(body)
	if len(posts) == 0 {
		w.WriteEasyJson(res, http.StatusCreated, posts)
		return
	}
	result, err, status := db.CreatePosts(posts, thread)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, http.StatusCreated, result)
}

func UpdatePost(res http.ResponseWriter, req *http.Request) {
	key, _ := mux.Vars(req)["id"]
	id, err := strconv.ParseInt(key, 10, 64)
	if id == 0 {
		id = -1
	}
	post, err, status := db.GetPost(id)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	modification := models.Post{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = modification.UnmarshalJSON(body)
	result, err, status := db.UpdatePost(post, modification)
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, status, result)
}
