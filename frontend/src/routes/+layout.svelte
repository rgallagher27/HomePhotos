<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200 px-6 py-4">
		<h1 class="text-xl font-semibold text-gray-900">HomePhotos</h1>
	</header>
	<main class="p-6">
		{@render children()}
	</main>
</div>
