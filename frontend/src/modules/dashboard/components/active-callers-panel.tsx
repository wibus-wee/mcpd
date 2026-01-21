// Input: ActiveCallers hook, motion animation
// Output: ActiveCallersPanel component showing connected callers
// Position: Dashboard visualization for active MCP callers

import {
  CircleIcon,
  MonitorIcon,
  UserIcon,
} from 'lucide-react'
import { m } from 'motion/react'
import { useMemo } from 'react'

import { useActiveCallers } from '@/hooks/use-active-callers'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { Spring } from '@/lib/spring'
import { formatRelativeTime } from '@/lib/time'

function generateColor(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 65%, 55%)`
}

function CallerAvatar({ name }: { name: string }) {
  const color = generateColor(name)
  const initial = name.charAt(0).toUpperCase()

  return (
    <div
      className="flex size-8 items-center justify-center rounded-full text-xs font-medium text-white"
      style={{ backgroundColor: color }}
    >
      {initial}
    </div>
  )
}

function CallerRow({
  caller,
  profile,
  lastHeartbeat,
  index,
}: {
  caller: string
  profile: string
  lastHeartbeat: string
  index: number
}) {
  const isRecent = useMemo(() => {
    if (!lastHeartbeat) return false
    const diff = Date.now() - new Date(lastHeartbeat).getTime()
    return diff < 30000
  }, [lastHeartbeat])

  return (
    <m.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      transition={Spring.smooth(0.3, index * 0.05)}
      className="flex items-center gap-3 rounded-lg px-2 py-2 transition-colors hover:bg-muted/50"
    >
      <CallerAvatar name={caller} />
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{caller}</span>
          <Tooltip>
            <TooltipTrigger>
              <CircleIcon
                className={`size-2 ${isRecent ? 'fill-emerald-500 text-emerald-500' : 'fill-muted text-muted'}`}
              />
            </TooltipTrigger>
            <TooltipContent>
              {isRecent ? 'Active' : 'Idle'}
            </TooltipContent>
          </Tooltip>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Badge variant="outline" size="sm">{profile}</Badge>
          {lastHeartbeat && (
            <span>{formatRelativeTime(lastHeartbeat)}</span>
          )}
        </div>
      </div>
    </m.div>
  )
}

export function ActiveCallersPanel() {
  const { data: callers, isLoading } = useActiveCallers()

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-sm font-medium">
            Active Callers
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="size-8 rounded-full" />
              <div className="space-y-1">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-3 w-16" />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    )
  }

  const activeCallers = callers ?? []

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-sm font-medium">
            Active Callers
          </CardTitle>
          {activeCallers.length > 0 && (
            <Badge variant="secondary" size="sm">
              {activeCallers.length}
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {activeCallers.length === 0 ? (
          <m.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex flex-col items-center justify-center py-6 text-center"
          >
            <UserIcon className="mb-2 size-8 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">No active callers</p>
            <p className="text-xs text-muted-foreground/60">
              Callers will appear when IDEs connect
            </p>
          </m.div>
        ) : (
          <ScrollArea className="h-48">
            <div className="space-y-1">
              {activeCallers.map((c, i) => (
                <CallerRow
                  key={`${c.caller}-${c.pid}`}
                  caller={c.caller}
                  profile={c.profile}
                  lastHeartbeat={c.lastHeartbeat}
                  index={i}
                />
              ))}
            </div>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
