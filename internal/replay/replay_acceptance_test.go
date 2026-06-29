package replay

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mahjong/internal/game"
)

func TestReplayCorruptionMatrix(t *testing.T) {
	base := replayStoreFixture(t, "matrix", time.Unix(60, 0).UTC())
	validJSON, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	mutated := func(change func(*game.ReplayFile)) []byte {
		file := base
		file.Frames = append([]game.ReplayFrame(nil), base.Frames...)
		change(&file)
		data, marshalErr := json.Marshal(file)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return data
	}
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{name: "empty", data: nil},
		{name: "truncated", data: []byte("{")},
		{name: "trailing", data: append(append([]byte(nil), validJSON...), []byte("\n{}")...)},
		{name: "checksum", data: mutated(func(file *game.ReplayFile) { file.ReplayID = "tampered" }), want: game.ErrReplayChecksum},
		{name: "missing-frames", data: mutated(func(file *game.ReplayFile) { file.Frames = nil }), want: game.ErrInvalidReplay},
		{name: "frame-gap", data: mutated(func(file *game.ReplayFile) { file.Frames[0].Index = 1 }), want: game.ErrInvalidReplay},
		{name: "invalid-mode", data: mutated(func(file *game.ReplayFile) { file.Mode = game.RuleMode("invalid") }), want: game.ErrInvalidReplay},
		{name: "incomplete", data: mutated(func(file *game.ReplayFile) { file.Complete = false }), want: game.ErrIncompleteReplay},
		{name: "schema-1", data: mutated(func(file *game.ReplayFile) { file.SchemaVersion = 1 }), want: game.ErrUnsupportedReplayVersion},
		{name: "schema-3", data: mutated(func(file *game.ReplayFile) { file.SchemaVersion = 3 }), want: game.ErrUnsupportedReplayVersion},
	}

	dir := t.TempDir()
	if _, err := Save(dir, base); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(dir, test.name+".json")
			if err := os.WriteFile(path, test.data, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil {
				t.Fatal("corrupt replay loaded successfully")
			}
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
	entries, issues, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ReplayID != base.ReplayID {
		t.Fatalf("entries=%#v", entries)
	}
	if len(issues) != len(cases) {
		t.Fatalf("issues=%d want=%d", len(issues), len(cases))
	}
}
