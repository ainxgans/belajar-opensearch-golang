import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { createProduct, getProduct, updateProduct, type Attribute, type ProductInput } from "../api";

const emptyForm: ProductInput = {
  sku: "",
  name: "",
  description: "",
  brand: "",
  category: "",
  price: 0,
  stock: 0,
  rating: 0,
  tags: [],
  attributes: [],
};

export default function ProductForm() {
  const { id } = useParams({ strict: false });
  const editing = id != null;
  const productId = editing ? Number(id) : undefined;
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data } = useQuery({
    queryKey: ["product", productId],
    queryFn: () => getProduct(productId!),
    enabled: editing,
  });

  const [form, setForm] = useState<ProductInput>(emptyForm);
  const [tagsText, setTagsText] = useState("");

  useEffect(() => {
    if (data) {
      setForm({
        sku: data.sku,
        name: data.name,
        description: data.description,
        brand: data.brand,
        category: data.category,
        price: data.price,
        stock: data.stock,
        rating: data.rating,
        tags: data.tags,
        attributes: data.attributes,
      });
      setTagsText(data.tags.join(", "));
    }
  }, [data]);

  const mutation = useMutation({
    mutationFn: (input: ProductInput) =>
      editing ? updateProduct(productId!, input) : createProduct(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["search"] });
      navigate({ to: "/" });
    },
  });

  function updateField<K extends keyof ProductInput>(key: K, value: ProductInput[K]) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  function updateAttribute(i: number, field: keyof Attribute, value: string) {
    setForm((f) => {
      const attributes = [...f.attributes];
      attributes[i] = { ...attributes[i], [field]: value };
      return { ...f, attributes };
    });
  }

  function addAttribute() {
    setForm((f) => ({ ...f, attributes: [...f.attributes, { name: "", value: "" }] }));
  }

  function removeAttribute(i: number) {
    setForm((f) => ({ ...f, attributes: f.attributes.filter((_, idx) => idx !== i) }));
  }

  function submit(e: React.FormEvent) {
    e.preventDefault();
    const tags = tagsText
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    mutation.mutate({ ...form, tags });
  }

  const inputClass =
    "mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9]";
  const labelClass = "text-sm font-medium";

  return (
    <form
      onSubmit={submit}
      className="flex max-w-[480px] flex-col gap-4 rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1e293b]"
    >
      <h2 className="text-lg font-semibold">{editing ? "Edit product" : "New product"}</h2>

      <label className={labelClass}>
        SKU
        <input value={form.sku} onChange={(e) => updateField("sku", e.target.value)} required className={inputClass} />
      </label>
      <label className={labelClass}>
        Name
        <input value={form.name} onChange={(e) => updateField("name", e.target.value)} required className={inputClass} />
      </label>
      <label className={labelClass}>
        Description
        <input
          value={form.description}
          onChange={(e) => updateField("description", e.target.value)}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Brand
        <input value={form.brand} onChange={(e) => updateField("brand", e.target.value)} required className={inputClass} />
      </label>
      <label className={labelClass}>
        Category
        <input
          value={form.category}
          onChange={(e) => updateField("category", e.target.value)}
          required
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Price
        <input
          type="number"
          value={form.price}
          onChange={(e) => updateField("price", Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Stock
        <input
          type="number"
          value={form.stock}
          onChange={(e) => updateField("stock", Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Rating
        <input
          type="number"
          step={0.1}
          min={0}
          max={5}
          value={form.rating}
          onChange={(e) => updateField("rating", Number(e.target.value))}
          className={inputClass}
        />
      </label>
      <label className={labelClass}>
        Tags (comma-separated)
        <textarea value={tagsText} onChange={(e) => setTagsText(e.target.value)} className={inputClass} />
      </label>

      <div>
        <strong className="text-sm font-semibold">Attributes</strong>
        {form.attributes.map((attr, i) => (
          <div key={i} className="mt-2 flex gap-2">
            <input
              placeholder="name"
              value={attr.name}
              onChange={(e) => updateAttribute(i, "name", e.target.value)}
              className={inputClass + " mt-0"}
            />
            <input
              placeholder="value"
              value={attr.value}
              onChange={(e) => updateAttribute(i, "value", e.target.value)}
              className={inputClass + " mt-0"}
            />
            <button
              type="button"
              onClick={() => removeAttribute(i)}
              className="shrink-0 rounded-md bg-red-500 px-3 py-1.5 text-sm text-white transition-colors hover:bg-red-600"
            >
              Remove
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={addAttribute}
          className="mt-3 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm transition-colors hover:bg-gray-50 dark:border-gray-600 dark:bg-[#1e293b] dark:hover:bg-gray-800"
        >
          Add attribute
        </button>
      </div>

      {mutation.error && (
        <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-950 dark:text-red-400">
          {(mutation.error as Error).message}
        </p>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="rounded-md bg-blue-500 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {mutation.isPending ? "Saving…" : editing ? "Save" : "Create"}
      </button>
    </form>
  );
}
