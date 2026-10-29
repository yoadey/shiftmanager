package usecase

// pageTotal returns the number of matches behind a page of n rows. The page
// itself proves the total when it is the last one: it ended before the limit
// (or there is no limit) and it is not empty or starts at the very beginning.
// Only otherwise is count (a separate COUNT query) needed.
func pageTotal(n, limit, offset int, count func() (int, error)) (int, error) {
	if (limit <= 0 || n < limit) && (n > 0 || offset <= 0) {
		return max(offset, 0) + n, nil
	}
	return count()
}
