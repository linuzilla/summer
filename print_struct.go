package summer

import (
	"fmt"
	"reflect"
)

func PrintStruct(something interface{}) {
	s := reflect.ValueOf(something).Elem()
	typeOfT := s.Type()
	fmt.Println(">> TypeOf:", reflect.TypeOf(something))
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if f.CanInterface() {
			fmt.Printf(">>   %d: %s %s = %v `%s`", i,
				typeOfT.Field(i).Name,
				f.Type(),
				f.Interface(),
				typeOfT.Field(i).Tag)
		} else {
			fmt.Printf(">>   %d: %s %s `%s`", i,
				typeOfT.Field(i).Name,
				f.Type(),
				typeOfT.Field(i).Tag)
		}

		if typeOfT.Field(i).Anonymous {
			fmt.Println(" (embedded)")
		} else {
			fmt.Println("")
		}
	}
}