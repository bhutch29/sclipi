export interface LogEntry {
  type: 'command' | 'query';
  scpi: string;
  response?: string;
  time: number;
  elapsed: number;
  errors: string[];
  serverError: string;
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
