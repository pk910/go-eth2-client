// Copyright © 2023, 2024 Attestant Limited.
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

	client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/deneb"
)

// BlobSidecars fetches the blobs sidecars given options.
func (s *Service) BlobSidecars(ctx context.Context,
	opts *api.BlobSidecarsOpts,
) (
	*api.Response[[]*deneb.BlobSidecar],
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

	endpoint := fmt.Sprintf("/eth/v1/beacon/blob_sidecars/%s", opts.Block)
	httpResponse, err := s.get(ctx, endpoint, "", &opts.Common, true)
	if err != nil {
		return nil, err
	}

	var response *api.Response[[]*deneb.BlobSidecar]
	switch httpResponse.contentType {
	case ContentTypeSSZ:
		response, err = s.blobSidecarsFromSSZ(httpResponse)
	case ContentTypeJSON:
		response, err = s.blobSidecarsFromJSON(httpResponse)
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (*Service) blobSidecarsFromSSZ(res *httpResponse) (*api.Response[[]*deneb.BlobSidecar], error) {
	response := &api.Response[[]*deneb.BlobSidecar]{}

	if len(res.body) == 0 {
		// This is a valid response when there are no blobs for the request.
		response.Data = make([]*deneb.BlobSidecar, 0)
		response.Metadata = make(map[string]any)

		return response, nil
	}

	data := &api.BlobSidecars{}
	if err := data.UnmarshalSSZ(res.body); err != nil {
		return nil, errors.Join(errors.New("failed to decode blob sidecars"), err)
	}

	response.Data = data.Sidecars

	return response, nil
}

func (*Service) blobSidecarsFromJSON(res *httpResponse) (*api.Response[[]*deneb.BlobSidecar], error) {
	response := &api.Response[[]*deneb.BlobSidecar]{}

	var err error
	response.Data, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), []*deneb.BlobSidecar{})
	if err != nil {
		return nil, err
	}

	return response, nil
}
