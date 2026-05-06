// Copyright © 2021 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multi

import (
	"context"

	consensusclient "github.com/ethpandaops/go-eth2-client"
	"github.com/ethpandaops/go-eth2-client/api"
	"github.com/ethpandaops/go-eth2-client/spec"
	"github.com/ethpandaops/go-eth2-client/spec/all"
)

// BeaconState fetches a beacon state.
func (s *Service) BeaconState(ctx context.Context, opts *api.BeaconStateOpts) (*api.Response[*spec.VersionedBeaconState], error) {
	res, err := s.doCall(ctx, func(ctx context.Context, client consensusclient.Service) (any, error) {
		beaconState, err := client.(consensusclient.BeaconStateProvider).BeaconState(ctx, opts)
		if err != nil {
			return nil, err
		}

		return beaconState, nil
	}, nil)
	if err != nil {
		return nil, err
	}

	response, isResponse := res.(*api.Response[*spec.VersionedBeaconState])
	if !isResponse {
		return nil, ErrIncorrectType
	}

	return response, nil
}

// AgnosticBeaconState fetches a beacon state as a fork-agnostic *all.BeaconState.
func (s *Service) AgnosticBeaconState(ctx context.Context, opts *api.BeaconStateOpts) (*api.Response[*all.BeaconState], error) {
	res, err := s.doCall(ctx, func(ctx context.Context, client consensusclient.Service) (any, error) {
		state, err := client.(consensusclient.BeaconStateProvider).AgnosticBeaconState(ctx, opts)
		if err != nil {
			return nil, err
		}

		return state, nil
	}, nil)
	if err != nil {
		return nil, err
	}

	response, isResponse := res.(*api.Response[*all.BeaconState])
	if !isResponse {
		return nil, ErrIncorrectType
	}

	return response, nil
}
