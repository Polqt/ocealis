package util

import "github.com/microcosm-cc/bluemonday"

// policy is created once and reused - bluemonday recommends this for performance.
var policy = bluemonday.StrictPolicy()

// Sanitize message strips all HTML tags from user-submitted message text.
// Input: raw string from the client, which may contain HTML tags or other potentially harmful content.
// Output: plain text safe to store and render
func SanitizeMessage(input string) string {
	return policy.Sanitize(input)
}