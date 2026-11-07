import type { GraphQLResponse } from './types';

const GRAPHQL_ENDPOINT = '/query';

export class GraphQLRequestError extends Error {
  public readonly status?: number;
  public readonly isNetworkError: boolean;

  constructor(message: string, options?: { status?: number; isNetworkError?: boolean }) {
    super(message);
    this.name = 'GraphQLRequestError';
    this.status = options?.status;
    this.isNetworkError = options?.isNetworkError ?? false;
  }
}

export async function executeGraphQL<TData>(
  query: string,
  variables?: Record<string, unknown>,
  signal?: AbortSignal,
): Promise<TData> {
  let response: Response;
  try {
    response = await fetch(GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        query,
        variables: variables ?? {},
      }),
      signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw err;
    }
    throw new GraphQLRequestError(
      'Unable to reach the local GraphQL server at /query. Make sure the backend is running.',
      { isNetworkError: true },
    );
  }

  if (!response.ok) {
    throw new GraphQLRequestError(
      `Server responded with HTTP ${response.status} (${response.statusText || 'Error'}).`,
      { status: response.status },
    );
  }

  let payload: GraphQLResponse<TData>;
  try {
    payload = (await response.json()) as GraphQLResponse<TData>;
  } catch {
    throw new GraphQLRequestError('Received an invalid JSON response from /query.');
  }

  if (Array.isArray(payload.errors) && payload.errors.length > 0) {
    const combinedMessage = payload.errors
      .map((item) => (typeof item?.message === 'string' ? item.message : 'GraphQL error'))
      .join('; ');
    throw new GraphQLRequestError(combinedMessage);
  }

  if (payload.data === undefined || payload.data === null) {
    throw new GraphQLRequestError('GraphQL response did not include data.');
  }

  return payload.data;
}
