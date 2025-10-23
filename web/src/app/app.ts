import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, httpResource } from '@angular/common/http';
import {
    Component,
    computed,
    effect,
    ElementRef,
    QueryList,
    Renderer2,
    Signal,
    signal,
    ViewChild,
    ViewChildren,
    WritableSignal
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MatAutocompleteModule, MatOption } from '@angular/material/autocomplete';
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
import { Commands, LogEntry, NodeInfo, ScpiNode, ScpiResponse } from './types';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';
import { AutocompleteTrigger } from './autocomplete/autocomplete-trigger';
import { cardinalityOf, getShortMnemonic, range, stripCardinality } from './utils';

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
    AutocompleteTrigger
  ],
})
export class App {
  public inputText = signal('');
  public log: WritableSignal<LogEntry[]> = signal([]);

  public activeToolbarButtons: WritableSignal<string[]> = signal([])

  private isScrolledToBottom = true;

  public autocomplete: Signal<Array<ScpiNode | string>> = computed(() => {
    if (!this.commands.hasValue()) {
      return [];
    }
    if (this.history.index() >= 0) {
      return [];
    }

    if (this.inputText().startsWith("*")) {
      return this.commands.value().starTree.children.filter(command => {
        return command.content.text.slice(1).toLowerCase().startsWith(this.inputText().slice(1).toLowerCase());
      });
    } else {
      const inputSegments = this.inputText().split(":").slice(1);
      const initialNode = this.commands.value().colonTree;
      let currentNodes = [initialNode];
      let finishedNode = initialNode;
      let matched = 0;

      const compare = (node: NodeInfo, typed: string) => {
        const cardinality = cardinalityOf(typed);
        const hasCardinality = cardinality !== undefined;
        const cardinalityMismatch = () => hasCardinality && !node.suffixed;
        const cardinalityOutOfRange = () => hasCardinality && node.suffixed && (cardinality < node.start || cardinality > node.stop);
        if (cardinalityMismatch() || cardinalityOutOfRange()) {
          return false;
        }

        const fullMatch = () => typed.toLowerCase() === node.text.toLowerCase();
        const shortMatch = () => typed.toLowerCase() === getShortMnemonic(node.text).toLowerCase();
        const noCardinalityMatch = () => node.suffixed && stripCardinality(typed).toLowerCase() === node.text.toLowerCase();
        return fullMatch() || shortMatch() || noCardinalityMatch();
      };

      for (const [segmentIndex, segment] of inputSegments.entries()) {
        if (segment === '') {
          continue;
        }

        if (!finishedNode.children) {
          break;
        }

        const lastSegment = segmentIndex !== inputSegments.length - 1;

        const matchingNodes = finishedNode.children.filter(x => compare(x.content, segment));
        if (matchingNodes.length > 0) {
          if (lastSegment) {
            currentNodes = matchingNodes
          }
          finishedNode = matchingNodes[0]; // TODO: explain why this is ok

          matched++;
        }
      }

      if (matched < inputSegments.length - 1) {
        // one or more completed mnemonic segments had no match, show no autocomplete options
        return [];
      }

      // TODO: handle cases where suffixed and unsuffixed nodes both exist
      // :AM{1:2}<click> should show cardinality options. :BB{1:1} works because there is no :BB

      const currentInputSegment = inputSegments[inputSegments.length - 1];
      const currentInputFinishesNode = currentInputSegment !== '' && currentInputSegment === finishedNode.content.text;
      const showSuffixes = currentInputFinishesNode && finishedNode.content.suffixed;
      if (showSuffixes) {
        return range(finishedNode.content.start, finishedNode.content.stop).map(x => `${x}`);
      }

      const allChildren: ScpiNode[] = [];
      for (const node of currentNodes) {
        if (node.children) {
          allChildren.push(...node.children);
        }
      }
      return allChildren.filter(x => x.content.text.toLowerCase().startsWith(inputSegments[inputSegments.length - 1]?.toLowerCase()));
    }
  });

  // TODO: handle selecting :AM{1:2}?, currently inserts :AM?
  private autocompleteValueTransform = (previous: string, selected: MatOption<any>): MatOption<any> => {
    if (typeof selected.value === 'string') {
      selected.value = previous + selected.value + ':';
      return selected;
    }

    // Have to capture copies of these early.
    // We modify the object further down in a hacky way by replacing the Object in `selected.value` with a string.
    const selectedScpiNode = selected.value as ScpiNode;
    const selectedText = selectedScpiNode.content.text;

    if (selectedText.startsWith('*')) {
      selected.value = selectedText;
      return selected;
    }

    if (this.preferences.preferShortScpi()) {
      selected.value = getShortMnemonic(selected.value);
    }

    const appendColonToUnfinishedMnemonics = (node: ScpiNode, option: MatOption<any>): MatOption<any> => {
      const noChildren = node.children && node.children.length !== 0;
      const noSuffix = !node.content.suffixed;
      if (noChildren && noSuffix) {
        option.value += ":";
      }
      return option;
    }

    const previousSplit = previous.split(':');

    if (previous === ':') {
      selected.value = previous + selectedText;
    } else if (previous === '*') {
      selected.value = selectedText;
      return selected;
    } else if (previousSplit.length === 1) {
      selected.value = previous + selectedText;
    } else {
      const trimmed = previousSplit.slice(0, -1).join(':') + ':';
      selected.value = trimmed + selectedText;
    }
    return appendColonToUnfinishedMnemonics(selectedScpiNode, selected);
  }

  public getDisplayValue = (node: ScpiNode): string => {
    let result = '';
    if (!node.content.text.startsWith('*')) {
      result += ':';
    }
    const isQuery = node.content.text.endsWith('?')
    if (isQuery) {
      result += node.content.text.slice(0, -1);
    } else {
      result += node.content.text;
    }
    if (node.content.suffixed) {
      if (node.content.start === node.content.stop) {
        result += `{${node.content.start}}`;
      } else {
        result += `{${node.content.start}:${node.content.stop}}`;
      }
    }
    if (isQuery) {
      result += '?';
    }
    return result;
  }

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
  @ViewChild('autocompleteRef') public autocompleteRef?: AutocompleteTrigger;

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

    if (this.autocompleteRef) {
      this.autocompleteRef.valueTransform = this.autocompleteValueTransform;
    }
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

  public insertCharacter(character: string) {
    this.inputText.set(character);
    this.scpiInput?.nativeElement.focus();
  }
}
