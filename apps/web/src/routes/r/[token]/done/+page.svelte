<script lang="ts">
	import * as Card from '$lib/components/ui/card';
</script>

<div class="bg-brand-hero flex min-h-screen items-center justify-center px-4 py-16">
	<div class="w-full max-w-lg space-y-6 text-center">
		<img src="/logo-mark.svg" alt="" class="mx-auto h-20 w-20" aria-hidden="true" />
		<h1 class="font-display text-[clamp(32px,5vw,48px)] leading-tight font-semibold text-[#060419]">
			<span class="block">Submitted</span>
			<span class="highlight-chip mt-2">Thank you</span>
		</h1>
		<Card.Root>
			<Card.Content class="py-6">
				<p class="text-sm text-muted-foreground">
					Your assessment has been recorded. You can close this tab.
				</p>
			</Card.Content>
		</Card.Root>
	</div>
</div>
