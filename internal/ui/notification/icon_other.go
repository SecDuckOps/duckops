//go:build !darwin

package notification

import (
	_ "embed"
)

//go:embed duckops-icon.png
var icon []byte

// Icon contains the embedded PNG icon data for desktop notifications.
var Icon any = icon
