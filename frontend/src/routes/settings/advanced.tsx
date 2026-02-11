// Input: TanStack Router, SystemService bindings, analytics toggle, settings hooks
// Output: Advanced settings page with telemetry, updates, and tray controls
// Position: /settings/advanced route

import { SystemService } from '@bindings/mcpv/internal/ui/services'
import type { UpdateCheckResult } from '@bindings/mcpv/internal/ui/types'
import { createFileRoute } from '@tanstack/react-router'
import { DownloadIcon, RefreshCwIcon } from 'lucide-react'
import { useCallback, useMemo, useState } from 'react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'
import { toastManager } from '@/components/ui/toast'
import {
  toggleAnalytics,
  useAnalyticsEnabledValue,
} from '@/lib/analytics'
import { formatRelativeTime } from '@/lib/time'
import { useTraySettings } from '@/modules/settings/hooks/use-tray-settings'
import { useUpdateSettings } from '@/modules/settings/hooks/use-update-settings'

const formatVersionLabel = (version?: string | null) => {
  if (!version) return 'unknown'
  return version.startsWith('v') ? version : `v${version}`
}

export const Route = createFileRoute('/settings/advanced')({
  component: AdvancedSettingsPage,
})

function AdvancedSettingsPage() {
  const analyticsEnabled = useAnalyticsEnabledValue()
  const {
    settings: updateSettings,
    error: updateError,
    isLoading: updateLoading,
    updateSettings: updateUpdateSettings,
  } = useUpdateSettings()

  const [manualCheckResult, setManualCheckResult] = useState<UpdateCheckResult | null>(null)
  const [manualCheckError, setManualCheckError] = useState<string | null>(null)
  const [manualCheckedAt, setManualCheckedAt] = useState<string | null>(null)
  const [isChecking, setIsChecking] = useState(false)

  const {
    settings: traySettings,
    error: trayError,
    isLoading: trayLoading,
    updateSettings: updateTraySettings,
  } = useTraySettings()

  const isMac = useMemo(() => {
    if (typeof navigator === 'undefined') return false
    return /Mac|iPhone|iPad|iPod/.test(navigator.platform)
  }, [])

  const handlePrereleaseToggle = useCallback(async (checked: boolean) => {
    try {
      await updateUpdateSettings({
        ...updateSettings,
        includePrerelease: checked,
      })
    }
    catch (err) {
      toastManager.add({
        type: 'error',
        title: 'Update preference failed',
        description: err instanceof Error ? err.message : 'Unable to update settings',
      })
    }
  }, [updateSettings, updateUpdateSettings])

  const handleOpenRelease = useCallback(async (url: string) => {
    const opened = window.open(url, '_blank', 'noopener,noreferrer')
    if (opened) return
    try {
      await navigator.clipboard.writeText(url)
      toastManager.add({
        type: 'success',
        title: 'Link copied',
        description: 'Download link copied to clipboard',
      })
    }
    catch {
      toastManager.add({
        type: 'error',
        title: 'Open failed',
        description: 'Unable to open the download link',
      })
    }
  }, [])

  const handleCheckNow = useCallback(async () => {
    if (isChecking) return
    setIsChecking(true)
    setManualCheckError(null)
    try {
      const result = await SystemService.CheckForUpdates()
      setManualCheckResult(result)
      setManualCheckedAt(new Date().toISOString())
    }
    catch (err) {
      const message = err instanceof Error ? err.message : 'Unable to check for updates'
      setManualCheckError(message)
      toastManager.add({
        type: 'error',
        title: 'Update check failed',
        description: message,
      })
    }
    finally {
      setIsChecking(false)
    }
  }, [isChecking])

  const handleTrayUpdate = useCallback(async (next: typeof traySettings) => {
    try {
      await updateTraySettings(next)
    }
    catch (err) {
      toastManager.add({
        type: 'error',
        title: 'Tray preference failed',
        description: err instanceof Error ? err.message : 'Unable to update tray settings',
      })
    }
  }, [updateTraySettings])

  const handleTrayToggle = useCallback((checked: boolean) => {
    handleTrayUpdate({
      ...traySettings,
      enabled: checked,
    })
  }, [handleTrayUpdate, traySettings])

  const handleHideDockToggle = useCallback((checked: boolean) => {
    handleTrayUpdate({
      ...traySettings,
      hideDock: checked,
    })
  }, [handleTrayUpdate, traySettings])

  const handleStartHiddenToggle = useCallback((checked: boolean) => {
    handleTrayUpdate({
      ...traySettings,
      startHidden: checked,
    })
  }, [handleTrayUpdate, traySettings])

  const updateDescription = useMemo(() => {
    const intervalHours = updateSettings.intervalHours || 24
    return `Checks for new releases every ${intervalHours} hours.`
  }, [updateSettings.intervalHours])

  const manualLatest = manualCheckResult?.latest ?? null
  const manualUpdateAvailable = manualCheckResult?.updateAvailable ?? false
  const manualStatus = useMemo(() => {
    if (isChecking) {
      return { label: 'Checking', variant: 'info' as const }
    }
    if (manualCheckError) {
      return { label: 'Failed', variant: 'error' as const }
    }
    if (!manualCheckResult) {
      return { label: 'Not checked', variant: 'outline' as const }
    }
    if (manualUpdateAvailable && manualLatest) {
      return { label: 'Update available', variant: 'warning' as const }
    }
    return { label: 'Up to date', variant: 'success' as const }
  }, [isChecking, manualCheckError, manualCheckResult, manualLatest, manualUpdateAvailable])

  const manualStatusDescription = useMemo(() => {
    if (isChecking) {
      return 'Checking GitHub releases for new builds.'
    }
    if (manualCheckError) {
      return manualCheckError
    }
    if (!manualCheckResult) {
      return 'Run a manual check to verify the latest release.'
    }
    if (manualUpdateAvailable && manualLatest) {
      const latestLabel = formatVersionLabel(manualLatest.version)
      const currentLabel = formatVersionLabel(manualCheckResult.currentVersion)
      return `Latest ${latestLabel} is available. Current ${currentLabel}.`
    }
    const currentLabel = formatVersionLabel(manualCheckResult.currentVersion)
    return currentLabel === 'unknown'
      ? 'You are up to date.'
      : `You are up to date (current ${currentLabel}).`
  }, [isChecking, manualCheckError, manualCheckResult, manualLatest, manualUpdateAvailable])

  return (
    <div className="space-y-3 p-3">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Updates</CardTitle>
          <CardDescription className="text-xs">
            {updateDescription}
          </CardDescription>
          <CardAction>
            <Button
              size="sm"
              variant="secondary"
              onClick={handleCheckNow}
              disabled={isChecking}
            >
              {isChecking ? (
                <Spinner className="size-3.5" />
              ) : (
                <RefreshCwIcon className="size-3.5" />
              )}
              {isChecking ? 'Checking...' : 'Check now'}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="rounded-lg border border-dashed bg-muted/20 p-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">Manual check</span>
                  <Badge variant={manualStatus.variant} size="sm" className="gap-1">
                    {manualStatus.label}
                  </Badge>
                </div>
                <p className={manualCheckError ? 'text-xs text-destructive' : 'text-xs text-muted-foreground'}>
                  {manualStatusDescription}
                </p>
                {manualCheckedAt && (
                  <p className="text-xs text-muted-foreground">
                    Last checked {formatRelativeTime(manualCheckedAt)}.
                  </p>
                )}
                {manualUpdateAvailable && manualLatest?.publishedAt && (
                  <p className="text-xs text-muted-foreground">
                    Released {formatRelativeTime(manualLatest.publishedAt)}.
                  </p>
                )}
              </div>
              {manualUpdateAvailable && manualLatest?.url && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleOpenRelease(manualLatest.url)}
                >
                  <DownloadIcon className="size-4" />
                  Download
                </Button>
              )}
            </div>
          </div>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="updates-prerelease" className="text-sm font-medium">
                Include pre-release versions
              </Label>
              <p className="text-xs text-muted-foreground">
                Enable this to receive early access builds with new features.
              </p>
            </div>
            <Switch
              checked={updateSettings.includePrerelease}
              disabled={updateLoading || !!updateError}
              id="updates-prerelease"
              onCheckedChange={handlePrereleaseToggle}
            />
          </div>
          {updateError ? (
            <p className="text-xs text-destructive">
              Unable to load update preferences.
            </p>
          ) : null}
          {import.meta.env.DEV && (
            <p className="text-xs text-muted-foreground">
              Update checks are disabled in development builds.
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Menu Bar & Tray</CardTitle>
          <CardDescription className="text-xs">
            Enable the menu bar tray for quick access to mcpv.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="tray-enabled" className="text-sm font-medium">
                Enable tray icon
              </Label>
              <p className="text-xs text-muted-foreground">
                Show a tray icon and keep the app running in the background.
              </p>
            </div>
            <Switch
              checked={traySettings.enabled}
              disabled={trayLoading}
              id="tray-enabled"
              onCheckedChange={handleTrayToggle}
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="tray-hide-dock" className="text-sm font-medium">
                Hide Dock icon when tray is enabled
              </Label>
              <p className="text-xs text-muted-foreground">
                Keep mcpv out of the Dock while the tray icon is active.
              </p>
            </div>
            <Switch
              checked={traySettings.hideDock}
              disabled={!traySettings.enabled || trayLoading || !isMac}
              id="tray-hide-dock"
              onCheckedChange={handleHideDockToggle}
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="tray-start-hidden" className="text-sm font-medium">
                Start hidden
              </Label>
              <p className="text-xs text-muted-foreground">
                Launch directly to the tray without showing the main window.
              </p>
            </div>
            <Switch
              checked={traySettings.startHidden}
              disabled={!traySettings.enabled || trayLoading}
              id="tray-start-hidden"
              onCheckedChange={handleStartHiddenToggle}
            />
          </div>

          {!isMac && (
            <p className="text-xs text-muted-foreground">
              Dock visibility controls are only available on macOS.
            </p>
          )}
          {trayError ? (
            <p className="text-xs text-destructive">
              Unable to load tray settings.
            </p>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Telemetry</CardTitle>
          <CardDescription className="text-xs">
            Help improve mcpv by sending anonymous usage data.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="analytics-toggle" className="text-sm font-medium">
                Send anonymous usage data
              </Label>
              <p className="text-xs text-muted-foreground">
                We collect anonymous usage statistics to improve the app. No personal data is collected.
              </p>
            </div>
            <Switch
              checked={analyticsEnabled}
              id="analytics-toggle"
              onCheckedChange={toggleAnalytics}
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            Debug & Diagnostics
            <Badge variant="secondary" size="sm">
              Coming Soon
            </Badge>
          </CardTitle>
          <CardDescription className="text-xs">
            Debug logs and diagnostic tools will be available here.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Advanced debugging features are currently under development.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
