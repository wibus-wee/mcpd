// Input: Sidebar components from ui/sidebar, icons from lucide-react, navigation atoms, router
// Output: AppSidebar component with navigation menu
// Position: App-specific sidebar for main layout

import { useAtom } from 'jotai'
import {
  LayoutDashboardIcon,
  ScrollTextIcon,
  SettingsIcon,
  WrenchIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import type { PageId } from '@/atoms/navigation'
import { activePageAtom } from '@/atoms/navigation'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { Spring } from '@/lib/spring'
import { cn } from '@/lib/utils'

interface NavItem {
  id: PageId
  label: string
  icon: typeof LayoutDashboardIcon
}

const navItems: NavItem[] = [
  {
    id: 'dashboard',
    label: 'Dashboard',
    icon: LayoutDashboardIcon,
  },
  {
    id: 'tools',
    label: 'Tools',
    icon: WrenchIcon,
  },
  {
    id: 'logs',
    label: 'Logs',
    icon: ScrollTextIcon,
  },
  {
    id: 'settings',
    label: 'Settings',
    icon: SettingsIcon,
  },
]

export function AppSidebar() {
  const [activePage, setActivePage] = useAtom(activePageAtom)

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader className="">
        <div className="flex h-10 items-center justify-center px-2" />
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item, index) => {
                const Icon = item.icon
                const isActive = activePage === item.id

                return (
                  <SidebarMenuItem key={item.id}>
                    <m.div
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={Spring.smooth(0.3, index * 0.05)}
                    >
                      <SidebarMenuButton
                        isActive={isActive}
                        onClick={() => setActivePage(item.id)}
                        tooltip={item.label}
                      >
                        <Icon className={cn(
                          'transition-colors',
                          isActive && 'text-sidebar-accent-foreground',
                        )}
                        />
                        <span>{item.label}</span>
                      </SidebarMenuButton>
                    </m.div>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-sidebar-border border-t">
        <m.div
          className="p-2 text-center text-muted-foreground text-xs group-data-[collapsible=icon]:hidden"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={Spring.smooth(0.4)}
        >
          {/* 评估版本 */}
          mcpd © 2025. All rights reserved.
        </m.div>
      </SidebarFooter>
    </Sidebar>
  )
}
