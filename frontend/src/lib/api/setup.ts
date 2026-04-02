import { client } from './gen/client.gen';

export function initClient(getToken: () => string | null) {
	client.setConfig({
		baseUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
		auth: () => getToken() ?? undefined
	});
}
