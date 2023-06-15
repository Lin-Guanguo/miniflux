package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"miniflux.app/model"
)

// AnotherCTagExists checks if another ctag exists with the same title.
func (s *Storage) AnotherCTagExists(userID, ctagID int64, title string) bool {
	var result bool
	query := `SELECT true FROM ctags WHERE user_id=$1 AND id != $2 AND lower(title)=lower($3) LIMIT 1`
	s.db.QueryRow(query, userID, ctagID, title).Scan(&result)
	return result
}

// CTagTitleExists checks if the given ctag exists into the database.
func (s *Storage) CTagTitleExists(userID int64, title string) bool {
	var result bool
	query := `SELECT true FROM ctags WHERE user_id=$1 AND lower(title)=lower($2) LIMIT 1`
	s.db.QueryRow(query, userID, title).Scan(&result)
	return result
}

// CTagIDExists checks if the given ctag exists into the database.
func (s *Storage) CTagIDExists(userID, ctagID int64) bool {
	var result bool
	query := `SELECT true FROM ctags WHERE user_id=$1 AND id=$2`
	s.db.QueryRow(query, userID, ctagID).Scan(&result)
	return result
}

// CTag returns a ctag from the database.
func (s *Storage) CTag(userID, ctagID int64) (*model.CTag, error) {
	var ctag model.CTag

	query := `SELECT id, user_id, title FROM ctags WHERE user_id=$1 AND id=$2`
	err := s.db.QueryRow(query, userID, ctagID).Scan(&ctag.ID, &ctag.UserID, &ctag.Title)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch ctag: %v`, err)
	default:
		return &ctag, nil
	}
}

// CTagByTitle finds a ctag by the title.
func (s *Storage) CTagByTitle(userID int64, title string) (*model.CTag, error) {
	var ctag model.CTag

	query := `SELECT id, user_id, title FROM ctags WHERE user_id=$1 AND title=$2`
	err := s.db.QueryRow(query, userID, title).Scan(&ctag.ID, &ctag.UserID, &ctag.Title)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch ctag: %v`, err)
	default:
		return &ctag, nil
	}
}

// CTags returns all ctags that belongs to the given user.
func (s *Storage) CTags(userID int64) (model.CTags, error) {
	query := `SELECT id, user_id, title FROM ctags WHERE user_id=$1 ORDER BY title ASC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch ctags: %v`, err)
	}
	defer rows.Close()

	ctags := make(model.CTags, 0)
	for rows.Next() {
		var ctag model.CTag
		if err := rows.Scan(&ctag.ID, &ctag.UserID, &ctag.Title); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch ctag row: %v`, err)
		}

		ctags = append(ctags, &ctag)
	}

	return ctags, nil
}

func (s *Storage) CTagsTree(userID int64) (model.CTagsTreeRoot, error) {
	ctags, err := s.CTags(userID)
	if err != nil {
		return nil, err
	}

	type TagInfo struct {
		ID         int64
		Title      string
		TitleParts []string
	}

	ctagsInfo := make([]TagInfo, 0)
	for _, ctag := range ctags {
		ctagsInfo = append(ctagsInfo, TagInfo{
			ID:         ctag.ID,
			Title:      ctag.Title,
			TitleParts: strings.Split(ctag.Title, "/"),
		})
	}

	// sort ctagsInfo by part len
	sort.SliceStable(ctagsInfo, func(i, j int) bool {
		return len(ctagsInfo[i].TitleParts) < len(ctagsInfo[j].TitleParts)
	})

	root := make(model.CTagsTreeRoot, 0)
	dict := make(map[string]*model.CTagsTree)

	for _, ctagInfo := range ctagsInfo {
		treeNode := &model.CTagsTree{
			ID:       ctagInfo.ID,
			Title:    ctagInfo.TitleParts[len(ctagInfo.TitleParts)-1],
			Children: []*model.CTagsTree{},
		}

		if len(ctagInfo.TitleParts) == 1 {
			root = append(root, treeNode)
			dict[ctagInfo.Title] = treeNode
			continue
		}

		parent := ctagInfo.Title[0:strings.LastIndex(ctagInfo.Title, "/")]
		if _, ok := dict[parent]; !ok {
			return nil, fmt.Errorf("store: unable to find parent tag %q for %q", parent, ctagInfo.Title)
		}
		dict[parent].Children = append(dict[parent].Children, treeNode)
		dict[ctagInfo.Title] = treeNode
	}

	return root, nil
}

// CreateCTag creates a new ctag.
func (s *Storage) CreateCTag(userID int64, request *model.CTagRequest) (*model.CTag, error) {
	var ctag model.CTag

	query := `
		INSERT INTO ctags
			(user_id, title)
		VALUES
			($1, $2)
		RETURNING
			id,
			user_id,
			title
	`
	err := s.db.QueryRow(
		query,
		userID,
		request.Title,
	).Scan(
		&ctag.ID,
		&ctag.UserID,
		&ctag.Title,
	)

	if err != nil {
		return nil, fmt.Errorf(`store: unable to create ctag %q: %v`, request.Title, err)
	}

	return &ctag, nil
}

// UpdateCTag updates an existing ctag.
func (s *Storage) UpdateCTag(ctag *model.CTag) error {
	query := `UPDATE ctags SET title=$1 WHERE id=$2 AND user_id=$3`
	_, err := s.db.Exec(
		query,
		ctag.Title,
		ctag.ID,
		ctag.UserID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update ctag: %v`, err)
	}

	return nil
}

