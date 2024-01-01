// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
	"miniflux.app/validator"
)

func (h *handler) updateCTag(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	ctagID := request.RouteInt64Param(r, "ctagID")
	ctag, err := h.store.CTag(request.UserID(r), ctagID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if ctag == nil {
		html.NotFound(w, r)
		return
	}

	ctagForm := form.NewCTagForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", ctagForm)
	view.Set("ctag", ctag)
	view.Set("menu", "ctags")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))

	ctagRequest := &model.CTagRequest{
		Title: ctagForm.Title,
	}

	if validationErr := validator.ValidateCTagModification(h.store, loggedUser.ID, ctag.ID, ctagRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.TranslationKey)
		html.OK(w, r, view.Render("create_ctag"))
		return
	}

	ctagRequest.Patch(ctag)
	if err := h.store.UpdateCTag(ctag); err != nil {
		logger.Error("[UI:UpdateCTag] %v", err)
		view.Set("errorMessage", "error.unable_to_update_ctag")
		html.OK(w, r, view.Render("edit_ctag"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "ctagEntriesAll", "ctagID", ctagID))
}