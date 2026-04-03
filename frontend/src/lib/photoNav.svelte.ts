// Stores the current list of photo IDs for arrow-key navigation on the detail page.
// Set by the grid page before navigating; read by the detail page.
let ids = $state<number[]>([]);

export function setPhotoNav(photoIds: number[]) {
	ids = photoIds;
}

export function getPhotoNav(): number[] {
	return ids;
}

export function getAdjacentId(currentId: number, direction: 'prev' | 'next'): number | null {
	const idx = ids.indexOf(currentId);
	if (idx === -1) return null;
	const target = direction === 'prev' ? idx - 1 : idx + 1;
	if (target < 0 || target >= ids.length) return null;
	return ids[target];
}
