import { createBrowserRouter, RouterProvider } from "react-router-dom";
import App from "./App";
import DetailPage from "./screens/DetailPage";

const router = createBrowserRouter([
  { path: "/", element: <App /> },
  { path: "/details", element: <DetailPage /> },
]);

export default function AppRouter() {
  return (
    <RouterProvider router={router} future={{ v7_startTransition: true }} />
  );
}
