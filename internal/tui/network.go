package tui

import "fmt"

type NetworkStatus string

const (
	NetworkLocal        NetworkStatus = ""
	NetworkOffline      NetworkStatus = "offline"
	NetworkReconnecting NetworkStatus = "reconnecting"
	NetworkReconnected  NetworkStatus = "reconnected"
	NetworkWaiting      NetworkStatus = "waiting"
	NetworkYourTurn     NetworkStatus = "your_turn"
)

func renderNetworkStatus(m Model) string {
	switch m.NetworkStatus {
	case NetworkOffline:
		return "Network: offline"
	case NetworkReconnecting:
		return fmt.Sprintf("Network: reconnecting %d/%d", m.ReconnectAttempt, m.ReconnectMax)
	case NetworkReconnected:
		return "Network: reconnected"
	case NetworkWaiting:
		return "Network: waiting for players"
	case NetworkYourTurn:
		return "Network: your turn"
	default:
		return "Network: local"
	}
}
