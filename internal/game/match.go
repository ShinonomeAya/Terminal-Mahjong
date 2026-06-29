package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

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

type MatchSnapshot struct {
	Mode                 RuleMode           `json:"mode"`
	RuleConfig           RuleConfig         `json:"rule_config"`
	Points               [4]int             `json:"points"`
	Dealer               int                `json:"dealer"`
	RoundNumber          int                `json:"round_number"`
	Complete             bool               `json:"complete"`
	Abandoned            bool               `json:"abandoned,omitempty"`
	Round                GameSnapshot       `json:"round"`
	LastMCRSettlement    *MCRSettlement     `json:"last_mcr_settlement,omitempty"`
	LastRiichiSettlement *RiichiSettlement  `json:"last_riichi_settlement,omitempty"`
	MCRSettlements       []MCRSettlement    `json:"mcr_settlements,omitempty"`
	RiichiSettlements    []RiichiSettlement `json:"riichi_settlements,omitempty"`
}

func NewMatch(seed int64, rules RuleSet) (*Match, error) {
	if rules == nil {
		return nil, fmt.Errorf("rule set is required")
	}
	round, err := NewGameWithRules(seed, rules)
	if err != nil {
		return nil, err
	}
	match := &Match{
		Mode:        rules.Mode(),
		RuleConfig:  rules.Config(),
		Points:      rules.InitialPoints(),
		RoundNumber: 1,
		Round:       round,
		rules:       rules,
	}
	match.Round.Dealer = match.Dealer
	match.Round.HandNumber = match.RoundNumber
	match.Round.Current = match.Dealer
	match.replay.initial = match.Snapshot()
	match.recordReplayFrame(nil)
	return match, nil
}

func (match *Match) Snapshot() MatchSnapshot {
	return match.snapshotWithRound(match.Round.Snapshot())
}

func (match *Match) SnapshotFor(playerID string) MatchSnapshot {
	return match.snapshotWithRound(match.Round.SnapshotFor(playerID))
}

