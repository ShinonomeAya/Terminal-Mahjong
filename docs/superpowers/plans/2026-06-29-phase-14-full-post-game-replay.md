# Phase 14 Full Post-Game Replay Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically save completed local and online matches as validated full-information replay files and provide a bilingual, read-only TUI browser and frame viewer.

**Architecture:** `game.Match` owns a compact authoritative replay journal so local play, bots, and the WebSocket server record the same accepted commands and full snapshots. A separate `internal/replay` package validates checksums and performs atomic disk persistence without becoming a second game engine. The TUI adds browser/viewer screens that reuse the Phase 13 table state and renderer in read-only mode.

**Tech Stack:** Go standard library (`encoding/json`, `crypto/sha256`, `os`, `path/filepath`, `runtime/debug`, `time`), Bubble Tea, Lip Gloss, Gorilla WebSocket, existing `game.MatchSnapshot`, protocol messages, and Go `testing`.

---

## Approved Contract And Assumptions

- `ReplayFileSchemaVersion` is `2`, as specified in the approved design. It is independent from the existing event-summary `ReplaySchemaVersion == 3`.
- Only completed matches produce valid files. Partial journals remain in memory and fail validation if exported as completed replays.
- A user quit is an abandoned match, not a completed match. It records the final quit command/frame for in-memory diagnostics but never autosaves a valid `ReplayFile`.
- Replay frames contain authoritative full `MatchSnapshot` values. Live `SnapshotFor` privacy behavior remains unchanged.
- Replay files never contain player reconnect tokens, socket addresses, or client session paths.
- Local TUI play retains a `*game.Match` instead of discarding the coordinator after startup. `Model.Game` remains an alias of `LocalMatch.Round` during migration so existing table code stays stable.
- The server retains completed replay payloads only with its existing in-memory room/session lifetime. No database, Redis, public matchmaking, video export, replay editing, or branch-from-frame behavior is added.
- Replay playback never calls `ApplyCommand`; it only selects recorded frames.
- Every completed task receives a step review against the current subphase goal. Every completed 14A-E subphase receives a phase review against the total auditable-client goal.

## File Map

- Create `internal/game/replay_file.go`: versioned replay types, checksum sealing, validation, and compatibility errors.
- Create `internal/game/replay_file_test.go`: schema, checksum, unsupported-version, incomplete-file, and mutation tests.
- Modify `internal/game/match.go`: authoritative command/frame journal, frame capture around settlement/round transitions, and replay export.
- Modify `internal/game/turn.go`: match-level AI advancement through `Match.ApplyCommand`.
- Modify `internal/game/match_test.go`: deterministic command/frame and round-transition coverage.
- Create `internal/replay/store.go`: application version, atomic save, validated load, newest-first listing, and corrupt-file reporting.
- Create `internal/replay/store_test.go`: atomicity, checksum validation, ordering, and corrupt-file isolation.
- Modify `internal/tui/model.go`: local match, replay browser/viewer state, replay messages, and screens.
- Modify `internal/tui/game_flow.go`: create and retain local matches.
- Modify `internal/tui/input.go`: route local actions through `Match`, browser navigation, viewer navigation, and timed playback.
- Modify `internal/tui/menu.go`: add the replay-browser entry and localized browser rendering.
- Create `internal/tui/replay.go`: asynchronous list/load/save commands and replay viewer rendering.
- Create `internal/tui/replay_test.go`: menu, browser, playback, read-only, localization, and width tests.
- Modify `internal/tui/table_state.go`: normalize replay frames into the shared table state.
- Modify `internal/tui/table_components.go`: replay-specific read-only controls and detail rail.
- Modify `internal/tui/i18n.go`: bilingual replay labels and status text.
- Modify `internal/protocol/message.go`: replay-ready, request, and replay-data messages.
- Modify `internal/online/server.go`: authoritative replay creation, completed-room retention, delivery, and reconnect retrieval.
- Modify `internal/online/client.go`: request completed replay payloads.
- Modify `internal/online/server_test.go` and `internal/online/client_test.go`: privacy, completion, delivery, and reconnect tests.
- Modify `cmd/client/main.go` and `cmd/client/main_test.go`: replay-directory option and completed replay saving in watch mode.
- Create `internal/tui/replay_screenshot_test.go`: deterministic Phase 14 visual artifacts.
- Modify `README.md`: replay location, controls, validation, and privacy.
- Modify `docs/workflow.md`: 14A-E step reviews, phase reviews, artifact paths, and acceptance evidence.

## Task 1: 14A Versioned Replay Schema And Checksum

**Files:**
- Create: `internal/game/replay_file.go`
- Create: `internal/game/replay_file_test.go`

- [x] **Step 1: Write failing schema and checksum tests**

Add table-driven tests that construct a minimal completed file and require:

```go
func TestReplayFileSealAndValidate(t *testing.T) {
	file := completedReplayFixture(t, ModeRiichi)
	sealed, err := SealReplay(file)
	if err != nil {
		t.Fatal(err)
	}
	if sealed.SchemaVersion != ReplayFileSchemaVersion || sealed.Checksum == "" {
		t.Fatalf("sealed replay = %#v", sealed)
	}
	if err := ValidateReplay(sealed); err != nil {
		t.Fatal(err)
	}

	sealed.Frames[0].Match.Points[0]++
	if !errors.Is(ValidateReplay(sealed), ErrReplayChecksum) {
		t.Fatal("mutated replay should fail checksum validation")
	}
}

func TestReplayFileRejectsUnsupportedAndIncompleteFiles(t *testing.T) {
	file := completedReplayFixture(t, ModeMCR)
	file.SchemaVersion = ReplayFileSchemaVersion + 1
	if !errors.Is(ValidateReplay(file), ErrUnsupportedReplayVersion) {
		t.Fatal("unsupported version was accepted")
	}
	file.SchemaVersion = ReplayFileSchemaVersion
	file.Complete = false
	if !errors.Is(ValidateReplay(file), ErrIncompleteReplay) {
		t.Fatal("incomplete replay was accepted")
	}
}
```

