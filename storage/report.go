package storage // import "miniflux.app/storage"

import (
	"fmt"

	"miniflux.app/model"
)

// Report message save to database
func (s *Storage) Report(userID int64, request *model.ReportRequest) (*model.Report, error) {
	var report model.Report

	query := `
		INSERT INTO reports
			(user_id, type, title, content)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			id,
			user_id,
			type,
			title,
			content,
			created_at
	`
	err := s.db.QueryRow(
		query,
		userID,
		request.Type,
		request.Title,
		request.Content,
	).Scan(
		&report.ID,
		&report.UserID,
		&report.Type,
		&report.Title,
		&report.Content,
		&report.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(`store: unable to save report %q: %v`, request.Title, err)
	}

	return &report, nil
}
