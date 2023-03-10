package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"unsafe"
)

type Tuple[L any, K any] struct {
	Left  L
	Right K
}

type List[T any] struct {
	Value T
	Next  *List[T]
}

type RouterHandle struct {
	Lock           *sync.Mutex
	Router         *List[*Tuple[string, *Tuple[string, http.HandlerFunc]]]
	NotFoundHandle http.HandlerFunc
}

func (routerHandle *RouterHandle) FindHandleByPath(path string, method string) *Tuple[string, *Tuple[string, http.HandlerFunc]] {
	tmp := routerHandle.Router

	for tmp != nil {

		if tmp.Value.Left == path && tmp.Value.Right.Left == method {
			return tmp.Value
		}

		tmp = tmp.Next
	}

	return nil
}

func (routerHandle *RouterHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Server", "smilex")

	handle := routerHandle.FindHandleByPath(r.RequestURI, r.Method)
	if handle != nil {
		handle.Right.Right(w, r)
		return
	}

	handle = routerHandle.FindHandleByPath(r.RequestURI[1:], r.Method)
	if handle != nil {
		handle.Right.Right(w, r)
		return
	}

	routerHandle.NotFoundHandle(w, r)
}

func (routerHandle *RouterHandle) AddRouterHandle(path string, method string, handle http.HandlerFunc) {
	routerHandle.Lock.Lock()
	defer routerHandle.Lock.Unlock()

	if routerHandle.FindHandleByPath(path, method) == nil {
		right := new(Tuple[string, http.HandlerFunc])
		right.Left = method
		right.Right = handle

		tuple := new(Tuple[string, *Tuple[string, http.HandlerFunc]])
		tuple.Left = path
		tuple.Right = right

		list := new(List[*Tuple[string, *Tuple[string, http.HandlerFunc]]])
		list.Value = tuple

		if routerHandle.Router != nil {
			list.Next = routerHandle.Router
		}

		routerHandle.Router = list
	}
}

func NewRouterHandle(notFountHandle http.HandlerFunc) *RouterHandle {
	routerHandle := new(RouterHandle)
	routerHandle.Lock = new(sync.Mutex)
	routerHandle.NotFoundHandle = notFountHandle
	return routerHandle
}

func StringConvToByteSlice(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func main() {
	router := NewRouterHandle(func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{}, 2)
		data["code"] = 404
		data["msg"] = "Not Found"

		respData, _ := json.Marshal(data)
		_, _ = w.Write(respData)
	})

	router.AddRouterHandle("/", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.Write(StringConvToByteSlice("Hello, World"))
	})

	router.AddRouterHandle("/json", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf")
		data := make(map[string]interface{}, 1)
		data["msg"] = "hello"

		respData, _ := json.Marshal(data)
		_, _ = w.Write(respData)
	})

	err := http.ListenAndServe("0.0.0.0:9000", router)
	if err != nil {
		panic(err)
	}
}
