const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export async function fetchHealth(): Promise<{ status: string; timestamp: string }> {
	const res = await fetch(`${API_BASE}/health`);
	return res.json();
}
