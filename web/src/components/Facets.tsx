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
    <div style={{ marginBottom: 16 }}>
      <strong>{title}</strong>
      {buckets.map((b) => (
        <label key={b.key} style={{ display: "block", fontSize: 14 }}>
          <input
            type="radio"
            name={title}
            checked={value === b.key}
            onChange={() => onSelect(value === b.key ? undefined : b.key)}
          />{" "}
          {b.key} ({b.doc_count})
        </label>
      ))}
    </div>
  );
}

export default function Facets({ facets, params, onChange, onReset }: Props) {
  const brands: Bucket[] = facets.brands?.buckets ?? [];
  const categories: Bucket[] = facets.categories?.buckets ?? [];
  const tags: Bucket[] = facets.tags?.buckets ?? [];
  const attrNames: AttrBucket[] = facets.attributes?.names?.buckets ?? [];

  return (
    <div style={{ width: 220 }}>
      <button onClick={onReset} style={{ marginBottom: 16 }}>
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

      <div style={{ marginBottom: 16 }}>
        <strong>Price</strong>
        <div style={{ display: "flex", gap: 4 }}>
          <input
            type="number"
            placeholder="min"
            value={params.price_min ?? ""}
            onChange={(e) => onChange({ price_min: e.target.value ? Number(e.target.value) : undefined })}
            style={{ width: "50%" }}
          />
          <input
            type="number"
            placeholder="max"
            value={params.price_max ?? ""}
            onChange={(e) => onChange({ price_max: e.target.value ? Number(e.target.value) : undefined })}
            style={{ width: "50%" }}
          />
        </div>
      </div>

      <div style={{ marginBottom: 16 }}>
        <strong>Min rating</strong>
        <input
          type="number"
          min={0}
          max={5}
          step={0.1}
          value={params.rating_min ?? ""}
          onChange={(e) => onChange({ rating_min: e.target.value ? Number(e.target.value) : undefined })}
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
