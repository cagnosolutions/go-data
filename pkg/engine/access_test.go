package engine

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/engine/buffer"
	"github.com/cagnosolutions/go-data/pkg/engine/page"
	"github.com/cagnosolutions/go-data/pkg/engine/storage"
)

type ExampleRecord struct {
	ID   int
	Data string
}

func (e *ExampleRecord) GetID() uint32 {
	return uint32(e.ID)
}

func (e *ExampleRecord) SetID(id uint32) {
	e.ID = int(id)
}

func (e *ExampleRecord) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

func (e *ExampleRecord) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

var config = &Config{
	BasePath:   "test/engine",
	PageFrames: 64,
}

func TestStorageEngine(t *testing.T) {

	// Open the storage engine
	// se, err := OpenStorageEngine(config)
	// if err != nil {
	// 	t.Errorf("open: %s", err)
	// }

	// Allocate a page
	// p0, err := se.Allocate()
	// if err != nil {
	// 	t.Errorf("allocate: %s", err)
	// }

	// Create something to map record ID's
	// rids := make(map[int]*RecordID)

	// Add 32 records
	// for i := 1; i <= 32; i++ {
	// 	rec := &ExampleRecord{
	// 		ID: i,
	// 		Data: fmt.Sprintf("this is record number %.4d", i),
	// 	}
	// 	rid, err := p0.AddRecord(page.Record{
	//
	// 	})
	// }

}

func TestStorageEngine_Allocate(t *testing.T) {
	type fields struct {
		conf  *Config
		store *storage.DiskStore
		pool  *buffer.BufferPoolManager
	}
	tests := []struct {
		name    string
		fields  fields
		want    page.Page
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &StorageEngine{
					conf:  tt.fields.conf,
					store: tt.fields.store,
					pool:  tt.fields.pool,
				}
				got, err := s.Allocate()
				if (err != nil) != tt.wantErr {
					t.Errorf("Allocate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Allocate() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorageEngine_Close(t *testing.T) {
	type fields struct {
		conf  *Config
		store *storage.DiskStore
		pool  *buffer.BufferPoolManager
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &StorageEngine{
					conf:  tt.fields.conf,
					store: tt.fields.store,
					pool:  tt.fields.pool,
				}
				if err := s.Close(); (err != nil) != tt.wantErr {
					t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestStorageEngine_Fetch(t *testing.T) {
	type fields struct {
		conf  *Config
		store *storage.DiskStore
		pool  *buffer.BufferPoolManager
	}
	type args struct {
		pid page.PageID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    page.Page
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &StorageEngine{
					conf:  tt.fields.conf,
					store: tt.fields.store,
					pool:  tt.fields.pool,
				}
				got, err := s.Fetch(tt.args.pid)
				if (err != nil) != tt.wantErr {
					t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Fetch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorageEngine_Flush(t *testing.T) {
	type fields struct {
		conf  *Config
		store *storage.DiskStore
		pool  *buffer.BufferPoolManager
	}
	type args struct {
		pid page.PageID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &StorageEngine{
					conf:  tt.fields.conf,
					store: tt.fields.store,
					pool:  tt.fields.pool,
				}
				if err := s.Flush(tt.args.pid); (err != nil) != tt.wantErr {
					t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
