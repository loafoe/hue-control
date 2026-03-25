package hue

import (
	"fmt"
	"time"

	"github.com/openhue/openhue-go"
)

// BridgeInfo contains basic information about a discovered bridge.
type BridgeInfo struct {
	IP string
	ID string
}

// Discover attempts to find a Hue Bridge on the local network.
func Discover() (*BridgeInfo, error) {
	discovery := openhue.NewBridgeDiscovery()
	info, err := discovery.Discover()
	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, fmt.Errorf("no bridge found")
	}

	return &BridgeInfo{
		IP: info.IpAddress,
		ID: info.Instance,
	}, nil
}

// Authenticate handles the Hue "link button" flow.
// It will wait for the user to press the button.
func Authenticate(bridgeIP string) (string, error) {
	auth, err := openhue.NewAuthenticator(bridgeIP, openhue.WithDeviceType("hue-control-go"))
	if err != nil {
		return "", err
	}

	fmt.Println("Please press the link button on your Hue Bridge...")

	for i := 60; i > 0; i-- {
		fmt.Printf("\rWaiting for button press... %ds  ", i)
		apiKey, wait, err := auth.Authenticate()
		if err == nil && apiKey != "" {
			fmt.Println("\nAuthentication successful!")
			return apiKey, nil
		}
		if !wait && err != nil {
			return "", fmt.Errorf("\nauthentication failed: %w", err)
		}
		// if wait is true, it just means the button hasn't been pressed yet
		time.Sleep(1 * time.Second)
	}

	return "", fmt.Errorf("\nauthentication timed out - button not pressed")
}
