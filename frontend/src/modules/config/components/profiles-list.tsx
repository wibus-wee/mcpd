// Input: ProfileSummary array, selection state
// Output: ProfilesList component - clean list view of profiles
// Position: Tab content in config page

import type { ProfileSummary } from '@bindings/mcpd/internal/ui'
import { LayersIcon, StarIcon } from 'lucide-react'
import { m } from 'motion/react'

import { ListItem } from '@/components/custom'
import { Badge } from '@/components/ui/badge'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'

interface ProfilesListProps {
  profiles: ProfileSummary[]
  selectedProfile: string | null
  onSelect: (name: string) => void
  isLoading: boolean
  onRefresh: () => void
}

function ProfilesListSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 3 }).map((_, i) => (
        <Skeleton key={i} className="h-16 w-full rounded-lg" />
      ))}
    </div>
  )
}

function ProfilesListEmpty() {
  return (
    <Empty className="py-12">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <LayersIcon className="size-5" />
        </EmptyMedia>
        <EmptyTitle>No profiles</EmptyTitle>
        <EmptyDescription>
          Create a profile in your configuration file to get started.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function ProfilesList({
  profiles,
  selectedProfile,
  onSelect,
  isLoading,
}: ProfilesListProps) {
  if (isLoading) {
    return <ProfilesListSkeleton />
  }

  if (profiles.length === 0) {
    return <ProfilesListEmpty />
  }

  return (
    <m.div
      className="space-y-1"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={Spring.smooth(0.3)}
    >
      {profiles.map((profile, index) => (
        <ListItem
          key={profile.name}
          index={index}
          selected={selectedProfile === profile.name}
          onClick={() => onSelect(profile.name)}
        >
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <div className="flex items-center gap-2">
              {profile.isDefault && (
                <StarIcon className="size-3.5 fill-warning text-warning shrink-0" />
              )}
              <span className="font-medium text-sm truncate">
                {profile.name}
              </span>
              {profile.isDefault && (
                <Badge variant="secondary" size="sm" className="shrink-0">
                  Default
                </Badge>
              )}
            </div>
            <p className="text-muted-foreground text-xs truncate">
              {profile.serverCount} server{profile.serverCount !== 1 ? 's' : ''} configured
            </p>
          </div>
        </ListItem>
      ))}
    </m.div>
  )
}