// RemoveCTag deletes a ctag.
func (s *Storage) RemoveCTag(userID, ctagID int64) error {
	query := `DELETE FROM ctags WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, ctagID, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to remove this ctag: %v`, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to remove this ctag: %v`, err)
	}

	if count == 0 {
		return errors.New(`store: no ctag has been removed`)
	}

	return nil
}

// Add entry custom tags
func (s *Storage) AddEntryCTags(entryCTagsReq *model.EntryCTagsRequest) ([]*model.EntryCTag, error) {
	var entryCTags []*model.EntryCTag

	query := `
		INSERT INTO entry_ctags
			(entry_id, ctag_id)
		VALUES
	`
	values := []interface{}{}
	index := 1

	for _, entryCTagReq := range entryCTagsReq.EntryCTags {
		query += fmt.Sprintf("($%d, $%d),", index, index+1)
		values = append(values, entryCTagReq.EntryID, entryCTagReq.CTagID)
		index += 2
	}

	query = query[:len(query)-1] + " RETURNING id, entry_id, ctag_id"

	rows, err := s.db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to create entry custom tag relations: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entryCTag model.EntryCTag
		err := rows.Scan(&entryCTag.ID, &entryCTag.EntryID, &entryCTag.CTagID)
		if err != nil {
			return nil, fmt.Errorf("store: unable to scan entry custom tag row: %v", err)
		}
		entryCTags = append(entryCTags, &entryCTag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: error occurred while iterating entry custom tag rows: %v", err)
	}

	return entryCTags, nil
}

// Remove entry custom tags
func (s *Storage) RemoveEntryCTags(entryCTagsReq *model.EntryCTagsRequest) error {
	query := `DELETE FROM entry_ctags WHERE (entry_id, ctag_id) IN (`
	values := []interface{}{}
	index := 1

	for _, entryCTagReq := range entryCTagsReq.EntryCTags {
		query += fmt.Sprintf("($%d, $%d),", index, index+1)
		values = append(values, entryCTagReq.EntryID, entryCTagReq.CTagID)
		index += 2
	}

	query = query[:len(query)-1] + ")"

	result, err := s.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("store: unable to remove entry custom tags: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: unable to retrieve affected rows: %v", err)
	}

	if count == 0 {
		return errors.New("store: no entry custom tags have been removed")
	}

	return nil
}

// Get custom tags on entry
func (s *Storage) GetEntryCTags(userID, entryID int64) (model.CTags, error) {
	query := `
		SELECT t.id, t.user_id, t.title
			FROM entry_ctags et
			JOIN ctags t ON et.ctag_id = t.id
		WHERE t.user_id = $1 AND et.entry_id = $2;
	`
	rows, err := s.db.Query(query, userID, entryID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch ctags: %v`, err)
	}
	defer rows.Close()

	ctags := make(model.CTags, 0)
	for rows.Next() {
		var ctag model.CTag
		if err := rows.Scan(&ctag.ID, &ctag.UserID, &ctag.Title); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch ctag row: %v`, err)
		}

		ctags = append(ctags, &ctag)
	}

	return ctags, nil
}
