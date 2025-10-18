import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, HttpErrorResponse, httpResource } from '@angular/common/http';
import {
  Component,
  computed,
  effect,
  ElementRef,
  Renderer2,
  Signal,
  signal,
  ViewChild,
  WritableSignal,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BehaviorSubject, delay, firstValueFrom, map, merge } from 'rxjs';
import { LocalStorageService } from './localStorage.service';

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

const defaultTimeout = 10;
const defaultSimulated = false;
const defaultAutoSystErr = true;
const defaultWrapLog = true;

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [FormsModule, DatePipe, CommonModule],
})
export class App {
  public simulated = signal(defaultSimulated);
  public autoSystErr = signal(defaultAutoSystErr);
  public wrapLog = signal(defaultWrapLog);
  public timeoutSeconds = signal(defaultTimeout);

  public port = signal(0);
  private committedPort = signal(0);
  public address = signal("");
  private committedAddress = signal("");

  public inputText = signal('');
  public error: WritableSignal<string> = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  private history: WritableSignal<string[]> = signal([]);
  private historyIndex = -1;
  private unsentScpiInput = '';

  public sending$ = new BehaviorSubject(false);
  public showSlowSendIndicator$ = merge(
    this.sending$.pipe(map((x) => (x ? 'sendStart' : 'sendEnd'))),
    this.sending$.pipe(
      delay(1000),
      map((x) => (x ? 'sendStartDelay' : 'sendEndDelay')),
    ),
  ).pipe(map((x) => x === 'sendStartDelay'));

  public health = httpResource.text(() => '/api/health');

  public idn = httpResource<ScpiResponse>(() => {
    if (this.committedPort() === 0 || this.committedAddress() === '') {
      return undefined;
    }
    return {
      url: '/api/scpi',
      method: 'POST',
      body: { scpi: '*IDN?', simulated: this.simulated(), port: this.committedPort(), address: this.committedAddress() },
    }
  });
  public idnFormatted = computed(() => {
    if (this.idn.hasValue()) {
      const [manufacturer, model, serial, version] = this.idn.value().response.split(',');
      if (!manufacturer || !model || !serial || !version) {
        return '';
      }
      return `
      Manufacturer: ${manufacturer}<br>
      Model: ${model}<br>
      Serial: ${serial}<br>
      Version: ${version}<br>
      `;
    } else {
      return '';
    }
  });
  public idnError = this.idn.error as Signal<HttpErrorResponse | undefined>;

  @ViewChild('scpiInput') scpiInput: ElementRef<HTMLInputElement> | undefined;

  constructor(
    private http: HttpClient,
    private renderer: Renderer2,
    localStorageService: LocalStorageService,
  ) {
    localStorageService.setFromStorage('simulated', this.simulated);
    localStorageService.setFromStorage('autoSystErr', this.autoSystErr);
    localStorageService.setFromStorage('wrapLog', this.wrapLog);
    localStorageService.setFromStorage('history', this.history);
    localStorageService.setFromStorage('timeoutSeconds', this.timeoutSeconds);

    effect(() => localStorageService.setItem('simulated', this.simulated()));
    effect(() => localStorageService.setItem('autoSystErr', this.autoSystErr()));
    effect(() => localStorageService.setItem('wrapLog', this.wrapLog()));
    effect(() => localStorageService.setItem('history', this.history()));
    effect(() => localStorageService.setItem('timeoutSeconds', this.timeoutSeconds()));

    this.loadPreferences();

    this.renderer.listen('window', 'focus', () => {
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
    const body = {
      scpi,
      simulated: this.simulated(),
      autoSystErr: this.autoSystErr(),
      timeoutSeconds: this.timeoutSeconds(),
      port: this.port()
    };
    this.http.post<ScpiResponse>('/api/scpi', body, { responseType: 'json' }).subscribe({
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
    if (this.history()[0] !== scpi) {
      this.history.update((x) => [scpi, ...x]);
    }
  }

  public arrowUp(event: Event) {
    event.preventDefault();
    if (this.history().length > this.historyIndex + 1) {
      if (this.historyIndex === -1) {
        this.unsentScpiInput = this.inputText();
      }
      this.inputText.set(this.history()[++this.historyIndex]);
    }
  }

  public arrowDown(event: Event) {
    event.preventDefault();
    if (this.historyIndex === 0) {
      this.inputText.set(this.unsentScpiInput);
      this.historyIndex--;
    }
    if (this.historyIndex > 0) {
      this.inputText.set(this.history()[--this.historyIndex]);
    }
  }

  public systErr() {
    this.sendInternal(':SYST:ERR?');
  }

  public onPortBlur() {
    this.setPort(this.port())
  }

  public onPortEnter(event: Event) {
    event.preventDefault();
    this.setPort(this.port())
  }

  private setPort(port: number) {
    if (!Number.isInteger(port)) {
      console.error('port must be an integer', port);
      this.port.set(this.committedPort());
      return;
    }

    if (port < 1 || port > 65535) {
      console.error('port must be between 1 and 65535', port);
      this.port.set(this.committedPort());
      return;
    }

    if (this.committedPort() != this.port()) {
      this.committedPort.set(this.port());
      if (port !== 0) {
        this.http.post('/api/scpiPort', this.committedPort(), {responseType: 'text'}).subscribe({
          next: x => console.log(x),
          error: x => console.error('Error posting port value', this.committedPort(), x)
        });
      }
    }
  }

  public onAddressBlur() {
    this.setAddress(this.address())
  }

  public onAddressEnter(event: Event) {
    event.preventDefault();
    this.setAddress(this.address())
  }

  private setAddress(address: string) {
    if (this.committedAddress() != this.address()) {
      this.committedAddress.set(this.address());
      if (address !== '') {
        this.http.post('/api/scpiAddress', this.committedAddress(), {responseType: 'text'}).subscribe({
          next: x => console.log(x),
          error: x => console.error('Error posting address value', this.committedAddress(), x)
        });
      }
    }
  }

  private async loadPreferences() {
    const port = await firstValueFrom(this.http.get('/api/scpiPort', {responseType: 'text'}))
    this.port.set(+port);
    this.committedPort.set(+port);

    const address = await firstValueFrom(this.http.get('/api/scpiAddress', {responseType: 'text'}));
    this.address.set(address);
    this.committedAddress.set(address);
  }

  public async resetPreferences() {
    await firstValueFrom(this.http.delete('/api/preferences', {responseType: 'text'}));
    await this.loadPreferences();
    this.simulated.set(defaultSimulated);
    this.wrapLog.set(defaultWrapLog);
    this.autoSystErr.set(defaultAutoSystErr);
    this.timeoutSeconds.set(defaultTimeout);
  }
}
