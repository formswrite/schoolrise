import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';
import { getObject } from '$lib/server/minio';

export const GET: RequestHandler = async ({ params }) => {
	const key = params.key;
	if (!key) throw error(400, 'missing key');

	try {
		const obj = await getObject(key);
		const body = obj.Body as ReadableStream | undefined;
		if (!body) throw error(404, 'not found');

		return new Response(body as ReadableStream, {
			headers: {
				'content-type': obj.ContentType ?? 'application/octet-stream',
				'content-length': String(obj.ContentLength ?? ''),
				'cache-control': 'private, max-age=300'
			}
		});
	} catch (e) {
		throw error(404, `object not found: ${(e as Error).message}`);
	}
};
