package game

import (
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

type mcrCatalogEntry struct {
	ID         string   `json:"id"`
	NameZH     string   `json:"name_zh"`
	NameEN     string   `json:"name_en"`
	Points     int      `json:"points"`
	Excludes   []string `json:"excludes"`
	Repeatable bool     `json:"repeatable"`
	Source     string   `json:"source"`
}

type mcrFanCoverageEntry struct {
	ID       string `json:"id"`
	Positive string `json:"positive"`
	NearMiss string `json:"near_miss"`
}

func TestMCRFanCatalogIsComplete(t *testing.T) {
	catalog := loadMCRCatalog(t, "../../testdata/rules/mcr/catalog.json")
	if len(catalog) != 81 {
		t.Fatalf("fan count = %d, want 81", len(catalog))
	}
	wantBands := map[int]int{88: 7, 64: 6, 48: 2, 32: 3, 24: 9, 16: 6, 12: 5, 8: 10, 6: 6, 4: 4, 2: 10, 1: 13}
	assertMCRPointBands(t, catalog, wantBands)
	assertUniqueFanIDsAndNames(t, catalog)
	assertCatalogExclusionsResolve(t, catalog)
}

func TestEveryMCRFanHasGoldenCoverage(t *testing.T) {
	catalog := loadMCRCatalog(t, "../../testdata/rules/mcr/catalog.json")
	file, err := os.Open("../../testdata/rules/mcr/fans/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var coverage []mcrFanCoverageEntry
	if err := decoder.Decode(&coverage); err != nil {
		t.Fatal(err)
	}
	if len(coverage) != len(catalog) {
		t.Fatalf("coverage entries = %d, want %d", len(coverage), len(catalog))
	}
	covered := make(map[string]bool, len(coverage))
	for _, fixture := range coverage {
		if fixture.ID == "" || fixture.Positive == "" || fixture.NearMiss == "" || covered[fixture.ID] {
			t.Fatalf("invalid fan coverage: %#v", fixture)
		}
		covered[fixture.ID] = true
	}
	for _, fan := range catalog {
		if !covered[fan.ID] {
			t.Errorf("%s has no positive and near-miss coverage", fan.ID)
		}
	}
}

func TestMCRProductionCatalogMatchesGoldenCatalog(t *testing.T) {
	catalog := loadMCRCatalog(t, "../../testdata/rules/mcr/catalog.json")
	if len(mcrFanMetadataByID) != len(catalog) {
		t.Fatalf("production catalog = %d, want %d", len(mcrFanMetadataByID), len(catalog))
	}
	for _, fan := range catalog {
		metadata, ok := mcrFanMetadataByID[FanID(fan.ID)]
		if !ok {
			t.Fatalf("production catalog is missing %s", fan.ID)
		}
		excludes := append([]FanID(nil), metadata.excludes...)
		sort.Slice(excludes, func(i, j int) bool { return excludes[i] < excludes[j] })
		wantExcludes := append([]string(nil), fan.Excludes...)
		sort.Strings(wantExcludes)
		gotExcludes := make([]string, len(excludes))
		for index, id := range excludes {
			gotExcludes[index] = string(id)
		}
		if metadata.zh != fan.NameZH || metadata.en != fan.NameEN || metadata.points != fan.Points || metadata.repeatable != fan.Repeatable || strings.Join(gotExcludes, ",") != strings.Join(wantExcludes, ",") {
			t.Fatalf("production metadata for %s does not match catalog: %#v", fan.ID, metadata)
		}
	}
}

func loadMCRCatalog(t *testing.T, path string) []mcrCatalogEntry {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var catalog []mcrCatalogEntry
	if err := decoder.Decode(&catalog); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("catalog has trailing JSON: %v", err)
	}
	return catalog
}

func assertMCRPointBands(t *testing.T, catalog []mcrCatalogEntry, want map[int]int) {
	t.Helper()
	got := make(map[int]int)
	for _, fan := range catalog {
		got[fan.Points]++
	}
	if len(got) != len(want) {
		t.Fatalf("point bands = %#v, want %#v", got, want)
	}
	for points, count := range want {
		if got[points] != count {
			t.Fatalf("%d-point fans = %d, want %d", points, got[points], count)
		}
	}
}

func assertUniqueFanIDsAndNames(t *testing.T, catalog []mcrCatalogEntry) {
	t.Helper()
	ids := make(map[string]bool)
	zh := make(map[string]bool)
	en := make(map[string]bool)
	for _, fan := range catalog {
		if fan.ID == "" || fan.NameZH == "" || fan.NameEN == "" || fan.Source == "" {
			t.Fatalf("fan has empty required field: %#v", fan)
		}
		if ids[fan.ID] || zh[fan.NameZH] || en[fan.NameEN] {
			t.Fatalf("duplicate fan identity: %#v", fan)
		}
		ids[fan.ID] = true
		zh[fan.NameZH] = true
		en[fan.NameEN] = true
	}
}

func assertCatalogExclusionsResolve(t *testing.T, catalog []mcrCatalogEntry) {
	t.Helper()
	ids := make(map[string]bool)
	for _, fan := range catalog {
		ids[fan.ID] = true
	}
	for _, fan := range catalog {
		for _, excluded := range fan.Excludes {
			if excluded == fan.ID || !ids[excluded] {
				t.Fatalf("%s has invalid exclusion %q", fan.ID, excluded)
			}
		}
	}
}
