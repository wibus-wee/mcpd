// Input: runtime form state, runtime profile metadata
// Output: Runtime settings card using compound component pattern
// Position: Settings page runtime section

import type { ProfileDetail } from '@bindings/mcpd/internal/ui'
import { AlertCircleIcon } from 'lucide-react'
import type * as React from 'react'
import type { UseFormReturn } from 'react-hook-form'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'

import {
  ACTIVATION_MODE_LABELS,
  ACTIVATION_MODE_OPTIONS,
  BOOTSTRAP_MODE_LABELS,
  BOOTSTRAP_MODE_OPTIONS,
  NAMESPACE_STRATEGY_LABELS,
  NAMESPACE_STRATEGY_OPTIONS,
  type RuntimeFormState,
} from '../lib/runtime-config'
import { SettingsCard } from './settings-card'

interface RuntimeSettingsCardProps {
  runtimeProfileName: string | null
  runtimeProfile: ProfileDetail | null | undefined
  runtimeError: unknown
  canEdit: boolean
  form: UseFormReturn<RuntimeFormState>
  statusLabel: string
  saveDisabledReason?: string
  showRuntimeSkeleton: boolean
  hasRuntimeProfile: boolean
  onSubmit: (event?: React.BaseSyntheticEvent) => void
}

export const RuntimeSettingsCard = ({
  runtimeProfileName,
  runtimeProfile,
  runtimeError,
  canEdit,
  form,
  statusLabel,
  saveDisabledReason,
  showRuntimeSkeleton,
  hasRuntimeProfile,
  onSubmit,
}: RuntimeSettingsCardProps) => {
  return (
    <SettingsCard form={form} canEdit={canEdit} onSubmit={onSubmit}>
      <SettingsCard.Header
        title="Runtime"
        description="Adjust timeouts, retries, and global defaults for runtime behavior."
        badge={runtimeProfileName ?? undefined}
      />
      <SettingsCard.Content>
        <SettingsCard.ReadOnlyAlert />
        <SettingsCard.ErrorAlert
          error={runtimeError}
          title="Failed to load runtime settings"
          fallbackMessage="Unable to load runtime configuration."
        />

        {!showRuntimeSkeleton && runtimeProfile === null && (
          <Alert variant="warning">
            <AlertCircleIcon />
            <AlertTitle>Runtime settings unavailable</AlertTitle>
            <AlertDescription>
              The default profile could not be loaded.
            </AlertDescription>
          </Alert>
        )}

        {showRuntimeSkeleton && <RuntimeSkeleton />}

        {hasRuntimeProfile && !showRuntimeSkeleton && (
          <>
            <SettingsCard.Section title="Core">
              <SettingsCard.SelectField<RuntimeFormState>
                name="bootstrapMode"
                label="Bootstrap Mode"
                description="How metadata is collected during startup"
                options={BOOTSTRAP_MODE_OPTIONS}
                labels={BOOTSTRAP_MODE_LABELS}
              />
              <SettingsCard.SelectField<RuntimeFormState>
                name="defaultActivationMode"
                label="Default Activation Mode"
                description="Applied when a server does not specify activationMode"
                options={ACTIVATION_MODE_OPTIONS}
                labels={ACTIVATION_MODE_LABELS}
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="routeTimeoutSeconds"
                label="Route Timeout"
                description="Maximum time to wait for routing requests"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="pingIntervalSeconds"
                label="Ping Interval"
                description="Interval for server health checks (0 to disable)"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="toolRefreshSeconds"
                label="Tool Refresh Interval"
                description="How often to refresh tool lists from servers"
                unit="seconds"
              />
            </SettingsCard.Section>

            <SettingsCard.Section title="Advanced">
              <SettingsCard.NumberField<RuntimeFormState>
                name="bootstrapConcurrency"
                label="Bootstrap Concurrency"
                description="Number of servers to initialize in parallel"
                unit="workers"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="bootstrapTimeoutSeconds"
                label="Bootstrap Timeout"
                description="Maximum time for server initialization"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="toolRefreshConcurrency"
                label="Tool Refresh Concurrency"
                description="Parallel tool refresh operations limit"
                unit="workers"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="callerCheckSeconds"
                label="Caller Check Interval"
                description="How often to check for inactive callers"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="callerInactiveSeconds"
                label="Caller Inactive Threshold"
                description="Time before marking caller as inactive"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="serverInitRetryBaseSeconds"
                label="Init Retry Base Delay"
                description="Initial delay for server initialization retry"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="serverInitRetryMaxSeconds"
                label="Init Retry Max Delay"
                description="Maximum delay for server initialization retry"
                unit="seconds"
              />
              <SettingsCard.NumberField<RuntimeFormState>
                name="serverInitMaxRetries"
                label="Init Max Retries"
                description="Maximum retry attempts for server initialization"
                unit="retries"
              />
              <SettingsCard.SwitchField<RuntimeFormState>
                name="exposeTools"
                label="Expose Tools"
                description="Expose tools to external callers"
              />
              <SettingsCard.SelectField<RuntimeFormState>
                name="toolNamespaceStrategy"
                label="Tool Namespace Strategy"
                description="How to namespace tool names from different servers"
                options={NAMESPACE_STRATEGY_OPTIONS}
                labels={NAMESPACE_STRATEGY_LABELS}
              />
            </SettingsCard.Section>
          </>
        )}
      </SettingsCard.Content>
      <SettingsCard.Footer
        statusLabel={statusLabel}
        saveDisabledReason={saveDisabledReason}
        customDisabled={!hasRuntimeProfile || !canEdit || !form.formState.isDirty || form.formState.isSubmitting}
      />
    </SettingsCard>
  )
}

const RuntimeSkeleton = () => (
  <div className="space-y-4">
    <div className="space-y-2">
      <Skeleton className="h-4 w-24" />
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, index) => (
          <div
            key={index}
            className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,280px)]"
          >
            <div className="space-y-2">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-48" />
            </div>
            <Skeleton className="h-9 w-full sm:max-w-64 sm:justify-self-end" />
          </div>
        ))}
      </div>
    </div>
  </div>
)
