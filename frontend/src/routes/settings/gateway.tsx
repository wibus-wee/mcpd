// Input: TanStack Router, gateway settings components
// Output: Gateway settings page
// Position: /settings/gateway route

import { createFileRoute } from '@tanstack/react-router'

import { useConfigMode } from '@/modules/servers/hooks'
import { GatewaySettingsCard } from '@/modules/settings/components/gateway-settings-card'
import { useGatewaySettings } from '@/modules/settings/hooks/use-gateway-settings'

export const Route = createFileRoute('/settings/gateway')({
  component: GatewaySettingsPage,
})

function GatewaySettingsPage() {
  const { data: configMode } = useConfigMode()
  const canEdit = Boolean(configMode?.isWritable)
  const gateway = useGatewaySettings({ canEdit })

  return (
    <div className="p-3">
      <GatewaySettingsCard
        canEdit={canEdit}
        form={gateway.form}
        statusLabel={gateway.statusLabel}
        saveDisabledReason={gateway.saveDisabledReason}
        gatewayLoading={gateway.gatewayLoading}
        gatewayError={gateway.gatewayError}
        validationError={gateway.validationError}
        endpointPreview={gateway.endpointPreview}
        visibilityMode={gateway.visibilityMode}
        accessMode={gateway.accessMode}
        enabled={gateway.enabled}
        onSubmit={gateway.handleSave}
      />
    </div>
  )
}
