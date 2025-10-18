import { Injectable, WritableSignal } from '@angular/core';

@Injectable({providedIn: 'root'})
export class LocalStorageService {
  public setFromStorage(key: string, signal: WritableSignal<any>): void {
    const stored = localStorage.getItem(key);
    if (stored) {
      if (typeof signal() === 'number') {
        signal.set(+JSON.parse(stored));
      } else {
        signal.set(JSON.parse(stored));
      }
    }
  }

  public setItem(key: string, data: any): void {
    localStorage.setItem(key, JSON.stringify(data));
  }

  public getItem(key: string): any {
    return JSON.parse(localStorage.getItem(key) || '[]');
  }

  public removeItem(key: string): void {
    localStorage.removeItem(key);
  }

  public clear() {
    localStorage.clear();
  }
}
