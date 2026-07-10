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
    <div style={{ position: "relative", maxWidth: 480 }}>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          submit(text);
        }}
        style={{ display: "flex", gap: 8 }}
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
          style={{ flex: 1, padding: 8 }}
        />
        <button type="submit">Search</button>
      </form>
      {showSuggestions && suggestions.length > 0 && (
        <ul
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            background: "#fff",
            border: "1px solid #ddd",
            listStyle: "none",
            margin: 0,
            padding: 4,
            zIndex: 10,
          }}
        >
          {suggestions.map((s) => (
            <li key={s} style={{ padding: 4, cursor: "pointer" }} onMouseDown={() => submit(s)}>
              {s}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
