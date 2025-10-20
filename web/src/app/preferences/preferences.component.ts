import { Component } from '@angular/core';
import { PreferencesService } from '../../services/preferences.service';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { HistoryService } from '../../services/history.service';

@Component({
  selector: 'app-preferences',
  imports: [FormsModule, MatButtonModule, MatFormFieldModule, MatInputModule, MatCheckboxModule],
  templateUrl: './preferences.component.html',
  styleUrl: './preferences.component.scss'
})
export class PreferencesComponent {
  constructor(
    public preferences: PreferencesService,
    private http: HttpClient,
    private history: HistoryService
  ) {
  }

  public onPortBlur() {
    this.setPort(this.preferences.port());
  }

  public onPortEnter(event: Event) {
    event.preventDefault();
    this.setPort(this.preferences.port());
  }

  private setPort(port: number) {
    if (!Number.isInteger(port)) {
      console.error('port must be an integer', port);
      this.preferences.port.set(this.preferences.committedPort());
      return;
    }

    if (port < 1 || port > 65535) {
      console.error('port must be between 1 and 65535', port);
      this.preferences.port.set(this.preferences.committedPort());
      return;
    }

    if (this.preferences.committedPort() != this.preferences.port()) {
      this.preferences.committedPort.set(this.preferences.port());
      if (port !== 0) {
        this.http.post('/api/scpiPort', this.preferences.committedPort(), { responseType: 'text' }).subscribe({
          next: (x) => console.log(x),
          error: (x) => console.error('Error posting port value', this.preferences.committedPort(), x),
        });
      }
    }
  }

  public onAddressBlur() {
    this.setAddress(this.preferences.address());
  }

  public onAddressEnter(event: Event) {
    event.preventDefault();
    this.setAddress(this.preferences.address());
  }

  private setAddress(address: string) {
    if (this.preferences.committedAddress() != this.preferences.address()) {
      this.preferences.committedAddress.set(this.preferences.address());
      if (address !== '') {
        this.http
          .post('/api/scpiAddress', this.preferences.committedAddress(), { responseType: 'text' })
          .subscribe({
            next: (x) => console.log(x),
            error: (x) => console.error('Error posting address value', this.preferences.committedAddress(), x),
          });
      }
    }
  }

  public clearHistory() {
    this.history.list.set([]);
  }
}
