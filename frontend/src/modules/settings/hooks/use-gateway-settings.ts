// Input: UI settings hook, react-hook-form, gateway config helpers
// Output: Gateway settings state + save handler
// Position: Settings gateway hook

import type { BaseSyntheticEvent } from 'react'
import { useEffect, useMemo, useRef } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { useForm, useWatch } from 'react-hook-form'

import { toastManager } from '@/components/ui/toast'
import { useUISettings } from '@/hooks/use-ui-settings'

import type { GatewayFormState } from '../lib/gateway-config'
import {
  buildEndpointPreview,
  buildGatewayPayload,
  DEFAULT_GATEWAY_FORM,
  GATEWAY_SECTION_KEY,
  isLocalHost,
  toGatewayFormState,
} from '../lib/gateway-config'

type UseGatewaySettingsOptions = {
  canEdit: boolean
}

type UseGatewaySettingsResult = {
  form: UseFormReturn<GatewayFormState>
  gatewayLoading: boolean
  gatewayError: unknown
  statusLabel: string
  saveDisabledReason?: string
  validationError?: string
  endpointPreview: string
  visibilityMode: GatewayFormState['visibilityMode']
  accessMode: GatewayFormState['accessMode']
  enabled: boolean
  handleSave: (event?: BaseSyntheticEvent) => void
}

const defaultNetworkHost = '0.0.0.0'

export const useGatewaySettings = ({ canEdit }: UseGatewaySettingsOptions): UseGatewaySettingsResult => {
  const form = useForm<GatewayFormState>({
    defaultValues: DEFAULT_GATEWAY_FORM,
  })
  const { reset, formState, setValue, control } = form
  const { isDirty } = formState

  const {
    error: gatewayError,
    isLoading: gatewayLoading,
    sections,
    updateUISettings,
  } = useUISettings({ scope: 'global' })

  const gatewaySnapshotRef = useRef<string | null>(null)

  useEffect(() => {
    const nextState = toGatewayFormState(sections?.[GATEWAY_SECTION_KEY])
    if (isDirty) {
      return
    }
    const snapshot = JSON.stringify(nextState)
    if (snapshot !== gatewaySnapshotRef.current) {
      gatewaySnapshotRef.current = snapshot
      reset(nextState, { keepDirty: false })
    }
  }, [isDirty, reset, sections])

  const visibilityMode = useWatch({
    control,
    name: 'visibilityMode',
    defaultValue: DEFAULT_GATEWAY_FORM.visibilityMode,
  })
  const accessMode = useWatch({
    control,
    name: 'accessMode',
    defaultValue: DEFAULT_GATEWAY_FORM.accessMode,
  })
  const httpAddr = useWatch({
    control,
    name: 'httpAddr',
    defaultValue: DEFAULT_GATEWAY_FORM.httpAddr,
  })
  const httpPath = useWatch({
    control,
    name: 'httpPath',
    defaultValue: DEFAULT_GATEWAY_FORM.httpPath,
  })
  const httpToken = useWatch({
    control,
    name: 'httpToken',
    defaultValue: DEFAULT_GATEWAY_FORM.httpToken,
  })
  const tagsInput = useWatch({
    control,
    name: 'tagsInput',
    defaultValue: DEFAULT_GATEWAY_FORM.tagsInput,
  })
  const serverName = useWatch({
    control,
    name: 'serverName',
    defaultValue: DEFAULT_GATEWAY_FORM.serverName,
  })
  const enabled = useWatch({
    control,
    name: 'enabled',
    defaultValue: DEFAULT_GATEWAY_FORM.enabled,
  })

  useEffect(() => {
    if (accessMode === 'local' && !isLocalHost(httpAddr)) {
      const nextAddr = coerceLocalAddr(httpAddr)
      setValue('httpAddr', nextAddr, { shouldDirty: true })
    }
    if (accessMode === 'network' && isLocalHost(httpAddr)) {
      const nextAddr = coerceNetworkAddr(httpAddr)
      setValue('httpAddr', nextAddr, { shouldDirty: true })
    }
  }, [accessMode, httpAddr, setValue])

  const endpointPreview = useMemo(
    () => buildEndpointPreview(httpAddr ?? '', httpPath ?? ''),
    [httpAddr, httpPath],
  )

  const validationError = useMemo(() => {
    if (!enabled) return
    if (!httpAddr || httpAddr.trim() === '') {
      return 'HTTP address is required'
    }
    if (accessMode === 'local' && !isLocalHost(httpAddr)) {
      return 'Local access requires a localhost address'
    }
    if (visibilityMode === 'tags' && !tagsInput.trim()) {
      return 'At least one tag is required'
    }
    if (visibilityMode === 'server' && !serverName.trim()) {
      return 'Server name is required'
    }
    if (accessMode === 'network' && !httpToken.trim()) {
      return 'Token is required for network access'
    }
  }, [accessMode, enabled, httpAddr, httpToken, serverName, tagsInput, visibilityMode])

  const statusLabel = useMemo(() => {
    if (gatewayLoading) {
      return 'Loading gateway settings'
    }
    if (gatewayError) {
      return 'Gateway settings unavailable'
    }
    if (validationError) {
      return validationError
    }
    if (isDirty) {
      return 'Unsaved changes'
    }
    return 'All changes saved'
  }, [gatewayError, gatewayLoading, isDirty, validationError])

  const saveDisabledReason = useMemo(() => {
    if (gatewayLoading) {
      return 'Gateway settings are still loading'
    }
    if (gatewayError) {
      return 'Gateway settings are unavailable'
    }
    if (validationError) {
      return validationError
    }
    if (!canEdit) {
      return 'Configuration is read-only'
    }
    if (!isDirty) {
      return 'No changes to save'
    }
    return
  }, [canEdit, gatewayError, gatewayLoading, isDirty, validationError])

  const handleSave = form.handleSubmit(async (values) => {
    if (!canEdit || validationError) {
      if (validationError) {
        toastManager.add({
          type: 'error',
          title: 'Validation required',
          description: validationError,
        })
      }
      return
    }
    try {
      const payload = buildGatewayPayload(values)
      const snapshot = await updateUISettings({
        [GATEWAY_SECTION_KEY]: payload,
      })
      reset(values, { keepDirty: false })
      toastManager.add({
        type: 'success',
        title: 'Gateway updated',
        description: 'Gateway settings applied successfully.',
      })
      return snapshot
    }
    catch (err) {
      toastManager.add({
        type: 'error',
        title: 'Update failed',
        description: err instanceof Error ? err.message : 'Unable to update gateway settings',
      })
    }
  })

  return {
    form,
    gatewayLoading,
    gatewayError,
    statusLabel,
    saveDisabledReason,
    validationError,
    endpointPreview,
    visibilityMode,
    accessMode,
    enabled: Boolean(enabled),
    handleSave,
  }
}

function coerceLocalAddr(addr: string) {
  const port = extractPort(addr) ?? '8090'
  return `127.0.0.1:${port}`
}

function coerceNetworkAddr(addr: string) {
  if (!addr.trim() || isLocalHost(addr)) {
    const port = extractPort(addr) ?? '8090'
    return `${defaultNetworkHost}:${port}`
  }
  return addr
}

function extractPort(addr: string) {
  const trimmed = addr.trim()
  if (!trimmed) return
  if (trimmed.includes('://')) {
    try {
      const url = new URL(trimmed)
      return url.port || undefined
    }
    catch {
      return
    }
  }
  const lastColon = trimmed.lastIndexOf(':')
  if (lastColon === -1) return
  const port = trimmed.slice(lastColon + 1)
  return port || undefined
}
