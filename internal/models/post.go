package models

import "time"

//easyjson:json
type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created,omitempty"`
	Forum    string    `json:"forum,omitempty"`
	ID       int64     `json:"id,omitempty"`
	IsEdited bool      `json:"isEdited,omitempty"`
	Message  string    `json:"message"`
	Parent   int64     `json:"parent,omitempty"`
	Thread   int32     `json:"thread,omitempty"`
	Path     []int64   `json:"-"`
}

//easyjson:json
type Posts []Post

//easyjson:json
type PostData struct {
	Author *User   `json:"author,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Post   *Post   `json:"post,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
}
