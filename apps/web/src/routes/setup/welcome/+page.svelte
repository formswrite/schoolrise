<script lang="ts">
	import LayeredButton from '$lib/components/layered-button.svelte';
</script>

<div class="space-y-8 text-center">
	<h2 class="font-display leading-[1.1] tracking-tight text-[#060419]">
		<span class="block text-[clamp(36px,5vw,56px)] font-semibold">Welcome to</span>
		<span class="highlight-chip mt-2 text-[clamp(36px,5vw,56px)]">SchoolRise</span>
	</h2>
	<p class="mx-auto max-w-xl text-[20px] leading-[30px] font-bold text-[#060419]">
		This installer will walk you through first-time setup.
	</p>

	<div class="relative mx-auto max-w-xl text-left">
		<span
			aria-hidden="true"
			class="absolute inset-0 translate-x-[6px] translate-y-[6px] rounded-[14px] border-2 border-[#0b0d2a] bg-[#f092dd]"
		></span>
		<span
			aria-hidden="true"
			class="absolute inset-0 translate-x-[3px] translate-y-[3px] rounded-[14px] border-2 border-[#0b0d2a] bg-[#ffc8eb]"
		></span>
		<div class="relative rounded-[14px] border-2 border-[#0b0d2a] bg-white p-8">
			<p class="text-sm font-semibold tracking-wider text-[#51545d] uppercase">You will</p>
			<ol class="mt-4 space-y-2 text-[16px] text-[#060419]">
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">1.</span> Paste the install token printed in your container
					logs
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">2.</span> Create the initial administrator account
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">3.</span> Configure system settings (instance name, locale,
					base URL)
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">4.</span> Define your hierarchy taxonomy
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">5.</span> Optionally bulk-import schools from CSV
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">6.</span> Optionally configure integrations (Resend,
					OpenAI, S3, SMTP)
				</li>
				<li class="flex gap-3">
					<span class="font-bold text-[#6439B5]">7.</span> Review and finalize
				</li>
			</ol>
			<p class="mt-6 text-sm text-[#51545d]">
				You can leave the installer at any step and resume later — your progress is saved.
			</p>
		</div>
	</div>

	<div class="flex justify-center pt-2">
		<LayeredButton href="/setup/unlock" size="lg">Begin</LayeredButton>
	</div>
</div>
