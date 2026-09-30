// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapper_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/config"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/frameworkvalidators"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/util"

	"github.com/google/go-cmp/cmp"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	yaml "go.yaml.in/yaml/v4"
)

func TestResourceMapper_basic_merges(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		createRequestSchema  *base.SchemaProxy
		createResponseSchema *base.SchemaProxy
		readResponseSchema   *base.SchemaProxy
		readParams           []*high.Parameter
		createParams         []*high.Parameter
		createCommonParams   []*high.Parameter
		updateRequestSchema  *base.SchemaProxy
		schemaOptions        explorer.SchemaOptions
		want                 resource.Attributes
	}{
		"merge primitives across all ops": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"bool_prop", "int64_prop"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"bool_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"boolean"},
						Description: "hey this is a bool, required!",
					}),
					"int64_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"integer"},
						Description: "hey this is an int64, required!",
					}),
				}),
			}),
			createResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"int64_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"integer"},
						Description: "this one already exists, so you shouldn't see this description!",
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"number_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"number"},
						Description: "hey this is a number!",
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name:        "string_prop",
					In:          "path",
					Description: "hey this is a string, overridden!",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"string"},
						Format:      util.OAS_format_password,
						Description: "you shouldn't see this because the description is overridden!",
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "bool_prop",
					Bool: &resource.BoolAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("hey this is a bool, required!"),
					},
				},
				{
					Name: "int64_prop",
					Int64: &resource.Int64Attribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("hey this is an int64, required!"),
					},
				},
				{
					Name: "number_prop",
					Number: &resource.NumberAttribute{
						ComputedOptionalRequired: schema.Computed,
						Description:              new("hey this is a number!"),
					},
				},
				{
					Name: "string_prop",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("hey this is a string, overridden!"),
						Sensitive:                new(true),
					},
				},
			},
		},
		"deep merge single nested object": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"nested_object_one"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"nested_object_one": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"object"},
						Required:    []string{"bool_prop"},
						Description: "hey this is an object, required!",
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"bool_prop": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"boolean"},
								Description: "hey this is a bool, required!",
							}),
						}),
					}),
				}),
			}),
			createResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"nested_object_one": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"object"},
						Description: "this one already exists, so you shouldn't see this description!",
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"string_prop": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"string"},
								Format:      util.OAS_format_password,
								Description: "hey this is a string!",
							}),
							"nested_object_two": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"object"},
								Description: "hey this is an object!",
								Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
									"bool_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"boolean"},
										Description: "hey this is a bool!",
									}),
									"number_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"number"},
										Description: "hey this is a number!",
									}),
								}),
							}),
						}),
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name: "nested_object_one",
					In:   "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"object"},
						Description: "this one already exists, so you shouldn't see this description!",
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"nested_object_two": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"object"},
								Required:    []string{"int64_prop"},
								Description: "this one already exists, so you shouldn't see this description!",
								Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
									"bool_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"boolean"},
										Description: "hey this is a bool!",
									}),
									"int64_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"integer"},
										Description: "hey this is a integer, switched to computed optional!",
									}),
								}),
							}),
						}),
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "nested_object_one",
					SingleNested: &resource.SingleNestedAttribute{
						Attributes: []resource.Attribute{
							{
								Name: "bool_prop",
								Bool: &resource.BoolAttribute{
									ComputedOptionalRequired: schema.Required,
									Description:              new("hey this is a bool, required!"),
								},
							},
							{
								Name: "nested_object_two",
								SingleNested: &resource.SingleNestedAttribute{
									Attributes: []resource.Attribute{
										{
											Name: "bool_prop",
											Bool: &resource.BoolAttribute{
												ComputedOptionalRequired: schema.Computed,
												Description:              new("hey this is a bool!"),
											},
										},
										{
											Name: "number_prop",
											Number: &resource.NumberAttribute{
												ComputedOptionalRequired: schema.Computed,
												Description:              new("hey this is a number!"),
											},
										},
										{
											Name: "int64_prop",
											Int64: &resource.Int64Attribute{
												ComputedOptionalRequired: schema.ComputedOptional,
												Description:              new("hey this is a integer, switched to computed optional!"),
											},
										},
									},
									ComputedOptionalRequired: schema.Computed,
									Description:              new("hey this is an object!"),
								},
							},
							{
								Name: "string_prop",
								String: &resource.StringAttribute{
									ComputedOptionalRequired: schema.Computed,
									Description:              new("hey this is a string!"),
									Sensitive:                new(true),
								},
							},
						},
						ComputedOptionalRequired: schema.Required,
						Description:              new("hey this is an object, required!"),
					},
				},
			},
		},
		"deep merge list nested array": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"array_prop"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"array_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array, required!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type:     []string{"object"},
								Required: []string{"nested_array_prop"},
								Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
									"float64_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"number"},
										Format:      "double",
										Description: "hey this is a float64!",
									}),
									"nested_array_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"array"},
										Description: "hey this is a nested array, required!",
										Items: &base.DynamicValue[*base.SchemaProxy, bool]{
											A: base.CreateSchemaProxy(&base.Schema{
												Type: []string{"object"},
												Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
													"super_nested_string": base.CreateSchemaProxy(&base.Schema{
														Type:        []string{"string"},
														Description: "hey this is a string!",
													}),
												}),
											}),
										},
									}),
								}),
							}),
						},
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"array_prop"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"array_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array, required!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type:     []string{"object"},
								Required: []string{"nested_array_prop"},
								Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
									"nested_array_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"array"},
										Description: "hey this is a nested array, required!",
										Items: &base.DynamicValue[*base.SchemaProxy, bool]{
											A: base.CreateSchemaProxy(&base.Schema{
												Type:     []string{"object"},
												Required: []string{"super_nested_bool_two"},
												Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
													"super_nested_bool_one": base.CreateSchemaProxy(&base.Schema{
														Type:        []string{"boolean"},
														Description: "hey this is a boolean!",
													}),
													"super_nested_bool_two": base.CreateSchemaProxy(&base.Schema{
														Type:        []string{"boolean"},
														Description: "hey this is a boolean, switched to computed optional!",
													}),
													"super_nested_int64": base.CreateSchemaProxy(&base.Schema{
														Type:        []string{"integer"},
														Description: "hey this is a integer!",
													}),
												}),
											}),
										},
									}),
									"number_prop": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"number"},
										Description: "hey this is a number!",
									}),
								}),
							}),
						},
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "array_prop",
					ListNested: &resource.ListNestedAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("hey this is an array, required!"),
						NestedObject: resource.NestedAttributeObject{
							Attributes: []resource.Attribute{
								{
									Name: "float64_prop",
									Float64: &resource.Float64Attribute{
										ComputedOptionalRequired: schema.ComputedOptional,
										Description:              new("hey this is a float64!"),
									},
								},
								{
									Name: "nested_array_prop",
									ListNested: &resource.ListNestedAttribute{
										ComputedOptionalRequired: schema.Required,
										Description:              new("hey this is a nested array, required!"),
										NestedObject: resource.NestedAttributeObject{
											Attributes: []resource.Attribute{
												{
													Name: "super_nested_string",
													String: &resource.StringAttribute{
														ComputedOptionalRequired: schema.ComputedOptional,
														Description:              new("hey this is a string!"),
													},
												},
												{
													Name: "super_nested_bool_one",
													Bool: &resource.BoolAttribute{
														ComputedOptionalRequired: schema.Computed,
														Description:              new("hey this is a boolean!"),
													},
												},
												{
													Name: "super_nested_bool_two",
													Bool: &resource.BoolAttribute{
														ComputedOptionalRequired: schema.Computed,
														Description:              new("hey this is a boolean, switched to computed optional!"),
													},
												},
												{
													Name: "super_nested_int64",
													Int64: &resource.Int64Attribute{
														ComputedOptionalRequired: schema.Computed,
														Description:              new("hey this is a integer!"),
													},
												},
											},
										},
									},
								},
								{
									Name: "number_prop",
									Number: &resource.NumberAttribute{
										ComputedOptionalRequired: schema.Computed,
										Description:              new("hey this is a number!"),
									},
								},
							},
						},
					},
				},
			},
		},
		"deep merge list array with object element types": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"array_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type: []string{"array"},
								Items: &base.DynamicValue[*base.SchemaProxy, bool]{
									A: base.CreateSchemaProxy(&base.Schema{
										Type: []string{"object"},
										Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
											"deep_nested_float64": base.CreateSchemaProxy(&base.Schema{
												Type:   []string{"number"},
												Format: "float",
											}),
											"deep_nested_string": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"string"},
											}),
										}),
									}),
								},
							}),
						},
					}),
				}),
			}),
			createResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"array_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type: []string{"array"},
								Items: &base.DynamicValue[*base.SchemaProxy, bool]{
									A: base.CreateSchemaProxy(&base.Schema{
										Type: []string{"object"},
										Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
											"deep_nested_list": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"array"},
												Items: &base.DynamicValue[*base.SchemaProxy, bool]{
													A: base.CreateSchemaProxy(&base.Schema{
														Type: []string{"object"},
														Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
															"deep_deep_nested_object": base.CreateSchemaProxy(&base.Schema{
																Type: []string{"object"},
																Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
																	"deep_deep_nested_string": base.CreateSchemaProxy(&base.Schema{
																		Type: []string{"string"},
																	}),
																}),
															}),
														}),
													}),
												},
											}),
											"deep_nested_bool": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"boolean"},
											}),
											"deep_nested_int64": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"integer"},
											}),
										}),
									}),
								},
							}),
						},
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"array_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type: []string{"array"},
								Items: &base.DynamicValue[*base.SchemaProxy, bool]{
									A: base.CreateSchemaProxy(&base.Schema{
										Type: []string{"object"},
										Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
											"deep_nested_list": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"array"},
												Items: &base.DynamicValue[*base.SchemaProxy, bool]{
													A: base.CreateSchemaProxy(&base.Schema{
														Type: []string{"object"},
														Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
															"deep_deep_nested_object": base.CreateSchemaProxy(&base.Schema{
																Type: []string{"object"},
																Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
																	"deep_deep_nested_bool": base.CreateSchemaProxy(&base.Schema{
																		Type: []string{"boolean"},
																	}),
																}),
															}),
														}),
													}),
												},
											}),
										}),
									}),
								},
							}),
						},
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "array_prop",
					List: &resource.ListAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("hey this is an array!"),
						ElementType: schema.ElementType{
							List: &schema.ListType{
								ElementType: schema.ElementType{
									Object: &schema.ObjectType{
										AttributeTypes: []schema.ObjectAttributeType{
											{
												Name:    "deep_nested_float64",
												Float64: &schema.Float64Type{},
											},
											{
												Name:   "deep_nested_string",
												String: &schema.StringType{},
											},
											{
												Name: "deep_nested_bool",
												Bool: &schema.BoolType{},
											},
											{
												Name:  "deep_nested_int64",
												Int64: &schema.Int64Type{},
											},
											{
												Name: "deep_nested_list",
												List: &schema.ListType{
													ElementType: schema.ElementType{
														Object: &schema.ObjectType{
															AttributeTypes: []schema.ObjectAttributeType{
																{
																	Name: "deep_deep_nested_object",
																	Object: &schema.ObjectType{
																		AttributeTypes: []schema.ObjectAttributeType{
																			{
																				Name:   "deep_deep_nested_string",
																				String: &schema.StringType{},
																			},
																			{
																				Name: "deep_deep_nested_bool",
																				Bool: &schema.BoolType{},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"precedence and configurability": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Required: []string{
					"create_request_required_create_request_only",
					"create_request_required_create_response",
					"create_request_required_read_parameter_optional",
					"create_request_required_read_parameter_required",
					"create_request_required_read_response",
				},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"create_request_optional_create_request_only": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_optional_create_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_optional_read_parameter_optional": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_optional_read_parameter_required": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_optional_read_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_required_create_request_only": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_required_create_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_required_read_parameter_optional": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_required_read_parameter_required": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_request_required_read_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			createResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					// Simulate API returning parameter in response
					"create_request_optional_create_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					// Simulate API returning parameter in response
					"create_request_required_create_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"create_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name:     "create_request_optional_read_parameter_optional",
					Required: new(false),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "create_request_optional_read_parameter_required",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "create_request_required_read_parameter_optional",
					Required: new(false),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "create_request_required_read_parameter_required",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				// Edge case where new read request properties do not align with
				// other request/response properties. Provider developers would
				// want to configure the converter to remap existing properties
				// to align with these or risk the converter returning API-level
				// details to practitioners, which are conventionally hidden.
				{
					Name:     "read_parameter_optional",
					Required: new(false),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "read_parameter_required",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					// Simulate API returning parameter in response
					"create_request_optional_read_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					// Simulate API returning parameter in response
					"create_request_required_read_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
					"read_response": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "create_request_optional_create_request_only",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "create_request_optional_create_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "create_request_optional_read_parameter_optional",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "create_request_optional_read_parameter_required",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "create_request_optional_read_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "create_request_required_create_request_only",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "create_request_required_create_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "create_request_required_read_parameter_optional",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "create_request_required_read_parameter_required",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "create_request_required_read_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "create_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Computed,
					},
				},
				{
					Name: "read_response",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Computed,
					},
				},
				{
					Name: "read_parameter_optional",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "read_parameter_required",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
			},
		},
		"parameter match for path and query params": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"attribute_required"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"attribute_required": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name:     "read_path_parameter",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "read_query_parameter",
					Required: new(false),
					In:       "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"boolean"},
					}),
				},
			},
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"attribute_computed": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"boolean"},
					}),
				}),
			}),
			schemaOptions: explorer.SchemaOptions{
				AttributeOptions: explorer.AttributeOptions{
					Aliases: map[string]string{
						"read_path_parameter":  "attribute_required",
						"read_query_parameter": "attribute_computed",
					},
				},
			},
			want: resource.Attributes{
				{
					Name: "attribute_required",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "attribute_computed",
					Bool: &resource.BoolAttribute{
						ComputedOptionalRequired: schema.Computed,
					},
				},
			},
		},
		"required read path param on the create path - required": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"name"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "child_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "parent_id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					// Required in the read path but absent from the create path, so the API
					// assigned it.
					Name: "child_id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
			},
		},
		"create path param declared on the path item - required": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			createCommonParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "parent_id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
			},
		},
		"query params are never promoted": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "zone",
					Required: new(true),
					In:       "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					Name:     "region",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					// Matches a create query param, not a create path param.
					Name:     "zone",
					Required: new(true),
					In:       "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
				{
					// Matches a create path param, but is itself a query param.
					Name:     "region",
					Required: new(true),
					In:       "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "zone",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "region",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
			},
		},
		"path param not marked required - not promoted": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			createParams: []*high.Parameter{
				{
					Name: "parent_id",
					In:   "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name: "parent_id",
					In:   "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "parent_id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
			},
		},
		"required param overrides a response-derived classification": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					// The response reaches the merge target first and is forced to Computed, so
					// without the promotion the parameter's classification would be discarded.
					"id": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"string"},
						Description: "the response description still wins",
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name:        "id",
					Required:    new(true),
					In:          "path",
					Description: "you shouldn't see this, the response was merged first",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						Description:              new("the response description still wins"),
					},
				},
			},
		},
		"required create path param promotes an optional create body property": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					// Optional in the body, but the value is also addressed by the create URL, so
					// the practitioner has to supply it and the parameter's Required wins.
					"name": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "name",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name:     "name",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
			},
		},
		"object-typed required path param keeps nested props computed_optional": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "selector",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"object"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name:     "selector",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:     []string{"object"},
						Required: []string{"key"},
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"key": base.CreateSchemaProxy(&base.Schema{
								Type: []string{"string"},
							}),
						}),
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "selector",
					SingleNested: &resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Required,
						Attributes: resource.Attributes{
							{
								Name: "key",
								String: &resource.StringAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
								},
							},
						},
					},
				},
			},
		},
		"update-only property in the read response - computed_optional with update validators": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
				Required: []string{"name"},
			}),
			// Required in the update body, but the practitioner is not obliged to manage it, and only
			// the update body's enum describes what may be sent.
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"state"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"state": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"string"},
						Description: "The state to move to.",
						Enum: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "active"},
							{Kind: yaml.ScalarNode, Value: "disabled"},
						},
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"state": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"string"},
						Description: "The current state.",
						Enum: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "active"},
							{Kind: yaml.ScalarNode, Value: "disabled"},
							{Kind: yaml.ScalarNode, Value: "expired"},
						},
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
				{
					Name: "state",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("The state to move to."),
						Validators: []schema.StringValidator{
							{
								Custom: frameworkvalidators.StringValidatorOneOf([]string{"active", "disabled"}),
							},
						},
					},
				},
			},
		},
		"update-only property absent from responses - added as computed_optional": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"language": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "language",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
			},
		},
		"create body property wins over update body": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"name"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type:      []string{"string"},
						MinLength: new(int64(1)),
					}),
				}),
			}),
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"name": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"string"},
						Description: "The name.",
						MaxLength:   new(int64(5)),
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "name",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
						// The create body has no description, so the update body's fills it.
						Description: new("The name."),
						Validators: []schema.StringValidator{
							{
								Custom: frameworkvalidators.StringValidatorLengthAtLeast(1),
							},
						},
					},
				},
			},
		},
		"nested update-only object - nested props computed_optional": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:     []string{"object"},
				Required: []string{"settings"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"settings": base.CreateSchemaProxy(&base.Schema{
						Type:     []string{"object"},
						Required: []string{"mode"},
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"mode": base.CreateSchemaProxy(&base.Schema{
								Type: []string{"string"},
							}),
						}),
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"settings": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"object"},
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"mode": base.CreateSchemaProxy(&base.Schema{
								Type: []string{"string"},
							}),
						}),
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "settings",
					SingleNested: &resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Attributes: resource.Attributes{
							{
								Name: "mode",
								String: &resource.StringAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
								},
							},
						},
					},
				},
			},
		},
		"update-only defaults are not mapped": {
			// A create body default still applies: it is the value the resource is created with.
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"private": base.CreateSchemaProxy(&base.Schema{
						Type:    []string{"boolean"},
						Default: &yaml.Node{Kind: yaml.ScalarNode, Value: "true"},
					}),
				}),
			}),
			// An update body default would make Terraform send it whenever the field is omitted,
			// overwriting the API's value, at the top level and nested alike.
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"archived": base.CreateSchemaProxy(&base.Schema{
						Type:    []string{"boolean"},
						Default: &yaml.Node{Kind: yaml.ScalarNode, Value: "false"},
					}),
					"settings": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"object"},
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"mode": base.CreateSchemaProxy(&base.Schema{
								Type:    []string{"string"},
								Default: &yaml.Node{Kind: yaml.ScalarNode, Value: "fast"},
							}),
						}),
					}),
				}),
			}),
			want: resource.Attributes{
				{
					Name: "private",
					Bool: &resource.BoolAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Default: &schema.BoolDefault{
							Static: new(true),
						},
					},
				},
				{
					Name: "archived",
					Bool: &resource.BoolAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
					},
				},
				{
					Name: "settings",
					SingleNested: &resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Attributes: resource.Attributes{
							{
								Name: "mode",
								String: &resource.StringAttribute{
									ComputedOptionalRequired: schema.ComputedOptional,
								},
							},
						},
					},
				},
			},
		},
		"required create path param still promotes over update body": {
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{}),
			}),
			updateRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"parent_id": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				}),
			}),
			createParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			readParams: []*high.Parameter{
				{
					Name:     "parent_id",
					Required: new(true),
					In:       "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "parent_id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Required,
					},
				},
			},
		},
		"ignore bool prop across all ops": {
			schemaOptions: explorer.SchemaOptions{
				Ignores: []string{
					"bool_prop",
					"nested_obj.bool_prop",
					"nested_array.deep_nested_bool",
					"nested_map.deep_nested_bool",
				},
			},
			createRequestSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"bool_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"boolean"},
						Description: "This boolean is going to be ignored!",
					}),
					"number_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"number"},
						Description: "hey this is a number!",
					}),
					"nested_map": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"object"},
						Description: "hey this is a map!",
						AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type: []string{"object"},
								Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
									"deep_nested_bool": base.CreateSchemaProxy(&base.Schema{
										Type: []string{"boolean"},
									}),
									"deep_nested_int64": base.CreateSchemaProxy(&base.Schema{
										Type:        []string{"integer"},
										Description: "hey this is an int64!",
									}),
								}),
							}),
						},
					}),
				}),
			}),
			createResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"bool_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"boolean"},
						Description: "This boolean is going to be ignored!",
					}),
					"number_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"number"},
						Description: "hey this is a number!",
					}),
					"nested_array": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"array"},
						Description: "hey this is an array!",
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{
							A: base.CreateSchemaProxy(&base.Schema{
								Type: []string{"array"},
								Items: &base.DynamicValue[*base.SchemaProxy, bool]{
									A: base.CreateSchemaProxy(&base.Schema{
										Type: []string{"object"},
										Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
											"deep_nested_bool": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"boolean"},
											}),
											"deep_nested_int64": base.CreateSchemaProxy(&base.Schema{
												Type: []string{"integer"},
											}),
										}),
									}),
								},
							}),
						},
					}),
				}),
			}),
			readResponseSchema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"object"},
				Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
					"bool_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"boolean"},
						Description: "This boolean is going to be ignored!",
					}),
					"number_prop": base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"number"},
						Description: "hey this is a number!",
					}),
					"nested_obj": base.CreateSchemaProxy(&base.Schema{
						Type: []string{"object"},
						Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
							"bool_prop": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"boolean"},
								Description: "This boolean is going to be ignored!",
							}),
							"string_prop": base.CreateSchemaProxy(&base.Schema{
								Type:        []string{"string"},
								Description: "hey this is a string!",
							}),
						}),
					}),
				}),
			}),
			readParams: []*high.Parameter{
				{
					Name:     "bool_prop",
					Required: new(true),
					In:       "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"boolean"},
						Description: "This boolean is going to be ignored!",
					}),
				},
				{
					Name: "float64_prop",
					In:   "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:        []string{"number"},
						Format:      "float",
						Description: "hey this is a float64!",
					}),
				},
			},
			want: resource.Attributes{
				{
					Name: "nested_map",
					MapNested: &resource.MapNestedAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("hey this is a map!"),
						NestedObject: resource.NestedAttributeObject{
							Attributes: []resource.Attribute{
								{
									Name: "deep_nested_int64",
									Int64: &resource.Int64Attribute{
										ComputedOptionalRequired: schema.ComputedOptional,
										Description:              new("hey this is an int64!"),
									},
								},
							},
						},
					},
				},
				{
					Name: "number_prop",
					Number: &resource.NumberAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("hey this is a number!"),
					},
				},
				{
					Name: "nested_array",
					List: &resource.ListAttribute{
						ComputedOptionalRequired: schema.Computed,
						Description:              new("hey this is an array!"),
						ElementType: schema.ElementType{
							List: &schema.ListType{
								ElementType: schema.ElementType{
									Object: &schema.ObjectType{
										AttributeTypes: []schema.ObjectAttributeType{
											{
												Name:  "deep_nested_int64",
												Int64: &schema.Int64Type{},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Name: "nested_obj",
					SingleNested: &resource.SingleNestedAttribute{
						ComputedOptionalRequired: schema.Computed,
						Attributes: []resource.Attribute{
							{
								Name: "string_prop",
								String: &resource.StringAttribute{
									ComputedOptionalRequired: schema.Computed,
									Description:              new("hey this is a string!"),
								},
							},
						},
					},
				},
				{
					Name: "float64_prop",
					Float64: &resource.Float64Attribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						Description:              new("hey this is a float64!"),
					},
				},
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mapper := mapper.NewResourceMapper(map[string]explorer.Resource{
				"test_resource": {
					CreateOp:               createTestCreateOp(testCase.createRequestSchema, testCase.createResponseSchema, testCase.createParams),
					ReadOp:                 createTestReadOp(testCase.readResponseSchema, testCase.readParams),
					UpdateOp:               createTestUpdateOp(testCase.updateRequestSchema),
					CreateCommonParameters: testCase.createCommonParams,
					SchemaOptions:          testCase.schemaOptions,
				},
			}, config.Config{})
			got, err := mapper.MapToIR(slog.Default())
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected only one resource, got: %d", len(got))
			}

			if diff := cmp.Diff(got[0].Schema.Attributes, testCase.want); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

// Two read parameters that map to the same attribute name have no deterministic merge: keeping
// either one silently discards the other. The resource is rejected so the specification can be fixed.
func TestResourceMapper_parameter_name_collisions(t *testing.T) {
	t.Parallel()

	stringObject := base.CreateSchemaProxy(&base.Schema{
		Type: []string{"object"},
		Properties: orderedmap.ToOrderedMap(map[string]*base.SchemaProxy{
			"name": base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
		}),
	})
	zonePath := &high.Parameter{
		Name:     "zone",
		In:       "path",
		Required: new(true),
		Schema:   base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
	}

	testCases := map[string]struct {
		readParams    []*high.Parameter
		schemaOptions explorer.SchemaOptions
		wantLog       string
	}{
		"same name in path and query": {
			readParams: []*high.Parameter{
				zonePath,
				{
					Name:   "zone",
					In:     "query",
					Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
				},
			},
			wantLog: "read operation parameters 'zone' (in: path) and 'zone' (in: query) both map to attribute 'zone'",
		},
		"alias onto another parameter's name": {
			readParams: []*high.Parameter{
				zonePath,
				{
					Name:   "region",
					In:     "query",
					Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
				},
			},
			schemaOptions: explorer.SchemaOptions{
				AttributeOptions: explorer.AttributeOptions{
					Aliases: map[string]string{"region": "zone"},
				},
			},
			wantLog: "read operation parameters 'zone' (in: path) and 'region' (in: query) both map to attribute 'zone'",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var logs bytes.Buffer
			mapper := mapper.NewResourceMapper(map[string]explorer.Resource{
				"test_resource": {
					CreateOp:      createTestCreateOp(stringObject, nil, []*high.Parameter{zonePath}),
					ReadOp:        createTestReadOp(stringObject, testCase.readParams),
					SchemaOptions: testCase.schemaOptions,
				},
			}, config.Config{})
			got, err := mapper.MapToIR(slog.New(slog.NewTextHandler(&logs, nil)))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(got) != 0 {
				t.Fatalf("expected the resource to be skipped, got: %+v", got)
			}

			if !strings.Contains(logs.String(), testCase.wantLog) {
				t.Errorf("expected log to contain %q, got: %s", testCase.wantLog, logs.String())
			}
		})
	}
}

func createTestCreateOp(request *base.SchemaProxy, response *base.SchemaProxy, params []*high.Parameter) *high.Operation {
	return &high.Operation{
		Parameters: params,
		RequestBody: &high.RequestBody{
			Content: orderedmap.ToOrderedMap(map[string]*high.MediaType{
				"application/json": {
					Schema: request,
				},
			}),
		},
		Responses: &high.Responses{
			Codes: orderedmap.ToOrderedMap(map[string]*high.Response{
				"201": {
					Content: orderedmap.ToOrderedMap(map[string]*high.MediaType{
						"application/json": {
							Schema: response,
						},
					}),
				},
			}),
		},
	}
}

func createTestUpdateOp(request *base.SchemaProxy) *high.Operation {
	if request == nil {
		return nil
	}

	return &high.Operation{
		RequestBody: &high.RequestBody{
			Content: orderedmap.ToOrderedMap(map[string]*high.MediaType{
				"application/json": {
					Schema: request,
				},
			}),
		},
	}
}

func createTestReadOp(response *base.SchemaProxy, params []*high.Parameter) *high.Operation {
	return &high.Operation{
		Responses: &high.Responses{
			Codes: orderedmap.ToOrderedMap(map[string]*high.Response{
				"200": {
					Content: orderedmap.ToOrderedMap(map[string]*high.MediaType{
						"application/json": {
							Schema: response,
						},
					}),
				},
			}),
		},
		Parameters: params,
	}
}
