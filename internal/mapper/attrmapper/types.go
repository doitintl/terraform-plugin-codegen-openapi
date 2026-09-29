// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmapper

import (
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
)

type ResourceNestedAttributeObject struct {
	Attributes ResourceAttributes
}

type DataSourceNestedAttributeObject struct {
	Attributes DataSourceAttributes
}

type ProviderNestedAttributeObject struct {
	Attributes ProviderAttributes
}

// mergeComputability lets a Required classification from the merge attribute override the target's.
//
// Merging is otherwise target-wins, which is correct for everything derived from a schema -- a
// resource's create request body outranks its response bodies. Required is different: it can only
// originate from a parameter the practitioner has to supply, because every response body is built
// with GlobalSchemaOpts.OverrideComputability set to schema.Computed. Discarding it would silently
// demote an attribute the API needs as an input, which is what happened to a read path parameter
// that shared its name with a response property.
func mergeComputability(target schema.ComputedOptionalRequired, merge schema.ComputedOptionalRequired) schema.ComputedOptionalRequired {
	if merge == schema.Required {
		return schema.Required
	}

	return target
}

func mergeElementType(target schema.ElementType, merge schema.ElementType) schema.ElementType {
	// Handle collection type
	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types#collection-types
	if target.List != nil && merge.List != nil {
		target.List.ElementType = mergeElementType(target.List.ElementType, merge.List.ElementType)
	} else if target.Map != nil && merge.Map != nil {
		target.Map.ElementType = mergeElementType(target.Map.ElementType, merge.Map.ElementType)
	} else if target.Set != nil && merge.Set != nil {
		target.Set.ElementType = mergeElementType(target.Set.ElementType, merge.Set.ElementType)
	}

	// Handle object type
	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types#object-type
	if target.Object != nil && merge.Object != nil {
		target.Object.AttributeTypes = mergeObjectAttributeTypes(target.Object.AttributeTypes, merge.Object.AttributeTypes)
	}

	return target
}

func mergeObjectAttributeTypes(targetAttrTypes []schema.ObjectAttributeType, mergeAttrTypes []schema.ObjectAttributeType) []schema.ObjectAttributeType {
	for _, mergeAttrType := range mergeAttrTypes {
		// As we compare attribute types, if we don't find a match, we should add this attribute type to the slice after
		isNewAttrType := true

		for i, targetAttrType := range targetAttrTypes {
			if targetAttrType.Name == mergeAttrType.Name {
				mergedElementType := mergeElementType(util.CreateElementType(targetAttrType), util.CreateElementType(mergeAttrType))
				targetAttrTypes[i] = util.CreateObjectAttributeType(targetAttrType.Name, mergedElementType)

				isNewAttrType = false
				break
			}
		}

		if isNewAttrType {
			// Add this back to the original slice to avoid adding duplicate attributes from different mergeSlices
			targetAttrTypes = append(targetAttrTypes, mergeAttrType)
		}
	}

	return targetAttrTypes
}
