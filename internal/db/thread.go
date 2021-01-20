package db

import (
	"db-hw/internal/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

func CreateThread(thread models.Thread) (models.Thread, error, int) {
	forum, err := GetForum(thread.Forum)
	if err != nil {
		return models.Thread{}, errors.Wrap(err, "CreateThread error"),
			http.StatusNotFound
	}
	thread.Forum = forum.Slug
	user, err := GetUser(thread.Author)
	if err != nil {
		return models.Thread{}, errors.Wrap(err, "CreateThread error"),
			http.StatusNotFound
	}
	thread.Author = user.Nickname
	tx, _ := DB.Begin()
	defer tx.Rollback()
	if thread.Slug == "" {
		err := tx.QueryRow(`
		INSERT INTO threads (author, created, forum, message, slug, title) 
		VALUES ($1, $2, $3, $4, NULL, $5) RETURNING id`,
			thread.Author, thread.Created,
			thread.Forum, thread.Message,
			thread.Title).
			Scan(&thread.ID)

		if err == pgx.ErrNoRows {
			return models.Thread{},
				errors.Wrap(err, "CreateThread error"), http.StatusConflict
		} else if err != nil {
			return models.Thread{},
				errors.Wrap(err, "CreateThread error"), http.StatusConflict
		}
	} else {
		err := tx.QueryRow(`
		INSERT INTO threads (author, created, forum, message, slug, title) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			thread.Author, thread.Created,
			thread.Forum, thread.Message, thread.Slug,
			thread.Title).Scan(&thread.ID)
		if err == pgx.ErrNoRows {
			return models.Thread{},
				errors.Wrap(err, "CreateThread error"), http.StatusConflict
		} else if err != nil {
			return models.Thread{},
				errors.Wrap(err, "CreateThread error"), http.StatusConflict
		}
	}
	tx.Commit()

	status := UpdateForum(forum, "thread", true, 1)
	if status != http.StatusOK {
		return models.Thread{}, errors.New("CreateThread error"), status
	}
	AddUserToForum(thread.Author, forum.Slug)
	return thread, nil, http.StatusOK
}

func GetThreads(slug string, limit int, since string, desc bool) (
	models.Threads, error, int) {
	threads := make([]models.Thread, 0, 1)
	forum, err := GetForum(slug)
	if err != nil {
		return []models.Thread{},
			errors.Wrap(err, "GetThreads error"), http.StatusNotFound
	}
	sql := "SELECT * FROM threads"
	sql += " WHERE forum = '" + forum.Slug + "'"
	if since != "" {
		if desc {
			sql += " AND created <= '" + since + "'"
		} else {
			sql += " AND created >= '" + since + "'"
		}
	}
	if desc {
		sql += " ORDER BY created DESC"
	} else {
		sql += " ORDER BY created"
	}
	if limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}
	res, _ := DB.Query(sql)
	defer res.Close()
	t := models.Thread{}
	varchar := &pgtype.Varchar{}
	for res.Next() {
		_ = res.Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message,
			varchar, &t.Title, &t.Votes)
		t.Slug = varchar.String
		threads = append(threads, t)
	}
	return threads, nil, http.StatusOK
}

func GetThread(id int, slug string) (models.Thread, error, int) {
	t := models.Thread{}
	if id == -1 {
		res, err := DB.Query(`
		SELECT * FROM threads WHERE slug = $1`,
			slug)
		defer res.Close()
		if err != nil {
			return models.Thread{},
				errors.Wrap(err, "GetThread error"), http.StatusNotFound
		}
		if res.Next() {
			text := pgtype.Text{}
			_ = res.Scan(&t.Author, &t.Created, &t.Forum,
				&t.ID, &t.Message, &text, &t.Title, &t.Votes)
			t.Slug = text.String
			return t, nil, http.StatusOK
		}
		return models.Thread{},
			errors.New("GetThread error"), http.StatusNotFound
	} else if slug == "" {
		res, err := DB.Query(`
		SELECT * FROM threads WHERE id = $1`, id)
		defer res.Close()
		if err != nil {
			return models.Thread{},
				errors.Wrap(err, "GetThread error"), http.StatusNotFound
		}
		if res.Next() {
			text := pgtype.Text{}
			res.Scan(&t.Author, &t.Created, &t.Forum, &t.ID,
				&t.Message, &text, &t.Title, &t.Votes)
			t.Slug = text.String
			return t, nil, http.StatusOK
		}
		return models.Thread{},
			errors.New("GetThread error"), http.StatusNotFound
	} else {
		res, err := DB.Query(`
		SELECT * FROM threads WHERE id = $1 OR slug = $2`,
			id, slug)
		defer res.Close()
		if err != nil {
			return models.Thread{},
				errors.Wrap(err, "GetThread error"), http.StatusNotFound
		}
		if res.Next() {
			text := pgtype.Text{}
			res.Scan(&t.Author, &t.Created, &t.Forum,
				&t.ID, &t.Message, &text, &t.Title, &t.Votes)
			t.Slug = text.String
			return t, nil, http.StatusOK
		}
		return models.Thread{},
			errors.New("GetThread error"), http.StatusNotFound
	}
}

func UpdateThread(thread models.Thread,
	update models.Thread) (models.Thread, error, int) {
	if update.Message == "" && update.Title == "" {
		return thread, nil, http.StatusOK
	}
	sql := "UPDATE threads SET"
	if update.Message == "" {
		sql += " message = message,"
	} else {
		sql += " message = '" + update.Message + "',"
	}
	if update.Title == "" {
		sql += " title = title"
	} else {
		sql += " title = '" + update.Title + "'"
	}
	sql += " WHERE slug = '" + thread.Slug + "'"
	res, err := DB.Exec(sql)
	if err != nil {
		return models.Thread{},
			errors.Wrap(err, "UpdateThread error"), http.StatusConflict
	}
	if res.RowsAffected() == 0 {
		return models.Thread{},
			errors.New("UpdateThread error"), http.StatusNotFound
	}
	result, _, _ := GetThread(-1, thread.Slug)
	return result, nil, http.StatusOK
}

func UpdateVote(nickname string, threadId int32, vote int8) (models.Vote, error) {
	res, err := DB.Exec(`
		UPDATE votes SET voice = $1 WHERE nickname = $2 AND thread = $3`,
		vote, nickname, threadId)
	if err != nil {
		return models.Vote{}, errors.Wrap(err, "UpdateVote error")
	}
	if res.RowsAffected() == 0 {
		return models.Vote{}, errors.New("UpdateVote error")
	}
	return models.Vote{
		Nickname: nickname,
		Voice:    vote,
		Thread:   threadId,
	}, nil
}

func UpdateThreadVote(threadId int32, vote int8) (models.Thread, error, int) {
	tx, _ := DB.Begin()
	defer tx.Rollback()
	thread := models.Thread{}
	slug := &pgtype.Varchar{}
	err := tx.QueryRow(`
	UPDATE threads SET votes = votes+$1 
	WHERE id = $2 
	RETURNING author, created, forum, "message", slug, title, id, votes`,
		vote, threadId).
		Scan(&thread.Author, &thread.Created,
			&thread.Forum, &thread.Message,
			slug, &thread.Title,
			&thread.ID, &thread.Votes)
	thread.Slug = slug.String
	if err != nil {
		return models.Thread{}, errors.New("GetThread error"), http.StatusNotFound
	}
	tx.Commit()
	return thread, nil, http.StatusOK
}

func GetVote(nickname string, threadId int32) (models.Vote, error) {
	res, err := DB.Query(`
		SELECT * FROM votes WHERE nickname = $1 AND thread = $2`,
		nickname, threadId)
	if err != nil {
		return models.Vote{},
			errors.Wrap(err, "GetVote error")
	}
	defer res.Close()
	vote := models.Vote{}
	if res.Next() {
		err := res.Scan(&vote.Nickname, &vote.Voice, &vote.Thread)
		if err != nil {
			return models.Vote{},
				errors.Wrap(err, "GetVote error")
		}
		return vote, nil
	}
	return models.Vote{}, errors.New("GetVote error")
}

func CreateVote(vote models.Vote) (models.Thread, error, int) {
	voice := vote.Voice
	result, _ := DB.Exec(`
		INSERT INTO votes (nickname, voice, thread) VALUES ($1, $2, $3)`,
		vote.Nickname, vote.Voice, vote.Thread)

	if result.RowsAffected() == 0 {
		oldVote, _ := GetVote(vote.Nickname, vote.Thread)
		vote, _ = UpdateVote(vote.Nickname, vote.Thread, vote.Voice)
		if vote.Voice == -1 && vote.Voice != oldVote.Voice {
			voice = -2
		} else if vote.Voice == 1 && vote.Voice != oldVote.Voice {
			voice = 2
		} else if vote.Voice == oldVote.Voice {
			voice = 0
		}
	}
	thread, err, status := UpdateThreadVote(vote.Thread, voice)
	if err != nil {
		return models.Thread{}, errors.Wrap(err, "cant update thread"), status
	}
	return thread, nil, http.StatusOK
}
