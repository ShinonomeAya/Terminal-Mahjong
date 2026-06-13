# Online Mahjong Final Client Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the online terminal Mahjong client as a playable, inspectable, reconnectable TUI/CLI experience while preserving the terminal-first identity.

**Architecture:** Keep the current Go game core as the source of truth, add protocol-level room discovery, make reconnect progress observable by clients, and move server lifetime settings behind explicit options. The plan borrows architecture ideas from `palemoky/fight-the-landlord` such as Bubble Tea model states, reconnect notifications, room/lobby separation, JSON/WebSocket transport, and bot decision interfaces, but it must not copy GPL-licensed implementation code.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Gorilla WebSocket, JSON protocol messages, Go `testing`, `httptest`, current `internal/game`, `internal/online`, `internal/tui`, and `cmd/client`.

---

## Status Check

The previous broad route is partially complete, not fully complete.

Completed in the current repository:
- Fair seed/shuffle proof and replay-facing snapshots exist in `internal/game`.
- Shared command/snapshot shape exists in `internal/game/snapshot.go` and `internal/game/action.go`.
- Heuristic bot engine exists in `internal/bot`.
- In-memory WebSocket room server exists in `internal/online/server.go`.
- CLI create/join/reconnect/ready/discard/win/kong/watch flows exist in `cmd/client`.
- TUI menu, Unicode tiles, keyboard selection, mouse selection, ready flow, online game-over flow, and reconnect token flow exist in `internal/tui`.

Remaining before calling the online client complete:
- Room discovery/lobby list, so players do not need to manually pass room codes for every flow.
- Observable reconnect progress in the TUI, modeled after the reference project's reconnect event channel.
- Configurable reconnect window and room lifecycle settings, instead of hard-coded server constants.
- A written acceptance ledger and final regression command set.
- A final pass that proves TUI/CLI/server behavior still works together.

Reference project checked locally:
- Source clone used for architecture study: `C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference`
- Checked commit: `97bb24f Merge pull request #63 from palemoky/dev`
- Relevant ideas observed: Bubble Tea online model split into lobby/game phases, reconnect callbacks sent through channels, room manager and session lifecycle tests, `DecisionEngine` bot boundary, server/client commands over WebSocket, configurable security/lifetime settings.
- UI reference observed:
  - `docs/lobby.png`: centered title, menu box plus chat/status box, muted bottom controls, restrained dark table background.
  - `docs/in-game.png`: centered game board, top status/counter area, middle opponent and last-play boxes, bottom hand tray, prompt below hand.
  - `internal/ui/view/renderer.go`: uses `lipgloss.Place`, `PlaceHorizontal`, `JoinHorizontal`, `JoinVertical`, fixed box widths, and compact centered rendering.
  - `internal/ui/view/lobby.go`: uses a lobby phase with a selectable menu panel and a secondary information panel.

## Strict Execution Rules

- Execute tasks in order.
- For each task, write the failing test first, run the targeted test, implement the smallest code change, rerun the targeted test, then commit.
- Do not copy code from the reference project. Use only the architectural ideas listed in this plan.
- Do not refactor unrelated files.
- Keep single-player TUI behavior unchanged.
- After every task, run `go test ./...` before committing if the task changed Go code.
- Commit messages must match the task scope, for example `feat: list online rooms`.

## Review Cadence

Every task is a stage. Every step must end with a short written review before moving on.

After each step, write this one-line step review in the working notes or commit preparation notes:

```text
Step Review: this step advanced <stage goal> by <specific verified fact>.
```

After each task, write this stage review in the commit body or the execution log:

```text
Stage Review: Task <N> now supports <stage outcome>; remaining total-goal risk is <specific risk or none>.
```

The total goal is: a terminal-first online Mahjong client with a clean reference-inspired lobby, a readable four-seat table, smooth keyboard/mouse play, room discovery, reconnect visibility, and passing automated acceptance.

## File Structure

- `docs/online-client-acceptance.md`: Human-readable acceptance ledger for the online client.
- `docs/tui-reference-study.md`: Design translation notes from the reference screenshots and renderer code.
- `README.md`: Short user-facing description of online room discovery, reconnect, and verification commands.
- `internal/protocol/message.go`: Adds room-list protocol types and message variants.
- `internal/online/server.go`: Handles room-list requests and server lifecycle options.
- `internal/online/options.go`: Defines server option defaults for reconnect and cleanup windows.
- `internal/online/client.go`: Adds list-room API and reconnect callbacks.
- `internal/online/server_test.go`: Server protocol, room discovery, and lifecycle tests.
- `internal/online/client_test.go`: Client list-room and reconnect callback tests.
- `cmd/client/main.go`: Adds `-list` CLI flow.
- `cmd/client/main_test.go`: CLI list output tests.
- `internal/tui/model.go`: Adds online lobby state and reconnect event channel state.
- `internal/tui/menu.go`: Adds browse-online-room menu screen and list controls.
- `internal/tui/online.go`: Wires reconnect progress events into Bubble Tea commands.
- `internal/tui/network.go`: Keeps reconnect status rendering stable.
- `internal/tui/style.go`: Adds reusable table, panel, and card-like tile styles inspired by the reference UI.
- `internal/tui/layout.go`: Replaces loose table strings with centered lobby/table sections and stable four-seat tabletop rendering.
- `internal/tui/model_test.go`: TUI lobby and reconnect rendering tests.
- `internal/tui/layout_test.go`: TUI tabletop, hand tray, width, and hitbox regression tests.
- `internal/tui/style_test.go`: Style width and text preservation tests.

---

## Task 1: Acceptance Ledger And README Contract

**Files:**
- Create: `docs/online-client-acceptance.md`
- Modify: `README.md`

- [ ] **Step 1: Create the acceptance ledger**

Create `docs/online-client-acceptance.md` with this content:

