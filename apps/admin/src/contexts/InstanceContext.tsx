import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api } from '../lib/api'

interface InstanceInfo {
  cloud: boolean
  plan: string
  licenseValid: boolean
  features: string[]
}

const defaultInfo: InstanceInfo = { cloud: false, plan: 'free', licenseValid: false, features: [] }

const InstanceContext = createContext<InstanceInfo>(defaultInfo)

export function InstanceProvider({ children }: { children: ReactNode }) {
  const [info, setInfo] = useState<InstanceInfo>(defaultInfo)

  useEffect(() => {
    api.getInstanceInfo().then((res) => {
      setInfo({
        cloud: res.cloud ?? false,
        plan: res.plan ?? 'free',
        licenseValid: res.license_valid ?? false,
        features: res.features ?? [],
      })
    }).catch(() => {
      // Non-fatal — default to no licensed features
    })
  }, [])

  return <InstanceContext.Provider value={info}>{children}</InstanceContext.Provider>
}

export function useInstance(): InstanceInfo {
  return useContext(InstanceContext)
}

export function useFeature(slug: string): boolean {
  const { features } = useInstance()
  return features.includes(slug)
}
