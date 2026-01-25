<!-- Once this directory changes, update this README.md -->

# Servers Module

Unified servers management module consolidating tools, configuration, and overview views.
Master-detail layout with tabbed detail panels (Overview, Tools, Configuration).
Replaces separate /tools and /config routes with unified /servers route.

## Files

- **servers-page.tsx**: Main page component with tabbed layout
- **atoms.ts**: Jotai atoms for server selection state
- **hooks.ts**: Data fetching hooks combining tools, config, and runtime hooks

## Directories

- **components/**: UI components for servers module
