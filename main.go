package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuarubin/go-sway"
)

func crash(err error) {
	summary := "Sway Autotiling Crashed!"
	body := "The autotiling daemon encountered an unexpected error."
	if err != nil {
		body = fmt.Sprintf("Error: %v", err)
	}

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

	fmt.Fprintf(os.Stderr, "[CRASH] %s: %s\n", summary, body)
	os.Exit(1)
}

type eventHandler struct {
	sway.EventHandler
	client sway.Client
}

func (h *eventHandler) updateLayout(ctx context.Context, conID int64) {
	// Sway'in ağacı güncellemesi ve yeni boyutları hesaplaması için mikro bekleme (race condition önleyici)
	time.Sleep(20 * time.Millisecond)

	tree, err := h.client.GetTree(ctx)
	if err != nil {
		return
	}

	// Odaklı pencereyi ağaçtan taze koordinatlarıyla bul
	focused := findFocused(tree)
	if focused == nil {
		return
	}

	// Yalnızca normal uygulama pencerelerini hedefle (floating, workspace veya output değil)
	if focused.Type != "con" || focused.ID == 0 {
		return
	}

	// Genişlik ve yükseklik kontrolü
	width := focused.Rect.Width
	height := focused.Rect.Height
	if width == 0 || height == 0 {
		return
	}

	var cmd string
	if width > height {
		cmd = fmt.Sprintf("[con_id=%d] splith", focused.ID)
	} else {
		cmd = fmt.Sprintf("[con_id=%d] splitv", focused.ID)
	}

	_, _ = h.client.RunCommand(ctx, cmd)
}

func findFocused(node *sway.Node) *sway.Node {
	if node == nil {
		return nil
	}
	if node.Focused {
		return node
	}
	for i := range node.Nodes {
		if f := findFocused(node.Nodes[i]); f != nil {
			return f
		}
	}
	for i := range node.FloatingNodes {
		if f := findFocused(node.FloatingNodes[i]); f != nil {
			return f
		}
	}
	return nil
}

func (h *eventHandler) Window(ctx context.Context, e sway.WindowEvent) {
	// Bir pencere kapandığında ("close") veya yeni pencereye odak geçildiğinde ("focus")
	if e.Change == "focus" || e.Change == "close" {
		go h.updateLayout(ctx, e.Container.ID)
	}
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
