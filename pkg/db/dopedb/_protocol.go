package dopedb

import (
	"fmt"
	"strings"
)

type Cmd struct {
	Op   int
	Argc int
	Argv []string
}

func makeCmd(s string) *Cmd {
	ss := strings.Fields(s)
	if len(ss) < 1 {
		return nil
	}
	if len(ss) == 1 {
		return &Cmd{
			Op:   parseOp(ss[0]),
			Argc: 1,
			Argv: nil,
		}
	}
	return &Cmd{
		Op:   parseOp(ss[0]),
		Argc: len(ss) - 1,
		Argv: ss[1:],
	}
}

const (
	ERR = 1<<iota + 1
	SET
	GET
	ZSET
	ZGET
	HSET
	HGET
	INCR
	DECR
)

var ops = map[string]int{
	"set":  SET,
	"SET":  SET,
	"get":  GET,
	"GET":  GET,
	"zset": ZSET,
	"ZSET": ZSET,
	"zget": ZGET,
	"ZGET": ZGET,
	"hset": HSET,
	"HSET": HSET,
	"incr": INCR,
	"INCR": INCR,
	"decr": DECR,
	"DECR": DECR,
}

func parseOp(s string) int {
	op, found := ops[string(s)]
	if !found {
		return ERR
	}
	return op
}

func (db *DB) Exec(s []byte) ([]byte, error) {
	cmd := makeCmd(string(s))
	if cmd == nil {
		goto ret
	}
	switch cmd.Op {
	case SET:
		return handleSet(db, cmd.Argv)
	case GET:
		return handleGet(db, cmd.Argv)
	case ZSET:
		return handleZset(db, cmd.Argv)
	case ZGET:
		return handleZget(db, cmd.Argv)
	case HSET:
		return handleHset(db, cmd.Argv)
	case HGET:
		return handleHget(db, cmd.Argv)
	case INCR:
		return handleIncr(db, cmd.Argv)
	case DECR:
		return handleDecr(db, cmd.Argv)
	}
ret:
	return nil, fmt.Errorf("invalid command: %q\n", s)
}

var ErrArgc = func(cmd string, atLeast bool, want, got int) error {
	if atLeast {
		return fmt.Errorf(
			"syntax error: wong number of arguments for %q (wanted at least %d, received %d)", cmd, want,
			got,
		)
	}
	return fmt.Errorf("syntax error: wong number of arguments for %q (wanted %d, received %d)", cmd, want, got)
}

func handleSet(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 2 {
		return nil, ErrArgc("set", false, 2, len(s))
	}
	db.set(s[0], s[1])
	return []byte("OK"), nil
}

func handleGet(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 1 {
		return nil, ErrArgc("get", false, 1, len(s))
	}
	return []byte(db.get(s[0])), nil
}

func handleZset(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) < 2 {
		return nil, ErrArgc("zset", true, 2, len(s))
	}
	db.zset(s[0], s[1:]...)
	return []byte("OK"), nil
}

func handleZget(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 1 {
		return nil, ErrArgc("zget", false, 1, len(s))
	}
	return []byte(strings.Join(db.zget(s[0]), " ")), nil
}

func handleHset(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) < 2 {
		return nil, ErrArgc("hset", true, 2, len(s))
	}
	db.hset(s[0], s[1:]...)
	return []byte("OK"), nil
}

func handleHget(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 1 {
		return nil, ErrArgc("hget", false, 1, len(s))
	}
	var ss []string
	for k, v := range db.hget(s[0]) {
		ss = append(ss, k+":"+v)
	}
	return []byte(strings.Join(ss, " ")), nil
}

func handleIncr(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 1 {
		return nil, ErrArgc("incr", false, 1, len(s))
	}
	return []byte(db.incr(s[0])), nil
}

func handleDecr(db *DB, s []string) ([]byte, error) {
	if s == nil || len(s) != 1 {
		return nil, ErrArgc("decr", false, 1, len(s))
	}
	return []byte(db.decr(s[0])), nil
}

func handleDefault(db *DB, s []string) ([]byte, error) {
	return []byte(fmt.Sprintf("got: %v\n", s)), nil
}
