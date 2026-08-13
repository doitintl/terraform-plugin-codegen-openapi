// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oas

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/util"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var ErrMultiTypeSchema = errors.New("unsupported multi-type, attribute cannot be created")
var ErrSchemaNotFound = errors.New("no compatible schema found")

// BuildSchemaFromRequest will extract and build the schema from the request body of an operation
//   - Media type will default to "application/json", then continue to the next available media type with a schema
func BuildSchemaFromRequest(op *high.Operation, schemaOpts SchemaOpts, globalOpts GlobalSchemaOpts) (*OASSchema, error) {
	if op == nil || op.RequestBody == nil || op.RequestBody.Content == nil || op.RequestBody.Content.Len() == 0 {
		return nil, ErrSchemaNotFound
	}

	return getSchemaFromMediaType(op.RequestBody.Content, schemaOpts, globalOpts)
}

// BuildSchemaFromResponse will extract and build the schema from the response body of an operation
//   - Response codes of 200 and then 201 will be prioritized, then will continue to the next available 2xx code
//   - Media type will default to "application/json", then continue to the next available media type with a schema
func BuildSchemaFromResponse(op *high.Operation, schemaOpts SchemaOpts, globalOpts GlobalSchemaOpts) (*OASSchema, error) {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil || op.Responses.Codes.Len() == 0 {
		return nil, ErrSchemaNotFound
	}

	okResponse, ok := op.Responses.Codes.Get(util.OAS_response_code_ok)
	if ok {
		return getSchemaFromMediaType(okResponse.Content, schemaOpts, globalOpts)
	}

	createdResponse, ok := op.Responses.Codes.Get(util.OAS_response_code_created)
	if ok {
		return getSchemaFromMediaType(createdResponse.Content, schemaOpts, globalOpts)
	}

	sortedCodes := orderedmap.SortAlpha(op.Responses.Codes)
	for pair := range orderedmap.Iterate(context.TODO(), sortedCodes) {
		responseCode := pair.Value()
		statusCode, err := strconv.Atoi(pair.Key())
		if err != nil {
			continue
		}

		if statusCode >= 200 && statusCode <= 299 {
			return getSchemaFromMediaType(responseCode.Content, schemaOpts, globalOpts)
		}
	}

	return nil, ErrSchemaNotFound
}

