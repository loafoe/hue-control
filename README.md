# Hue Control for Humans and AIs

A robust tool for Philips Hue, written in Go. It operates in two modes:
1. **CLI Mode:** Directly control your lights from the command line.
2. **MCP Mode:** A Model Context Protocol (MCP) server for AI assistants (like Gemini, Claude, LMStudio) to interact with your Hue system.

## Features

- **Modern API:** Built on the Philips Hue V2 API for improved reliability and performance.
- **Auto-Discovery:** Automatic detection of Hue Bridges on your local network using mDNS.
- **Robust Authentication:** Simple "link button" pairing flow with persistent configuration.
- **Dual Mode:** CLI for humans, MCP for AI.
- **Type-Safe:** Leverages Go's strong typing and the official MCP Go SDK.
- **Containerized:** Multi-arch OCI images (linux/amd64, linux/arm64) on GitHub Container Registry.

## Installation

### Prerequisites

- A Philips Hue Bridge on your local network.

### Container Image (OCI)

Multi-arch images (linux/amd64, linux/arm64) are available on GitHub Container Registry:

```bash
docker pull ghcr.io/loafoe/hue-control:latest
```

For MCP client integration (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "hue-go": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-v", "${HOME}/.hue-control:/root/.hue-control",
        "--network", "host",
        "ghcr.io/loafoe/hue-control:latest",
        "mcp"
      ]
    }
  }
}
```

> **Note:** `--network host` is required for mDNS bridge discovery. The config volume mount persists your bridge credentials across restarts.

### Go Install

```bash
go install github.com/loafoe/hue-control/cmd/hue-control@latest
```

This installs the `hue-control` binary to your `$GOPATH/bin` (or `$HOME/go/bin` by default).

### Build from Source

```bash
git clone https://github.com/loafoe/hue-control.git
cd hue-control
go build -o hue-control cmd/hue-control/main.go
```

## Setup

The first time you run any command, the tool will attempt to discover your Hue Bridge and prompt you to press the link button:

```bash
./hue-control lights list
```

1. The tool will search for the bridge using mDNS.
2. When prompted, press the large round button on top of your Hue Bridge.
3. The configuration will be saved to `~/.hue-control/config.json`.

## Usage

### CLI Mode

```bash
# List all lights
./hue-control lights list

# Turn on a light
./hue-control lights on [light-id]

# Turn off a light
./hue-control lights off [light-id]

# List motion sensors
./hue-control sensors motion

# List temperature sensors
./hue-control sensors temp
```

### MCP Mode

To use this with an MCP client like Claude Desktop, add the following to your configuration file (e.g., `~/Library/Application Support/Claude/claude_desktop_config.json`):

**Native binary:**
```json
{
  "mcpServers": {
    "hue-go": {
      "command": "/path/to/hue-control",
      "args": ["mcp"]
    }
  }
}
```

**Docker container:**
```json
{
  "mcpServers": {
    "hue-go": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-v", "${HOME}/.hue-control:/root/.hue-control",
        "--network", "host",
        "ghcr.io/loafoe/hue-control:latest",
        "mcp"
      ]
    }
  }
}
```

## Available MCP Tools

- **Lights:** `get_all_lights`, `get_light`, `turn_on_light`, `turn_off_light`, `set_brightness`, `set_color_rgb`, `set_color_temperature`, `alert_light`, `set_light_effect`, `find_light_by_name`, `set_color_preset`.
- **Groups:** `get_all_groups`, `get_all_rooms`, `get_all_zones`, `turn_on_group`, `turn_off_group`, `set_group_brightness`, `set_group_color_rgb`, `set_group_color_preset`.
- **Scenes:** `get_all_scenes`, `set_scene`, `get_motion_sensors`, `get_temperature_sensors`.

## License

MIT - See [LICENSE](LICENSE) for details.