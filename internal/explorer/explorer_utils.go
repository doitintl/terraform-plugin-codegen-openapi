// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package explorer

import high "github.com/pb33f/libopenapi/datamodel/high/v3"

func mergeParameters(commonParameters []*high.Parameter, operation *high.Operation) []*high.Parameter {
	mergedParameters := make([]*high.Parameter, len(commonParameters))
	copy(mergedParameters, commonParameters)
	if operation != nil {
		for _, operationParameter := range operation.Parameters {
			found := false
			for i, mergedParameter := range mergedParameters {
				// A parameter is identified by name and location, so an operation parameter only
				// overrides a path item parameter when both match. Matching on the name alone
				// would let, say, an operation's `zone` query parameter displace the path item's
				// `zone` path parameter, which the specification treats as a different parameter.
				if operationParameter.Name == mergedParameter.Name && operationParameter.In == mergedParameter.In {
					found = true
					mergedParameters[i] = operationParameter
					break
				}
			}
			if !found {
				mergedParameters = append(mergedParameters, operationParameter) // nolint: makezero
			}
		}
	}
	return mergedParameters
}

func (e *Resource) ReadOpParameters() []*high.Parameter {
	return mergeParameters(e.CommonParameters, e.ReadOp)
}

func (e *Resource) CreateOpParameters() []*high.Parameter {
	return mergeParameters(e.CreateCommonParameters, e.CreateOp)
}

func (e *DataSource) ReadOpParameters() []*high.Parameter {
	return mergeParameters(e.CommonParameters, e.ReadOp)
}
