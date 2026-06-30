# Terminal Mahjong

[![CI](https://github.com/ShinonomeAya/Terminal-Mahjong/actions/workflows/ci.yml/badge.svg)](https://github.com/ShinonomeAya/Terminal-Mahjong/actions/workflows/ci.yml)
[![Go 1.23](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个使用 Go、Bubble Tea 和 Lip Gloss 编写的终端麻将客户端。它支持 Unicode 麻将牌、键盘和鼠标操作、三种规则模式、单机机器人、WebSocket 联网、断线重连，以及可校验的完整赛后回放。

项目保持终端优先，不依赖图形界面。

## 功能

- **终端牌桌**：四家固定座位、牌河、手牌托盘、当前行动提示和战术分析栏。
- **现代终端操作**：方向键选牌，回车或空格出牌，支持鼠标选择和再次点击出牌。
- **中英文界面**：可在开始菜单切换语言。
- **三种规则模式**：
  - 经典兼容模式；
  - 国标麻将（MCR），包括 144 张牌、花牌、完整番种与结算；
  - 四人日麻，包括立直、振听、宝牌、里宝牌、赤宝牌、役种、符番与东南战结算。
- **单机游戏**：一名玩家与三名启发式机器人对局。
- **联网游戏**：内存房间、准备状态、空座机器人、服务端权威动作、WebSocket 状态同步。
- **断线重连**：客户端保存 reconnect token，服务端默认保留离线会话两分钟。
- **公平发牌**：默认种子来自 `crypto/rand`，牌墙使用确定性洗牌，并记录 SHA-256 牌墙证明。
- **完整回放**：本地和联网完赛后自动保存经过 schema 与 checksum 校验的逐帧回放。

## 运行要求

- Go 1.23 或更高版本；
- 支持 UTF-8、ANSI 颜色和 Unicode 麻将字符的现代终端；
- Windows 推荐使用 Windows Terminal。

麻将字符的大小和形状由终端字体决定。如果字符显示不完整，请选择包含 Mahjong Tiles Unicode 区块的字体或启用终端字体回退。

## 开始单机游戏

克隆并运行：

```powershell
git clone https://github.com/ShinonomeAya/Terminal-Mahjong.git
cd Terminal-Mahjong
go run ./cmd/mahjong
```

也可以构建独立程序：

```powershell
go build -o terminal-mahjong.exe ./cmd/mahjong
.\terminal-mahjong.exe
```

Linux 和 macOS 可以省略 `.exe` 后缀。

## 操作

| 场景 | 按键或操作 |
| --- | --- |
| 菜单 | `Up` / `Down` 选择，`Enter` 确认 |
| 手牌 | `Left` / `Right` 选择牌 |
| 出牌 | `Enter` / `Space` |
| 鼠标 | 单击选择，再次单击同一张牌出牌 |
| 胡牌 | `H` |
| 杠 | `K` |
| 立直 | `L` |
| 响应弃牌 | `H` 胡、`P` 碰、`C` 吃、`Space` 或 `Esc` 过 |
| 多个吃牌方案 | `Left` / `Right` 选择，`C` 确认 |
| 战术栏 | `Tab` |
| 退出当前对局 | `Q` |

界面只会展示服务端或规则引擎给出的合法动作。TUI 不会自行判断联网动作是否合法。

## 启动联网游戏

先启动内存房间服务器：

```powershell
go run ./cmd/server -addr :8080
```

TUI 默认连接 `ws://127.0.0.1:8080/ws`，可以在开始菜单创建房间、浏览房间、加入房间或使用已保存的会话重连。

命令行客户端适合测试或连接其他主机：

```powershell
# 创建房间
go run ./cmd/client -server ws://127.0.0.1:8080/ws -name Alice

# 加入房间
go run ./cmd/client -server ws://127.0.0.1:8080/ws -name Bob -join 000001 -session .mahjong-bob.json

# 准备并持续接收状态
go run ./cmd/client -reconnect -ready -watch

# 使用日麻规则创建房间，并关闭赤宝牌
go run ./cmd/client -mode riichi -red-fives 0

# 查看等待中的房间
go run ./cmd/client -server ws://127.0.0.1:8080/ws -list
```

服务器目前不使用数据库。重启服务会清空房间，未开始房间在默认十分钟空闲后清理。

## 查看回放

完成的本地和联网比赛默认保存在 `replays/`：

```text
<UTC 时间>-<规则模式>-<回放 ID>.json
```

在开始菜单选择 **回放 / Replays**：

| 操作 | 按键 |
| --- | --- |
| 选择回放 | `Up` / `Down` |
| 打开 | `Enter` |
| 刷新列表 | `R` |
| 前后切帧 | `Left` / `Right` |
| 跳到首尾 | `Home` / `End` |
| 播放或暂停 | `Space` |
| 完整手牌与结算 | `Tab` |
| 返回 | `Esc` |

浏览器会跳过损坏、未完成或版本不兼容的文件，并显示跳过数量。`ReplayFile` 保存完整比赛帧；较早的 `ReplayLog` 仅用于简短的事件和结果摘要。

## 公平性与隐私

- 新比赛使用系统加密随机源生成种子；
- 固定种子测试可以复现相同发牌和事件顺序；
- 回放包含 seed 和牌墙 hash，便于审计；
- 实时联网快照只包含接收者自己的暗牌，不暴露其他玩家手牌、seed 或里宝牌；
- 完整信息只在比赛结束后通过封存回放交付；
- 回放不包含 reconnect token、WebSocket 地址、IP 地址或本地会话路径。

## 测试

运行完整测试：

```powershell
go test ./...
```

发布前使用的完整检查：

```powershell
go mod verify
go test ./... -count=1 -shuffle=on
go test -race ./... -count=1
go vet ./...
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

规则一致性资料位于：

- [规则符合性说明](docs/rules/conformance.md)
- [国标番种目录](docs/rules/mcr-fan-catalog.md)
- [日麻规则来源说明](docs/rules/riichi-source-notes.md)
- [联网客户端验收清单](docs/online-client-acceptance.md)

## 项目结构

```text
cmd/mahjong      TUI 客户端
cmd/server       WebSocket 房间服务器
cmd/client       最小命令行联网客户端
internal/game    规则、比赛状态、结算和权威回放
internal/bot     启发式机器人
internal/online  房间、客户端、隐私和重连
internal/replay  回放校验与原子存储
internal/tui     Bubble Tea 界面与输入
testdata/rules   MCR 与日麻规则 fixture
```

## 当前限制

- 房间、玩家和重连会话仅保存在服务端内存中；
- 没有账号、匹配、排行榜或观战系统；
- 公网部署需要自行配置 TLS、反向代理和访问控制；
- 回放 schema 暂不提供自动迁移，未知版本会被安全跳过。

## 许可证

[MIT License](LICENSE)