```markdown
# Online Client Acceptance

This document is the acceptance checklist for the terminal-first online Mahjong client.

## Reference Boundary

The project learns architecture ideas from `palemoky/fight-the-landlord`: Bubble Tea phase separation, WebSocket room flow, reconnect progress notifications, room/session tests, and bot decision boundaries.

No GPL implementation code from that repository is copied into this repository.

## Required Behaviors

- `cmd/server` starts an in-memory WebSocket room server.
- `cmd/client -list` prints joinable waiting rooms.
- `cmd/client` creates a room and saves a reconnect session.
- `cmd/client -join <room>` joins a waiting room.
- `cmd/client -ready` broadcasts ready room state.
- `cmd/client -watch` survives one dropped socket by reconnecting with the saved token.
- The TUI main menu can create, browse, join, reconnect, return to menu, and quit.
- The TUI shows one of these network states: local, waiting, your turn, reconnecting, reconnected, offline.
- The server rejects joins after a room has started.
- The server rejects non-current-player commands.
- The server keeps disconnected sessions reconnectable within the configured reconnect window.
- Expired disconnected sessions cannot reconnect.
- Expired idle rooms are not listed in room discovery.

## Verification Commands

Run these commands from the repository root:

```powershell
go test ./...
go test ./internal/online ./cmd/client ./internal/tui
go run ./cmd/server
go run ./cmd/client -list
```

Manual smoke flow:

```powershell
go run ./cmd/server
go run ./cmd/client -name Host -session .host-session.json
go run ./cmd/client -name Guest -join <printed-room-code> -session .guest-session.json
go run ./cmd/client -session .host-session.json -reconnect -ready -watch
```
```

- [ ] **Step 2: Verify the ledger has no placeholder markers**

Run:

```powershell
$patterns = @("TB"+"D", "TO"+"DO", "implement"+" later", "fill in"+" details", "Similar"+" to", "appropriate"+" error")
foreach ($pattern in $patterns) { rg -n $pattern docs/online-client-acceptance.md }
```

Expected: no matches and exit code `1`.

- [ ] **Step 3: Update README online section**

Add this concise contract near the existing online play documentation in `README.md`:

```markdown
### Online client acceptance

The online client remains terminal-first. It supports in-memory WebSocket rooms, room discovery, ready/start synchronization, token-based reconnect, bot-filled empty seats, and CLI/TUI smoke flows. The detailed acceptance checklist lives in `docs/online-client-acceptance.md`.
```

- [ ] **Step 4: Run documentation checks**

Run:

```powershell
rg -n "Online client acceptance|docs/online-client-acceptance.md" README.md docs/online-client-acceptance.md
```

Expected: both files are reported.

- [ ] **Step 5: Commit**

Run:

```powershell
git add README.md docs/online-client-acceptance.md
git commit -m "docs: add online client acceptance ledger"
```

Expected: commit succeeds.

---

## Task 2: Room Discovery Protocol And Server Handler

**Files:**
- Modify: `internal/protocol/message.go`
- Modify: `internal/online/server.go`
- Test: `internal/online/server_test.go`

- [ ] **Step 1: Write the failing server room-list test**

Add this test to `internal/online/server_test.go`:

```go
func TestWebSocketServerListsWaitingRooms(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom, Name: "first"})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	lister := dialTestClient(t, url)
	defer lister.Close()
	sendMessage(t, lister, protocol.Message{Type: protocol.MsgListRooms})
	list := readUntil(t, lister, protocol.MsgRoomList)

	if len(list.Rooms) != 1 {
		t.Fatalf("rooms = %#v, want one waiting room", list.Rooms)
	}
	room := list.Rooms[0]
	if room.Code != created.RoomCode || room.Occupied != 1 || room.Ready != 0 || room.Started {
		t.Fatalf("room = %#v, created = %#v", room, created)
	}
}
```

- [ ] **Step 2: Run the failing test**

Run:

```powershell
go test ./internal/online -run TestWebSocketServerListsWaitingRooms -count=1
```

Expected: compile fails because `protocol.MsgListRooms`, `protocol.MsgRoomList`, and `protocol.RoomSummary` do not exist.

- [ ] **Step 3: Add protocol room-list types**

In `internal/protocol/message.go`, add these message constants:

```go
MsgListRooms    MessageType = "list_rooms"
MsgRoomList     MessageType = "room_list"
```

Add this type:

```go
type RoomSummary struct {
	Code     string `json:"code"`
	Occupied int    `json:"occupied"`
	Ready    int    `json:"ready"`
	Started  bool   `json:"started"`
	Wall     int    `json:"wall"`
}
```

Add this field to `Message`:

```go
Rooms []RoomSummary `json:"rooms,omitempty"`
```

- [ ] **Step 4: Add the server handler**

In `internal/online/server.go`, add a `protocol.MsgListRooms` case to `handleMessage`:

```go
case protocol.MsgListRooms:
	writeJSON(conn, protocol.Message{
		Type:  protocol.MsgRoomList,
		Rooms: s.roomSummaries(),
	})
```

Add this method:

```go
func (s *Server) roomSummaries() []protocol.RoomSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	rooms := make([]protocol.RoomSummary, 0, len(s.rooms))
	for _, room := range s.rooms {
		if room.started {
			continue
		}
		rooms = append(rooms, protocol.RoomSummary{
			Code:     room.code,
			Occupied: len(occupiedSeats(room)),
			Ready:    len(readySeats(room)),
			Started:  room.started,
			Wall:     room.game.Snapshot().WallCount,
		})
	}
	return rooms
}
```

- [ ] **Step 5: Run the room-list test**

Run:

```powershell
go test ./internal/online -run TestWebSocketServerListsWaitingRooms -count=1
```

Expected: pass.

- [ ] **Step 6: Run the online package tests**

Run:

```powershell
go test ./internal/online -count=1
```

Expected: pass.

- [ ] **Step 7: Commit**

Run:

```powershell
git add internal/protocol/message.go internal/online/server.go internal/online/server_test.go
git commit -m "feat: list waiting online rooms"
```

Expected: commit succeeds.

---

## Task 3: Client API And CLI Room Listing

**Files:**
- Modify: `internal/online/client.go`
- Modify: `internal/online/client_test.go`
- Modify: `cmd/client/main.go`
- Modify: `cmd/client/main_test.go`

- [ ] **Step 1: Write the failing client API test**

Add this test to `internal/online/client_test.go`:

```go
func TestClientListsWaitingRooms(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := NewClient(url+"/ws", "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	lister := NewClient(url+"/ws", "lister")
	defer lister.Close()
	rooms, err := lister.ListRooms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 || rooms[0].Code != created.RoomCode {
		t.Fatalf("rooms = %#v, created = %#v", rooms, created)
	}
}
```

- [ ] **Step 2: Run the failing client API test**

Run:

```powershell
go test ./internal/online -run TestClientListsWaitingRooms -count=1
```

Expected: compile fails because `ListRooms` does not exist.

