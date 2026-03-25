package hue

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/openhue/openhue-go"
)

// Client is a wrapper around the openhue-go client.
type Client struct {
	api openhue.ClientWithResponsesInterface
}

// NewClient creates a new Hue client.
func NewClient(bridgeIP, apiKey string) (*Client, error) {
	authFn := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("hue-application-key", apiKey)
		return nil
	}

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	api, err := openhue.NewClientWithResponses("https://"+bridgeIP, openhue.WithRequestEditorFn(authFn))
	if err != nil {
		return nil, err
	}

	return &Client{api: api}, nil
}

// GetLights returns information about all lights.
func (c *Client) GetLights() (map[string]openhue.LightGet, error) {
	resp, err := c.api.GetLightsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	lights := make(map[string]openhue.LightGet)
	for _, light := range data {
		lights[*light.Id] = light
	}
	return lights, nil
}

// GetRooms returns information about all rooms.
func (c *Client) GetRooms() (map[string]openhue.RoomGet, error) {
	resp, err := c.api.GetRoomsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	rooms := make(map[string]openhue.RoomGet)
	for _, room := range data {
		rooms[*room.Id] = room
	}
	return rooms, nil
}

// GetZones returns information about all zones.
func (c *Client) GetZones() (map[string]openhue.RoomGet, error) {
	resp, err := c.api.GetZonesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	zones := make(map[string]openhue.RoomGet)
	for _, zone := range data {
		zones[*zone.Id] = zone
	}
	return zones, nil
}

// GetScenes returns information about all scenes.
func (c *Client) GetScenes() (map[string]openhue.SceneGet, error) {
	resp, err := c.api.GetScenesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	scenes := make(map[string]openhue.SceneGet)
	for _, scene := range data {
		scenes[*scene.Id] = scene
	}
	return scenes, nil
}

// UpdateLightState updates the state of a specific light.
func (c *Client) UpdateLightState(id string, update openhue.LightPut) error {
	resp, err := c.api.UpdateLightWithResponse(context.Background(), id, update)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}
	return nil
}

// UpdateGroupedLightState updates the state of a grouped_light resource.
func (c *Client) UpdateGroupedLightState(id string, update openhue.GroupedLightPut) error {
	resp, err := c.api.UpdateGroupedLightWithResponse(context.Background(), id, update)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}
	return nil
}

// UpdateSceneState updates a scene.
func (c *Client) UpdateSceneState(id string, update openhue.ScenePut) error {
	resp, err := c.api.UpdateSceneWithResponse(context.Background(), id, update)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}
	return nil
}

// GetMotionSensors returns information about all motion sensors.
func (c *Client) GetMotionSensors() (map[string]openhue.MotionGet, error) {
	resp, err := c.api.GetMotionSensorsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	sensors := make(map[string]openhue.MotionGet)
	for _, sensor := range data {
		sensors[*sensor.Id] = sensor
	}
	return sensors, nil
}

// GetTemperatureSensors returns information about all temperature sensors.
func (c *Client) GetTemperatureSensors() (map[string]openhue.TemperatureGet, error) {
	resp, err := c.api.GetTemperaturesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	data := *(*resp.JSON200).Data
	sensors := make(map[string]openhue.TemperatureGet)
	for _, sensor := range data {
		sensors[*sensor.Id] = sensor
	}
	return sensors, nil
}

// GetGroupedLightByResourceID returns the grouped_light RID for a given room or zone ID.
func (c *Client) GetGroupedLightByResourceID(id string) (string, error) {
	// Try room first
	rooms, err := c.GetRooms()
	if err == nil {
		if room, ok := rooms[id]; ok && room.Services != nil {
			for _, service := range *room.Services {
				if service.Rtype != nil && string(*service.Rtype) == "grouped_light" {
					return *service.Rid, nil
				}
			}
		}
	}

	// Try zone
	zones, err := c.GetZones()
	if err == nil {
		if zone, ok := zones[id]; ok && zone.Services != nil {
			for _, service := range *zone.Services {
				if service.Rtype != nil && string(*service.Rtype) == "grouped_light" {
					return *service.Rid, nil
				}
			}
		}
	}

	return "", fmt.Errorf("grouped_light service not found for resource %s", id)
}
