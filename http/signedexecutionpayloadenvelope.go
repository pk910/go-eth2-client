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
	"bytes"
	"context"
	"errors"
	"fmt"

	client "github.com/ethpandaops/go-eth2-client"
	"github.com/ethpandaops/go-eth2-client/api"
	"github.com/ethpandaops/go-eth2-client/spec"
	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
)

// SignedExecutionPayloadEnvelope fetches a signed execution payload envelope given a block ID.
func (s *Service) SignedExecutionPayloadEnvelope(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope],
	error,
) {
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

	var response *api.Response[*spec.VersionedSignedExecutionPayloadEnvelope]
	switch httpResponse.contentType {
	case ContentTypeSSZ:
		response, err = s.signedExecutionPayloadEnvelopeFromSSZ(ctx, httpResponse)
	case ContentTypeJSON:
		response, err = s.signedExecutionPayloadEnvelopeFromJSON(httpResponse)
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) signedExecutionPayloadEnvelopeFromSSZ(ctx context.Context,
	res *httpResponse,
) (
	*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope],
	error,
) {
	if res.consensusVersion != spec.DataVersionGloas && res.consensusVersion != spec.DataVersionHeze {
		return nil, fmt.Errorf("execution payload envelope not available for block version %s", res.consensusVersion)
	}

	dynSSZ, err := s.dynSSZForRequest(ctx)
	if err != nil {
		return nil, err
	}

	envelope := &gloas.SignedExecutionPayloadEnvelope{}
	if err := dynSSZ.UnmarshalSSZ(envelope, res.body); err != nil {
		return nil, errors.Join(errors.New("failed to decode gloas signed execution payload envelope contents"), err)
	}

	return &api.Response[*spec.VersionedSignedExecutionPayloadEnvelope]{
		Data: &spec.VersionedSignedExecutionPayloadEnvelope{
			Version: res.consensusVersion,
			Gloas:   envelope,
		},
		Metadata: metadataFromHeaders(res.headers),
	}, nil
}

// AgnosticExecutionPayload fetches a signed execution payload envelope and
// returns the inner ExecutionPayload as a fork-agnostic *all.ExecutionPayload.
// Envelope-specific fields and the signature are dropped; use
// SignedExecutionPayloadEnvelope when those are needed.
func (s *Service) AgnosticExecutionPayload(ctx context.Context,
	opts *api.SignedExecutionPayloadEnvelopeOpts,
) (
	*api.Response[*all.ExecutionPayload],
	error,
) {
	versioned, err := s.SignedExecutionPayloadEnvelope(ctx, opts)
	if err != nil {
		return nil, err
	}

	if versioned.Data == nil ||
		versioned.Data.Gloas == nil ||
		versioned.Data.Gloas.Message == nil ||
		versioned.Data.Gloas.Message.Payload == nil {
		return nil, errors.New("execution payload envelope contains no payload")
	}

	// The envelope's Payload is *gloas.ExecutionPayload regardless of whether
	// the response was Gloas or Heze (Heze reuses the Gloas execution-payload
	// schema), so pin the Version explicitly from the versioned wrapper rather
	// than letting FromView's type-switch downgrade Heze to Gloas.
	payload := &all.ExecutionPayload{Version: versioned.Data.Version}
	if err := payload.FromView(versioned.Data.Gloas.Message.Payload); err != nil {
		return nil, errors.Join(errors.New("failed to convert execution payload to agnostic"), err)
	}

	return &api.Response[*all.ExecutionPayload]{
		Data:     payload,
		Metadata: versioned.Metadata,
	}, nil
}

func (*Service) signedExecutionPayloadEnvelopeFromJSON(res *httpResponse) (
	*api.Response[*spec.VersionedSignedExecutionPayloadEnvelope],
	error,
) {
	if res.consensusVersion != spec.DataVersionGloas && res.consensusVersion != spec.DataVersionHeze {
		return nil, fmt.Errorf("execution payload envelope not available for block version %s", res.consensusVersion)
	}

	envelope, metadata, err := decodeJSONResponse(bytes.NewReader(res.body),
		&gloas.SignedExecutionPayloadEnvelope{},
	)
	if err != nil {
		return nil, err
	}

	return &api.Response[*spec.VersionedSignedExecutionPayloadEnvelope]{
		Data: &spec.VersionedSignedExecutionPayloadEnvelope{
			Version: res.consensusVersion,
			Gloas:   envelope,
		},
		Metadata: metadata,
	}, nil
}