- [ ] **Step 3: Implement `Client.ListRooms`**

Add this method to `internal/online/client.go`:

```go
func (c *Client) ListRooms(ctx context.Context) ([]protocol.RoomSummary, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	if err := c.write(protocol.Message{Type: protocol.MsgListRooms}); err != nil {
		return nil, err
	}
	message, err := c.ReadUntil(ctx, 2*time.Second, protocol.MsgRoomList, protocol.MsgError)
	if err != nil {
		return nil, err
	}
	return append([]protocol.RoomSummary(nil), message.Rooms...), nil
}
```

- [ ] **Step 4: Run the client API test**

Run:

```powershell
go test ./internal/online -run TestClientListsWaitingRooms -count=1
```

Expected: pass.

- [ ] **Step 5: Write the failing CLI list test**

Add this test to `cmd/client/main_test.go`:

```go
func TestRunListsRooms(t *testing.T) {
	server := online.NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := online.NewClient(url+"/ws", "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	err = run(context.Background(), []string{"-server", url + "/ws", "-list"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"rooms=1", created.RoomCode, "occupied=1", "ready=0"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}
```

If `cmd/client/main_test.go` does not already import `strings`, `context`, or `mahjong/internal/online`, add only the missing imports.

- [ ] **Step 6: Run the failing CLI list test**

Run:

```powershell
go test ./cmd/client -run TestRunListsRooms -count=1
```

Expected: compile fails because the `-list` flag is not implemented.

- [ ] **Step 7: Implement CLI `-list`**

In `cmd/client/main.go`, add this flag near the other flags:

```go
listRooms := flags.Bool("list", false, "list waiting rooms and exit")
```

After flag parsing and before the existing `connectCtx` block, add:

```go
if *listRooms {
	client := online.NewClient(*serverURL, *name)
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rooms, err := client.ListRooms(ctx)
	if err != nil {
		return err
	}
	printRoomList(out, rooms)
	return nil
}
```

Add this helper to `cmd/client/main.go`:

```go
func printRoomList(out io.Writer, rooms []protocol.RoomSummary) {
	fmt.Fprintf(out, "rooms=%d\n", len(rooms))
	for _, room := range rooms {
		fmt.Fprintf(out, "room=%s occupied=%d ready=%d started=%t wall=%d\n",
			room.Code,
			room.Occupied,
			room.Ready,
			room.Started,
			room.Wall,
		)
	}
}
```

- [ ] **Step 8: Run CLI tests**

Run:

```powershell
go test ./cmd/client -count=1
```

Expected: pass.

- [ ] **Step 9: Run full tests**

Run:

```powershell
go test ./...
```

Expected: pass.

- [ ] **Step 10: Commit**

Run:

```powershell
git add internal/online/client.go internal/online/client_test.go cmd/client/main.go cmd/client/main_test.go
git commit -m "feat: add cli room listing"
```

Expected: commit succeeds.

---

## Task 4: Reconnect Progress Events For Client And TUI

**Files:**
- Modify: `internal/online/client.go`
- Modify: `internal/online/client_test.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/online.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing client reconnect callback test**

Add this test to `internal/online/client_test.go`:

```go
func TestClientReadUntilWithReconnectReportsAttemptsAndSuccess(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	client := NewClient(url+"/ws", "first")
	defer client.Close()
	created, err := client.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.conn.Close(); err != nil {
		t.Fatal(err)
	}

	var attempts []int
	successes := 0
	reconnected, err := client.ReadUntilWithReconnect(
		context.Background(),
		2*time.Second,
		ReconnectPolicy{
			MaxAttempts: 5,
			BaseDelay:   time.Millisecond,
			OnAttempt: func(attempt int, max int) {
				if max != 5 {
					t.Fatalf("max = %d, want 5", max)
				}
				attempts = append(attempts, attempt)
			},
			OnSuccess: func() {
				successes++
			},
		},
		protocol.MsgReconnected,
	)
	if err != nil {
		t.Fatal(err)
	}
	if reconnected.PlayerID != created.PlayerID {
		t.Fatalf("reconnected player = %q, want %q", reconnected.PlayerID, created.PlayerID)
	}
	if len(attempts) == 0 || attempts[0] != 1 {
		t.Fatalf("attempts = %#v, want first attempt reported", attempts)
	}
	if successes != 1 {
		t.Fatalf("successes = %d, want 1", successes)
	}
}
```

- [ ] **Step 2: Run the failing callback test**

Run:

```powershell
go test ./internal/online -run TestClientReadUntilWithReconnectReportsAttemptsAndSuccess -count=1
```

Expected: compile fails because `ReconnectPolicy.OnAttempt` and `ReconnectPolicy.OnSuccess` do not exist.

- [ ] **Step 3: Add reconnect callbacks to the client policy**

Modify `ReconnectPolicy` in `internal/online/client.go`:

```go
type ReconnectPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	OnAttempt   func(attempt int, max int)
	OnSuccess   func()
}
```

Inside `reconnectWithBackoff`, call the callbacks:

```go
for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
	if policy.OnAttempt != nil {
		policy.OnAttempt(attempt+1, policy.MaxAttempts)
	}
	c.Close()
	message, err := c.Reconnect(ctx, session)
	if err == nil {
		if policy.OnSuccess != nil {
			policy.OnSuccess()
		}
		return message, nil
	}
	lastErr = err
	delay := policy.BaseDelay << attempt
	timer := time.NewTimer(delay)
	select {
	case <-ctx.Done():
		timer.Stop()
		return protocol.Message{}, ctx.Err()
	case <-timer.C:
	}
}
```

- [ ] **Step 4: Run reconnect client tests**

Run:

```powershell
go test ./internal/online -run "TestClientReadUntilWithReconnect" -count=1
```

Expected: pass.

- [ ] **Step 5: Write the failing TUI reconnect event tests**

Add these tests to `internal/tui/model_test.go`:

```go
func TestOnlineReconnectAttemptMessageUpdatesNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable

	next, cmd := model.Update(onlineReconnectAttemptMsg{Attempt: 2, Max: 5})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command for direct reconnect message without event channel")
	}
	if updated.NetworkStatus != NetworkReconnecting {
		t.Fatalf("network status = %q, want reconnecting", updated.NetworkStatus)
	}
	if updated.ReconnectAttempt != 2 || updated.ReconnectMax != 5 {
		t.Fatalf("attempt = %d/%d, want 2/5", updated.ReconnectAttempt, updated.ReconnectMax)
	}
	if !strings.Contains(updated.View(), "Network: reconnecting 2/5") {
		t.Fatalf("view missing reconnecting status:\n%s", updated.View())
	}
}

