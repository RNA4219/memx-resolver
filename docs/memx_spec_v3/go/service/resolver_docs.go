package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// DocsIngestRequest methods

func (r *DocsIngestRequest) normalize() {
	r.DocType = strings.TrimSpace(strings.ToLower(r.DocType))
	r.Title = strings.TrimSpace(r.Title)
	r.SourcePath = strings.TrimSpace(r.SourcePath)
	r.Version = strings.TrimSpace(r.Version)
	r.UpdatedAt = strings.TrimSpace(r.UpdatedAt)
	r.Summary = strings.TrimSpace(r.Summary)
	r.Body = strings.TrimSpace(r.Body)
	r.DocID = strings.TrimSpace(r.DocID)
	r.Chunking.Mode = strings.TrimSpace(strings.ToLower(r.Chunking.Mode))
	if r.Chunking.Mode == "" {
		r.Chunking.Mode = "heading"
	}
	if r.Chunking.MaxChars <= 0 {
		r.Chunking.MaxChars = 4000
	}
	if r.UpdatedAt == "" {
		r.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	r.Tags = uniqueTrimmedStrings(r.Tags)
	r.FeatureKeys = uniqueTrimmedStrings(r.FeatureKeys)
	r.TaskIDs = uniqueTrimmedStrings(r.TaskIDs)
	r.AcceptanceCriteria = uniqueTrimmedStrings(r.AcceptanceCriteria)
	r.ForbiddenPatterns = uniqueTrimmedStrings(r.ForbiddenPatterns)
	r.DefinitionOfDone = uniqueTrimmedStrings(r.DefinitionOfDone)
	r.Dependencies = uniqueTrimmedStrings(r.Dependencies)
	if r.DocID == "" {
		r.DocID = buildDocID(r.DocType, r.Title)
	}
}

func (r *DocsIngestRequest) validate() error {
	if r.DocType == "" || r.Title == "" || r.Version == "" || r.Body == "" {
		return fmt.Errorf("%w: doc_type/title/version/body is required", ErrInvalidArgument)
	}
	switch r.Chunking.Mode {
	case "heading", "fixed":
	default:
		return fmt.Errorf("%w: unsupported chunking.mode", ErrInvalidArgument)
	}
	return nil
}

// Service methods for resolver docs

func (s *Service) DocsIngest(ctx context.Context, req DocsIngestRequest) (ResolverDocument, int, error) {
	req.normalize()
	if err := req.validate(); err != nil {
		return ResolverDocument{}, 0, err
	}
	sections := parseSectionsForChunking(req.Title, req.Body, req.DocType, req.Chunking.Mode)
	if len(req.AcceptanceCriteria) == 0 {
		req.AcceptanceCriteria = extractSectionItems(sections, []string{"acceptance criteria", "acceptance"})
	}
	if len(req.ForbiddenPatterns) == 0 {
		req.ForbiddenPatterns = extractSectionItems(sections, []string{"forbidden patterns", "forbidden", "anti-pattern"})
	}
	if len(req.DefinitionOfDone) == 0 {
		req.DefinitionOfDone = extractSectionItems(sections, []string{"definition of done", "done"})
	}
	if len(req.Dependencies) == 0 {
		req.Dependencies = extractSectionItems(sections, []string{"dependency", "dependencies"})
	}
	if req.Summary == "" {
		req.Summary = buildDocumentSummary(req.Body)
	}
	if existing, err := s.getResolverDocument(ctx, req.DocID); err == nil {
		if compareVersions(req.Version, existing.Version) < 0 {
			return ResolverDocument{}, 0, fmt.Errorf("%w: version conflict", ErrConflict)
		}
	} else if err != ErrNotFound {
		return ResolverDocument{}, 0, err
	}
	doc := ResolverDocument{
		DocID:              req.DocID,
		DocType:            req.DocType,
		Title:              req.Title,
		SourcePath:         req.SourcePath,
		Version:            req.Version,
		UpdatedAt:          req.UpdatedAt,
		Summary:            req.Summary,
		Body:               req.Body,
		Tags:               req.Tags,
		FeatureKeys:        req.FeatureKeys,
		TaskIDs:            req.TaskIDs,
		AcceptanceCriteria: req.AcceptanceCriteria,
		ForbiddenPatterns:  req.ForbiddenPatterns,
		DefinitionOfDone:   req.DefinitionOfDone,
		Dependencies:       req.Dependencies,
		Importance:         defaultDocImportance(req.DocType),
	}
	chunks := buildResolverChunks(req.DocID, sections, req.Chunking)
	tx, err := s.resolverDB().BeginTx(ctx, nil)
	if err != nil {
		return ResolverDocument{}, 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM resolver_document_links WHERE src_doc_id = ?;`, doc.DocID); err != nil {
		return ResolverDocument{}, 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM resolver_chunks WHERE doc_id = ?;`, doc.DocID); err != nil {
		return ResolverDocument{}, 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM resolver_documents WHERE doc_id = ?;`, doc.DocID); err != nil {
		return ResolverDocument{}, 0, err
	}
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO resolver_documents(
	  doc_id, doc_type, title, source_path, version, updated_at,
	  summary, body, tags_json, feature_keys_json, task_ids_json,
	  acceptance_criteria_json, forbidden_patterns_json, definition_of_done_json,
	  dependencies_json, importance
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`, doc.DocID, doc.DocType, doc.Title, doc.SourcePath, doc.Version, doc.UpdatedAt,
		doc.Summary, doc.Body, mustJSON(doc.Tags), mustJSON(doc.FeatureKeys), mustJSON(doc.TaskIDs),
		mustJSON(doc.AcceptanceCriteria), mustJSON(doc.ForbiddenPatterns), mustJSON(doc.DefinitionOfDone),
		mustJSON(doc.Dependencies), doc.Importance); err != nil {
		return ResolverDocument{}, 0, err
	}
	for _, chunk := range chunks {
		if _, err := tx.ExecContext(ctx, `
	INSERT INTO resolver_chunks(
	  chunk_id, doc_id, heading, heading_path_json, ordinal, body, token_estimate, importance
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?);
	`, chunk.ChunkID, chunk.DocID, chunk.Heading, mustJSON(chunk.HeadingPath), chunk.Ordinal, chunk.Body, chunk.TokenEstimate, chunk.Importance); err != nil {
			return ResolverDocument{}, 0, err
		}
	}
	for _, dep := range doc.Dependencies {
		if _, err := tx.ExecContext(ctx, `
	INSERT INTO resolver_document_links(src_doc_id, dst_doc_id, link_type) VALUES(?, ?, 'depends_on');
	`, doc.DocID, dep); err != nil {
			return ResolverDocument{}, 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ResolverDocument{}, 0, err
	}
	return doc, len(chunks), nil
}

func (s *Service) DocsResolve(ctx context.Context, req DocsResolveRequest) ([]ResolveEntry, []ResolveEntry, error) {
	req.Feature = strings.TrimSpace(req.Feature)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.Topic = strings.TrimSpace(req.Topic)
	if req.Feature == "" && req.TaskID == "" && req.Topic == "" {
		return nil, nil, fmt.Errorf("%w: feature/task_id/topic is required", ErrInvalidArgument)
	}
	docs, err := s.listResolverDocuments(ctx)
	if err != nil {
		return nil, nil, err
	}
	scored := scoreResolverDocuments(docs, req.Feature, req.TaskID, req.Topic)
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	required := make([]ResolveEntry, 0, limit)
	recommended := make([]ResolveEntry, 0, limit)
	count := 0
	query := firstNonEmpty(req.Topic, req.Feature, req.TaskID)
	for _, item := range scored {
		if count >= limit {
			break
		}
		entry, err := s.buildResolveEntry(ctx, item.Doc, item.Why, query)
		if err != nil {
			return nil, nil, err
		}
		if item.Doc.Importance == "required" {
			required = append(required, entry)
		} else {
			recommended = append(recommended, entry)
		}
		count++
	}
	return required, recommended, nil
}

func (s *Service) DocsSearch(ctx context.Context, req DocsSearchRequest) ([]ResolveEntry, error) {
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidArgument)
	}
	req.DocTypes = uniqueTrimmedStrings(req.DocTypes)
	req.Tags = uniqueTrimmedStrings(req.Tags)
	req.FeatureKeys = uniqueTrimmedStrings(req.FeatureKeys)
	docs, err := s.listResolverDocuments(ctx)
	if err != nil {
		return nil, err
	}
	docs = filterResolverDocumentsForSearch(docs, req)
	scored := scoreResolverDocuments(docs, "", "", req.Query)
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	results := make([]ResolveEntry, 0, limit)
	for _, item := range scored {
		if len(results) >= limit {
			break
		}
		entry, err := s.buildResolveEntry(ctx, item.Doc, item.Why, req.Query)
		if err != nil {
			return nil, err
		}
		results = append(results, entry)
	}
	return results, nil
}

