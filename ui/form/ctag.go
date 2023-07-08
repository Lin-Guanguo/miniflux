// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
)

// CTagForm represents a feed form in the UI
type CTagForm struct {
	Title string
}

// NewCTagForm returns a new CTagForm.
func NewCTagForm(r *http.Request) *CTagForm {
	return &CTagForm{
		Title: r.FormValue("title"),
	}
}
