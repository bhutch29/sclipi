import { httpResource } from '@angular/common/http';
import { Component, signal } from '@angular/core';

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  public readonly title = signal('sclipi-web');
  public temp = httpResource.text(() => '/api/health');
}
