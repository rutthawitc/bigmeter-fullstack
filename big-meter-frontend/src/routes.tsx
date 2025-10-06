import { createBrowserRouter, RouterProvider } from "react-router-dom";
import App from "./App";
import DetailPage from "./screens/DetailPage";
import AdminPage from "./screens/AdminPage";

const router = createBrowserRouter([
  { path: "/", element: <App /> },
  { path: "/details", element: <DetailPage /> },
  { path: "/admin", element: <AdminPage /> },
]);

export default function AppRouter() {
  return (
    <RouterProvider router={router} future={{ v7_startTransition: true }} />
  );
}
