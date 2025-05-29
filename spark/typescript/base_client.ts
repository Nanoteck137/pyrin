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
  headers: Map<string, string>;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
    this.headers = new Map<string, string>();
  }

  createEndpointUrl(endpoint: string) {
    return new URL(this.baseUrl + endpoint);
  }

  private getInitialHeaders() {
    const headers: Record<string, string> = {};

    this.headers.forEach((v, k) => {
      headers[k] = v;
    });

    return headers;
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
    const url = this.createEndpointUrl(endpoint);
    const headers = this.getInitialHeaders();

    if (body) {
      headers["Content-Type"] = "application/json";
    }

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
    const url = this.createEndpointUrl(endpoint);
    const headers = this.getInitialHeaders();

    if (body) {
      headers["Content-Type"] = "multipart/form-data";
    }

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
