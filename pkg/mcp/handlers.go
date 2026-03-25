package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/loafoe/hue-control/internal/color"
	"github.com/loafoe/hue-control/pkg/hue"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openhue/openhue-go"
)

// Helper functions for pointer types
func ptr[T any](v T) *T {
	return &v
}

// RegisterHandlers registers all Hue MCP tools with the server.
func RegisterHandlers(server *mcp.Server, client *hue.Client) {
	// 1. Get All Lights
	type emptyArgs struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_lights",
		Description: "List all lights with their UUIDs, names, and current state. Call this first to discover light IDs needed by other tools.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		lights, err := client.GetLights()
		if err != nil {
			return errorResult(err), nil, nil
		}

		data, _ := json.MarshalIndent(lights, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 2. Get Light
	type getLightArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_light",
		Description: "Get detailed information about a specific light including color, brightness, and capabilities.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getLightArgs) (*mcp.CallToolResult, any, error) {
		lights, err := client.GetLights()
		if err != nil {
			return errorResult(err), nil, nil
		}

		light, ok := lights[args.LightID]
		if !ok {
			return errorResult(fmt.Errorf("light %s not found", args.LightID)), nil, nil
		}

		data, _ := json.MarshalIndent(light, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 3. Turn On Light
	type lightIDArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "turn_on_light",
		Description: "Turn on a specific light by UUID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args lightIDArgs) (*mcp.CallToolResult, any, error) {
		update := openhue.LightPut{
			On: &openhue.On{On: ptr(true)},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s turned on.", args.LightID)), nil, nil
	})

	// 4. Turn Off Light
	mcp.AddTool(server, &mcp.Tool{
		Name:        "turn_off_light",
		Description: "Turn off a specific light by UUID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args lightIDArgs) (*mcp.CallToolResult, any, error) {
		update := openhue.LightPut{
			On: &openhue.On{On: ptr(false)},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s turned off.", args.LightID)), nil, nil
	})

	// 5. Set Brightness
	type setBrightnessArgs struct {
		LightID    string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
		Brightness int    `json:"brightness" jsonschema:"description:Brightness level 0-100 (0=off, 50=half, 100=full)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_brightness",
		Description: "Set light brightness. Also turns the light on if it was off.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setBrightnessArgs) (*mcp.CallToolResult, any, error) {
		if args.Brightness < 0 || args.Brightness > 100 {
			return errorResult(fmt.Errorf("brightness must be between 0 and 100")), nil, nil
		}

		update := openhue.LightPut{
			Dimming: &openhue.Dimming{Brightness: ptr(float32(args.Brightness))},
			On:      &openhue.On{On: ptr(true)}, // Turn on if setting brightness
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s brightness set to %d%%.", args.LightID, args.Brightness)), nil, nil
	})

	// 6. Set Color RGB
	type setColorRGBArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
		Red     int    `json:"red" jsonschema:"description:Red value 0-255"`
		Green   int    `json:"green" jsonschema:"description:Green value 0-255"`
		Blue    int    `json:"blue" jsonschema:"description:Blue value 0-255"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_rgb",
		Description: "Set light to a specific RGB color. For common colors, consider set_color_preset instead. Also turns the light on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setColorRGBArgs) (*mcp.CallToolResult, any, error) {
		xy := color.RGBToXY(args.Red, args.Green, args.Blue)
		update := openhue.LightPut{
			Color: &openhue.Color{
				Xy: &openhue.GamutPosition{
					X: ptr(float32(xy.X)),
					Y: ptr(float32(xy.Y)),
				},
			},
			On: &openhue.On{On: ptr(true)},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s color set to RGB(%d, %d, %d).", args.LightID, args.Red, args.Green, args.Blue)), nil, nil
	})

	// 7. Set Color Temperature
	type setColorTemperatureArgs struct {
		LightID     string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
		Temperature int    `json:"temperature" jsonschema:"description:Color temperature in Kelvin: 2000=warm candlelight, 2700=soft white, 4000=neutral, 5000=cool daylight, 6500=bright blue-white"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_temperature",
		Description: "Set white color temperature. Lower values are warmer/orange, higher values are cooler/blue. Also turns the light on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setColorTemperatureArgs) (*mcp.CallToolResult, any, error) {
		if args.Temperature < 2000 || args.Temperature > 6500 {
			return errorResult(fmt.Errorf("temperature must be between 2000K and 6500K")), nil, nil
		}

		mired := color.KelvinToMired(args.Temperature)
		update := openhue.LightPut{
			ColorTemperature: &openhue.ColorTemperature{
				Mirek: ptr(mired),
			},
			On: &openhue.On{On: ptr(true)},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s color temperature set to %dK (%d mirek).", args.LightID, args.Temperature, *update.ColorTemperature.Mirek)), nil, nil
	})

	// 8. Get All Groups (Rooms and Zones)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_groups",
		Description: "List all rooms AND zones combined with their IDs. Use this when you need to control any group regardless of type. For only rooms or zones, use get_all_rooms or get_all_zones.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		rooms, err := client.GetRooms()
		if err != nil {
			return errorResult(err), nil, nil
		}
		zones, err := client.GetZones()
		if err != nil {
			return errorResult(err), nil, nil
		}

		groups := make(map[string]any)
		for id, room := range rooms {
			groups[id] = room
		}
		for id, zone := range zones {
			groups[id] = zone
		}

		data, _ := json.MarshalIndent(groups, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 9. Get All Rooms
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_rooms",
		Description: "List all rooms (physical spaces like 'Living Room', 'Bedroom') with their IDs. Each light belongs to exactly one room. Use room IDs with group control tools.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		rooms, err := client.GetRooms()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(rooms, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 10. Get All Zones
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_zones",
		Description: "List all zones (cross-room groupings like 'Downstairs', 'All Lights') with their IDs. Zones can span multiple rooms. Use zone IDs with group control tools.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		zones, err := client.GetZones()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(zones, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 11. Turn On Group
	type groupIDArgs struct {
		GroupID string `json:"group_id" jsonschema:"description:The UUID of the room or zone (use get_all_rooms, get_all_zones, or get_all_groups to discover IDs)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "turn_on_group",
		Description: "Turn on all lights in a room or zone at once.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args groupIDArgs) (*mcp.CallToolResult, any, error) {
		groupedLightID, err := client.GetGroupedLightByResourceID(args.GroupID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		update := openhue.GroupedLightPut{
			On: &openhue.On{On: ptr(true)},
		}
		err = client.UpdateGroupedLightState(groupedLightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Group %s turned on.", args.GroupID)), nil, nil
	})

	// 12. Turn Off Group
	mcp.AddTool(server, &mcp.Tool{
		Name:        "turn_off_group",
		Description: "Turn off all lights in a room or zone at once.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args groupIDArgs) (*mcp.CallToolResult, any, error) {
		groupedLightID, err := client.GetGroupedLightByResourceID(args.GroupID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		update := openhue.GroupedLightPut{
			On: &openhue.On{On: ptr(false)},
		}
		err = client.UpdateGroupedLightState(groupedLightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Group %s turned off.", args.GroupID)), nil, nil
	})

	// 13. Set Group Brightness
	type setGroupBrightnessArgs struct {
		GroupID    string `json:"group_id" jsonschema:"description:The UUID of the room or zone (use get_all_rooms, get_all_zones, or get_all_groups to discover IDs)"`
		Brightness int    `json:"brightness" jsonschema:"description:Brightness level 0-100 (0=off, 50=half, 100=full)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_brightness",
		Description: "Set brightness for all lights in a room or zone. Also turns lights on if they were off.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setGroupBrightnessArgs) (*mcp.CallToolResult, any, error) {
		if args.Brightness < 0 || args.Brightness > 100 {
			return errorResult(fmt.Errorf("brightness must be between 0 and 100")), nil, nil
		}

		groupedLightID, err := client.GetGroupedLightByResourceID(args.GroupID)
		if err != nil {
			return errorResult(err), nil, nil
		}

		update := openhue.GroupedLightPut{
			Dimming: &openhue.Dimming{Brightness: ptr(float32(args.Brightness))},
			On:      &openhue.On{On: ptr(true)},
		}
		err = client.UpdateGroupedLightState(groupedLightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Group %s brightness set to %d%%.", args.GroupID, args.Brightness)), nil, nil
	})

	// 14. Set Group Color RGB
	type setGroupColorRGBArgs struct {
		GroupID string `json:"group_id" jsonschema:"description:The UUID of the room or zone (use get_all_rooms, get_all_zones, or get_all_groups to discover IDs)"`
		Red     int    `json:"red" jsonschema:"description:Red value 0-255"`
		Green   int    `json:"green" jsonschema:"description:Green value 0-255"`
		Blue    int    `json:"blue" jsonschema:"description:Blue value 0-255"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_color_rgb",
		Description: "Set all lights in a room or zone to a specific RGB color. For common colors, consider set_group_color_preset instead. Also turns lights on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setGroupColorRGBArgs) (*mcp.CallToolResult, any, error) {
		xy := color.RGBToXY(args.Red, args.Green, args.Blue)
		groupedLightID, err := client.GetGroupedLightByResourceID(args.GroupID)
		if err != nil {
			return errorResult(err), nil, nil
		}

		update := openhue.GroupedLightPut{
			Color: &openhue.Color{
				Xy: &openhue.GamutPosition{
					X: ptr(float32(xy.X)),
					Y: ptr(float32(xy.Y)),
				},
			},
			On: &openhue.On{On: ptr(true)},
		}
		err = client.UpdateGroupedLightState(groupedLightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Group %s color set to RGB(%d, %d, %d).", args.GroupID, args.Red, args.Green, args.Blue)), nil, nil
	})

	// 15. Get All Scenes
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_scenes",
		Description: "List all scenes with their IDs and names. Scenes are pre-configured lighting states (colors, brightness) for groups of lights. Call this to discover scene IDs for set_scene.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		scenes, err := client.GetScenes()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(scenes, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 16. Set Scene
	type setSceneArgs struct {
		SceneID string `json:"scene_id" jsonschema:"description:The UUID of the scene (use get_all_scenes to discover available scenes)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_scene",
		Description: "Activate a pre-configured scene. Scenes apply saved colors and brightness to multiple lights at once. Use get_all_scenes to discover available scenes.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setSceneArgs) (*mcp.CallToolResult, any, error) {
		update := openhue.ScenePut{
			Recall: &openhue.SceneRecall{
				Action: ptr(openhue.SceneRecallActionActive),
			},
		}
		err := client.UpdateSceneState(args.SceneID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Scene %s applied.", args.SceneID)), nil, nil
	})

	// 17. Find Light By Name
	type findByNameArgs struct {
		Name string `json:"name" jsonschema:"description:Partial or full name to search for (case-insensitive)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_light_by_name",
		Description: "Search lights by name (case-insensitive partial match). Faster than get_all_lights when you know the light's name. Returns matching lights with their UUIDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args findByNameArgs) (*mcp.CallToolResult, any, error) {
		lights, err := client.GetLights()
		if err != nil {
			return errorResult(err), nil, nil
		}

		matches := make(map[string]openhue.LightGet)
		searchName := strings.ToLower(args.Name)
		for id, light := range lights {
			if light.Metadata != nil && light.Metadata.Name != nil {
				if strings.Contains(strings.ToLower(*light.Metadata.Name), searchName) {
					matches[id] = light
				}
			}
		}

		data, _ := json.MarshalIndent(matches, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 18. Alert Light
	mcp.AddTool(server, &mcp.Tool{
		Name:        "alert_light",
		Description: "Make a light flash briefly (breathe effect) to help identify which physical light corresponds to an ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args lightIDArgs) (*mcp.CallToolResult, any, error) {
		update := openhue.LightPut{
			Alert: &openhue.Alert{Action: ptr("breathe")},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Light %s alerted.", args.LightID)), nil, nil
	})

	// 19. Set Light Effect
	type setEffectArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
		Effect  string `json:"effect" jsonschema:"description:Effect type: 'colorloop' cycles through colors, 'none' stops effects"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_light_effect",
		Description: "Set a dynamic effect on a light. Use 'colorloop' for continuous color cycling, 'none' to stop. Also turns the light on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setEffectArgs) (*mcp.CallToolResult, any, error) {
		update := openhue.LightPut{
			Effects: &openhue.Effects{Effect: ptr(openhue.SupportedEffects(args.Effect))},
			On:      &openhue.On{On: ptr(true)},
		}
		err := client.UpdateLightState(args.LightID, update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Effect %s set on light %s.", args.Effect, args.LightID)), nil, nil
	})

	// 20. Set Color Preset
	type setPresetArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light (use get_all_lights or find_light_by_name to discover IDs)"`
		Preset  string `json:"preset" jsonschema:"description:Preset name - Activity: concentration (bright cool), relax (dim warm), reading (medium neutral), energize (bright cool). White: warm, cool, daylight. Color: red, green, blue, purple, orange"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_preset",
		Description: "Apply a named color/mood preset to a light. Easier than specifying RGB or temperature values. Also turns the light on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setPresetArgs) (*mcp.CallToolResult, any, error) {
		update, err := getPresetLightPut(args.Preset)
		if err != nil {
			return errorResult(err), nil, nil
		}
		err = client.UpdateLightState(args.LightID, *update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Applied preset %s to light %s.", args.Preset, args.LightID)), nil, nil
	})

	// 21. Set Group Color Preset
	type setGroupPresetArgs struct {
		GroupID string `json:"group_id" jsonschema:"description:The UUID of the room or zone (use get_all_rooms, get_all_zones, or get_all_groups to discover IDs)"`
		Preset  string `json:"preset" jsonschema:"description:Preset name - Activity: concentration (bright cool), relax (dim warm), reading (medium neutral), energize (bright cool). White: warm, cool, daylight. Color: red, green, blue, purple, orange"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_color_preset",
		Description: "Apply a named color/mood preset to all lights in a room or zone. Easier than specifying RGB values. Also turns lights on.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setGroupPresetArgs) (*mcp.CallToolResult, any, error) {
		update, err := getPresetGroupedLightPut(args.Preset)
		if err != nil {
			return errorResult(err), nil, nil
		}
		groupedLightID, err := client.GetGroupedLightByResourceID(args.GroupID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		err = client.UpdateGroupedLightState(groupedLightID, *update)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Applied preset %s to group %s.", args.Preset, args.GroupID)), nil, nil
	})

	// 22. Get Motion Sensors
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_motion_sensors",
		Description: "List all motion sensors with their IDs, names, and current motion detection state.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		sensors, err := client.GetMotionSensors()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(sensors, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 23. Get Temperature Sensors
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_temperature_sensors",
		Description: "List all temperature sensors with their IDs, names, and current temperature readings.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		sensors, err := client.GetTemperatureSensors()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(sensors, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// Register Prompts
	server.AddPrompt(&mcp.Prompt{
		Name:        "control_lights",
		Description: "Basic light control guidance.",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: "You are connected to a Philips Hue lighting system. Help me control my lights.",
					},
				},
			},
		}, nil
	})
}

func getPresetLightPut(preset string) (*openhue.LightPut, error) {
	update := &openhue.LightPut{On: &openhue.On{On: ptr(true)}}
	switch preset {
	case "warm":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(400)} // 2500K
	case "cool":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(222)} // 4500K
	case "daylight":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(153)} // 6500K
	case "concentration":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(217)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(100))}
	case "relax":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(370)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(50))}
	case "reading":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(312)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(80))}
	case "energize":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(166)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(100))}
	case "red":
		xy := color.RGBToXY(255, 0, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "green":
		xy := color.RGBToXY(0, 255, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "blue":
		xy := color.RGBToXY(0, 0, 255)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "purple":
		xy := color.RGBToXY(128, 0, 255)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "orange":
		xy := color.RGBToXY(255, 165, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	default:
		return nil, fmt.Errorf("unknown preset: %s", preset)
	}
	return update, nil
}

func getPresetGroupedLightPut(preset string) (*openhue.GroupedLightPut, error) {
	update := &openhue.GroupedLightPut{On: &openhue.On{On: ptr(true)}}
	switch preset {
	case "warm":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(400)} // 2500K
	case "cool":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(222)} // 4500K
	case "daylight":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(153)} // 6500K
	case "concentration":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(217)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(100))}
	case "relax":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(370)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(50))}
	case "reading":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(312)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(80))}
	case "energize":
		update.ColorTemperature = &openhue.ColorTemperature{Mirek: ptr(166)}
		update.Dimming = &openhue.Dimming{Brightness: ptr(float32(100))}
	case "red":
		xy := color.RGBToXY(255, 0, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "green":
		xy := color.RGBToXY(0, 255, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "blue":
		xy := color.RGBToXY(0, 0, 255)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "purple":
		xy := color.RGBToXY(128, 0, 255)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	case "orange":
		xy := color.RGBToXY(255, 165, 0)
		update.Color = &openhue.Color{Xy: &openhue.GamutPosition{X: ptr(float32(xy.X)), Y: ptr(float32(xy.Y))}}
	default:
		return nil, fmt.Errorf("unknown preset: %s", preset)
	}
	return update, nil
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Error: %v", err),
			},
		},
		IsError: true,
	}
}

func successResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: text,
			},
		},
	}
}
