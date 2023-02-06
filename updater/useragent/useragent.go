// Copyright (C) 2023 Huntress Labs, Inc.
//
// This file is part of Huntress. Unauthorized copying of this file, via any medium is
// strictly prohibited without the express written consent of Huntress Labs, Inc.

package useragent

import (
	"fmt"
	"runtime"
)

const (
	WSUpdaterServiceName = "Huntress-WSUpdater"
	WSUPDATER_VER        = "NA"
)

// Returns a formatted string suitable for usage as a User-Agent header for HTTP requests
func GetUserAgentString() string {
	// Follow RFC7231 Spec
	// Ref: https://www.rfc-editor.org/rfc/rfc7231#section-5.5.3
	// Ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	return fmt.Sprintf("%s/%s (%s; %s)",
		WSUpdaterServiceName, WSUPDATER_VER, runtime.GOOS, runtime.GOARCH)
}
