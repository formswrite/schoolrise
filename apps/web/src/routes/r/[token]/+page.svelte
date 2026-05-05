<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import LayeredButton from '$lib/components/layered-button.svelte';
	import { computeVisibleQuestions, type Answers, type LogicRule } from '$lib/forms/logic';
	import type { Question } from '$lib/forms/field-types';
	import katex from 'katex';
	import 'katex/dist/katex.min.css';
	import FileUploadInput from '$lib/components/forms/file-upload-input.svelte';
	import SignaturePad from '$lib/components/forms/signature-pad.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	function isMultiChoice(t: string) {
		return ['MULTIPLE_CHOICE', 'RADIO', 'YES_NO'].includes(t);
	}
	function isCheckboxList(t: string) {
		return t === 'CHECKBOX' || t === 'DROPDOWN';
	}

	function asStringArray(v: unknown, fallback: string[]): string[] {
		if (!Array.isArray(v)) return fallback;
		return v.filter((x): x is string => typeof x === 'string');
	}

	function renderKatex(latex: string): string {
		try {
			return katex.renderToString(latex, { throwOnError: false, displayMode: false });
		} catch {
			return latex;
		}
	}

	function fillBlanksParts(template: string): string[] {
		return template.split(/\[\[\d+\]\]/);
	}

	let answers = $state<Answers>({});

	function logicRules(): LogicRule[] {
		const settings = (data.version?.snapshot?.settings ?? {}) as Record<string, unknown>;
		const raw = settings.logic_rules;
		if (!Array.isArray(raw)) return [];
		return raw as LogicRule[];
	}

	function questions(): Question[] {
		return (data.version?.snapshot?.questions ?? []) as Question[];
	}

	const visibleQuestions = $derived(
		data.version ? computeVisibleQuestions(questions(), logicRules(), answers) : []
	);
	const submittable = $derived(
		visibleQuestions.filter((q) => !['SECTION', 'STATEMENT'].includes(q.type))
	);

	function setAnswer(cid: string, value: string | string[]) {
		answers[cid] = value;
	}
</script>

