import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient, httpResource } from '@angular/common/http';
import {
  Component,
  computed,
  effect,
  ElementRef,
  HostListener,
  QueryList,
  Renderer2,
  Signal,
  signal,
  ViewChild,
  ViewChildren,
  WritableSignal,
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
import { Commands, HealthResponse, LogEntry, NodeInfo, ScpiNode, ScpiResponse } from './types';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';
import { AutocompleteTrigger } from './autocomplete/autocomplete-trigger';
import {
  cardinalityOf,
  childrenOf,
  findCardinalNode,
  getClipboardText,
  getShortMnemonic,
  getTimestamp,
  range,
  removeDuplicateNodes,
  stripCardinality,
} from './utils';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { SyntaxHighlightPipe } from './syntax-highlight/syntax-highlight.pipe';

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  providers: [DatePipe],
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
    AutocompleteTrigger,
    MatProgressBarModule,
    SyntaxHighlightPipe,
  ],
})
export class App {
  public inputText = signal('');
  public scriptedLog: WritableSignal<LogEntry[]> = signal([]);
  public interactiveLog: WritableSignal<LogEntry[]> = signal([]);
  public selectedLogIndices = signal<number[]>([]);
  private dragStartIndex: number | null = null;
  private isDragging = false;
  private lastSelectedAutocompletionHasSuffix = signal(false);
  private lastSelectedAutocompletionIsQuery = signal(false);
  public activeToolbarButtons: WritableSignal<string[]> = signal([]);
  public isScrolledToBottom = signal(true);
  public isScrolledToTop = signal(true);
  public forceShowLog = signal(false);

  public script: WritableSignal<string[]> = signal([]);
  public scriptSource: WritableSignal<'file' | 'clipboard'> = signal('file');
  public scriptFileName: WritableSignal<string> = signal('');
  public scriptRunning: WritableSignal<boolean> = signal(false);
  public scriptProgressPercentage: WritableSignal<number> = signal(0);
  public scriptCancelled: WritableSignal<boolean> = signal(false);

  @ViewChild('fileInput') fileInput!: ElementRef<HTMLInputElement>;
  private resolveFile?: (value: string) => void;
  private rejectFile?: (reason?: any) => void;

