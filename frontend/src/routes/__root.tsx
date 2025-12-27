// Input: TanStack Router, Devtools, SidebarProvider, AppSidebar, MainContent
// Output: Root route with sidebar layout
// Position: Root layout component for all routes

import { TanStackDevtools } from '@tanstack/react-devtools'
import { createRootRoute, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'

import { AppSidebar } from '@/components/common/app-sidebar'
import { MainContent } from '@/components/common/main-content'
import { SidebarProvider } from '@/components/ui/sidebar'
import { RootProvider } from '@/providers/root-provider'

export const Route = createRootRoute({
  component: () => (
    <RootProvider>
      <SidebarProvider defaultOpen>
        <AppSidebar />
        <MainContent>
          <Outlet />
        </MainContent>
        <TanStackDevtools
          config={{
            position: 'bottom-right',
          }}
          plugins={[
            {
              name: 'Tanstack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
          ]}
        />
      </SidebarProvider>
    </RootProvider>
  ),
})
