package mapper

import (
	"fmt"

	"github.com/rancher/norman/types"
)

type Embed struct {
	Field          string
	ignoreOverride bool
	embeddedFields []string
}

func (e *Embed) FromInternal(data map[string]interface{}) {
	sub, _ := data[e.Field].(map[string]interface{})
	for _, fieldName := range e.embeddedFields {
		if v, ok := sub[fieldName]; ok {
			data[fieldName] = v
		}
	}
	delete(data, e.Field)
}

func (e *Embed) ToInternal(data map[string]interface{}) {
	sub := map[string]interface{}{}
	for _, fieldName := range e.embeddedFields {
		if v, ok := data[fieldName]; ok {
			sub[fieldName] = v
		}

		delete(data, fieldName)
	}

	data[e.Field] = sub
}

func (e *Embed) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	internalSchema, err := validateInternalField(e.Field, schema)
	if err != nil {
		return err
	}

	e.embeddedFields = []string{}

	embeddedSchemaID := internalSchema.ResourceFields[e.Field].Type
	embeddedSchema := schemas.Schema(&schema.Version, embeddedSchemaID)
	if embeddedSchema == nil {
		return fmt.Errorf("failed to find schema %s for embedding", embeddedSchemaID)
	}

	for name, field := range embeddedSchema.ResourceFields {
		if !e.ignoreOverride {
			if _, ok := schema.ResourceFields[name]; ok {
				return fmt.Errorf("embedding field %s on %s will overwrite the field %s",
					e.Field, schema.ID, name)
			}
		}
		schema.ResourceFields[name] = field
		e.embeddedFields = append(e.embeddedFields, name)
	}

	delete(schema.ResourceFields, e.Field)
	return nil
}
