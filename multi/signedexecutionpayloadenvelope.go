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

// SignedExecutionPayloadEnvelope fetches a signed execution payload envelope given a block ID.
func (s *Service) SignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope],
	error,
) {
	res, err := s.doCall(ctx, func(ctx context.Context, client consensusclient.Service) (any, error) {
		block, err := client.(consensusclient.ExecutionPayloadProvider).SignedExecutionPayloadEnvelope(ctx, opts)
		if err != nil {
			return nil, err
		}

		return block, nil
	}, nil)
	if err != nil {
		return nil, err
	}

	response, isResponse := res.(*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope])
	if !isResponse {
		return nil, ErrIncorrectType
	}

	return response, nil
}

// AgnosticSignedExecutionPayloadEnvelope fetches a signed execution payload
// envelope as a fork-agnostic *all.SignedExecutionPayloadEnvelope.
func (s *Service) AgnosticSignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*all.SignedExecutionPayloadEnvelope],
	error,
) {
	res, err := s.doCall(ctx, func(ctx context.Context, client consensusclient.Service) (any, error) {
		envelope, err := client.(consensusclient.ExecutionPayloadProvider).AgnosticSignedExecutionPayloadEnvelope(ctx, opts)
		if err != nil {
			return nil, err
		}

		return envelope, nil
	}, nil)
	if err != nil {
		return nil, err
	}

	response, isResponse := res.(*api.Response[*all.SignedExecutionPayloadEnvelope])
	if !isResponse {
		return nil, ErrIncorrectType
	}

	return response, nil
}
