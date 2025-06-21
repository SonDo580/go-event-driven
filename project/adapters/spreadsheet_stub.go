package adapters

import (
	"context"
	"slices"
	"sync"
)

type SpreadsheetsAPIStub struct {
	lock sync.Mutex
	Rows map[string][][]string // spreadsheetName -> rows
}

func (stub *SpreadsheetsAPIStub) AppendRow(
	ctx context.Context, spreadsheetName string, row []string,
) error {
	stub.lock.Lock()
	defer stub.lock.Unlock()

	if stub.Rows == nil {
		stub.Rows = make(map[string][][]string)
	}

	copiedRow := slices.Clone(row)
	stub.Rows[spreadsheetName] = append(stub.Rows[spreadsheetName], copiedRow)

	return nil
}
