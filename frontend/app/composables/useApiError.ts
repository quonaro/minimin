export interface ApiErrorData {
  error?: string;
  detail?: string;
  message?: string;
}

export function getApiErrorMessage(err: unknown, fallback?: string): string {
  if (err === null || err === undefined) {
    return fallback ?? "Request failed";
  }

  const e = err as any;

  const data = e?.data as ApiErrorData | undefined;
  if (data?.error) {
    return data.error;
  }
  if (data?.detail) {
    return data.detail;
  }
  if (data?.message) {
    return data.message;
  }

  if (typeof e?.message === "string" && e.message) {
    return e.message;
  }
  if (typeof e?.statusText === "string" && e.statusText) {
    return e.statusText;
  }

  return fallback ?? "Request failed";
}

export function getApiErrorStatus(err: unknown): number | undefined {
  const e = err as any;
  return e?.status || e?.statusCode;
}
