package i18n

import (
	"sort"
	"strings"
)

// Error texts (docs/api/errors.md). The server sends only an error's code;
// the phone shows the text for it, so the texts live here with every other
// string a player sees. A code's text is under "error.<code>". A code that
// means two things at the table has a second text under
// "error.<code>.<situation>", and the client picks it from the message it
// sent. A code with no text of its own gets ErrorGeneric: only a buggy client
// or a bug in the server causes it.
const ErrorGeneric = "error.generic"

// Situations of a code with a second text.
const (
	// SituationJoin: wrong_phase answering a join, the game has started.
	SituationJoin = "join"
	// SituationToken: bad_credentials answering a rejoin with the saved
	// token, which another phone or the host made stop working.
	SituationToken = "token"
)

// errorAliases are codes whose text is another string the player already
// knows, so the two always read the same.
var errorAliases = map[string]string{
	"blocked": NightBlocked, // the notice on the blocked player's own turn
}

const errorPrefix = "error."

// ErrorText is the text of code in situation ("" for the usual one), args
// filling its %s. own is false when the code has no text of its own and the
// text is ErrorGeneric's.
func ErrorText(code, situation string, args ...string) (text string, own bool) {
	key, own := errorKey(code, situation)
	return Vi(key, args...), own
}

// ErrorTemplate is ErrorText before its %s are filled, as docs/api/errors.md
// prints it.
func ErrorTemplate(code, situation string) (template string, own bool) {
	key, own := errorKey(code, situation)
	return vi[key], own
}

func errorKey(code, situation string) (key string, own bool) {
	if situation != "" {
		if k := errorPrefix + code + "." + situation; vi[k] != "" {
			return k, true
		}
	}
	if k, ok := errorAliases[code]; ok {
		return k, true
	}
	if k := errorPrefix + code; vi[k] != "" {
		return k, true
	}
	return ErrorGeneric, false
}

// ErrorKeys lists every own error text as "code" or "code.situation",
// sorted, aliases included: what docs/api/errors.md must account for.
func ErrorKeys() []string {
	var keys []string
	for k := range vi {
		if strings.HasPrefix(k, errorPrefix) && k != ErrorGeneric {
			keys = append(keys, strings.TrimPrefix(k, errorPrefix))
		}
	}
	for code := range errorAliases {
		keys = append(keys, code)
	}
	sort.Strings(keys)
	return keys
}
