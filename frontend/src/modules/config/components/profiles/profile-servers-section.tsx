// Input: ProfileDetail servers, ServerItem component, profile actions hook
// Output: ProfileServersSection - servers list section
// Position: Section component in profile detail view

import type { ProfileDetail } from '@bindings/mcpd/internal/ui'
import { ServerIcon } from 'lucide-react'

import { Accordion } from '@/components/ui/accordion'

import { ServerItem, type ServerSpecWithKey } from '../profile-detail/server-item'

interface ProfileServersSectionProps {
  profile: ProfileDetail
  canEdit: boolean
  disabledHint?: string
  pendingServerName: string | null
  onToggleDisabled: (server: ServerSpecWithKey, disabled: boolean) => void
  onDelete: (server: ServerSpecWithKey) => void
}

/**
 * Servers section displaying list of MCP servers with controls.
 */
export function ProfileServersSection({
  profile,
  canEdit,
  disabledHint,
  pendingServerName,
  onToggleDisabled,
  onDelete,
}: ProfileServersSectionProps) {
  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-sm font-medium">Servers</h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            MCP servers configured for this profile
          </p>
        </div>
      </div>

      {profile.servers.length === 0 ? (
        <div className="py-12 text-center">
          <ServerIcon className="size-8 text-muted-foreground/50 mx-auto mb-3" />
          <p className="text-sm font-medium text-muted-foreground">No servers</p>
          <p className="text-xs text-muted-foreground mt-1">
            Import servers or add manually to get started
          </p>
        </div>
      ) : (
        <div className="space-y-1">
          <Accordion multiple>
            {profile.servers.map(server => (
              <ServerItem
                key={server.name}
                server={server}
                canEdit={canEdit}
                isBusy={pendingServerName === server.name}
                disabledHint={disabledHint}
                onToggleDisabled={onToggleDisabled}
                onDelete={onDelete}
              />
            ))}
          </Accordion>
        </div>
      )}
    </section>
  )
}
