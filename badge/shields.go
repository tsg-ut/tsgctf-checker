package badge

import (
	"fmt"
	"strings"
	"time"

	"github.com/tsg-ut/tsgctf-checker/checker"
)

func toShieldsString(s string) string {
	s = strings.ReplaceAll(s, "-", "--")
	s = strings.ReplaceAll(s, "_", "__")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// Convert to shields.io URL.
// eg: https://img.shields.io/badge/<LABEL>-<MESSAGE>-<COLOR>
func toShieldsUrl(label string, message string, color string) string {
	label = toShieldsString(label)
	message = toShieldsString(message)
	value := fmt.Sprintf("%s-%s-%s", label, message, color)
	return fmt.Sprintf("https://img.shields.io/badge/%s", value)
}

func GetBadge(chall_name string, result checker.TestResult, timestamp time.Time) (string, error) {
	label := result.ToMessage()
	message := fmt.Sprintf("%s", timestamp.Format("01/02 15:04:05 UTC"))
	color := result.ToColor()
	url := toShieldsUrl(label, message, color)

	return url, nil
}
