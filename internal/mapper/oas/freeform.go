// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oas

import (
	"github.com/hashicorp/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
	"github.com/hashicorp/terraform-plugin-codegen-spec/code"
	"github.com/hashicorp/terraform-plugin-codegen-spec/datasource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/provider"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
)

// jsonNormalizedCustomType returns the CustomType for jsontypes.Normalized,
// which provides plan-time JSON validation and semantic equality.
func jsonNormalizedCustomType() *schema.CustomType {
	return &schema.CustomType{
		Import: &code.Import{
			Path: "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
		},
		Type:      "jsontypes.NormalizedType{}",
		ValueType: "jsontypes.Normalized",
	}
}

// freeformDescription appends a JSON encoding hint to the original description.
func freeformDescription(original *string) *string {
	suffix := "Value is JSON-encoded."
	if original == nil || *original == "" {
		return &suffix
	}
	sep := " "
	last := (*original)[len(*original)-1]
	if last != '.' && last != '!' && last != '?' {
		sep = ". "
	}
	desc := *original + sep + suffix
	return &desc
}

func (s *OASSchema) BuildFreeformResource(name string, computability schema.ComputedOptionalRequired) (attrmapper.ResourceAttribute, *SchemaError) {
	desc := freeformDescription(s.GetDescription())

	result := &attrmapper.ResourceStringAttribute{
		Name: name,
		StringAttribute: resource.StringAttribute{
			ComputedOptionalRequired: computability,
			CustomType:               jsonNormalizedCustomType(),
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              desc,
		},
	}

	return result, nil
}

func (s *OASSchema) BuildFreeformDataSource(name string, computability schema.ComputedOptionalRequired) (attrmapper.DataSourceAttribute, *SchemaError) {
	desc := freeformDescription(s.GetDescription())

	result := &attrmapper.DataSourceStringAttribute{
		Name: name,
		StringAttribute: datasource.StringAttribute{
			ComputedOptionalRequired: computability,
			CustomType:               jsonNormalizedCustomType(),
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              desc,
		},
	}

	return result, nil
}

func (s *OASSchema) BuildFreeformProvider(name string, optionalOrRequired schema.OptionalRequired) (attrmapper.ProviderAttribute, *SchemaError) {
	desc := freeformDescription(s.GetDescription())

	result := &attrmapper.ProviderStringAttribute{
		Name: name,
		StringAttribute: provider.StringAttribute{
			OptionalRequired:   optionalOrRequired,
			CustomType:         jsonNormalizedCustomType(),
			DeprecationMessage: s.GetDeprecationMessage(),
			Description:        desc,
		},
	}

	return result, nil
}

// BuildFreeformElementType returns a string element type for free-form objects
// used as element types (e.g., inside an array). The custom type cannot be
// expressed at the element type level, so we fall back to a plain string.
func (s *OASSchema) BuildFreeformElementType() (schema.ElementType, *SchemaError) {
	return schema.ElementType{
		String: &schema.StringType{},
	}, nil
}
