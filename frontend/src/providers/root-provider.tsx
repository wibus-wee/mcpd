'use client'

import { Provider } from 'jotai'
import { LazyMotion, MotionConfig } from 'motion/react'
import { ThemeProvider } from 'next-themes'

import { jotaiStore } from '@/lib/jotai'
import { Spring } from '@/lib/spring'

const loadFeatures = () =>
  import('@/lib/framer-lazy-feature').then(res => res.default)

export function RootProvider({ children }: { children: React.ReactNode }) {
  return (
    <LazyMotion features={loadFeatures} strict>
      <MotionConfig transition={Spring.presets.smooth}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Provider store={jotaiStore}>
            {children}
          </Provider>
        </ThemeProvider>
      </MotionConfig>
    </LazyMotion>
  )
}