Also assert frame indexes are contiguous, mode/config validation runs, participants have unique seats, at least one frame exists, the last frame agrees with `FinalStandings`, and sealing does not mutate the input.

- [x] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/game -run "ReplayFile|ReplayChecksum" -count=1
```

Expected: compile failure because `ReplayFile`, `SealReplay`, and validation errors do not exist.

- [x] **Step 3: Add replay file types**

Implement these public types in `internal/game/replay_file.go`:

```go
const ReplayFileSchemaVersion = 2

var (
	ErrUnsupportedReplayVersion = errors.New("unsupported replay version")
	ErrIncompleteReplay         = errors.New("replay is incomplete")
	ErrReplayChecksum           = errors.New("replay checksum mismatch")
	ErrInvalidReplay            = errors.New("invalid replay")
)

type ReplayParticipant struct {
	Seat int    `json:"seat"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ReplayFrame struct {
	Index   int            `json:"index"`
	Command *GameCommand   `json:"command,omitempty"`
	Match   MatchSnapshot  `json:"match"`
}

type ReplayFile struct {
	SchemaVersion     int                   `json:"schema_version"`
	ApplicationVersion string               `json:"application_version"`
	ReplayID          string                `json:"replay_id"`
	CreatedAt         time.Time             `json:"created_at"`
	Mode              RuleMode              `json:"mode"`
	RuleConfig        RuleConfig            `json:"rule_config"`
	ShuffleProof      ShuffleProof          `json:"shuffle_proof"`
	Participants      []ReplayParticipant   `json:"participants"`
	Initial           MatchSnapshot         `json:"initial"`
	Commands          []GameCommand         `json:"commands"`
	Frames            []ReplayFrame         `json:"frames"`
	MCRSettlements    []MCRSettlement       `json:"mcr_settlements,omitempty"`
	RiichiSettlements []RiichiSettlement    `json:"riichi_settlements,omitempty"`
	FinalStandings    [4]int                `json:"final_standings"`
	Complete          bool                  `json:"complete"`
	Checksum          string                `json:"checksum"`
}
```

Keep `ReplayFileSchemaVersion` separate from `ReplaySchemaVersion`. Do not rename or reinterpret the existing `ReplayLog`.

- [x] **Step 4: Implement sealing and validation**

Use canonical standard-library JSON over a copy with `Checksum == ""`:

```go
func replayChecksum(file ReplayFile) (string, error) {
	file.Checksum = ""
	data, err := json.Marshal(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func SealReplay(file ReplayFile) (ReplayFile, error) {
	if file.SchemaVersion == 0 {
		file.SchemaVersion = ReplayFileSchemaVersion
	}
	sum, err := replayChecksum(file)
	if err != nil {
		return ReplayFile{}, err
	}
	file.Checksum = sum
	if err := ValidateReplay(file); err != nil {
		return ReplayFile{}, err
	}
	return file, nil
}
```

`ValidateReplay` must check version before all other checks, require `Complete`, validate `RuleConfig` for `Mode`, verify participant seats and frame indexes, compare the last frame's points with `FinalStandings`, and compare the SHA-256 checksum with `subtle.ConstantTimeCompare`.

- [x] **Step 5: Verify and commit**

Run:

```powershell
go test ./internal/game -run "ReplayFile|ReplayChecksum" -count=20
go test ./internal/game -count=1
```

Commit:

```powershell
git add internal/game/replay_file.go internal/game/replay_file_test.go
git commit -m "feat: add validated replay file schema"
```

Step review:
- Subphase goal: define a stable, independently verifiable replay envelope.
- Evidence: sealed files validate; mutations, unsupported versions, invalid modes, gaps, and incomplete files fail with typed errors.
- Next step: record authoritative match transitions into this envelope.

## Task 2: 14B Authoritative Match Recorder

**Files:**
- Modify: `internal/game/match.go`
- Modify: `internal/game/turn.go`
- Modify: `internal/game/match_test.go`

- [ ] **Step 1: Write failing recorder tests**

Create a fixed-seed Riichi match, record the initial draw and two accepted commands, and assert:

```go
func TestMatchReplayJournalCapturesAcceptedCommandsAndFrames(t *testing.T) {
	match := mustMatch(t, 140014, NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi))
	match.EnsureCurrentTurnDraw()
	before := match.ReplayFrameCount()
	action := firstLegalDiscard(t, match.Round.Snapshot().LegalActions)
	result := match.ApplyCommand(GameCommand{
		PlayerID: "0",
		Kind: CommandDiscard,
		TileIndex: action.TileIndex,
	})
	if !result.OK {
		t.Fatal(result.Error)
	}
	if match.ReplayCommandCount() != 1 || match.ReplayFrameCount() <= before {
		t.Fatalf("journal commands=%d frames=%d", match.ReplayCommandCount(), match.ReplayFrameCount())
	}
}
```

Add tests proving rejected commands create no command/frame, draw-only frames have `Command == nil`, final-round state is captured before a next-round frame, and `ReplayFile` returns deep copies.

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/game -run "MatchReplayJournal|ReplayFrame|RejectedCommandDoesNotRecord" -count=1
```

Expected: compile failure because the match journal API does not exist.

- [ ] **Step 3: Add an unexported match journal**

Extend `Match` without exposing mutable journal slices. Add `Abandoned bool` to both `Match` and `MatchSnapshot` so quit state cannot be confused with rule-complete state:

```go
type matchReplayJournal struct {
	initial  MatchSnapshot
	commands []GameCommand
	frames   []ReplayFrame
}

type Match struct {
	Mode                 RuleMode
	RuleConfig           RuleConfig
	Points               [4]int
	Dealer               int
	RoundNumber          int
	Complete             bool
	Abandoned            bool
	Round                *Game
	LastMCRSettlement    *MCRSettlement
	LastRiichiSettlement *RiichiSettlement
	MCRSettlements       []MCRSettlement
	RiichiSettlements    []RiichiSettlement
	rules                RuleSet
	replay               matchReplayJournal
}

func (match *Match) recordReplayFrame(command *GameCommand) {
	frame := ReplayFrame{
		Index: len(match.replay.frames),
		Match: match.Snapshot(),
	}
	if command != nil {
		copyCommand := *command
		frame.Command = &copyCommand
		match.replay.commands = append(match.replay.commands, copyCommand)
	}
	match.replay.frames = append(match.replay.frames, frame)
}
```

Initialize `replay.initial` and frame zero at the end of `NewMatch`.

- [ ] **Step 4: Record draws, commands, settlement, and round transitions**

In `Match.EnsureCurrentTurnDraw`, record a nil-command frame only when state changed. In `Match.ApplyCommand`, treat an accepted quit as abandoned and do not run settlement/round advancement:

```go
func (match *Match) ApplyCommand(command GameCommand) CommandResult {
	result := match.Round.ApplyCommand(command)
	if !result.OK {
		return result
	}
	match.recordReplayFrame(&result.Command)
	if command.Kind == CommandQuit {
		match.Abandoned = true
		return result
	}
	if match.Round.Over {
		if match.Mode == ModeMCR {
			match.completeMCRRound()
		} else if match.Mode == ModeRiichi {
			match.completeRiichiRound()
		} else {
			match.Complete = true
		}
		match.recordReplayFrame(nil)
	}
	return result
}
```

The command frame preserves the completed round before settlement changes `match.Round`; the nil-command frame preserves settlement, standings, and the next-round or completed-match state.

- [ ] **Step 5: Route match AI through accepted commands**

Add `func (match *Match) AdvanceAIUntilHumanTurn()` in `turn.go`. It must:

1. return on a human claim window;
2. call `match.ApplyCommand(match.Round.aiClaimCommand())` for bot claims;
3. call `match.EnsureCurrentTurnDraw()` for bot draws;
4. prefer the first legal `CommandWin`, then first legal `CommandKong`;
5. otherwise select `ChooseAIDiscard` and call `match.ApplyCommand(CommandDiscard)`;
6. stop on match completion or when seat zero has a playable turn.

Do not reconstruct state from events and do not call `discardCurrent`, `finish`, or `resolveAIKongs` directly from the new match loop.

- [ ] **Step 6: Export a sealed completed replay**

Add:

```go
func (match *Match) ReplayFrameCount() int
func (match *Match) ReplayCommandCount() int
func (match *Match) CompletedReplay(
	applicationVersion string,
	createdAt time.Time,
	participants []ReplayParticipant,
) (ReplayFile, error)
```

`CompletedReplay` must reject `!match.Complete || match.Abandoned`, derive a deterministic `ReplayID` from mode, initial wall hash, and UTC creation timestamp, copy all frames/commands/settlements, set final points, and call `SealReplay`.

- [ ] **Step 7: Verify and commit**

Run:

```powershell
go test ./internal/game -run "MatchReplay|ReplayFrame|AdvanceMatchAI" -count=20
go test ./internal/game -count=1
```

Commit:

```powershell
git add internal/game/match.go internal/game/turn.go internal/game/match_test.go
git commit -m "feat: record authoritative match replay frames"
```

Step review:
- Subphase goal: capture replay data from the same coordinator that settles and advances matches.
- Evidence: all accepted human/bot commands and draw/settlement transitions produce deterministic copied frames; rejected commands do not.
- Next step: persist completed local matches atomically.

## Task 3: 14B Atomic Storage And Local Autosave

**Files:**
- Create: `internal/replay/store.go`
- Create: `internal/replay/store_test.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/game_flow.go`
- Modify: `internal/tui/input.go`
- Modify: `internal/tui/menu.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing atomic storage tests**

Use `t.TempDir()` and fixed UTC time. Assert `Save` creates exactly one validated `.json`, leaves no temporary file, `Load` returns canonical JSON-equivalent data, and a checksum-invalid file fails without replacing a valid destination.

Add:

```go
func TestListSkipsCorruptReplayAndKeepsValidFiles(t *testing.T) {
	dir := t.TempDir()
	saveReplayFixture(t, dir, "older", time.Unix(10, 0))
	saveReplayFixture(t, dir, "newer", time.Unix(20, 0))
	os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644)

	entries, issues, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].ReplayID != "newer" || len(issues) != 1 {
		t.Fatalf("entries=%#v issues=%#v", entries, issues)
	}
}
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/replay -count=1
```

Expected: package/functions do not exist.

- [ ] **Step 3: Implement validated atomic storage**

Create:

```go
type Entry struct {
	Path      string
	ReplayID  string
	Mode      game.RuleMode
	CreatedAt time.Time
	Frames    int
}

type FileIssue struct {
	Path string
	Err  error
}

func ApplicationVersion() string
func Save(dir string, file game.ReplayFile) (string, error)
func Load(path string) (game.ReplayFile, error)
func List(dir string) ([]Entry, []FileIssue, error)
```

`ApplicationVersion` reads `debug.ReadBuildInfo` and returns `"dev"` for an empty or `(devel)` main version. `Save` validates before writing, uses `os.MkdirAll`, `os.CreateTemp(dir, ".replay-*.tmp")`, `file.Sync`, `file.Close`, and `os.Rename`. Its final name is `<UTC timestamp>-<mode>-<replay id>.json`. Deferred cleanup removes the temporary file on all failures. `Load` applies a 32 MiB `io.LimitReader`, rejects trailing JSON values, then calls `game.ValidateReplay`. `List` treats a missing directory as empty, validates each `.json`, records per-file issues, and sorts entries by `CreatedAt` descending then path.

- [ ] **Step 4: Retain the local Match coordinator**

Add these model fields:

```go
LocalMatch     *game.Match
ReplayDir      string
LastReplayPath string
```

Set `ReplayDir: "replays"` in `NewModel`. Add:

```go
func newStartedMatchWithRules(mode game.RuleMode, config game.RuleConfig) *game.Match

func syncLocalRound(m Model) Model {
	if m.LocalMatch != nil {
		m.Game = m.LocalMatch.Round
	}
	return m
}
```

Keep `newStartedGameWithRules` as a compatibility helper that returns `newStartedMatchWithRules(...).Round`.

- [ ] **Step 5: Route local actions through Match**

Replace direct local calls to `HumanDiscardSelected`, `Game.ApplyCommand`, and `Game.AdvanceAIUntilHumanTurn` with:

```go
func applyLocalCommand(m Model, command game.GameCommand) (Model, game.CommandResult) {
	command.PlayerID = "0"
	result := m.LocalMatch.ApplyCommand(command)
	m = syncLocalRound(m)
	return m, result
}

func advanceLocalAI(m Model) Model {
	m.LocalMatch.AdvanceAIUntilHumanTurn()
	return syncLocalRound(m)
}
```

For MCR/Riichi, remain on `ScreenTable` after a round transition and reset selection/claim indexes. Enter `ScreenGameOver` only when `LocalMatch.Complete`. Compatibility mode remains a one-round match.

Route `Q` through `LocalMatch.ApplyCommand(CommandQuit)`, enter `ScreenGameOver`, and do not schedule autosave because `LocalMatch.Abandoned` is true. The game-over screen must say the match was abandoned rather than claiming a replay was saved.

- [ ] **Step 6: Save a completed local replay asynchronously**

Define:

```go
type replaySavedMsg struct {
	Path string
}

type replaySaveErrorMsg struct {
	Err error
}

func saveCompletedReplayCmd(match *game.Match, dir string) tea.Cmd
```

Build participants from the four player IDs/names, call `CompletedReplay(replay.ApplicationVersion(), time.Now().UTC(), participants)`, then `replay.Save`. On `replaySavedMsg`, store `LastReplayPath` and show a localized success status. A save failure must not erase match results or quit the TUI.

- [ ] **Step 7: Verify and commit**

Run:

```powershell
go test ./internal/replay -count=20
go test ./internal/tui -run "LocalMatch|LocalReplay|GameOver" -count=20
go test ./internal/game ./internal/replay ./internal/tui -count=1
```

Commit:

```powershell
git add internal/replay internal/tui/model.go internal/tui/game_flow.go internal/tui/input.go internal/tui/menu.go internal/tui/model_test.go
git commit -m "feat: save completed local match replays"
```

14B phase review:
- Total goal: completed local and online matches must become durable, auditable replays without a second rules engine.
- Achieved: the local TUI now retains the authoritative `Match`, journals every transition, and atomically saves only sealed completed files.
- Evidence: deterministic journal tests, atomic storage tests, and local multi-round/game-over tests.
- Remaining risk: online completion, privacy, delivery, and reconnect retrieval are not yet connected.

## Task 4: 14C Online Replay Privacy And Server Delivery

**Files:**
- Modify: `internal/protocol/message.go`
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`

- [ ] **Step 1: Write failing protocol/privacy tests**

Add protocol round-trip tests and a WebSocket integration test that:

1. creates/starts a room;
2. verifies all pre-completion messages have `Replay == nil` and empty `ReplayID`;
3. places the room's match in a legal final-hand fixture;
4. sends the winning command;
5. receives `replay_ready` followed by `replay_data`;
6. validates the replay checksum and full concealed hands;
7. verifies no reconnect token appears in marshaled replay JSON.

Add a request test where a disconnected player reconnects inside the existing window, sends `request_replay`, and receives byte-equivalent replay data. Invalid room/session and incomplete match requests must return `MsgError`.

The pre-completion privacy test must exercise the real message builder:

```go
func TestReplayPrivacyBeforeCompletion(t *testing.T) {
	match, err := game.NewMatch(140014, game.NewRiichiRuleSet(game.DefaultRuleConfig(game.ModeRiichi).Riichi))
	if err != nil {
		t.Fatal(err)
	}
	room := &room{code: "140014", match: match}
	session := &session{playerID: "player", seat: 0}

	message := stateMessageForSession(room, session, protocol.Message{Type: protocol.MsgGameSnapshot})

	if message.Replay != nil || message.ReplayID != "" {
		t.Fatalf("live message leaked replay: %#v", message)
	}
	if message.Snapshot.Seed != 0 || message.Snapshot.Players[1].Hand != nil {
		t.Fatalf("live privacy regressed: %#v", message.Snapshot)
	}
}
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/online -run "ReplayDelivery|ReplayPrivacy|ReplayReconnect" -count=1
```

Expected: compile failure because replay protocol messages and room replay retention do not exist.

- [ ] **Step 3: Extend the JSON protocol**

Add:

```go
const (
	MsgReplayReady  MessageType = "replay_ready"
	MsgRequestReplay MessageType = "request_replay"
	MsgReplayData   MessageType = "replay_data"
)

type Message struct {
	Type           MessageType        `json:"type"`
	PlayerID       string             `json:"player_id,omitempty"`
	ReconnectToken string             `json:"reconnect_token,omitempty"`
	RoomCode       string             `json:"room_code,omitempty"`
	Name           string             `json:"name,omitempty"`
	Seat           int                `json:"seat,omitempty"`
	ReadySeats     []int              `json:"ready_seats,omitempty"`
	Started        bool               `json:"started,omitempty"`
	OccupiedSeats  []int              `json:"occupied_seats,omitempty"`
	Command        game.GameCommand   `json:"command,omitempty"`
	Result         game.CommandResult `json:"result,omitempty"`
	Snapshot       game.GameSnapshot  `json:"snapshot,omitempty"`
	Mode           game.RuleMode      `json:"mode,omitempty"`
	RuleConfig     game.RuleConfig    `json:"rule_config,omitempty"`
	Match          game.MatchSnapshot `json:"match,omitempty"`
	Rooms          []RoomSummary      `json:"rooms,omitempty"`
	ReplayID       string             `json:"replay_id,omitempty"`
	Replay         *game.ReplayFile   `json:"replay,omitempty"`
	Error          string             `json:"error,omitempty"`
}
```

Do not add session or address fields to `game.ReplayFile`.

- [ ] **Step 4: Retain a sealed replay on completed rooms**

Extend `room`:

```go
replay *game.ReplayFile
```

Add:

```go
func (s *Server) ensureCompletedReplayLocked(room *room) error
func (s *Server) replayParticipantsLocked(room *room) []game.ReplayParticipant
```

After `room.match.ApplyCommand` and bot advancement, call `ensureCompletedReplayLocked`. If `room.match.Complete && room.replay == nil`, create participants from occupied sessions plus stable names for bot seats, call `CompletedReplay(replay.ApplicationVersion(), time.Now().UTC(), participants)`, validate it, and retain a pointer to the sealed value. Before completion, `room.replay` must remain nil.

- [ ] **Step 5: Deliver and retrieve completed replay data**

Add `MsgRequestReplay` handling that requires a joined/reconnected session and a completed room replay. After the final `MsgGameSnapshot`, broadcast:

```go
protocol.Message{Type: protocol.MsgReplayReady, ReplayID: room.replay.ReplayID}
protocol.Message{Type: protocol.MsgReplayData, ReplayID: room.replay.ReplayID, Replay: room.replay}
```

The sealed `room.replay` value is immutable after assignment, so it may be serialized directly. Include only `ReplayID` in `MsgReconnected` when a completed replay is available; send the payload only in response to `MsgRequestReplay` or the original completion broadcast.

- [ ] **Step 6: Verify and commit**

Run:

```powershell
go test ./internal/online -run "ReplayDelivery|ReplayPrivacy|ReplayReconnect" -count=20
go test ./internal/protocol ./internal/online -count=1
```

Commit:

```powershell
git add internal/protocol/message.go internal/online/server.go internal/online/server_test.go
git commit -m "feat: deliver completed online replays"
```

Step review:
- Subphase goal: reveal full replay data only after authoritative match completion.
- Evidence: live messages remain private; completion and reconnect requests return the same sealed full-information payload.
- Next step: make TUI/CLI clients request and save that payload reliably.

## Task 5: 14C Client Retrieval And Saving

**Files:**
- Modify: `internal/online/client.go`
- Modify: `internal/online/client_test.go`
- Modify: `internal/tui/online.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/model_test.go`
- Modify: `cmd/client/main.go`
- Modify: `cmd/client/main_test.go`

- [ ] **Step 1: Write failing client tests**

Add:

```go
func TestClientRequestsCompletedReplay(t *testing.T) {
	server, wsURL, session := completedReplayClientFixture(t)
	defer server.Close()

	client := NewClient(wsURL, session.Name)
	defer client.Close()
	if _, err := client.Reconnect(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	if err := client.RequestReplay(context.Background()); err != nil {
		t.Fatal(err)
	}
	message, err := client.ReadUntil(context.Background(), 2*time.Second, protocol.MsgReplayData)
	if err != nil {
		t.Fatal(err)
	}
	if message.Replay == nil {
		t.Fatal("replay payload missing")
	}
	if err := game.ValidateReplay(*message.Replay); err != nil {
		t.Fatal(err)
	}
}
```

`completedReplayClientFixture` must use the existing `httptest` WebSocket server pattern, finish a room through an accepted final command, disconnect the client, and return its saved `ClientSession`; it must not assign `room.replay` directly.

Add TUI tests that `MsgReplayReady` triggers one request, `MsgReplayData` saves under `ReplayDir`, and reconnect with `ReplayID` requests the payload again. Add a CLI test that `-watch -replay-dir <temp>` saves one validated file.

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/online ./internal/tui ./cmd/client -run "RequestReplay|ReplayData|ReplayDir" -count=1
```

- [ ] **Step 3: Add the client request**

Implement:

```go
func (c *Client) RequestReplay(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.write(protocol.Message{Type: protocol.MsgRequestReplay})
}
```

Use the client's existing `connect` and `write` methods; do not open a second WebSocket.

- [ ] **Step 4: Handle replay messages in the TUI**

Include `MsgReplayReady` and `MsgReplayData` in `waitOnlineSnapshot`. Add:

```go
type onlineReplayReadyMsg struct{ ReplayID string }
type onlineReplayDataMsg struct{ Replay game.ReplayFile }
```

On ready/reconnect availability, issue `requestOnlineReplayCmd` once per ID. On data, validate then reuse the same atomic save command as local replay storage. Keep game-over state visible if saving fails.

- [ ] **Step 5: Save replay data in the CLI watcher**

Add `-replay-dir` with default `replays`. When watch mode receives `MsgReplayData`, call `replay.Save`, print `replay=<path>`, and continue reading until context cancellation. Do not print the full payload or token.

- [ ] **Step 6: Verify and commit**

Run:

```powershell
go test ./internal/online ./internal/tui ./cmd/client -run "RequestReplay|ReplayData|ReplayDir" -count=20
go test ./internal/online ./internal/tui ./cmd/client -count=1
```

Commit:

```powershell
git add internal/online/client.go internal/online/client_test.go internal/tui/online.go internal/tui/model.go internal/tui/model_test.go cmd/client/main.go cmd/client/main_test.go
git commit -m "feat: save received online replays"
```

14C phase review:
- Total goal: completed online matches must be replayable without exposing hidden information during live play.
- Achieved: the server owns full replay creation; clients receive/save it only after completion and can request it again after reconnect.
- Evidence: live privacy, completion delivery, checksum, reconnect retrieval, TUI save, and CLI save tests.
- Remaining risk: files exist but users cannot browse or play them inside the TUI yet.

## Task 6: 14D Replay Browser

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/menu.go`
- Modify: `internal/tui/i18n.go`
- Create: `internal/tui/replay.go`
- Create: `internal/tui/replay_test.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing browser tests**

Require a localized `回放 / Replays` menu item. With two valid files and one corrupt file, Enter must open `ScreenReplayBrowser`, display newest first, show the skipped-file count without failing the screen, support Up/Down selection, Enter load, R refresh, and Esc return.

Use the real Tea command result:

```go
func TestReplayBrowserListsNewestValidFiles(t *testing.T) {
	dir := t.TempDir()
	saveTUIReplayFixture(t, dir, game.ModeMCR, time.Unix(10, 0))
	newest := saveTUIReplayFixture(t, dir, game.ModeRiichi, time.Unix(20, 0))
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	message := listReplaysCmd(dir)().(replayListMsg)
	model := NewModel()
	model.ReplayDir = dir
	next, _ := model.Update(message)
	updated := next.(Model)
	if len(updated.ReplayEntries) != 2 || updated.ReplayEntries[0].Path != newest || len(updated.ReplayIssues) != 1 {
		t.Fatalf("entries=%#v issues=%#v", updated.ReplayEntries, updated.ReplayIssues)
	}
}
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/tui -run "ReplayMenu|ReplayBrowser|ReplayList" -count=1
```

- [ ] **Step 3: Add replay screens and model state**

Add:

```go
const (
	ScreenMenu Screen = iota
	ScreenJoinOnline
	ScreenOnlineRooms
	ScreenHelp
	ScreenTable
	ScreenGameOver
	ScreenReplayBrowser
	ScreenReplayViewer
)
```

Add these fields to `Model`:

```go
	ReplayEntries     []replay.Entry
	ReplayIssues      []replay.FileIssue
	ReplayIndex       int
	ReplayFile        *game.ReplayFile
	ReplayFrame       int
	ReplayPlaying     bool
	ReplayShowDetails bool
	ReplayRequestedID string
```

Keep `ReplayDir` from Task 3.

- [ ] **Step 4: Add asynchronous list/load commands**

Implement:

```go
type replayListMsg struct {
	Entries []replay.Entry
	Issues  []replay.FileIssue
}

type replayLoadedMsg struct {
	File game.ReplayFile
}

func listReplaysCmd(dir string) tea.Cmd
func loadReplayCmd(path string) tea.Cmd
```

All file reads occur inside Tea commands. Errors become localized status lines and do not close the browser.

- [ ] **Step 5: Render and navigate the browser**

Show timestamp, rule mode, frame count, replay ID, and path basename. The selected row uses the existing selected style. Keep lines within `m.Width`; truncate only path basenames, never replay IDs or mode. Add bilingual controls: Up/Down, Enter, R, Esc.

Implement these exact screen entry points:

```go
func updateReplayBrowser(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd)
func renderReplayBrowser(m Model) string
func replayEntryLabel(m Model, entry replay.Entry) string
```

`updateReplayBrowser` clamps `ReplayIndex` after every refresh, returns `loadReplayCmd` only for a valid selected entry, and never deletes or rewrites files.

- [ ] **Step 6: Verify and commit**

Run:

```powershell
go test ./internal/tui -run "ReplayMenu|ReplayBrowser|ReplayList" -count=20
go test ./internal/tui -count=1
```

Commit:

```powershell
git add internal/tui/model.go internal/tui/menu.go internal/tui/i18n.go internal/tui/replay.go internal/tui/replay_test.go internal/tui/model_test.go
git commit -m "feat: add replay browser"
```

Step review:
- Subphase goal: make valid saved replays discoverable without letting one corrupt file block the library.
- Evidence: newest-first listing, isolated errors, bilingual controls, and safe asynchronous loading.
- Next step: add read-only frame navigation and playback.

## Task 7: 14D Read-Only Frame Navigation And Playback

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/input.go`
- Modify: `internal/tui/replay.go`
- Modify: `internal/tui/replay_test.go`

- [ ] **Step 1: Write failing viewer control tests**

Load a three-frame fixture and assert:

- Right increments but never exceeds the final frame.
- Left decrements but never goes below zero.
- Home/End jump to first/last.
- Space toggles playback.
- A synthetic `replayTickMsg` advances one frame and schedules the next tick.
- Playback stops at the last frame.
- Tab toggles detail mode.
- Esc returns to the replay browser.
- Discard/win/kong/claim keys never mutate recorded snapshots or commands.

The read-only test must compare canonical JSON before and after unrelated live-game keys:

```go
func TestReplayViewerIgnoresLiveGameCommands(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))
	before, err := json.Marshal(model.ReplayFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"h", "k", "l", "p", "c", "q"} {
		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		model = next.(Model)
	}
	after, err := json.Marshal(model.ReplayFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("replay viewer mutated recorded data")
	}
}
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/tui -run "ReplayViewer|ReplayPlayback|ReplayReadOnly" -count=1
```

- [ ] **Step 3: Add viewer input and timed ticks**

Implement:

```go
type replayTickMsg struct{}

