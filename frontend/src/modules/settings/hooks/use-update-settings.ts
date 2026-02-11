// Input: UISettingsService bindings, UI settings hook
// Output: Update settings hook with load and save helpers
// Position: Settings module hook for update preferences

import type { UISettingsSnapshot } from '@bindings/mcpv/internal/ui/types'

import { useUISettings } from '@/hooks/use-ui-settings'

export type UpdateSettings = {
  intervalHours: number
  includePrerelease: boolean
}

const updateSectionKey = 'updates'

const defaultUpdateSettings: UpdateSettings = {
  intervalHours: 24,
  includePrerelease: true,
}

function parseUpdateSettings(section: unknown): UpdateSettings {
  if (!section) return defaultUpdateSettings
  const data = resolveSectionObject(section)
  if (!data) return defaultUpdateSettings
  return {
    intervalHours: resolveIntervalHours(data.intervalHours),
    includePrerelease: Boolean(data.includePrerelease),
  }
}

function resolveSectionObject(section: unknown): Partial<UpdateSettings> | null {
  if (typeof section === 'string') {
    try {
      const parsed = JSON.parse(section) as unknown
      return typeof parsed === 'object' && parsed ? parsed as Partial<UpdateSettings> : null
    }
    catch {
      return null
    }
  }
  if (typeof section === 'object') {
    return section as Partial<UpdateSettings>
  }
  return null
}

function resolveIntervalHours(value: unknown): number {
  if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
    return value
  }
  if (typeof value === 'string') {
    const parsed = Number(value)
    if (Number.isFinite(parsed) && parsed > 0) {
      return parsed
    }
  }
  return defaultUpdateSettings.intervalHours
}

export type UseUpdateSettingsResult = {
  settings: UpdateSettings
  isLoading: boolean
  error: unknown
  updateSettings: (next: UpdateSettings) => Promise<UISettingsSnapshot>
}

export function useUpdateSettings(): UseUpdateSettingsResult {
  const { error, isLoading, sections, updateUISettings } = useUISettings({ scope: 'global' })

  const settings = parseUpdateSettings(sections?.[updateSectionKey])

  const updateSettings = async (next: UpdateSettings) => {
    return updateUISettings({
      [updateSectionKey]: next,
    })
  }

  return {
    settings,
    isLoading,
    error,
    updateSettings,
  }
}
