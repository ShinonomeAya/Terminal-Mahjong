package game

import (
	"bufio"
	"strings"
	"testing"
)

func TestNewGameDealsThirteenTilesToEachPlayer(t *testing.T) {
	game := NewGame(1)
	if len(game.Wall) != 84 {
		t.Fatalf("wall length = %d, want 84", len(game.Wall))
	}
	for i, player := range game.Players {
		if len(player.Hand) != 13 {
			t.Fatalf("player %d hand length = %d, want 13", i, len(player.Hand))
		}
	}
}

func TestChooseAIDiscardReturnsValidIndex(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "E", "E")
	index := ChooseAIDiscard(hand)
	if index < 0 || index >= len(hand) {
		t.Fatalf("discard index = %d, want valid hand index", index)
	}
}

func TestConcealedKongTileFindsFourCopies(t *testing.T) {
	player := Player{Hand: mustTiles(t, "1m", "1m", "1m", "1m", "E")}
	tile, ok := concealedKongTile(player)
	if !ok {
		t.Fatal("expected concealed kong to be found")
	}
	if tile.String() != "1m" {
		t.Fatalf("kong tile = %s, want 1m", tile)
	}
}

func TestConcealedKongTileRejectsTriplet(t *testing.T) {
	player := Player{Hand: mustTiles(t, "1m", "1m", "1m", "E")}
	_, ok := concealedKongTile(player)
	if ok {
		t.Fatal("did not expect concealed kong for three copies")
	}
}

func TestPlayCanQuitAtFirstPrompt(t *testing.T) {
	game := NewGame(1)
	var out strings.Builder
	game.Play(strings.NewReader("q\n"), &out)
	if !game.Over {
		t.Fatal("expected game to be over")
	}
	if game.Reason != "quit" {
		t.Fatalf("reason = %q, want quit\nOutput:\n%s", game.Reason, out.String())
	}
}

func TestParseActionDiscardByNumber(t *testing.T) {
	action, err := ParseAction("3")
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionDiscard || action.Index != 2 {
		t.Fatalf("action = %#v, want discard index 2", action)
	}
}

func TestParseActionChow(t *testing.T) {
	action, err := ParseAction("c 2m 3m")
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionChow || len(action.Tiles) != 2 {
		t.Fatalf("action = %#v, want chow with two tiles", action)
	}
}

func TestHumanCanWinOnDiscard(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n"), &out, 1, mustTile(t, "E"))
	if !claimed {
		t.Fatal("expected discard to be claimed")
	}
	if !game.Over || game.Winner != 0 {
		t.Fatalf("expected human discard-win, winner=%d over=%v\nOutput:\n%s", game.Winner, game.Over, out.String())
	}
}

func TestHumanCanPongAndDiscard(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Players[0].Hand = mustTiles(t, "1m", "1m", "2m", "3m", "4m", "5p", "6p", "7p", "1s", "2s", "3s", "E", "F")
	game.Players[1].Hand = mustTiles(t, "9m", "9m", "9m", "2p", "2p", "2p", "4s", "5s", "6s", "E", "E", "B", "B")
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n1\n"), &out, 1, mustTile(t, "1m"))
	if !claimed {
		t.Fatal("expected pong to be claimed")
	}
	if len(game.Players[0].Melds) != 1 || game.Players[0].Melds[0].Kind != MeldPong {
		t.Fatalf("expected one pong meld, got %#v", game.Players[0].Melds)
	}
	if len(game.Players[0].Discards) != 1 {
		t.Fatalf("expected human to discard after pong, got %d discards\nOutput:\n%s", len(game.Players[0].Discards), out.String())
	}
}

func TestCanChowPreviousDiscardWithTwoHandTiles(t *testing.T) {
	player := Player{Hand: mustTiles(t, "2m", "4m", "E")}
	options := ChowOptions(player, mustTile(t, "3m"))
	if len(options) != 1 || FormatTiles(options[0]) != "2m 3m 4m" {
		t.Fatalf("options = %#v, want 2m 3m 4m", options)
	}
}

func TestCannotChowHonorTile(t *testing.T) {
	player := Player{Hand: mustTiles(t, "E", "E")}
	options := ChowOptions(player, mustTile(t, "E"))
	if len(options) != 0 {
		t.Fatalf("options = %#v, want none", options)
	}
}

func TestHumanCanChowPreviousDiscardAndDiscard(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Current = 3
	game.Players[0].Hand = mustTiles(t, "2m", "4m", "5p", "6p", "7p", "1s", "2s", "3s", "E", "E", "F", "F", "B")
	game.Players[1].Hand = mustTiles(t, "9m", "9m", "2p", "2p", "2p", "4s", "5s", "6s", "E", "E", "W", "W", "B")
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n1\n"), &out, 3, mustTile(t, "3m"))
	if !claimed {
		t.Fatal("expected chow to be claimed")
	}
	if len(game.Players[0].Melds) != 1 || game.Players[0].Melds[0].Kind != MeldChow {
		t.Fatalf("expected one chow meld, got %#v", game.Players[0].Melds)
	}
	if len(game.Players[0].Discards) != 1 {
		t.Fatalf("expected human to discard after chow, got %d discards\nOutput:\n%s", len(game.Players[0].Discards), out.String())
	}
}

