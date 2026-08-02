/**
 * Cookie name constants shared by the Edge middleware and the Node.js
 * route handlers. Kept in a tiny module with zero runtime deps so the
 * Edge bundle never pulls in `next/headers` or server env helpers.
 */
export const COOKIE_ACCESS = "cv_access";
export const COOKIE_REFRESH = "cv_refresh";
export const COOKIE_UNIT = "cv_unit";
