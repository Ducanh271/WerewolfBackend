// Command demo starts the server and five scripted bots that play a short
// game, printing the voice lines and public announcements so a game can be
// watched in the terminal. Run it with `make demo`.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	_ "masoi/internal/roles"
	"masoi/internal/room"
	"masoi/internal/transport/ws"
	"masoi/internal/transport/ws/wsclient"
)

var nicknames = []string{"An", "Binh", "Chi", "Dung", "Em"}

// defaultRoles is the five-player game; villagers fill whatever the list leaves.
const defaultRoles = "Werewolf=1,Guard=1,Seer=1"

var (
	playersFlag = flag.Int("players", 5, "number of bots at the table")
	speedFlag   = flag.Float64("speed", 1, "the room clock runs this many times faster than real time")
	rolesFlag   = flag.String("roles", defaultRoles, "non-villager roles, e.g. CursedWolf=1,WolfSeer=1,Seer=1; Villager fills the rest")
)

func main() {
	flag.Parse()
	count := *playersFlag
	roleList, err := parseRoles(*rolesFlag, count)
	if err != nil {
		log.Fatal(err)
	}
	nicknames = makeNicknames(count)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	if *speedFlag <= 0 {
		log.Fatal("-speed must be positive")
	}
	var clock room.Clock = room.RealClock{}
	if *speedFlag != 1 {
		clock = newScaledClock(*speedFlag)
	}
	rooms := room.NewManager(room.ManagerConfig{Clock: clock})
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
	fmt.Printf("Phòng %s trên %s, %d bot vào chơi.\n\n", code, base, count)

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

	startGame(ctx, bots, roleList)
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
	if path := os.Getenv("DEMO_DUMP"); path != "" {
		dump(path, bots)
	}
}

// startGame is the host: seats, roles, short times, deal.
func startGame(ctx context.Context, bots []*bot, roleList map[string]int) {
	host := bots[0]
	order := make([]string, len(bots))
	for i, b := range bots {
		order[i] = b.id
	}
	send := func(name string, args map[string]any) {
		must(host.c.Send(ctx, map[string]any{"type": "command", "name": name, "args": args}))
	}
	send("setSeats", map[string]any{"order": order})
	send("configureRoles", map[string]any{"roles": roleList})
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

// parseRoles reads "Role=n,Role=n" and fills the table up with villagers.
func parseRoles(spec string, players int) (map[string]int, error) {
	out := map[string]int{}
	total := 0
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, num, ok := strings.Cut(part, "=")
		n := 1
		if ok {
			var err error
			if n, err = strconv.Atoi(num); err != nil {
				return nil, fmt.Errorf("vai %q: số lượng không hợp lệ", part)
			}
		}
		out[name] += n
		total += n
	}
	if total > players {
		return nil, fmt.Errorf("danh sách vai có %d vai, nhiều hơn %d người", total, players)
	}
	out["Villager"] += players - total
	return out, nil
}

// makeNicknames gives the first five the classic names, then P6, P7, ...
func makeNicknames(n int) []string {
	out := make([]string, n)
	for i := range out {
		if i < 5 {
			out[i] = []string{"An", "Binh", "Chi", "Dung", "Em"}[i]
		} else {
			out[i] = "P" + strconv.Itoa(i+1)
		}
	}
	return out
}

// dump writes what every phone received (raw frames), what each bot sent and
// the errors the server answered, as JSON lines, so a game can be inspected.
func dump(path string, bots []*bot) {
	f, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, b := range bots {
		snap := b.c.Snapshot()
		_ = enc.Encode(map[string]any{"bot": b.nick, "role": b.role, "sent": b.sent, "errors": snap.Errors, "frames": snap.Frames})
	}
}
