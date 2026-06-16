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
	if m.chinese() {
		switch m.NetworkStatus {
		case NetworkOffline:
			return "网络：离线"
		case NetworkReconnecting:
			return fmt.Sprintf("网络：重连中 %d/%d", m.ReconnectAttempt, m.ReconnectMax)
		case NetworkReconnected:
			return "网络：已重连"
		case NetworkWaiting:
			return "网络：等待玩家"
		case NetworkYourTurn:
			return "网络：轮到你"
		default:
			return "网络：本地"
		}
	}
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
