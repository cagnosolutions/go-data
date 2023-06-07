package dopedb

import (
	"testing"
	"time"
)

func TestDBServer(t *testing.T) {
	db, err := NewDB(nil)
	if err != nil {
		t.Errorf("open db: %s", err)
	}
	srv := NewDBServer(db, 5*time.Minute)
	err = srv.ListenAndServe(":9999")
	if err != nil {
		t.Errorf("listen and serve: %s", err)
	}
}
