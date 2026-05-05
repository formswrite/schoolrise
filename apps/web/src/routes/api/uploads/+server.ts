import type { RequestHandler } from './$types';
import { error, json } from '@sveltejs/kit';
import { uploadStream, publicUrlFor } from '$lib/server/minio';

const MAX_BYTES = 25 * 1024 * 1024;

const ALLOWED_MIME_PREFIXES = [
	'image/',
	'application/pdf',
	'application/msword',
	'application/vnd.openxmlformats-officedocument',
	'audio/',
	'video/mp4',
	'text/plain',
	'text/csv'
];

function extFromMime(mime: string): string {
	const map: Record<string, string> = {
		'image/png': 'png',
		'image/jpeg': 'jpg',
		'image/jpg': 'jpg',
		'image/gif': 'gif',
		'image/webp': 'webp',
		'image/svg+xml': 'svg',
		'application/pdf': 'pdf',
		'audio/mpeg': 'mp3',
		'audio/wav': 'wav',
		'video/mp4': 'mp4',
		'text/plain': 'txt',
		'text/csv': 'csv'
	};
	return map[mime] ?? 'bin';
}

function randomKey(ext: string): string {
	const now = new Date();
	const yyyy = now.getUTCFullYear();
	const mm = String(now.getUTCMonth() + 1).padStart(2, '0');
	const id = crypto.randomUUID();
	return `uploads/${yyyy}/${mm}/${id}.${ext}`;
}

export const POST: RequestHandler = async ({ request }) => {
	const ctype = request.headers.get('content-type') ?? '';
	if (!ctype.startsWith('multipart/form-data')) {
		throw error(415, 'expected multipart/form-data');
	}

	const form = await request.formData();
	const file = form.get('file');
	if (!(file instanceof File)) throw error(400, 'no file uploaded');

	if (file.size > MAX_BYTES) {
		throw error(413, `file too large (${file.size} bytes; max ${MAX_BYTES})`);
	}

	const mime = file.type || 'application/octet-stream';
	if (!ALLOWED_MIME_PREFIXES.some((p) => mime.startsWith(p))) {
		throw error(415, `unsupported content-type: ${mime}`);
	}

	const ext = file.name.includes('.')
		? file.name.split('.').pop()!.toLowerCase().slice(0, 8)
		: extFromMime(mime);
	const key = randomKey(ext);

	const buf = new Uint8Array(await file.arrayBuffer());
	await uploadStream(key, buf, mime);

	return json({
		key,
		url: publicUrlFor(key),
		content_type: mime,
		size: file.size,
		original_name: file.name
	});
};