func replayTickCmd() tea.Cmd {
	return tea.Tick(750*time.Millisecond, func(time.Time) tea.Msg {
		return replayTickMsg{}
	})
}
```

`updateReplayViewer` handles only Left, Right, Home, End, Space, Tab, and Esc. `Model.Update` handles `replayTickMsg` only when `ScreenReplayViewer && ReplayPlaying`. Never call game commands from this screen.

- [ ] **Step 4: Add the viewer shell**

Render a frame header containing localized mode, `current/total`, paused/playing state, creation time, and replay ID. Render the shared table body below it and a single replay control line. At the final frame show final standings and stop playback.

Add:

```go
func updateReplayViewer(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd)
func applyReplayTick(m Model) (Model, tea.Cmd)
func renderReplayViewer(m Model) string
func currentReplayFrame(m Model) (game.ReplayFrame, bool)
```

`currentReplayFrame` returns a copied frame and false for nil/empty/out-of-range state. Both update functions clamp before indexing.

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./internal/tui -run "ReplayViewer|ReplayPlayback|ReplayReadOnly" -count=20
go test ./internal/tui -count=1
```

Commit:

```powershell
git add internal/tui/model.go internal/tui/input.go internal/tui/replay.go internal/tui/replay_test.go
git commit -m "feat: add replay frame playback"
```

