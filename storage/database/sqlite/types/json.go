package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type SQLiteJSON[J any] struct {
	JSON J
}

func NewSQLiteJSON[J any](val J) SQLiteJSON[J] {
	return SQLiteJSON[J]{JSON: val}
}

var _ sql.Scanner = (*SQLiteJSON[any])(nil)
var _ driver.Valuer = (*SQLiteJSON[any])(nil)

func (sj *SQLiteJSON[J]) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid input type for converting json %T", src)
	}

	return json.Unmarshal([]byte(str), &sj.JSON)
}

func (sj SQLiteJSON[J]) Value() (driver.Value, error) {
	j, err := json.Marshal(sj.JSON)
	return driver.Value(string(j)), err
}
