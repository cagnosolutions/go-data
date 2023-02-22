package ember

import (
	"fmt"
	"log"
	"time"
)

type EmberDB struct {
	db  *shardedHashMap
	aof *aof
}

func (edb *EmberDB) Open(path string) (*EmberDB, error) {
	f, err := open(path)
	if err != nil {
		return nil, err
	}
	db := &EmberDB{
		db:  newShardedHashMap(64, nil),
		aof: f,
	}
	done := make(chan bool)
	background(
		done, func() {
			if err := db.aof.checkPrune(); err != nil {
				panic(err)
			}
		},
	)
	return db, nil
}

func background(done <-chan bool, f func()) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-done:
				log.Print("Stopping ticker")
				ticker.Stop()
				return
			case <-ticker.C:
				f()
				fmt.Println("TICK!")
			}
		}
	}()
}