func TestOnlineReconnectSuccessMessageUpdatesNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.NetworkStatus = NetworkReconnecting
	model.ReconnectAttempt = 2
	model.ReconnectMax = 5

	next, cmd := model.Update(onlineReconnectSuccessMsg{})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command for direct reconnect success without event channel")
	}
	if updated.NetworkStatus != NetworkReconnected {
		t.Fatalf("network status = %q, want reconnected", updated.NetworkStatus)
	}
	if !strings.Contains(updated.View(), "Network: reconnected") {
		t.Fatalf("view missing reconnected status:\n%s", updated.View())
	}
}
```

- [ ] **Step 6: Run the failing TUI reconnect tests**

Run:

```powershell
go test ./internal/tui -run "TestOnlineReconnect.*Message" -count=1
```

Expected: compile fails because `onlineReconnectAttemptMsg` and `onlineReconnectSuccessMsg` do not exist.

- [ ] **Step 7: Add TUI reconnect message types and update handling**

In `internal/tui/online.go`, add:

```go
type onlineReconnectAttemptMsg struct {
	Attempt int
	Max     int
}

type onlineReconnectSuccessMsg struct{}

func listenOnlineEvents(events <-chan tea.Msg) tea.Cmd {
	if events == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return nil
		}
		return msg
	}
}

func sendOnlineEvent(events chan<- tea.Msg, msg tea.Msg) {
	if events == nil {
		return
	}
	select {
	case events <- msg:
	default:
	}
}
```

In `internal/tui/model.go`, add this field to `Model`:

```go
OnlineEvents chan tea.Msg
```

In `internal/tui/model.go`, add cases to `Update`:

```go
case onlineReconnectAttemptMsg:
	m.NetworkStatus = NetworkReconnecting
	m.ReconnectAttempt = msg.Attempt
	m.ReconnectMax = msg.Max
	return m, listenOnlineEvents(m.OnlineEvents)
case onlineReconnectSuccessMsg:
	m.NetworkStatus = NetworkReconnected
	m.StatusLine = "Reconnected"
	return m, listenOnlineEvents(m.OnlineEvents)
```

- [ ] **Step 8: Wire reconnect callbacks in `waitOnlineSnapshot`**

In `internal/tui/online.go`, change `waitOnlineSnapshot` to accept an event channel:

```go
func waitOnlineSnapshot(client *online.Client, events chan<- tea.Msg) tea.Cmd {
	if client == nil {
		return nil
	}
	return func() tea.Msg {
		message, err := client.ReadUntilWithReconnect(
			context.Background(),
			24*time.Hour,
			online.ReconnectPolicy{
				MaxAttempts: 5,
				BaseDelay:   200 * time.Millisecond,
				OnAttempt: func(attempt int, max int) {
					sendOnlineEvent(events, onlineReconnectAttemptMsg{Attempt: attempt, Max: max})
				},
				OnSuccess: func() {
					sendOnlineEvent(events, onlineReconnectSuccessMsg{})
				},
			},
			protocol.MsgGameSnapshot,
			protocol.MsgRoomState,
			protocol.MsgReconnected,
			protocol.MsgError,
		)
		if err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineSnapshotMsg{Message: message}
	}
}
```

In `applyOnlineConnected`, initialize the channel before returning the model:

```go
if m.OnlineEvents == nil {
	m.OnlineEvents = make(chan tea.Msg, 8)
}
```

In `Model.Update`, change the online connected and snapshot cases:

```go
case onlineConnectedMsg:
	updated := applyOnlineConnected(m, msg)
	return updated, tea.Batch(waitOnlineSnapshot(msg.Client, updated.OnlineEvents), listenOnlineEvents(updated.OnlineEvents))
case onlineSnapshotMsg:
	updated := applyOnlineSnapshot(m, msg.Message)
	return updated, waitOnlineSnapshot(updated.OnlineClient, updated.OnlineEvents)
```

Update all direct test calls from:

```go
msg = waitOnlineSnapshot(client)()
```

to:

```go
msg = waitOnlineSnapshot(client, nil)()
```

- [ ] **Step 9: Run TUI reconnect status tests**

Run:

```powershell
go test ./internal/tui -run "TestOnlineReconnect.*Message|TestOnlineConnectedMessageShowsSnapshotTable" -count=1
```

Expected: pass.

- [ ] **Step 10: Run full tests**

Run:

```powershell
go test ./...
```

Expected: pass.

- [ ] **Step 11: Commit**

Run:

```powershell
git add internal/online/client.go internal/online/client_test.go internal/tui/model.go internal/tui/online.go internal/tui/model_test.go
git commit -m "feat: report reconnect progress"
```

Expected: commit succeeds.

---

## Task 5: Configurable Server Lifecycle

**Files:**
- Create: `internal/online/options.go`
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`

- [ ] **Step 1: Write the failing reconnect-window option test**

Add this test to `internal/online/server_test.go`:

```go
func TestWebSocketServerRejectsReconnectAfterConfiguredWindow(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{ReconnectWindow: time.Nanosecond})
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	first.Close()

	time.Sleep(2 * time.Millisecond)

	reconnecting := dialTestClient(t, url)
	defer reconnecting.Close()
	sendMessage(t, reconnecting, protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       created.PlayerID,
		ReconnectToken: created.ReconnectToken,
	})
	errMsg := readUntil(t, reconnecting, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "reconnect window expired") {
		t.Fatalf("error = %#v", errMsg)
	}
}
```

- [ ] **Step 2: Run the failing option test**

Run:

```powershell
go test ./internal/online -run TestWebSocketServerRejectsReconnectAfterConfiguredWindow -count=1
```

Expected: compile fails because `NewServerWithOptions` and `ServerOptions` do not exist.

- [ ] **Step 3: Add server options**

Create `internal/online/options.go`:

```go
package online

import "time"

type ServerOptions struct {
	ReconnectWindow time.Duration
	RoomIdleTTL     time.Duration
}

func (o ServerOptions) withDefaults() ServerOptions {
	if o.ReconnectWindow <= 0 {
		o.ReconnectWindow = 2 * time.Minute
	}
	if o.RoomIdleTTL <= 0 {
		o.RoomIdleTTL = 10 * time.Minute
	}
	return o
}
```