Step review:
- Subphase goal: navigate authoritative frames without replaying rules or mutating state.
- Evidence: bounded navigation, deterministic ticks, read-only key isolation, and final-frame stopping.
- Next step: reuse the Phase 13 table and expose full post-game information clearly.

## Task 8: 14D Shared Table Rendering And Full Details

**Files:**
- Modify: `internal/tui/table_state.go`
- Modify: `internal/tui/table_components.go`
- Modify: `internal/tui/replay.go`
- Modify: `internal/tui/replay_test.go`
- Modify: `internal/tui/layout_test.go`
- Modify: `internal/tui/i18n.go`

- [ ] **Step 1: Write failing shared-renderer tests**

Require the replay view to contain the same four seat labels, discard rivers, points, and hand tray as the Phase 13 table. Require full concealed hands for all four players in replay detail mode, mode-specific dora/ura or flowers, settlement deltas, final standings, and no live action labels such as discard/win/kong.

Add 140-, 96-, and 80-column line-width tests in Chinese and English.

```go
func TestReplayDetailsRevealAllHandsOnlyInCompletedReplay(t *testing.T) {
	model := replayViewerModel(t, completedRiichiReplay(t))
	model.ReplayShowDetails = true
	view := renderReplayViewer(model)
	frame, _ := currentReplayFrame(model)
	for _, player := range frame.Match.Round.Players {
		for _, tile := range player.Hand {
			if !strings.Contains(view, game.TileLabel(tile, model.UnicodeTiles)) {
				t.Fatalf("detail view missing tile %s for %s:\n%s", tile, player.Name, view)
			}
		}
	}
	if strings.Contains(view, "[H]") || strings.Contains(view, "[K]") {
		t.Fatalf("replay exposed live actions:\n%s", view)
	}
}
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/tui -run "ReplayTable|ReplayDetails|ReplayWidth" -count=1
```