  public autocomplete: Signal<Array<ScpiNode | string>> = computed(() => {
    if (!this.commands.hasValue()) {
      return [];
    }
    if (this.history.index() >= 0) {
      return [];
    }
    if (this.inputText().endsWith('?') || this.inputText().endsWith(' ')) {
      return [];
    }

    if (this.inputText().startsWith('*')) {
      return this.commands.value().starTree?.children.filter((command) => {
        return command.content.text
          .slice(1)
          .toLowerCase()
          .startsWith(this.inputText().slice(1).toLowerCase());
      });
    } else {
      const inputSegments = this.inputText().split(':').slice(1);
      const initialNode = this.commands.value().colonTree;
      let currentNodes = [initialNode];
      let finishedNodes = [initialNode];
      let matched = 0;

      const compare = (node: NodeInfo, typed: string) => {
        const cardinality = cardinalityOf(typed);
        const hasCardinality = cardinality !== undefined;
        const cardinalityMismatch = () => hasCardinality && !node.suffixed;
        const cardinalityOutOfRange = () =>
          hasCardinality && node.suffixed && (cardinality < node.start || cardinality > node.stop);
        if (cardinalityMismatch() || cardinalityOutOfRange()) {
          return false;
        }

        const fullMatch = () => typed.toLowerCase() === node.text.toLowerCase();
        const shortMatch = () => typed.toLowerCase() === getShortMnemonic(node.text).toLowerCase();
        const noCardinalityFullMatch = () =>
          node.suffixed && stripCardinality(typed).toLowerCase() === node.text.toLowerCase();
        const noCardinalityShortMatch = () =>
          node.suffixed &&
          stripCardinality(typed).toLowerCase() === getShortMnemonic(node.text).toLowerCase();
        return fullMatch() || shortMatch() || noCardinalityFullMatch() || noCardinalityShortMatch();
      };

      for (const [segmentIndex, segment] of inputSegments.entries()) {
        if (segment === '') {
          continue;
        }

        if (childrenOf(finishedNodes).length === 0) {
          break;
        }

        const lastSegment = segmentIndex !== inputSegments.length - 1;

        const matchingNodes = childrenOf(finishedNodes).filter((x) => compare(x.content, segment));
        if (matchingNodes.length > 0) {
          if (lastSegment) {
            currentNodes = matchingNodes;
          }

          // If the currently typed segment doesn't have cardinality, we include matches with and without cardinality (since cardinality is optional)
          // Otherwise, when it has cardinality, we take only the last item, which based on backend sorting should always be the one with cardinality.
          const cardinality = cardinalityOf(segment);
          const hasCardinality = cardinality !== undefined;
          finishedNodes = hasCardinality
            ? [matchingNodes[matchingNodes.length - 1]]
            : matchingNodes;

          matched++;
        }
      }

      if (matched < inputSegments.length - 1) {
        // one or more completed mnemonic segments had no match, show no autocomplete options
        return [];
      }

      if (finishedNodes.length > 2) {
        console.error(
          'Unexpected, should have at most 2, with and withous suffixes',
          finishedNodes,
        );
      }

      const cardinalNode = findCardinalNode(finishedNodes);
      if (cardinalNode) {
        const currentInputSegment = inputSegments[inputSegments.length - 1];
        const currentInputFinishesNode =
          currentInputSegment !== '' &&
          (currentInputSegment === cardinalNode.content.text ||
            currentInputSegment === getShortMnemonic(cardinalNode.content.text));
        if (currentInputFinishesNode && this.lastSelectedAutocompletionHasSuffix()) {
          return range(cardinalNode.content.start, cardinalNode.content.stop).map((x) => {
            if (this.lastSelectedAutocompletionIsQuery()) {
              return `${x}?`;
            } else {
              return `${x}`;
            }
          });
        }
      }

      const result = childrenOf(currentNodes).filter((x) =>
        x.content.text
          .toLowerCase()
          .startsWith(inputSegments[inputSegments.length - 1]?.toLowerCase()),
      );
      return removeDuplicateNodes(result);
    }
  });

  public autocompleteValueTransform = (
    previous: string,
    selected: MatOption<any>,
  ): MatOption<any> => {
    if (typeof selected.value === 'string') {
      if (selected.value.endsWith('?')) {
        selected.value = previous + selected.value;
      } else {
        selected.value = previous + selected.value + ':';
      }
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

    const appendColonToUnfinishedMnemonics = (
      node: ScpiNode,
      option: MatOption<any>,
    ): MatOption<any> => {
      const noChildren = node.children && node.children.length !== 0;
      const noSuffix = !node.content.suffixed;
      if (noChildren && noSuffix) {
        option.value += ':';
      }
      return option;
    };

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

    if (this.preferences.preferShortScpi()) {
      selected.value = getShortMnemonic(selected.value);
    }

    this.lastSelectedAutocompletionHasSuffix.set(selectedScpiNode.content.suffixed);
    this.lastSelectedAutocompletionIsQuery.set(selectedScpiNode.content.text.endsWith('?'));

    if (this.lastSelectedAutocompletionHasSuffix() && this.lastSelectedAutocompletionIsQuery()) {
      selected.value = selected.value.slice(0, -1); // Remove ?, will be added back by auto-completion options
    }

    return appendColonToUnfinishedMnemonics(selectedScpiNode, selected);
  };

  public getDisplayValue = (node: ScpiNode): string => {
    let result = '';
    if (!node.content.text.startsWith('*')) {
      result += ':';
    }
    const isQuery = node.content.text.endsWith('?');
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
  };

  private unsentScpiInput = '';

  public sending$: BehaviorSubject<boolean> = new BehaviorSubject(false);
  public showSlowSendIndicator$ = combineLatest([
    this.sending$.pipe(map((x) => (x ? 'start' : 'end'))),
    this.sending$.pipe(
      delay(500),
      map((x) => (x ? 'start' : 'end')),
    ),
  ]).pipe(map(([sending, sendingDelayed]) => sending === 'start' && sendingDelayed === 'start'));

  public health = httpResource<HealthResponse>(() => '/api/health');

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
    private datePipe: DatePipe,
  ) {
    localStorageService.setFromStorage('activeToolbarButtons', this.activeToolbarButtons);
    effect(() => localStorageService.setItem('activeToolbarButtons', this.activeToolbarButtons()));

    localStorageService.setFromStorage('interactiveLog', this.interactiveLog);
    effect(() => localStorageService.setItem('interactiveLog', this.interactiveLog()));

    localStorageService.setFromStorage('scriptedLog', this.scriptedLog);
    effect(() => localStorageService.setItem('scriptedLog', this.scriptedLog()));

    localStorageService.setFromStorage('script', this.script);
    effect(() => localStorageService.setItem('script', this.script()));

    localStorageService.setFromStorage('scriptSource', this.scriptSource);
    effect(() => localStorageService.setItem('scriptSource', this.scriptSource()));

    localStorageService.setFromStorage('scriptFileName', this.scriptFileName);
    effect(() => localStorageService.setItem('scriptFileName', this.scriptFileName()));

    effect(() => {
      this.preferences.operationMode();
      this.checkScrollPosition();
    });

    effect(() => {
      this.log();
      this.checkScrollPosition();
      if (this.preferences.scrollToNewLogOutput() || this.isScrolledToBottom()) {
        setTimeout(() => { // wait to scroll to bottom until after Angular has a chance to redraw things
          this.scrollToBottom();
        })
      }
    });

    this.renderer.listen('window', 'focus', () => {
      this.scpiInput?.nativeElement.focus();
    });
  }