<div class="bg-brand-hero flex min-h-screen items-start justify-center px-4 py-16">
	<div class="w-full max-w-2xl space-y-6">
		<header class="text-center">
			<img src="/logo-mark.svg" alt="" class="mx-auto h-16 w-16" aria-hidden="true" />
			<h1
				class="mt-4 font-display text-[clamp(28px,4vw,40px)] leading-tight font-semibold tracking-tight text-[#060419]"
			>
				SchoolRise assessment
			</h1>
		</header>

		{#if !data.found}
			<Card.Root>
				<Card.Content class="py-12 text-center">
					<p class="text-lg font-semibold text-destructive">Invalid link</p>
					<p class="mt-2 text-sm text-muted-foreground">
						This assessment link is not recognized. Please check with your teacher.
					</p>
				</Card.Content>
			</Card.Root>
		{:else if data.alreadySubmitted}
			<Card.Root>
				<Card.Content class="py-12 text-center">
					<p class="text-lg font-semibold text-[#6439B5]">Already submitted</p>
					<p class="mt-2 text-sm text-muted-foreground">
						You've already completed <strong>{data.lookup.campaign.title}</strong>.
					</p>
				</Card.Content>
			</Card.Root>
		{:else if data.version}
			{@const v = data.version}
			<Card.Root>
				<Card.Header>
					<Card.Title>{v.title}</Card.Title>
					{#if v.description}<Card.Description>{v.description}</Card.Description>{/if}
				</Card.Header>
				<Card.Content>
					<form method="POST" action="?/submit" class="space-y-6">
						<input type="hidden" name="_total_questions" value={submittable.length} />
						{#each visibleQuestions as q, i (q.client_id)}
							{#if q.type === 'SECTION'}
								<h2 class="border-b-2 border-[#0b0d2a] pb-2 text-xl font-bold text-[#060419]">
									{q.title}
								</h2>
							{:else if q.type === 'STATEMENT'}
								<p class="rounded-md bg-[#ffc8eb]/40 p-3 text-sm text-[#060419]">{q.title}</p>
							{:else}
								<div class="space-y-2">
									<Label for="q_{q.client_id}" class="font-semibold">
										{i + 1}. {q.title}
										{#if q.required}<span class="text-destructive">*</span>{/if}
									</Label>
									{#if q.description}<p class="text-xs text-muted-foreground">
											{q.description}
										</p>{/if}

									{#if q.type === 'PARAGRAPH' || q.type === 'ESSAY'}
										<textarea
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											rows="4"
											required={q.required}
											class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
										></textarea>
									{:else if q.type === 'NUMBER' || q.type === 'DECIMAL'}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											type="number"
											step={q.type === 'DECIMAL' ? '0.01' : '1'}
											required={q.required}
										/>
									{:else if q.type === 'EMAIL'}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											type="email"
											required={q.required}
										/>
									{:else if q.type === 'PHONE' || q.type === 'HOME_NUMBER'}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											type="tel"
											required={q.required}
										/>
									{:else if q.type === 'DATE'}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											type="date"
											required={q.required}
										/>
									{:else if q.type === 'TIME'}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											type="time"
											required={q.required}
										/>
									{:else if q.type === 'LINEAR_SCALE' || q.type === 'RATING'}
										<div class="flex items-center gap-3">
											<span class="text-xs text-muted-foreground">{q.scale_min ?? 1}</span>
											<input
												id="q_{q.client_id}"
												name="q_{q.client_id}"
												type="range"
												min={q.scale_min ?? 1}
												max={q.scale_max ?? 5}
												value={Math.floor(((q.scale_min ?? 1) + (q.scale_max ?? 5)) / 2)}
												class="w-full"
											/>
											<span class="text-xs text-muted-foreground">{q.scale_max ?? 5}</span>
										</div>
									{:else if isMultiChoice(q.type)}
										<div class="space-y-2">
											{#if q.type === 'YES_NO'}
												<label class="flex items-center gap-2 text-sm">
													<input
														type="radio"
														name="q_{q.client_id}"
														value="yes"
														required={q.required}
														onchange={(e) =>
															setAnswer(q.client_id, (e.target as HTMLInputElement).value)}
													/> Yes
												</label>
												<label class="flex items-center gap-2 text-sm">
													<input
														type="radio"
														name="q_{q.client_id}"
														value="no"
														required={q.required}
														onchange={(e) =>
															setAnswer(q.client_id, (e.target as HTMLInputElement).value)}
													/> No
												</label>
											{:else}
												{#each q.options as { label?: string; value?: string }[] as opt, oi}
													<label class="flex items-center gap-2 text-sm">
														<input
															type="radio"
															name="q_{q.client_id}"
															value={opt.value ?? opt.label ?? `opt_${oi}`}
															required={q.required}
															onchange={(e) =>
																setAnswer(q.client_id, (e.target as HTMLInputElement).value)}
														/>
														{opt.label ?? opt.value ?? `Option ${oi + 1}`}
													</label>
												{/each}
											{/if}
										</div>
									{:else if isCheckboxList(q.type)}
										<select
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											required={q.required}
											class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
										>
											<option value=""></option>
											{#each q.options as { label?: string; value?: string }[] as opt, oi}
												<option value={opt.value ?? opt.label ?? `opt_${oi}`}
													>{opt.label ?? opt.value ?? `Option ${oi + 1}`}</option
												>
											{/each}
										</select>
									{:else if q.type === 'COUNTRY_REGION'}
										{@const regions = asStringArray((q.extra as { regions?: unknown })?.regions, [
											'Boké',
											'Conakry',
											'Faranah',
											'Kankan',
											'Kindia',
											'Labé',
											'Mamou',
											'Nzérékoré'
										])}
										<div class="grid grid-cols-2 gap-2">
											<select
												disabled
												class="h-9 rounded-md border border-input bg-muted/30 px-2 text-sm"
											>
												<option>Guinée</option>
											</select>
											<select
												name="q_{q.client_id}"
												required={q.required}
												class="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
											>
												<option value="">— région —</option>
												{#each regions as r}<option value={r}>{r}</option>{/each}
											</select>
										</div>
									{:else if q.type === 'ADDRESS'}
										{@const fields = asStringArray((q.extra as { fields?: unknown })?.fields, [
											'Quartier',
											'Commune',
											'Préfecture'
										])}
										<div class="space-y-2">
											{#each fields as f, fi}
												<Input name="q_{q.client_id}_{fi}" placeholder={f} required={q.required} />
											{/each}
										</div>
									{:else if q.type === 'TABLE'}
										{@const rows = asStringArray((q.extra as { rows?: unknown })?.rows, [])}
										{@const cols = asStringArray((q.extra as { columns?: unknown })?.columns, [])}
										<table class="w-full border text-sm">
											<thead>
												<tr>
													<th class="border bg-muted/30 px-2 py-1"></th>
													{#each cols as c}
														<th class="border bg-muted/30 px-2 py-1 font-medium">{c}</th>
													{/each}
												</tr>
											</thead>
											<tbody>
												{#each rows as r, ri}
													<tr>
														<td class="border bg-muted/30 px-2 py-1 font-medium">{r}</td>
														{#each cols as _, ci}
															<td class="border p-1">
																<input
																	name="q_{q.client_id}_{ri}_{ci}"
																	class="w-full rounded border border-input bg-transparent px-1 py-0.5 text-sm"
																/>
															</td>
														{/each}
													</tr>
												{/each}
											</tbody>
										</table>
									{:else if q.type === 'ORDERING'}
										{@const items = (q.options as Array<{ label?: string; value?: string }>) ?? []}
										<ol class="space-y-1">
											{#each items as opt, oi}
												<li
													class="flex items-center gap-2 rounded-md border bg-muted/20 px-2 py-1 text-sm"
												>
													<input
														type="number"
														name="q_{q.client_id}_pos_{oi}"
														min="1"
														max={items.length}
														placeholder={`${oi + 1}`}
														class="w-12 rounded border border-input bg-transparent px-1 py-0.5 text-sm"
													/>
													<span>{opt.label ?? opt.value ?? `Item ${oi + 1}`}</span>
												</li>
											{/each}
										</ol>
									{:else if q.type === 'MATCHING'}
										{@const pairs =
											(q.extra as { pairs?: Array<{ left?: string; right?: string }> })?.pairs ??
											[]}
										{@const rights = pairs.map((p, i) => p.right ?? `R${i + 1}`)}
										<div class="space-y-1">
											{#each pairs as p, pi}
												<div class="grid grid-cols-2 items-center gap-2">
													<div class="rounded-md border bg-muted/20 px-2 py-1 text-sm">
														{p.left ?? `L${pi + 1}`}
													</div>
													<select
														name="q_{q.client_id}_match_{pi}"
														required={q.required}
														class="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
													>
														<option value="">— choisir —</option>
														{#each rights as r}<option value={r}>{r}</option>{/each}
													</select>
												</div>
											{/each}
										</div>
									{:else if q.type === 'FILL_IN_BLANK'}
										{@const tpl =
											typeof (q.extra as { template?: string })?.template === 'string'
												? (q.extra as { template?: string }).template!
												: q.title}
										{@const parts = fillBlanksParts(tpl)}
										<p class="text-sm leading-relaxed">
											{#each parts as part, pi}
												<span>{part}</span>
												{#if pi < parts.length - 1}
													<input
														name="q_{q.client_id}_blank_{pi}"
														required={q.required}
														class="mx-1 inline-block w-32 border-b-2 border-foreground/40 bg-transparent text-center focus:border-[#6439B5] focus:outline-none"
													/>
												{/if}
											{/each}
										</p>
									{:else if q.type === 'EQUATION'}
										{@const latex = ((q.extra as { latex?: string })?.latex ?? '') as string}
										<div class="space-y-2">
											{#if latex}
												<div class="rounded-md border bg-muted/20 px-3 py-2">
													{@html renderKatex(latex)}
												</div>
											{/if}
											<Input name="q_{q.client_id}" required={q.required} placeholder="Réponse" />
										</div>
									{:else if q.type === 'CODE_BLOCK'}
										<textarea
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											rows="6"
											required={q.required}
											placeholder={`// ${(q.extra as { language?: string })?.language ?? 'code'}`}
											class="w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-xs shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
										></textarea>
									{:else if q.type === 'FILE_UPLOAD' || q.type === 'ATTACHMENT' || q.type === 'IMAGE'}
										<FileUploadInput
											name="q_{q.client_id}"
											required={q.required}
											accept={q.type === 'IMAGE' ? 'image/*' : undefined}
											label={q.type === 'IMAGE' ? 'Choose an image' : 'Choose a file'}
										/>
									{:else if q.type === 'SIGNATURE'}
										<SignaturePad name="q_{q.client_id}" required={q.required} />
									{:else if q.type === 'HOTSPOT'}
										<input
											name="q_{q.client_id}"
											required={q.required}
											placeholder="x,y coordinates (placeholder)"
											class="block w-full rounded-md border border-input bg-muted/20 px-3 py-2 text-sm"
										/>
									{:else}
										<Input
											id="q_{q.client_id}"
											name="q_{q.client_id}"
											required={q.required}
											placeholder={`(${q.type})`}
										/>
									{/if}
								</div>
							{/if}
						{/each}

						{#if form?.error}
							<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>
						{/if}

						<div class="flex justify-center pt-2">
							<LayeredButton size="lg">Submit</LayeredButton>
						</div>
					</form>
				</Card.Content>
			</Card.Root>
		{/if}
	</div>
</div>
