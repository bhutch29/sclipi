import { computed, effect, Injectable, signal, WritableSignal } from '@angular/core';
import { LocalStorageService } from './localStorage.service';

@Injectable({providedIn: 'root'})
export class HistoryService {
  public list: WritableSignal<string[]> = signal([]);
  public index = -1;

  constructor(
    localStorageService: LocalStorageService
  ){
    localStorageService.setFromStorage('history', this.list);
    effect(() => localStorageService.setItem('history', this.list()));
  }

  public add(scpi: string) {
    if (this.list()[0] !== scpi) {
      this.list.update((x) => [scpi, ...x]);
    }
  }

}
