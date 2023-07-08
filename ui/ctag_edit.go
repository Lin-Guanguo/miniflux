// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showEditCTagPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
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

	ctagForm := form.CTagForm{
		Title: ctag.Title,
	}

	view.Set("form", ctagForm)
	view.Set("ctag", ctag)
	view.Set("menu", "ctags")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	html.OK(w, r, view.Render("edit_ctag"))
}
