// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oas_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/oas"
)

func TestOASSchemaGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   oas.OASSchema
		expected *string
	}{
		"deprecated-nil": {
			schema: oas.OASSchema{
				Schema: &base.Schema{
					Deprecated: nil,
				},
			},
			expected: nil,
		},
		"deprecated-false": {
			schema: oas.OASSchema{
				Schema: &base.Schema{
					Deprecated: new(false),
				},
			},
			expected: nil,
		},
		"deprecated-true": {
			schema: oas.OASSchema{
				Schema: &base.Schema{
					Deprecated: new(true),
				},
			},
			expected: new("This attribute is deprecated."),
		},
		"deprecated-true-override-empty": {
			schema: oas.OASSchema{
				Schema: &base.Schema{
					Deprecated: new(true),
				},
				SchemaOpts: oas.SchemaOpts{
					OverrideDeprecationMessage: "",
				},
			},
			expected: new("This attribute is deprecated."),
		},
		"deprecated-true-override-non-empty": {
			schema: oas.OASSchema{
				Schema: &base.Schema{
					Deprecated: new(true),
				},
				SchemaOpts: oas.SchemaOpts{
					OverrideDeprecationMessage: "Use test attribute instead.",
				},
			},
			expected: new("Use test attribute instead."),
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetDeprecationMessage()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestGetDescription_Override(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema              oas.OASSchema
		expectedDescription string
	}{
		"override description": {
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					OverrideDescription: "this is the correct description!",
				},
				Schema: &base.Schema{
					Description: "this shouldn't show up!",
				},
			},
			expectedDescription: "this is the correct description!",
		},
		"no override of description": {
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{},
				Schema: &base.Schema{
					Description: "this is the correct description!",
				},
			},
			expectedDescription: "this is the correct description!",
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetDescription()
			if *got != testCase.expectedDescription {
				t.Fatalf("unexpected difference, got: %s, wanted: %s", *got, testCase.expectedDescription)
			}
		})
	}
}

func TestIsPropertyIgnored(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema       oas.OASSchema
		propertyName string
		want         bool
	}{
		"propery is ignored": {
			propertyName: "ignored_prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: []string{
						"ignored_prop",
					},
				},
			},
			want: true,
		},
		"propery not ignored": {
			propertyName: "prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: []string{
						"ignored_prop",
					},
				},
			},
			want: false,
		},
		"nested propery is not ignored": {
			propertyName: "prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: []string{
						"prop.ignored_prop",
					},
				},
			},
			want: false,
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.IsPropertyIgnored(testCase.propertyName)
			if got != testCase.want {
				t.Fatalf("unexpected difference, got: %t, wanted: %t", got, testCase.want)
			}
		})
	}
}

func TestGetIgnoresForNested(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema       oas.OASSchema
		propertyName string
		want         []string
	}{
		"ignores are empty": {
			propertyName: "prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: make([]string, 0),
				},
			},
			want: make([]string, 0),
		},
		"ignores are invalid": {
			propertyName: "prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: []string{
						".prop",
						"prop.",
						".",
						"",
					},
				},
			},
			want: make([]string, 0),
		},
		"nested ignores exist": {
			propertyName: "prop",
			schema: oas.OASSchema{
				SchemaOpts: oas.SchemaOpts{
					Ignores: []string{
						"prop.ignore_me_1",
						"not_me.prop",
						"prop.nested.ignore_me_2",
						"prop.ignore_me_3",
					},
				},
			},
			want: []string{
				"ignore_me_1",
				"nested.ignore_me_2",
				"ignore_me_3",
			},
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetIgnoresForNested(testCase.propertyName)
			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestIsFreeformObject(t *testing.T) {
	t.Parallel()

	// Helper to create a DynamicValue with boolean (N=1 means IsB())
	boolAdditionalProps := func(val bool) *base.DynamicValue[*base.SchemaProxy, bool] {
		return &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: val}
	}

	// Helper to create a DynamicValue with a schema ref (N=0 means IsA())
	schemaAdditionalProps := func() *base.DynamicValue[*base.SchemaProxy, bool] {
		return &base.DynamicValue[*base.SchemaProxy, bool]{N: 0, A: &base.SchemaProxy{}}
	}

	// Helper to create a Properties map with one entry
	withProperties := func() *orderedmap.Map[string, *base.SchemaProxy] {
		m := orderedmap.New[string, *base.SchemaProxy]()
		m.Set("some_property", &base.SchemaProxy{})
		return m
	}

	testCases := map[string]struct {
		schema oas.OASSchema
		want   bool
	}{
		// --- additionalProperties: true ---
		"additionalProperties-true-no-properties": {
			schema: oas.OASSchema{
				Type:   "object",
				Schema: &base.Schema{AdditionalProperties: boolAdditionalProps(true)},
			},
			want: true,
		},
		"additionalProperties-true-with-properties": {
			schema: oas.OASSchema{
				Type: "object",
				Schema: &base.Schema{
					AdditionalProperties: boolAdditionalProps(true),
					Properties:           withProperties(),
				},
			},
			want: false, // has properties, should NOT be freeform
		},

		// --- additionalProperties: false ---
		"additionalProperties-false-no-properties": {
			schema: oas.OASSchema{
				Type:   "object",
				Schema: &base.Schema{AdditionalProperties: boolAdditionalProps(false)},
			},
			want: false,
		},
		"additionalProperties-false-with-properties": {
			schema: oas.OASSchema{
				Type: "object",
				Schema: &base.Schema{
					AdditionalProperties: boolAdditionalProps(false),
					Properties:           withProperties(),
				},
			},
			want: false,
		},

		// --- additionalProperties: <schema> (map) ---
		"additionalProperties-schema-no-properties": {
			schema: oas.OASSchema{
				Type:   "object",
				Schema: &base.Schema{AdditionalProperties: schemaAdditionalProps()},
			},
			want: false, // this is a map, not freeform
		},

		// --- additionalProperties not set ---
		"no-additionalProperties-no-properties-object": {
			schema: oas.OASSchema{
				Type:   "object",
				Schema: &base.Schema{},
			},
			want: true, // type: object with nothing defined
		},
		"no-additionalProperties-with-properties": {
			schema: oas.OASSchema{
				Type: "object",
				Schema: &base.Schema{
					Properties: withProperties(),
				},
			},
			want: false, // has properties
		},
		"no-additionalProperties-no-properties-string": {
			schema: oas.OASSchema{
				Type:   "string",
				Schema: &base.Schema{},
			},
			want: false, // not type: object
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.IsFreeformObject()
			if got != testCase.want {
				t.Errorf("IsFreeformObject() = %t, want %t", got, testCase.want)
			}
		})
	}
}
