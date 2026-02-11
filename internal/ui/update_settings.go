package ui

import (
	"encoding/json"
	"time"

	"mcpv/internal/ui/types"
)

const UpdateSectionKey = "updates"

type UpdateSettings struct {
	IntervalHours     int  `json:"intervalHours"`
	IncludePrerelease bool `json:"includePrerelease"`
}

type updateSettingsPayload struct {
	IntervalHours     *int  `json:"intervalHours,omitempty"`
	IncludePrerelease *bool `json:"includePrerelease,omitempty"`
}

func DefaultUpdateSettings() UpdateSettings {
	return UpdateSettings{
		IntervalHours:     int(defaultUpdateInterval / time.Hour),
		IncludePrerelease: true,
	}
}

func ParseUpdateSettings(raw json.RawMessage) (UpdateSettings, error) {
	settings := DefaultUpdateSettings()
	if len(raw) == 0 {
		return settings, nil
	}
	var payload updateSettingsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, err
	}
	if payload.IntervalHours != nil {
		settings.IntervalHours = *payload.IntervalHours
	}
	if payload.IncludePrerelease != nil {
		settings.IncludePrerelease = *payload.IncludePrerelease
	}
	return normalizeUpdateSettings(settings), nil
}

func (s UpdateSettings) ToUpdateCheckOptions() types.UpdateCheckOptions {
	return types.UpdateCheckOptions{
		IntervalHours:     s.IntervalHours,
		IncludePrerelease: s.IncludePrerelease,
	}
}

func normalizeUpdateSettings(settings UpdateSettings) UpdateSettings {
	if settings.IntervalHours <= 0 {
		settings.IntervalHours = int(defaultUpdateInterval / time.Hour)
	}
	return settings
}
