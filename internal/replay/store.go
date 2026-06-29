package replay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"
	"unicode"

	"mahjong/internal/game"
)

const maxReplayFileSize = 32 << 20

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

func ApplicationVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}

func Save(dir string, file game.ReplayFile) (string, error) {
	if err := game.ValidateReplay(file); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')

	temp, err := os.CreateTemp(dir, ".replay-*.tmp")
	if err != nil {
		return "", err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}

	name := fmt.Sprintf(
		"%s-%s-%s.json",
		file.CreatedAt.UTC().Format("20060102T150405.000000000Z"),
		safeFilenamePart(string(file.Mode)),
		safeFilenamePart(file.ReplayID),
	)
	path := filepath.Join(dir, name)
	if err := os.Rename(tempPath, path); err != nil {
		return "", err
	}
	return path, nil
}

func Load(path string) (game.ReplayFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return game.ReplayFile{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxReplayFileSize+1))
	if err != nil {
		return game.ReplayFile{}, err
	}
	if len(data) > maxReplayFileSize {
		return game.ReplayFile{}, fmt.Errorf("replay file exceeds %d bytes", maxReplayFileSize)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var replay game.ReplayFile
	if err := decoder.Decode(&replay); err != nil {
		return game.ReplayFile{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return game.ReplayFile{}, fmt.Errorf("replay file contains trailing JSON")
		}
		return game.ReplayFile{}, err
	}
	if err := game.ValidateReplay(replay); err != nil {
		return game.ReplayFile{}, err
	}
	return replay, nil
}

func List(dir string) ([]Entry, []FileIssue, error) {
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	entries := make([]Entry, 0, len(files))
	issues := make([]FileIssue, 0)
	for _, file := range files {
		if file.IsDir() || !strings.EqualFold(filepath.Ext(file.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, file.Name())
		replay, err := Load(path)
		if err != nil {
			issues = append(issues, FileIssue{Path: path, Err: err})
			continue
		}
		entries = append(entries, Entry{
			Path:      path,
			ReplayID:  replay.ReplayID,
			Mode:      replay.Mode,
			CreatedAt: replay.CreatedAt,
			Frames:    len(replay.Frames),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].CreatedAt.Equal(entries[j].CreatedAt) {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
	return entries, issues, nil
}

func safeFilenamePart(value string) string {
	if value == "" {
		return "unknown"
	}
	var out strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			out.WriteRune(r)
		} else {
			out.WriteByte('_')
		}
	}
	return out.String()
}
