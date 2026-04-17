package ginlogctx

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

var scopedFields sync.Map

func bindFields(fields logrus.Fields) func() {
	gid := currentGoroutineID()
	if gid == 0 {
		return func() {}
	}

	scopedFields.Store(gid, cloneFields(fields))
	return func() {
		scopedFields.Delete(gid)
	}
}

func currentFields() logrus.Fields {
	gid := currentGoroutineID()
	if gid == 0 {
		return nil
	}

	fields, ok := scopedFields.Load(gid)
	if !ok {
		return nil
	}

	scoped, ok := fields.(logrus.Fields)
	if !ok {
		return nil
	}

	return cloneFields(scoped)
}

func cloneFields(fields logrus.Fields) logrus.Fields {
	if len(fields) == 0 {
		return nil
	}

	cloned := make(logrus.Fields, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}

func currentGoroutineID() uint64 {
	var stack [64]byte
	n := runtime.Stack(stack[:], false)
	if n == 0 {
		return 0
	}

	const prefix = "goroutine "
	line := stack[:n]
	if !bytes.HasPrefix(line, []byte(prefix)) {
		return 0
	}

	line = line[len(prefix):]
	stop := bytes.IndexByte(line, ' ')
	if stop < 0 {
		return 0
	}

	gid, err := strconv.ParseUint(string(line[:stop]), 10, 64)
	if err != nil {
		return 0
	}

	return gid
}
