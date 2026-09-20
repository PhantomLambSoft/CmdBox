package handlers

type ListHistoryArgs struct {
	Limit    int
	HasLimit bool
	Alias    string
	HasAlias bool
}

type ClearHistoryArgs struct {
	Yes      bool
	Alias    string
	HasAlias bool
}

func RunListHistory(args ListHistoryArgs) error {
	return nil
}

func RunShowHistory(indexOrID string) error {
	return nil
}

func RunRerunHistory(indexOrID string) error {
	return nil
}

func RunLastHistory() error {
	return nil
}

func RunClearHistory(args ClearHistoryArgs) error {
	return nil
}
