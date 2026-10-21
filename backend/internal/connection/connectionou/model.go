// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package connectionou

import tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"

// Errors are the errors a kind of connection answers with when it refuses the organization unit of a request.
type Errors struct {
	// Required is returned when neither an id nor a handle is given for a new connection.
	Required tidcommon.ServiceError
	// NotFound is returned when the organization unit named does not exist.
	NotFound tidcommon.ServiceError
	// Changed is returned when the organization unit named for an existing connection is not the one it has.
	Changed tidcommon.ServiceError
}
