<!-- Once this directory changes, update this README.md -->

# Modules/Config/Components/Profiles

Profile-related components for flat, desktop-style configuration UI.

## Files

- **profile-overview-section.tsx**: Overview section with active callers and stats
- **profile-subagent-section.tsx**: SubAgent enable/disable section
- **profile-servers-section.tsx**: Servers list section with ServerItem components
- **profile-runtime-section.tsx**: Runtime configuration section (moved to Settings page)
- **index.ts**: Centralized exports

## Design Pattern

All section components follow the flat, VS Code-style layout:
- No Card shadow or heavy borders
- Section headers with description
- SettingRow for configuration fields
- Separator between sections
