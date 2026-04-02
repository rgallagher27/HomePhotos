<script lang="ts">
	import type { PhotoDetailResponse } from '$lib/api/gen/types.gen';

	let { photo }: { photo: PhotoDetailResponse } = $props();

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	function formatAperture(aperture: number): string {
		return `f/${aperture % 1 === 0 ? aperture.toFixed(0) : aperture.toFixed(1)}`;
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			weekday: 'short',
			month: 'long',
			day: 'numeric',
			year: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	type Field = { label: string; value: string };

	const fields: Field[] = $derived.by(() => {
		const f: Field[] = [];

		if (photo.captured_at) f.push({ label: 'Date', value: formatDate(photo.captured_at) });

		const camera = [photo.camera_make, photo.camera_model].filter(Boolean).join(' ');
		if (camera) f.push({ label: 'Camera', value: camera });
		if (photo.lens_model) f.push({ label: 'Lens', value: photo.lens_model });
		if (photo.focal_length_mm != null) f.push({ label: 'Focal Length', value: `${photo.focal_length_mm}mm` });
		if (photo.aperture != null) f.push({ label: 'Aperture', value: formatAperture(photo.aperture) });
		if (photo.shutter_speed) f.push({ label: 'Shutter Speed', value: photo.shutter_speed });
		if (photo.iso != null) f.push({ label: 'ISO', value: `${photo.iso}` });
		if (photo.width && photo.height) f.push({ label: 'Dimensions', value: `${photo.width} × ${photo.height}` });

		f.push({ label: 'File Size', value: formatFileSize(photo.file_size_bytes) });
		f.push({ label: 'Format', value: photo.format.toUpperCase() });

		if (photo.gps_latitude != null && photo.gps_longitude != null) {
			f.push({ label: 'GPS', value: `${photo.gps_latitude.toFixed(6)}, ${photo.gps_longitude.toFixed(6)}` });
		}

		return f;
	});
</script>

<div class="space-y-1">
	<h3 class="text-sm font-medium text-gray-900 mb-2">Details</h3>
	{#each fields as field}
		<div class="flex justify-between py-1 text-sm">
			<span class="text-gray-500">{field.label}</span>
			<span class="text-gray-900 text-right ml-4">{field.value}</span>
		</div>
	{/each}
</div>
