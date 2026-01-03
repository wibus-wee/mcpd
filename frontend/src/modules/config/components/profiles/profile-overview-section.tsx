// Input: ActiveCaller type, ProfileDetail, CallerChipGroup component
// Output: ProfileOverviewSection - overview section with active callers and stats
// Position: Section component in profile detail view

import type { ActiveCaller, ProfileDetail } from '@bindings/mcpd/internal/ui'

import { Badge } from '@/components/ui/badge'
import { CallerChipGroup } from '@/components/common/caller-chip-group'

interface ProfileOverviewSectionProps {
  profile: ProfileDetail
  activeCallers: ActiveCaller[]
}

/**
 * Overview section displaying active callers and quick stats.
 */
export function ProfileOverviewSection({
  profile,
  activeCallers,
}: ProfileOverviewSectionProps) {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-sm font-medium">Overview</h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          Active callers and quick statistics
        </p>
      </div>

      <div className="space-y-0.5">
        <div className="flex items-center justify-between py-2.5 px-3 rounded-md hover:bg-muted/50 transition-colors">
          <div className="flex-1">
            <div className="text-sm font-medium">Active Callers</div>
            <div className="text-xs text-muted-foreground">
              Currently connected clients
            </div>
          </div>
          <div className="shrink-0">
            <CallerChipGroup
              callers={activeCallers}
              maxVisible={3}
              showPid
              emptyText="No active callers"
            />
          </div>
        </div>

        <div className="flex items-center justify-between py-2.5 px-3 rounded-md hover:bg-muted/50 transition-colors">
          <div className="flex-1">
            <div className="text-sm font-medium">Total Servers</div>
            <div className="text-xs text-muted-foreground">
              Configured MCP servers
            </div>
          </div>
          <div className="shrink-0">
            <Badge variant="secondary">{profile.servers.length}</Badge>
          </div>
        </div>
      </div>
    </section>
  )
}
