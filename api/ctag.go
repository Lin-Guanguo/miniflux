package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
)

func (h *handler) createCTag(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var ctagRequest model.CTagRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&ctagRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	// if validationErr := validator.ValidateCTagCreation(h.store, userID, &ctagRequest); validationErr != nil {
	// 	json.BadRequest(w, r, validationErr.Error())
	// 	return
	// }

	ctag, err := h.store.CreateCTag(userID, &ctagRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, ctag)
}

func (h *handler) updateCTag(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	ctagID := request.RouteInt64Param(r, "ctagID")

	ctag, err := h.store.CTag(userID, ctagID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if ctag == nil {
		json.NotFound(w, r)
		return
	}

	var ctagRequest model.CTagRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&ctagRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	// if validationErr := validator.ValidateCTagModification(h.store, userID, ctag.ID, &ctagRequest); validationErr != nil {
	// 	json.BadRequest(w, r, validationErr.Error())
	// 	return
	// }

	ctagRequest.Patch(ctag)
	err = h.store.UpdateCTag(ctag)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, ctag)
}

func (h *handler) getCTags(w http.ResponseWriter, r *http.Request) {
	ctags, err := h.store.CTags(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, ctags)
}

func (h *handler) removeCTag(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	ctagID := request.RouteInt64Param(r, "ctagID")

	if !h.store.CTagIDExists(userID, ctagID) {
		json.NotFound(w, r)
		return
	}

	if err := h.store.RemoveCTag(userID, ctagID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) addEntryCTags(w http.ResponseWriter, r *http.Request) {
	// userID := request.UserID(r)

	var entryCTagsRequest model.EntryCTagsRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entryCTagsRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	// if validationErr := validator.ValidateCTagCreation(h.store, userID, &ctagRequest); validationErr != nil {
	// 	json.BadRequest(w, r, validationErr.Error())
	// 	return
	// }

	entryCTags, err := h.store.AddEntryCTags(&entryCTagsRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, entryCTags)
}

func (h *handler) removeEntryCTags(w http.ResponseWriter, r *http.Request) {
	// userID := request.UserID(r)

	var entryCTagsRequest model.EntryCTagsRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entryCTagsRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	err := h.store.RemoveEntryCTags(&entryCTagsRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) getEntryCTags(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	entryID := request.RouteInt64Param(r, "entryID")

	ctags, err := h.store.GetEntryCTags(userID, entryID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, ctags)
}
