package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/artempak/bluetooth-chat/internal/bluetooth"
	"github.com/artempak/bluetooth-chat/internal/chat"
	"github.com/artempak/bluetooth-chat/internal/ui"
)

func main() {
	name := flag.String("name", "Anonymous", "Display name advertised and used as message sender")
	flag.Parse()

	handler := chat.NewHandler(*name)

	var hub *bluetooth.Hub
	var err error
	hub, err = bluetooth.NewHub(*name, func(peerKey string, payload []byte) {
		if err := handler.HandlePayload(peerKey, payload); err != nil {
			fmt.Fprintf(os.Stderr, "invalid message from %s: %v\n", peerKey, err)
		}
	}, func(peerKey string, connected bool) {
		if connected {
			handler.RegisterPeer(peerKey, chat.HubTransport{
				SendFn: func(data []byte) error {
					return hub.SendTo(peerKey, data)
				},
			})
		} else {
			handler.UnregisterPeer(peerKey)
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start BLE: %v\n", err)
		os.Exit(1)
	}

	if err := hub.StartAdvertising(); err != nil {
		fmt.Fprintf(os.Stderr, "note: advertising unavailable (%v). Outbound connect still works.\n", err)
	}

	console := ui.NewConsole(*name)
	console.PrintStartupBanner()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var promptMu sync.Mutex
	refreshPrompt := func() {
		promptMu.Lock()
		defer promptMu.Unlock()
		console.PromptReady()
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-handler.Incoming():
				promptMu.Lock()
				console.PrintMessage(msg.Peer, msg.Message.Sender, msg.Message.Text, msg.Message.Timestamp)
				promptMu.Unlock()
				refreshPrompt()
			}
		}
	}()

	go func() {
		<-ctx.Done()
		_ = hub.StopScan()
		os.Exit(0)
	}()

	for {
		cmd, err := console.ReadCommand()
		if err != nil {
			fmt.Fprintf(os.Stderr, "read command: %v\n", err)
			return
		}
		if cmd.Name == "" {
			continue
		}

		switch cmd.Name {
		case "help":
			console.PrintHelp()
		case "quit", "exit":
			console.PrintInfo("Goodbye.")
			return
		case "scan":
			if err := hub.StartScan(); err != nil {
				console.PrintError(err)
				break
			}
			console.PrintInfo("Scanning for chat peers (10s)...")
			go func() {
				time.Sleep(10 * time.Second)
				_ = hub.StopScan()
				peers := hub.DiscoveredPeers()
				promptMu.Lock()
				list := make([]struct {
					Name    string
					Address string
					RSSI    int16
				}, len(peers))
				for i, p := range peers {
					list[i].Name = p.Name
					list[i].Address = p.Address
					list[i].RSSI = p.RSSI
				}
				console.PrintDiscovered(list)
				promptMu.Unlock()
				refreshPrompt()
			}()
		case "connect":
			if len(cmd.Args) < 1 {
				console.PrintError(fmt.Errorf("usage: connect <peer>"))
				break
			}
			peer := strings.Join(cmd.Args, " ")
			key, err := hub.Connect(peer)
			if err != nil {
				console.PrintError(err)
				break
			}
			handler.RegisterPeer(key, chat.HubTransport{
				SendFn: func(data []byte) error {
					return hub.SendTo(key, data)
				},
			})
			console.PrintInfo(fmt.Sprintf("Connected to %s as %q", peer, key))
		case "disconnect":
			if len(cmd.Args) < 1 {
				console.PrintError(fmt.Errorf("usage: disconnect <peer>"))
				break
			}
			peer := strings.Join(cmd.Args, " ")
			if err := hub.Disconnect(peer); err != nil {
				console.PrintError(err)
				break
			}
			handler.UnregisterPeer(peer)
			console.PrintInfo(fmt.Sprintf("Disconnected from %s", peer))
		case "peers":
			discovered := hub.DiscoveredPeers()
			if len(discovered) > 0 {
				console.PrintInfo("Discovered (last scan):")
				list := make([]struct {
					Name    string
					Address string
					RSSI    int16
				}, len(discovered))
				for i, p := range discovered {
					list[i].Name = p.Name
					list[i].Address = p.Address
					list[i].RSSI = p.RSSI
				}
				console.PrintDiscovered(list)
			}
			console.PrintPeers(handler.ConnectedPeers())
		case "send":
			if len(cmd.Args) < 1 {
				console.PrintError(fmt.Errorf("usage: send <message>"))
				break
			}
			text := strings.Join(cmd.Args, " ")
			if err := handler.BroadcastText(text); err != nil {
				console.PrintError(err)
			} else {
				console.PrintInfo("Message sent.")
			}
		default:
			console.PrintError(fmt.Errorf("unknown command %q (try help)", cmd.Name))
		}
	}
}
