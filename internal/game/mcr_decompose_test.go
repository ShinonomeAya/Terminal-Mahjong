package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

type mcrDecompositionFixture struct {
	ID           string   `json:"id"`
	Source       string   `json:"source"`
	Hand         []string `json:"hand"`
	WinningTile  string   `json:"winning_tile"`
	ExpectedKind string   `json:"expected_kind"`
}

func TestMCRDecomposeGoldenFixtures(t *testing.T) {
	for _, fixture := range loadMCRDecompositionFixtures(t) {
		t.Run(fixture.ID, func(t *testing.T) {
			hand := parseFixtureTiles(t, fixture.Hand)
			winning := mustTile(t, fixture.WinningTile)

			got := MCRDecompose(hand, nil, winning)

			if !hasMCRDecompositionKind(got, MCRShapeKind(fixture.ExpectedKind)) {
				t.Fatalf("kinds = %v, want %s", mcrDecompositionKinds(got), fixture.ExpectedKind)
			}
			assertMCRDecompositionsUseTiles(t, got, append(hand, winning))
			assertUniqueMCRDecompositions(t, got)
		})
	}
}

func TestMCRDecomposeEnumeratesAmbiguousStandardHand(t *testing.T) {
	hand := mustTiles(t, "1m", "1m", "1m", "2m", "2m", "2m", "3m", "3m", "3m", "4m", "4m", "4m", "5m")

	got := MCRDecompose(hand, nil, mustTile(t, "5m"))

	standard := 0
	for _, decomposition := range got {
		if decomposition.Kind == MCRShapeStandard {
			standard++
		}
	}
	if standard < 2 {
		t.Fatalf("standard decompositions = %d, want at least 2: %#v", standard, got)
	}
}

func TestMCRDecomposeIsPermutationInvariant(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "2p", "3p", "4p", "7s", "7s", "7s", "E")
	reversed := append([]Tile(nil), hand...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}

	first := MCRDecompose(hand, nil, mustTile(t, "E"))
	second := MCRDecompose(reversed, nil, mustTile(t, "E"))

	if mcrDecompositionKeys(first) != mcrDecompositionKeys(second) {
		t.Fatalf("decompositions changed with tile order:\n%s\n%s", mcrDecompositionKeys(first), mcrDecompositionKeys(second))
	}
}

func loadMCRDecompositionFixtures(t *testing.T) []mcrDecompositionFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/rules/mcr/decompositions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []mcrDecompositionFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func hasMCRDecompositionKind(values []MCRDecomposition, kind MCRShapeKind) bool {
	for _, value := range values {
		if value.Kind == kind {
			return true
		}
	}
	return false
}

func mcrDecompositionKinds(values []MCRDecomposition) []MCRShapeKind {
	kinds := make([]MCRShapeKind, len(values))
	for i, value := range values {
		kinds[i] = value.Kind
	}
	return kinds
}

func assertMCRDecompositionsUseTiles(t *testing.T, values []MCRDecomposition, expected []Tile) {
	t.Helper()
	want := append([]Tile(nil), expected...)
	SortTiles(want)
	for _, value := range values {
		got := append([]Tile(nil), value.Tiles...)
		SortTiles(got)
		if FormatTiles(got) != FormatTiles(want) {
			t.Fatalf("%s tiles = %s, want %s", value.Kind, FormatTiles(got), FormatTiles(want))
		}
	}
}

func assertUniqueMCRDecompositions(t *testing.T, values []MCRDecomposition) {
	t.Helper()
	seen := make(map[string]bool)
	for _, value := range values {
		key := mcrDecompositionKey(value)
		if seen[key] {
			t.Fatalf("duplicate decomposition %s", key)
		}
		seen[key] = true
	}
}

func mcrDecompositionKeys(values []MCRDecomposition) string {
	keys := make([]string, len(values))
	for i, value := range values {
		keys[i] = mcrDecompositionKey(value)
	}
	sort.Strings(keys)
	return strings.Join(keys, "\n")
}

func mcrDecompositionKey(value MCRDecomposition) string {
	groups := make([]string, len(value.Groups))
	for i, group := range value.Groups {
		groups[i] = fmt.Sprintf("%s:%s", group.Kind, FormatTiles(group.Tiles))
	}
	sort.Strings(groups)
	return fmt.Sprintf("%s|%s", value.Kind, strings.Join(groups, ";"))
}
