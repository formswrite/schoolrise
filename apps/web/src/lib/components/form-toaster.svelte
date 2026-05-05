<script lang="ts">
	import { toast } from 'svelte-sonner';

	type Result = {
		success?: boolean;
		error?: string;
		message?: string;
		toast?: { type?: 'success' | 'error' | 'info' | 'warning'; message?: string };
	} | null | undefined;

	type Props = { form: Result };
	let { form }: Props = $props();
	let lastSig = '';

	$effect(() => {
		if (!form) return;
		const sig = JSON.stringify({
			s: form.success ?? null,
			e: form.error ?? null,
			t: form.toast ?? null
		});
		if (sig === lastSig) return;
		lastSig = sig;

		const f = form as NonNullable<Result>;
		if (f.toast?.message) {
			const t = f.toast.type ?? 'info';
			toast[t](f.toast.message);
			return;
		}
		if (f.success) {
			toast.success(f.message ?? 'Saved.');
		} else if (f.error) {
			toast.error(f.error);
		}
	});
</script>
