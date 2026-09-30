// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapper

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/config"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/log"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/oas"
	"github.com/doitintl/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
	high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var _ ResourceMapper = resourceMapper{}

type ResourceMapper interface {
	MapToIR(*slog.Logger) ([]resource.Resource, error)
}

type resourceMapper struct {
	resources map[string]explorer.Resource
	//nolint:unused // Might be useful later!
	cfg config.Config
}

func NewResourceMapper(resources map[string]explorer.Resource, cfg config.Config) ResourceMapper {
	return resourceMapper{
		resources: resources,
		cfg:       cfg,
	}
}

func (m resourceMapper) MapToIR(logger *slog.Logger) ([]resource.Resource, error) {
	resourceSchemas := []resource.Resource{}

	// Guarantee the order of processing
	resourceNames := util.SortedKeys(m.resources)
	for _, name := range resourceNames {
		explorerResource := m.resources[name]
		rLogger := logger.With("resource", name)

		schema, err := generateResourceSchema(rLogger, explorerResource)
		if err != nil {
			log.WarnLogOnError(rLogger, err, "skipping resource schema mapping")
			continue
		}

		resourceSchemas = append(resourceSchemas, resource.Resource{
			Name:   name,
			Schema: schema,
		})
	}

	return resourceSchemas, nil
}

func generateResourceSchema(logger *slog.Logger, explorerResource explorer.Resource) (*resource.Schema, error) {
	resourceSchema := &resource.Schema{
		Attributes: []resource.Attribute{},
	}

	// ********************
	// Create Request Body (required)
	// ********************
	logger.Debug("searching for create operation request body")

	schemaOpts := oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	createRequestSchema, err := oas.BuildSchemaFromRequest(explorerResource.CreateOp, schemaOpts, oas.GlobalSchemaOpts{})
	if err != nil {
		return nil, err
	}
	createRequestAttributes, schemaErr := createRequestSchema.BuildResourceAttributes()
	if schemaErr != nil {
		return nil, schemaErr
	}

	// *********************
	// Create Response Body (optional)
	// *********************
	logger.Debug("searching for create operation response body")

	createResponseAttributes := attrmapper.ResourceAttributes{}
	schemaOpts = oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	globalSchemaOpts := oas.GlobalSchemaOpts{
		OverrideComputability: schema.Computed,
	}
	createResponseSchema, err := oas.BuildSchemaFromResponse(explorerResource.CreateOp, schemaOpts, globalSchemaOpts)
	if err != nil {
		if errors.Is(err, oas.ErrSchemaNotFound) {
			// Demote log to INFO if there was no schema found
			logger.Info("skipping mapping of create operation response body", "err", err)
		} else {
			logger.Warn("skipping mapping of create operation response body", "err", err)
		}
	} else {
		createResponseAttributes, schemaErr = createResponseSchema.BuildResourceAttributes()
		if schemaErr != nil {
			log.WarnLogOnError(logger, schemaErr, "skipping mapping of create operation response body")
		}
	}

	// *******************
	// READ Response Body (optional)
	// *******************
	logger.Debug("searching for read operation response body")

	readResponseAttributes := attrmapper.ResourceAttributes{}

	schemaOpts = oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	globalSchemaOpts = oas.GlobalSchemaOpts{
		OverrideComputability: schema.Computed,
	}
	readResponseSchema, err := oas.BuildSchemaFromResponse(explorerResource.ReadOp, schemaOpts, globalSchemaOpts)
	if err != nil {
		if errors.Is(err, oas.ErrSchemaNotFound) {
			// Demote log to INFO if there was no schema found
			logger.Info("skipping mapping of read operation response body", "err", err)
		} else {
			logger.Warn("skipping mapping of read operation response body", "err", err)
		}
	} else {
		readResponseAttributes, schemaErr = readResponseSchema.BuildResourceAttributes()
		if schemaErr != nil {
			log.WarnLogOnError(logger, schemaErr, "skipping mapping of read operation response body")
		}
	}

	// ****************
	// READ Parameters (optional)
	// ****************
	//
	// A read path parameter that also addresses the create operation is supplied by the
	// practitioner, so it maps to Required. One that only appears on the read path identifies
	// something the API assigned, so it stays ComputedOptional. `required` alone cannot tell the
	// two apart: on /parents/{parent_id}/children/{child_id} both are `required: true`.
	createPathParams := map[string]struct{}{}
	for _, param := range explorerResource.CreateOpParameters() {
		if param.In == util.OAS_param_path {
			createPathParams[param.Name] = struct{}{}
		}
	}

	readParameterAttributes := attrmapper.ResourceAttributes{}
	// Keyed by attribute name, i.e. after aliasing, since that is where two parameters collide.
	mappedParams := map[string]*high.Parameter{}
	for _, param := range explorerResource.ReadOpParameters() {
		if param.In != util.OAS_param_path && param.In != util.OAS_param_query {
			continue
		}

		pLogger := logger.With("param", param.Name)
		schemaOpts := oas.SchemaOpts{
			Ignores:             explorerResource.SchemaOptions.Ignores,
			OverrideDescription: param.Description,
		}
		// Nested properties of an object-typed parameter are not themselves addressed by the URL,
		// so the override stays ComputedOptional even when the parameter itself is Required.
		globalSchemaOpts := oas.GlobalSchemaOpts{OverrideComputability: schema.ComputedOptional}

		s, schemaErr := oas.BuildSchema(param.Schema, schemaOpts, globalSchemaOpts)
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		computability := schema.ComputedOptional
		if param.In == util.OAS_param_path && param.Required != nil && *param.Required {
			if _, ok := createPathParams[param.Name]; ok {
				computability = schema.Required
			}
		}

		// Check for any aliases and replace the paramater name if found
		paramName := param.Name
		if aliasedName, ok := explorerResource.SchemaOptions.AttributeOptions.Aliases[param.Name]; ok {
			pLogger = pLogger.With("param_alias", aliasedName)
			paramName = aliasedName
		}

		if s.IsPropertyIgnored(paramName) {
			continue
		}

		// Parameters are identified by name and location, but attributes only by name. Neither
		// parameter can win without silently discarding the other, so the specification (or the
		// aliases) must be fixed instead.
		if mapped, ok := mappedParams[paramName]; ok {
			return nil, fmt.Errorf("read operation parameters '%s' (in: %s) and '%s' (in: %s) both map to attribute '%s'", mapped.Name, mapped.In, param.Name, param.In, paramName)
		}
		mappedParams[paramName] = param

		parameterAttribute, schemaErr := s.BuildResourceAttribute(paramName, computability)
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		readParameterAttributes = append(readParameterAttributes, parameterAttribute)
	}

	// TODO: currently, no errors can be returned from merging, but in the future we should consider raising errors/warnings for unexpected scenarios, like type mismatches between attribute schemas
	resourceAttributes, _ := createRequestAttributes.Merge(createResponseAttributes, readResponseAttributes, readParameterAttributes)

	// TODO: handle error for overrides
	resourceAttributes, _ = resourceAttributes.ApplyOverrides(explorerResource.SchemaOptions.AttributeOptions.Overrides)

	resourceSchema.Attributes = resourceAttributes.ToSpec()

	// Set the resource-level description from OpenAPI tags if available
	if explorerResource.Description != "" {
		resourceSchema.Description = &explorerResource.Description
		resourceSchema.MarkdownDescription = &explorerResource.Description
	}

	return resourceSchema, nil
}
