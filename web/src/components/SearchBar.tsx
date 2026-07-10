import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { autocomplete } from "../api";

interface Props {
  value: string;
  onSearch: (q: string) => void;
}

export default function SearchBar({ value, onSearch }: Props) {
  const [text, setText] = useState(value);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const { data } = useQuery({
    queryKey: ["autocomplete", text],
    queryFn: () => autocomplete(text),
    enabled: text.length >= 2,
  });

  const suggestions = data?.suggestions ?? [];

  function submit(q: string) {
    setText(q);
    setShowSuggestions(false);
    onSearch(q);
  }

  return (
    <div className="relative max-w-[480px]">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          submit(text);
        }}
        className="flex gap-2"
      >
        <input
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          onBlur={() => setTimeout(() => setShowSuggestions(false), 150)}
          placeholder="Search products…"
          className="flex-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-[#1a202c] focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-[#1e293b] dark:text-[#f1f5f9]"
        />
        <button
          type="submit"
          className="rounded-md bg-blue-500 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-600"
        >
          Search
        </button>
      </form>
      {showSuggestions && suggestions.length > 0 && (
        <ul className="absolute left-0 right-0 top-full z-10 mt-1 list-none rounded-md border border-gray-200 bg-white p-1 shadow-lg dark:border-gray-700 dark:bg-[#1e293b]">
          {suggestions.map((s) => (
            <li
              key={s}
              className="cursor-pointer rounded px-2 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-800"
              onMouseDown={() => submit(s)}
            >
              {s}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
