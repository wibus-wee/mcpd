// Input: Card, Badge, Progress, Tooltip components, dashboard data hooks, lucide icons
// Output: StatusCards component displaying core status metrics
// Position: Dashboard status overview section

import {
  ActivityIcon,
  ClockIcon,
  FileTextIcon,
  ServerIcon,
  WrenchIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useCoreState } from '@/hooks/use-core-state'
import { Spring } from '@/lib/spring'
import { formatDuration } from '@/lib/time'

import { usePrompts, useResources, useTools } from '../hooks'

interface StatCardProps {
  title: string
  value: number | string
  icon: React.ReactNode
  description?: string
  delay?: number
  loading?: boolean
}

function StatCard({ title, value, icon, description, delay = 0, loading }: StatCardProps) {
  return (
    <m.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={Spring.smooth(0.3, delay)}
    >
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-1">
          <CardTitle className="text-muted-foreground text-xs font-medium">
            {title}
          </CardTitle>
          <Tooltip>
            <TooltipTrigger>
              <span className="text-muted-foreground">{icon}</span>
            </TooltipTrigger>
            {description && (
              <TooltipContent>{description}</TooltipContent>
            )}
          </Tooltip>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-5 w-12" />
          ) : (
            <div className="text-lg font-semibold">{value}</div>
          )}
        </CardContent>
      </Card>
    </m.div>
  )
}

export function StatusCards() {
  const { coreStatus, data: coreState, isLoading } = useCoreState()
  const { tools, isLoading: toolsLoading } = useTools()
  const { resources, isLoading: resourcesLoading } = useResources()
  const { prompts, isLoading: promptsLoading } = usePrompts()

  const statusBadgeVariant = {
    running: 'success' as const,
    starting: 'warning' as const,
    stopped: 'secondary' as const,
    stopping: 'warning' as const,
    error: 'error' as const,
  }

  return (
    <div className="grid gap-2 md:grid-cols-3 lg:grid-cols-5">
      <m.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={Spring.smooth(0.3)}
      >
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-1">
            <CardTitle className="text-muted-foreground text-xs font-medium">
              Core Status
            </CardTitle>
            <ServerIcon className="size-3.5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-1.5">
              {isLoading ? (
                <Skeleton className="h-5 w-16" />
              ) : (
                <>
                  <Badge variant={statusBadgeVariant[coreStatus]} size="default" className="mt-2.5">
                    {coreStatus.charAt(0).toUpperCase() + coreStatus.slice(1)}
                  </Badge>
                  {coreStatus === 'starting' && (
                    <Progress value={null} className="h-1 w-12" />
                  )}
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </m.div>

      <m.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={Spring.smooth(0.3, 0.03)}
      >
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-1">
            <CardTitle className="text-muted-foreground text-xs font-medium">
              Uptime
            </CardTitle>
            <ClockIcon className="size-3.5 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-5 w-12" />
            ) : (
              <div className="text-lg font-semibold">
                {coreState?.uptime ? formatDuration(coreState.uptime) : '--'}
              </div>
            )}
          </CardContent>
        </Card>
      </m.div>

      <StatCard
        title="Tools"
        value={tools.length}
        icon={<WrenchIcon className="size-3.5" />}
        description="Available MCP tools"
        delay={0.06}
        loading={toolsLoading}
      />

      <StatCard
        title="Resources"
        value={resources.length}
        icon={<FileTextIcon className="size-3.5" />}
        description="Available resources"
        delay={0.09}
        loading={resourcesLoading}
      />

      <StatCard
        title="Prompts"
        value={prompts.length}
        icon={<ActivityIcon className="size-3.5" />}
        description="Available prompt templates"
        delay={0.12}
        loading={promptsLoading}
      />
    </div>
  )
}
