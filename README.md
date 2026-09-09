sway-autotile

A lightweight automatic tiling script for Sway written in Go. It dynamically switches between horizontal and vertical splits depending on the active window dimensions.
Features

-    Fast and low memory footprint (written in Go).

-   Dynamically sets splitv or splith based on the focused window's aspect ratio.

-  Seamless Sway IPC integration.

Installation
Prerequisites

-    Go (version 1.20 or newer)

-    Sway

Build from Source

Clone the repository and build the binary:

```
git clone https://github.com/username/sway-autotile.git
cd sway-autotile
go build -ldflags="-s -w" -o sway-autotile
```

Move the compiled binary to your $PATH (e.g., ~/.local/bin or /usr/local/bin):

`install -Dm755 sway-autotile ~/.local/bin/sway-autotile`


Ensure ~/.local/bin is in your $PATH.


Configuration

To automatically start the daemon when Sway launches, add the following line to your Sway config (~/.config/sway/config):

`exec_always --no-startup-id sway-autotile`

Reload Sway to apply the changes:

`swaymsg reload`
