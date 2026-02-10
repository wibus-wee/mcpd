package ui

import "encoding/json"

const TraySectionKey = "tray"

type TrayClickAction string

const (
	TrayClickActionOpenMenu TrayClickAction = "openMenu"
	TrayClickActionToggle   TrayClickAction = "toggle"
	TrayClickActionShow     TrayClickAction = "show"
)

type TraySettings struct {
	Enabled     bool            `json:"enabled"`
	HideDock    bool            `json:"hideDock"`
	StartHidden bool            `json:"startHidden"`
	ClickAction TrayClickAction `json:"clickAction"`
}

func DefaultTraySettings() TraySettings {
	return TraySettings{
		Enabled:     false,
		HideDock:    false,
		StartHidden: false,
		ClickAction: TrayClickActionOpenMenu,
	}
}

func ParseTraySettings(raw json.RawMessage) (TraySettings, error) {
	settings := DefaultTraySettings()
	if len(raw) == 0 {
		return settings, nil
	}
	var decoded TraySettings
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return settings, err
	}
	settings.Enabled = decoded.Enabled
	settings.HideDock = decoded.HideDock
	settings.StartHidden = decoded.StartHidden
	settings.ClickAction = normalizeTrayClickAction(decoded.ClickAction)
	return settings, nil
}

func normalizeTrayClickAction(action TrayClickAction) TrayClickAction {
	switch action {
	case TrayClickActionOpenMenu, TrayClickActionToggle, TrayClickActionShow:
		return action
	default:
		return TrayClickActionOpenMenu
	}
}
