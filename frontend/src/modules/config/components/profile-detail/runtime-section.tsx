// Input: ProfileDetail type
// Output: RuntimeSection accordion component
// Position: Profile runtime configuration display

import type { ProfileDetail } from '@bindings/mcpd/internal/ui'
import { WailsService } from '@bindings/mcpd/internal/ui'
import { NetworkIcon, SettingsIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSWRConfig } from 'swr'

import {
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { toastManager } from '@/components/ui/toast'

import { DetailRow } from './detail-row'
import { reloadConfig } from '../../lib/reload-config'

interface RuntimeSectionProps {
  profile: ProfileDetail
  canEdit?: boolean
  disabledHint?: string
}

type RuntimeFormState = {
  routeTimeoutSeconds: number
  pingIntervalSeconds: number
  toolRefreshSeconds: number
  toolRefreshConcurrency: number
  callerCheckSeconds: number
  callerInactiveSeconds: number
  serverInitRetryBaseSeconds: number
  serverInitRetryMaxSeconds: number
  serverInitMaxRetries: number
  startupStrategy: string
  bootstrapConcurrency: number
  bootstrapTimeoutSeconds: number
  exposeTools: boolean
  toolNamespaceStrategy: string
}

/**
 * Displays the runtime configuration section within an accordion.
 */
export function RuntimeSection({ profile, canEdit, disabledHint }: RuntimeSectionProps) {
  const { runtime } = profile
  const { mutate } = useSWRConfig()
  const [isEditing, setIsEditing] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [formState, setFormState] = useState<RuntimeFormState>(() => ({
    routeTimeoutSeconds: runtime.routeTimeoutSeconds,
    pingIntervalSeconds: runtime.pingIntervalSeconds,
    toolRefreshSeconds: runtime.toolRefreshSeconds,
    toolRefreshConcurrency: runtime.toolRefreshConcurrency,
    callerCheckSeconds: runtime.callerCheckSeconds,
    callerInactiveSeconds: runtime.callerInactiveSeconds,
    serverInitRetryBaseSeconds: runtime.serverInitRetryBaseSeconds,
    serverInitRetryMaxSeconds: runtime.serverInitRetryMaxSeconds,
    serverInitMaxRetries: runtime.serverInitMaxRetries,
    startupStrategy: runtime.startupStrategy || 'lazy',
    bootstrapConcurrency: runtime.bootstrapConcurrency,
    bootstrapTimeoutSeconds: runtime.bootstrapTimeoutSeconds,
    exposeTools: runtime.exposeTools,
    toolNamespaceStrategy: runtime.toolNamespaceStrategy || 'prefix',
  }))

  useEffect(() => {
    setFormState({
      routeTimeoutSeconds: runtime.routeTimeoutSeconds,
      pingIntervalSeconds: runtime.pingIntervalSeconds,
      toolRefreshSeconds: runtime.toolRefreshSeconds,
      toolRefreshConcurrency: runtime.toolRefreshConcurrency,
      callerCheckSeconds: runtime.callerCheckSeconds,
      callerInactiveSeconds: runtime.callerInactiveSeconds,
      serverInitRetryBaseSeconds: runtime.serverInitRetryBaseSeconds,
      serverInitRetryMaxSeconds: runtime.serverInitRetryMaxSeconds,
      serverInitMaxRetries: runtime.serverInitMaxRetries,
      startupStrategy: runtime.startupStrategy || 'lazy',
      bootstrapConcurrency: runtime.bootstrapConcurrency,
      bootstrapTimeoutSeconds: runtime.bootstrapTimeoutSeconds,
      exposeTools: runtime.exposeTools,
      toolNamespaceStrategy: runtime.toolNamespaceStrategy || 'prefix',
    })
  }, [runtime])

  const handleSave = async () => {
    if (!canEdit || isSaving) {
      return
    }
    setIsSaving(true)
    try {
      await WailsService.UpdateRuntimeConfig({
        routeTimeoutSeconds: formState.routeTimeoutSeconds,
        pingIntervalSeconds: formState.pingIntervalSeconds,
        toolRefreshSeconds: formState.toolRefreshSeconds,
        toolRefreshConcurrency: formState.toolRefreshConcurrency,
        callerCheckSeconds: formState.callerCheckSeconds,
        callerInactiveSeconds: formState.callerInactiveSeconds,
        serverInitRetryBaseSeconds: formState.serverInitRetryBaseSeconds,
        serverInitRetryMaxSeconds: formState.serverInitRetryMaxSeconds,
        serverInitMaxRetries: formState.serverInitMaxRetries,
        startupStrategy: formState.startupStrategy,
        bootstrapConcurrency: formState.bootstrapConcurrency,
        bootstrapTimeoutSeconds: formState.bootstrapTimeoutSeconds,
        exposeTools: formState.exposeTools,
        toolNamespaceStrategy: formState.toolNamespaceStrategy,
      })

      const reloadResult = await reloadConfig()
      if (!reloadResult.ok) {
        toastManager.add({
          type: 'error',
          title: 'Reload failed',
          description: reloadResult.message,
        })
        return
      }

      await Promise.all([
        mutate(['profile', profile.name]),
        mutate('profiles'),
      ])

      toastManager.add({
        type: 'success',
        title: 'Runtime updated',
        description: 'Changes applied.',
      })
      setIsEditing(false)
    } catch (err) {
      toastManager.add({
        type: 'error',
        title: 'Update failed',
        description: err instanceof Error ? err.message : 'Update failed',
      })
    } finally {
      setIsSaving(false)
    }
  }

  const handleCancel = () => {
    setFormState({
      routeTimeoutSeconds: runtime.routeTimeoutSeconds,
      pingIntervalSeconds: runtime.pingIntervalSeconds,
      toolRefreshSeconds: runtime.toolRefreshSeconds,
      toolRefreshConcurrency: runtime.toolRefreshConcurrency,
      callerCheckSeconds: runtime.callerCheckSeconds,
      callerInactiveSeconds: runtime.callerInactiveSeconds,
      serverInitRetryBaseSeconds: runtime.serverInitRetryBaseSeconds,
      serverInitRetryMaxSeconds: runtime.serverInitRetryMaxSeconds,
      serverInitMaxRetries: runtime.serverInitMaxRetries,
      startupStrategy: runtime.startupStrategy || 'lazy',
      bootstrapConcurrency: runtime.bootstrapConcurrency,
      bootstrapTimeoutSeconds: runtime.bootstrapTimeoutSeconds,
      exposeTools: runtime.exposeTools,
      toolNamespaceStrategy: runtime.toolNamespaceStrategy || 'prefix',
    })
    setIsEditing(false)
  }

  return (
    <AccordionItem value="runtime" className="border-none">
      <AccordionTrigger className="py-2 hover:no-underline">
        <div className="flex items-center gap-2">
          <SettingsIcon className="size-3.5 text-muted-foreground" />
          <span className="text-sm font-medium">Runtime Configuration</span>
        </div>
      </AccordionTrigger>
      <AccordionContent className="pb-0">
        <div className="flex items-center justify-between pb-2">
          <span className="text-xs text-muted-foreground">Runtime defaults shared across profiles.</span>
          <Button
            variant="ghost"
            size="xs"
            onClick={() => setIsEditing(true)}
            disabled={!canEdit || isEditing}
            title={disabledHint}
          >
            Edit
          </Button>
        </div>

        {isEditing ? (
          <div className="space-y-4 pb-3">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>Startup Strategy</Label>
                <Select
                  value={formState.startupStrategy}
                  onValueChange={(value) => setFormState(prev => ({ ...prev, startupStrategy: value }))}
                >
                  <SelectTrigger size="sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="lazy">Lazy</SelectItem>
                    <SelectItem value="eager">Eager</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Bootstrap Concurrency</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.bootstrapConcurrency}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      bootstrapConcurrency: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Bootstrap Timeout (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.bootstrapTimeoutSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      bootstrapTimeoutSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Route Timeout (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.routeTimeoutSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      routeTimeoutSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Ping Interval (s)</Label>
                <Input
                  type="number"
                  min={0}
                  value={formState.pingIntervalSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      pingIntervalSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Tool Refresh (s)</Label>
                <Input
                  type="number"
                  min={0}
                  value={formState.toolRefreshSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      toolRefreshSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Tool Refresh Concurrency</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.toolRefreshConcurrency}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      toolRefreshConcurrency: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Caller Check (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.callerCheckSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      callerCheckSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Caller Inactive (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.callerInactiveSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      callerInactiveSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Init Retry Base (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.serverInitRetryBaseSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      serverInitRetryBaseSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Init Retry Max (s)</Label>
                <Input
                  type="number"
                  min={1}
                  value={formState.serverInitRetryMaxSeconds}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      serverInitRetryMaxSeconds: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Init Max Retries</Label>
                <Input
                  type="number"
                  min={0}
                  value={formState.serverInitMaxRetries}
                  onChange={(event) => {
                    const next = Number.parseInt(event.target.value, 10)
                    setFormState(prev => ({
                      ...prev,
                      serverInitMaxRetries: Number.isNaN(next) ? 0 : next,
                    }))
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label>Namespace Strategy</Label>
                <Select
                  value={formState.toolNamespaceStrategy}
                  onValueChange={(value) => setFormState(prev => ({ ...prev, toolNamespaceStrategy: value }))}
                >
                  <SelectTrigger size="sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="prefix">Prefix</SelectItem>
                    <SelectItem value="flat">Flat</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2 pt-6">
                <Switch
                  checked={formState.exposeTools}
                  onCheckedChange={(checked) => setFormState(prev => ({ ...prev, exposeTools: checked }))}
                />
                <span className="text-xs text-muted-foreground">Expose Tools</span>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={handleCancel} disabled={isSaving}>
                Cancel
              </Button>
              <Button size="sm" onClick={handleSave} disabled={!canEdit || isSaving}>
                {isSaving ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </div>
        ) : (
          <div className="divide-y divide-border/50 pb-3">
            <DetailRow
              label="Startup Strategy"
              value={(
                <Badge variant="outline" size="sm">
                  {runtime.startupStrategy || 'lazy'}
                </Badge>
              )}
            />
            <DetailRow label="Bootstrap Concurrency" value={`${runtime.bootstrapConcurrency}`} mono />
            <DetailRow label="Bootstrap Timeout" value={`${runtime.bootstrapTimeoutSeconds}s`} mono />
            <DetailRow label="Route Timeout" value={`${runtime.routeTimeoutSeconds}s`} mono />
            <DetailRow label="Ping Interval" value={`${runtime.pingIntervalSeconds}s`} mono />
            <DetailRow label="Tool Refresh" value={`${runtime.toolRefreshSeconds}s`} mono />
            <DetailRow label="Tool Refresh Concurrency" value={`${runtime.toolRefreshConcurrency}`} mono />
            <DetailRow label="Caller Check" value={`${runtime.callerCheckSeconds}s`} mono />
            <DetailRow label="Caller Inactive" value={`${runtime.callerInactiveSeconds}s`} mono />
            <DetailRow label="Init Retry Base" value={`${runtime.serverInitRetryBaseSeconds}s`} mono />
            <DetailRow label="Init Retry Max" value={`${runtime.serverInitRetryMaxSeconds}s`} mono />
            <DetailRow label="Init Max Retries" value={`${runtime.serverInitMaxRetries}`} mono />
            <DetailRow
              label="Expose Tools"
              value={
                <Badge variant={runtime.exposeTools ? 'success' : 'secondary'} size="sm">
                  {runtime.exposeTools ? 'Yes' : 'No'}
                </Badge>
              }
            />
            <DetailRow
              label="Namespace Strategy"
              value={(
                <Badge variant="outline" size="sm">
                  {runtime.toolNamespaceStrategy || 'prefix'}
                </Badge>
              )}
            />
          </div>
        )}

        <div className="border-t pt-3 pb-2">
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground mb-2">
            <NetworkIcon className="size-3" />
            RPC Configuration
          </div>
          <div className="divide-y divide-border/50">
            <DetailRow
              label="Listen Address"
              value={
                <span className="font-mono text-xs truncate max-w-40 block text-right">
                  {runtime.rpc.listenAddress}
                </span>
              }
            />
            <DetailRow label="Socket Mode" value={runtime.rpc.socketMode || '0660'} mono />
          </div>
        </div>
      </AccordionContent>
    </AccordionItem>
  )
}
