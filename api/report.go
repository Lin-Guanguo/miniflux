package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
)

func (h *handler) report(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var reportRequest model.ReportRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&reportRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	report, err := h.store.Report(userID, &reportRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, report)
}
