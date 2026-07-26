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

export function normalizeApiError(error: unknown, status: number): ApiError {
  if (error && typeof error === "object" && "error" in error) {
    const body = error as _Error;
    if (typeof body.error === "string") {
      return new ApiError(body, status);
    }
  }
  if (typeof error === "string") {
    return new ApiError({ error }, status);
  }
  return new ApiError({ error: "request failed" }, status);
}
