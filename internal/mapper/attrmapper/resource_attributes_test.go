// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmapper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
)

func TestResourceAttributes_Merge(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		targetAttributes     attrmapper.ResourceAttributes
		mergeAttributeSlices []attrmapper.ResourceAttributes
		expectedAttributes   attrmapper.ResourceAttributes
	}{
		"matches and appends": {
			targetAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("string description"),
						Sensitive:                new(true),
					},
				},
			},
			mergeAttributeSlices: []attrmapper.ResourceAttributes{
				{
					&attrmapper.ResourceStringAttribute{
						Name: "string_attribute",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Computed,
							Description:              new("this will be ignored"),
							Sensitive:                new(false),
						},
					},
					&attrmapper.ResourceBoolAttribute{
						Name: "bool_attribute",
						BoolAttribute: resource.BoolAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("bool description"),
						},
					},
				},
				{
					&attrmapper.ResourceFloat64Attribute{
						Name: "float64_attribute",
						Float64Attribute: resource.Float64Attribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("float64 description"),
						},
					},
				},
			},
			expectedAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("string description"),
						Sensitive:                new(true),
					},
				},
				&attrmapper.ResourceBoolAttribute{
					Name: "bool_attribute",
					BoolAttribute: resource.BoolAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("bool description"),
					},
				},
				&attrmapper.ResourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: resource.Float64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("float64 description"),
					},
				},
			},
		},
		"recursive - matches and appends": {
			targetAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceSingleNestedAttribute{
					Name: "single_nested_attribute",
					Attributes: attrmapper.ResourceAttributes{
						&attrmapper.ResourceStringAttribute{
							Name: "string_attribute",
							StringAttribute: resource.StringAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("string description"),
								Sensitive:                new(true),
							},
						},
					},
					SingleNestedAttribute: resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("single nested description"),
					},
				},
			},
			mergeAttributeSlices: []attrmapper.ResourceAttributes{
				{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "single_nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "string_attribute",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Computed,
									Description:              new("this will be ignored"),
									Sensitive:                new(false),
								},
							},
							&attrmapper.ResourceBoolAttribute{
								Name: "bool_attribute",
								BoolAttribute: resource.BoolAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("bool description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("single nested description"),
						},
					},
				},
				{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "single_nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceFloat64Attribute{
								Name: "float64_attribute",
								Float64Attribute: resource.Float64Attribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("float64 description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("single nested description"),
						},
					},
				},
			},
			expectedAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceSingleNestedAttribute{
					Name: "single_nested_attribute",
					Attributes: attrmapper.ResourceAttributes{
						&attrmapper.ResourceStringAttribute{
							Name: "string_attribute",
							StringAttribute: resource.StringAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("string description"),
								Sensitive:                new(true),
							},
						},
						&attrmapper.ResourceBoolAttribute{
							Name: "bool_attribute",
							BoolAttribute: resource.BoolAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("bool description"),
							},
						},
						&attrmapper.ResourceFloat64Attribute{
							Name: "float64_attribute",
							Float64Attribute: resource.Float64Attribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("float64 description"),
							},
						},
					},
					SingleNestedAttribute: resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("single nested description"),
					},
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.targetAttributes.Merge(testCase.mergeAttributeSlices...)

			if diff := cmp.Diff(got, testCase.expectedAttributes); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestResourceAttributes_ApplyOverrides(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		overrides          map[string]explorer.Override
		attributes         attrmapper.ResourceAttributes
		expectedAttributes attrmapper.ResourceAttributes
	}{
		// TODO: this may eventually return an error, but for now just returns without modification
		"no matching overrides": {
			overrides: map[string]explorer.Override{
				"": {
					Description: "new description",
				},
				"attribute_that_doesnt_exist": {
					Description: "new description",
				},
				"string_attribute.attribute_that_doesnt_exist": {
					Description: "new description",
				},
			},
			attributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
			},
		},
		"matching overrides": {
			overrides: map[string]explorer.Override{
				"string_attribute": {
					Description: "new string description",
				},
				"float64_attribute": {
					Description: "new float64 description",
				},
			},
			attributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
				&attrmapper.ResourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: resource.Float64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("new string description"),
					},
				},
				&attrmapper.ResourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: resource.Float64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("new float64 description"),
					},
				},
			},
		},
		"matching nested overrides": {
			overrides: map[string]explorer.Override{
				"single_nested": {
					Description: "new description",
				},
				"single_nested.list_nested": {
					Description: "new description",
				},
				"single_nested.list_nested.string_attribute": {
					Description: "new description",
				},
			},
			attributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceSingleNestedAttribute{
					Name: "single_nested",
					Attributes: attrmapper.ResourceAttributes{
						&attrmapper.ResourceListNestedAttribute{
							Name: "list_nested",
							NestedObject: attrmapper.ResourceNestedAttributeObject{
								attrmapper.ResourceAttributes{
									&attrmapper.ResourceStringAttribute{
										Name: "string_attribute",
										StringAttribute: resource.StringAttribute{
											ComputedOptionalRequired: schema.Required,
											Description:              new("old description"),
										},
									},
								},
							},
							ListNestedAttribute: resource.ListNestedAttribute{
								ComputedOptionalRequired: schema.Optional,
								Description:              new("old description"),
							},
						},
					},
					SingleNestedAttribute: resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Optional,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.ResourceAttributes{
				&attrmapper.ResourceSingleNestedAttribute{
					Name: "single_nested",
					Attributes: attrmapper.ResourceAttributes{
						&attrmapper.ResourceListNestedAttribute{
							Name: "list_nested",
							NestedObject: attrmapper.ResourceNestedAttributeObject{
								attrmapper.ResourceAttributes{
									&attrmapper.ResourceStringAttribute{
										Name: "string_attribute",
										StringAttribute: resource.StringAttribute{
											ComputedOptionalRequired: schema.Required,
											Description:              new("new description"),
										},
									},
								},
							},
							ListNestedAttribute: resource.ListNestedAttribute{
								ComputedOptionalRequired: schema.Optional,
								Description:              new("new description"),
							},
						},
					},
					SingleNestedAttribute: resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Optional,
						Description:              new("new description"),
					},
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.attributes.ApplyOverrides(testCase.overrides)

			if diff := cmp.Diff(got, testCase.expectedAttributes); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
