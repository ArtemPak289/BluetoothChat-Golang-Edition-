# Bluetooth Chat (Go Edition)

[Русская версия](README.ru.md)

Console-based peer-to-peer chat over **Bluetooth Low Energy (BLE)** only. No internet, encryption, authentication, or GUI.

Built with [Go 1.24+](https://go.dev/) and [tinygo.org/x/bluetooth](https://github.com/tinygo-org/bluetooth) (module path for `github.com/tinygo-org/bluetooth`).

## Features

- **scan** — discover nearby devices advertising the chat service
- **Advertising** — configurable local name (`-name` flag)
- **connect / disconnect / peers** — connection management, multiple peers
- **send** — JSON messages broadcast to all connected peers
- **help / quit** — CLI help and clean exit

### Message format (JSON)

```json
{
  "id": "uuid",
  "sender": "Dante",
  "timestamp": "2026-06-01T12:00:00Z",
  "text": "Hello"
}
```

## Project layout

```
cmd/main.go
internal/bluetooth/   advertiser, scanner, service, connection, hub
internal/chat/        message, handler
internal/ui/          console
```

## Requirements

- Go 1.24 or newer
- Working BLE adapter
- **Linux**: BlueZ 5.x (`bluetoothd` running). You may need CAP permissions or run as root for advertising/GATT.
- **Windows**: Bluetooth enabled
- **macOS**: Central role only (scan, connect, send/receive). The upstream library does not yet expose GATT peripheral/advertising on macOS, so this machine cannot accept inbound BLE connections. Use Linux or Windows on at least one peer, or connect *from* macOS *to* a peer that is advertising.

## Build

```bash
go build -o bluetooth-chat ./cmd
```

## Run

```bash
./bluetooth-chat -name Dante
```

Example session on two Linux machines:

**Peer A**

```
./bluetooth-chat -name Alice
> scan
Scanning for chat peers (10s)...
...
> connect Bob
Connected to Bob as "Bob"
> send Hello from Alice
```

**Peer B**

```
./bluetooth-chat -name Bob
> scan
> connect Alice
> send Hi Alice
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | Scan ~10 seconds for chat peers |
| `connect <peer>` | Connect by advertised name or address |
| `disconnect <peer>` | Close a connection |
| `peers` | Show discovered (last scan) and connected peers |
| `send <message>` | Send text to all connected peers |
| `help` | Show help |
| `quit` | Exit |

## BLE protocol

- Custom 16-bit UUID service `0xFFE0` with RX (`0xFFE1`) and TX (`0xFFE2`) characteristics (NUS-style).
- Payloads are length-prefixed JSON frames (2-byte big-endian length + UTF-8 JSON).
- Outbound writes use chunked **write without response** for compatibility.

## License

MIT (see repository license if present).
