import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class TokenStorageService {
  private readonly TOKEN_KEY = 'jwt';
  private _isValid = signal<boolean>(this.isTokenValid());

  public readonly isValid = this._isValid.asReadonly();

  // Salva il token JWT e aggiorna lo stato di validità
  public saveToken(token: string): void {
    window.sessionStorage.setItem(this.TOKEN_KEY, token);
    this._isValid.set(this.isTokenValid());
  }

  public getToken(): string | null {
    return window.sessionStorage.getItem(this.TOKEN_KEY);
  }

  public clearToken(): void {
    window.sessionStorage.removeItem(this.TOKEN_KEY);
    this._isValid.set(false);
  }

  // TODO: Non so se come lo sto facendo ha actually senso
  public isTokenValid(): boolean {
    const token = this.getToken();
    if (!token) {
      return false;
    }

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const exp = payload.exp;
      const currentTime = Math.floor(Date.now() / 1000);
      return exp > currentTime;
    } catch {
      return false;
    }
  }
}
