import { z } from "zod";

function createApiResponse<
  Data extends z.ZodTypeAny,
  Errors extends z.ZodTypeAny,
>(data: Data, errors: Errors) {
  const error = z.object({
    code: z.number(),
    message: z.string(),
    errors,
  });

  return z.discriminatedUnion("status", [
    z.object({ status: z.literal("success"), data }),
    z.object({ status: z.literal("error"), error }),
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

  async request<T extends z.ZodTypeAny>(
    endpoint: string,
    method: string,
    bodySchema: T,
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

    const Schema = createApiResponse(bodySchema, z.undefined());

    const data = await res.json();
    const parsedData = await Schema.parseAsync(data);

    return parsedData;
  }
}
