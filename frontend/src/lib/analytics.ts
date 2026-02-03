// Input: @umami/node, jotai store, localStorage
// Output: Analytics service with enable/disable toggle and track function
// Position: Core analytics utility for Umami integration

import umami from '@umami/node'
import { atom } from 'jotai'

import { createAtomHooks, jotaiStore } from './jotai'

const STORAGE_KEY = 'mcpv-analytics-enabled'
const UMAMI_WEBSITE_ID = import.meta.env.VITE_UMAMI_WEBSITE_ID || '__UMAMI_WEBSITE_ID__'
const UMAMI_HOST_URL = import.meta.env.VITE_UMAMI_HOST_URL || '__UMAMI_HOST_URL__'

const analyticsEnabledAtom = atom(
  typeof localStorage !== 'undefined'
    ? localStorage.getItem(STORAGE_KEY) !== 'false'
    : true,
)

export const [
  useAnalyticsEnabled,
  useSetAnalyticsEnabled,
  useAnalyticsEnabledValue,,
  getAnalyticsEnabled,
  setAnalyticsEnabled,
] = createAtomHooks(analyticsEnabledAtom)

let initialized = false

const ensureInit = () => {
  if (initialized) return
  if (
    UMAMI_WEBSITE_ID === '__UMAMI_WEBSITE_ID__'
    || UMAMI_HOST_URL === '__UMAMI_HOST_URL__'
  ) {
    return
  }
  umami.init({
    userAgent: navigator.userAgent,
    websiteId: UMAMI_WEBSITE_ID,
    hostUrl: UMAMI_HOST_URL,
  })
  initialized = true
}

export const track = (name: string, data?: Record<string, unknown>) => {
  if (!jotaiStore.get(analyticsEnabledAtom)) return
  ensureInit()
  if (!initialized) return
  umami.track({ name, data })
}

export const trackPageView = (url: string, title?: string) => {
  if (!jotaiStore.get(analyticsEnabledAtom)) return
  ensureInit()
  if (!initialized) return
  umami.track({ url, title })
}

export const toggleAnalytics = (enabled: boolean) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, String(enabled))
  }
  setAnalyticsEnabled(enabled)
}

export const AnalyticsEvents = {
  APP_LAUNCH: 'app_launch',
  PAGE_VIEW: 'page_view',
  SERVER_START: 'server_start',
  SERVER_STOP: 'server_stop',
  SERVER_DELETE: 'server_delete',
  SERVER_CREATE: 'server_create',
  SETTINGS_SAVE: 'settings_save',
  PLUGIN_INSTALL: 'plugin_install',
  PLUGIN_REMOVE: 'plugin_remove',
} as const
