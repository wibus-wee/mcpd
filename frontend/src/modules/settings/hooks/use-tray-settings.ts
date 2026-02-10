// Input: UI settings hook
// Output: Tray settings hook with defaults
// Position: Settings module hook

import type { UISettingsSnapshot } from '@bindings/mcpv/internal/ui/types'

import { useUISettings } from '@/hooks/use-ui-settings'

export type TrayClickAction = 'openMenu' | 'toggle' | 'show'

export type TraySettings = {
  enabled: boolean
  hideDock: boolean
  startHidden: boolean
  clickAction: TrayClickAction
}

const traySectionKey = 'tray'

const defaultTraySettings: TraySettings = {
  enabled: false,
  hideDock: false,
  startHidden: false,
  clickAction: 'openMenu',
}

function parseTraySettings(section: unknown): TraySettings {
  if (!section) return defaultTraySettings
  const data = resolveSectionObject(section)
  if (!data) return defaultTraySettings
  const clickAction = resolveClickAction(data.clickAction)
  return {
    enabled: Boolean(data.enabled),
    hideDock: Boolean(data.hideDock),
    startHidden: Boolean(data.startHidden),
    clickAction,
  }
}

function resolveSectionObject(section: unknown): Partial<TraySettings> | null {
  if (typeof section === 'string') {
    try {
      const parsed = JSON.parse(section) as unknown
      return typeof parsed === 'object' && parsed ? parsed as Partial<TraySettings> : null
    }
    catch {
      return null
    }
  }
  if (typeof section === 'object') {
    return section as Partial<TraySettings>
  }
  return null
}

function resolveClickAction(value: unknown): TrayClickAction {
  switch (value) {
    case 'toggle':
    case 'show':
    case 'openMenu':
      return value
    default:
      return 'openMenu'
  }
}

export type UseTraySettingsResult = {
  settings: TraySettings
  isLoading: boolean
  error: unknown
  updateSettings: (next: TraySettings) => Promise<UISettingsSnapshot>
}

export function useTraySettings(): UseTraySettingsResult {
  const { error, isLoading, sections, updateUISettings } = useUISettings({ scope: 'global' })

  const settings = parseTraySettings(sections?.[traySectionKey])

  const updateSettings = async (next: TraySettings) => {
    return updateUISettings({
      [traySectionKey]: next,
    })
  }

  return {
    settings,
    isLoading,
    error,
    updateSettings,
  }
}
