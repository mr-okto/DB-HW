package db

import (
	"db-hw/internal/models"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetPost(id int64) (models.Post, error, int) {
	res, err := DB.Query(`
		SELECT author, created, forum, id, isedited, message, parent, thread, path 
		FROM posts WHERE id = $1`,
		id)
	if err != nil {
		return models.Post{},
			errors.Wrap(err, "GetPost error"), http.StatusNotFound
	}
	defer res.Close()
	post := models.Post{}
	for res.Next() {
		_ = res.Scan(&post.Author, &post.Created, &post.Forum,
			&post.ID, &post.IsEdited, &post.Message, &post.Parent,
			&post.Thread, pq.Array(&post.Path))
		return post, nil, http.StatusOK
	}
	return models.Post{},
		errors.New("GetPost error"), http.StatusNotFound
}

func CreatePosts(posts []models.Post,
	existingThread models.Thread) (models.Posts, error, int) {
	tx, _ := DB.Begin()
	defer tx.Rollback()

	parents := make(map[int64]models.Post)
	users := make(map[string]string)
	for _, post := range posts {
		if _, ok := parents[post.Parent]; !ok && post.Parent != 0 {
			parentPostQuery, err, _ := GetPost(post.Parent)
			if err != nil {
				return models.Posts{},
					errors.Wrap(err, "CreatePosts error"), http.StatusConflict
			}
			if parentPostQuery.Thread != existingThread.ID {
				return models.Posts{},
					errors.New("CreatePosts error"), http.StatusConflict
			}
			parents[post.Parent] = parentPostQuery
		}
		if _, ok := users[post.Author]; !ok {
			users[post.Author] = post.Author
		}
	}
	qRes, err := tx.Query(fmt.Sprintf(`
		SELECT nextval(pg_get_serial_sequence('posts', 'id')) 
		FROM generate_series(1, %d);`,
		len(posts)))

	if err != nil {
		log.Println(errors.Wrap(err, "CreatePosts error"))
		return models.Posts{},
			errors.Wrap(err, "CreatePosts error"), http.StatusNotFound
	}
	var postIds []int64
	for qRes.Next() {
		var availableId int64
		_ = qRes.Scan(&availableId)
		postIds = append(postIds, availableId)
	}
	qRes.Close()
	posts[0].Path = append(parents[posts[0].Parent].Path, postIds[0])

	err = tx.QueryRow(`
		INSERT INTO posts (id, author, forum, message, parent, thread, path)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created`,
		postIds[0], posts[0].Author, existingThread.Forum,
		posts[0].Message, posts[0].Parent,
		existingThread.ID,
		"{"+
			strings.Trim(strings.Replace(
				fmt.Sprint(posts[0].Path),
				" ", ",", -1), "[]")+"}").
		Scan(&posts[0].Created)

	if err != nil {
		log.Println(errors.Wrap(err, "CreatePosts error"))
		return models.Posts{}, errors.Wrap(err, "CreatePosts error"),
			http.StatusNotFound
	}

	now := posts[0].Created
	posts[0].Forum = existingThread.Forum
	posts[0].Thread = existingThread.ID
	posts[0].Created = time.Time(now)
	posts[0].ID = postIds[0]

	for i, post := range posts {
		if i == 0 {
			continue
		}
		post.Path = append(parents[post.Parent].Path, postIds[i])
		resInsert, err := tx.Exec(`
		INSERT INTO posts (id, author, created, forum, message, parent, thread, path) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			postIds[i], post.Author, now, existingThread.Forum,
			post.Message, post.Parent, existingThread.ID,
			"{"+
				strings.Trim(strings.Replace(fmt.Sprint(post.Path),
					" ", ",", -1),
					"[]")+"}")
		if err != nil {
			log.Println(errors.Wrap(err, "CreatePosts error"))
			return models.Posts{}, errors.Wrap(err, "CreatePosts error"), http.StatusNotFound
		}
		if resInsert.RowsAffected() == 0 {
			log.Println(errors.Wrap(err, "CreatePosts error"))
			return models.Posts{},
				errors.Wrap(err, "CreatePosts error"), http.StatusNotFound
		}
		posts[i].Forum = existingThread.Forum
		posts[i].Thread = existingThread.ID
		posts[i].Created = time.Time(now)
		posts[i].ID = postIds[i]
	}
	tx.Commit()
	status := UpdateForum(
		models.Forum{Slug: existingThread.Forum},
		"post",
		true,
		len(posts))
	if status != http.StatusOK {
		log.Println(errors.Wrap(err, "CreatePosts error"))
		return models.Posts{}, errors.New("CreatePosts error"), status
	}
	go func() {
		for _, val := range users {
			AddUserToForum(val, existingThread.Forum)
		}
	}()
	return posts, nil, http.StatusOK
}

func GetPostDetails(existingPost models.Post, related []string) (models.PostData, error, int) {
	sql := ""
	postData := models.PostData{}
	for _, val := range related {
		switch val {
		case "user":
			sql = `
			SELECT about, email, fullname, nickname 
			FROM users 
			WHERE nickname = $1`
			res, _ := DB.Query(sql, existingPost.Author)
			u := models.User{}
			for res.Next() {
				_ = res.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
			}
			postData.Author = &u
			res.Close()
		case "forum":
			sql = `
			SELECT posts, slug, threads, title, "user" 
			FROM forums 
			WHERE slug = $1`
			res, _ := DB.Query(sql, existingPost.Forum)
			f := models.Forum{}
			for res.Next() {
				_ = res.Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
			}
			postData.Forum = &f
			res.Close()
		case "thread":
			sql = `
			SELECT author, created, forum, id, message, slug, title, votes 
			FROM threads 
			WHERE id = $1`
			res, _ := DB.Query(sql, existingPost.Thread)
			t := models.Thread{}
			varchar := pgtype.Varchar{}
			for res.Next() {
				_ = res.Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message, &varchar, &t.Title, &t.Votes)
			}
			t.Slug = varchar.String
			postData.Thread = &t
			res.Close()
		}
	}
	return postData, nil, http.StatusOK
}

func SortFlat(parent models.Thread, limit int, from int,
	desc bool) string {
	id := strconv.FormatInt(int64(parent.ID), 10)
	sql := ""
	sql = `
	SELECT author, created, forum, id, isedited, message, parent, thread 
	FROM posts WHERE thread = ` + id

	if from != 0 {
		if desc {
			sql += " AND id < " + strconv.Itoa(from)
		} else {
			sql += " AND id > " + strconv.Itoa(from)
		}
	}
	if desc {
		sql += " ORDER BY id DESC"
	} else {
		sql += " ORDER BY id"
	}
	sql += " LIMIT " + strconv.Itoa(limit)
	return sql
}

func SortTree(parent models.Thread, limit int, from int,
	desc bool) string {
	id := strconv.FormatInt(int64(parent.ID), 10)
	sql := ""
	sql = `SELECT author, created, forum, id, isedited, message, parent, thread
	FROM posts WHERE thread = ` + id

	if from != 0 {
		if desc {
			sql += " AND path < (SELECT path FROM posts WHERE id = " +
				strconv.Itoa(from) + ")"
		} else {
			sql += " AND path > (SELECT path FROM posts WHERE id = " +
				strconv.Itoa(from) + ")"
		}
	}
	if desc {
		sql += " ORDER BY path DESC, id DESC"
	} else {
		sql += " ORDER BY path, id"
	}
	sql += " LIMIT " + strconv.Itoa(limit)
	return sql
}

func SortParentTree(parent models.Thread, limit int, since int,
	desc bool) string {
	baseSQL := ""
	baseSQL = `SELECT author, created, forum, id, isedited, message, parent, thread
		FROM posts 
		WHERE path[1] IN
		(SELECT id FROM posts WHERE thread = ` +
		strconv.FormatInt(int64(parent.ID), 10) +
		" AND parent = 0"
	if since != 0 {
		if desc {
			baseSQL += ` AND path[1] < (SELECT path[1]
			FROM posts WHERE id = ` + strconv.Itoa(since) + ")"
		} else {
			baseSQL += ` AND path[1] > (SELECT path[1]
			FROM posts WHERE id = ` + strconv.Itoa(since) + ")"
		}
	}
	if desc {
		baseSQL += " ORDER BY id DESC"
	} else {
		baseSQL += " ORDER BY id"
	}
	baseSQL += " LIMIT " + strconv.Itoa(limit) + ")"
	if desc {
		baseSQL += " ORDER BY path[1] DESC, path, id"
	} else {
		baseSQL += " ORDER BY path"
	}
	return baseSQL
}

func GetPosts(parent models.Thread,
	limit int, since int, sort string, desc bool) (
	models.Posts, error, int) {
	if sort == "" {
		sort = "flat"
	}
	sql := ""
	posts := make([]models.Post, 0, 1)
	switch sort {
	case "flat":
		sql = SortFlat(parent, limit, since, desc)
	case "tree":
		sql = SortTree(parent, limit, since, desc)
	case "parent_tree":
		sql = SortParentTree(parent, limit, since, desc)
	}
	res, _ := DB.Query(sql)
	defer res.Close()
	post := models.Post{}
	for res.Next() {
		_ = res.Scan(&post.Author, &post.Created,
			&post.Forum, &post.ID, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread)
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		return models.Posts{}, nil, http.StatusOK
	}
	return posts, nil, http.StatusOK
}

func UpdatePost(post models.Post, update models.Post) (models.Post, error, int) {
	if update.Message == "" {
		return post, nil, http.StatusOK
	}
	if post.Message == update.Message {
		return post, nil, http.StatusOK
	}
	res, err := DB.Exec(`
		UPDATE posts 
		SET message = $1, isedited = true WHERE id = $2`,
		update.Message, post.ID)
	if err != nil {
		return models.Post{},
			errors.Wrap(err, "UpdatePost error"), http.StatusConflict
	}
	if res.RowsAffected() == 0 {
		return models.Post{},
			errors.New("UpdatePost"), http.StatusNotFound
	}
	result, _, _ := GetPost(post.ID)
	return result, nil, http.StatusOK
}
