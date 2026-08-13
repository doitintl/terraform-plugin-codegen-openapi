// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmapper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-codegen-spec/datasource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
)

func TestDataSourceAttributes_Merge(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		targetAttributes     attrmapper.DataSourceAttributes
		mergeAttributeSlices []attrmapper.DataSourceAttributes
		expectedAttributes   attrmapper.DataSourceAttributes
	}{
		"matches and appends": {
			targetAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("string description"),
						Sensitive:                new(true),
					},
				},
			},
			mergeAttributeSlices: []attrmapper.DataSourceAttributes{
				{
					&attrmapper.DataSourceStringAttribute{
						Name: "string_attribute",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Computed,
							Description:              new("this will be ignored"),
							Sensitive:                new(false),
						},
					},
					&attrmapper.DataSourceBoolAttribute{
						Name: "bool_attribute",
						BoolAttribute: datasource.BoolAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("bool description"),
						},
					},
				},
				{
					&attrmapper.DataSourceFloat64Attribute{
						Name: "float64_attribute",
						Float64Attribute: datasource.Float64Attribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("float64 description"),
						},
					},
				},
			},
			expectedAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("string description"),
						Sensitive:                new(true),
					},
				},
				&attrmapper.DataSourceBoolAttribute{
					Name: "bool_attribute",
					BoolAttribute: datasource.BoolAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("bool description"),
					},
				},
				&attrmapper.DataSourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: datasource.Float64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("float64 description"),
					},
				},
			},
		},
		"recursive - matches and appends": {
			targetAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceSingleNestedAttribute{
					Name: "single_nested_attribute",
					Attributes: attrmapper.DataSourceAttributes{
						&attrmapper.DataSourceStringAttribute{
							Name: "string_attribute",
							StringAttribute: datasource.StringAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("string description"),
								Sensitive:                new(true),
							},
						},
					},
					SingleNestedAttribute: datasource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("single nested description"),
					},
				},
			},
			mergeAttributeSlices: []attrmapper.DataSourceAttributes{
				{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "single_nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "string_attribute",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Computed,
									Description:              new("this will be ignored"),
									Sensitive:                new(false),
								},
							},
							&attrmapper.DataSourceBoolAttribute{
								Name: "bool_attribute",
								BoolAttribute: datasource.BoolAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("bool description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("single nested description"),
						},
					},
				},
				{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "single_nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceFloat64Attribute{
								Name: "float64_attribute",
								Float64Attribute: datasource.Float64Attribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("float64 description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("single nested description"),
						},
					},
				},
			},
			expectedAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceSingleNestedAttribute{
					Name: "single_nested_attribute",
					Attributes: attrmapper.DataSourceAttributes{
						&attrmapper.DataSourceStringAttribute{
							Name: "string_attribute",
							StringAttribute: datasource.StringAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("string description"),
								Sensitive:                new(true),
							},
						},
						&attrmapper.DataSourceBoolAttribute{
							Name: "bool_attribute",
							BoolAttribute: datasource.BoolAttribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("bool description"),
							},
						},
						&attrmapper.DataSourceFloat64Attribute{
							Name: "float64_attribute",
							Float64Attribute: datasource.Float64Attribute{
								ComputedOptionalRequired: schema.Required,
								Description:              new("float64 description"),
							},
						},
					},
					SingleNestedAttribute: datasource.SingleNestedAttribute{
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

func TestDataSourceAttributes_ApplyOverrides(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		overrides          map[string]explorer.Override
		attributes         attrmapper.DataSourceAttributes
		expectedAttributes attrmapper.DataSourceAttributes
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
			attributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
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
			attributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
				&attrmapper.DataSourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: datasource.Float64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceStringAttribute{
					Name: "string_attribute",
					StringAttribute: datasource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("new string description"),
					},
				},
				&attrmapper.DataSourceFloat64Attribute{
					Name: "float64_attribute",
					Float64Attribute: datasource.Float64Attribute{
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
			attributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceSingleNestedAttribute{
					Name: "single_nested",
					Attributes: attrmapper.DataSourceAttributes{
						&attrmapper.DataSourceListNestedAttribute{
							Name: "list_nested",
							NestedObject: attrmapper.DataSourceNestedAttributeObject{
								attrmapper.DataSourceAttributes{
									&attrmapper.DataSourceStringAttribute{
										Name: "string_attribute",
										StringAttribute: datasource.StringAttribute{
											ComputedOptionalRequired: schema.Required,
											Description:              new("old description"),
										},
									},
								},
							},
							ListNestedAttribute: datasource.ListNestedAttribute{
								ComputedOptionalRequired: schema.Optional,
								Description:              new("old description"),
							},
						},
					},
					SingleNestedAttribute: datasource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Optional,
						Description:              new("old description"),
					},
				},
			},
			expectedAttributes: attrmapper.DataSourceAttributes{
				&attrmapper.DataSourceSingleNestedAttribute{
					Name: "single_nested",
					Attributes: attrmapper.DataSourceAttributes{
						&attrmapper.DataSourceListNestedAttribute{
							Name: "list_nested",
							NestedObject: attrmapper.DataSourceNestedAttributeObject{
								attrmapper.DataSourceAttributes{
									&attrmapper.DataSourceStringAttribute{
										Name: "string_attribute",
										StringAttribute: datasource.StringAttribute{
											ComputedOptionalRequired: schema.Required,
											Description:              new("new description"),
										},
									},
								},
							},
							ListNestedAttribute: datasource.ListNestedAttribute{
								ComputedOptionalRequired: schema.Optional,
								Description:              new("new description"),
							},
						},
					},
					SingleNestedAttribute: datasource.SingleNestedAttribute{
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
