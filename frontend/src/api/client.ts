// Typed access to the two read-only endpoints. Types come only from contract.gen.ts.
import type { ErrorResponse, ModelsResponse, SummaryResponse } from './contract.gen';

/** Raised for every failure; the message is plain text safe to render. */
export class ApiError extends Error {
  readonly status: number | null;
  readonly code: string | null;

  constructor(message: string, status: number | null, code: string | null) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

// Returns the error body's code and message when the text is a well-formed ErrorResponse.
function readErrorBody(text: string): ErrorResponse['error'] | null {
  try {
    const body: unknown = JSON.parse(text);
    if (isRecord(body) && isRecord(body.error)) {
      const { code, message } = body.error;
      if (typeof code === 'string' && typeof message === 'string') return { code, message };
    }
  } catch {
    // Not JSON: fall through to the status-only message.
  }
  return null;
}

async function getJson<T>(path: string): Promise<T> {
  let res: Response;
  let text: string;
  try {
    res = await fetch(path);
    text = await res.text();
  } catch (e) {
    const reason = e instanceof Error ? e.message : String(e);
    throw new ApiError(`request to ${path} failed: ${reason}`, null, null);
  }

  if (!res.ok) {
    const err = readErrorBody(text);
    if (err) throw new ApiError(`${err.code}: ${err.message} (HTTP ${res.status})`, res.status, err.code);
    throw new ApiError(`HTTP ${res.status} from ${path} with no readable error body`, res.status, null);
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    throw new ApiError(`${path} returned a body that is not JSON`, res.status, null);
  }
  if (!isRecord(parsed)) {
    throw new ApiError(`${path} returned JSON that is not an object`, res.status, null);
  }
  // The server is the contract's other half; the drift tests compare key sets.
  return parsed as unknown as T;
}

export function getSummary(): Promise<SummaryResponse> {
  return getJson<SummaryResponse>('/api/summary');
}

export function getModels(): Promise<ModelsResponse> {
  return getJson<ModelsResponse>('/api/models');
}
