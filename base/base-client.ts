import { z } from "zod";

export function createError<
  ErrorTypes extends [string, ...string[]],
  ErrorExtra extends z.ZodTypeAny,
>(types: z.ZodEnum<ErrorTypes>, extra: ErrorExtra) {
  return z.object({
    code: z.number(),
    message: z.string(),
    type: types,
    extra: extra,
  });
}

export function createApiResponse<
  Data extends z.ZodTypeAny,
  Error extends z.ZodTypeAny,
>(data: Data, error: Error) {
  return z.discriminatedUnion("success", [
    z.object({ success: z.literal(true), data }),
    z.object({ success: z.literal(false), error }),
  ]);
}

export type ExtraOptions = {
  headers?: Record<string, string>;
  query?: Record<string, string>;
};

export class BaseApiClient {
  baseUrl: string;
  token?: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  setToken(token?: string) {
    this.token = token;
  }

  async request<B extends z.ZodTypeAny, E extends z.ZodTypeAny>(
    endpoint: string,
    method: string,
    bodySchema: B,
    errorSchema: E,
    body?: any,
    extra?: ExtraOptions,
  ) {
    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    if (body) {
      headers["Content-Type"] = "application/json";
    }

    const url = new URL(this.baseUrl + endpoint);

    if (extra) {
      if (extra.headers) {
        for (const [key, value] of Object.entries(extra.headers)) {
          headers[key] = value;
        }
      }

      if (extra.query) {
        for (const [key, value] of Object.entries(extra.query)) {
          url.searchParams.set(key, value);
        }
      }
    }

    const res = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : null,
    });

    const Schema = createApiResponse(bodySchema, errorSchema);

    const data = await res.json();
    const parsedData = await Schema.parseAsync(data);

    return parsedData;
  }
}
