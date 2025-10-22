import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, httpResource } from '@angular/common/http';
import {
    Component,
    computed,
    effect,
    ElementRef,
    QueryList,
    Renderer2,
    signal,
    ViewChild,
    ViewChildren,
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
import { Commands, LogEntry, ScpiResponse } from './types';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';

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
    MatButtonToggleModule,
    MatSnackBarModule,
    MatDividerModule,
  ],
})
export class App {
  public inputText = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  public activeToolbarButtons: WritableSignal<string[]> = signal([])

  private isScrolledToBottom = true;

  public autocomplete = computed(() => {
    if (!this.commands.hasValue()) {
      return [];
    }
    if (this.history.index() >= 0) {
      return [];
    }
    if (this.inputText().startsWith("*")) {
      return this.commands.value().starTree.children.map(x => x.content.text);
    } else {
      const inputCommands = this.inputText().split(":").slice(1);
      const initial = this.commands.value().colonTree.children;
      return initial.map(x => x.content.text).filter(x => x.toLowerCase().includes(inputCommands[0]?.toLowerCase()));
    }
  });

  private unsentScpiInput = '';

  public sending$ = new BehaviorSubject(false);
  public showSlowSendIndicator$ = combineLatest([
    this.sending$.pipe(map((x) => (x ? 'start' : 'end'))),
    this.sending$.pipe(
      delay(500),
      map((x) => (x ? 'start' : 'end')),
    ),
  ]).pipe(map(([sending, sendingDelayed]) => sending === 'start' && sendingDelayed === 'start'));

  public health = httpResource.text(() => '/api/health');

  public commands = httpResource<Commands>(() => {
    if (this.preferences.port() === 0 || this.preferences.address() === '') {
      return undefined;
    }
    return {
      url: '/api/commands',
      method: 'GET',
      params: {
        port: this.preferences.port(),
        address: this.preferences.address(),
      },
    };
  });

  @ViewChild('scpiInput') scpiInput: ElementRef<HTMLInputElement> | undefined;
  @ViewChild('logContainer') logContainer: ElementRef<any> | undefined;
  @ViewChildren('entry') public entryElements?: QueryList<any>;

  constructor(
    private http: HttpClient,
    private renderer: Renderer2,
    public preferences: PreferencesService,
    public idn: IdnService,
    public history: HistoryService,
    private snackBar: MatSnackBar,
    localStorageService: LocalStorageService,
  ) {
    localStorageService.setFromStorage('activeToolbarButtons', this.activeToolbarButtons);
    effect(() => localStorageService.setItem('activeToolbarButtons', this.activeToolbarButtons()));

    this.renderer.listen('window', 'focus', () => {
      this.scpiInput?.nativeElement.focus();
    });
  }

  ngAfterViewInit() {
    this.entryElements?.changes.subscribe(() => {
      if (this.isScrolledToBottom) {
        this.scrollToBottom();
      }
    });
  }

  private scrollToBottom() {
    if (this.logContainer) {
      this.logContainer.nativeElement.scrollTop = Number.MAX_SAFE_INTEGER;
    }
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
    this.history.index.set(-1);
  }

  private sendInternal(scpi: string) {
    this.sending$.next(true);
    this.inputText.set('');

    this.isScrolledToBottom = this.logContainer?.nativeElement.scrollHeight - this.logContainer?.nativeElement.clientHeight <= this.logContainer?.nativeElement.scrollTop + 1; // allows for 1px inaccuracy

    scpi = scpi.startsWith(':') || scpi.startsWith('*') ? scpi : `:${scpi}`;
    this.history.add(scpi);

    const time = Date.now();
    const type = scpi.includes('?') ? 'query' : 'command';
    this.log.update((log) => [
      ...log,
      { type, scpi, response: undefined, time, elapsed: 0, errors: [], serverError: "" },
    ]);
    const params = {
      simulated: this.preferences.simulated(),
      autoSystErr: this.preferences.autoSystErr(),
      timeoutSeconds: this.preferences.timeoutSeconds(),
      port: this.preferences.port(),
      address: this.preferences.address(),
    };
    this.http.post<ScpiResponse>('/api/scpi', scpi, { params, responseType: 'json' }).subscribe({
      next: (x) => {
        const response = type === 'query' ? x.response : undefined;
        this.log.update((log) => {
          const lastElement = log[log.length - 1];
          lastElement.response = response;
          lastElement.errors = x.errors;
          lastElement.serverError = x.serverError;
          lastElement.elapsed = Date.now() - time;
          return log;
        });
        this.sending$.next(false);
      },
      error: (x) => {
        this.log.update((log) => {
          const lastElement = log[log.length - 1];
          lastElement.serverError = x.error ?? x.message;
          lastElement.elapsed = Date.now() - time;
          return log;
        });
        this.snackBar.open(x.error ?? x.message, "Close", {duration: 5000});
        this.sending$.next(false);
      },
    });
  }

  // FYI: preventDefault and listening to options observables doesn't work here
  // MatAutocomplete runs before we get a chance here.
  public arrowUp(event: Event) {
    event.preventDefault();
    if (this.history.index() === -1 && this.inputText() !== "") {
      return;
    }
    if (this.history.list().length > this.history.index() + 1) {
      if (this.history.index() === -1) {
        this.unsentScpiInput = this.inputText();
      }
      this.history.index.update(x => x + 1)
      this.inputText.set(this.history.list()[this.history.index()]);
    }
  }

  public arrowDown(event: Event) {
    event.preventDefault();
    if (this.history.index() === 0) {
      this.inputText.set(this.unsentScpiInput);
      this.history.index.update(x => x - 1);
    }
    if (this.history.index() > 0) {
      this.history.index.update(x => x - 1)
      this.inputText.set(this.history.list()[this.history.index()]);
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
