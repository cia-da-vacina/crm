/** Shared API origin for fetch client + MSW handlers. */
export const API_BASE =
  (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1").replace(
    /\/$/,
    "",
  );
