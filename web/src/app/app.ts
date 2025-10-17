import { HttpClient, httpResource } from '@angular/common/http';
import { Component, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.scss',
  imports: [FormsModule]
})
export class App {
  public temp = httpResource.text(() => '/api/health');
  public inputText = "";
  public response: WritableSignal<string> = signal("");
  public error: WritableSignal<string> = signal("");

  constructor(
    private http: HttpClient
  ) {

  }

  public send() {
    this.http.post("/api/scpi", {scpi: this.inputText, simulated: true}, {responseType: 'text'}).subscribe({
      next: x => {
        this.error.set("");
        this.response.set(x);
      },
      error: x => {
        this.error.set(x.error);
        this.response.set("");
      }
    });
  }
}
