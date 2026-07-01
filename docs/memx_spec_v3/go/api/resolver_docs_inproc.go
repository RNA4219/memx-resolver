package api

import (
	"context"

	"memx/service"
)

func (c *InProcClient) DocsIngest(ctx context.Context, req DocsIngestRequest) (DocsIngestResponse, *Error) {
	doc, chunkCount, err := c.Svc.DocsIngest(ctx, req)
	if err != nil {
		return DocsIngestResponse{}, mapError(err)
	}
	return DocsIngestResponse{DocID: doc.DocID, Version: doc.Version, ChunkCount: chunkCount, Status: "ingested"}, nil
}

func (c *InProcClient) DocsResolve(ctx context.Context, req DocsResolveRequest) (DocsResolveResponse, *Error) {
	required, recommended, err := c.Svc.DocsResolve(ctx, req)
	if err != nil {
		return DocsResolveResponse{}, mapError(err)
	}
	if required == nil {
		required = []ResolveEntry{}
	}
	if recommended == nil {
		recommended = []ResolveEntry{}
	}
	return DocsResolveResponse{Required: required, Recommended: recommended}, nil
}

func (c *InProcClient) ChunksGet(ctx context.Context, req ChunksGetRequest) (ChunksGetResponse, *Error) {
	docID, chunks, err := c.Svc.ChunksGet(ctx, req)
	if err != nil {
		return ChunksGetResponse{}, mapError(err)
	}
	if chunks == nil {
		chunks = []ResolverChunk{}
	}
	cards := service.BuildRankedResolverMemoryCards(chunks, req.Query, 64, 0)
	if cards == nil {
		cards = []ResolverMemoryCard{}
	}
	return ChunksGetResponse{DocID: docID, Chunks: chunks, MemoryCards: cards}, nil
}

func (c *InProcClient) DocsSearch(ctx context.Context, req DocsSearchRequest) (DocsSearchResponse, *Error) {
	results, err := c.Svc.DocsSearch(ctx, req)
	if err != nil {
		return DocsSearchResponse{}, mapError(err)
	}
	if results == nil {
		results = []ResolveEntry{}
	}
	return DocsSearchResponse{Results: results}, nil
}

func (c *InProcClient) CardsSearch(ctx context.Context, req CardsSearchRequest) (CardsSearchResponse, *Error) {
	cards, err := c.Svc.CardsSearch(ctx, req)
	if err != nil {
		return CardsSearchResponse{}, mapError(err)
	}
	if cards == nil {
		cards = []ResolverMemoryCard{}
	}
	return CardsSearchResponse{Cards: cards}, nil
}

func (c *InProcClient) CardFeedback(ctx context.Context, req CardFeedbackRequest) (CardFeedbackResponse, *Error) {
	feedback, err := c.Svc.CardFeedback(ctx, req)
	if err != nil {
		return CardFeedbackResponse{}, mapError(err)
	}
	return CardFeedbackResponse{Feedback: feedback}, nil
}

func (c *InProcClient) PromptBundle(ctx context.Context, req PromptBundleRequest) (PromptBundleResponse, *Error) {
	bundle, err := c.Svc.PromptBundle(ctx, req)
	if err != nil {
		return PromptBundleResponse{}, mapError(err)
	}
	if bundle.Cards == nil {
		bundle.Cards = []ResolverMemoryCard{}
	}
	if bundle.SourceRefs == nil {
		bundle.SourceRefs = []string{}
	}
	return bundle, nil
}

func (c *InProcClient) TaskStateExport(ctx context.Context, req TaskStateExportRequest) (TaskStateExportResponse, *Error) {
	out, err := c.Svc.TaskStateExport(ctx, req)
	if err != nil {
		return TaskStateExportResponse{}, mapError(err)
	}
	if out.RequiredDocs == nil {
		out.RequiredDocs = []ResolveEntry{}
	}
	if out.ReadReceipts == nil {
		out.ReadReceipts = []ResolverReadReceipt{}
	}
	if out.StaleReasons == nil {
		out.StaleReasons = []ResolverStaleReason{}
	}
	if out.SourceRefs == nil {
		out.SourceRefs = []string{}
	}
	return out, nil
}

func (c *InProcClient) ReadsAck(ctx context.Context, req ReadsAckRequest) (ReadsAckResponse, *Error) {
	receipt, err := c.Svc.ReadsAck(ctx, req)
	if err != nil {
		return ReadsAckResponse{}, mapError(err)
	}
	return ReadsAckResponse{Receipt: receipt}, nil
}

func (c *InProcClient) DocsStaleCheck(ctx context.Context, req DocsStaleCheckRequest) (DocsStaleCheckResponse, *Error) {
	stale, err := c.Svc.DocsStaleCheck(ctx, req)
	if err != nil {
		return DocsStaleCheckResponse{}, mapError(err)
	}
	if stale == nil {
		stale = []ResolverStaleReason{}
	}
	status := "fresh"
	if len(stale) > 0 {
		status = "stale"
	}
	return DocsStaleCheckResponse{TaskID: req.TaskID, Status: status, StaleReasons: stale, Stale: stale}, nil
}

func (c *InProcClient) ContractsResolve(ctx context.Context, req ContractsResolveRequest) (ContractsResolveResponse, *Error) {
	required, acceptance, forbidden, done, dependencies, err := c.Svc.ContractsResolve(ctx, req)
	if err != nil {
		return ContractsResolveResponse{}, mapError(err)
	}
	if required == nil {
		required = []ResolveEntry{}
	}
	if acceptance == nil {
		acceptance = []string{}
	}
	if forbidden == nil {
		forbidden = []string{}
	}
	if done == nil {
		done = []string{}
	}
	if dependencies == nil {
		dependencies = []string{}
	}
	return ContractsResolveResponse{
		Required:           required,
		AcceptanceCriteria: acceptance,
		ForbiddenPatterns:  forbidden,
		DefinitionOfDone:   done,
		Dependencies:       dependencies,
	}, nil
}
