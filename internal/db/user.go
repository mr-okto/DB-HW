package db

import (
	"db-hw/internal/models"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

func GetUser(nickname string) (models.User, error) {
	u := models.User{}
	res, err := DB.Query(`
		SELECT about, email, fullname, nickname 
		FROM users WHERE nickname = $1`,
		nickname)

	if err != nil {
		return models.User{}, errors.Wrap(err, "GetUser error")
	}
	defer res.Close()
	if res.Next() {
		_ = res.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
		return u, nil
	}
	return models.User{}, errors.New("GetUser error")
}

func GetExistingUsers(nickname string, email string) (models.Users, error) {
	result := make([]models.User, 0, 1)
	u := models.User{}
	res, err := DB.Query(`
		SELECT about, email, fullname, nickname 
		FROM users 
		WHERE email = $1 OR nickname = $2`,
		email, nickname)

	if err != nil {
		return []models.User{}, errors.Wrap(err, "GetExistingUsers error")
	}
	defer res.Close()

	for res.Next() {
		err := res.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
		if err != nil {
			return []models.User{}, errors.Wrap(err, "db query result parsing error")
		}
		result = append(result, u)
	}
	return result, nil
}

func CreateUser(user models.User) (models.User, error) {
	res, err := DB.Exec(`
		INSERT INTO users (nickname, fullname, email, about) 
		VALUES ($1, $2, $3, $4)`,
		user.Nickname, user.Fullname,
		user.Email, user.About)

	if err != nil {
		return models.User{}, errors.Wrap(err, "CreateUser error")
	}
	if res.RowsAffected() == 0 {
		return models.User{}, errors.Wrap(err, "CreateUser error")
	}
	return user, nil
}

func UpdateUser(user models.User) (models.User, error, int) {
	if user.About == "" &&
		user.Email == "" &&
		user.Fullname == "" {
		updatedUser, _ := GetUser(user.Nickname)
		return updatedUser, nil, http.StatusOK
	}
	sql := "Update users SET"
	if user.Fullname == "" {
		sql += " fullname = fullname,"
	} else {
		sql += " fullname = '" + user.Fullname + "',"
	}
	if user.Email == "" {
		sql += " email = email,"
	} else {
		sql += " email = '" + user.Email + "',"
	}
	if user.About == "" {
		sql += " about = about"
	} else {
		sql += " about = '" + user.About + "'"
	}
	sql += " WHERE nickname = '" + user.Nickname + "'"

	res, err := DB.Exec(sql)
	if err != nil {
		return models.User{}, errors.Wrap(err, "UpdateUser error"),
			http.StatusConflict
	}
	if res.RowsAffected() == 0 {
		return models.User{}, errors.New("UpdateUser user not found"),
			http.StatusNotFound
	}
	result, _ := GetUser(user.Nickname)
	return result, nil, http.StatusOK
}

func GetUsers(forum models.Forum,
	limit int, since string, desc bool) (models.Users, error, int) {
	sql := `
		SELECT about, email, fullname, u.nickname
		FROM forum_users fu
		JOIN users u
			ON fu.nickname = u.nickname`

	sql += ` where slug = '` + forum.Slug + `'`
	if since != "" {
		if desc {
			sql += ` AND u.nickname < '` + since + `'`
		} else {
			sql += ` AND u.nickname > '` + since + `'`
		}
	}
	if desc {
		sql += " ORDER BY nickname DESC"
	} else {
		sql += " ORDER BY nickname ASC"
	}
	if limit != 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}
	res, _ := DB.Query(sql)
	defer res.Close()
	users := make([]models.User, 0, 1)
	u := models.User{}
	for res.Next() {
		_ = res.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
		users = append(users, u)
	}
	return users, nil, http.StatusOK
}

func AddUserToForum(nickname string, slug string) {
	_, _ = DB.Exec(`
		INSERT INTO forum_users (nickname, slug) 
		VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		nickname, slug)
	return
}
