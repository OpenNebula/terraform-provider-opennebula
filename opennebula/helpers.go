package opennebula

import (
	"fmt"
	"strings"
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
