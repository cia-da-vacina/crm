/** Offset/page-based list response. */
export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

/** Cursor-based list response, used for high-volume/real-time feeds (inbox, messages). */
export interface CursorPage<T> {
  items: T[];
  next_cursor: string | null;
}

/** Standard error body returned by the backend (and mirrored by the BFF). */
export interface ApiErrorBody {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}
