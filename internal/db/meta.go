package db

import (
	"db-hw/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
)

var DB *pgx.ConnPool

func InitDB(db *pgx.ConnPool) error {
	file, err := ioutil.ReadFile("init/tables_init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(file))
	return err
}

func GetDBCountData() (models.Info, error, int) {
	s := models.Info{}
	_ = DB.QueryRow("SELECT count(posts) from forums").Scan(&s.Forum)
	_ = DB.QueryRow("SELECT count(id) from posts").Scan(&s.Post)
	_ = DB.QueryRow("SELECT count(id) from threads").Scan(&s.Thread)
	_ = DB.QueryRow("SELECT count(nickname) from users").Scan(&s.User)
	return s, nil, http.StatusOK
}

func ClearDB() {
	res, _ := DB.Query(`
		TRUNCATE TABLE users, forums, 
			threads, posts, votes CASCADE`)
	defer res.Close()
}
