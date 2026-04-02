export function thumbUrl(id: number): string {
	return `/img/${id}/thumb`;
}

export function previewUrl(id: number): string {
	return `/img/${id}/preview`;
}

export function fullUrl(id: number): string {
	return `/img/${id}/full`;
}
