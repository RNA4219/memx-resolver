package api

import "context"

func (c *HTTPClient) DocsIngest(ctx context.Context, req DocsIngestRequest) (DocsIngestResponse, *Error) {
	var out DocsIngestResponse
	if err := c.post(ctx, "/v1/docs:ingest", req, &out); err != nil {
		return DocsIngestResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) DocsResolve(ctx context.Context, req DocsResolveRequest) (DocsResolveResponse, *Error) {
	var out DocsResolveResponse
	if err := c.post(ctx, "/v1/docs:resolve", req, &out); err != nil {
		return DocsResolveResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) ChunksGet(ctx context.Context, req ChunksGetRequest) (ChunksGetResponse, *Error) {
	var out ChunksGetResponse
	if err := c.post(ctx, "/v1/chunks:get", req, &out); err != nil {
		return ChunksGetResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) DocsSearch(ctx context.Context, req DocsSearchRequest) (DocsSearchResponse, *Error) {
	var out DocsSearchResponse
	if err := c.post(ctx, "/v1/docs:search", req, &out); err != nil {
		return DocsSearchResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) CardsSearch(ctx context.Context, req CardsSearchRequest) (CardsSearchResponse, *Error) {
	var out CardsSearchResponse
	if err := c.post(ctx, "/v1/cards:search", req, &out); err != nil {
		return CardsSearchResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) CardFeedback(ctx context.Context, req CardFeedbackRequest) (CardFeedbackResponse, *Error) {
	var out CardFeedbackResponse
	if err := c.post(ctx, "/v1/cards:feedback", req, &out); err != nil {
		return CardFeedbackResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) PromptBundle(ctx context.Context, req PromptBundleRequest) (PromptBundleResponse, *Error) {
	var out PromptBundleResponse
	if err := c.post(ctx, "/v1/cards:bundle", req, &out); err != nil {
		return PromptBundleResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) TaskStateExport(ctx context.Context, req TaskStateExportRequest) (TaskStateExportResponse, *Error) {
	var out TaskStateExportResponse
	if err := c.post(ctx, "/v1/taskstate:export", req, &out); err != nil {
		return TaskStateExportResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) ReadsAck(ctx context.Context, req ReadsAckRequest) (ReadsAckResponse, *Error) {
	var out ReadsAckResponse
	if err := c.post(ctx, "/v1/reads:ack", req, &out); err != nil {
		return ReadsAckResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) DocsStaleCheck(ctx context.Context, req DocsStaleCheckRequest) (DocsStaleCheckResponse, *Error) {
	var out DocsStaleCheckResponse
	if err := c.post(ctx, "/v1/docs:stale-check", req, &out); err != nil {
		return DocsStaleCheckResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) ContractsResolve(ctx context.Context, req ContractsResolveRequest) (ContractsResolveResponse, *Error) {
	var out ContractsResolveResponse
	if err := c.post(ctx, "/v1/contracts:resolve", req, &out); err != nil {
		return ContractsResolveResponse{}, err
	}
	return out, nil
}
