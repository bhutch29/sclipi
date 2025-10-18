import { DatePipe } from '@angular/common';
import { HttpClient, HttpErrorResponse, httpResource } from '@angular/common/http';
import { Component, Signal, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';

interface LogEntry {
  type: 'command' | 'query',
  scpi: string,
  response?: string
  time: number
}
@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [FormsModule, DatePipe]
})
export class App {
  public temp = httpResource.text(() => '/api/health');
  public blah = undefined;
  public inputText = "";
  public error: WritableSignal<string> = signal("");
  public idn = httpResource.text(() => ({url: '/api/scpi', method: "POST", body: {scpi: "*IDN?", simulated: true}}));
  public idnError = this.idn.error as Signal<HttpErrorResponse | undefined>;
  public log: WritableSignal<LogEntry[]> = signal([]);

  constructor(
    private http: HttpClient
  ) {
  }

  public send() {
    const scpi = this.inputText;
    this.inputText = "";
    const time = Date.now()
    this.http.post("/api/scpi", {scpi, simulated: true}, {responseType: 'text'}).subscribe({
      next: x => {
        this.error.set("");
        const type = scpi.includes("?") ? 'query' : 'command';
        const response = type === 'query' ? x : undefined;
        this.log.update(log => [...log, {type, scpi, response, time}]);
      },
      error: x => {
        this.error.set(x.error);
      }
    });
  }
}
