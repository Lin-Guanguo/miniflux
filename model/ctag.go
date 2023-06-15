package model // import "miniflux.app/model"

import "fmt"

// CTag represents a entry custom tag.
type CTag struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
}

func (c *CTag) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s", c.ID, c.UserID, c.Title)
}

// EntryCTag represents entry custom tag relation.
type EntryCTag struct {
	ID      int64 `json:"id"`
	EntryID int64 `json:"entry_id"`
	CTagID  int64 `json:"ctag_id"`
}

// CTagRequest represents the request to create or update a custom tag.
type CTagRequest struct {
	Title string `json:"title"`
}

// Patch updates custom tag fields.
func (cr *CTagRequest) Patch(ctag *CTag) {
	ctag.Title = cr.Title
}

type EntryCTagRequest struct {
	EntryID int64 `json:"entry_id"`
	CTagID  int64 `json:"ctag_id"`
}

type EntryCTagsRequest struct {
	EntryCTags []EntryCTagRequest `json:"entry_ctags"`
}

type CTags []*CTag

type EntryCTags []*EntryCTag

type CTagsTree struct {
	ID       int64        `json:"id"`
	Title    string       `json:"title"`
	Children []*CTagsTree `json:"children"`
}

type CTagsTreeRoot []*CTagsTree
