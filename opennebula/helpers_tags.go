package opennebula

import (
	"log"
	"strings"

	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
)

func pairsToMap(tpl dyn.Template) map[string]interface{} {

	tags := make(map[string]interface{})

	for i, _ := range tpl.Elements {
		pair, ok := tpl.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		tags[pair.Key()] = pair.Value
	}

	return tags
}

func pairsToMapFilter(tpl dyn.Template, elements map[string]interface{}) map[string]interface{} {

	tags := make(map[string]interface{})

	for k, v := range elements {
		uK := strings.ToUpper(k)
		pair, err := tpl.GetPair(uK)
		if err != nil {
			continue
		}

		if pair.Value != v {
			continue
		}

		tags[k] = pair.Value
	}

	return tags
}

func matchTags(tpl dyn.Template, tags map[string]interface{}) bool {

	for k, v := range tags {
		uK := strings.ToUpper(k)
		pair, err := tpl.GetPair(uK)
		if err != nil {
			log.Printf("[DEBUG] tag key %s: %s\n", uK, err)
			return false
		}

		if pair.Value != v {
			log.Printf("[DEBUG] tag value %s: doesn't match\n", v)
			return false
		}
	}

	return true
}