func getSchemaFromMediaType(mediaTypes *orderedmap.Map[string, *high.MediaType], schemaOpts SchemaOpts, globalOpts GlobalSchemaOpts) (*OASSchema, error) {
	if mediaTypes == nil {
		return nil, ErrSchemaNotFound
	}

	jsonMediaType, ok := mediaTypes.Get(util.OAS_mediatype_json)
	if ok && jsonMediaType.Schema != nil {
		s, err := BuildSchema(jsonMediaType.Schema, schemaOpts, globalOpts)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	sortedMediaTypes := orderedmap.SortAlpha(mediaTypes)
	for pair := range orderedmap.Iterate(context.TODO(), sortedMediaTypes) {
		mediaType := pair.Value()
		if mediaType.Schema != nil {
			s, err := BuildSchema(mediaType.Schema, schemaOpts, globalOpts)
			if err != nil {
				return nil, err
			}
			return s, nil
		}
	}

	return nil, ErrSchemaNotFound
}

// BuildSchema will build a schema from a schema proxy. It can also handle nullable schemas/types,
// implemented with oneOf/anyOf OAS keywords or an array on the "type" property
func BuildSchema(proxy *base.SchemaProxy, schemaOpts SchemaOpts, globalOpts GlobalSchemaOpts) (*OASSchema, *SchemaError) {
	resp := OASSchema{}

	s, err := buildSchemaProxy(proxy)
	if err != nil {
		return nil, err
	}

	resp.SchemaOpts = schemaOpts
	resp.GlobalSchemaOpts = globalOpts
	resp.Schema = s

	oasType, err := retrieveType(resp.Schema)
	if err != nil {
		return nil, err
	}

	resp.Type = oasType
	resp.Format = resp.Schema.Format

	return &resp, nil
}

// buildSchemaProxy is a helper function that builds a schema proxy. If needed, it will recursively resolve a specific set of [schema composition] keywords:
//   - allOf: If len == 1, will resolve with that one item.
//   - anyOf: If len == 2, will resolve nullable or stringable types
//   - oneOf: If len == 2, will resolve nullable or stringable types
//
// # Any other combinations of allOf, anyOf, or oneOf will return a SchemaError
//
// [schema composition]: https://json-schema.org/understanding-json-schema/reference/combining
func buildSchemaProxy(proxy *base.SchemaProxy) (*base.Schema, *SchemaError) {
	s, err := proxy.BuildSchema()
	if err != nil {
		return nil, SchemaErrorFromProxy(fmt.Errorf("failed to build schema proxy - %w", err), proxy)
	}

	// If there are no schema composition keywords, return the schema
	if len(s.AllOf) == 0 && len(s.AnyOf) == 0 && len(s.OneOf) == 0 {
		return s, nil
	}

	if len(s.AnyOf) > 0 {
		if len(s.AnyOf) == 2 {
			schema, err := getMultiTypeSchema(s, s.AnyOf[0], s.AnyOf[1])
			if err != nil {
				return nil, err
			}

			return schema, nil
		}

		// Dynamic type currently not supported
		return nil, SchemaErrorFromNode(fmt.Errorf("found %d anyOf subschema(s), schema composition is currently not supported", len(s.AnyOf)), s, AnyOf)
	}

	if len(s.OneOf) > 0 {
		if len(s.OneOf) == 2 {
			schema, err := getMultiTypeSchema(s, s.OneOf[0], s.OneOf[1])
			if err != nil {
				return nil, err
			}

			return schema, nil
		}

		// Dynamic type currently not supported
		return nil, SchemaErrorFromNode(fmt.Errorf("found %d oneOf subschema(s), schema composition is currently not supported", len(s.OneOf)), s, OneOf)
	}

	// If there is just one allOf, we can use it as the schema, with any keywords
	// set alongside the allOf overlaid on top of it.
	if len(s.AllOf) == 1 {
		allOfSchema, err := buildSchemaProxy(s.AllOf[0])
		if err != nil {
			return nil, err
		}

		return overlaySiblings(s, allOfSchema, overlayTypeInfo), nil
	}

	// Combining multiple allOf schemas and their properties is possible here, but currently not supported
	// See: https://github.com/doitintl/terraform-plugin-codegen-openapi/issues/56
	return nil, SchemaErrorFromNode(fmt.Errorf("found %d allOf subschema(s), schema composition is currently not supported", len(s.AllOf)), s, AllOf)
}

// getMultiTypeSchema will check the types of both schemas provided and will return the non-null schema, with any
// keywords set alongside the oneOf/anyOf on the composing schema overlaid on top of it. If a null schema type is not
// detected, an error will be returned as multi-types are not supported
func getMultiTypeSchema(composer *base.Schema, proxyOne *base.SchemaProxy, proxyTwo *base.SchemaProxy) (*base.Schema, *SchemaError) {
	firstSchema, err := buildSchemaProxy(proxyOne)
	if err != nil {
		return nil, err
	}

	secondSchema, err := buildSchemaProxy(proxyTwo)
	if err != nil {
		return nil, err
	}

	firstType, err := retrieveType(firstSchema)
	if err != nil {
		return nil, err
	}

	secondType, err := retrieveType(secondSchema)
	if err != nil {
		return nil, err
	}

	// Check for null type, if found, return the other type
	if firstType == util.OAS_type_null {
		return overlaySiblings(composer, secondSchema, dontOverlayTypeInfo), nil
	} else if secondType == util.OAS_type_null {
		return overlaySiblings(composer, firstSchema, dontOverlayTypeInfo), nil
	}

	// Check for string type, if the other type can be represented as a string, return the string type
	if firstType == util.OAS_type_string && isStringableType(secondType) {
		return overlaySiblings(composer, firstSchema, dontOverlayTypeInfo), nil
	} else if secondType == util.OAS_type_string && isStringableType(firstType) {
		return overlaySiblings(composer, secondSchema, dontOverlayTypeInfo), nil
	}

	return nil, SchemaErrorFromNode(fmt.Errorf("[%s %s] - %w", firstType, secondType, ErrMultiTypeSchema), firstSchema, Type)
}

// Arguments for overlaySiblings' allowTypeInfo parameter.
const (
	// overlayTypeInfo carries `type` and `format` over from the composing schema. Correct for allOf, where the
	// composing schema and the subschema describe the same value.
	overlayTypeInfo = true

	// dontOverlayTypeInfo leaves the resolved subschema's `type` and `format` alone. Correct for oneOf/anyOf, where
	// the composing schema describes the union rather than the branch that was selected from it.
	dontOverlayTypeInfo = false
)

// overlaySiblings applies the keywords set alongside a schema composition keyword ("sibling keys") on the composing
// schema to the subschema that the composition resolved to.
//
// Only keywords this generator actually reads are carried over. Everything else -- nullable, readOnly, title,
// examples, extensions and so on -- is deliberately dropped, because carrying it would have no observable effect.
//
// Merge rules:
//   - description, deprecated, default, enum, pattern and the numeric/size constraints: the composing schema wins
//     when it sets them. allOf is formally an intersection, so a composing schema that widens a bound is
//     technically incorrect here, but in practice these siblings are only ever written to refine the referenced
//     schema at one use site, and they feed generated Terraform validators rather than a validator of record.
//   - required: unioned, never replaced. Replacing would discard the referenced schema's own required list and
//     silently demote real attributes from required to computed_optional.
//   - properties: merged into a fresh map, with the composing schema winning on key collisions.
//   - type: only fills a gap. `type` selects which attribute kind is generated, and because allOf is an
//     intersection a type declared on both sides has to agree -- so when they disagree the $ref target is the
//     authoritative one and overriding it would generate the wrong kind of attribute.
//
// overlaySiblings never mutates either argument. libopenapi caches one *base.Schema per $ref target and hands the
// same pointer to every reference site, so writing through resolved would leak these keywords into unrelated
// attributes elsewhere in the document.
func overlaySiblings(composer *base.Schema, resolved *base.Schema, allowTypeInfo bool) *base.Schema {
	if composer == nil || resolved == nil || composer == resolved {
		return resolved
	}

	merged := *resolved

	// Annotations.
	if composer.Description != "" {
		merged.Description = composer.Description
	}
	if composer.Deprecated != nil {
		merged.Deprecated = composer.Deprecated
	}
	if composer.Default != nil {
		merged.Default = composer.Default
	}

	// Type information.
	if allowTypeInfo {
		if len(merged.Type) == 0 && len(composer.Type) > 0 {
			merged.Type = composer.Type
		}
		if composer.Format != "" {
			merged.Format = composer.Format
		}
	}

	// Structure.
	if orderedmap.Len(composer.Properties) > 0 {
		merged.Properties = mergeProperties(resolved.Properties, composer.Properties)
	}
	if len(composer.Required) > 0 {
		merged.Required = unionRequired(resolved.Required, composer.Required)
	}
	if composer.AdditionalProperties != nil {
		merged.AdditionalProperties = composer.AdditionalProperties
	}
	if composer.Items != nil {
		merged.Items = composer.Items
	}

	// Validation.
	if len(composer.Enum) > 0 {
		merged.Enum = composer.Enum
	}
	if composer.Pattern != "" {
		merged.Pattern = composer.Pattern
	}
	if composer.Minimum != nil {
		merged.Minimum = composer.Minimum
	}
	if composer.Maximum != nil {
		merged.Maximum = composer.Maximum
	}
	if composer.MinLength != nil {
		merged.MinLength = composer.MinLength
	}
	if composer.MaxLength != nil {
		merged.MaxLength = composer.MaxLength
	}
	if composer.MinItems != nil {
		merged.MinItems = composer.MinItems
	}
	if composer.MaxItems != nil {
		merged.MaxItems = composer.MaxItems
	}
	if composer.UniqueItems != nil {
		merged.UniqueItems = composer.UniqueItems
	}
	if composer.MinProperties != nil {
		merged.MinProperties = composer.MinProperties
	}
	if composer.MaxProperties != nil {
		merged.MaxProperties = composer.MaxProperties
	}

	return &merged
}

// mergeProperties returns a new ordered map holding the resolved subschema's properties followed by the composing
// schema's own properties. Keys defined on both sides resolve to the composing schema's proxy. Neither input is
// mutated.
func mergeProperties(resolved *orderedmap.Map[string, *base.SchemaProxy], composer *orderedmap.Map[string, *base.SchemaProxy]) *orderedmap.Map[string, *base.SchemaProxy] {
	merged := orderedmap.New[string, *base.SchemaProxy]()

	for name, proxy := range resolved.FromOldest() {
		merged.Set(name, proxy)
	}

	for name, proxy := range composer.FromOldest() {
		merged.Set(name, proxy)
	}

	return merged
}

// unionRequired returns a new slice containing every name in either list, in resolved-then-composer order. A new
// slice is always allocated: appending onto resolved could write into the shared *base.Schema's backing array.
func unionRequired(resolved []string, composer []string) []string {
	merged := make([]string, 0, len(resolved)+len(composer))
	merged = append(merged, resolved...)

	for _, name := range composer {
		if !slices.Contains(merged, name) {
			merged = append(merged, name)
		}
	}

	return merged
}

// retrieveType will return the JSON schema type. Support for multi-types is restricted to combinations of "null" and another type, i.e. ["null", "string"]
func retrieveType(schema *base.Schema) (string, *SchemaError) {
	switch len(schema.Type) {
	case 0:
		// Properties are only valid applying to objects, it's possible tools might omit the type
		// https://github.com/doitintl/terraform-plugin-codegen-openapi/issues/79
		if schema.Properties != nil && schema.Properties.Len() > 0 {
			return util.OAS_type_object, nil
		}

		return "", SchemaErrorFromProxy(errors.New("no 'type' array or supported allOf, oneOf, anyOf constraint - attribute cannot be created"), schema.ParentProxy)
	case 1:
		return schema.Type[0], nil
	case 2:
		// Check for null type, if found, return the other type
		if schema.Type[0] == util.OAS_type_null {
			return schema.Type[1], nil
		} else if schema.Type[1] == util.OAS_type_null {
			return schema.Type[0], nil
		}

		// Check for string type, if the other type can be represented as a string, return the string type
		if schema.Type[0] == util.OAS_type_string && isStringableType(schema.Type[1]) {
			return schema.Type[0], nil
		} else if schema.Type[1] == util.OAS_type_string && isStringableType(schema.Type[0]) {
			return schema.Type[1], nil
		}
	}

	return "", SchemaErrorFromNode(fmt.Errorf("%v - %w", schema.Type, ErrMultiTypeSchema), schema, Type)
}

func isStringableType(t string) bool {
	switch t {
	case util.OAS_type_integer:
		return true
	case util.OAS_type_number:
		return true
	case util.OAS_type_boolean:
		return true
	default:
		return false
	}
}
