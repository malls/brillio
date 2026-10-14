import type { ApiError, FieldError, ValidationError } from './types';

export type { ApiError, FieldError, ValidationError };

// The rejection value of every client call (docs/SPEC.md section 7.3). typescript-eslint forbids
// throwing or rejecting a plain object, so the ApiError travels inside an Error subclass, together
// with the response's X-Request-ID (null when no response arrived).
export class ApiRequestError extends Error {
  readonly apiError: ApiError;
  readonly requestId: string | null;

  constructor(apiError: ApiError, requestId: string | null, options?: ErrorOptions) {
    super(apiError.message, options);
    this.name = 'ApiRequestError';
    this.apiError = apiError;
    this.requestId = requestId;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isFieldError(value: unknown): value is FieldError {
  return (
    isRecord(value) &&
    typeof value.field === 'string' &&
    typeof value.code === 'string' &&
    typeof value.message === 'string'
  );
}

// Reads a body as JSON. An empty body, a non-JSON body (such as the dev proxy's HTML 502) or a
// failed read all count as no body.
async function readJson(res: Response): Promise<unknown> {
  try {
    const body: unknown = await res.json();
    return body;
  } catch {
    return null;
  }
}

// Turns a non-2xx response into an ApiError (SPEC section 5.5). A validation_error keeps only its
// well-formed field entries; anything else is an http error. The message falls back to a generic
// sentence when the body carries none.
export async function parseErrorResponse(res: Response): Promise<ApiError> {
  const body = await readJson(res);
  const error = isRecord(body) && isRecord(body.error) ? body.error : null;
  const message =
    typeof error?.message === 'string' && error.message !== ''
      ? error.message
      : `Request failed with status ${res.status}.`;

  if (error?.code === 'validation_error') {
    const fields = Array.isArray(error.fields) ? error.fields.filter(isFieldError) : [];
    return { kind: 'validation', message, fields };
  }
  return { kind: 'http', status: res.status, message };
}

// Recovers the ApiError from a client rejection. Anything that is not an ApiRequestError is
// reported as a network error.
export function toApiError(err: unknown): ApiError {
  if (err instanceof ApiRequestError) {
    return err.apiError;
  }
  const message = err instanceof Error && err.message !== '' ? err.message : 'Request failed.';
  return { kind: 'network', message };
}
