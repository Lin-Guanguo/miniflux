// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"fmt"
	"net/http"

	"miniflux.app/http/response/html"
)

func (h *handler) markCTagAsRead(w http.ResponseWriter, r *http.Request) {
	// userID := request.UserID(r)
	// categoryID := request.RouteInt64Param(r, "categoryID")

	// category, err := h.store.Category(userID, categoryID)
	// if err != nil {
	// 	html.ServerError(w, r, err)
	// 	return
	// }

	// if category == nil {
	// 	html.NotFound(w, r)
	// 	return
	// }

	// if err = h.store.MarkCategoryAsRead(userID, categoryID, time.Now()); err != nil {
	// 	html.ServerError(w, r, err)
	// 	return
	// }

	// html.Redirect(w, r, route.Path(h.router, "categories"))
	// TODO:
	html.ServerError(w, r, fmt.Errorf("TODO: unimplemented"))
}