- [ ] **Step 3: Normalize replay frames into table state**

Extend:

```go
type tableViewState struct {
	Snapshot   game.GameSnapshot
	Match      game.MatchSnapshot
	ViewerSeat int
	Mode       game.RuleMode
	Online     bool
	Started    bool
	RoomCode   string
	Replay     bool
	ReadOnly   bool
}
```

When `ScreenReplayViewer`, `tableStateFor` selects `ReplayFile.Frames[ReplayFrame].Match`, uses viewer seat zero only for orientation, and preserves all four hands from the full snapshot.

- [ ] **Step 4: Add replay-specific controls and detail rail**

When `state.ReadOnly`:

- replace normal table actions with `←/→ frame | Home/End | Space play/pause | Tab details | Esc library`;
- do not build tactical recommendations from future/full hidden information;
- render a replay detail rail with all four hands, the command attached to the frame, new events since the previous frame, round settlement, and final standings;
- use `ReplayShowDetails` to switch between the normal table and the detail/settlement presentation.

Use bare Unicode Mahjong tiles and existing ANSI-safe width helpers. Bound event rows and hand summaries explicitly; do not rely on terminal wrapping.

Add these render boundaries:

```go
func renderReplayControls(m Model) string
func renderReplayDetailRail(m Model, file game.ReplayFile, frame game.ReplayFrame) string
func renderReplayHands(m Model, players []game.PlayerView) string
func renderReplaySettlement(m Model, frame game.ReplayFrame) string
```

