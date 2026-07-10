import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { flexRender, getCoreRowModel, useReactTable, createColumnHelper } from "@tanstack/react-table";
import { deleteProduct, exportUrl, searchProducts, type Product, type SearchParams } from "../api";
import SearchBar from "../components/SearchBar";
import Facets from "../components/Facets";

const SORT_OPTIONS = [
  { value: "", label: "Relevance" },
  { value: "price_asc", label: "Price: low to high" },
  { value: "price_desc", label: "Price: high to low" },
  { value: "rating", label: "Rating" },
  { value: "newest", label: "Newest" },
];

const columnHelper = createColumnHelper<Product>();

export default function SearchPage() {
  const [params, setParams] = useState<SearchParams>({ page: 1, size: 20 });
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ["search", params],
    queryFn: () => searchProducts(params),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["search"] }),
  });

  function patch(p: Partial<SearchParams>) {
    setParams((prev) => ({ ...prev, ...p, page: 1 }));
  }

  const columns = useMemo(
    () => [
      columnHelper.accessor("sku", { header: "SKU" }),
      columnHelper.accessor("name", { header: "Name" }),
      columnHelper.accessor("brand", { header: "Brand" }),
      columnHelper.accessor("category", { header: "Category" }),
      columnHelper.accessor("price", { header: "Price", cell: (c) => c.getValue().toLocaleString() }),
      columnHelper.accessor("stock", { header: "Stock" }),
      columnHelper.accessor("rating", { header: "Rating" }),
      columnHelper.display({
        id: "actions",
        header: "Actions",
        cell: (c) => (
          <div style={{ display: "flex", gap: 8 }}>
            <button onClick={() => navigate({ to: "/edit/$id", params: { id: String(c.row.original.id) } })}>
              Edit
            </button>
            <button
              onClick={() => {
                if (confirm(`Delete ${c.row.original.sku}?`)) deleteMutation.mutate(c.row.original.id);
              }}
            >
              Delete
            </button>
          </div>
        ),
      }),
    ],
    [navigate, deleteMutation],
  );

  const table = useReactTable({
    data: data?.items ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  const size = params.size ?? 20;
  const page = params.page ?? 1;
  const total = data?.total ?? 0;
  const hasNext = page * size < total;

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <SearchBar value={params.q ?? ""} onSearch={(q) => patch({ q })} />
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <select value={params.sort ?? ""} onChange={(e) => patch({ sort: e.target.value || undefined })}>
            {SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <a href={exportUrl(params)}>Export Excel</a>
        </div>
      </div>

      {data?.suggestions && data.suggestions.length > 0 && (
        <p>
          Did you mean:{" "}
          {data.suggestions.map((s) => (
            <button key={s} onClick={() => patch({ q: s })} style={{ marginRight: 8 }}>
              {s}
            </button>
          ))}
        </p>
      )}

      <div style={{ display: "flex", gap: 24 }}>
        <Facets facets={data?.facets ?? {}} params={params} onChange={patch} onReset={() => setParams({ page: 1, size: 20 })} />

        <div style={{ flex: 1 }}>
          {isLoading && <p>Loading…</p>}
          {error && <p style={{ color: "red" }}>{(error as Error).message}</p>}

          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              {table.getHeaderGroups().map((hg) => (
                <tr key={hg.id}>
                  {hg.headers.map((h) => (
                    <th key={h.id} style={{ textAlign: "left", borderBottom: "1px solid #ddd", padding: 8 }}>
                      {flexRender(h.column.columnDef.header, h.getContext())}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows.map((row) => (
                <tr key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} style={{ padding: 8, borderBottom: "1px solid #eee" }}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>

          <div style={{ display: "flex", gap: 8, marginTop: 16, alignItems: "center" }}>
            <button disabled={page <= 1} onClick={() => setParams((p) => ({ ...p, page: page - 1 }))}>
              Prev
            </button>
            <span>
              Page {page} · {total} results
            </span>
            <button disabled={!hasNext} onClick={() => setParams((p) => ({ ...p, page: page + 1 }))}>
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