- [ ] **Step 4: Store options on the server**

In `internal/online/server.go`, remove the package-level reconnect constant and add `options ServerOptions` to `Server`:

```go
type Server struct {
	mu       sync.Mutex
	rooms    map[string]*room
	sessions map[string]*session
	bots     bot.BotEngine
	upgrader websocket.Upgrader
	nextRoom int
	options  ServerOptions
}
```

Replace `NewServer` with:

```go
func NewServer() *Server {
	return NewServerWithOptions(ServerOptions{})
}

func NewServerWithOptions(options ServerOptions) *Server {
	options = options.withDefaults()
	return &Server{
		rooms:    make(map[string]*room),
		sessions: make(map[string]*session),
		bots:     bot.NewHeuristicBot(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		options: options,
	}
}
```

In `reconnect`, replace the old window check with:

```go
if !session.offlineAt.IsZero() && time.Since(session.offlineAt) > s.options.ReconnectWindow {
	return nil, nil, game.GameSnapshot{}, fmt.Errorf("reconnect window expired")
}
```

- [ ] **Step 5: Run the reconnect-window option test**

Run:

```powershell
go test ./internal/online -run TestWebSocketServerRejectsReconnectAfterConfiguredWindow -count=1
```

Expected: pass.

- [ ] **Step 6: Write the failing idle-room list pruning test**

Add this test to `internal/online/server_test.go`:

```go
func TestWebSocketServerDoesNotListExpiredIdleRooms(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{RoomIdleTTL: time.Nanosecond})
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := dialTestClient(t, url)
	sendMessage(t, host, protocol.Message{Type: protocol.MsgCreateRoom})
	_ = readUntil(t, host, protocol.MsgRoomCreated)
	host.Close()

	time.Sleep(2 * time.Millisecond)

	lister := dialTestClient(t, url)
	defer lister.Close()
	sendMessage(t, lister, protocol.Message{Type: protocol.MsgListRooms})
	list := readUntil(t, lister, protocol.MsgRoomList)
	if len(list.Rooms) != 0 {
		t.Fatalf("rooms = %#v, want expired room hidden", list.Rooms)
	}
}
```

- [ ] **Step 7: Run the failing idle-room test**

Run:

```powershell
go test ./internal/online -run TestWebSocketServerDoesNotListExpiredIdleRooms -count=1
```

Expected: fail because rooms do not track idle timestamps.

- [ ] **Step 8: Track room activity and prune before list**

Add `updatedAt time.Time` to `room`:

```go
type room struct {
	code      string
	game      *game.Game
	seats     [4]string
	ready     map[string]bool
	started   bool
	updatedAt time.Time
}
```

Set it in `createRoom`:

```go
created := &room{
	code:      code,
	game:      game.NewGame(0),
	ready:     make(map[string]bool),
	updatedAt: time.Now(),
}
```

Add this helper:

```go
func (s *Server) touchRoomLocked(room *room) {
	if room != nil {
		room.updatedAt = time.Now()
	}
}
```

Call `s.touchRoomLocked(room)` in `joinRoom`, `setReady`, and `playCommand` after successful state changes.

Add this helper:

```go
func (s *Server) pruneExpiredRoomsLocked(now time.Time) {
	for code, room := range s.rooms {
		if room.started {
			continue
		}
		if now.Sub(room.updatedAt) <= s.options.RoomIdleTTL {
			continue
		}
		for _, playerID := range room.seats {
			if playerID != "" {
				delete(s.sessions, playerID)
			}
		}
		delete(s.rooms, code)
	}
}
```

Call it at the top of `roomSummaries` after acquiring the lock:

```go
s.pruneExpiredRoomsLocked(time.Now())
```

- [ ] **Step 9: Run lifecycle tests**

Run:

```powershell
go test ./internal/online -run "ConfiguredWindow|ExpiredIdleRooms|ListsWaitingRooms" -count=1
```

Expected: pass.

- [ ] **Step 10: Run full tests**

Run:

```powershell
go test ./...
```

Expected: pass.

- [ ] **Step 11: Commit**

Run:

```powershell
git add internal/online/options.go internal/online/server.go internal/online/server_test.go
git commit -m "feat: configure online room lifecycle"
```

Expected: commit succeeds.

---

## Task 6: TUI Online Lobby And Browse Flow

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/menu.go`
- Modify: `internal/tui/online.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing menu option test**

Update `TestMenuViewContainsOptions` in `internal/tui/model_test.go` so the expected texts include:

```go
"Browse Online Rooms"
```

Run:

```powershell
go test ./internal/tui -run TestMenuViewContainsOptions -count=1
```

Expected: fail because the menu item is missing.

- [ ] **Step 2: Add the menu item and screen state**

In `internal/tui/model.go`, add a screen:

```go
ScreenOnlineRooms
```

Add fields to `Model`:

```go
OnlineRooms []protocol.RoomSummary
RoomIndex   int
```

Add the protocol import in `internal/tui/model.go`:

```go
"mahjong/internal/protocol"
```

In `internal/tui/menu.go`, change `menuItems` to:

```go
var menuItems = []string{"Solo Game", "Create Online Room", "Browse Online Rooms", "Join Online Room", "Reconnect Online", "How to Play", "Quit"}
```

Adjust `updateMenu` cases so:
- Case `0`: solo game.
- Case `1`: create room.
- Case `2`: browse rooms and return `listOnlineRoomsCmd(m)`.
- Case `3`: manual join room.
- Case `4`: reconnect.
- Case `5`: help.
- Case `6`: quit.

- [ ] **Step 3: Add room-list command messages**

In `internal/tui/online.go`, add:

```go
type onlineRoomsMsg struct {
	Rooms []protocol.RoomSummary
}
```

Add this command:

```go
func listOnlineRoomsCmd(m Model) tea.Cmd {
	serverURL := m.OnlineServerURL
	name := m.OnlineName
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := online.NewClient(serverURL, name)
		defer client.Close()
		rooms, err := client.ListRooms(ctx)
		if err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineRoomsMsg{Rooms: rooms}
	}
}
```

In `internal/tui/model.go`, add an update case:

