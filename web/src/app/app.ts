import { HttpClient, httpResource } from '@angular/common/http';
import { Component, signal } from '@angular/core';
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

  constructor(
    private http: HttpClient
  ) {

  }

  public send() {
    this.http.post("/api/scpi", {scpi: this.inputText, simulated: true}, {responseType: 'text'}).subscribe({
      next: x => console.log(x),
      error: x => console.error(x.error)
    });
  }
}
