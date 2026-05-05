import { encoreApiUrl } from './encore';

export type RowError = {
	row_number: number;
	field?: string;
	error: string;
};

export type ImportJob = {
	id: number;
	kind: string;
	institution_id: number;
	status: 'pending' | 'running' | 'completed' | 'failed';
	total_rows: number;
	succeeded: number;
	failed: number;
	dry_run: boolean;
	errors: RowError[];
	created_at: string;
	completed_at?: string;
};

type Opts = { token: string };

export async function importStudents(
	{ token }: Opts,
	body: { institution_id: number; csv_data: string; dry_run: boolean }
): Promise<{ ok: boolean; status: number; job?: ImportJob; message?: string }> {
	const res = await fetch(`${encoreApiUrl}/v1/imports/students`, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		return { ok: false, status: res.status, message: data?.message ?? 'Import failed' };
	}
	return { ok: true, status: res.status, job: data as ImportJob };
}
