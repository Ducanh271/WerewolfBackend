package enginetest

import "masoi/internal/engine"

// Announcement is implemented by events the narrator reads aloud. The key is
// the i18n key of the message (internal/i18n), so scenarios do not depend on
// the Vietnamese wording. The dawn task (T6) makes its events implement it.
type Announcement interface {
	AnnouncementKey() string
}

// announcementKeys lists, in order, the keys of the announcements in events.
func announcementKeys(events []engine.Event) []string {
	var keys []string
	for _, e := range events {
		if a, ok := e.(Announcement); ok {
			keys = append(keys, a.AnnouncementKey())
		}
	}
	return keys
}
