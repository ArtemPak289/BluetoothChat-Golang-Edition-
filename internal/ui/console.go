package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Command is a parsed interactive CLI instruction.
type Command struct {
	Name string
	Args []string
}

// Console provides a line-oriented terminal UI for chat commands.
type Console struct {
	in   *bufio.Reader
	out  io.Writer
	name string
}

// NewConsole creates a terminal UI using the given display name.
func NewConsole(name string) *Console {
	return &Console{
		in:   bufio.NewReader(os.Stdin),
		out:  os.Stdout,
		name: name,
	}
}

// PrintStartupBanner shows welcome text and command hints.
func (c *Console) PrintStartupBanner() {
	fmt.Fprintf(c.out, "Bluetooth Chat — %s\n", c.name)
	fmt.Fprintln(c.out, "Type help for commands.")
	c.printPrompt()
}

// printPrompt writes the input prompt.
func (c *Console) printPrompt() {
	fmt.Fprint(c.out, "> ")
}

// ReadCommand blocks until the user submits a line and parses it into a command.
func (c *Console) ReadCommand() (Command, error) {
	line, err := c.in.ReadString('\n')
	if err != nil {
		return Command{}, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{Name: ""}, nil
	}
	fields := strings.Fields(line)
	cmd := Command{Name: strings.ToLower(fields[0])}
	if len(fields) > 1 {
		cmd.Args = fields[1:]
	}
	return cmd, nil
}

// PrintHelp lists supported commands.
func (c *Console) PrintHelp() {
	fmt.Fprintln(c.out, "Commands:")
	fmt.Fprintln(c.out, "  scan                     Scan for nearby chat peers")
	fmt.Fprintln(c.out, "  connect <peer>           Connect by name or address")
	fmt.Fprintln(c.out, "  disconnect <peer>        Disconnect from a peer")
	fmt.Fprintln(c.out, "  peers                    List connected peers")
	fmt.Fprintln(c.out, "  send <message>           Broadcast to all connected peers")
	fmt.Fprintln(c.out, "  help                     Show this help")
	fmt.Fprintln(c.out, "  quit                     Exit")
}

// PrintDiscovered lists devices from a scan.
func (c *Console) PrintDiscovered(peers []struct {
	Name    string
	Address string
	RSSI    int16
}) {
	if len(peers) == 0 {
		fmt.Fprintln(c.out, "No chat peers discovered yet.")
		return
	}
	fmt.Fprintln(c.out, "Discovered peers:")
	for _, p := range peers {
		fmt.Fprintf(c.out, "  %s  (%s)  RSSI %d\n", p.Name, p.Address, p.RSSI)
	}
}

// PrintPeers shows connected peer keys.
func (c *Console) PrintPeers(peers []string) {
	if len(peers) == 0 {
		fmt.Fprintln(c.out, "No connected peers.")
		return
	}
	fmt.Fprintln(c.out, "Connected peers:")
	for _, p := range peers {
		fmt.Fprintf(c.out, "  %s\n", p)
	}
}

// PrintMessage formats an inbound chat line for the terminal.
func (c *Console) PrintMessage(peer, sender, text, timestamp string) {
	fmt.Fprintf(c.out, "[%s] %s (%s): %s\n", timestamp, sender, peer, text)
}

// PrintInfo writes an informational line.
func (c *Console) PrintInfo(msg string) {
	fmt.Fprintln(c.out, msg)
}

// PrintError writes an error line.
func (c *Console) PrintError(err error) {
	fmt.Fprintf(c.out, "error: %v\n", err)
}

// PromptReady reprints the input prompt after background output.
func (c *Console) PromptReady() {
	c.printPrompt()
}
