package sqlite

func calcNumPages(requestPageSize int, count int64) (numPages int, pageSize int) {
	pageSize = (requestPageSize)
	if pageSize <= 0 {
		pageSize = int(count)
	}

	if pageSize <= 0 {
		return 1, int(count)
	}

	return divCeil(int(count), pageSize), pageSize
}

func divCeil(a, b int) int {
	mod := a % b
	if mod == 0 {
		return a / b
	}

	return a/b + 1
}
