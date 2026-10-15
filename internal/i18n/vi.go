package i18n

import "strings"

// vi maps announcement keys to the Vietnamese text read aloud. %s is replaced
// by the names, if the message takes any.
var vi = map[string]string{
	DawnSomeoneDied: "Đêm qua, %s đã chết.",
	DawnProtected:   "Không có ai chết, được bảo vệ.",
	DawnNobodyDied:  "Không có ai chết.",
}

// Vi returns the Vietnamese text for key with names joined in the order given.
// An unknown key returns the key itself so a missing string is easy to spot.
func Vi(key string, names ...string) string {
	text, ok := vi[key]
	if !ok {
		return key
	}
	return strings.Replace(text, "%s", strings.Join(names, ", "), 1)
}
