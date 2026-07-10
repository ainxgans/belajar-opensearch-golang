const BASE = "/api/products";

export interface Attribute {
  name: string;
  value: string;
}

export interface Product {
  id: number;
  sku: string;
  name: string;
  description: string;
  brand: string;
  category: string;
  price: number;
  stock: number;
  rating: number;
  tags: string[];
  attributes: Attribute[];
  created_at: string;
  updated_at: string;
}

// ProductInput is the writable subset of Product used for create/update.
export type ProductInput = Pick<
  Product,
  "sku" | "name" | "description" | "brand" | "category" | "price" | "stock" | "rating" | "tags" | "attributes"
>;

export interface SearchParams {
  q?: string;
  brand?: string;
  category?: string;
  tags?: string[];
  price_min?: number;
  price_max?: number;
  rating_min?: number;
  attrs?: Record<string, string>;
  sort?: string;
  page?: number;
  size?: number;
}

export interface SearchResult {
  items: Product[];
  total: number;
  facets: Record<string, any>;
  suggestions: string[];
  took_ms: number;
}

export interface GenerateConfig {
  count?: number;
  prefix?: string;
  seed?: number;
  price_min?: number;
  price_max?: number;
  brands?: string[];
  categories?: string[];
  tags?: string[];
  attr_max?: number;
  truncate?: boolean;
}

export interface GenerateResult {
  inserted: number;
  note?: string;
}

function paramsToQuery(params: SearchParams): string {
  const q = new URLSearchParams();
  if (params.q) q.set("q", params.q);
  if (params.brand) q.set("brand", params.brand);
  if (params.category) q.set("category", params.category);
  if (params.tags?.length) q.set("tags", params.tags.join(","));
  if (params.price_min != null) q.set("price_min", String(params.price_min));
  if (params.price_max != null) q.set("price_max", String(params.price_max));
  if (params.rating_min != null) q.set("rating_min", String(params.rating_min));
  if (params.sort) q.set("sort", params.sort);
  if (params.page != null) q.set("page", String(params.page));
  if (params.size != null) q.set("size", String(params.size));
  for (const [name, value] of Object.entries(params.attrs ?? {})) {
    q.set(`attr.${name}`, value);
  }
  return q.toString();
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `request failed: ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export function searchProducts(params: SearchParams): Promise<SearchResult> {
  return request(`${BASE}/search?${paramsToQuery(params)}`);
}

export function autocomplete(q: string): Promise<{ suggestions: string[] }> {
  return request(`${BASE}/autocomplete?q=${encodeURIComponent(q)}`);
}

export function getProduct(id: number): Promise<Product> {
  return request(`${BASE}/${id}`);
}

export function createProduct(input: ProductInput): Promise<Product> {
  return request(BASE, { method: "POST", body: JSON.stringify(input) });
}

export function updateProduct(id: number, input: ProductInput): Promise<Product> {
  return request(`${BASE}/${id}`, { method: "PUT", body: JSON.stringify(input) });
}

export function deleteProduct(id: number): Promise<void> {
  return request(`${BASE}/${id}`, { method: "DELETE" });
}

export function generate(cfg: GenerateConfig): Promise<GenerateResult> {
  return request(`${BASE}/generate`, { method: "POST", body: JSON.stringify(cfg) });
}

export function exportUrl(params: SearchParams): string {
  return `${BASE}/export.xlsx?${paramsToQuery(params)}`;
}
