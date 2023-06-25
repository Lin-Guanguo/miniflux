// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showCTagEntriesPage(w http.ResponseWriter, r *http.Request) {
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

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithCTagID(ctag.ID)
	builder.WithOrder(user.EntryOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("ctag", ctag)
	view.Set("total", count)
	view.Set("entries", entries)
	view.Set("pagination", getPagination(route.Path(h.router, "ctagEntries", "ctagID", ctag.ID), count, offset, user.EntriesPerPage))
	view.Set("menu", "ctags")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("showOnlyUnreadEntries", true)

	html.OK(w, r, view.Render("ctag_entries"))
}
