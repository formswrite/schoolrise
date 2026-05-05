<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Variant = 'primary' | 'dark' | 'light';
	type Size = 'md' | 'lg';

	type Props = {
		children: Snippet;
		variant?: Variant;
		size?: Size;
		class?: string;
		href?: string;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		onclick?: (e: MouseEvent) => void;
	};

	let {
		children,
		variant = 'primary',
		size = 'md',
		class: className,
		href,
		type = 'submit',
		disabled,
		onclick
	}: Props = $props();

	const face = $derived(
		{
			primary: 'bg-[#6439B5] text-white border-2 border-[#6439B5]',
			dark: 'bg-[#060419] text-white border-2 border-[#060419]',
			light: 'bg-white text-[#060419] border-2 border-[#060419]'
		}[variant]
	);

	const padding = $derived(
		size === 'lg'
			? 'px-10 py-[18px] text-[18px] leading-[28px]'
			: 'px-6 py-3 text-[15px] leading-[22px]'
	);

	const wrapperClass = $derived(
		cn(
			'group relative inline-flex items-center justify-center font-medium',
			disabled && 'pointer-events-none opacity-60',
			className
		)
	);
	const faceClass = $derived(
		cn(
			'relative rounded-[10px] transition-transform group-hover:-translate-y-[2px]',
			face,
			padding
		)
	);
</script>

{#if href}
	<a {href} class={wrapperClass}>
		<span class="absolute inset-0 rounded-[10px] bg-[#f092dd] border-2 border-[#0b0d2a] translate-x-[4px] translate-y-[4px]" aria-hidden="true"></span>
		<span class="absolute inset-0 rounded-[10px] bg-[#ffc8eb] border-2 border-[#0b0d2a] translate-x-[2px] translate-y-[2px]" aria-hidden="true"></span>
		<span class={faceClass}>{@render children()}</span>
	</a>
{:else}
	<button {type} {disabled} {onclick} class={wrapperClass}>
		<span class="absolute inset-0 rounded-[10px] bg-[#f092dd] border-2 border-[#0b0d2a] translate-x-[4px] translate-y-[4px]" aria-hidden="true"></span>
		<span class="absolute inset-0 rounded-[10px] bg-[#ffc8eb] border-2 border-[#0b0d2a] translate-x-[2px] translate-y-[2px]" aria-hidden="true"></span>
		<span class={faceClass}>{@render children()}</span>
	</button>
{/if}
