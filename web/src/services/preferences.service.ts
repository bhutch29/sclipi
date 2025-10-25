import { effect, Injectable, signal } from '@angular/core';
import { LocalStorageService } from './localStorage.service';
import { firstValueFrom } from 'rxjs';
import { HttpClient } from '@angular/common/http';

const defaultTimeout = 10;
const defaultSimulated = false;
const defaultAutoSystErr = true;
const defaultWrapLog = true;
const defaultShowDate = false;
const defaultPreferShortScpi = false;
const defaultScrollToNewLogOutput = true;

@Injectable({providedIn: 'root'})
export class PreferencesService {
  public simulated = signal(defaultSimulated);
  public autoSystErr = signal(defaultAutoSystErr);
  public wrapLog = signal(defaultWrapLog);
  public showDate = signal(defaultShowDate);
  public uncommittedTimeoutSeconds = signal(defaultTimeout);
  public timeoutSeconds = signal(defaultTimeout);
  public preferShortScpi = signal(defaultPreferShortScpi);
  public scrollToNewLogOutput = signal(defaultScrollToNewLogOutput);

  public uncommittedPort = signal(0);
  public port = signal(0);
  public uncommittedAddress = signal('');
  public address = signal('');

  constructor(
    localStorageService: LocalStorageService,
    private http: HttpClient
  ) {
    localStorageService.setFromStorage('simulated', this.simulated);
    localStorageService.setFromStorage('autoSystErr', this.autoSystErr);
    localStorageService.setFromStorage('wrapLog', this.wrapLog);
    localStorageService.setFromStorage('timeoutSeconds', this.uncommittedTimeoutSeconds);
    localStorageService.setFromStorage('timeoutSeconds', this.timeoutSeconds);
    localStorageService.setFromStorage('showDate', this.showDate);
    localStorageService.setFromStorage('preferShortScpi', this.preferShortScpi);
    localStorageService.setFromStorage('scrollToNewLogOutput', this.scrollToNewLogOutput);

    effect(() => localStorageService.setItem('simulated', this.simulated()));
    effect(() => localStorageService.setItem('autoSystErr', this.autoSystErr()));
    effect(() => localStorageService.setItem('wrapLog', this.wrapLog()));
    effect(() => localStorageService.setItem('timeoutSeconds', this.timeoutSeconds()));
    effect(() => localStorageService.setItem('showDate', this.showDate()));
    effect(() => localStorageService.setItem('preferShortScpi', this.preferShortScpi()));
    effect(() => localStorageService.setItem('scrollToNewLogOutput', this.scrollToNewLogOutput()));

    this.loadServerPreferences();
  }

  private async loadServerPreferences() {
    const port = await firstValueFrom(this.http.get('/api/scpiPort', { responseType: 'text' }));
    this.uncommittedPort.set(+port);
    this.port.set(+port);

    const address = await firstValueFrom(
      this.http.get('/api/scpiAddress', { responseType: 'text' }),
    );
    this.uncommittedAddress.set(address);
    this.address.set(address);
  }

  public async resetServerPreferences() {
    await firstValueFrom(this.http.delete('/api/preferences', { responseType: 'text' }));
    await this.loadServerPreferences();
  }

  public async resetClientPreferences() {
    this.simulated.set(defaultSimulated);
    this.wrapLog.set(defaultWrapLog);
    this.autoSystErr.set(defaultAutoSystErr);
    this.timeoutSeconds.set(defaultTimeout);
    this.uncommittedTimeoutSeconds.set(defaultTimeout);
    this.showDate.set(defaultShowDate);
    this.preferShortScpi.set(defaultPreferShortScpi);
    this.scrollToNewLogOutput.set(defaultScrollToNewLogOutput);
  }

}
