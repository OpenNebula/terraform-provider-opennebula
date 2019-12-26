package opennebula

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func inArray(val string, array []string) (index int) {
	var ok bool
	for i := range array {
		if ok = array[i] == val; ok {
			return i
		}
	}
	return -1
}

// appendTemplate add attribute and value to an existing string
func appendTemplate(template, attribute, value string) string {
	return fmt.Sprintf("%s\n%s = \"%s\"", template, attribute, value)
}

func ArrayToString(list []interface{}, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(list)), delim), "[]")
}

func StringToLockLevel(str string, lock *shared.LockLevel) error {
	if str == "USE" {
		*lock = shared.LockUse
		return nil
	}
	if str == "MANAGE" {
		*lock = shared.LockManage
		return nil
	}
	if str == "ADMIN" {
		*lock = shared.LockAdmin
		return nil
	}
	if str == "ALL" {
		*lock = shared.LockAll
		return nil
	}
	return fmt.Errorf("Unexpected Lock level %s", str)
}

func LockLevelToString(lock int) string {
	if lock == 1 {
		return "USE"
	}
	if lock == 2 {
		return "MANAGE"
	}
	if lock == 3 {
		return "ADMIN"
	}
	if lock == 4 {
		return "ALL"
	}
	return ""
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
