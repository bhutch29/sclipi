import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, httpResource } from '@angular/common/http';
import {
    Component,
    computed,
    effect,
    ElementRef,
    Renderer2,
    signal,
    ViewChild,
    WritableSignal
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { BehaviorSubject, combineLatest, delay, map } from 'rxjs';
import { HistoryService } from '../services/history.service';
import { IdnService } from '../services/idn.service';
import { LocalStorageService } from '../services/localStorage.service';
import { PreferencesService } from '../services/preferences.service';
import { PreferencesComponent } from './preferences/preferences.component';
import { LogEntry, ScpiResponse } from './types';

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
    MatTooltipModule,
    MatButtonToggleModule
  ],
})
export class App {
  public inputText = signal('');
  public error: WritableSignal<string> = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  public activeToolbarButtons: WritableSignal<string[]> = signal([])

  public autocompleteHistory = computed(() =>
    this.history.list()
      .filter((x) => x.toLowerCase().includes(this.inputText().toLowerCase()))
      .filter((elem, i, self) => i === self.indexOf(elem)),
  );

  private unsentScpiInput = '';

  public sending$ = new BehaviorSubject(false);
  public showSlowSendIndicator$ = combineLatest([
    this.sending$.pipe(map((x) => (x ? 'start' : 'end'))),
    this.sending$.pipe(
      delay(1000),
      map((x) => (x ? 'start' : 'end')),
    ),
  ]).pipe(map(([sending, sendingDelayed]) => sending === 'start' && sendingDelayed === 'start'));

  public health = httpResource.text(() => '/api/health');

  @ViewChild('scpiInput') scpiInput: ElementRef<HTMLInputElement> | undefined;

  constructor(
    private http: HttpClient,
    private renderer: Renderer2,
    public preferences: PreferencesService,
    public idn: IdnService,
    public history: HistoryService,
    localStorageService: LocalStorageService,
  ) {
    localStorageService.setFromStorage('activeToolbarButtons', this.activeToolbarButtons);
    effect(() => localStorageService.setItem('activeToolbarButtons', this.activeToolbarButtons()));

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

    scpi = scpi.startsWith(':') || scpi.startsWith('*') ? scpi : `:${scpi}`;
    this.history.add(scpi);

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

  public arrowUp(event: Event) {
    event.preventDefault();
    if (this.history.list().length > this.history.index + 1) {
      if (this.history.index === -1) {
        this.unsentScpiInput = this.inputText();
      }
      this.inputText.set(this.history.list()[++this.history.index]);
    }
  }

  public arrowDown(event: Event) {
    event.preventDefault();
    if (this.history.index === 0) {
      this.inputText.set(this.unsentScpiInput);
      this.history.index--;
    }
    if (this.history.index > 0) {
      this.inputText.set(this.history.list()[--this.history.index]);
    }
  }

  public systErr() {
    this.sendInternal(':SYST:ERR?');
  }

  public onHistoryEntrySelect(entry: string) {
    this.inputText.set(entry);
    this.send();
  }
}
