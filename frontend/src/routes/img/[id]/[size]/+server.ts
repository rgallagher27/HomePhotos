import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

const API_URL = env.API_URL ?? 'http://localhost:8080';
const VALID_SIZES = new Set(['thumb', 'preview', 'full']);

export const GET: RequestHandler = async ({ params, cookies }) => {
	const { id, size } = params;

	if (!VALID_SIZES.has(size)) {
		return new Response('Invalid size', { status: 400 });
	}

	const token = cookies.get('token');
	if (!token) {
		return new Response('Unauthorized', { status: 401 });
	}

	const res = await fetch(`${API_URL}/api/v1/photos/${id}/image?size=${size}`, {
		headers: { Authorization: `Bearer ${token}` }
	});

	if (!res.ok) {
		return new Response(res.statusText, { status: res.status });
	}

	return new Response(res.body, {
		headers: {
			'Content-Type': res.headers.get('Content-Type') ?? 'image/jpeg',
			'Cache-Control': res.headers.get('Cache-Control') ?? 'public, max-age=31536000, immutable'
		}
	});
};
