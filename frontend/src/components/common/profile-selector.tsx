// Input: Select components from ui, selectedProfileNameAtom, useProfiles hook
// Output: ProfileSelector dropdown component for switching profiles
// Position: Common component for top bar profile selection

import { useAtom } from 'jotai'
import { m } from 'motion/react'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Spring } from '@/lib/spring'
import { selectedProfileNameAtom } from '@/modules/config/atoms'
import { useProfiles } from '@/modules/config/hooks'

export function ProfileSelector() {
  const { data: profiles, isLoading } = useProfiles()
  const [selectedProfile, setSelectedProfile] = useAtom(selectedProfileNameAtom)

  // Don't render if no profiles available
  if (isLoading || !profiles || profiles.length === 0) {
    return null
  }

  // Don't render if only one profile (no need to switch)
  if (profiles.length === 1) {
    return null
  }

  return (
    <m.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={Spring.snappy(0.3)}
    >
      <Select
        value={selectedProfile ?? undefined}
        onValueChange={(value) => setSelectedProfile(value)}
      >
        <SelectTrigger className="h-8 w-40 text-xs">
          <SelectValue placeholder="Select profile..." />
        </SelectTrigger>
        <SelectContent>
          {profiles.map((profile) => (
            <SelectItem key={profile.name} value={profile.name}>
              <span className="font-medium">{profile.name}</span>
              {profile.isDefault && (
                <span className="ml-2 text-muted-foreground text-xs">
                  (default)
                </span>
              )}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </m.div>
  )
}
