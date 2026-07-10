import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { generate, type GenerateConfig } from "../api";

function splitList(text: string): string[] {
  return text
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

export default function GeneratorPage() {
  const [count, setCount] = useState(100);
  const [seed, setSeed] = useState(1);
  const [attrMax, setAttrMax] = useState(3);
  const [prefix, setPrefix] = useState("GEN");
  const [priceMin, setPriceMin] = useState(500);
  const [priceMax, setPriceMax] = useState(500000);
  const [brands, setBrands] = useState("");
  const [categories, setCategories] = useState("");
  const [tags, setTags] = useState("");
  const [truncate, setTruncate] = useState(false);

  const mutation = useMutation({
    mutationFn: (cfg: GenerateConfig) => generate(cfg),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    mutation.mutate({
      count,
      seed,
      attr_max: attrMax,
      prefix,
      price_min: priceMin,
      price_max: priceMax,
      brands: splitList(brands),
      categories: splitList(categories),
      tags: splitList(tags),
      truncate,
    });
  }

  const inputClass =
    "mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9]";
  const labelClass = "text-sm font-medium";

  return (
    <form
      onSubmit={submit}
      className="flex max-w-[480px] flex-col gap-4 rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1e293b]"
    >
      <h2 className="text-lg font-semibold">Bulk generator</h2>

      <label className={labelClass}>
        Count
        <input
          type="number"
          value={count}
          onChange={(e) => setCount(Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Seed
        <input type="number" value={seed} onChange={(e) => setSeed(Number(e.target.value))} className={inputClass} />
      </label>
      <label className={labelClass}>
        Max attributes per product
        <input
          type="number"
          value={attrMax}
          onChange={(e) => setAttrMax(Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        SKU/name prefix
        <input value={prefix} onChange={(e) => setPrefix(e.target.value)} className={inputClass} />
      </label>
      <label className={labelClass}>
        Price min
        <input
          type="number"
          value={priceMin}
          onChange={(e) => setPriceMin(Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Price max
        <input
          type="number"
          value={priceMax}
          onChange={(e) => setPriceMax(Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Brands (comma-separated)
        <textarea value={brands} onChange={(e) => setBrands(e.target.value)} className={inputClass} />
      </label>
      <label className={labelClass}>
        Categories (comma-separated)
        <textarea value={categories} onChange={(e) => setCategories(e.target.value)} className={inputClass} />
      </label>
      <label className={labelClass}>
        Tags (comma-separated)
        <textarea value={tags} onChange={(e) => setTags(e.target.value)} className={inputClass} />
      </label>
      <label className="flex items-center gap-2 text-sm font-medium">
        <input
          type="checkbox"
          checked={truncate}
          onChange={(e) => setTruncate(e.target.checked)}
          className="accent-amber-500"
        />
        Truncate existing products first
      </label>
      {truncate && (
        <p className="text-sm text-amber-500 dark:text-amber-400">
          Warning: this will delete existing products before generating new ones.
        </p>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="rounded-md bg-blue-500 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {mutation.isPending ? "Generating…" : "Generate"}
      </button>

      {mutation.isSuccess && (
        <p className="rounded-md bg-green-50 px-3 py-2 text-sm text-green-600 dark:bg-green-950 dark:text-green-400">
          Inserted {mutation.data.inserted} products.
          {mutation.data.note && <> {mutation.data.note}</>}
        </p>
      )}
      {mutation.error && (
        <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-950 dark:text-red-400">
          {(mutation.error as Error).message}
        </p>
      )}
    </form>
  );
}
