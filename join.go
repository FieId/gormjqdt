package gormjqdt

import "reflect"

type RelationMappings struct {
	ModelSchemaNames map[int]string
	DbColumns        map[int]map[int]string
	DbColumnTypes    map[int]map[string]reflect.Kind
}

func RelationSetters(givenRelations map[int]interface{}) *RelationMappings {
	relMaps := &RelationMappings{}

	// Set the model schema names
	relMaps.ModelSchemaNames = GetPointerName(givenRelations)

	// Set DB columns
	relMaps.DbColumns = make(map[int]map[int]string)
	for kc, vc := range givenRelations {
		relMaps.DbColumns[kc] = GetDbColumns(vc)
	}

	// Set DB column types
	relMaps.DbColumnTypes = make(map[int]map[string]reflect.Kind)
	for kct, vct := range givenRelations {
		relMaps.DbColumnTypes[kct] = GetDbColumnTypes(vct)
	}

	return relMaps
}
