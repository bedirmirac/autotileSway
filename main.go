package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joshuarubin/go-sway"
)

// crash notifies the desktop environment and exits the program.
func crash(err error) {
	summary := "Sway Autotiling Crashed!"
	body := "The autotiling daemon encountered an unexpected error."
	if err != nil {
		body = fmt.Sprintf("Error: %v", err)
	}

	// 1. Primary: Send desktop notification via D-Bus abstraction (notify-send)
	// Works transparently across Wayland/X11 and all notification daemons (swaync, dunst, mako, etc.)
	if notifyPath, lookupErr := exec.LookPath("notify-send"); lookupErr == nil {
		cmd := exec.Command(
			notifyPath,
			"-u", "critical",
			"-a", "sway-autotiling",
			summary,
			body,
		)
		_ = cmd.Run()
	} else {
		// 2. Fallback: Spawn an available terminal emulator
		terminals := []string{
			os.Getenv("TERMINAL"),
			"foot",
			"ghostty",
			"kitty",
			"alacritty",
			"wezterm",
		}

		shCmd := fmt.Sprintf("echo '[ERROR] %s'; echo '%s'; echo; read -p 'Press Enter to exit...'", summary, body)

		for _, term := range terminals {
			if term == "" {
				continue
			}
			if termPath, lookupErr := exec.LookPath(term); lookupErr == nil {
				_ = exec.Command(termPath, "-e", "sh", "-c", shCmd).Run()
				break
			}
		}
	}

	// 3. Fallback: Always print to stderr for systemd/journald logs
	fmt.Fprintf(os.Stderr, "[CRASH] %s: %s\n", summary, body)
	os.Exit(1)
}

type eventHandler struct {
	sway.EventHandler // dummy handler
	client            sway.Client
}

func (h *eventHandler) Window(ctx context.Context, e sway.WindowEvent) {
	if e.Change != "focus" {
		return
	}

	if e.Container.Type == "floating_con" {
		return
	}

	width := e.Container.Rect.Width
	height := e.Container.Rect.Height

	var cmd string
	if width > height {
		cmd = "splith"
	} else {
		cmd = "splitv"
	}

	h.client.RunCommand(ctx, cmd)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := sway.New(ctx)
	if err != nil {
		crash(err)
	}

	handler := &eventHandler{
		client: client,
	}

	err = sway.Subscribe(ctx, handler, sway.EventTypeWindow)
	if err != nil {
		crash(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
