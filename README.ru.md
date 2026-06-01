# Bluetooth Chat (Go Edition)

[English version](README.md)

Консольный одноранговый чат только через **Bluetooth Low Energy (BLE)**. Без интернета, шифрования, аутентификации и графического интерфейса.

Собран на [Go 1.24+](https://go.dev/) и [tinygo.org/x/bluetooth](https://github.com/tinygo-org/bluetooth) (модуль `github.com/tinygo-org/bluetooth`).

## Возможности

- **scan** — поиск устройств поблизости, которые рекламируют сервис чата
- **Реклама** — настраиваемое локальное имя (флаг `-name`)
- **connect / disconnect / peers** — управление соединениями, несколько пиров
- **send** — JSON-сообщения, рассылаемые всем подключённым пирам
- **help / quit** — справка в CLI и корректный выход

### Формат сообщения (JSON)

```json
{
  "id": "uuid",
  "sender": "Dante",
  "timestamp": "2026-06-01T12:00:00Z",
  "text": "Hello"
}
```

## Структура проекта

```
cmd/main.go
internal/bluetooth/   advertiser, scanner, service, connection, hub
internal/chat/        message, handler
internal/ui/          console
```

## Требования

- Go 1.24 или новее
- Рабочий BLE-адаптер
- **Linux**: BlueZ 5.x (должен работать `bluetoothd`). Для рекламы и GATT могут понадобиться CAP-права или запуск от root.
- **Windows**: включённый Bluetooth
- **macOS**: только роль центрального устройства (сканирование, подключение, отправка/приём). В используемой библиотеке пока нет GATT peripheral/рекламы на macOS, поэтому этот компьютер не может принимать входящие BLE-соединения. Хотя бы один пир должен быть на Linux или Windows с рекламой, либо подключайтесь *с* macOS *к* пиру, который рекламируется.

## Сборка

```bash
go build -o bluetooth-chat ./cmd
```

## Запуск

```bash
./bluetooth-chat -name Dante
```

Пример сессии на двух машинах с Linux:

**Пир A**

```
./bluetooth-chat -name Alice
> scan
Scanning for chat peers (10s)...
...
> connect Bob
Connected to Bob as "Bob"
> send Hello from Alice
```

**Пир B**

```
./bluetooth-chat -name Bob
> scan
> connect Alice
> send Hi Alice
```

## Команды

| Команда | Описание |
|---------|----------|
| `scan` | Сканирование пиров чата ~10 секунд |
| `connect <peer>` | Подключение по рекламируемому имени или адресу |
| `disconnect <peer>` | Закрыть соединение |
| `peers` | Показать обнаруженных (последнее сканирование) и подключённых пиров |
| `send <message>` | Отправить текст всем подключённым пирам |
| `help` | Справка |
| `quit` | Выход |

## BLE-протокол

- Пользовательский 16-битный UUID-сервис `0xFFE0` с характеристиками RX (`0xFFE1`) и TX (`0xFFE2`) (в стиле NUS).
- Полезная нагрузка — JSON-кадры с префиксом длины (2 байта big-endian + UTF-8 JSON).
- Исходящие записи разбиваются на чанки с **write without response** для совместимости.

## Лицензия

MIT (см. лицензию в репозитории, если она указана).
