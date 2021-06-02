package summer

import (
	"container/list"
	"reflect"
)

type stream struct {
	channel chan interface{}
}

func Stream(dataSlice interface{}) *stream {
	s := &stream{channel: make(chan interface{})}

	go func() {
		for _, data := range interfaceSlice(dataSlice) {
			s.channel <- data
		}
		close(s.channel)
	}()
	return s
}

func StreamOfList(aList *list.List) *stream {
	s := &stream{channel: make(chan interface{})}

	go func() {
		for e := aList.Front(); e != nil; e = e.Next() {
			s.channel <- e.Value
		}
		close(s.channel)
	}()
	return s
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func (inputStream *stream) ForEach(accept func(i interface{})) {
	for data := range inputStream.channel {
		accept(data)
	}
}

func (inputStream *stream) Filter(predicate func(i interface{}) bool) *stream {
	outputStream := &stream{channel: make(chan interface{})}

	go func() {
		for data := range inputStream.channel {
			if predicate(data) {
				outputStream.channel <- data
			}
		}
		close(outputStream.channel)
	}()

	return outputStream
}

func (inputStream *stream) Map(apply func(i interface{}) interface{}) *stream {
	outputStream := &stream{channel: make(chan interface{})}

	go func() {
		for data := range inputStream.channel {
			outputStream.channel <- apply(data)
		}
		close(outputStream.channel)
	}()

	return outputStream
}
