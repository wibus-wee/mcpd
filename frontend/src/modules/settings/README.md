<!-- Once this directory changes, update this README.md -->

# Modules/Settings

Settings module for configuring mcpd runtime and profile-level options.
Currently focuses on SubAgent LLM-based tool filtering configuration.
Follows the domain-based component organization pattern.

## Files

- **components/subagent-config-form.tsx**: Runtime-level SubAgent LLM provider configuration form (read-only, displays values from runtime.yaml)
- **components/profile-subagent-toggle.tsx**: Per-profile SubAgent enable/disable toggle switch with Wails backend integration

## Architecture

### State Management
- Uses Jotai atoms from `@/atoms/subagent.ts` for global state
- Fetches config via Wails bindings from Go backend
- Updates persist to YAML files via backend API

### Configuration Levels
1. **Runtime Config** (runtime.yaml): LLM provider settings (model, provider, apiKeyEnvVar, etc.) - shared across all profiles
2. **Profile Config** (profiles/xxx.yaml): Per-profile `enabled` flag - controls whether SubAgent is active for that profile

### Backend Integration
- `GetSubAgentConfig()` - Fetch runtime-level LLM config
- `GetProfileSubAgentConfig(profileName)` - Fetch per-profile enabled state
- `SetProfileSubAgentEnabled(req)` - Update per-profile enabled state
- `IsSubAgentAvailable()` - Check if SubAgent infrastructure is configured
