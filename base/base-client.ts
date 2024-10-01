import { z } from "zod";

export function createApiResponse<
  Data extends z.ZodTypeAny,
  ErrorExtra extends z.ZodTypeAny,
>(data: Data, errorExtra: ErrorExtra) {
  return z.discriminatedUnion("success", [
    z.object({ success: z.literal(true), data }),
    z.object({
      success: z.literal(false),
      error: z.object({
        code: z.number(),
        message: z.string(),
        type: z.string(),
        extra: errorExtra,
      }),
    }),
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

  createEndpointUrl(endpoint: string) {
    return new URL(this.baseUrl + endpoint);
  }

  async request<
    DataSchema extends z.ZodTypeAny,
    ErrorExtraSchema extends z.ZodTypeAny,
  >(
    endpoint: string,
    method: string,
    dataSchema: DataSchema,
    errorExtraSchema: ErrorExtraSchema,
    body?: unknown,
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

    const Schema = createApiResponse(dataSchema, errorExtraSchema);

    const data = await res.json();
    const parsedData = await Schema.parseAsync(data);

    return parsedData;
  }

  async requestWithFormData<
    DataSchema extends z.ZodTypeAny,
    ErrorExtraSchema extends z.ZodTypeAny,
  >(
    endpoint: string,
    method: string,
    dataSchema: DataSchema,
    errorExtraSchema: ErrorExtraSchema,
    body: FormData,
    extra?: ExtraOptions,
  ) {
    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
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
      body,
    });

    const Schema = createApiResponse(dataSchema, errorExtraSchema);

    const data = await res.json();
    const parsedData = await Schema.parseAsync(data);

    return parsedData;
  }
}
