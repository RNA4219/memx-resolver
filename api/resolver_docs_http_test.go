package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP_DocsEndpoints(t *testing.T) {
	server, cleanup := setupHTTPTestServer(t)
	defer cleanup()

	ctx := context.Background()
	_, apiErr := server.InProc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "spec",
		Title:       "Resolver API Spec",
		Version:     "2026-03-10",
		FeatureKeys: []string{"resolver-api"},
		TaskIDs:     []string{"task:feature:local:resolver"},
		Body: `# Resolver API Spec

## Acceptance Criteria
- returns required docs`,
	})
	if apiErr != nil {
		t.Fatalf("DocsIngest: %s", apiErr.Message)
	}

	jsonBody, _ := json.Marshal(DocsResolveRequest{Feature: "resolver-api"})
	resolveReq := httptest.NewRequest(http.MethodPost, "/v1/docs:resolve", bytes.NewReader(jsonBody))
	resolveReq.Header.Set("Content-Type", "application/json")
	resolveRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(resolveRec, resolveReq)
	if resolveRec.Code != http.StatusOK {
		t.Fatalf("resolve status=%d body=%s", resolveRec.Code, resolveRec.Body.String())
	}
	var resolveResp DocsResolveResponse
	if err := json.Unmarshal(resolveRec.Body.Bytes(), &resolveResp); err != nil {
		t.Fatalf("decode resolve response: %v", err)
	}
	if len(resolveResp.Required) != 1 {
		t.Fatalf("unexpected resolve response: %#v", resolveResp)
	}

	chunksBody, _ := json.Marshal(ChunksGetRequest{DocID: resolveResp.Required[0].DocID})
	chunksReq := httptest.NewRequest(http.MethodPost, "/v1/chunks:get", bytes.NewReader(chunksBody))
	chunksReq.Header.Set("Content-Type", "application/json")
	chunksRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(chunksRec, chunksReq)
	if chunksRec.Code != http.StatusOK {
		t.Fatalf("chunks status=%d body=%s", chunksRec.Code, chunksRec.Body.String())
	}
	var chunksResp ChunksGetResponse
	if err := json.Unmarshal(chunksRec.Body.Bytes(), &chunksResp); err != nil {
		t.Fatalf("decode chunks response: %v", err)
	}
	if len(chunksResp.MemoryCards) == 0 || chunksResp.MemoryCards[0].MemoryType != "acceptance" {
		t.Fatalf("expected prompt-ready memory cards, got %#v", chunksResp.MemoryCards)
	}

	cardsBody, _ := json.Marshal(CardsSearchRequest{
		Query:       "required docs",
		FeatureKeys: []string{"resolver-api"},
		MemoryTypes: []string{"acceptance"},
		Limit:       2,
		TokenBudget: 40,
	})
	cardsReq := httptest.NewRequest(http.MethodPost, "/v1/cards:search", bytes.NewReader(cardsBody))
	cardsReq.Header.Set("Content-Type", "application/json")
	cardsRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(cardsRec, cardsReq)
	if cardsRec.Code != http.StatusOK {
		t.Fatalf("cards status=%d body=%s", cardsRec.Code, cardsRec.Body.String())
	}
	var cardsResp CardsSearchResponse
	if err := json.Unmarshal(cardsRec.Body.Bytes(), &cardsResp); err != nil {
		t.Fatalf("decode cards response: %v", err)
	}
	if len(cardsResp.Cards) != 1 || cardsResp.Cards[0].MemoryType != "acceptance" || cardsResp.Cards[0].Score <= 0 {
		t.Fatalf("unexpected cards response: %#v", cardsResp)
	}

	feedbackBody, _ := json.Marshal(CardFeedbackRequest{
		CardID:     cardsResp.Cards[0].CardID,
		DocID:      cardsResp.Cards[0].DocID,
		ChunkID:    cardsResp.Cards[0].ChunkID,
		MemoryType: cardsResp.Cards[0].MemoryType,
		Signal:     "helpful",
		Query:      "required docs",
	})
	feedbackReq := httptest.NewRequest(http.MethodPost, "/v1/cards:feedback", bytes.NewReader(feedbackBody))
	feedbackReq.Header.Set("Content-Type", "application/json")
	feedbackRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(feedbackRec, feedbackReq)
	if feedbackRec.Code != http.StatusOK {
		t.Fatalf("feedback status=%d body=%s", feedbackRec.Code, feedbackRec.Body.String())
	}

	bundleBody, _ := json.Marshal(PromptBundleRequest{Query: "required docs", Feature: "resolver-api", Limit: 2})
	bundleReq := httptest.NewRequest(http.MethodPost, "/v1/cards:bundle", bytes.NewReader(bundleBody))
	bundleReq.Header.Set("Content-Type", "application/json")
	bundleRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(bundleRec, bundleReq)
	if bundleRec.Code != http.StatusOK {
		t.Fatalf("bundle status=%d body=%s", bundleRec.Code, bundleRec.Body.String())
	}
	var bundleResp PromptBundleResponse
	if err := json.Unmarshal(bundleRec.Body.Bytes(), &bundleResp); err != nil {
		t.Fatalf("decode bundle response: %v", err)
	}
	if bundleResp.Prompt == "" || len(bundleResp.SourceRefs) == 0 {
		t.Fatalf("unexpected bundle response: %#v", bundleResp)
	}

	ackBody, _ := json.Marshal(ReadsAckRequest{TaskID: "task:feature:local:resolver", DocID: resolveResp.Required[0].DocID})
	ackReq := httptest.NewRequest(http.MethodPost, "/v1/reads:ack", bytes.NewReader(ackBody))
	ackReq.Header.Set("Content-Type", "application/json")
	ackRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(ackRec, ackReq)
	if ackRec.Code != http.StatusOK {
		t.Fatalf("ack status=%d body=%s", ackRec.Code, ackRec.Body.String())
	}
	var ackResp ReadsAckResponse
	if err := json.Unmarshal(ackRec.Body.Bytes(), &ackResp); err != nil {
		t.Fatalf("decode ack response: %v", err)
	}
	if len(ackResp.Receipt.ChunkSnapshots) == 0 {
		t.Fatalf("expected read receipt snapshots, got %#v", ackResp)
	}

	staleBody, _ := json.Marshal(DocsStaleCheckRequest{TaskID: "task:feature:local:resolver"})
	staleReq := httptest.NewRequest(http.MethodPost, "/v1/docs:stale-check", bytes.NewReader(staleBody))
	staleReq.Header.Set("Content-Type", "application/json")
	staleRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(staleRec, staleReq)
	if staleRec.Code != http.StatusOK {
		t.Fatalf("stale status=%d body=%s", staleRec.Code, staleRec.Body.String())
	}
	var staleResp DocsStaleCheckResponse
	if err := json.Unmarshal(staleRec.Body.Bytes(), &staleResp); err != nil {
		t.Fatalf("decode stale response: %v", err)
	}
	if len(staleResp.Stale) != 0 {
		t.Fatalf("expected no stale docs, got %#v", staleResp.Stale)
	}
	if staleResp.Status != "fresh" {
		t.Fatalf("expected fresh status, got %q", staleResp.Status)
	}
	if len(staleResp.StaleReasons) != 0 {
		t.Fatalf("expected no stale reasons, got %#v", staleResp.StaleReasons)
	}

	_, apiErr = server.InProc.DocsIngest(ctx, DocsIngestRequest{
		DocID:       resolveResp.Required[0].DocID,
		DocType:     "spec",
		Title:       "Resolver API Spec",
		Version:     "2026-03-11",
		FeatureKeys: []string{"resolver-api"},
		TaskIDs:     []string{"task:feature:local:resolver"},
		Body: `# Resolver API Spec

## Acceptance Criteria
- returns required docs after update`,
	})
	if apiErr != nil {
		t.Fatalf("DocsIngest update: %s", apiErr.Message)
	}

	staleReqAfterUpdate := httptest.NewRequest(http.MethodPost, "/v1/docs:stale-check", bytes.NewReader(staleBody))
	staleReqAfterUpdate.Header.Set("Content-Type", "application/json")
	staleRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(staleRec, staleReqAfterUpdate)
	if staleRec.Code != http.StatusOK {
		t.Fatalf("stale after update status=%d body=%s", staleRec.Code, staleRec.Body.String())
	}
	if err := json.Unmarshal(staleRec.Body.Bytes(), &staleResp); err != nil {
		t.Fatalf("decode stale response after update: %v", err)
	}
	if staleResp.Status != "stale" {
		t.Fatalf("expected stale status, got %q", staleResp.Status)
	}
	if len(staleResp.StaleReasons) != 1 {
		t.Fatalf("expected one stale reason, got %#v", staleResp.StaleReasons)
	}
	if len(staleResp.Stale) != 1 {
		t.Fatalf("expected backward-compatible stale field, got %#v", staleResp.Stale)
	}
	if staleResp.StaleReasons[0].Reason != "semantic_diff" || len(staleResp.StaleReasons[0].ImpactScope) == 0 {
		t.Fatalf("expected semantic stale impact, got %#v", staleResp.StaleReasons[0])
	}

	exportBody, _ := json.Marshal(TaskStateExportRequest{TaskID: "task:feature:local:resolver", Feature: "resolver-api"})
	exportReq := httptest.NewRequest(http.MethodPost, "/v1/taskstate:export", bytes.NewReader(exportBody))
	exportReq.Header.Set("Content-Type", "application/json")
	exportRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("taskstate export status=%d body=%s", exportRec.Code, exportRec.Body.String())
	}
	var exportResp TaskStateExportResponse
	if err := json.Unmarshal(exportRec.Body.Bytes(), &exportResp); err != nil {
		t.Fatalf("decode taskstate export response: %v", err)
	}
	if exportResp.TaskRef != "agent-taskstate:task:local:task_feature_local_resolver" || len(exportResp.SourceRefs) == 0 {
		t.Fatalf("unexpected taskstate export response: %#v", exportResp)
	}
}
