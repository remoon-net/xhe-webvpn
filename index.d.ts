export interface Config {
  Link?: string[];
  Key: string;
  // Port: number;
  Tun?: string;
  Route?: string[];
  ICE?: string[];
  Peer?: Peer[];
  signal?: AbortSignal;
  WASM?: string;
}

export interface Peer {
  Pubkey: string;
  PSK?: string;
  Allow?: string[];
  Auto?: number;
  WHIP?: string[];
  ICE?: string[];
}

interface Network {
  listen(port: string, handler: Handler): Promise<Server>;
}

interface Handler {
  fetch: (req: Request) => Response;
  signal?: AbortSignal;
}

interface Server {
  close(): void;
}

export function connect(config: Config): Promise<Network>;

export as namespace VPN;

declare global {
  const vpn: VPN;
}
