package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

func xSuperProperties() (string, error) {
	props := map[string]interface{}{
		"os":                       "Windows",
		"browser":                  "Chrome",
		"device":                   "",
		"system_locale":            "en-GB",
		"has_client_mods":          false,
		"browser_user_agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64)AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"browser_version":          "120.0",
		"os_version":               "10",
		"referrer":                 "",
		"referring_domain":         "",
		"referrer_current":         "",
		"referring_domain_current": "",
		"release_channel":          "stable",
		"client_build_number":      rand.Intn(100_000) + 400_000,
		"client_event_source":      nil,
		"client_launch_id":         uuid.New().String(),
		"client_app_state":         "unfocused",
	}

	compact, err := json.Marshal(props)
	if err != nil {
		return "", fmt.Errorf("failed to marshal super properties: %w", err)
	}

	return base64.StdEncoding.EncodeToString(compact), nil
}