func TestCannotChowNonPreviousDiscard(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Players[0].Hand = mustTiles(t, "2m", "4m", "5p", "6p", "7p", "1s", "2s", "3s", "E", "E", "F", "F", "B")
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n1\n"), &out, 1, mustTile(t, "3m"))
	if claimed {
		t.Fatalf("did not expect chow from non-previous discard\nOutput:\n%s", out.String())
	}
}

func TestAIChowChoosesLowUsefulnessTiles(t *testing.T) {
	player := Player{Hand: mustTiles(t, "2m", "4m", "2p", "3p", "4p", "E", "E")}
	options := ChowOptions(player, mustTile(t, "3m"))
	index, ok := shouldAIChow(player, mustTile(t, "3m"), options)
	if !ok || index != 0 {
		t.Fatalf("index=%d ok=%v, want option 0", index, ok)
	}
}

func TestClaimPriorityWinBeatsPong(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s",
		"E", "E",
	)
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n"), &out, 1, mustTile(t, "7s"))
	if !claimed {
		t.Fatal("expected discard to be claimed")
	}
	if !game.Over || game.Winner != 0 {
		t.Fatalf("expected win claim before pong, winner=%d over=%v\nOutput:\n%s", game.Winner, game.Over, out.String())
	}
	if len(game.Players[0].Melds) != 0 {
		t.Fatalf("expected no pong meld when win is claimed, got %#v", game.Players[0].Melds)
	}
}

func TestClaimPriorityPongBeatsChow(t *testing.T) {
	game := NewGame(1)
	game.Players = NewPlayers()
	game.Current = 3
	game.Players[0].Hand = mustTiles(t, "2m", "4m", "5p", "6p", "7p", "1s", "2s", "3s", "E", "E", "F", "F", "B")
	game.Players[1].Hand = mustTiles(t, "3m", "3m", "9m", "2p", "2p", "2p", "4s", "5s", "6s", "E", "W", "W", "B")
	var out strings.Builder
	claimed := game.resolveDiscardClaims(bufioReader("y\n1\n"), &out, 3, mustTile(t, "3m"))
	if !claimed {
		t.Fatal("expected discard to be claimed")
	}
	if len(game.Players[1].Melds) != 1 || game.Players[1].Melds[0].Kind != MeldPong {
		t.Fatalf("expected player 1 pong before player 0 chow, got %#v\nOutput:\n%s", game.Players[1].Melds, out.String())
	}
	if len(game.Players[0].Melds) != 0 {
		t.Fatalf("expected player 0 chow not to happen, got %#v", game.Players[0].Melds)
	}
}

func TestHumanConcealedKongCommandDrawsReplacement(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t, "1m", "1m", "1m", "1m", "2m", "3m", "4m", "5p", "6p", "7p", "E", "E", "F", "B")
	game.Wall = mustTiles(t, "9s")
	var out strings.Builder
	if !game.tryHumanKong("1m", &out) {
		t.Fatal("expected concealed kong command to succeed")
	}
	if len(game.Players[0].Melds) != 1 || game.Players[0].Melds[0].Kind != MeldKong {
		t.Fatalf("expected one kong meld, got %#v", game.Players[0].Melds)
	}
	if game.Players[0].Count(mustTile(t, "9s")) != 1 {
		t.Fatalf("expected replacement draw in hand\nOutput:\n%s", out.String())
	}
}

func TestScoreSelfDraw(t *testing.T) {
	result := ScoreRound(WinContext{WinType: WinSelfDraw, Melds: []Meld{}})
	if result.Points != 2 || result.Label != "self-draw +2" {
		t.Fatalf("result = %#v", result)
	}
}

func TestScoreKongBonus(t *testing.T) {
	result := ScoreRound(WinContext{
		WinType: WinDiscard,
		Melds:   []Meld{{Kind: MeldKong, Tiles: mustTiles(t, "1m", "1m", "1m", "1m")}},
	})
	if result.Points != 2 {
		t.Fatalf("points = %d, want 2", result.Points)
	}
}

func TestPrintResultIncludesScore(t *testing.T) {
	game := NewGame(1)
	game.finish(0, "discard-win on 5s from AI-1", WinDiscard)
	var out strings.Builder
	game.printResult(&out)
	output := out.String()
	if !strings.Contains(output, "Winner: You") || !strings.Contains(output, "Win: discard-win on 5s from AI-1") || !strings.Contains(output, "Score: discard-win +1") {
		t.Fatalf("result output missing settlement details:\n%s", output)
	}
}

func mustTile(t *testing.T, text string) Tile {
	t.Helper()
	tile, ok := ParseTile(text)
	if !ok {
		t.Fatalf("bad tile in test: %s", text)
	}
	return tile
}

func bufioReader(text string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(text))
}