func (s *Service) ChunksGet(ctx context.Context, req ChunksGetRequest) (string, []ResolverChunk, error) {
	if len(req.ChunkIDs) > 0 {
		chunks, err := s.getResolverChunksByIDs(ctx, req.ChunkIDs)
		if err != nil {
			return "", nil, err
		}
		docID := ""
		if len(chunks) > 0 {
			docID = chunks[0].DocID
		}
		return docID, chunks, nil
	}
	req.DocID = strings.TrimSpace(req.DocID)
	if req.DocID == "" {
		return "", nil, fmt.Errorf("%w: doc_id or chunk_ids is required", ErrInvalidArgument)
	}
	chunks, err := s.getResolverChunksByDocID(ctx, req.DocID)
	if err != nil {
		return "", nil, err
	}
	return req.DocID, filterResolverChunks(chunks, strings.TrimSpace(req.Heading), strings.TrimSpace(req.Query), req.Limit), nil
}

func (s *Service) CardsSearch(ctx context.Context, req CardsSearchRequest) ([]ResolverMemoryCard, error) {
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidArgument)
	}
	req.DocTypes = uniqueTrimmedStrings(req.DocTypes)
	req.Tags = uniqueTrimmedStrings(req.Tags)
	req.FeatureKeys = uniqueTrimmedStrings(req.FeatureKeys)
	req.MemoryTypes = uniqueTrimmedStrings(req.MemoryTypes)
	docs, err := s.listResolverDocuments(ctx)
	if err != nil {
		return nil, err
	}
	docs = filterResolverDocumentsForSearch(docs, DocsSearchRequest{
		Query:       req.Query,
		DocTypes:    req.DocTypes,
		Tags:        req.Tags,
		FeatureKeys: req.FeatureKeys,
	})
	scored := scoreResolverDocuments(docs, "", "", req.Query)
	var chunks []ResolverChunk
	for _, item := range scored {
		docChunks, err := s.getResolverChunksByDocID(ctx, item.Doc.DocID)
		if err != nil {
			return nil, err
		}
		filtered := filterResolverChunks(docChunks, "", req.Query, 0)
		if len(filtered) == 0 {
			filtered = docChunks
		}
		chunks = append(chunks, filtered...)
	}
	feedback, err := s.memoryCardFeedbackWeights(ctx)
	if err != nil {
		return nil, err
	}
	cards := BuildRankedResolverMemoryCardsWithWeights(chunks, req.Query, 0, 0, req.RankingWeights, feedback)
	cards = filterResolverMemoryCards(cards, req.MemoryTypes)
	return applyMemoryCardLimits(cards, req.Limit, req.TokenBudget), nil
}

