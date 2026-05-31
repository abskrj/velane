declare module '@nangohq/frontend' {
  interface ConnectUIOptions {
    onSuccess?: () => void
    onError?: (err: Error) => void
  }

  class Nango {
    constructor(opts: { connectSessionToken: string })
    openConnectUI(opts?: ConnectUIOptions): Promise<void>
  }

  export default Nango
}
