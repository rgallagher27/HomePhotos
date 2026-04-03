<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="flex min-h-[60vh] items-center justify-center">
	<div class="text-center">
		<h1 class="text-6xl font-bold text-gray-300">{page.status}</h1>
		<p class="mt-4 text-lg text-gray-600">{page.error?.message ?? 'Something went wrong'}</p>
		<a href="/" class="mt-6 inline-block text-sm text-blue-600 hover:underline">Back to photos</a>
	</div>
</div>
