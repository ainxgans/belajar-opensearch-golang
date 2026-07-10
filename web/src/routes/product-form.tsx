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

  return (
    <form onSubmit={submit} style={{ maxWidth: 480, display: "flex", flexDirection: "column", gap: 12 }}>
      <h2>{editing ? "Edit product" : "New product"}</h2>

      <label>
        SKU
        <input value={form.sku} onChange={(e) => updateField("sku", e.target.value)} required />
      </label>
      <label>
        Name
        <input value={form.name} onChange={(e) => updateField("name", e.target.value)} required />
      </label>
      <label>
        Description
        <input value={form.description} onChange={(e) => updateField("description", e.target.value)} />
      </label>
      <label>
        Brand
        <input value={form.brand} onChange={(e) => updateField("brand", e.target.value)} required />
      </label>
      <label>
        Category
        <input value={form.category} onChange={(e) => updateField("category", e.target.value)} required />
      </label>
      <label>
        Price
        <input type="number" value={form.price} onChange={(e) => updateField("price", Number(e.target.value))} />
      </label>
      <label>
        Stock
        <input type="number" value={form.stock} onChange={(e) => updateField("stock", Number(e.target.value))} />
      </label>
      <label>
        Rating
        <input
          type="number"
          step={0.1}
          min={0}
          max={5}
          value={form.rating}
          onChange={(e) => updateField("rating", Number(e.target.value))}
        />
      </label>
      <label>
        Tags (comma-separated)
        <textarea value={tagsText} onChange={(e) => setTagsText(e.target.value)} />
      </label>

      <div>
        <strong>Attributes</strong>
        {form.attributes.map((attr, i) => (
          <div key={i} style={{ display: "flex", gap: 8, marginTop: 4 }}>
            <input
              placeholder="name"
              value={attr.name}
              onChange={(e) => updateAttribute(i, "name", e.target.value)}
            />
            <input
              placeholder="value"
              value={attr.value}
              onChange={(e) => updateAttribute(i, "value", e.target.value)}
            />
            <button type="button" onClick={() => removeAttribute(i)}>
              Remove
            </button>
          </div>
        ))}
        <button type="button" onClick={addAttribute} style={{ marginTop: 8 }}>
          Add attribute
        </button>
      </div>

      {mutation.error && <p style={{ color: "red" }}>{(mutation.error as Error).message}</p>}

      <button type="submit" disabled={mutation.isPending}>
        {editing ? "Save" : "Create"}
      </button>
    </form>
  );
}
