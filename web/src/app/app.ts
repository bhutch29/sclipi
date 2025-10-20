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
import { BehaviorSubject, delay, map, merge } from 'rxjs';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatButtonModule } from '@angular/material/button';
import { LocalStorageService } from '../services/localStorage.service';
import { PreferencesService } from '../services/preferences.service';
import { PreferencesComponent } from './preferences/preferences.component';
import { MatMenuModule } from '@angular/material/menu';
import { MatIconModule } from '@angular/material/icon';
import { IDN, LogEntry, ScpiResponse } from './types';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [
    FormsModule,
    DatePipe,
    CommonModule,
    MatAutocompleteModule,
    MatInputModule,
    MatFormFieldModule,
    MatButtonModule,
    PreferencesComponent,
    MatMenuModule,
    MatIconModule,
    MatToolbarModule,
    MatTooltipModule
  ],
})
export class App {
  public inputText = signal('');
  public error: WritableSignal<string> = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  public history: WritableSignal<string[]> = signal([]);
  public historyIndex = -1;
  public autocompleteHistory = computed(() =>
    this.history()
      .filter((x) => x.toLowerCase().includes(this.inputText().toLowerCase()))
      .filter((elem, i, self) => i === self.indexOf(elem)),
  );
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
    if (this.preferences.committedPort() === 0 || this.preferences.committedAddress() === '') {
      return undefined;
    }
    return {
      url: '/api/scpi',
      method: 'POST',
      body: {
        scpi: '*IDN?',
        simulated: this.preferences.simulated(),
        port: this.preferences.committedPort(),
        address: this.preferences.committedAddress(),
      },
    };
  });

  public idnStruct: Signal<IDN | undefined> = computed(() => {
    if (this.idn.hasValue()) {
      const [manufacturer, model, serial, version] = this.idn.value().response.split(',');
      if (!manufacturer || !model || !serial || !version) {
        return undefined;
      }
      return { manufacturer, model, serial, version };
    } else {
      return undefined;
    }
  });

  public idnFormatted = computed(() => {
    const x = this.idnStruct();
    if (x) {
      return `
      Manufacturer: ${x.manufacturer}
      Model: ${x.model}
      Serial: ${x.serial}
      Version: ${x.version}
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
    public preferences: PreferencesService,
    localStorageService: LocalStorageService,
  ) {
    localStorageService.setFromStorage('history', this.history);
    effect(() => localStorageService.setItem('history', this.history()));

    this.renderer.listen('window', 'focus', () => {
      this.scpiInput?.nativeElement.focus();
    });
  }

  public send() {
    if (this.sending$.value) {
      return;
    }
    if (this.inputText().length === 0) {
      return;
    }
    if (
      this.inputText().length === 1 &&
      (this.inputText()[0] === ':' || this.inputText()[0] === '*')
    ) {
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
      simulated: this.preferences.simulated(),
      autoSystErr: this.preferences.autoSystErr(),
      timeoutSeconds: this.preferences.timeoutSeconds(),
      port: this.preferences.port(),
      address: this.preferences.address(),
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
    const entry = scpi.startsWith(':') || scpi.startsWith('*') ? scpi : `:${scpi}`;
    if (this.history()[0] !== entry) {
      this.history.update((x) => [entry, ...x]);
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

  public clearHistory() {
    this.history.set([]);
  }

  public onHistoryEntrySelect(entry: string) {
    console.log(entry);
  }
}
