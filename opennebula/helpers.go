package opennebula

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
