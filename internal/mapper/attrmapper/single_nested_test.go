// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmapper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-codegen-spec/datasource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
)

func TestResourceSingleNestedAttribute_Merge(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		targetAttribute   attrmapper.ResourceSingleNestedAttribute
		mergeAttribute    attrmapper.ResourceAttribute
		expectedAttribute attrmapper.ResourceAttribute
	}{
		"mismatch type - no merge": {
			targetAttribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.ResourceStringAttribute{
				Name: "string_attribute",
				StringAttribute: resource.StringAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("nested string description"),
				},
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
		"populated description - no merge": {
			targetAttribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old single nested description"),
				},
			},
			mergeAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old single nested description"),
				},
			},
		},
		"nil description - merge": {
			targetAttribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new single nested description"),
				},
			},
		},
		"empty description - merge": {
			targetAttribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new(""),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new(""),
				},
			},
			mergeAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new single nested description"),
				},
			},
		},
		"nested object - merge": {
			targetAttribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("nested string description"),
						},
					},
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
								},
							},
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceBoolAttribute{
						Name: "nested_bool",
						BoolAttribute: resource.BoolAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("nested bool description"),
						},
					},
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested string description"),
								},
							},
							&attrmapper.ResourceBoolAttribute{
								Name: "double_nested_bool",
								BoolAttribute: resource.BoolAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested bool description"),
								},
							},
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("single nested description"),
				},
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("nested string description"),
						},
					},
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("double nested string description"),
								},
							},
							&attrmapper.ResourceBoolAttribute{
								Name: "double_nested_bool",
								BoolAttribute: resource.BoolAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested bool description"),
								},
							},
						},
					},
					&attrmapper.ResourceBoolAttribute{
						Name: "nested_bool",
						BoolAttribute: resource.BoolAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("nested bool description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("single nested description"),
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.targetAttribute.Merge(testCase.mergeAttribute)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestResourceSingleNestedAttribute_ApplyOverride(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute         attrmapper.ResourceSingleNestedAttribute
		override          explorer.Override
		expectedAttribute attrmapper.ResourceAttribute
	}{
		"override description": {
			attribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "test_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old description"),
				},
			},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "test_attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceStringAttribute{
						Name: "nested_string",
						StringAttribute: resource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new description"),
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.attribute.ApplyOverride(testCase.override)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestResourceSingleNestedAttribute_ApplyNestedOverride(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute         attrmapper.ResourceSingleNestedAttribute
		overridePath      []string
		override          explorer.Override
		expectedAttribute attrmapper.ResourceAttribute
	}{
		"override nested attribute": {
			attribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			overridePath: []string{"nested_attribute"},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
		"override double nested attribute": {
			attribute: attrmapper.ResourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			overridePath: []string{"nested_attribute", "double_nested_attribute"},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.ResourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.ResourceAttributes{
					&attrmapper.ResourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.ResourceAttributes{
							&attrmapper.ResourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: resource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("new description"),
								},
							},
						},
						SingleNestedAttribute: resource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: resource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.attribute.ApplyNestedOverride(testCase.overridePath, testCase.override)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestDataSourceSingleNestedAttribute_Merge(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		targetAttribute   attrmapper.DataSourceSingleNestedAttribute
		mergeAttribute    attrmapper.DataSourceAttribute
		expectedAttribute attrmapper.DataSourceAttribute
	}{
		"mismatch type - no merge": {
			targetAttribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.DataSourceStringAttribute{
				Name: "string_attribute",
				StringAttribute: datasource.StringAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("nested string description"),
				},
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
		"populated description - no merge": {
			targetAttribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old single nested description"),
				},
			},
			mergeAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old single nested description"),
				},
			},
		},
		"nil description - merge": {
			targetAttribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new single nested description"),
				},
			},
		},
		"empty description - merge": {
			targetAttribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new(""),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new(""),
				},
			},
			mergeAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("new single nested description"),
				},
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new nested string description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new single nested description"),
				},
			},
		},
		"nested object - merge": {
			targetAttribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("nested string description"),
						},
					},
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
								},
							},
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			mergeAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceBoolAttribute{
						Name: "nested_bool",
						BoolAttribute: datasource.BoolAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("nested bool description"),
						},
					},
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested string description"),
								},
							},
							&attrmapper.DataSourceBoolAttribute{
								Name: "double_nested_bool",
								BoolAttribute: datasource.BoolAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested bool description"),
								},
							},
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.ComputedOptional,
					Description:              new("single nested description"),
				},
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "single_nested_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("nested string description"),
						},
					},
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_object",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_string",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("double nested string description"),
								},
							},
							&attrmapper.DataSourceBoolAttribute{
								Name: "double_nested_bool",
								BoolAttribute: datasource.BoolAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
									Description:              new("double nested bool description"),
								},
							},
						},
					},
					&attrmapper.DataSourceBoolAttribute{
						Name: "nested_bool",
						BoolAttribute: datasource.BoolAttribute{
							ComputedOptionalRequired: schema.ComputedOptional,
							Description:              new("nested bool description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("single nested description"),
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.targetAttribute.Merge(testCase.mergeAttribute)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestDataSourceSingleNestedAttribute_ApplyOverride(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute         attrmapper.DataSourceSingleNestedAttribute
		override          explorer.Override
		expectedAttribute attrmapper.DataSourceAttribute
	}{
		"override description": {
			attribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "test_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("old description"),
				},
			},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "test_attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceStringAttribute{
						Name: "nested_string",
						StringAttribute: datasource.StringAttribute{
							ComputedOptionalRequired: schema.Required,
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
					Description:              new("new description"),
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.attribute.ApplyOverride(testCase.override)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestDataSourceSingleNestedAttribute_ApplyNestedOverride(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute         attrmapper.DataSourceSingleNestedAttribute
		overridePath      []string
		override          explorer.Override
		expectedAttribute attrmapper.DataSourceAttribute
	}{
		"override nested attribute": {
			attribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			overridePath: []string{"nested_attribute"},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("new description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
		"override double nested attribute": {
			attribute: attrmapper.DataSourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("old description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
			overridePath: []string{"nested_attribute", "double_nested_attribute"},
			override: explorer.Override{
				Description: "new description",
			},
			expectedAttribute: &attrmapper.DataSourceSingleNestedAttribute{
				Name: "attribute",
				Attributes: attrmapper.DataSourceAttributes{
					&attrmapper.DataSourceSingleNestedAttribute{
						Name: "nested_attribute",
						Attributes: attrmapper.DataSourceAttributes{
							&attrmapper.DataSourceStringAttribute{
								Name: "double_nested_attribute",
								StringAttribute: datasource.StringAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("new description"),
								},
							},
						},
						SingleNestedAttribute: datasource.SingleNestedAttribute{
							ComputedOptionalRequired: schema.Required,
							Description:              new("old description"),
						},
					},
				},
				SingleNestedAttribute: datasource.SingleNestedAttribute{
					ComputedOptionalRequired: schema.Required,
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := testCase.attribute.ApplyNestedOverride(testCase.overridePath, testCase.override)

			if diff := cmp.Diff(got, testCase.expectedAttribute); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
