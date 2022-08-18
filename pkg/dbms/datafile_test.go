package dbms

import (
	"testing"
)

func TestDataFile_All(t *testing.T) {
	// open a data file
	df, err := OpenDataFile("testing", 0)
	if err != nil {
		t.Error(err)
	}
	defer func(df *DataFile) {
		err := df.Close()
		if err != nil {
			t.Error(err)
		}
	}(df)
}
