// Command demo starts the server and five scripted bots that play a short
// game, printing the voice lines and public announcements so a game can be
// watched in the terminal. Run it with `make demo`.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "masoi/internal/roles"
	"masoi/internal/room"
	"masoi/internal/transport/ws"
	"masoi/internal/transport/ws/wsclient"
)

var nicknames = []string{"An", "Binh", "Chi", "Dung", "Em"}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	rooms := room.NewManager(room.ManagerConfig{})
	defer rooms.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Handler: ws.New(ws.Config{Rooms: rooms, AllowAnyOrigin: true}), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()
	base := ln.Addr().String()

	code, err := createRoom(base)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Phòng %s trên %s, 5 bot vào chơi.\n\n", code, base)

	out := newPrinter()
	var bots []*bot
	for i, nick := range nicknames {
		var onFrame func(wsclient.Frame)
		if i == 0 {
			onFrame = out.host // the host phone is the speaker
		}
		c, err := wsclient.DialWith(ctx, "ws://"+base+"/ws", nil, onFrame)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		must(c.Send(ctx, map[string]any{"type": "join", "room": code, "nickname": nick}))
		if !c.Wait(ctx, func(s wsclient.Snapshot) bool { return s.Welcome != nil }) {
			log.Fatalf("%s không vào được phòng: %v", nick, c.Snapshot().Errors)
		}
		b := newBot(nick, c)
		b.id = c.Snapshot().Welcome.Str("player")
		bots = append(bots, b)
		out.names[b.id] = nick
	}

	startGame(ctx, bots)
	for _, b := range bots {
		b.role = b.c.Snapshot().View.Str("view", "me", "role")
		fmt.Printf("  (demo) %s là %s\n", b.nick, b.role)
	}
	fmt.Println()

	done := make(chan struct{})
	for _, b := range bots {
		go func() { b.run(ctx); done <- struct{}{} }()
	}
	for range bots {
		<-done
	}
	time.Sleep(500 * time.Millisecond) // let the last announcements print
	fmt.Println("\nVán đấu kết thúc.")
}

// startGame is the host: seats, roles, short times, deal.
func startGame(ctx context.Context, bots []*bot) {
	host := bots[0]
	order := make([]string, len(bots))
	for i, b := range bots {
		order[i] = b.id
	}
	send := func(name string, args map[string]any) {
		must(host.c.Send(ctx, map[string]any{"type": "command", "name": name, "args": args}))
	}
	send("setSeats", map[string]any{"order": order})
	send("configureRoles", map[string]any{"roles": map[string]int{"Werewolf": 1, "Guard": 1, "Seer": 1, "Villager": 2}})
	send("configureTimers", map[string]any{"timers": map[string]any{
		"roleView": 5000, "chiefElection": 5000, "nightStep": 5000, "announce": 5000, "day": 60000,
		"nominationLock": 5000, "defense": 5000, "verdict": 5000, "hunterShot": 5000, "chiefSuccession": 5000,
	}})
	send("dealRoles", nil)
	for _, b := range bots {
		if !b.c.Wait(ctx, func(s wsclient.Snapshot) bool { return s.View.Str("view", "me", "role") != "" }) {
			log.Fatalf("%s không nhận được vai: %v", b.nick, b.c.Snapshot().Errors)
		}
	}
}

func createRoom(base string) (string, error) {
	resp, err := http.Post("http://"+base+"/rooms", "application/json", strings.NewReader(""))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var body struct{ Room string }
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Room == "" {
		return "", fmt.Errorf("tạo phòng thất bại: %v", err)
	}
	return body.Room, nil
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
