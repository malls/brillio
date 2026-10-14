import { describe, expect, it } from 'vitest';
import { internalErrorBody, jsonResponse, validationErrorBody } from '../test/fixtures';
import { ApiRequestError, parseErrorResponse, toApiError } from './errors';

describe('parseErrorResponse', () => {
  it('parses a validation_error body with every field', async () => {
    const res = jsonResponse(validationErrorBody, { status: 400 });

    await expect(parseErrorResponse(res)).resolves.toEqual({
      kind: 'validation',
      message: 'Invalid search parameters.',
      fields: validationErrorBody.error.fields,
    });
  });

  it('drops malformed field entries and keeps the valid ones', async () => {
    const valid = {
      field: 'minBedrooms',
      code: 'negative',
      message: 'Min bedrooms must be 0 or more.',
    };
    const res = jsonResponse(
      {
        error: {
          code: 'validation_error',
          message: 'Invalid search parameters.',
          fields: [
            valid,
            null,
            'minPrice',
            { field: 'city' },
            { field: 1, code: 'x', message: 'y' },
          ],
        },
      },
      { status: 400 },
    );

    await expect(parseErrorResponse(res)).resolves.toEqual({
      kind: 'validation',
      message: 'Invalid search parameters.',
      fields: [valid],
    });
  });

  it('gives no fields when a validation_error has none', async () => {
    const res = jsonResponse(
      { error: { code: 'validation_error', message: 'Invalid search parameters.' } },
      { status: 400 },
    );

    await expect(parseErrorResponse(res)).resolves.toEqual({
      kind: 'validation',
      message: 'Invalid search parameters.',
      fields: [],
    });
  });

  it('parses any other error code as an http error with the body message', async () => {
    const res = jsonResponse(internalErrorBody, { status: 500 });

    await expect(parseErrorResponse(res)).resolves.toEqual({
      kind: 'http',
      status: 500,
      message: 'Internal server error.',
    });
  });

  it.each([
    ['an empty body', new Response(null, { status: 503 })],
    ['an HTML body', new Response('<html>Bad Gateway</html>', { status: 503 })],
    ['a JSON array', jsonResponse([1, 2], { status: 503 })],
    ['a body without error', jsonResponse({ message: 'nope' }, { status: 503 })],
    ['an empty message', jsonResponse({ error: { code: 'x', message: '' } }, { status: 503 })],
  ])('falls back to a generic message for %s', async (_name, res) => {
    await expect(parseErrorResponse(res)).resolves.toEqual({
      kind: 'http',
      status: 503,
      message: 'Request failed with status 503.',
    });
  });
});

describe('ApiRequestError', () => {
  it('is an Error carrying the ApiError, the request ID and the cause', () => {
    const cause = new TypeError('Failed to fetch');
    const err = new ApiRequestError({ kind: 'network', message: 'Failed to fetch' }, null, {
      cause,
    });

    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe('ApiRequestError');
    expect(err.message).toBe('Failed to fetch');
    expect(err.apiError).toEqual({ kind: 'network', message: 'Failed to fetch' });
    expect(err.requestId).toBeNull();
    expect(err.cause).toBe(cause);
  });
});

describe('toApiError', () => {
  it('returns the ApiError of an ApiRequestError', () => {
    const apiError = { kind: 'http', status: 404, message: 'Resource not found.' } as const;

    expect(toApiError(new ApiRequestError(apiError, 'req-1'))).toBe(apiError);
  });

  it('reports any other Error as a network error with its message', () => {
    expect(toApiError(new Error('boom'))).toEqual({ kind: 'network', message: 'boom' });
  });

  it('reports a non-Error as a network error with a generic message', () => {
    expect(toApiError('boom')).toEqual({ kind: 'network', message: 'Request failed.' });
  });
});
