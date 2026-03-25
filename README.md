# Hue Control & MCP Server (Go)

A robust tool for Philips Hue, written in Go. It operates in two modes:
1. **CLI Mode:** Directly control your lights from the command line.
2. **MCP Mode:** A Model Context Protocol (MCP) server for AI assistants (like Claude) to interact with your Hue system.

## Features

- **Modern API:** Built on the Philips Hue V2 API for improved reliability and performance.
- **Auto-Discovery:** Automatic detection of Hue Bridges on your local network using mDNS.
- **Robust Authentication:** Simple "link button" pairing flow with persistent configuration.
- **Dual Mode:** CLI for humans, MCP for AI.
- **Type-Safe:** Leverages Go's strong typing and the official MCP Go SDK.

## Installation

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- A Philips Hue Bridge on your local network.

### Build

```bash
cd go
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

# List motion sensors
./hue-control sensors motion

# List temperature sensors
./hue-control sensors temp
./hue-control lights off [light-id]
```

### MCP Mode

To use this with an MCP client like Claude Desktop, add the following to your configuration file (e.g., `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "hue-go": {
      "command": "/Users/andy/DEV/Personal/hue-control/hue-control",
      "args": ["mcp"]
    }
  }
}
```

## Available MCP Tools

- **Lights:** `get_all_lights`, `get_light`, `turn_on_light`, `turn_off_light`, `set_brightness`, `set_color_rgb`, `set_color_temperature`, `alert_light`, `set_light_effect`, `find_light_by_name`, `refresh_lights`, `set_color_preset`.
- **Groups:** `get_all_groups`, `get_all_rooms`, `get_all_zones`, `turn_on_group`, `turn_off_group`, `set_group_brightness`, `set_group_color_rgb`, `set_group_color_preset`.
- **Scenes:** `get_all_scenes`, `set_scene`, `get_motion_sensors`, `get_temperature_sensors`.

## License

MIT - See [LICENSE](LICENSE) for details.