  public get log(): WritableSignal<LogEntry[]> {
    return this.preferences.operationMode() === 'scripted' ? this.scriptedLog : this.interactiveLog;
  }

  public scrollToBottom() {
    if (this.logContainer) {
      this.logContainer.nativeElement.scrollTop = Number.MAX_SAFE_INTEGER;
    }
  }

  public scrollToTop() {
    if (this.logContainer) {
      this.logContainer.nativeElement.scrollTop = 0;
    }
  }

  public onLogScroll() {
    this.checkScrollPosition();
  }

  @HostListener('window:resize')
  onResize() {
    this.checkScrollPosition();
  }

  private checkScrollPosition() {
    if (this.log().length === 0) {
      this.isScrolledToBottom.set(true);
      this.isScrolledToTop.set(true);
    } else {
      // Allows for 1px inaccuracy
      this.isScrolledToBottom.set(
        this.logContainer?.nativeElement.scrollHeight -
          this.logContainer?.nativeElement.clientHeight <=
          this.logContainer?.nativeElement.scrollTop + 1,
      );
      this.isScrolledToTop.set(this.logContainer?.nativeElement.scrollTop <= 1);
    }
  }

  public onEntryMouseDown(index: number, event: MouseEvent) {
    this.dragStartIndex = index;
    this.isDragging = true;
    event.preventDefault();
  }

  public onEntryMouseEnter(index: number) {
    if (this.isDragging && this.dragStartIndex !== null) {
      const start = Math.min(this.dragStartIndex, index);
      const end = Math.max(this.dragStartIndex, index);
      const selected: number[] = [];
      for (let i = start; i <= end; i++) {
        selected.push(i);
      }
      this.selectedLogIndices.set(selected);
    }
  }

  @HostListener('window:mouseup')
  onMouseUp() {
    this.isDragging = false;
    this.dragStartIndex = null;
  }

  public onEntryClick(index: number) {
    if (!this.isDragging) {
      this.selectedLogIndices.set([index]);
    }
  }

  public previousLogEntry() {
    const current = this.selectedLogIndices();
    if (current.length === 0) {
      this.selectedLogIndices.set([this.log().length - 1]);
    } else {
      const minIndex = Math.min(...current);
      if (minIndex > 0) {
        this.selectedLogIndices.set([minIndex - 1]);
      }
    }
    this.scrollToSelectedEntry();
  }

