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

  return (
    <form onSubmit={submit} style={{ maxWidth: 480, display: "flex", flexDirection: "column", gap: 12 }}>
      <h2>Bulk generator</h2>

      <label>
        Count
        <input type="number" value={count} onChange={(e) => setCount(Number(e.target.value))} />
      </label>
      <label>
        Seed
        <input type="number" value={seed} onChange={(e) => setSeed(Number(e.target.value))} />
      </label>
      <label>
        Max attributes per product
        <input type="number" value={attrMax} onChange={(e) => setAttrMax(Number(e.target.value))} />
      </label>
      <label>
        SKU/name prefix
        <input value={prefix} onChange={(e) => setPrefix(e.target.value)} />
      </label>
      <label>
        Price min
        <input type="number" value={priceMin} onChange={(e) => setPriceMin(Number(e.target.value))} />
      </label>
      <label>
        Price max
        <input type="number" value={priceMax} onChange={(e) => setPriceMax(Number(e.target.value))} />
      </label>
      <label>
        Brands (comma-separated)
        <textarea value={brands} onChange={(e) => setBrands(e.target.value)} />
      </label>
      <label>
        Categories (comma-separated)
        <textarea value={categories} onChange={(e) => setCategories(e.target.value)} />
      </label>
      <label>
        Tags (comma-separated)
        <textarea value={tags} onChange={(e) => setTags(e.target.value)} />
      </label>
      <label>
        <input type="checkbox" checked={truncate} onChange={(e) => setTruncate(e.target.checked)} />
        Truncate existing products first
      </label>

      <button type="submit" disabled={mutation.isPending}>
        Generate
      </button>

      {mutation.isSuccess && (
        <p>
          Inserted {mutation.data.inserted} products.
          {mutation.data.note && <> {mutation.data.note}</>}
        </p>
      )}
      {mutation.error && <p style={{ color: "red" }}>{(mutation.error as Error).message}</p>}
    </form>
  );
}
