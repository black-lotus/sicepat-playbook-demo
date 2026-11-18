const baseURL = import.meta.env.VITE_API_BASE_URL ?? '';

/** Error dari API, berisi status HTTP dan kode error dari backend. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export interface Envelope<TData, TMeta> {
  data: TData;
  meta: TMeta;
}

type QueryParams = Record<string, string | number | undefined>;

function buildURL(path: string, params: QueryParams): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      search.set(key, String(value));
    }
  }
  const query = search.toString();
  return `${baseURL}${path}${query ? `?${query}` : ''}`;
}

function isErrorBody(body: unknown): body is { error: { code: string; message: string } } {
  if (typeof body !== 'object' || body === null || !('error' in body)) return false;
  const { error } = body;
  return typeof error === 'object' && error !== null && 'code' in error && 'message' in error;
}

function isEnvelope(body: unknown): body is Envelope<unknown, unknown> {
  return typeof body === 'object' && body !== null && 'data' in body;
}

/** GET ke API dan mengembalikan body berformat {data, meta}. Melempar ApiError bila gagal. */
export async function apiGet<TData, TMeta>(
  path: string,
  params: QueryParams = {},
  signal?: AbortSignal,
): Promise<Envelope<TData, TMeta>> {
  const response = await fetch(buildURL(path, params), {
    headers: { Accept: 'application/json' },
    signal,
  });

  let body: unknown;
  try {
    body = await response.json();
  } catch {
    throw new ApiError(response.status, 'invalid_response', 'Response server tidak valid');
  }

  if (!response.ok) {
    if (isErrorBody(body)) {
      throw new ApiError(response.status, body.error.code, body.error.message);
    }
    throw new ApiError(response.status, 'unknown_error', 'Terjadi kesalahan pada server');
  }
  if (!isEnvelope(body)) {
    throw new ApiError(response.status, 'invalid_response', 'Response server tidak valid');
  }
  return body as Envelope<TData, TMeta>;
}
