// Input: ProfileDetail runtime config, SettingRow component, type definitions
// Output: ProfileRuntimeSection - flat settings section for runtime config
// Position: Section component in profile detail view

import type { ProfileDetail } from '@bindings/mcpd/internal/ui'
import { WailsService } from '@bindings/mcpd/internal/ui'
import { ChevronRightIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSWRConfig } from 'swr'

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { toastManager } from '@/components/ui/toast'
import { SettingRow } from '@/components/custom/setting-row'
import { cn } from '@/lib/utils'

import {
  type RuntimeFormState,
  toRuntimeFormState,
  STARTUP_STRATEGY_OPTIONS,
  NAMESPACE_STRATEGY_OPTIONS,
} from '../../types/runtime'
import { reloadConfig } from '../../lib/reload-config'

interface ProfileRuntimeSectionProps {
  profile: ProfileDetail
  canEdit: boolean
}

/**
 * Runtime configuration section with flat, VS Code-style settings layout.
 * Organized in Basic and Advanced subsections with auto-save on change.
 */
export function ProfileRuntimeSection({
  profile,
  canEdit,
}: ProfileRuntimeSectionProps) {
  const { mutate } = useSWRConfig()
  const [formState, setFormState] = useState<RuntimeFormState>(() =>
    toRuntimeFormState(profile.runtime)
  )
  const [isSaving, setIsSaving] = useState(false)

  // Sync form state when profile runtime changes
  useEffect(() => {
    setFormState(toRuntimeFormState(profile.runtime))
  }, [profile.runtime])

  const handleUpdate =
    (field: keyof RuntimeFormState) => async (value: string | number | boolean) => {
      const newState = { ...formState, [field]: value }
      setFormState(newState)

      // Auto-save
      if (!canEdit || isSaving) return

      setIsSaving(true)
      try {
        await WailsService.UpdateRuntimeConfig(newState)
        const reloadResult = await reloadConfig()

        if (!reloadResult.ok) {
          toastManager.add({
            type: 'error',
            title: 'Reload failed',
            description: reloadResult.message,
          })
          // Revert on failure
          setFormState(toRuntimeFormState(profile.runtime))
          return
        }

        await Promise.all([mutate(['profile', profile.name]), mutate('profiles')])

        toastManager.add({
          type: 'success',
          title: 'Runtime updated',
          description: 'Changes applied successfully',
        })
      } catch (err) {
        toastManager.add({
          type: 'error',
          title: 'Update failed',
          description: err instanceof Error ? err.message : 'Unknown error',
        })
        // Revert on error
        setFormState(toRuntimeFormState(profile.runtime))
      } finally {
        setIsSaving(false)
      }
    }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-sm font-medium">Runtime Configuration</h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          Timeout, retry, and concurrency settings for this profile
        </p>
      </div>

      {/* Basic Settings */}
      <div className="space-y-1">
        <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider px-3 py-1">
          Basic
        </h3>
        <div className="space-y-0.5">
          <SettingRow
            label="Startup Strategy"
            description="How servers are initialized when profile activates"
            value={formState.startupStrategy}
            type="select"
            options={STARTUP_STRATEGY_OPTIONS}
            onChange={handleUpdate('startupStrategy')}
            disabled={!canEdit || isSaving}
          />
          <SettingRow
            label="Route Timeout"
            description="Maximum time to wait for routing requests"
            value={formState.routeTimeoutSeconds}
            type="number"
            unit="seconds"
            onChange={handleUpdate('routeTimeoutSeconds')}
            disabled={!canEdit || isSaving}
          />
          <SettingRow
            label="Ping Interval"
            description="Interval for server health checks (0 to disable)"
            value={formState.pingIntervalSeconds}
            type="number"
            unit="seconds"
            onChange={handleUpdate('pingIntervalSeconds')}
            disabled={!canEdit || isSaving}
          />
          <SettingRow
            label="Tool Refresh Interval"
            description="How often to refresh tool lists from servers"
            value={formState.toolRefreshSeconds}
            type="number"
            unit="seconds"
            onChange={handleUpdate('toolRefreshSeconds')}
            disabled={!canEdit || isSaving}
          />
        </div>
      </div>

      {/* Advanced Settings - Collapsible */}
      <Collapsible>
        <CollapsibleTrigger
          className={cn(
            'flex items-center gap-2 text-xs font-medium text-muted-foreground',
            'uppercase tracking-wider px-3 py-1 hover:text-foreground transition-colors w-full',
            'data-[state=open]:[&>svg]:rotate-90'
          )}
        >
          <ChevronRightIcon className="size-3 transition-transform" />
          Advanced
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="space-y-0.5 mt-1">
            <SettingRow
              label="Bootstrap Concurrency"
              description="Number of servers to initialize in parallel"
              value={formState.bootstrapConcurrency}
              type="number"
              onChange={handleUpdate('bootstrapConcurrency')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Bootstrap Timeout"
              description="Maximum time for server initialization"
              value={formState.bootstrapTimeoutSeconds}
              type="number"
              unit="seconds"
              onChange={handleUpdate('bootstrapTimeoutSeconds')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Tool Refresh Concurrency"
              description="Parallel tool refresh operations limit"
              value={formState.toolRefreshConcurrency}
              type="number"
              onChange={handleUpdate('toolRefreshConcurrency')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Caller Check Interval"
              description="How often to check for inactive callers"
              value={formState.callerCheckSeconds}
              type="number"
              unit="seconds"
              onChange={handleUpdate('callerCheckSeconds')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Caller Inactive Threshold"
              description="Time before marking caller as inactive"
              value={formState.callerInactiveSeconds}
              type="number"
              unit="seconds"
              onChange={handleUpdate('callerInactiveSeconds')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Init Retry Base Delay"
              description="Initial delay for server initialization retry"
              value={formState.serverInitRetryBaseSeconds}
              type="number"
              unit="seconds"
              onChange={handleUpdate('serverInitRetryBaseSeconds')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Init Retry Max Delay"
              description="Maximum delay for server initialization retry"
              value={formState.serverInitRetryMaxSeconds}
              type="number"
              unit="seconds"
              onChange={handleUpdate('serverInitRetryMaxSeconds')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Init Max Retries"
              description="Maximum retry attempts for server initialization"
              value={formState.serverInitMaxRetries}
              type="number"
              onChange={handleUpdate('serverInitMaxRetries')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Expose Tools"
              description="Expose tools to external callers"
              value={formState.exposeTools}
              type="switch"
              onChange={handleUpdate('exposeTools')}
              disabled={!canEdit || isSaving}
            />
            <SettingRow
              label="Tool Namespace Strategy"
              description="How to namespace tool names from different servers"
              value={formState.toolNamespaceStrategy}
              type="select"
              options={NAMESPACE_STRATEGY_OPTIONS}
              onChange={handleUpdate('toolNamespaceStrategy')}
              disabled={!canEdit || isSaving}
            />
          </div>
        </CollapsibleContent>
      </Collapsible>
    </section>
  )
}
