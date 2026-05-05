import { S3Client, GetObjectCommand, HeadObjectCommand } from '@aws-sdk/client-s3';
import { Upload } from '@aws-sdk/lib-storage';
import { env } from '$env/dynamic/private';

function requireEnv(name: string): string {
	const v = env[name];
	if (!v) throw new Error(`${name} is required (no fallback for credentials)`);
	return v;
}

const ENDPOINT = env.MINIO_ENDPOINT ?? 'http://minio:9000';
const ACCESS_KEY = requireEnv('MINIO_ACCESS_KEY');
const SECRET_KEY = requireEnv('MINIO_SECRET_KEY');
const REGION = env.MINIO_REGION ?? 'us-east-1';

export const MINIO_BUCKET = env.MINIO_BUCKET ?? 'schoolrise-uploads';

export const minio = new S3Client({
	endpoint: ENDPOINT,
	region: REGION,
	credentials: { accessKeyId: ACCESS_KEY, secretAccessKey: SECRET_KEY },
	forcePathStyle: true
});

// Returns a same-origin URL the browser can fetch. The actual MinIO bucket is
// never publicly exposed; this path resolves through SvelteKit's
// /api/uploads/[...key]/+server.ts proxy, which streams the object from MinIO
// with session-aware headers.
export function publicUrlFor(key: string): string {
	return `/api/uploads/${key.split('/').map(encodeURIComponent).join('/')}`;
}

export async function uploadStream(
	key: string,
	body: Uint8Array | Buffer | ReadableStream,
	contentType: string
): Promise<void> {
	const upload = new Upload({
		client: minio,
		params: { Bucket: MINIO_BUCKET, Key: key, Body: body, ContentType: contentType }
	});
	await upload.done();
}

export async function getObject(key: string) {
	const cmd = new GetObjectCommand({ Bucket: MINIO_BUCKET, Key: key });
	return await minio.send(cmd);
}

export async function headObject(key: string) {
	const cmd = new HeadObjectCommand({ Bucket: MINIO_BUCKET, Key: key });
	return await minio.send(cmd);
}
