import { DatePipe } from '@angular/common';
import { HttpClient, HttpErrorResponse, httpResource } from '@angular/common/http';
import { Component, Signal, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';

interface LogEntry {
  type: 'command' | 'query',
  scpi: string,
  response?: string
  time: number
  errors: string[]
}

interface ScpiResponse {
  response: string,
  errors: string[]
}

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [FormsModule, DatePipe]
})
export class App {
  public simulated = signal(false);
  public autoSystErr = signal(false);
  public inputText = signal("");
  public error: WritableSignal<string> = signal("");
  public log: WritableSignal<LogEntry[]> = signal([]);

  public temp = httpResource.text(() => '/api/health');

  public idn = httpResource<ScpiResponse>(() => ({url: '/api/scpi', method: "POST", body: {scpi: "*IDN?", simulated: true}}));
  public idnError = this.idn.error as Signal<HttpErrorResponse | undefined>;

  constructor(
    private http: HttpClient
  ) {
  }

  public send() {
    const scpi = this.inputText();
    this.inputText.set("");
    const time = Date.now()
    this.http.post<ScpiResponse>("/api/scpi", {scpi, simulated: this.simulated(), autoSystErr: this.autoSystErr()}, {responseType: 'json'}).subscribe({
      next: x => {
        this.error.set("");
        const type = scpi.includes("?") ? 'query' : 'command';
        const response = type === 'query' ? x.response : undefined;
        this.log.update(log => [...log, {type, scpi, response, time, errors: x.errors}]);
      },
      error: x => {
        this.error.set(x.error);
      }
    });
  }
}
