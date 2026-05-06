// Copyright © 2020 - 2024 Attestant Limited.
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

package http

import (
	"context"
	"errors"
	"fmt"

	client "github.com/ethpandaops/go-eth2-client"
	"github.com/ethpandaops/go-eth2-client/api"
	"github.com/ethpandaops/go-eth2-client/spec"
	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
)

// SignedExecutionPayloadEnvelope fetches a signed execution payload envelope
// given a block ID and decodes it directly into the per-fork view stored on a
// *spec.VersionedSignedExecutionPayloadEnvelope. No intermediate copy.
func (s *Service) SignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope],
	error,
) {
	httpResponse, err := s.fetchSignedExecutionPayloadEnvelope(ctx, opts)
	if err != nil {
		return nil, err
	}

	// All current envelope-bearing forks (Gloas, Heze) reuse the gloas
	// schema, so the wire bytes always parse into *gloas.SignedExecutionPayloadEnvelope.
	envelope := &gloas.SignedExecutionPayloadEnvelope{}

	switch httpResponse.contentType {
	case ContentTypeSSZ:
		ds, err := s.dynSSZForRequest(ctx)
		if err != nil {
			return nil, err
		}

		if err := ds.UnmarshalSSZ(envelope, httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s signed execution payload envelope", httpResponse.consensusVersion), err)
		}
	case ContentTypeJSON:
		if err := envelope.UnmarshalJSON(httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s signed execution payload envelope", httpResponse.consensusVersion), err)
		}
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}

	return &api.Response[*spec.VersionedSignedExecutionPayloadEnvelope]{
		Data: &spec.VersionedSignedExecutionPayloadEnvelope{
			Version: httpResponse.consensusVersion,
			Gloas:   envelope,
		},
		Metadata: metadataFromHeaders(httpResponse.headers),
	}, nil
}

// AgnosticSignedExecutionPayloadEnvelope fetches a signed execution payload
// envelope and decodes it directly into a fork-agnostic
// *all.SignedExecutionPayloadEnvelope. The Version is set from the consensus
// version header before unmarshaling so the union type's view-aware codec
// dispatches into the correct fork's schema. No intermediate copy.
func (s *Service) AgnosticSignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*all.SignedExecutionPayloadEnvelope],
	error,
) {
	httpResponse, err := s.fetchSignedExecutionPayloadEnvelope(ctx, opts)
	if err != nil {
		return nil, err
	}

	envelope := &all.SignedExecutionPayloadEnvelope{Version: httpResponse.consensusVersion}

	switch httpResponse.contentType {
	case ContentTypeSSZ:
		ds, err := s.dynSSZForRequest(ctx)
		if err != nil {
			return nil, err
		}

		if err := envelope.UnmarshalSSZDyn(ds, httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s signed execution payload envelope", httpResponse.consensusVersion), err)
		}
	case ContentTypeJSON:
		if err := envelope.UnmarshalJSON(httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s signed execution payload envelope", httpResponse.consensusVersion), err)
		}
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}

	return &api.Response[*all.SignedExecutionPayloadEnvelope]{
		Data:     envelope,
		Metadata: metadataFromHeaders(httpResponse.headers),
	}, nil
}

// fetchSignedExecutionPayloadEnvelope performs the GET request shared by both
// SignedExecutionPayloadEnvelope and AgnosticSignedExecutionPayloadEnvelope:
// validates opts, hits the endpoint, and rejects responses for forks where
// the envelope doesn't apply.
func (s *Service) fetchSignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (*httpResponse, error) {
	if err := s.assertIsActive(ctx); err != nil {
		return nil, err
	}

	if opts == nil {
		return nil, client.ErrNoOptions
	}

	if opts.Block == "" {
		return nil, errors.Join(errors.New("no block specified"), client.ErrInvalidOptions)
	}

	endpoint := fmt.Sprintf("/eth/v1/beacon/execution_payload_envelope/%s", opts.Block)

	httpResponse, err := s.get(ctx, endpoint, "", &opts.Common, true)
	if err != nil {
		return nil, err
	}

	if httpResponse.consensusVersion != spec.DataVersionGloas && httpResponse.consensusVersion != spec.DataVersionHeze {
		return nil, fmt.Errorf("execution payload envelope not available for block version %s", httpResponse.consensusVersion)
	}

	return httpResponse, nil
}
