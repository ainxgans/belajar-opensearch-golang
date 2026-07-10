import type { SearchParams } from "../api";

interface Bucket {
  key: string;
  doc_count: number;
}

interface AttrBucket extends Bucket {
  values?: { buckets: Bucket[] };
}

interface Props {
  facets: Record<string, any>;
  params: SearchParams;
  onChange: (patch: Partial<SearchParams>) => void;
  onReset: () => void;
}

function RadioGroup({
  title,
  buckets,
  value,
  onSelect,
}: {
  title: string;
  buckets: Bucket[];
  value?: string;
  onSelect: (v: string | undefined) => void;
}) {
  if (!buckets.length) return null;
  return (
    <div className="mb-4">
      <strong className="text-sm font-semibold">{title}</strong>
      <div className="mt-1 flex flex-col gap-1">
        {buckets.map((b) => (
          <label key={b.key} className="flex items-center gap-1.5 text-sm">
            <input
              type="radio"
              name={title}
              checked={value === b.key}
              onChange={() => onSelect(value === b.key ? undefined : b.key)}
              className="accent-blue-500"
            />
            {b.key} ({b.doc_count})
          </label>
        ))}
      </div>
    </div>
  );
}

export default function Facets({ facets, params, onChange, onReset }: Props) {
  const brands: Bucket[] = facets.brands?.buckets ?? [];
  const categories: Bucket[] = facets.categories?.buckets ?? [];
  const tags: Bucket[] = facets.tags?.buckets ?? [];
  const attrNames: AttrBucket[] = facets.attributes?.names?.buckets ?? [];

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-[#1e293b]">
      <button
        onClick={onReset}
        className="mb-4 w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-[#1a202c] transition-colors hover:bg-gray-50 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9] dark:hover:bg-gray-800"
      >
        Reset filters
      </button>

      <RadioGroup title="Brand" buckets={brands} value={params.brand} onSelect={(v) => onChange({ brand: v })} />
      <RadioGroup
        title="Category"
        buckets={categories}
        value={params.category}
        onSelect={(v) => onChange({ category: v })}
      />
      <RadioGroup
        title="Tags"
        buckets={tags}
        value={params.tags?.[0]}
        onSelect={(v) => onChange({ tags: v ? [v] : undefined })}
      />

      <div className="mb-4">
        <strong className="text-sm font-semibold">Price</strong>
        <div className="mt-1 flex gap-2">
          <input
            type="number"
            placeholder="min"
            value={params.price_min ?? ""}
            onChange={(e) => onChange({ price_min: e.target.value ? Number(e.target.value) : undefined })}
            className="w-1/2 rounded-md border border-gray-300 bg-white px-2 py-1 text-sm text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#0f172a] dark:text-[#f1f5f9]"
          />
          <input
            type="number"
            placeholder="max"
            value={params.price_max ?? ""}
            onChange={(e) => onChange({ price_max: e.target.value ? Number(e.target.value) : undefined })}
            className="w-1/2 rounded-md border border-gray-300 bg-white px-2 py-1 text-sm text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#0f172a] dark:text-[#f1f5f9]"
          />
        </div>
      </div>

      <div className="mb-4">
        <strong className="text-sm font-semibold">Min rating</strong>
        <input
          type="number"
          min={0}
          max={5}
          step={0.1}
          value={params.rating_min ?? ""}
          onChange={(e) => onChange({ rating_min: e.target.value ? Number(e.target.value) : undefined })}
          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-2 py-1 text-sm text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#0f172a] dark:text-[#f1f5f9]"
        />
      </div>

      {attrNames.map((attr) => (
        <RadioGroup
          key={attr.key}
          title={attr.key}
          buckets={attr.values?.buckets ?? []}
          value={params.attrs?.[attr.key]}
          onSelect={(v) => {
            const attrs = { ...(params.attrs ?? {}) };
            if (v) attrs[attr.key] = v;
            else delete attrs[attr.key];
            onChange({ attrs });
          }}
        />
      ))}
    </div>
  );
}
