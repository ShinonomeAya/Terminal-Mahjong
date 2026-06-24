package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiCatalogEntry struct {
	ID             string `json:"id"`
	NameZH         string `json:"name_zh"`
	NameEN         string `json:"name_en"`
	ClosedHan      int    `json:"closed_han"`
	OpenHan        int    `json:"open_han"`
	Yakuman        int    `json:"yakuman"`
	Limit          string `json:"limit"`
	IsYaku         bool   `json:"is_yaku"`
	RequiresClosed bool   `json:"requires_closed"`
	Repeatable     bool   `json:"repeatable"`
	Source         string `json:"source"`
}

func TestRiichiCatalogMatchesSourceNotes(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var catalog []riichiCatalogEntry
	if err := decoder.Decode(&catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 45 {
		t.Fatalf("catalog entries = %d, want 45", len(catalog))
	}
	seen := make(map[string]bool, len(catalog))
	yaku, yakuman, bonuses := 0, 0, 0
	for _, entry := range catalog {
		if entry.ID == "" || entry.NameZH == "" || entry.NameEN == "" || entry.Source == "" || seen[entry.ID] {
			t.Fatalf("invalid catalog entry: %#v", entry)
		}
		seen[entry.ID] = true
		if entry.IsYaku {
			yaku++
		} else {
			bonuses++
		}
		if entry.Yakuman > 0 {
			yakuman++
		}
		if !entry.IsYaku && (entry.ClosedHan != 0 || entry.OpenHan != 0 || entry.Yakuman != 0 || entry.Limit != "") {
			t.Fatalf("bonus entry is marked as yaku value: %#v", entry)
		}
	}
	if yaku != 41 || yakuman != 12 || bonuses != 4 {
		t.Fatalf("catalog counts yaku=%d yakuman=%d bonuses=%d", yaku, yakuman, bonuses)
	}
}
