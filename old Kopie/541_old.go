/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package noop

import (
	"encoding/json"

	dto_discovery "github.com/mysteriumnetwork/node/service_discovery/dto"
)

// Bootstrap is called on program initialization time and registers various deserializers related to opepnvpn service
func Bootstrap() {
	dto_discovery.RegisterServiceDefinitionUnserializer(
		ServiceType,
		func(rawDefinition *json.RawMessage) (dto_discovery.ServiceDefinition, error) {
			var definition ServiceDefinition
			err := json.Unmarshal(*rawDefinition, &definition)

			return definition, err
		},
	)

	dto_discovery.RegisterPaymentMethodUnserializer(
		PaymentMethodNoop,
		func(rawDefinition *json.RawMessage) (dto_discovery.PaymentMethod, error) {
			var method PaymentNoop
			err := json.Unmarshal(*rawDefinition, &method)

			return method, err
		},
	)
}