  public nextLogEntry() {
    const current = this.selectedLogIndices();
    if (current.length === 0) {
      this.selectedLogIndices.set([0]);
    } else {
      const maxIndex = Math.max(...current);
      if (maxIndex < this.log().length - 1) {
        this.selectedLogIndices.set([maxIndex + 1]);
      }
    }
    this.scrollToSelectedEntry();
  }

  private scrollToSelectedEntry() {
    const indices = this.selectedLogIndices();
    if (indices.length > 0) {
      const lastIndex = indices[indices.length - 1];
      if (this.entryElements?.get(lastIndex)) {
        this.entryElements?.get(lastIndex).nativeElement.scrollIntoView({
          behavior: 'instant',
          block: 'nearest',
        });
      }
    }
  }

  public sendInteractive() {
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
    this.sendInteractiveInternal(this.inputText());
  }

  private async sendInteractiveInternal(scpi: string): Promise<void> {
    this.history.index.set(-1);
    this.inputText.set('');

    scpi = scpi.startsWith(':') || scpi.startsWith('*') ? scpi : `:${scpi}`;
    setTimeout(() => this.history.add(scpi), 100); // Delay to avoid history dropdown updating before it has a chance to close.

    this.sending$.next(true);
    await this.sendInternal(scpi);
    this.sending$.next(false);
  }

  private async sendInternal(scpi: string): Promise<void> {
    scpi = scpi.startsWith(':') || scpi.startsWith('*') ? scpi : `:${scpi}`;

    const time = Date.now();
    const type = scpi.includes('?') ? 'query' : 'command';
    this.log.update((log) => [
      ...log,
      {
        type,
        scpi,
        response: undefined,
        time,
        elapsed: 0,
        isServerError: false,
        uniqueId: crypto.randomUUID(),
      },
    ]);
    const params = {
      simulated: this.preferences.simulated(),
      autoSystErr: this.preferences.autoSystErr(),
      timeoutSeconds: this.preferences.timeoutSeconds(),
      port: this.preferences.port(),
      address: this.preferences.address(),
    };

    return new Promise<void>((resolve) => {
      this.http.post<ScpiResponse>('/api/scpi', scpi, { params, responseType: 'json' }).subscribe({
        next: (x) => {
          const response = type === 'query' ? x.response : undefined;
          this.log.update((log) => {
            const clone = structuredClone(log); // Can't modify existing log, have to write a new one, otherwise signals don't work
            const lastElement = clone[clone.length - 1];
            lastElement.response = (response ? response : x.serverError).trim();
            lastElement.isServerError = !response;
            lastElement.elapsed = Date.now() - time;
            for (const error of x.errors ?? []) {
              clone.push({
                type: 'query',
                scpi: ':SYST:ERR?',
                response: error,
                uniqueId: crypto.randomUUID(),
                time,
                hideTime: true,
                isServerError: false,
              });
            }
            return clone;
          });
          resolve();
        },
        error: (x) => {
          this.log.update((log) => {
            const clone = structuredClone(log); // Can't modify existing log, have to write a new one, otherwise signals don't work
            const lastElement = clone[clone.length - 1];
            lastElement.response = x.error ?? x.message;
            lastElement.isServerError = true;
            lastElement.elapsed = Date.now() - time;
            return clone;
          });
          this.snackBar.open(x.error ?? x.message, 'Close', { duration: 5000 });
          resolve();
        },
      });
    });
  }

