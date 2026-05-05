<script lang="ts">
	import type { ActionData } from './$types';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import LayeredButton from '$lib/components/layered-button.svelte';

	let { form }: { form: ActionData } = $props();
</script>

<div
	class="bg-brand-hero relative flex min-h-screen items-center justify-center overflow-hidden px-4 py-16"
>
	<div aria-hidden="true" class="pointer-events-none absolute inset-0 overflow-hidden">
		<div
			class="absolute top-1/2 left-1/2 h-[800px] w-[800px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#d4c8ff] opacity-30"
		></div>
		<div
			class="absolute top-1/2 left-1/2 h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#d4c8ff] opacity-20"
		></div>
		<div
			class="absolute top-1/2 left-1/2 h-[400px] w-[400px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#d4c8ff] opacity-15"
		></div>
	</div>

	<div class="relative w-full max-w-md text-center">
		<img src="/logo-mark.svg" alt="" class="mx-auto h-20 w-20" aria-hidden="true" />
		<h1 class="mt-6 font-display leading-[1.1] tracking-tight text-[#060419]">
			<span class="block text-[clamp(32px,5vw,48px)] font-semibold">Sign in to</span>
			<span class="highlight-chip mt-2 text-[clamp(32px,5vw,48px)]">SchoolRise</span>
		</h1>
		<p class="mt-4 text-[18px] leading-[28px] text-[#51545d]">
			The assessment platform for educators.
		</p>

		<div class="relative mt-8">
			<span
				aria-hidden="true"
				class="absolute inset-0 translate-x-[6px] translate-y-[6px] rounded-[14px] border-2 border-[#0b0d2a] bg-[#f092dd]"
			></span>
			<span
				aria-hidden="true"
				class="absolute inset-0 translate-x-[3px] translate-y-[3px] rounded-[14px] border-2 border-[#0b0d2a] bg-[#ffc8eb]"
			></span>
			<div class="relative rounded-[14px] border-2 border-[#0b0d2a] bg-white p-8 text-left">
				<form method="POST" class="space-y-5">
					<div class="space-y-2">
						<Label for="email" class="font-semibold text-[#060419]">Email</Label>
						<Input
							id="email"
							type="email"
							name="email"
							required
							autocomplete="email"
							value={form?.email ?? ''}
							class="border-2 border-[#0b0d2a] bg-white focus-visible:ring-[#6439B5]"
						/>
					</div>

					<div class="space-y-2">
						<Label for="password" class="font-semibold text-[#060419]">Password</Label>
						<Input
							id="password"
							type="password"
							name="password"
							required
							autocomplete="current-password"
							class="border-2 border-[#0b0d2a] bg-white focus-visible:ring-[#6439B5]"
						/>
					</div>

					{#if form?.error}
						<Alert variant="destructive">
							<AlertDescription>{form.error}</AlertDescription>
						</Alert>
					{/if}

					<div class="flex justify-center pt-2">
						<LayeredButton type="submit" size="lg">Sign in</LayeredButton>
					</div>
				</form>
			</div>
		</div>

		<p class="mt-6 text-[14px] text-[#51545d]">Need help? Contact your school administrator.</p>
	</div>
</div>