func (s *Service) ReadsAck(ctx context.Context, req ReadsAckRequest) (ResolverReadReceipt, error) {
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.DocID = strings.TrimSpace(req.DocID)
	req.Version = strings.TrimSpace(req.Version)
	req.Reader = strings.TrimSpace(req.Reader)
	if req.TaskID == "" || req.DocID == "" {
		return ResolverReadReceipt{}, fmt.Errorf("%w: task_id/doc_id is required", ErrInvalidArgument)
	}
	doc, err := s.getResolverDocument(ctx, req.DocID)
	if err != nil {
		return ResolverReadReceipt{}, err
	}
	if req.Version == "" {
		req.Version = doc.Version
	}
	if req.Reader == "" {
		req.Reader = "agent"
	}
	chunkIDs := uniqueTrimmedStrings(req.ChunkIDs)
	snapshots, err := s.buildReadChunkSnapshots(ctx, req.DocID, chunkIDs)
	if err != nil {
		return ResolverReadReceipt{}, err
	}
	if len(chunkIDs) == 0 {
		for _, snapshot := range snapshots {
			chunkIDs = append(chunkIDs, snapshot.ChunkID)
		}
	}
	receipt := ResolverReadReceipt{
		TaskID:         req.TaskID,
		DocID:          req.DocID,
		Version:        req.Version,
		ChunkIDs:       chunkIDs,
		ChunkSnapshots: snapshots,
		Reader:         req.Reader,
		ReadAt:         time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := s.resolverDB().ExecContext(ctx, `
	INSERT INTO resolver_read_receipts(task_id, doc_id, version, chunk_ids_json, chunk_snapshots_json, reader, read_at)
	VALUES(?, ?, ?, ?, ?, ?, ?);
	`, receipt.TaskID, receipt.DocID, receipt.Version, mustJSON(receipt.ChunkIDs), mustJSON(receipt.ChunkSnapshots), receipt.Reader, receipt.ReadAt); err != nil {
		return ResolverReadReceipt{}, err
	}
	return receipt, nil
}

func (s *Service) DocsStaleCheck(ctx context.Context, req DocsStaleCheckRequest) ([]ResolverStaleReason, error) {
	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		return nil, fmt.Errorf("%w: task_id is required", ErrInvalidArgument)
	}
	rows, err := s.resolverDB().QueryContext(ctx, `
	SELECT rr.task_id, rr.doc_id, rr.version, rr.chunk_ids_json, rr.chunk_snapshots_json
	FROM resolver_read_receipts rr
	JOIN (
	  SELECT doc_id, MAX(id) AS latest_id
	  FROM resolver_read_receipts
	  WHERE task_id = ?
	  GROUP BY doc_id
	) latest ON latest.latest_id = rr.id
	WHERE rr.task_id = ?
	ORDER BY rr.id ASC;
	`, req.TaskID, req.TaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stale []ResolverStaleReason
	for rows.Next() {
		var taskID, docID, version, chunkIDsJSON, snapshotsJSON string
		if err := rows.Scan(&taskID, &docID, &version, &chunkIDsJSON, &snapshotsJSON); err != nil {
			return nil, err
		}
		chunkIDs := decodeStringSlice(chunkIDsJSON)
		var snapshots []ResolverChunkSnapshot
		decodeJSON(snapshotsJSON, &snapshots)
		doc, err := s.getResolverDocument(ctx, docID)
		if err != nil {
			if err == ErrNotFound {
				stale = append(stale, ResolverStaleReason{TaskID: taskID, DocID: docID, PreviousVersion: version, Reason: "document_missing", Severity: "critical", ImpactScope: []string{"document"}, DetectedAt: time.Now().UTC().Format(time.RFC3339)})
				continue
			}
			return nil, err
		}
		if doc.Version != version {
			reason, err := s.buildSemanticStaleReason(ctx, taskID, docID, version, doc.Version, chunkIDs, snapshots)
			if err != nil {
				return nil, err
			}
			stale = append(stale, reason)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stale, nil
}

func (s *Service) CardFeedback(ctx context.Context, req CardFeedbackRequest) (CardFeedbackRecord, error) {
	req.CardID = strings.TrimSpace(req.CardID)
	req.DocID = strings.TrimSpace(req.DocID)
	req.ChunkID = strings.TrimSpace(req.ChunkID)
	req.MemoryType = strings.TrimSpace(strings.ToLower(req.MemoryType))
	req.Signal = strings.TrimSpace(strings.ToLower(req.Signal))
	req.Query = strings.TrimSpace(req.Query)
	if req.CardID == "" || req.Signal == "" {
		return CardFeedbackRecord{}, fmt.Errorf("%w: card_id/signal is required", ErrInvalidArgument)
	}
	switch req.Signal {
	case "used", "helpful", "pinned", "irrelevant", "skipped":
	default:
		return CardFeedbackRecord{}, fmt.Errorf("%w: unsupported signal", ErrInvalidArgument)
	}
	if req.Weight <= 0 {
		req.Weight = 1
	}
	record := CardFeedbackRecord{
		CardID:     req.CardID,
		DocID:      req.DocID,
		ChunkID:    req.ChunkID,
		MemoryType: req.MemoryType,
		Signal:     req.Signal,
		Weight:     req.Weight,
		Query:      req.Query,
		RecordedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := s.resolverDB().ExecContext(ctx, `
	INSERT INTO resolver_memory_card_feedback(card_id, doc_id, chunk_id, memory_type, signal, weight, query, recorded_at)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?);
	`, record.CardID, record.DocID, record.ChunkID, record.MemoryType, record.Signal, record.Weight, record.Query, record.RecordedAt); err != nil {
		return CardFeedbackRecord{}, err
	}
	return record, nil
}

func (s *Service) PromptBundle(ctx context.Context, req PromptBundleRequest) (PromptBundleResponse, error) {
	req.Query = strings.TrimSpace(req.Query)
	req.Feature = strings.TrimSpace(req.Feature)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.Format = strings.TrimSpace(strings.ToLower(req.Format))
	if req.Format == "" {
		req.Format = "markdown"
	}
	query := firstNonEmpty(req.Query, req.Feature, req.TaskID)
	if query == "" {
		return PromptBundleResponse{}, fmt.Errorf("%w: query/feature/task_id is required", ErrInvalidArgument)
	}
	cards, err := s.CardsSearch(ctx, CardsSearchRequest{
		Query:       query,
		FeatureKeys: nonEmptySlice(req.Feature),
		MemoryTypes: uniqueTrimmedStrings(req.MemoryTypes),
		Limit:       req.Limit,
		TokenBudget: req.TokenBudget,
	})
	if err != nil {
		return PromptBundleResponse{}, err
	}
	prompt, sourceRefs, tokenEstimate := buildPromptBundleText(query, cards, req.Format)
	return PromptBundleResponse{
		BundleID:      fmt.Sprintf("prompt_bundle_%s", time.Now().UTC().Format("20060102_150405")),
		Query:         query,
		Format:        req.Format,
		TokenEstimate: tokenEstimate,
		Cards:         cards,
		Prompt:        prompt,
		SourceRefs:    sourceRefs,
	}, nil
}

func (s *Service) TaskStateExport(ctx context.Context, req TaskStateExportRequest) (TaskStateExportResponse, error) {
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.Feature = strings.TrimSpace(req.Feature)
	if req.TaskID == "" {
		return TaskStateExportResponse{}, fmt.Errorf("%w: task_id is required", ErrInvalidArgument)
	}
	required, _, err := s.DocsResolve(ctx, DocsResolveRequest{Feature: req.Feature, TaskID: req.TaskID, Limit: 20})
	if err != nil && req.Feature != "" {
		return TaskStateExportResponse{}, err
	}
	receipts, err := s.listLatestReadReceipts(ctx, req.TaskID)
	if err != nil {
		return TaskStateExportResponse{}, err
	}
	stale, err := s.DocsStaleCheck(ctx, DocsStaleCheckRequest{TaskID: req.TaskID})
	if err != nil {
		return TaskStateExportResponse{}, err
	}
	sourceRefs := make([]string, 0, len(required)+len(receipts))
	for _, entry := range required {
		sourceRefs = append(sourceRefs, typedRef("memx", "doc", entry.DocID))
		for _, chunkID := range entry.TopChunks {
			sourceRefs = append(sourceRefs, typedRef("memx", "chunk", chunkID))
		}
	}
	for _, receipt := range receipts {
		sourceRefs = append(sourceRefs, typedRef("memx", "doc", receipt.DocID))
		for _, chunkID := range receipt.ChunkIDs {
			sourceRefs = append(sourceRefs, typedRef("memx", "chunk", chunkID))
		}
	}
	return TaskStateExportResponse{
		TaskRef:      typedRef("agent-taskstate", "task", req.TaskID),
		TaskID:       req.TaskID,
		RequiredDocs: required,
		ReadReceipts: receipts,
		StaleReasons: stale,
		SourceRefs:   uniqueTrimmedStrings(sourceRefs),
		ExportedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) buildReadChunkSnapshots(ctx context.Context, docID string, chunkIDs []string) ([]ResolverChunkSnapshot, error) {
	var chunks []ResolverChunk
	var err error
	if len(chunkIDs) > 0 {
		chunks, err = s.getResolverChunksByIDs(ctx, chunkIDs)
	} else {
		chunks, err = s.getResolverChunksByDocID(ctx, docID)
	}
	if err != nil {
		return nil, err
	}
	snapshots := make([]ResolverChunkSnapshot, 0, len(chunks))
	for _, chunk := range chunks {
		snapshots = append(snapshots, chunkSnapshot(chunk))
	}
	return snapshots, nil
}

func (s *Service) buildSemanticStaleReason(ctx context.Context, taskID string, docID string, previousVersion string, currentVersion string, chunkIDs []string, snapshots []ResolverChunkSnapshot) (ResolverStaleReason, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	reason := ResolverStaleReason{
		TaskID:          taskID,
		DocID:           docID,
		PreviousVersion: previousVersion,
		CurrentVersion:  currentVersion,
		Reason:          "version_mismatch",
		Severity:        "low",
		ImpactScope:     []string{"metadata"},
		DetectedAt:      now,
	}
	if len(snapshots) == 0 {
		return reason, nil
	}
	currentChunks, err := s.getResolverChunksByDocID(ctx, docID)
	if err != nil {
		return ResolverStaleReason{}, err
	}
	currentByID := make(map[string]ResolverChunk, len(currentChunks))
	for _, chunk := range currentChunks {
		currentByID[chunk.ChunkID] = chunk
	}
	readIDs := make(map[string]struct{}, len(chunkIDs))
	for _, chunkID := range chunkIDs {
		readIDs[chunkID] = struct{}{}
	}
	impact := make(map[string]struct{})
	var changes []ResolverChunkChange
	for _, snapshot := range snapshots {
		if len(readIDs) > 0 {
			if _, ok := readIDs[snapshot.ChunkID]; !ok {
				continue
			}
		}
		current, ok := currentByID[snapshot.ChunkID]
		if !ok {
			changes = append(changes, ResolverChunkChange{
				ChunkID:     snapshot.ChunkID,
				ChangeType:  "removed",
				HeadingPath: snapshot.HeadingPath,
				MemoryType:  snapshot.MemoryType,
				Impact:      "read_chunk_removed",
			})
			addImpactScope(impact, snapshot.HeadingPath, snapshot.MemoryType)
			continue
		}
		currentSnapshot := chunkSnapshot(current)
		if currentSnapshot.BodyHash != snapshot.BodyHash {
			changes = append(changes, ResolverChunkChange{
				ChunkID:     snapshot.ChunkID,
				ChangeType:  "modified",
				HeadingPath: currentSnapshot.HeadingPath,
				MemoryType:  currentSnapshot.MemoryType,
				Impact:      "read_chunk_modified",
			})
			addImpactScope(impact, currentSnapshot.HeadingPath, currentSnapshot.MemoryType)
		}
	}
	if len(changes) == 0 {
		return reason, nil
	}
	reason.Reason = "semantic_diff"
	reason.Severity = "high"
	if len(changes) == 1 {
		reason.Severity = "medium"
	}
	reason.ChangedChunks = changes
	reason.ImpactScope = sortedImpactScope(impact)
	return reason, nil
}

func (s *Service) memoryCardFeedbackWeights(ctx context.Context) (map[string]int, error) {
	rows, err := s.resolverDB().QueryContext(ctx, `
	SELECT card_id, memory_type, signal, weight
	FROM resolver_memory_card_feedback
	ORDER BY id DESC
	LIMIT 500;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var cardID, memoryType, signal string
		var weight int
		if err := rows.Scan(&cardID, &memoryType, &signal, &weight); err != nil {
			return nil, err
		}
		delta := weight
		switch signal {
		case "irrelevant", "skipped":
			delta = -weight
		}
		out[cardID] += delta
		if memoryType != "" {
			out["type:"+memoryType] += delta
		}
	}
	return out, rows.Err()
}

func (s *Service) listLatestReadReceipts(ctx context.Context, taskID string) ([]ResolverReadReceipt, error) {
	rows, err := s.resolverDB().QueryContext(ctx, `
	SELECT rr.task_id, rr.doc_id, rr.version, rr.chunk_ids_json, rr.chunk_snapshots_json, rr.reader, rr.read_at
	FROM resolver_read_receipts rr
	JOIN (
	  SELECT doc_id, MAX(id) AS latest_id
	  FROM resolver_read_receipts
	  WHERE task_id = ?
	  GROUP BY doc_id
	) latest ON latest.latest_id = rr.id
	WHERE rr.task_id = ?
	ORDER BY rr.id ASC;
	`, taskID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var receipts []ResolverReadReceipt
	for rows.Next() {
		var receipt ResolverReadReceipt
		var chunkIDsJSON, snapshotsJSON string
		if err := rows.Scan(&receipt.TaskID, &receipt.DocID, &receipt.Version, &chunkIDsJSON, &snapshotsJSON, &receipt.Reader, &receipt.ReadAt); err != nil {
			return nil, err
		}
		receipt.ChunkIDs = decodeStringSlice(chunkIDsJSON)
		decodeJSON(snapshotsJSON, &receipt.ChunkSnapshots)
		receipts = append(receipts, receipt)
	}
	return receipts, rows.Err()
}

func buildPromptBundleText(query string, cards []ResolverMemoryCard, format string) (string, []string, int) {
	var b strings.Builder
	sourceRefs := make([]string, 0, len(cards)*3)
	tokenEstimate := 0
	if format == "jsonl" {
		for _, card := range cards {
			b.WriteString(mustJSON(card))
			b.WriteByte('\n')
			sourceRefs = append(sourceRefs, typedRef("memx", "card", card.CardID), typedRef("memx", "chunk", card.ChunkID), typedRef("memx", "doc", card.DocID))
			tokenEstimate += card.TokenEstimate
		}
		return b.String(), uniqueTrimmedStrings(sourceRefs), tokenEstimate
	}
	b.WriteString("# Memory Cards\n\n")
	b.WriteString("Query: ")
	b.WriteString(query)
	b.WriteString("\n\n")
	for _, card := range cards {
		b.WriteString("- [")
		b.WriteString(card.MemoryType)
		b.WriteString("] ")
		b.WriteString(card.Statement)
		b.WriteString("\n  source: ")
		b.WriteString(card.DocID)
		b.WriteString(" / ")
		b.WriteString(card.ChunkID)
		b.WriteString("\n")
		sourceRefs = append(sourceRefs, typedRef("memx", "card", card.CardID), typedRef("memx", "chunk", card.ChunkID), typedRef("memx", "doc", card.DocID))
		tokenEstimate += card.TokenEstimate
	}
	return b.String(), uniqueTrimmedStrings(sourceRefs), tokenEstimate
}

func chunkSnapshot(chunk ResolverChunk) ResolverChunkSnapshot {
	hydrateResolverChunkMemoryFields(&chunk)
	return ResolverChunkSnapshot{
		ChunkID:       chunk.ChunkID,
		BodyHash:      hashText(chunk.Body),
		HeadingPath:   append([]string(nil), chunk.HeadingPath...),
		MemoryType:    chunk.MemoryType,
		Importance:    chunk.Importance,
		TokenEstimate: chunk.TokenEstimate,
	}
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:])
}

func addImpactScope(scopes map[string]struct{}, headingPath []string, memoryType string) {
	if memoryType != "" {
		scopes["memory_type:"+memoryType] = struct{}{}
	}
	if len(headingPath) > 0 {
		scopes["heading:"+strings.Join(headingPath, " > ")] = struct{}{}
	}
}

func sortedImpactScope(scopes map[string]struct{}) []string {
	out := make([]string, 0, len(scopes))
	for scope := range scopes {
		out = append(out, scope)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return []string{"document"}
	}
	return out
}

func typedRef(domain string, entityType string, id string) string {
	id = strings.TrimSpace(id)
	id = strings.ReplaceAll(id, ":", "_")
	return fmt.Sprintf("%s:%s:local:%s", domain, entityType, id)
}

func (s *Service) ContractsResolve(ctx context.Context, req ContractsResolveRequest) ([]ResolveEntry, []string, []string, []string, []string, error) {
	req.Feature = strings.TrimSpace(req.Feature)
	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.Feature == "" && req.TaskID == "" {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: feature or task_id is required", ErrInvalidArgument)
	}
	required, recommended, err := s.DocsResolve(ctx, DocsResolveRequest{Feature: req.Feature, TaskID: req.TaskID, Limit: 20})
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	all := append(append([]ResolveEntry{}, required...), recommended...)
	var acceptance, forbidden, done, dependencies []string
	for _, entry := range all {
		doc, err := s.getResolverDocument(ctx, entry.DocID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		acceptance = append(acceptance, doc.AcceptanceCriteria...)
		forbidden = append(forbidden, doc.ForbiddenPatterns...)
		done = append(done, doc.DefinitionOfDone...)
		dependencies = append(dependencies, doc.Dependencies...)
	}
	return required, uniqueTrimmedStrings(acceptance), uniqueTrimmedStrings(forbidden), uniqueTrimmedStrings(done), uniqueTrimmedStrings(dependencies), nil
}

// DB access methods

func (s *Service) getResolverDocument(ctx context.Context, docID string) (ResolverDocument, error) {
	var doc ResolverDocument
	var tagsJSON, featureJSON, taskJSON, acceptanceJSON, forbiddenJSON, doneJSON, dependenciesJSON string
	err := s.resolverDB().QueryRowContext(ctx, `
	SELECT doc_id, doc_type, title, source_path, version, updated_at, summary, body,
	       tags_json, feature_keys_json, task_ids_json,
	       acceptance_criteria_json, forbidden_patterns_json, definition_of_done_json,
	       dependencies_json, importance
	FROM resolver_documents
	WHERE doc_id = ?;
	`, docID).Scan(
		&doc.DocID, &doc.DocType, &doc.Title, &doc.SourcePath, &doc.Version, &doc.UpdatedAt, &doc.Summary, &doc.Body,
		&tagsJSON, &featureJSON, &taskJSON, &acceptanceJSON, &forbiddenJSON, &doneJSON, &dependenciesJSON, &doc.Importance,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ResolverDocument{}, ErrNotFound
		}
		return ResolverDocument{}, err
	}
	doc.Tags = decodeStringSlice(tagsJSON)
	doc.FeatureKeys = decodeStringSlice(featureJSON)
	doc.TaskIDs = decodeStringSlice(taskJSON)
	doc.AcceptanceCriteria = decodeStringSlice(acceptanceJSON)
	doc.ForbiddenPatterns = decodeStringSlice(forbiddenJSON)
	doc.DefinitionOfDone = decodeStringSlice(doneJSON)
	doc.Dependencies = decodeStringSlice(dependenciesJSON)
	return doc, nil
}

func (s *Service) listResolverDocuments(ctx context.Context) ([]ResolverDocument, error) {
	rows, err := s.resolverDB().QueryContext(ctx, `
	SELECT doc_id, doc_type, title, source_path, version, updated_at, summary, body,
	       tags_json, feature_keys_json, task_ids_json,
	       acceptance_criteria_json, forbidden_patterns_json, definition_of_done_json,
	       dependencies_json, importance
	FROM resolver_documents;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResolverDocument
	for rows.Next() {
		var doc ResolverDocument
		var tagsJSON, featureJSON, taskJSON, acceptanceJSON, forbiddenJSON, doneJSON, dependenciesJSON string
		if err := rows.Scan(&doc.DocID, &doc.DocType, &doc.Title, &doc.SourcePath, &doc.Version, &doc.UpdatedAt, &doc.Summary, &doc.Body, &tagsJSON, &featureJSON, &taskJSON, &acceptanceJSON, &forbiddenJSON, &doneJSON, &dependenciesJSON, &doc.Importance); err != nil {
			return nil, err
		}
		doc.Tags = decodeStringSlice(tagsJSON)
		doc.FeatureKeys = decodeStringSlice(featureJSON)
		doc.TaskIDs = decodeStringSlice(taskJSON)
		doc.AcceptanceCriteria = decodeStringSlice(acceptanceJSON)
		doc.ForbiddenPatterns = decodeStringSlice(forbiddenJSON)
		doc.DefinitionOfDone = decodeStringSlice(doneJSON)
		doc.Dependencies = decodeStringSlice(dependenciesJSON)
		out = append(out, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) getResolverChunksByDocID(ctx context.Context, docID string) ([]ResolverChunk, error) {
	rows, err := s.resolverDB().QueryContext(ctx, `
	SELECT chunk_id, doc_id, heading, heading_path_json, ordinal, body, token_estimate, importance
	FROM resolver_chunks
	WHERE doc_id = ?
	ORDER BY ordinal ASC;
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResolverChunk
	for rows.Next() {
		var chunk ResolverChunk
		var headingPathJSON string
		if err := rows.Scan(&chunk.ChunkID, &chunk.DocID, &chunk.Heading, &headingPathJSON, &chunk.Ordinal, &chunk.Body, &chunk.TokenEstimate, &chunk.Importance); err != nil {
			return nil, err
		}
		chunk.HeadingPath = decodeStringSlice(headingPathJSON)
		hydrateResolverChunkMemoryFields(&chunk)
		out = append(out, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		if _, err := s.getResolverDocument(ctx, docID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Service) getResolverChunksByIDs(ctx context.Context, chunkIDs []string) ([]ResolverChunk, error) {
	var out []ResolverChunk
	for _, chunkID := range uniqueTrimmedStrings(chunkIDs) {
		var chunk ResolverChunk
		var headingPathJSON string
		err := s.resolverDB().QueryRowContext(ctx, `
	SELECT chunk_id, doc_id, heading, heading_path_json, ordinal, body, token_estimate, importance
	FROM resolver_chunks
	WHERE chunk_id = ?;
	`, chunkID).Scan(&chunk.ChunkID, &chunk.DocID, &chunk.Heading, &headingPathJSON, &chunk.Ordinal, &chunk.Body, &chunk.TokenEstimate, &chunk.Importance)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrNotFound
			}
			return nil, err
		}
		chunk.HeadingPath = decodeStringSlice(headingPathJSON)
		hydrateResolverChunkMemoryFields(&chunk)
		out = append(out, chunk)
	}
	return out, nil
}

func (s *Service) buildResolveEntry(ctx context.Context, doc ResolverDocument, reason string, query string) (ResolveEntry, error) {
	chunks, err := s.getResolverChunksByDocID(ctx, doc.DocID)
	if err != nil {
		return ResolveEntry{}, err
	}
	return ResolveEntry{DocID: doc.DocID, Title: doc.Title, Version: doc.Version, Importance: doc.Importance, Reason: reason, TopChunks: pickTopChunkIDs(chunks, query, 3)}, nil
}

func (s *Service) resolverDB() *sql.DB {
	if s != nil && s.Conn != nil && s.Conn.ResolverDB != nil {
		return s.Conn.ResolverDB
	}
	if s != nil && s.Conn != nil {
		return s.Conn.DB
	}
	return nil
}
