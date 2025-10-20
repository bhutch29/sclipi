import { HttpErrorResponse, httpResource } from '@angular/common/http';
import { computed, Injectable, Signal } from '@angular/core';
import { IDN, ScpiResponse } from '../app/types';
import { PreferencesService } from './preferences.service';

@Injectable({providedIn: 'root'})
export class IdnService {
  public idn = httpResource<ScpiResponse>(() => {
    if (this.preferences.committedPort() === 0 || this.preferences.committedAddress() === '') {
      return undefined;
    }
    return {
      url: '/api/scpi',
      method: 'POST',
      body: {
        scpi: '*IDN?',
        simulated: this.preferences.simulated(),
        port: this.preferences.committedPort(),
        address: this.preferences.committedAddress(),
      },
    };
  });

  public data: Signal<IDN | undefined> = computed(() => {
    if (this.idn.hasValue()) {
      const [manufacturer, model, serial, version] = this.idn.value().response.split(',');
      if (!manufacturer || !model || !serial || !version) {
        return undefined;
      }
      return { manufacturer, model, serial, version };
    } else {
      return undefined;
    }
  });

  public formatted = computed(() => {
    const x = this.data();
    if (x) {
      return `
      Manufacturer: ${x.manufacturer}
      Model: ${x.model}
      Serial: ${x.serial}
      Version: ${x.version}
      `;
    } else {
      return '';
    }
  });


  public error = this.idn.error as Signal<HttpErrorResponse | undefined>;

  constructor(
    private preferences: PreferencesService
  ){}
}