func (match *Match) ApplyCommand(command GameCommand) CommandResult {
	result := match.Round.ApplyCommand(command)
	if !result.OK {
		return result
	}
	if command.Kind == CommandQuit {
		match.Abandoned = true
		match.recordReplayFrame(&result.Command)
		return result
	}
	match.recordReplayFrame(&result.Command)
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

func (match *Match) EnsureCurrentTurnDraw() (Tile, bool) {
	tile, ok := match.Round.EnsureCurrentTurnDraw()
	if ok || match.Round.Over {
		match.recordReplayFrame(nil)
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
	return tile, ok
}

func (match *Match) snapshotWithRound(round GameSnapshot) MatchSnapshot {
	return MatchSnapshot{
		Mode:                 match.Mode,
		RuleConfig:           match.RuleConfig,
		Points:               match.Points,
		Dealer:               match.Dealer,
		RoundNumber:          match.RoundNumber,
		Complete:             match.Complete,
		Abandoned:            match.Abandoned,
		Round:                round,
		LastMCRSettlement:    copyMCRSettlement(match.LastMCRSettlement),
		LastRiichiSettlement: copyRiichiSettlement(match.LastRiichiSettlement),
		MCRSettlements:       copyMCRSettlements(match.MCRSettlements),
		RiichiSettlements:    copyRiichiSettlements(match.RiichiSettlements),
	}
}

type matchReplayJournal struct {
	initial  MatchSnapshot
	commands []GameCommand
	frames   []ReplayFrame
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

func (match *Match) ReplayFrameCount() int {
	return len(match.replay.frames)
}

func (match *Match) ReplayCommandCount() int {
	return len(match.replay.commands)
}

func (match *Match) CompletedReplay(applicationVersion string, createdAt time.Time, participants []ReplayParticipant) (ReplayFile, error) {
	if !match.Complete || match.Abandoned {
		return ReplayFile{}, ErrIncompleteReplay
	}
	createdAt = createdAt.UTC()
	file := ReplayFile{
		ApplicationVersion: applicationVersion,
		ReplayID:           matchReplayID(match.Mode, match.replay.initial.Round.ShuffleProof.WallHash, createdAt),
		CreatedAt:          createdAt,
		Mode:               match.Mode,
		RuleConfig:         match.RuleConfig,
		ShuffleProof:       match.replay.initial.Round.ShuffleProof,
		Participants:       append([]ReplayParticipant(nil), participants...),
		Initial:            match.replay.initial,
		Commands:           append([]GameCommand(nil), match.replay.commands...),
		Frames:             append([]ReplayFrame(nil), match.replay.frames...),
		MCRSettlements:     copyMCRSettlements(match.MCRSettlements),
		RiichiSettlements:  copyRiichiSettlements(match.RiichiSettlements),
		FinalStandings:     match.Points,
		Complete:           true,
	}
	copyFile, err := cloneReplayFile(file)
	if err != nil {
		return ReplayFile{}, err
	}
	return SealReplay(copyFile)
}

func matchReplayID(mode RuleMode, wallHash string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", mode, wallHash, createdAt.Format(time.RFC3339Nano))))
	return hex.EncodeToString(sum[:8])
}

func cloneReplayFile(file ReplayFile) (ReplayFile, error) {
	data, err := json.Marshal(file)
	if err != nil {
		return ReplayFile{}, err
	}
	var copyFile ReplayFile
	if err := json.Unmarshal(data, &copyFile); err != nil {
		return ReplayFile{}, err
	}
	return copyFile, nil
}

func (match *Match) completeMCRRound() {
	if match.Complete || match.Mode != ModeMCR || match.Round == nil || !match.Round.Over {
		return
	}
	if match.Round.Winner >= 0 && match.Round.MCRScore != nil {
		settlement := SettleMCR(*match.Round.MCRScore, match.Round.Winner, match.Round.Discarder, match.Round.WinType)
		match.LastMCRSettlement = &settlement
		match.MCRSettlements = append(match.MCRSettlements, *copyMCRSettlement(&settlement))
		for player, delta := range settlement.Deltas {
			match.Points[player] += delta
		}
	} else {
		match.LastMCRSettlement = nil
	}

	match.Dealer = (match.Dealer + 1) % 4
	if match.RoundNumber >= 16 {
		match.Complete = true
		return
	}
	match.RoundNumber++
	next, err := NewGameWithRules(match.Round.Seed+1, match.rules)
	if err != nil {
		match.Complete = true
		return
	}
	next.Dealer = match.Dealer
	next.HandNumber = match.RoundNumber
	next.Current = match.Dealer
	match.Round = next
}

func (match *Match) completeRiichiRound() {
	if match.Complete || match.Mode != ModeRiichi || match.Round == nil || !match.Round.Over {
		return
	}
	dealerRepeats := false
	if match.Round.Winner >= 0 && match.Round.RiichiScore != nil {
		settlement := SettleRiichi(RiichiSettlementInput{
			Winners:      []int{match.Round.Winner},
			Discarder:    match.Round.Discarder,
			Dealer:       match.Dealer,
			WinType:      match.Round.WinType,
			Scores:       []RiichiScoreBreakdown{*match.Round.RiichiScore},
			Honba:        riichiHonba(match.Round),
			RiichiSticks: riichiSticks(match.Round),
		})
		match.LastRiichiSettlement = &settlement
		match.RiichiSettlements = append(match.RiichiSettlements, *copyRiichiSettlement(&settlement))
		for player, delta := range settlement.Deltas {
			match.Points[player] += delta
		}
		dealerRepeats = match.Round.Winner == match.Dealer
	} else {
		match.LastRiichiSettlement = nil
	}

	if !dealerRepeats {
		match.Dealer = (match.Dealer + 1) % 4
		match.RoundNumber++
	}
	if match.RoundNumber > 8 {
		match.Complete = true
		return
	}
	next, err := NewGameWithRules(match.Round.Seed+1, match.rules)
	if err != nil {
		match.Complete = true
		return
	}
	next.Dealer = match.Dealer
	next.HandNumber = match.RoundNumber
	next.Current = match.Dealer
	if next.Riichi != nil && match.LastRiichiSettlement != nil {
		next.Riichi.Honba = match.LastRiichiSettlement.HonbaAfter
		next.Riichi.RiichiSticks = match.LastRiichiSettlement.SticksAfter
	}
	match.Round = next
}

func copyMCRSettlement(settlement *MCRSettlement) *MCRSettlement {
	if settlement == nil {
		return nil
	}
	copyValue := *settlement
	copyValue.Score = *copyMCRScore(&settlement.Score)
	return &copyValue
}

func copyMCRSettlements(settlements []MCRSettlement) []MCRSettlement {
	if len(settlements) == 0 {
		return nil
	}
	result := make([]MCRSettlement, len(settlements))
	for index := range settlements {
		result[index] = *copyMCRSettlement(&settlements[index])
	}
	return result
}

func copyRiichiSettlement(settlement *RiichiSettlement) *RiichiSettlement {
	if settlement == nil {
		return nil
	}
	copyValue := *settlement
	copyValue.Winners = append([]int(nil), settlement.Winners...)
	copyValue.Scores = append([]RiichiScoreBreakdown(nil), settlement.Scores...)
	return &copyValue
}

func copyRiichiSettlements(settlements []RiichiSettlement) []RiichiSettlement {
	if len(settlements) == 0 {
		return nil
	}
	result := make([]RiichiSettlement, len(settlements))
	for index := range settlements {
		result[index] = *copyRiichiSettlement(&settlements[index])
	}
	return result
}

func riichiHonba(round *Game) int {
	if round != nil && round.Riichi != nil {
		return round.Riichi.Honba
	}
	return 0
}

func riichiSticks(round *Game) int {
	if round != nil && round.Riichi != nil {
		return round.Riichi.RiichiSticks
	}
	return 0
}
