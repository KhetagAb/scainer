import type { _Error } from "@/client/types.gen";

/** Ошибка HTTP API — соответствует components/schemas/Error в OpenAPI. */
export class ApiError extends Error implements _Error {
  readonly error: string;

  constructor(body: _Error, readonly status: number) {
    super(body.error);
    this.name = "ApiError";
    this.error = body.error;
  }
}

function echoMessage(error: unknown): string | null {
  if (!error || typeof error !== "object" || !("message" in error)) return null;
  const message = (error as { message: unknown }).message;
  return typeof message === "string" ? message : null;
}

export function normalizeApiError(error: unknown, status: number): ApiError {
  if (error && typeof error === "object" && "error" in error) {
    const body = error as _Error;
    if (typeof body.error === "string") {
      return new ApiError(body, status);
    }
  }
  const message = echoMessage(error);
  if (message) {
    if (status === 404 && message === "Not Found") {
      return new ApiError(
        { error: "метод API не найден на сервере" },
        status,
      );
    }
    return new ApiError({ error: message }, status);
  }
  if (typeof error === "string") {
    return new ApiError({ error }, status);
  }
  return new ApiError({ error: "request failed" }, status);
}

const MAX_PLAIN_BODY = 200;

/** Человекочитаемое сообщение для UI и логов job-хуков. */
export function formatApiError(error: unknown, status?: number): string {
  const prefix = status != null ? `HTTP ${status}: ` : "";

  if (error instanceof ApiError) {
    return `${prefix}${error.error}`;
  }
  if (error && typeof error === "object" && "error" in error) {
    const body = error as _Error;
    if (typeof body.error === "string") {
      return `${prefix}${body.error}`;
    }
  }
  const message = echoMessage(error);
  if (message) {
    if (status === 404 && message === "Not Found") {
      return `${prefix}метод API не найден на сервере`;
    }
    return `${prefix}${message}`;
  }
  if (typeof error === "string") {
    const trimmed =
      error.length > MAX_PLAIN_BODY ? `${error.slice(0, MAX_PLAIN_BODY)}…` : error;
    return `${prefix}${trimmed}`;
  }
  if (error instanceof TypeError && /fetch/i.test(error.message)) {
    return "сеть: не удалось достучаться до API";
  }
  if (error instanceof Error) {
    return `${prefix}${error.message}`;
  }
  return `${prefix}request failed`;
}
