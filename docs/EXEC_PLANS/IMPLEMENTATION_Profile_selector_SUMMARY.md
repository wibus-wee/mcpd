# Profile Selector Implementation Summary

## Completed Work

This directory contains the implementation of the Profile Selector feature for mcpd. The changes enable users to switch between different profiles in the UI.

### Backend Changes (Go)

1. **`internal/app/control_plane.go`** (line ~442)
   - Added `GetProfileByName(profileName string) (*profileRuntime, error)` method
   - Allows direct profile lookup without caller resolution
   - Validates profile exists and is active

2. **`internal/ui/service.go`** (lines ~238-328)
   - Added `ListToolsForProfile(ctx, profileName)` - Lists tools for specific profile
   - Added `ListResourcesForProfile(ctx, profileName, cursor)` - Lists resources for specific profile
   - Added `ListPromptsForProfile(ctx, profileName, cursor)` - Lists prompts for specific profile
   - All three methods bypass caller mapping and directly access profiles

3. **Error Handling**
   - Uses existing `ErrCodeProfileNotFound` error code (internal/ui/errors.go:34)
   - Proper error responses for missing or inactive profiles

### Frontend Changes (TypeScript/React)

1. **`frontend/src/modules/config/atoms.ts`**
   - Updated `selectedProfileNameAtom` to use `atomWithStorage` for localStorage persistence
   - Profile selection now persists across page reloads

2. **`frontend/src/modules/config/hooks.ts`**
   - Added `useInitializeSelectedProfile()` hook
   - Auto-selects default profile on app load if none selected
   - Imported `useAtom` for state management

3. **`frontend/src/components/common/profile-selector.tsx`** (NEW FILE)
   - Created dropdown component using Select UI component
   - Displays all available profiles with default indicator
   - Hides when zero or one profile available
   - Animated with Motion library using Spring config

4. **`frontend/src/components/common/app-topbar.tsx`**
   - Integrated `ProfileSelector` component
   - Positioned between status badge and theme toggle in top bar
   - Maintains consistent styling with existing components

## Next Steps (To Complete in Main Repo)

The following tasks require completing in the main mcpd repository:

### 1. Regenerate Wails Bindings
```bash
cd /Users/wibus/dev/mcpd
make wails-bindings
```

This will generate TypeScript bindings for the new backend methods:
- `ListToolsForProfile`
- `ListResourcesForProfile`
- `ListPromptsForProfile`

### 2. Update Frontend Data Hooks

**File: `frontend/src/modules/tools/hooks.ts`**
```typescript
export function useToolsByServer(profileName?: string | null) {
  const { data: tools, isLoading, error } = useSWR<ToolEntry[]>(
    profileName ? ['tools', profileName] : 'tools',
    () => profileName
      ? WailsService.ListToolsForProfile(profileName)
      : WailsService.ListTools(),
  )
  // ... rest of implementation
}
```

**File: `frontend/src/modules/dashboard/hooks.ts`**
```typescript
export function useTools(profileName?: string | null) {
  const swr = useSWR(
    profileName ? ['tools', profileName] : 'tools',
    () => profileName
      ? WailsService.ListToolsForProfile(profileName)
      : WailsService.ListTools(),
    { revalidateOnFocus: false, dedupingInterval: 10000 },
  )
  return { ...swr, tools: swr.data ?? [] }
}

// Similar updates for useResources() and usePrompts()
```

### 3. Wire Up Data Consumers

**File: `frontend/src/routes/tools.tsx`**
```typescript
import { useAtomValue } from 'jotai'
import { selectedProfileNameAtom } from '@/modules/config/atoms'

function ToolsPage() {
  const selectedProfile = useAtomValue(selectedProfileNameAtom)
  const { servers, isLoading, error } = useToolsByServer(selectedProfile)
  // ... rest
}
```

**File: `frontend/src/routes/index.tsx`** (Dashboard)
```typescript
function DashboardPage() {
  const selectedProfile = useAtomValue(selectedProfileNameAtom)
  const { tools } = useTools(selectedProfile)
  const { resources } = useResources(selectedProfile)
  const { prompts } = usePrompts(selectedProfile)
  // ... rest
}
```

### 4. Initialize Profile Selection

**File: `frontend/src/routes/__root.tsx`**
```typescript
import { useInitializeSelectedProfile } from '@/modules/config/hooks'

function RootComponent() {
  useInitializeSelectedProfile() // Add this line
  return (
    // ... existing layout
  )
}
```

### 5. Testing

After completing steps 1-4, test the following:

1. **Profile persistence**: Select a profile, refresh page → should remember selection
2. **Profile switching**: Switch profiles → tools/resources/prompts should update
3. **Loading states**: Switch profiles → should show loading indicator
4. **Edge cases**:
   - No profiles configured → selector should hide
   - Single profile → selector should hide
   - Multiple profiles → selector shows all options
5. **Default profile**: First load → should auto-select default profile

### 6. Run Tests

```bash
cd /Users/wibus/dev/mcpd
make test
```

Ensure all backend tests pass, especially:
- `internal/app/control_plane_test.go`
- `internal/ui/service_test.go`

## Architecture Notes

### Dual API Design

The implementation uses a **dual API approach**:

1. **Caller-based API** (existing): `ListTools(ctx, caller)`
   - Used by MCP protocol clients
   - Maps caller → profile → data
   - Maintains backward compatibility

2. **Profile-based API** (new): `ListToolsForProfile(ctx, profileName)`
   - Used by UI for direct profile selection
   - Bypasses caller mapping
   - Enables multi-profile view

This design:
- Maintains backward compatibility
- Separates concerns (MCP protocol vs UI preferences)
- Allows UI to view any profile without modifying caller mappings

### State Management

- **Profile selection**: Persisted in localStorage via `atomWithStorage`
- **Tool data**: Cached by SWR with profile-specific cache keys: `['tools', profileName]`
- **Auto-initialization**: Default profile selected on first load

### UI Integration

- **Profile selector**: Top bar between status badge and theme toggle
- **Visual feedback**: Shows "(default)" indicator for default profile
- **Smart hiding**: Only shows when 2+ profiles available
- **Animations**: Consistent with app design using Motion + Spring config

## Files Modified

### Backend
- `internal/app/control_plane.go` - Added GetProfileByName helper
- `internal/ui/service.go` - Added 3 profile-scoped API methods

### Frontend
- `frontend/src/modules/config/atoms.ts` - Added localStorage persistence
- `frontend/src/modules/config/hooks.ts` - Added initialization hook
- `frontend/src/components/common/profile-selector.tsx` - NEW component
- `frontend/src/components/common/app-topbar.tsx` - Integrated selector

## Implementation Status

✅ **Completed in `/Users/wibus/dev/profile_selector`:**
- Backend API methods
- Frontend state management
- Profile selector component
- UI integration

⏳ **Remaining (to do in main repo):**
- Regenerate Wails bindings
- Update data fetching hooks
- Wire up route consumers
- Add initialization to root component
- Test end-to-end functionality

## Testing the Implementation

Once all steps are completed in the main repo, the feature can be tested by:

1. Starting the application with multiple profiles configured
2. Verifying the profile selector appears in the top bar
3. Switching between profiles and observing tool list changes
4. Refreshing the page to confirm persistence
5. Checking edge cases (0 profiles, 1 profile, etc.)

The implementation follows mcpd's existing patterns and maintains consistency with the codebase architecture.