```go
case onlineRoomsMsg:
	m.Screen = ScreenOnlineRooms
	m.OnlineRooms = append([]protocol.RoomSummary(nil), msg.Rooms...)
	m.RoomIndex = 0
	m.StatusLine = fmt.Sprintf("Rooms found: %d", len(msg.Rooms))
	return m, nil
```

Add the `fmt` import to `internal/tui/model.go` if it is not present.

- [ ] **Step 4: Add online-room list rendering and controls**

In `internal/tui/menu.go`, add:

```go
func updateOnlineRooms(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.Screen = ScreenMenu
	case tea.KeyDown:
		if len(m.OnlineRooms) > 0 {
			m.RoomIndex = (m.RoomIndex + 1) % len(m.OnlineRooms)
		}
	case tea.KeyUp:
		if len(m.OnlineRooms) > 0 {
			m.RoomIndex = (m.RoomIndex + len(m.OnlineRooms) - 1) % len(m.OnlineRooms)
		}
	case tea.KeyEnter:
		if len(m.OnlineRooms) == 0 {
			m.StatusLine = "No waiting rooms"
			return m, nil
		}
		m.JoinRoomCode = m.OnlineRooms[m.RoomIndex].Code
		m.NetworkStatus = NetworkWaiting
		m.StatusLine = "Joining room " + m.JoinRoomCode + "..."
		return m, joinOnlineRoomCmd(m)
	}
	switch key.String() {
	case "r":
		m.StatusLine = "Refreshing rooms..."
		return m, listOnlineRoomsCmd(m)
	}
	return m, nil
}

func renderOnlineRooms(m Model) string {
	var out strings.Builder
	out.WriteString(styleTitle("ONLINE ROOMS") + "\n\n")
	out.WriteString(styleSectionTitle("Waiting Rooms") + "\n")
	if len(m.OnlineRooms) == 0 {
		out.WriteString(styleMuted("No waiting rooms") + "\n")
	} else {
		for i, room := range m.OnlineRooms {
			line := fmt.Sprintf("%s  players:%d ready:%d wall:%d", room.Code, room.Occupied, room.Ready, room.Wall)
			if i == m.RoomIndex {
				out.WriteString(styleSelectedTile("> "+line) + "\n")
			} else {
				out.WriteString("  " + line + "\n")
			}
		}
	}
	out.WriteString("\n" + renderStatus(m))
	out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
	out.WriteString(styleMuted("Up/Down choose | Enter join | R refresh | Esc menu") + "\n")
	return out.String()
}
```

In `Model.Update`, add:

```go
case ScreenOnlineRooms:
	return updateOnlineRooms(m, msg)
```

In `Model.View`, add:

```go
case ScreenOnlineRooms:
	return renderOnlineRooms(m)
```

- [ ] **Step 5: Write the browse-room rendering tests**

Add these tests to `internal/tui/model_test.go`:

```go
func TestOnlineRoomsMessageShowsRoomList(t *testing.T) {
	model := NewModel()

	next, _ := model.Update(onlineRoomsMsg{Rooms: []protocol.RoomSummary{
		{Code: "000123", Occupied: 1, Ready: 0, Wall: 67},
	}})
	updated := next.(Model)

	if updated.Screen != ScreenOnlineRooms {
		t.Fatalf("screen = %v, want online rooms", updated.Screen)
	}
	view := updated.View()
	for _, want := range []string{"ONLINE ROOMS", "000123", "players:1", "Enter join"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func TestOnlineRoomsEnterJoinsSelectedRoom(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	host := online.NewClient(serverURL, "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	model := NewModel()
	model.Screen = ScreenOnlineRooms
	model.OnlineServerURL = serverURL
	model.OnlineSession = t.TempDir() + "/session.json"
	model.OnlineRooms = []protocol.RoomSummary{{Code: created.RoomCode, Occupied: 1, Ready: 0, Wall: 67}}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected join command")
	}
	if updated.JoinRoomCode != created.RoomCode {
		t.Fatalf("join room code = %q, want %q", updated.JoinRoomCode, created.RoomCode)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if !updated.Online || updated.OnlineRoomCode != created.RoomCode {
		t.Fatalf("online=%v room=%q want room %q", updated.Online, updated.OnlineRoomCode, created.RoomCode)
	}
}
```

If `internal/tui/model_test.go` does not already import `mahjong/internal/protocol`, add it.

- [ ] **Step 6: Run TUI browse tests**

Run:

```powershell
go test ./internal/tui -run "MenuViewContainsOptions|OnlineRooms" -count=1
```

Expected: pass.

- [ ] **Step 7: Run full tests**

Run:

```powershell
go test ./...
```

Expected: pass.

- [ ] **Step 8: Commit**

Run:

```powershell
git add internal/tui/model.go internal/tui/menu.go internal/tui/online.go internal/tui/model_test.go
git commit -m "feat: browse online rooms in tui"
```

Expected: commit succeeds.

---

## Task 7: Reference-Inspired TUI Table Skin

**Stage Goal:** Convert the existing terminal table from sectioned text into a centered, readable, four-seat Mahjong table inspired by the reference project's screenshots and Lip Gloss renderer structure.

**Reference Inputs:**
- Screenshot: `C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference\docs\lobby.png`
- Screenshot: `C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference\docs\in-game.png`
- Code study only, no copying: `C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference\internal\ui\view\renderer.go`
- Code study only, no copying: `C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference\internal\ui\view\lobby.go`

**Files:**
- Create: `docs/tui-reference-study.md`
- Modify: `internal/tui/style.go`
- Modify: `internal/tui/style_test.go`
- Modify: `internal/tui/layout.go`
- Modify: `internal/tui/layout_test.go`

- [ ] **Step 1: Record the UI translation contract**

Create `docs/tui-reference-study.md` with this content:

```markdown
# TUI Reference Study

Reference project: `palemoky/fight-the-landlord`

Local reference clone used for study:
`C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference`

## Boundary

This project studies the reference screenshots and layout architecture only.
No GPL implementation code is copied.

## Screenshot Observations

`docs/lobby.png`:
- Main content is centered in the terminal.
- A primary menu panel sits beside a secondary info/chat panel.
- The bottom control hint is muted and visually separate.

`docs/in-game.png`:
- The game screen is centered vertically and horizontally.
- The top area shows compact status/counter information.
- The middle area shows opponent boxes and the latest play.
- The bottom area is a framed hand tray.
- The prompt/action hint sits below the hand, not mixed into the table.

## Mahjong Translation

- Lobby: keep terminal menu, but make it centered and panel-based.
- Table: render four seats around a central discard/event area.
- Hand: render player's hand as a stable tray with clear selected tile focus.
- Width: keep important lines within 96 visible cells on normal terminals and 80 cells in compact mode.
- Mouse: hitboxes must continue to match the rendered hand row.
```

