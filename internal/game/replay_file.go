package game

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

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
	Index   int           `json:"index"`
	Command *GameCommand  `json:"command,omitempty"`
	Match   MatchSnapshot `json:"match"`
}

type ReplayFile struct {
	SchemaVersion      int                 `json:"schema_version"`
	ApplicationVersion string              `json:"application_version"`
	ReplayID           string              `json:"replay_id"`
	CreatedAt          time.Time           `json:"created_at"`
	Mode               RuleMode            `json:"mode"`
	RuleConfig         RuleConfig          `json:"rule_config"`
	ShuffleProof       ShuffleProof        `json:"shuffle_proof"`
	Participants       []ReplayParticipant `json:"participants"`
	Initial            MatchSnapshot       `json:"initial"`
	Commands           []GameCommand       `json:"commands"`
	Frames             []ReplayFrame       `json:"frames"`
	MCRSettlements     []MCRSettlement     `json:"mcr_settlements,omitempty"`
	RiichiSettlements  []RiichiSettlement  `json:"riichi_settlements,omitempty"`
	FinalStandings     [4]int              `json:"final_standings"`
	Complete           bool                `json:"complete"`
	Checksum           string              `json:"checksum"`
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

func ValidateReplay(file ReplayFile) error {
	if file.SchemaVersion != ReplayFileSchemaVersion {
		return fmt.Errorf("%w: %d", ErrUnsupportedReplayVersion, file.SchemaVersion)
	}
	if !file.Complete {
		return ErrIncompleteReplay
	}
	if file.ApplicationVersion == "" {
		return invalidReplay("application version is required")
	}
	if file.ReplayID == "" {
		return invalidReplay("replay id is required")
	}
	if file.CreatedAt.IsZero() {
		return invalidReplay("creation time is required")
	}
	if err := file.RuleConfig.Validate(file.Mode); err != nil {
		return invalidReplay(err.Error())
	}
	if file.Initial.Mode != file.Mode || file.Initial.RuleConfig != file.RuleConfig {
		return invalidReplay("initial match mode or configuration differs")
	}
	if file.Initial.Round.ShuffleProof != file.ShuffleProof {
		return invalidReplay("shuffle proof differs from initial match")
	}
	if err := validateReplayParticipants(file.Participants); err != nil {
		return err
	}
	if err := validateReplayFrames(file); err != nil {
		return err
	}

	expected, err := replayChecksum(file)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(file.Checksum)) != 1 {
		return ErrReplayChecksum
	}
	return nil
}

func validateReplayParticipants(participants []ReplayParticipant) error {
	if len(participants) != 4 {
		return invalidReplay("exactly four participants are required")
	}
	var seen [4]bool
	for _, participant := range participants {
		if participant.Seat < 0 || participant.Seat >= len(seen) {
			return invalidReplay("participant seat is out of range")
		}
		if seen[participant.Seat] {
			return invalidReplay("participant seats must be unique")
		}
		seen[participant.Seat] = true
		if participant.ID == "" || participant.Name == "" {
			return invalidReplay("participant id and name are required")
		}
	}
	return nil
}

func validateReplayFrames(file ReplayFile) error {
	if len(file.Frames) == 0 {
		return invalidReplay("at least one frame is required")
	}
	commands := make([]GameCommand, 0, len(file.Commands))
	for index, frame := range file.Frames {
		if frame.Index != index {
			return invalidReplay("frame indexes must be contiguous")
		}
		if frame.Match.Mode != file.Mode || frame.Match.RuleConfig != file.RuleConfig {
			return invalidReplay("frame match mode or configuration differs")
		}
		if frame.Command != nil {
			commands = append(commands, *frame.Command)
		}
	}
	if len(commands) != len(file.Commands) {
		return invalidReplay("command list differs from frame commands")
	}
	for index := range commands {
		if commands[index] != file.Commands[index] {
			return invalidReplay("command list differs from frame commands")
		}
	}
	last := file.Frames[len(file.Frames)-1].Match
	if !last.Complete {
		return invalidReplay("last frame is not complete")
	}
	if last.Points != file.FinalStandings {
		return invalidReplay("final standings differ from last frame")
	}
	return nil
}

func replayChecksum(file ReplayFile) (string, error) {
	file.Checksum = ""
	data, err := json.Marshal(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func invalidReplay(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidReplay, message)
}
