/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause

   Generated by: https://github.com/swagger-api/swagger-codegen.git */

package upgrade

// Gateway statistics
type GatewayStats struct {

	// Gateways with status DEPLOYING
	Deploying int64 `json:"deploying,omitempty"`

	// Gateways with status DOWN
	Down int64 `json:"down,omitempty"`

	// Gateways with status UP
	Up int64 `json:"up,omitempty"`
}
