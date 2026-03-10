package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"sync"
	"unsafe"

	"truco-tui/internal/appcore"
)

var (
	handleMu   sync.Mutex
	nextHandle uintptr = 1
	runtimes           = map[uintptr]*appcore.Runtime{}
)

func main() {}

//export TrucoCoreCreate
func TrucoCoreCreate() C.uintptr_t {
	rt := appcore.NewRuntime()
	handleMu.Lock()
	id := nextHandle
	nextHandle++
	runtimes[id] = rt
	handleMu.Unlock()
	return C.uintptr_t(id)
}

//export TrucoCoreDestroy
func TrucoCoreDestroy(handle C.uintptr_t) {
	rt := getRuntime(uintptr(handle))
	if rt != nil {
		_ = rt.Close()
	}
	handleMu.Lock()
	delete(runtimes, uintptr(handle))
	handleMu.Unlock()
}

//export TrucoCoreDispatchIntentJSON
func TrucoCoreDispatchIntentJSON(handle C.uintptr_t, payload *C.char) *C.char {
	rt := getRuntime(uintptr(handle))
	if rt == nil {
		return encodeError("invalid_handle", "runtime não encontrado")
	}
	var intent appcore.AppIntent
	if err := json.Unmarshal([]byte(C.GoString(payload)), &intent); err != nil {
		return encodeError("invalid_json", err.Error())
	}
	if err := rt.DispatchIntent(intent); err != nil {
		return encodeError("dispatch_failed", err.Error())
	}
	return nil
}

//export TrucoCorePollEventJSON
func TrucoCorePollEventJSON(handle C.uintptr_t) *C.char {
	rt := getRuntime(uintptr(handle))
	if rt == nil {
		return encodeError("invalid_handle", "runtime não encontrado")
	}
	ev, ok := rt.PollEvent()
	if !ok {
		return nil
	}
	return mustJSON(ev)
}

//export TrucoCoreSnapshotJSON
func TrucoCoreSnapshotJSON(handle C.uintptr_t) *C.char {
	rt := getRuntime(uintptr(handle))
	if rt == nil {
		return encodeError("invalid_handle", "runtime não encontrado")
	}
	s, err := rt.SnapshotJSON()
	if err != nil {
		return encodeError("snapshot_failed", err.Error())
	}
	return C.CString(s)
}

//export TrucoCoreVersionsJSON
func TrucoCoreVersionsJSON() *C.char {
	return mustJSON(appcore.NewRuntime().Versions())
}

//export TrucoCoreFreeString
func TrucoCoreFreeString(ptr *C.char) {
	if ptr == nil {
		return
	}
	C.free(unsafe.Pointer(ptr))
}

func getRuntime(handle uintptr) *appcore.Runtime {
	handleMu.Lock()
	defer handleMu.Unlock()
	return runtimes[handle]
}

func mustJSON(v any) *C.char {
	b, err := json.Marshal(v)
	if err != nil {
		return encodeError("marshal_failed", err.Error())
	}
	return C.CString(string(b))
}

func encodeError(code, message string) *C.char {
	return mustJSON(map[string]string{
		"code":    code,
		"message": message,
	})
}
