package db

import (
	"db-hw/internal/models"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

func CreateForum(forum models.Forum) (models.Forum, error) {
	user, err := GetUser(forum.User)
	if err != nil {
		return models.Forum{}, errors.Wrap(err, "CreateForum error")
	}
	forum.User = user.Nickname
	_, err = DB.Exec(`
		INSERT INTO forums (posts, slug, threads, title, "user") 
		VALUES ($1, $2, $3, $4, $5)`,
		forum.Posts, forum.Slug,
		forum.Threads, forum.Title, forum.User)
	if err != nil {
		return models.Forum{}, errors.Wrap(err, "CreateForum error")
	}
	return forum, nil
}

func GetForum(slug string) (models.Forum, error) {
	res, err := DB.Query(`
		SELECT * FROM forums WHERE slug = $1`,
		slug)

	if err != nil {
		return models.Forum{}, errors.Wrap(err, "GetForum error")
	}
	defer res.Close()

	f := models.Forum{}
	if res.Next() {
		err := res.Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
		if err != nil {
			return models.Forum{}, errors.Wrap(err, "GetForum error")
		}
	}
	if f.Slug == "" {
		return models.Forum{}, errors.New("cannot get forum by slug")
	}
	return f, nil
}

func UpdateForum(forum models.Forum,
	vars string, inc bool, diff int) int {
	sql := "UPDATE forums SET"
	switch vars {
	case "post":
		sql += " posts = posts"
		if inc {
			sql += "+"
		} else {
			sql += "-"
		}
		sql += strconv.Itoa(diff)
	case "thread":
		sql += " threads = threads"
		if inc {
			sql += "+"
		} else {
			sql += "-"
		}
		sql += strconv.Itoa(diff)
	default:
		return http.StatusInternalServerError
	}
	sql += " WHERE slug = $1"
	res, err := DB.Exec(sql, forum.Slug)
	if err != nil {
		return http.StatusConflict
	}
	if res.RowsAffected() == 0 {
		return http.StatusNotFound
	}
	return http.StatusOK
}
