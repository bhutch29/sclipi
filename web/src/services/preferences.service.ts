import { effect, Injectable, signal, WritableSignal } from '@angular/core';
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
  public operationMode: WritableSignal<'interactive' | 'scripted'> = signal('interactive')

  public simulated = signal(defaultSimulated);
  public autoSystErr = signal(defaultAutoSystErr);
  public wrapLog = signal(defaultWrapLog);
  public showDate = signal(defaultShowDate);
  public uncommittedTimeoutSeconds = signal(defaultTimeout);
  public timeoutSeconds = signal(defaultTimeout);
  public preferShortScpi = signal(defaultPreferShortScpi);
  public scrollToNewLogOutput = signal(defaultScrollToNewLogOutput);

  public uncommittedPort = signal(5025);
  public port = signal(5025);
  public perClientPort = signal(5025);
  public uncommittedAddress = signal('');
  public address = signal('');
  public perClientAddress = signal('');

  constructor(
    private localStorageService: LocalStorageService,
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
    localStorageService.setFromStorage('operationMode', this.operationMode);
    localStorageService.setFromStorage('perClientPort', this.port);
    localStorageService.setFromStorage('perClientAddress', this.address);

    effect(() => localStorageService.setItem('simulated', this.simulated()));
    effect(() => localStorageService.setItem('autoSystErr', this.autoSystErr()));
    effect(() => localStorageService.setItem('wrapLog', this.wrapLog()));
    effect(() => localStorageService.setItem('timeoutSeconds', this.timeoutSeconds()));
    effect(() => localStorageService.setItem('showDate', this.showDate()));
    effect(() => localStorageService.setItem('preferShortScpi', this.preferShortScpi()));
    effect(() => localStorageService.setItem('scrollToNewLogOutput', this.scrollToNewLogOutput()));
    effect(() => localStorageService.setItem('operationMode', this.operationMode()));
    effect(() => localStorageService.setItem('perClientPort', this.port()));
    effect(() => localStorageService.setItem('perClientAddress', this.address()));
  }

  public async loadServerPreferences(loadFromServer = false) {
    if (loadFromServer) {
      const port = await firstValueFrom(this.http.get('/api/scpiPort', { responseType: 'text' }));
      this.uncommittedPort.set(+port);
      this.port.set(+port);

      const address = await firstValueFrom(
        this.http.get('/api/scpiAddress', { responseType: 'text' }),
      );
      this.uncommittedAddress.set(address);
      this.address.set(address);
    } else {
      this.localStorageService.setFromStorage('perClientAddress', this.address)
      this.localStorageService.setFromStorage('perClientAddress', this.uncommittedAddress)
      this.localStorageService.setFromStorage('perClientAddress', this.perClientAddress)
      this.localStorageService.setFromStorage('perClientPort', this.port)
      this.localStorageService.setFromStorage('perClientPort', this.uncommittedPort)
      this.localStorageService.setFromStorage('perClientPort', this.perClientPort)
    }
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
    // Operation mode is skipped intentionally
  }

  public onPortBlur(updateServer = false) {
    this.setPort(this.uncommittedPort(), updateServer);
  }

  public onPortEnter(event: Event, updateServer = false) {
    event.preventDefault();
    this.setPort(this.uncommittedPort(), updateServer);
  }

  private setPort(port: number | null, updateServer: boolean) {
    if (port === null) {
      this.uncommittedPort.set(this.port());
      return;
    }

    if (!Number.isInteger(port)) {
      console.error('port must be an integer', port);
      this.uncommittedPort.set(this.port());
      return;
    }

    if (port < 1 || port > 65535) {
      console.error('port must be between 1 and 65535', port);
      this.uncommittedPort.set(this.port());
      return;
    }

    if (this.port() !== port) {
      this.port.set(port);
      if (updateServer && port !== 0) {
        this.http.post('/api/scpiPort', port, { responseType: 'text' }).subscribe({
          next: (x) => console.log(x),
          error: (x) => console.error('Error posting port value', this.port(), x),
        });
      }
    }
  }

  public onAddressBlur(updateServer = false) {
    this.setAddress(this.uncommittedAddress(), updateServer);
  }

  public onAddressEnter(event: Event, updateServer = false) {
    event.preventDefault();
    this.setAddress(this.uncommittedAddress(), updateServer);
  }

  private setAddress(address: string, updateServer: boolean) {
    if (this.address() != address) {
      this.address.set(address);
      if (updateServer && address !== '') {
        this.http
          .post('/api/scpiAddress', address, { responseType: 'text' })
          .subscribe({
            next: (x) => console.log(x),
            error: (x) => console.error('Error posting address value', this.address(), x),
          });
      }
    }
  }
}
