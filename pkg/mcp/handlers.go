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
		Description: "Get information about all lights connected to the Hue bridge.",
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
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_light",
		Description: "Get detailed information about a specific light.",
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
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light"`
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
		LightID    string `json:"light_id" jsonschema:"description:The UUID of the light"`
		Brightness int    `json:"brightness" jsonschema:"description:Brightness level (0-100)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_brightness",
		Description: "Set the brightness of a light.",
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
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light"`
		Red     int    `json:"red" jsonschema:"description:Red value (0-255)"`
		Green   int    `json:"green" jsonschema:"description:Green value (0-255)"`
		Blue    int    `json:"blue" jsonschema:"description:Blue value (0-255)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_rgb",
		Description: "Set light color using RGB values.",
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
		LightID     string `json:"light_id" jsonschema:"description:The UUID of the light"`
		Temperature int    `json:"temperature" jsonschema:"description:Color temperature in Kelvin (2000-6500)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_temperature",
		Description: "Set the color temperature of a light in Kelvin.",
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
		Description: "Get information about all groups (rooms and zones).",
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
		Description: "Get information about all rooms.",
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
		Description: "Get information about all zones.",
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
		GroupID string `json:"group_id" jsonschema:"description:The UUID of the room or zone"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "turn_on_group",
		Description: "Turn on all lights in a specific group (room or zone).",
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
		Description: "Turn off all lights in a specific group (room or zone).",
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
		GroupID    string `json:"group_id" jsonschema:"description:The UUID of the room or zone"`
		Brightness int    `json:"brightness" jsonschema:"description:Brightness level (0-100)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_brightness",
		Description: "Set the brightness of all lights in a group.",
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
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_color_rgb",
		Description: "Set color for all lights in a group using RGB values.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args setColorRGBArgs) (*mcp.CallToolResult, any, error) {
		xy := color.RGBToXY(args.Red, args.Green, args.Blue)
		groupedLightID, err := client.GetGroupedLightByResourceID(args.LightID) // setColorRGBArgs uses LightID field
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
		return successResult(fmt.Sprintf("Group %s color set to RGB(%d, %d, %d).", args.LightID, args.Red, args.Green, args.Blue)), nil, nil
	})

	// 15. Get All Scenes
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_scenes",
		Description: "Get information about all scenes.",
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
		SceneID string `json:"scene_id" jsonschema:"description:The UUID of the scene"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_scene",
		Description: "Apply a scene.",
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
		Name string `json:"name" jsonschema:"description:Partial or full name to search for"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_light_by_name",
		Description: "Find lights by searching their names.",
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
		Description: "Make a light flash briefly to identify it.",
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
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light"`
		Effect  string `json:"effect" jsonschema:"description:Effect type (none or colorloop)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_light_effect",
		Description: "Set a dynamic effect on a light.",
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

	// 20. Refresh Lights
	mcp.AddTool(server, &mcp.Tool{
		Name:        "refresh_lights",
		Description: "Refresh information for all lights.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		lights, err := client.GetLights()
		if err != nil {
			return errorResult(err), nil, nil
		}
		return successResult(fmt.Sprintf("Refreshed information for %d lights.", len(lights))), nil, nil
	})

	// 21. Set Color Preset
	type setPresetArgs struct {
		LightID string `json:"light_id" jsonschema:"description:The UUID of the light"`
		Preset  string `json:"preset" jsonschema:"description:Color preset name (warm, cool, daylight, concentration, relax, reading, energize, red, green, blue, purple, orange)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_color_preset",
		Description: "Apply a color preset to a light.",
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

	// 22. Set Group Color Preset
	type setGroupPresetArgs struct {
		GroupID string `json:"group_id" jsonschema:"description:The UUID of the group"`
		Preset  string `json:"preset" jsonschema:"description:Color preset name (warm, cool, daylight, concentration, relax, reading, energize, red, green, blue, purple, orange)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_group_color_preset",
		Description: "Apply a color preset to a group.",
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

	// 23. Get Motion Sensors
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_motion_sensors",
		Description: "Get information about all motion sensors.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		sensors, err := client.GetMotionSensors()
		if err != nil {
			return errorResult(err), nil, nil
		}
		data, _ := json.MarshalIndent(sensors, "", "  ")
		return successResult(string(data)), nil, nil
	})

	// 24. Get Temperature Sensors
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_temperature_sensors",
		Description: "Get information about all temperature sensors.",
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
