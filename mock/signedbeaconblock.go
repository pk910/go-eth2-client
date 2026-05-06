// Copyright © 2020, 2023 Attestant Limited.
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

package mock

import (
	"context"

	"github.com/ethpandaops/go-eth2-client/api"
	"github.com/ethpandaops/go-eth2-client/spec"
	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
)

// SignedBeaconBlock fetches a signed beacon block given a block ID.
func (s *Service) SignedBeaconBlock(ctx context.Context,
	opts *api.SignedBeaconBlockOpts,
) (
	*api.Response[*spec.VersionedSignedBeaconBlock],
	error,
) {
	if s.SignedBeaconBlockFunc != nil {
		return s.SignedBeaconBlockFunc(ctx, opts)
	}

	return &api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Version: spec.DataVersionPhase0,
			Phase0: &phase0.SignedBeaconBlock{
				Message: &phase0.BeaconBlock{
					Body: &phase0.BeaconBlockBody{
						ETH1Data: &phase0.ETH1Data{},
					},
				},
			},
		},
		Metadata: make(map[string]any),
	}, nil
}

// AgnosticSignedBeaconBlock returns a stub fork-agnostic signed beacon block.
func (s *Service) AgnosticSignedBeaconBlock(ctx context.Context,
	opts *api.SignedBeaconBlockOpts,
) (
	*api.Response[*all.SignedBeaconBlock],
	error,
) {
	if s.AgnosticSignedBeaconBlockFunc != nil {
		return s.AgnosticSignedBeaconBlockFunc(ctx, opts)
	}

	return &api.Response[*all.SignedBeaconBlock]{
		Data: &all.SignedBeaconBlock{
			Version: version.DataVersionPhase0,
			Message: &all.BeaconBlock{
				Version: version.DataVersionPhase0,
				Body: &all.BeaconBlockBody{
					Version:  version.DataVersionPhase0,
					ETH1Data: &phase0.ETH1Data{},
				},
			},
		},
		Metadata: make(map[string]any),
	}, nil
}
