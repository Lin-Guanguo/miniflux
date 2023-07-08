// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package validator // import "miniflux.app/validator"

import (
	"strings"

	"miniflux.app/model"
	"miniflux.app/storage"
)

// ValidateCTagCreation validates ctag creation.
func ValidateCTagCreation(store *storage.Storage, userID int64, request *model.CTagRequest) *ValidationError {
	if request.Title == "" {
		return NewValidationError("error.title_required")
	}

	if store.CTagTitleExists(userID, request.Title) {
		return NewValidationError("error.ctag_already_exists")
	}

	titleParts := strings.Split(request.Title, "/")
	for i := 1; i < len(titleParts); i++ {
		if !store.CTagTitleExists(userID, strings.Join(titleParts[:i], "/")) {
			return NewValidationError("error.ctag_parent_not_exists")
		}
	}

	return nil
}

// ValidateCTagModification validates ctag modification.
func ValidateCTagModification(store *storage.Storage, userID, ctagID int64, request *model.CTagRequest) *ValidationError {
	if request.Title == "" {
		return NewValidationError("error.title_required")
	}

	if store.AnotherCTagExists(userID, ctagID, request.Title) {
		return NewValidationError("error.ctag_already_exists")
	}

	if store.CTagChildExists(userID, ctagID) {
		return NewValidationError("error.ctag_child_exists")
	}

	titleParts := strings.Split(request.Title, "/")
	for i := 1; i < len(titleParts); i++ {
		if !store.CTagTitleExists(userID, strings.Join(titleParts[:i], "/")) {
			return NewValidationError("error.ctag_parent_not_exists")
		}
	}

	return nil
}

// ValidateCTagRemoval validates ctag removal.
func ValidateCTagRemoval(store *storage.Storage, userID, ctagID int64) *ValidationError {
	if store.CTagChildExists(userID, ctagID) {
		return NewValidationError("error.ctag_child_exists")
	}
	return nil
}
