// Input: ProfileSummary array, selection state
// Output: ProfileList component with profile items
// Position: Left sidebar in config page

import type { ProfileSummary } from '@bindings/mcpd/internal/ui'
import {
  LayersIcon,
  RefreshCwIcon,
  StarIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'
import { cn } from '@/lib/utils'

interface ProfileListProps {
  profiles: ProfileSummary[]
  selectedProfile: string | null
  onSelect: (name: string) => void
  isLoading: boolean
  onRefresh: () => void
}

export function ProfileList({
  profiles,
  selectedProfile,
  onSelect,
  isLoading,
  onRefresh,
}: ProfileListProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-sm">
            <LayersIcon className="size-4" />
            Profiles
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <m.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      transition={Spring.smooth(0.3)}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2 text-sm">
              <LayersIcon className="size-4" />
              Profiles
              <Badge variant="secondary" size="sm">
                {profiles.length}
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
        <CardContent className="space-y-1">
          {profiles.length === 0 ? (
            <p className="text-muted-foreground text-sm text-center py-4">
              No profiles found
            </p>
          ) : (
            profiles.map((profile, index) => (
              <m.button
                key={profile.name}
                initial={{ opacity: 0, y: 5 }}
                animate={{ opacity: 1, y: 0 }}
                transition={Spring.smooth(0.2, index * 0.05)}
                onClick={() => onSelect(profile.name)}
                className={cn(
                  'w-full flex items-center justify-between p-2.5 rounded-md text-left transition-colors',
                  selectedProfile === profile.name
                    ? 'bg-accent text-accent-foreground'
                    : 'hover:bg-muted/50',
                )}
              >
                <div className="flex items-center gap-2 min-w-0">
                  {profile.isDefault && (
                    <StarIcon className="size-3.5 text-warning shrink-0" />
                  )}
                  <span className="font-medium text-sm truncate">
                    {profile.name}
                  </span>
                </div>
                <Badge variant="outline" size="sm">
                  {profile.serverCount} server{profile.serverCount !== 1 ? 's' : ''}
                </Badge>
              </m.button>
            ))
          )}
        </CardContent>
      </Card>
    </m.div>
  )
}