Step Review: this step advances the stage goal by converting visual inspiration into a written non-GPL implementation contract.

- [ ] **Step 2: Verify the reference contract text**

Run:

```powershell
rg -n "TUI Reference Study|No GPL implementation code|four seats|stable tray" docs/tui-reference-study.md
```

Expected: all four phrases are reported.

Step Review: this step advances the stage goal by proving the UI contract exists and covers the no-copy boundary, table layout, and hand tray.

- [ ] **Step 3: Write failing style tests for reusable table panels**

Add these tests to `internal/tui/style_test.go`:

```go
func TestStylePanelKeepsContentVisible(t *testing.T) {
	panel := stylePanel("Seat", "AI-1\nhand:13")

	for _, want := range []string{"Seat", "AI-1", "hand:13"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("panel missing %q:\n%s", want, panel)
		}
	}
}

func TestStylePanelRespectsVisibleWidth(t *testing.T) {
	panel := stylePanel("Center", "Wall: 67")

	if got := visibleWidth(panel); got > 40 {
		t.Fatalf("panel width = %d, want <= 40:\n%s", got, panel)
	}
}

func TestStyleTileFaceKeepsUnicodeTile(t *testing.T) {
	tile := styleTileFace("🀇", true)

	if !strings.Contains(tile, "🀇") {
		t.Fatalf("tile face missing unicode tile: %q", tile)
	}
	if got := visibleWidth(tile); got > 6 {
		t.Fatalf("tile face width = %d, want <= 6: %q", got, tile)
	}
}
```

Step Review: this step advances the stage goal by defining reusable visual primitives before changing the renderer.

- [ ] **Step 4: Run the failing style tests**

Run:

```powershell
go test ./internal/tui -run "TestStylePanel|TestStyleTileFace" -count=1
```

Expected: compile fails because `stylePanel` and `styleTileFace` do not exist.

Step Review: this step advances the stage goal by confirming the next code change is driven by explicit visual tests.

- [ ] **Step 5: Implement reusable panel and tile styles**

In `internal/tui/style.go`, add this style:

```go
panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D6CCC2")).
		Padding(0, 1)

tileFaceStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#111827")).
		Background(lipgloss.Color("#F8FAFC")).
		Bold(true).
		Padding(0, 1)
```

Add these helpers:

```go
func stylePanel(title string, body string) string {
	content := strings.TrimRight(title+"\n"+body, "\n")
	return panelStyle.Render(content)
}

func styleTileFace(label string, selected bool) string {
	style := tileFaceStyle
	if selected {
		style = style.Background(lipgloss.Color("#FDE68A"))
	}
	return style.Render(label)
}
```

Add this import to `internal/tui/style.go`:

```go
import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)
```

Step Review: this step advances the stage goal by adding the shared panel and tile-face primitives needed for a table-like TUI.

- [ ] **Step 6: Run the style tests**

Run:

```powershell
go test ./internal/tui -run "TestStylePanel|TestStyleTileFace|TestVisibleWidth" -count=1
```

Expected: pass.

Step Review: this step advances the stage goal by verifying the new visual primitives preserve text and visible width.

- [ ] **Step 7: Write failing table-layout tests**

Add these tests to `internal/tui/layout_test.go`:

```go
func TestRenderTableUsesReferenceInspiredTabletop(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, want := range []string{"TERMINAL MAHJONG", "AI-2", "AI-1", "AI-3", "CENTER", "Hand Tray"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if lineIndexContaining(view, "CENTER") <= lineIndexContaining(view, "AI-2") {
		t.Fatalf("center should appear below opposite seat:\n%s", view)
	}
	if lineIndexContaining(view, "Hand Tray") <= lineIndexContaining(view, "CENTER") {
		t.Fatalf("hand tray should appear below center table:\n%s", view)
	}
}

func TestRenderTableCentersMainBoardWhenWide(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	line := firstLineContaining(view, "TERMINAL MAHJONG")
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("wide title should be centered with leading space:\n%s", view)
	}
}

func TestRenderTableKeepsReferenceInspiredWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 120 {
			t.Fatalf("line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}
```

Add this helper to `internal/tui/layout_test.go`:

```go
func firstLineContaining(text string, part string) string {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, part) {
			return line
		}
	}
	return ""
}
```

Step Review: this step advances the stage goal by defining the required four-seat table order, centered board, and width budget.

- [ ] **Step 8: Run the failing table-layout tests**

Run:

```powershell
go test ./internal/tui -run "ReferenceInspired|CentersMainBoard|ReferenceInspiredWidth" -count=1
```

Expected: at least `TestRenderTableCentersMainBoardWhenWide` fails because the current table renderer is not centered with `lipgloss.PlaceHorizontal`.

Step Review: this step advances the stage goal by proving the current screen does not yet match the reference-inspired layout requirement.

- [ ] **Step 9: Implement centered table board helpers**

In `internal/tui/layout.go`, add the Lip Gloss import:

```go
"github.com/charmbracelet/lipgloss"
```

Add these helpers:

```go
func tableWidth(m Model) int {
	width := m.Width
	if width <= 0 {
		width = 96
	}
	if width < 80 {
		return width
	}
	if width > 120 {
		return 120
	}
	return width
}

func centerLine(width int, text string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, text)
}

func renderBoardFrame(m Model, topSeat string, middle string, hand string, prompt string) string {
	width := tableWidth(m)
	sections := []string{
		centerLine(width, styleTitle("TERMINAL MAHJONG")),
		centerLine(width, styleMuted(renderNetworkStatus(m))),
		centerLine(width, topSeat),
		centerLine(width, middle),
		centerLine(width, hand),
		centerLine(width, prompt),
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(sections, "\n"))
}

func renderSeatPanel(label string, name string, handCount int, melds string, discards string, active bool) string {
	if discards == "" {
		discards = "-"
	}
	title := label + "  " + name
	if active {
		title = "▶ " + title
	}
	body := fmt.Sprintf("hand:%02d  melds:%s\nlast:%s", handCount, melds, discards)
	return stylePanel(title, body)
}
```

