package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func contains(value string, values []string) bool {
	for i := range values {
		if values[i] == value {
			return true
		}
	}
	return false
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

func emptyOrEqual(config interface{}, value interface{}) bool {
	return isEmptyValue(reflect.ValueOf(config)) || value == config
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

// NoExists indicate if an entity exists in checking the error code returned from an Info call
func NoExists(err error) bool {

	respErr, ok := err.(*errors.ResponseError)

	// expected case, the entity does not exists so we doesn't return an error
	if ok && respErr.Code == errors.OneNoExistsError {
		return true
	}

	return false
}

// returns the diff of two lists of schemas, making diff on attrNames only
func diffListConfig(refVecs, vecs []interface{}, s *schema.Resource, attrNames ...string) ([]interface{}, []interface{}) {

	// remove schema fields that are not listed in attrNames
	for scKey := range s.Schema {
		present := false
		for _, attrName := range attrNames {
			if scKey == attrName {
				present = true
				break
			}
		}
		if !present {
			delete(s.Schema, scKey)
		}
	}

	refSet := schema.NewSet(schema.HashResource(s), []interface{}{})
	for _, iface := range refVecs {
		sc := iface.(map[string]interface{})

		refSet.Add(sc)
	}

	set := schema.NewSet(schema.HashResource(s), []interface{}{})
	for _, iface := range vecs {
		sc := iface.(map[string]interface{})

		set.Add(sc)
	}

	pSet := refSet.Difference(set)
	mSet := set.Difference(refSet)

	return mSet.List(), pSet.List()
}

func mergeSchemas(schema map[string]*schema.Schema, schemas ...map[string]*schema.Schema) map[string]*schema.Schema {
	if len(schemas) == 0 {
		return schema
	}

	for _, m := range schemas {
		for k, v := range m {
			schema[k] = v
		}
	}

	return schema
}

func ParseIntFromInterface(i interface{}) (int, error) {
	switch v := i.(type) {
	case float64:
		return int(v), nil
	case string:
		if r, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int(r), nil
		}
	}
	return -1, fmt.Errorf("Does not look like a number")
}

func ArraysAreEqual[T comparable](a, b []T) bool {
    if len(a) != len(b) {
        return false
    }

    mapA := make(map[T]int)
    mapB := make(map[T]int)

    for _, v := range a {
        mapA[v]++
    }

    for _, v := range b {
        mapB[v]++
    }

    for k, v := range mapA {
        if mapB[k] != v {
            return false
        }
    }

    return true
}


// ArrayDifference returns the elements that are in "src" but not in "other"
func ArrayDifference[T comparable](src, other []T) []T {
    counter := map[T]int{}

    for _, elem := range other {
        counter[elem]++
    }

    diff := []T{}
    for _, elem := range src {
        if counter[elem] == 0 {
            diff = append(diff, elem)
        }
    }

    return diff
}


// returns the elements from src that are not present in other based on a key value
func MapArrayDifferenceByKeyValue (src, other []any, key string) ([]any, error) {
    if len(src) == 0 {
        return nil, nil
    }

    counter := map[string]int{}
    for _, elem := range other {
        v, ok := elem.(map[string]any)
        if !ok {
            return nil, fmt.Errorf("element %v is not a map", elem)
        }
        if val, exists := v[key]; exists {
            counter[fmt.Sprintf("%v", val)]++
        }
    }

    diff := []any{}
    for _, elem := range src {
        v, ok := elem.(map[string]any)
        if !ok {
            return nil, fmt.Errorf("element %v is not a map", elem)
        }
        if val, exists := v[key]; exists {
            if counter[fmt.Sprintf("%v", val)] == 0 {
                diff = append(diff, elem)
            }
        } else {
            diff = append(diff, elem)
        }
    }
    return diff, nil
}


// returns the elements from src that has the same key value as existing elements in other
func MapArrayIntersectionByKeyValue(src, other[]any, key string)([]any, error) {

    if len(src) == 0 || len(other) == 0 {
        return nil, nil
    }

    counter := map[string]int{}
    for _, elem := range other {
        v, ok := elem.(map[string]any)
        if !ok {
            return nil, fmt.Errorf("Element %v is not a map", elem)
        }
        if val, exists := v[key]; exists {
            counter[fmt.Sprintf("%v", val)]++
        }
    }

    intersection := []any{}
    for _, elem := range src {
        v, ok := elem.(map[string]any)
        if !ok {
            return nil, fmt.Errorf("Element %v is not a map", elem)
        }
        if val, exists := v[key]; exists {
            if counter[fmt.Sprintf("%v", val)] > 0 {
                intersection = append(intersection, elem)
            }
        }
    }
    return intersection, nil
}

// Generic function to check if a value is present in array
func ValueInArray[T comparable](value T, array []T) bool {
    for _, v := range array {
        if v == value {
            return true
        }
    }
    return false
}
