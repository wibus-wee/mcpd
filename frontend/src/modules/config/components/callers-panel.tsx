// Input: Callers mapping, profiles list
// Output: CallersPanel component showing caller-to-profile mappings
// Position: Left sidebar in config page, below profile list

import {
  RefreshCwIcon,
  UsersIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'

interface CallersPanelProps {
  callers: Record<string, string>
  isLoading: boolean
  onRefresh: () => void
}

export function CallersPanel({
  callers,
  isLoading,
  onRefresh,
}: CallersPanelProps) {
  const callerEntries = Object.entries(callers)

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-sm">
            <UsersIcon className="size-4" />
            Callers
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {Array.from({ length: 2 }).map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <m.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      transition={Spring.smooth(0.3, 0.1)}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2 text-sm">
              <UsersIcon className="size-4" />
              Callers
              <Badge variant="secondary" size="sm">
                {callerEntries.length}
              </Badge>
            </CardTitle>
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={onRefresh}
            >
              <RefreshCwIcon className="size-3" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {callerEntries.length === 0 ? (
            <p className="text-muted-foreground text-sm text-center py-4">
              No caller mappings defined
            </p>
          ) : (
            <div className="space-y-2">
              {callerEntries.map(([caller, profile], index) => (
                <m.div
                  key={caller}
                  initial={{ opacity: 0, y: 5 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={Spring.smooth(0.2, index * 0.05)}
                  className="flex items-center justify-between p-2 rounded-md bg-muted/30"
                >
                  <span className="font-mono text-sm truncate">
                    {caller}
                  </span>
                  <Badge variant="outline" size="sm">
                    {profile}
                  </Badge>
                </m.div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </m.div>
  )
}
