import { createRootRoute, createRoute, createRouter, Link, Outlet } from "@tanstack/react-router";
import SearchPage from "./routes/search";
import ProductForm from "./routes/product-form";
import GeneratorPage from "./routes/generator";

function RootLayout() {
  return (
    <div>
      <nav style={{ display: "flex", gap: 16, padding: 12, borderBottom: "1px solid #ddd" }}>
        <Link to="/">Search</Link>
        <Link to="/new">New product</Link>
        <Link to="/generator">Generator</Link>
      </nav>
      <main style={{ padding: 16 }}>
        <Outlet />
      </main>
    </div>
  );
}

const rootRoute = createRootRoute({ component: RootLayout });

const searchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: SearchPage,
});

const newRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/new",
  component: ProductForm,
});

const editRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/edit/$id",
  component: ProductForm,
});

const generatorRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/generator",
  component: GeneratorPage,
});

const routeTree = rootRoute.addChildren([searchRoute, newRoute, editRoute, generatorRoute]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
