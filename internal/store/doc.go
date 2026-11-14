// Package store keeps what the rooms produce: the event log of every room
// (EventStore) and a record of every finished game (GameArchive), in memory
// (Memory) or in MySQL (subpackage mysql). Events are stored with the codec in
// codec.go, which reads the engine's event registry. See docs/database.md.
package store
