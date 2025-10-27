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
