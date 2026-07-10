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
          <div className="flex gap-2">
            <button
              onClick={() => navigate({ to: "/edit/$id", params: { id: String(c.row.original.id) } })}
              className="rounded-md border border-gray-300 bg-white px-3 py-1 text-sm transition-colors hover:bg-gray-50 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9] dark:hover:bg-gray-800"
            >
              Edit
            </button>
            <button
              onClick={() => {
                if (confirm(`Delete ${c.row.original.sku}?`)) deleteMutation.mutate(c.row.original.id);
              }}
              className="rounded-md bg-red-500 px-3 py-1 text-sm text-white transition-colors hover:bg-red-600"
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

  const items = data?.items ?? [];

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
        <SearchBar value={params.q ?? ""} onSearch={(q) => patch({ q })} />
        <div className="flex items-center gap-2">
          <select
            value={params.sort ?? ""}
            onChange={(e) => patch({ sort: e.target.value || undefined })}
            className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9]"
          >
            {SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <a href={exportUrl(params)} className="font-medium text-blue-500 hover:text-blue-600 hover:underline">
            Export Excel
          </a>
        </div>
      </div>

      {data?.suggestions && data.suggestions.length > 0 && (
        <p className="mb-4 text-sm">
          Did you mean:{" "}
          {data.suggestions.map((s) => (
            <button
              key={s}
              onClick={() => patch({ q: s })}
              className="mr-2 rounded-md border border-gray-300 bg-white px-2 py-1 text-sm transition-colors hover:bg-gray-50 dark:border-gray-600 dark:bg-[#1e293b] dark:hover:bg-gray-800"
            >
              {s}
            </button>
          ))}
        </p>
      )}

      <div className="flex flex-col gap-6 lg:flex-row">
        <div className="lg:w-60 lg:shrink-0">
          <Facets facets={data?.facets ?? {}} params={params} onChange={patch} onReset={() => setParams({ page: 1, size: 20 })} />
        </div>

        <div className="flex-1">
          {isLoading && <p className="text-sm text-gray-500 dark:text-gray-400">Loading…</p>}
          {error && (
            <p className="mb-4 rounded-md bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-950 dark:text-red-400">
              {(error as Error).message}
            </p>
          )}

          {!isLoading && !error && items.length === 0 ? (
            <p className="rounded-lg border border-gray-200 bg-white p-6 text-center text-sm text-gray-500 shadow-sm dark:border-gray-700 dark:bg-[#1e293b] dark:text-gray-400">
              No products found.
            </p>
          ) : (
            <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-[#1e293b]">
              <table className="w-full border-collapse">
                <thead>
                  {table.getHeaderGroups().map((hg) => (
                    <tr key={hg.id}>
                      {hg.headers.map((h) => (
                        <th
                          key={h.id}
                          className="border-b border-gray-200 px-3 py-2 text-left text-sm font-semibold dark:border-gray-700"
                        >
                          {flexRender(h.column.columnDef.header, h.getContext())}
                        </th>
                      ))}
                    </tr>
                  ))}
                </thead>
                <tbody>
                  {table.getRowModel().rows.map((row) => (
                    <tr key={row.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                      {row.getVisibleCells().map((cell) => (
                        <td key={cell.id} className="border-b border-gray-100 px-3 py-2 text-sm dark:border-gray-800">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <div className="mt-4 flex items-center gap-3 text-sm">
            <button
              disabled={page <= 1}
              onClick={() => setParams((p) => ({ ...p, page: page - 1 }))}
              className="rounded-md border border-gray-300 bg-white px-3 py-1.5 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-[#1e293b] dark:hover:bg-gray-800"
            >
              Prev
            </button>
            <span>
              Page {page} · {total} results
            </span>
            <button
              disabled={!hasNext}
              onClick={() => setParams((p) => ({ ...p, page: page + 1 }))}
              className="rounded-md border border-gray-300 bg-white px-3 py-1.5 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-[#1e293b] dark:hover:bg-gray-800"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