Step Review: this step advances the stage goal by creating centered board primitives instead of scattering string padding across the renderer.

- [ ] **Step 10: Update local and online table renderers to use the board frame**

In `renderTable`, keep the existing data sources but replace the top-level loose section appends with these parts:

```go
topSeat := renderSeatPanel("AI-2", g.Players[2].Name, len(g.Players[2].Hand), g.Players[2].MeldSummary(), game.FormatTileLabels(recentTiles(g.Players[2].Discards, 4), m.UnicodeTiles), g.Current == 2)
leftSeat := renderSeatPanel("AI-1", g.Players[1].Name, len(g.Players[1].Hand), g.Players[1].MeldSummary(), game.FormatTileLabels(recentTiles(g.Players[1].Discards, 4), m.UnicodeTiles), g.Current == 1)
rightSeat := renderSeatPanel("AI-3", g.Players[3].Name, len(g.Players[3].Hand), g.Players[3].MeldSummary(), game.FormatTileLabels(recentTiles(g.Players[3].Discards, 4), m.UnicodeTiles), g.Current == 3)
center := stylePanel("CENTER", renderCenter(g, m.UnicodeTiles))
middle := lipgloss.JoinHorizontal(lipgloss.Top, leftSeat, "  ", center, "  ", rightSeat)
hand := stylePanel("Hand Tray", renderStatus(m)+renderLastAction(m)+fmt.Sprintf("Melds: %s\n", g.Players[0].MeldSummary())+fmt.Sprintf("Discards: %s\n", game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles))+renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles)+renderActionBar(m))
prompt := styleMuted(tableControls(m))
return renderBoardFrame(m, topSeat, middle, hand, prompt)
```

In `renderOnlineTable`, apply the same structure using `onlinePlayer`, `m.OnlineSnapshot.Current`, `renderOnlineCenter`, `renderOnlineActionBar`, and `renderOnlineRoomState`.

Step Review: this step advances the stage goal by moving both solo and online views to the same four-seat table structure.

- [ ] **Step 11: Keep mouse hitboxes aligned with the hand tray**

Update constants in `internal/tui/layout.go` only if the rendered first `Hand:` line moved:

```go
const (
	handStartX = 6
	handRowY   = 18
	handRowGap = 2
	handCellW  = 10
	handCols   = 7
)
```

Then update `TestDefaultHandHitBoxesMatchRenderedHandStartLine` expected behavior through the existing assertion, not by hard-coding a second expected value in the test.

Step Review: this step advances the stage goal by preserving mouse click accuracy after the visual table moves.

- [ ] **Step 12: Run table, style, and TUI tests**

Run:

```powershell
go test ./internal/tui -run "ReferenceInspired|CentersMainBoard|ReferenceInspiredWidth|HandHitBoxes|RenderTable|StylePanel|StyleTileFace" -count=1
```

Expected: pass.

Step Review: this step advances the stage goal by proving the reference-inspired screen remains readable, bounded, and clickable.

- [ ] **Step 13: Run full tests**

Run:

```powershell
go test ./...
```

Expected: pass.

Step Review: this step advances the stage goal by proving the visual upgrade did not regress game, online, CLI, or bot behavior.

- [ ] **Step 14: Commit**

Run:

```powershell
git add docs/tui-reference-study.md internal/tui/style.go internal/tui/style_test.go internal/tui/layout.go internal/tui/layout_test.go
git commit -m "feat: add reference-inspired tui table skin"
```

Expected: commit succeeds.

Step Review: this step advances the stage goal by saving the UI upgrade as an isolated reviewable commit.

**Stage Review:** Task 7 makes the terminal Mahjong table visually closer to the reference project's centered card-table style while keeping terminal controls, mouse hitboxes, Unicode tiles, and automated tests intact. The total-goal risk remaining after this task is final end-to-end acceptance only.

---

## Task 8: Final Acceptance Verification

**Files:**
- Modify: `docs/online-client-acceptance.md`

- [ ] **Step 1: Run full automated tests**

Run:

```powershell
go test ./...
```

Expected: pass.

- [ ] **Step 2: Build all commands**

Run:

```powershell
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

Expected: pass and no compiler output.

- [ ] **Step 3: Run focused package tests**

Run:

```powershell
go test ./internal/online ./cmd/client ./internal/tui -count=1
```

Expected: pass.

- [ ] **Step 4: Update acceptance ledger with verification result**

Append this section to `docs/online-client-acceptance.md` after the verification commands:

```markdown
## Latest Verified Result

- `go test ./...`: pass
- `go build ./cmd/mahjong ./cmd/server ./cmd/client`: pass
- `go test ./internal/online ./cmd/client ./internal/tui -count=1`: pass
```

- [ ] **Step 5: Commit**

Run:

```powershell
git add docs/online-client-acceptance.md
git commit -m "test: record online client acceptance"
```

Expected: commit succeeds.

---

## Final Self-Review

- Spec coverage:
  - fight-the-landlord reference architecture is documented and bounded by the GPL no-copy rule.
  - fight-the-landlord UI screenshots and renderer structure are documented and translated into a terminal Mahjong table-skin task.
  - Room discovery covers lobby-style client flow.
  - Reconnect progress covers client callbacks and TUI-visible states.
  - Configurable reconnect and idle-room lifecycle cover server robustness.
  - CLI/TUI/server acceptance commands are listed.
- Placeholder scan:
  - Run `$patterns = @("TB"+"D", "TO"+"DO", "implement"+" later", "fill in"+" details", "Similar"+" to", "appropriate"+" error"); foreach ($pattern in $patterns) { rg -n $pattern docs/superpowers/plans/2026-06-13-online-mahjong-final-client.md }`.
  - Expected: no matches and exit code `1`.
- Type consistency:
  - `protocol.RoomSummary` is introduced before all usages.
  - `protocol.MsgListRooms` and `protocol.MsgRoomList` are introduced before client/server/TUI usage.
  - `ReconnectPolicy.OnAttempt` and `ReconnectPolicy.OnSuccess` are introduced before callback tests.
  - `ScreenOnlineRooms`, `onlineRoomsMsg`, and `Model.OnlineRooms` are introduced before browse-flow tests.
  - `stylePanel`, `styleTileFace`, `renderBoardFrame`, and `renderSeatPanel` are introduced before the table-skin rendering tests depend on them.
