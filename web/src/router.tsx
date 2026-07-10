import { createRootRoute, createRoute, createRouter, Link, Outlet } from "@tanstack/react-router";
import SearchPage from "./routes/search";
import ProductForm from "./routes/product-form";
import GeneratorPage from "./routes/generator";

function RootLayout() {
  return (
    <div className="min-h-screen bg-[#f8f9fa] text-[#1a202c] dark:bg-[#0f172a] dark:text-[#f1f5f9]">
      <nav className="flex gap-4 border-b border-gray-200 bg-white px-4 py-3 shadow-sm dark:border-gray-700 dark:bg-[#1e293b]">
        <Link to="/" className="font-medium text-blue-500 hover:text-blue-600 [&.active]:underline">
          Search
        </Link>
        <Link to="/new" className="font-medium text-blue-500 hover:text-blue-600 [&.active]:underline">
          New product
        </Link>
        <Link to="/generator" className="font-medium text-blue-500 hover:text-blue-600 [&.active]:underline">
          Generator
        </Link>
      </nav>
      <main className="p-4">
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
