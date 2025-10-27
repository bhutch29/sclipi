import { Component, input } from '@angular/core';
import { PreferencesService } from '../../services/preferences.service';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { HistoryService } from '../../services/history.service';
import { ConnectionMode } from '../types';

@Component({
  selector: 'app-preferences',
  imports: [FormsModule, MatButtonModule, MatFormFieldModule, MatInputModule, MatCheckboxModule],
  templateUrl: './preferences.component.html',
  styleUrl: './preferences.component.scss'
})
export class PreferencesComponent {
  public connectionMode = input<ConnectionMode | undefined>()

  constructor(
    public preferences: PreferencesService,
    private http: HttpClient,
    private history: HistoryService
  ) {
  }

  public onPortBlur() {
    this.setPort(this.preferences.uncommittedPort());
  }

  public onPortEnter(event: Event) {
    event.preventDefault();
    this.setPort(this.preferences.uncommittedPort());
  }

  private setPort(port: number | null) {
    if (port === null) {
      this.preferences.uncommittedPort.set(this.preferences.port());
      return;
    }

    if (!Number.isInteger(port)) {
      console.error('port must be an integer', port);
      this.preferences.uncommittedPort.set(this.preferences.port());
      return;
    }

    if (port < 1 || port > 65535) {
      console.error('port must be between 1 and 65535', port);
      this.preferences.uncommittedPort.set(this.preferences.port());
      return;
    }

    if (this.preferences.port() != port) {
      this.preferences.port.set(port);
      if (port !== 0) {
        this.http.post('/api/scpiPort', port, { responseType: 'text' }).subscribe({
          next: (x) => console.log(x),
          error: (x) => console.error('Error posting port value', this.preferences.port(), x),
        });
      }
    }
  }

  public onAddressBlur() {
    this.setAddress(this.preferences.uncommittedAddress());
  }

  public onAddressEnter(event: Event) {
    event.preventDefault();
    this.setAddress(this.preferences.uncommittedAddress());
  }

  private setAddress(address: string) {
    if (this.preferences.address() != address) {
      this.preferences.address.set(address);
      if (address !== '') {
        this.http
          .post('/api/scpiAddress', address, { responseType: 'text' })
          .subscribe({
            next: (x) => console.log(x),
            error: (x) => console.error('Error posting address value', this.preferences.address(), x),
          });
      }
    }
  }

  public onTimeoutBlur() {
    this.setTimeout(this.preferences.uncommittedTimeoutSeconds());
  }

  public onTimeoutEnter(event: Event) {
    event.preventDefault();
    this.setTimeout(this.preferences.uncommittedTimeoutSeconds());
  }

  private setTimeout(timeout: number | null) {
    if (timeout === null) {
      this.preferences.uncommittedTimeoutSeconds.set(this.preferences.timeoutSeconds());
      return;
    }

    if (!Number.isInteger(timeout)) {
      console.error('timeout must be an integer', timeout);
      this.preferences.uncommittedTimeoutSeconds.set(this.preferences.timeoutSeconds());
      return;
    }

    if (timeout <= 0) {
      this.preferences.uncommittedTimeoutSeconds.set(this.preferences.timeoutSeconds());
      return;
    }

    this.preferences.timeoutSeconds.set(this.preferences.uncommittedTimeoutSeconds());
  }

  public clearHistory() {
    this.history.list.set([]);
  }
}
