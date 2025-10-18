import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, HttpErrorResponse, httpResource } from '@angular/common/http';
import { Component, computed, ElementRef, Renderer2, Signal, signal, ViewChild, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BehaviorSubject, delay, map, merge } from 'rxjs';

interface LogEntry {
  type: 'command' | 'query';
  scpi: string;
  response?: string;
  time: number;
  errors: string[];
  serverError: string;
}

interface ScpiResponse {
  response: string;
  errors: string[];
  serverError: string;
}

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [FormsModule, DatePipe, CommonModule],
})
export class App {
  public simulated = signal(false);
  public autoSystErr = signal(true);
  public wrapLog = signal(true);
  public inputText = signal('');
  public error: WritableSignal<string> = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  private history: string[] = [];
  private historyIndex = -1;
  private unsentScpiInput = "";

  public sending$ = new BehaviorSubject(false);
  public showSlowSendIndicator$ = merge(
    this.sending$.pipe(map((x) => (x ? 'sendStart' : 'sendEnd'))),
    this.sending$.pipe(
      delay(1000),
      map((x) => (x ? 'sendStartDelay' : 'sendEndDelay')),
    ),
  ).pipe(map((x) => x === 'sendStartDelay'));

  public health = httpResource.text(() => '/api/health');

  public idn = httpResource<ScpiResponse>(() => ({
    url: '/api/scpi',
    method: 'POST',
    body: { scpi: '*IDN?', simulated: this.simulated() },
  }));
  public idnFormatted = computed(() => {
    if (this.idn.hasValue()) {
      const [manufacturer, model, serial, version] = this.idn.value().response.split(',');
      if (!manufacturer || !model || !serial || !version) {
        return "";
      }
      return `
      Manufacturer: ${manufacturer}<br>
      Model: ${model}<br>
      Serial: ${serial}<br>
      Version: ${version}<br>
      `
    } else {
      return "";
    }
  });
  public idnError = this.idn.error as Signal<HttpErrorResponse | undefined>;

  @ViewChild('scpiInput') scpiInput: ElementRef<HTMLInputElement> | undefined;

  constructor(
    private http: HttpClient,
    private renderer: Renderer2
  ) {
    this.renderer.listen("window", "focus", () => {
      this.scpiInput?.nativeElement.focus();
    });
  }

  public send() {
    if (this.sending$.value) {
      return;
    }
    this.sendInternal(this.inputText());
  }

  private sendInternal(scpi: string) {
    this.sending$.next(true);
    this.inputText.set('');
    this.addToHistory(scpi);
    const time = Date.now();
    this.http
      .post<ScpiResponse>(
        '/api/scpi',
        { scpi, simulated: this.simulated(), autoSystErr: this.autoSystErr() },
        { responseType: 'json' },
      )
      .subscribe({
        next: (x) => {
          this.error.set('');
          const type = scpi.includes('?') ? 'query' : 'command';
          const response = type === 'query' ? x.response : undefined;
          this.log.update((log) => [
            ...log,
            { type, scpi, response, time, errors: x.errors, serverError: x.serverError },
          ]);
          this.sending$.next(false);
        },
        error: (x) => {
          this.error.set(x.error);
          this.sending$.next(false);
        },
      });
  }

  private addToHistory(scpi: string) {
    if (this.history[0] !== scpi) {
      this.history = [scpi, ...this.history];
    }
  }

  public arrowUp(event: Event) {
    event.preventDefault();
    if (this.history.length > this.historyIndex + 1) {
      if (this.historyIndex === -1) {
        this.unsentScpiInput = this.inputText();
      }
      this.inputText.set(this.history[++this.historyIndex]);
    }
  }

  public arrowDown(event: Event) {
    event.preventDefault();
    if (this.historyIndex === 0) {
      this.inputText.set(this.unsentScpiInput);
      this.historyIndex--;
    }
    if (this.historyIndex > 0) {
      this.inputText.set(this.history[--this.historyIndex]);
    }
  }

  public systErr() {
    this.sendInternal(':SYST:ERR?');
  }
}
