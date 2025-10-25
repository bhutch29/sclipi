export interface LogEntry {
  type: 'command' | 'query';
  scpi: string;
  response?: string;
  time: number;
  elapsed?: number;
  hideTime?: boolean
  serverError?: string;
  uniqueId: string;
  minimized?: boolean;
}

export interface ScpiResponse {
  response: string;
  errors: string[];
  serverError: string;
}

export interface IDN {
  manufacturer: string;
  model: string;
  serial: string;
  version: string;
}

export interface ScpiNode {
  content: NodeInfo;
  children: ScpiNode[];
}

export interface NodeInfo {
  text: string;
  start: number;
  stop: number;
  suffixed: boolean;
}

export interface Commands {
  starTree: ScpiNode;
  colonTree: ScpiNode;
}