  public onInputKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowDown') {
      this.arrowDown(event);
    } else if (event.key === 'ArrowUp') {
      this.arrowUp(event);
    }

    // Not handling Enter here, for some reason it affects whether `stopImmediatePropagation` works in autocomplete-trigger

    this.trackLastSelectedAutocompletion(event.key);
    this.trackModificationsToResetHistoryIndex(event.key);
  }

  public arrowUp(event: Event) {
    event.preventDefault();

    if (this.history.index() === -1 && this.inputText() !== '') {
      return;
    }

    if (this.history.list().length > this.history.index() + 1) {
      if (this.history.index() === -1) {
        this.unsentScpiInput = this.inputText();
      }
      this.history.index.update((x) => x + 1);
      this.inputText.set(this.history.list()[this.history.index()]);
    }
  }

  public arrowDown(event: Event) {
    event.preventDefault();

    if (this.history.index() === 0) {
      this.inputText.set(this.unsentScpiInput);
      this.history.index.update((x) => x - 1);
    }

    if (this.history.index() > 0) {
      this.history.index.update((x) => x - 1);
      this.inputText.set(this.history.list()[this.history.index()]);
    }
  }

  private trackLastSelectedAutocompletion(key: string) {
    const isNumber = /^[0-9]$/.test(key);
    const ignoredKeys = [
      'ArrowUp',
      'ArrowDown',
      'ArrowLeft',
      'ArrowRight',
      'Backspace',
      'Enter',
      'Shift',
      'Control',
      'Meta',
    ];
    if (!isNumber && !ignoredKeys.includes(key)) {
      this.lastSelectedAutocompletionHasSuffix.set(false);
      this.lastSelectedAutocompletionIsQuery.set(false);
    }
  }

  private trackModificationsToResetHistoryIndex(key: string) {
    const isNumeric = /^[a-zA-Z0-9]$/.test(key);
    const isSymbol = /^[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?`~]$/.test(key);
    const watchedKeys = ['Backspace', 'Delete'];
    if (isNumeric || isSymbol || watchedKeys.includes(key)) {
      this.history.index.set(-1);
    }
  }

  public async systErr() {
    await this.sendInteractiveInternal(':SYST:ERR?');
  }

  public onHistoryEntrySelect(entry: string) {
    this.inputText.set(entry);
  }

  public insertCharacter(character: string) {
    this.inputText.set(character);
    this.scpiInput?.nativeElement.focus();
  }

  private tryGetLogText(): string | undefined {
    return this.log()
      .map((x) => this.tryGetCommandText(x))
      .join('\n');
  }

  private tryGetCommandText(entry?: LogEntry): string | undefined {
    if (!entry) {
      return undefined;
    }

    const timestamp = this.datePipe.transform(
      entry.time,
      this.preferences.showDate() ? 'MMM dd, hh:mm:ss.SS a' : 'hh:mm:ss.SS a',
    );
    return `[${timestamp}] ${entry.scpi} ${entry.response}`;
  }

  public clearLog() {
    this.log.set([]);
    this.forceShowLog.set(true);
  }

  public async copyFullLog() {
    const text = this.tryGetLogText();
    if (text) {
      await navigator.clipboard.writeText(text);
      const count = this.log().length;
      this.snackBar.open(
        `${count} ${count === 1 ? 'line' : 'lines'} copied to clipboard`,
        'Close',
        { duration: 2000 },
      );
    } else {
      this.snackBar.open('Copy log failed', 'Close', { duration: 5000 });
    }
  }

  public async copyCommands() {
    const text = this.log()
      .map((x) => x.scpi)
      .join('\n');
    if (text) {
      await navigator.clipboard.writeText(text);
      const count = this.log().length;
      this.snackBar.open(
        `${count} ${count === 1 ? 'command' : 'commands'} copied to clipboard`,
        'Close',
        { duration: 2000 },
      );
    } else {
      this.snackBar.open('Copy commands failed', 'Close', { duration: 5000 });
    }
  }

  public async sendSelectedCommands() {
    const indices = this.selectedLogIndices();
    for (const index of indices) {
      if (this.preferences.operationMode() === 'interactive') {
        await this.sendInteractiveInternal(this.log()[index].scpi)
      } else {
        this.sending$.next(true);
        await this.sendInternal(this.log()[index].scpi);
        this.sending$.next(false);
      }
    }
  }

  public async copySelectedCommands() {
    const indices = this.selectedLogIndices();
    if (indices.length === 0) {
      this.snackBar.open('No entries selected', 'Close', { duration: 2000 });
      return;
    }

    const result = indices
      .map(index => this.tryGetCommandText(this.log()[index]))
      .filter(text => text !== undefined)
      .join('\n');

    if (result) {
      await navigator.clipboard.writeText(result);
      this.snackBar.open(
        `Copied ${indices.length} ${indices.length === 1 ? 'entry' : 'entries'} to clipboard`,
        'Close',
        { duration: 2000 }
      );
    } else {
      this.snackBar.open('Copy failed', 'Close', { duration: 5000 });
    }
  }

  public async downloadFullLog() {
    const text = this.tryGetLogText();
    if (!text) {
      this.snackBar.open('Download log failed', 'Close', { duration: 5000 });
      return;
    }
    const blob = new Blob([text], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');

    link.href = url;
    link.download = `sclipi_web_log_${getTimestamp()}.txt`;
    link.click();

    window.URL.revokeObjectURL(url);
  }

  public async downloadCommands() {
    const text = this.log()
      .map((x) => x.scpi)
      .join('\n');
    if (!text) {
      this.snackBar.open('Download commands failed', 'Close', { duration: 5000 });
      return;
    }
    const blob = new Blob([text], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');

    link.href = url;
    link.download = `sclipi_web_commands_${getTimestamp()}.txt`;
    link.click();

    window.URL.revokeObjectURL(url);
  }

  public minimizeSelectedEntry() {
    const index = this.selectedLogIndices()[0]; // Only supported on single selection
    this.log.update((x) => {
      const clone = structuredClone(x);
      clone[index].minimized = true;
      return clone;
    });
  }

  public maximizeSelectedEntry() {
    const index = this.selectedLogIndices()[0]; // Only supported on single selection
    this.log.update((x) => {
      const clone = structuredClone(x);
      clone[index].minimized = false;
      return clone;
    });
  }

  public async selectScriptFromClipboard(): Promise<void> {
    try {
      const text = await getClipboardText();
      const split = text.trim().split(/\r?\n/);
      this.snackBar.open(
        `Copied ${split.length} ${split.length === 1 ? 'command' : 'commands'} from clipboard`,
        'Close',
        { duration: 2000 },
      );
      this.script.set(split);
      this.scriptSource.set('clipboard');
    } catch (err) {
      this.snackBar.open(`Failed to read clipboard: ${err}`, 'Close', { duration: 5000 });
      this.script.set([]);
    }
  }

  public async selectScriptFromFile(): Promise<void> {
    const filePromise = new Promise<string>((resolve, reject) => {
      this.resolveFile = resolve;
      this.rejectFile = reject;
    });

    this.fileInput.nativeElement.click();

    try {
      const content = await filePromise;
      const split = content.trim().split(/\r?\n/);
      this.snackBar.open(
        `Copied ${split.length} ${split.length === 1 ? 'command' : 'commands'} from file`,
        'Close',
        { duration: 2000 },
      );
      this.script.set(split);
      this.scriptSource.set('file');
    } catch (err) {
      this.snackBar.open(`Failed to read from file: ${err}`, 'Close', { duration: 5000 });
    }
  }

  public onFileSelected(event: Event): void {
    const file = (event.target as HTMLInputElement).files?.[0];

    if (!file) {
      this.rejectFile?.('No file selected');
      return;
    }

    this.scriptFileName.set(file.name);

    const reader = new FileReader();
    reader.onload = () => this.resolveFile?.(reader.result as string);
    reader.onerror = () => this.rejectFile?.(reader.error);
    reader.readAsText(file);

    this.fileInput.nativeElement.value = '';
  }

  public async runScript() {
    this.log.set([]);
    this.scriptCancelled.set(false);
    this.scriptProgressPercentage.set(0);
    this.scriptRunning.set(true);
    this.sending$.next(true);

    const percentPerCommand = 100.0 / this.script().length;
    let count = 0;

    for (const entry of this.script()) {
      if (this.scriptCancelled()) {
        this.snackBar.open(`Script was aborted early. Skipped ${this.script().length - count} commands`, 'Close', { duration: 5000 });
        break;
      }
      await this.sendInternal(entry);
      count++
      this.scriptProgressPercentage.update(x => x + percentPerCommand);
    }

    this.sending$.next(false);
    this.scriptRunning.set(false);
  }
}
