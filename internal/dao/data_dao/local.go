package datadao

import (
	"activity_log/api/apperror"
	"activity_log/api/constructs"
	"fmt"
	"os"
	"sort"
)

type DataDAO struct {
	path string
}

func NewDataDAO(path string) *DataDAO {
	return &DataDAO{
		path: path,
	}
}

// TODO(luca): take column headers into consideration
func (dd *DataDAO) Append(data *constructs.UserData) error {
	f, err := os.OpenFile(dd.path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		if apperror.IsNotFoundError(err) {
			// TODO(luca): log this to messenger
			fmt.Printf("No data file found at %sq. Creating one.\n", dd.path)
			f, err = os.OpenFile(dd.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
		} else {
			return fmt.Errorf("error opening file: %w", err)
		}
	}
	defer f.Close()

	keys := []string{}
	for k := range data.Data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	row := ""
	for _, k := range keys {
		row += fmt.Sprintf("%v,", data.Data[k])
	}

	if _, err = f.WriteString(fmt.Sprintf("%d,%s\n", data.TimestampMS, row)); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
