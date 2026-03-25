package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/loafoe/hue-control/pkg/config"
	"github.com/loafoe/hue-control/pkg/hue"
	hue_mcp "github.com/loafoe/hue-control/pkg/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
)

var (
	bridgeIP   string
	apiKey     string
	jsonOutput bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "hue-control",
		Short: "Philips Hue Control & MCP Server",
		Long:  `A command-line tool to control Philips Hue lights and an MCP server for AI assistants.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip discovery/auth for certain commands if needed
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}

			cfg, err := config.Load()
			if err != nil || !cfg.IsValid() {
				if err != nil {
					fmt.Printf("No existing configuration found (%v). Starting onboarding...\n", err)
				} else {
					fmt.Println("Incomplete configuration found. Starting onboarding...")
				}

				// Discover Bridge
				bridge, err := hue.Discover()
				if err != nil {
					return fmt.Errorf("discovery failed: %w", err)
				}
				fmt.Printf("Discovered bridge at %s (ID: %s)\n", bridge.IP, bridge.ID)

				// Authenticate
				key, err := hue.Authenticate(bridge.IP)
				if err != nil {
					return fmt.Errorf("authentication failed: %w", err)
				}

				// Save Configuration
				cfg = &config.Config{
					BridgeIP: bridge.IP,
					Username: key,
				}
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				fmt.Println("Configuration saved successfully.")
			}
			bridgeIP = cfg.BridgeIP
			apiKey = cfg.Username
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// MCP Command
	var mcpCmd = &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server",
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}

			server := mcp.NewServer(&mcp.Implementation{
				Name:    "Philips Hue Controller (Go)",
				Version: "0.1.0",
			}, nil)

			hue_mcp.RegisterHandlers(server, hueClient)

			if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
				log.Fatalf("Error serving MCP: %v", err)
			}
		},
	}

	// Lights Command
	var lightsCmd = &cobra.Command{
		Use:   "lights",
		Short: "Manage Philips Hue lights",
	}

	var listLightsCmd = &cobra.Command{
		Use:   "list",
		Short: "List all lights",
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}
			lights, err := hueClient.GetLights()
			if err != nil {
				log.Fatalf("Error getting lights: %v", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(lights, "", "  ")
				fmt.Println(string(data))
				return
			}

			for id, light := range lights {
				name := "Unknown"
				if light.Metadata != nil && light.Metadata.Name != nil {
					name = *light.Metadata.Name
				}
				status := "off"
				if light.On != nil && light.On.On != nil && *light.On.On {
					status = "on"
				}
				fmt.Printf("[%s] %s (status: %s)\n", id, name, status)
			}
		},
	}

	var onCmd = &cobra.Command{
		Use:   "on [light-id]",
		Short: "Turn on a light",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}
			err = hueClient.UpdateLightState(args[0], openhue.LightPut{On: &openhue.On{On: func() *bool { b := true; return &b }()}})
			if err != nil {
				log.Fatalf("Error turning on light: %v", err)
			}
			fmt.Printf("Light %s turned on.\n", args[0])
		},
	}

	var offCmd = &cobra.Command{
		Use:   "off [light-id]",
		Short: "Turn off a light",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}
			err = hueClient.UpdateLightState(args[0], openhue.LightPut{On: &openhue.On{On: func() *bool { b := false; return &b }()}})
			if err != nil {
				log.Fatalf("Error turning off light: %v", err)
			}
			fmt.Printf("Light %s turned off.\n", args[0])
		},
	}

	lightsCmd.AddCommand(listLightsCmd, onCmd, offCmd)

	// Sensors Command
	var sensorsCmd = &cobra.Command{
		Use:   "sensors",
		Short: "Read Philips Hue sensors",
	}

	var listMotionCmd = &cobra.Command{
		Use:   "motion",
		Short: "List all motion sensors",
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}
			sensors, err := hueClient.GetMotionSensors()
			if err != nil {
				log.Fatalf("Error getting motion sensors: %v", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(sensors, "", "  ")
				fmt.Println(string(data))
				return
			}

			for id, s := range sensors {
				motion := false
				if s.Motion != nil && s.Motion.MotionReport != nil && s.Motion.MotionReport.Motion != nil {
					motion = *s.Motion.MotionReport.Motion
				}
				fmt.Printf("[%s] Motion: %v\n", id, motion)
			}
		},
	}

	var listTempCmd = &cobra.Command{
		Use:   "temp",
		Short: "List all temperature sensors",
		Run: func(cmd *cobra.Command, args []string) {
			hueClient, err := hue.NewClient(bridgeIP, apiKey)
			if err != nil {
				log.Fatalf("Failed to initialize Hue client: %v", err)
			}
			sensors, err := hueClient.GetTemperatureSensors()
			if err != nil {
				log.Fatalf("Error getting temperature sensors: %v", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(sensors, "", "  ")
				fmt.Println(string(data))
				return
			}

			for id, s := range sensors {
				temp := float32(0)
				if s.Temperature != nil && s.Temperature.TemperatureReport != nil && s.Temperature.TemperatureReport.Temperature != nil {
					temp = *s.Temperature.TemperatureReport.Temperature
				}
				fmt.Printf("[%s] Temperature: %.2fC\n", id, temp)
			}
		},
	}

	sensorsCmd.AddCommand(listMotionCmd, listTempCmd)

	rootCmd.AddCommand(mcpCmd, lightsCmd, sensorsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
