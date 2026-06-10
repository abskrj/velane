import type { Monaco } from '@monaco-editor/react'

const VELANE_TYPES_URI = 'inmemory://velane/types.d.ts'

const VELANE_TYPES = `
declare module '@velane/integrations' {
  export interface IntegrationClient {
    get(endpoint: string): Promise<unknown>
    post(endpoint: string, body?: unknown): Promise<unknown>
    patch(endpoint: string, body?: unknown): Promise<unknown>
    put(endpoint: string, body?: unknown): Promise<unknown>
    delete(endpoint: string): Promise<unknown>
  }

  export function integration(provider: string, opts?: { alias?: string }): IntegrationClient
  export default integration
}

declare module '@velane/*' {
  const mod: unknown
  export default mod
}
`

let configured = false

/** Registers ambient types for built-in @velane/* imports in the Monaco editor. */
export function setupMonacoVelaneTypes(monaco: Monaco) {
  if (configured) return
  configured = true

  const ts = monaco.languages.typescript
  const compilerOptions = {
    ...ts.typescriptDefaults.getCompilerOptions(),
    allowNonTsExtensions: true,
    moduleResolution: ts.ModuleResolutionKind.NodeNext,
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ESNext,
    allowJs: true,
    checkJs: false,
  }

  ts.typescriptDefaults.setCompilerOptions(compilerOptions)
  ts.javascriptDefaults.setCompilerOptions(compilerOptions)
  ts.typescriptDefaults.addExtraLib(VELANE_TYPES, VELANE_TYPES_URI)
  ts.javascriptDefaults.addExtraLib(VELANE_TYPES, VELANE_TYPES_URI)
}
