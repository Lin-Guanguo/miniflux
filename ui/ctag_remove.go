// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"fmt"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/validator"
)

func (h *handler) removeCTag(w http.ResponseWriter, r *http.Request) {
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

	if validationErr := validator.ValidateCTagRemoval(h.store, user.ID, ctagID); validationErr != nil {
		html.ServerError(w, r, fmt.Errorf("remove invalidate %s", validationErr.TranslationKey))
		return
	}

	if err := h.store.RemoveCTag(user.ID, ctag.ID); err != nil {
		html.ServerError(w, r, err)
		return
	}
}
