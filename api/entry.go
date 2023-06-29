// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
	"miniflux.app/proxy"
	"miniflux.app/reader/processor"
	"miniflux.app/storage"
	"miniflux.app/url"
	"miniflux.app/validator"
)

func (h *handler) getEntryFromBuilder(w http.ResponseWriter, r *http.Request, b *storage.EntryQueryBuilder) {
	entry, err := b.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	entry.Content = proxy.AbsoluteProxyRewriter(h.router, r.Host, entry.Content)
	proxyOption := config.Opts.ProxyOption()

	for i := range entry.Enclosures {
		if proxyOption == "all" || proxyOption != "none" && !url.IsHTTPS(entry.Enclosures[i].URL) {
			for _, mediaType := range config.Opts.ProxyMediaTypes() {
				if strings.HasPrefix(entry.Enclosures[i].MimeType, mediaType+"/") {
					entry.Enclosures[i].URL = proxy.AbsoluteProxifyURL(h.router, r.Host, entry.Enclosures[i].URL)
					break
				}
			}
		}
	}

	json.OK(w, r, entry)
}

func (h *handler) getFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getCategoryEntry(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getFeedEntries(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	h.findEntries(w, r, feedID, 0, 0, "")
}

func (h *handler) getCategoryEntries(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	h.findEntries(w, r, 0, categoryID, 0, "")
}

func (h *handler) getCTagEntries(w http.ResponseWriter, r *http.Request) {
	ctagID := request.RouteInt64Param(r, "ctagID")
	h.findEntries(w, r, 0, 0, ctagID, "")
}

func (h *handler) getEntries(w http.ResponseWriter, r *http.Request) {
	h.findEntries(w, r, 0, 0, 0, "")
}

func (h *handler) searchEntries(w http.ResponseWriter, r *http.Request) {
	var entriesSearchRequest model.EntriesSearchRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesSearchRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}
	h.findEntries(w, r, 0, 0, 0, entriesSearchRequest.Condition)
}

func (h *handler) searchEntriesEnclosures(w http.ResponseWriter, r *http.Request) {
	var entriesSearchRequest model.EntriesSearchRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesSearchRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}
	h.findEntriesEnclosures(w, r, 0, 0, 0, entriesSearchRequest.Condition)
}

func (h *handler) buildQuery(r *http.Request, feedID int64, categoryID int64, ctagID int64, condition string) (*storage.EntryQueryBuilder, error) {
	statuses := request.QueryStringParamList(r, "status")
	for _, status := range statuses {
		if err := validator.ValidateEntryStatus(status); err != nil {
			return nil, err
		}
	}

	order := request.QueryStringParam(r, "order", model.DefaultSortingOrder)
	if err := validator.ValidateEntryOrder(order); err != nil {
		return nil, err
	}

	direction := request.QueryStringParam(r, "direction", model.DefaultSortingDirection)
	if err := validator.ValidateDirection(direction); err != nil {
		return nil, err
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := validator.ValidateRange(offset, limit); err != nil {
		return nil, err
	}

	userID := request.UserID(r)
	categoryID = request.QueryInt64Param(r, "category_id", categoryID)
	if categoryID > 0 && !h.store.CategoryIDExists(userID, categoryID) {
		return nil, errors.New("Invalid category ID")
	}

	feedID = request.QueryInt64Param(r, "feed_id", feedID)
	if feedID > 0 && !h.store.FeedExists(userID, feedID) {
		return nil, errors.New("Invalid feed ID")
	}

	ctagID = request.QueryInt64Param(r, "ctag_id", ctagID)
	if ctagID > 0 && !h.store.CTagIDExists(userID, ctagID) {
		return nil, errors.New("Invalid CTag ID")
	}

	tags := request.QueryStringParamList(r, "tags")

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
	builder.WithCategoryID(categoryID)
	builder.WithCTagID(ctagID)
	builder.WithCondition(condition)
	builder.WithStatuses(statuses)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	builder.WithTags(tags)
	configureFilters(builder, r)
	return builder, nil
}

func (h *handler) findEntries(w http.ResponseWriter, r *http.Request, feedID int64, categoryID int64, ctagID int64, condition string) {
	builder, err := h.buildQuery(r, feedID, categoryID, ctagID, condition)
	if err != nil {
		json.BadRequest(w, r, err)
	}

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	for i := range entries {
		entries[i].Content = proxy.AbsoluteProxyRewriter(h.router, r.Host, entries[i].Content)
	}

	json.OK(w, r, &entriesResponse{Total: count, Entries: entries})
}

func (h *handler) findEntriesEnclosures(w http.ResponseWriter, r *http.Request, feedID int64, categoryID int64, ctagID int64, condition string) {
	builder, err := h.buildQuery(r, feedID, categoryID, ctagID, condition)
	if err != nil {
		json.BadRequest(w, r, err)
	}

	entriesEnclosures, err := builder.GetEntriesEnclosures()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntriesEnclosures()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, &entriesEnclosuresResponse{Total: count, EntriesEnclosures: entriesEnclosures})
}

func (h *handler) setEntryStatus(w http.ResponseWriter, r *http.Request) {
	var entriesStatusUpdateRequest model.EntriesStatusUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntriesStatusUpdateRequest(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, entriesStatusUpdateRequest.Status); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) toggleBookmark(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if err := h.store.ToggleBookmark(request.UserID(r), entryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) fetchContent(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	entryID := request.RouteInt64Param(r, "entryID")

	entryBuilder := h.store.NewEntryQueryBuilder(loggedUserID)
	entryBuilder.WithEntryID(entryID)
	entryBuilder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := entryBuilder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	user, err := h.store.UserByID(entry.UserID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	feedBuilder := storage.NewFeedQueryBuilder(h.store, loggedUserID)
	feedBuilder.WithFeedID(entry.FeedID)
	feed, err := feedBuilder.GetFeed()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if feed == nil {
		json.NotFound(w, r)
		return
	}

	if err := processor.ProcessEntryWebPage(feed, entry, user); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, map[string]string{"content": entry.Content})
}

func configureFilters(builder *storage.EntryQueryBuilder, r *http.Request) {
	beforeEntryID := request.QueryInt64Param(r, "before_entry_id", 0)
	if beforeEntryID > 0 {
		builder.BeforeEntryID(beforeEntryID)
	}

	afterEntryID := request.QueryInt64Param(r, "after_entry_id", 0)
	if afterEntryID > 0 {
		builder.AfterEntryID(afterEntryID)
	}

	beforeTimestamp := request.QueryInt64Param(r, "before", 0)
	if beforeTimestamp > 0 {
		builder.BeforeDate(time.Unix(beforeTimestamp, 0))
	}

	afterTimestamp := request.QueryInt64Param(r, "after", 0)
	if afterTimestamp > 0 {
		builder.AfterDate(time.Unix(afterTimestamp, 0))
	}

	categoryID := request.QueryInt64Param(r, "category_id", 0)
	if categoryID > 0 {
		builder.WithCategoryID(categoryID)
	}

	if request.HasQueryParam(r, "starred") {
		starred, err := strconv.ParseBool(r.URL.Query().Get("starred"))
		if err == nil {
			builder.WithStarred(starred)
		}
	}

	searchQuery := request.QueryStringParam(r, "search", "")
	if searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
