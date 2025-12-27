// Input: Card components from ui, UniversalEmptyState, icons, Motion, atoms
// Output: DashboardPage component displaying core status and overview
// Position: Main dashboard page in dashboard module

import { useAtomValue } from 'jotai'
import {
  ActivityIcon,
  AlertCircleIcon,
  ServerIcon,
  WrenchIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { coreStatusAtom } from '@/atoms/core'
import { UniversalEmptyState } from '@/components/common/universal-empty-state'
import { Card } from '@/components/ui/card'
import { Spring } from '@/lib/spring'

export function DashboardPage() {
  const coreStatus = useAtomValue(coreStatusAtom)

  if (coreStatus === 'stopped') {
    return (
      <UniversalEmptyState
        icon={ServerIcon}
        title="Core is not running"
        description="Start the mcpd core to see your dashboard and manage MCP servers."
        action={{
          label: 'Start Core',
          onClick: () => {
            // TODO: Implement start core action
          },
        }}
      />
    )
  }

  if (coreStatus === 'error') {
    return (
      <UniversalEmptyState
        icon={AlertCircleIcon}
        title="Core encountered an error"
        description="The mcpd core failed to start or encountered a runtime error. Check the logs for details."
        action={{
          label: 'View Logs',
          onClick: () => {
            // TODO: Navigate to logs page
          },
        }}
      />
    )
  }

  return (
    <div className="flex flex-1 flex-col gap-6 p-6 overflow-auto">
      {/* Status Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <m.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={Spring.smooth(0.3)}
        >
          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="flex size-12 items-center justify-center rounded-lg bg-success/10">
                <ServerIcon className="size-6 text-success" />
              </div>
              <div>
                <p className="text-muted-foreground text-sm">Core Status</p>
                <p className="font-semibold text-2xl">
                  {coreStatus === 'running' ? 'Running' : 'Starting'}
                </p>
              </div>
            </div>
          </Card>
        </m.div>

        <m.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={Spring.smooth(0.3, 0.05)}
        >
          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="flex size-12 items-center justify-center rounded-lg bg-primary/10">
                <ActivityIcon className="size-6 text-primary" />
              </div>
              <div>
                <p className="text-muted-foreground text-sm">Active Callers</p>
                <p className="font-semibold text-2xl">0</p>
              </div>
            </div>
          </Card>
        </m.div>

        <m.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={Spring.smooth(0.3, 0.1)}
        >
          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="flex size-12 items-center justify-center rounded-lg bg-info/10">
                <WrenchIcon className="size-6 text-info" />
              </div>
              <div>
                <p className="text-muted-foreground text-sm">Available Tools</p>
                <p className="font-semibold text-2xl">0</p>
              </div>
            </div>
          </Card>
        </m.div>
      </div>

      {/* Recent Activity */}
      <m.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={Spring.smooth(0.4, 0.15)}
      >
        <Card className="flex-1">
          <div className="border-border border-b p-4">
            <h2 className="font-semibold text-lg">Recent Activity</h2>
          </div>
          <div className="p-4">
            <UniversalEmptyState
              icon={ActivityIcon}
              title="No recent activity"
              description="Tool calls and system events will appear here once callers start connecting."
            />
          </div>
        </Card>
      </m.div>
    </div>
  )
}
