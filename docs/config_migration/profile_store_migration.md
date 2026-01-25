Profile Store Migration Guide

Overview
- This document describes how to migrate configuration management from a monolithic runtime.yaml to a profile-store layout:
  - runtime.yaml at the repository root (global runtime settings)
  - profiles/*.yaml containing per-profile catalog definitions
  - callers.yaml mapping callers to profiles

Why migrate?
- Improves modularity and multi-profile support
- Aligns with runtime behavior tested in tests like ProfileStoreLoader_RuntimeOverrideFromStore
- Simplifies hot-reload and isolated updates per profile

Preconditions
- You have an existing runtime.yaml with global runtime configuration
- You have server definitions currently in a single location (or multiple servers per profile)
- You can create a profiles/ directory and a callers.yaml file in your profile store path

Recommended directory structure (example)
- profile store path (e.g. ./profile-store)
  - runtime.yaml (global runtime settings)
  - profiles/
    - default.yaml (default profile with its servers)
    - chat.yaml
    - vscode.yaml
  - callers.yaml (caller -> profile mapping)

Migration steps (conservative, low-risk)
1. Create the profile-store scaffold
   - Create profiles/ directory if not exists
   - Create default.yaml under profiles/ with servers block copied from the current global runtime where appropriate
   - Create any additional profile yaml files (e.g. chat.yaml) mirroring per-profile server definitions
   - Copy runtime settings that are global into root runtime.yaml; if some settings are per-profile in the old format, consider moving them into the default profile or new per-profile overrides as needed
2. Create a minimal callers.yaml
   - Example content:
     callers:
       default: default
   - This maps a default caller to the default profile
3. Validate the migration locally
   - Run: go test ./... to ensure tests pass with the new layout
   - Run loader against the profile-store directory to verify loading works as intended
4. Update runtime onboarding / startup scripts
   - Ensure the application loads the profile-store directory and not the old single-runtime config when starting in profile-store mode
5. Document the rollout
   - Update docs/EXEC_PLANS and AGENTS.md to reflect the new recommended configuration approach

Validation criteria
- The loader loads all profiles from profiles/*, applies Runtime overrides from runtime.yaml when provided, and respects the callers.yaml mapping
- go test passes
- Community/CI tests pass with the new layout

Notes
- If you rely on CI to produce a stable configuration, consider pinning a known-good profile-store layout in docs/examples
- The existing tests include scenarios for default profile creation, runtime override, and caller validation which align with this migration approach
