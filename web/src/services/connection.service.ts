import { computed, effect, Injectable, signal } from '@angular/core';
import { LocalStorageService } from './localStorage.service';
import { httpResource } from '@angular/common/http';
import { PreferencesService } from './preferences.service';
import { toObservable } from '@angular/core/rxjs-interop';

@Injectable({
  providedIn: 'root'
})
export class ConnectionService {
  public desiredPerClientConnected = signal(false);
  public desiredPerClientConnected$ = toObservable(this.desiredPerClientConnected);

  public perClientConnected = computed(() => {
    return this.desiredPerClientConnected() && this.backendConfirmsConnected.hasValue() && this.backendConfirmsConnected.value();
  });

  public backendConfirmsConnected = httpResource<boolean>(() => {
    if (!this.desiredPerClientConnected() || this.preferences.perClientPort() === 0 || this.preferences.perClientAddress() === '') {
      return undefined;
    }
    return {
      url: '/api/isConnected',
      method: 'GET',
      params: {
        simulated: this.preferences.simulated(),
        port: this.preferences.perClientPort(),
        address: this.preferences.perClientAddress(),
        timeoutSeconds: this.preferences.timeoutSeconds(),
      },
    };
  });

  constructor(
    localStorageService: LocalStorageService,
    private preferences: PreferencesService
  ) {
    localStorageService.setFromStorage('desiredPerClientConnected', this.desiredPerClientConnected);
    effect(() => localStorageService.setItem('desiredPerClientConnected', this.desiredPerClientConnected()));
  }

  public perClientConnect() {
    this.desiredPerClientConnected.set(true);
    this.preferences.perClientPort.set(this.preferences.port());
    this.preferences.perClientAddress.set(this.preferences.address());
  }

  public perClientDisconnect() {
    this.desiredPerClientConnected.set(false);
  }
}