`renderWideHandAndActions` selects `renderReplayControls` when `state.ReadOnly`. `renderWideTable` selects `renderReplayDetailRail` instead of `renderTacticalRail` for replay state, preventing tactical calculations from using revealed future information.

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./internal/tui -run "ReplayTable|ReplayDetails|ReplayWidth" -count=20
go test ./internal/tui -count=1
```

Commit:

```powershell
git add internal/tui/table_state.go internal/tui/table_components.go internal/tui/replay.go internal/tui/replay_test.go internal/tui/layout_test.go internal/tui/i18n.go
git commit -m "feat: render full replay details"
```

14D phase review:
- Total goal: saved full-information matches must be understandable and controllable from the terminal.
- Achieved: users can browse, load, step, play, pause, inspect all hands, inspect settlements, and return safely without entering a live command path.
- Evidence: shared-renderer, localization, read-only, playback, and width tests.
- Remaining risk: deterministic cross-mode fixtures, corruption matrices, network completeness, and visual acceptance still need the 14E gate.

## Task 9: 14E Dual-Mode Acceptance, Visual Evidence, And Documentation

**Files:**
- Create: `internal/game/replay_acceptance_test.go`
- Create: `internal/online/replay_acceptance_test.go`
- Create: `internal/tui/replay_screenshot_test.go`
- Create: `artifacts/phase14/` generated visual artifacts
- Modify: `README.md`
- Modify: `docs/workflow.md`
- Modify: `docs/superpowers/plans/2026-06-29-phase-14-full-post-game-replay.md`

- [ ] **Step 1: Add deterministic dual-mode replay fixtures**

For fixed seeds, create completed MCR and Riichi match fixtures. For every frame assert:

- indexes are contiguous;
- canonical JSON remains identical after save/load;
- frame snapshots equal the authoritative source snapshots captured during play;
- command order equals accepted command order;
- final settlements and standings equal the source match;
- replay checksum is stable for fixed creation time/application version.

```go
func TestDualModeReplayAcceptance(t *testing.T) {
	for _, fixture := range dualModeReplayFixtures(t) {
		t.Run(string(fixture.mode), func(t *testing.T) {
			source, file := fixture.completedMatchAndReplay(t)
			loaded := saveAndLoadReplay(t, file)
			assertCanonicalJSONEqual(t, file, loaded)
			assertReplayFramesEqualSource(t, source, loaded)
			if loaded.FinalStandings != source.Points {
				t.Fatalf("standings=%v source=%v", loaded.FinalStandings, source.Points)
			}
		})
	}
}
```

- [ ] **Step 2: Add corruption and compatibility matrices**

Table-test empty files, truncated JSON, trailing JSON, checksum mutation, missing frames, frame-index gaps, invalid mode/config, incomplete match, and schema versions `1` and `3`. The browser must skip each bad file, retain valid files, and expose a bounded issue count.

```go
func TestReplayCorruptionMatrix(t *testing.T) {
	for _, fixture := range replayCorruptionFixtures(t) {
		t.Run(fixture.name, func(t *testing.T) {
			err := fixture.load(t)
			if !errors.Is(err, fixture.want) {
				t.Fatalf("error=%v want=%v", err, fixture.want)
			}
		})
	}
}
```

- [ ] **Step 3: Add online completeness and privacy acceptance**

For both modes:

- no live message contains replay payload, full opponent hands, shuffle seed, or ura indicators;
- completion payload validates and contains all full hands, proof, commands, frames, settlement, and standings;
- reconnect request returns canonical JSON-equivalent data;
- replay JSON contains no `reconnect_token`, WebSocket URL, IP address, or session path.

Marshal every received `MsgReplayData.Replay` and scan the JSON bytes for those forbidden field names before comparing it with the server's retained sealed replay.

- [ ] **Step 4: Generate deterministic visual artifacts**

Add an environment-gated generator:

```powershell
$env:MAHJONG_PHASE14_CAPTURE_DIR = Join-Path (Get-Location) 'artifacts\phase14'
go test ./internal/tui -run TestGeneratePhase14ReplaySnapshots -count=1
```

Generate:

- `riichi-replay-wide-zh.html`;
- `riichi-replay-wide-en.html`;
- `mcr-replay-wide-zh.html`;
- `mcr-replay-wide-en.html`;
- `riichi-replay-details-zh.html`;
- `riichi-replay-fallback-80.html`.

Each generator case uses a fixed replay fixture and asserts at most 42 lines and no line wider than the requested viewport. Inspect a real PTY when available. Do not bypass desktop/browser security policy to manufacture a PNG.

- [ ] **Step 5: Run full acceptance**

Run:

```powershell
go test ./internal/game -run "ReplayFile|ReplayFrame|ReplayAcceptance" -count=20
go test ./internal/replay -count=20
go test ./internal/online -run "Replay|Private|Reconnect" -count=20
go test ./internal/tui ./cmd/client -run "Replay|Width|Menu|GameOver" -count=20
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go build ./cmd/mahjong ./cmd/server ./cmd/client
gofmt -l internal cmd
git diff --check
```

- [ ] **Step 6: Update user documentation**

Document:

- completed replay location and filename;
- local and online autosave behavior;
- browser and viewer controls;
- checksum/unsupported-version/corrupt-file behavior;
- full-information post-game privacy boundary;
- the difference between `ReplayLog` event summaries and `ReplayFile` frame replays.

- [ ] **Step 7: Record reviews and commit**

Update `docs/workflow.md` with each 14A-E step review, each subphase review, exact commands, and artifact paths. Mark all plan checkboxes complete only after the commands pass.

Commit:

```powershell
git add internal/game/replay_acceptance_test.go internal/online/replay_acceptance_test.go internal/tui/replay_screenshot_test.go artifacts/phase14 README.md docs/workflow.md docs/superpowers/plans/2026-06-29-phase-14-full-post-game-replay.md
git commit -m "test: record phase 14 replay acceptance"
```

14E phase review:
- Total goal: provide an auditable terminal Mahjong client with complete rules, private live networking, a readable tactical table, and durable post-game replay.
- Acceptance question: do saved/restored frames, settlements, standings, privacy boundaries, controls, and visual layouts match their authoritative source in both rule modes?
- Exit criterion: begin no later phase until all functional, race, build, privacy, corruption, width, and visual-artifact checks pass.

## Plan Self-Review

- Spec coverage: 14A schema/checksum, 14B recorder/local atomic save, 14C post-completion online delivery/reconnect, 14D browser/viewer/shared renderer, and 14E dual-mode/privacy/corruption/visual acceptance each have explicit tasks.
- Placeholder scan: no placeholder, deferred implementation marker, or unspecified test step remains.
- Type consistency: `ReplayFileSchemaVersion == 2` is separate from `ReplaySchemaVersion == 3`; all frames use existing `MatchSnapshot`; persistence consumes `game.ReplayFile`; protocol transports `*game.ReplayFile`; TUI viewer consumes `replay.Entry` and `game.ReplayFile`.
- Authority check: only `game.Match` records frames and accepted commands. `internal/replay`, protocol, clients, and TUI never recompute rules.
- Privacy check: live snapshots remain recipient-filtered; full hands/seeds/ura appear only in a sealed completed payload.
- Scope check: no database, Redis, accounts, rankings, replay editing, video export, GUI, or new Mahjong rules are included.
- Migration check: `Model.Game` remains the current-round alias while `LocalMatch` becomes authoritative; accepted quit marks `Match.Abandoned` and cannot be sealed as a completed replay.
