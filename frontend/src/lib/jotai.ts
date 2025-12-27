import type { Atom, PrimitiveAtom } from 'jotai'
import { createStore, useAtom, useAtomValue, useSetAtom } from 'jotai'
import { selectAtom } from 'jotai/utils'
import { useCallback } from 'react'

export const jotaiStore = createStore()

export const createAtomAccessor = <T>(atom: PrimitiveAtom<T>) =>
  [
    () => jotaiStore.get(atom),
    (value: T) => jotaiStore.set(atom, value),
  ] as const

const options = { store: jotaiStore }
/**
 * @param atom - jotai
 * @returns - [useAtom, useSetAtom, useAtomValue, atom, jotaiStore.get, jotaiStore.set]
 */
export const createAtomHooks = <T>(atom: PrimitiveAtom<T>) =>
  [
    () => useAtom(atom, options),
    () => useSetAtom(atom, options),
    () => useAtomValue(atom, options),
    atom,
    ...createAtomAccessor(atom),
  ] as const

export const createAtomSelector = <T>(atom: Atom<T>) => {
  const useHook = <R>(selector: (a: T) => R, deps: any[] = []) =>
    useAtomValue(
      selectAtom(
        atom,
        // eslint-disable-next-line react-hooks/exhaustive-deps
        useCallback(a => selector(a as T), deps),
      ),
    )

  useHook.__atom = atom
  return useHook
}
