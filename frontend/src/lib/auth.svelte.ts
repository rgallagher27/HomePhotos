import type { AuthResponse, UserResponse } from './api/gen/types.gen';
import { getAuthMe } from './api/gen/sdk.gen';

const TOKEN_KEY = 'homephotos_token';

function setTokenCookie(token: string | null) {
	if (token) {
		document.cookie = `token=${token}; path=/; SameSite=Strict`;
	} else {
		document.cookie = 'token=; path=/; SameSite=Strict; Max-Age=0';
	}
}

export class AuthState {
	user = $state<{ id: number; username: string; role: 'admin' | 'viewer' } | null>(null);
	token = $state<string | null>(null);

	get isAuthenticated() {
		return this.token !== null;
	}

	get isAdmin() {
		return this.user?.role === 'admin';
	}

	login(auth: AuthResponse) {
		this.token = auth.token;
		this.user = { id: auth.id, username: auth.username, role: auth.role };
		localStorage.setItem(TOKEN_KEY, auth.token);
		setTokenCookie(auth.token);
	}

	logout() {
		this.token = null;
		this.user = null;
		localStorage.removeItem(TOKEN_KEY);
		setTokenCookie(null);
	}

	async restore(): Promise<boolean> {
		const stored = localStorage.getItem(TOKEN_KEY);
		if (!stored) return false;

		this.token = stored;
		setTokenCookie(stored);

		const res = await getAuthMe();
		if (res.error || !res.data) {
			this.logout();
			return false;
		}

		const me = res.data as unknown as UserResponse;
		this.user = { id: me.id, username: me.username, role: me.role };
		return true;
	}
}
