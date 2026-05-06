package service

import (
	"context"
	"database/sql"
	"fmt"
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
	required, recommended, err := s.DocsResolve(ctx, DocsResolveRequest{Topic: req.Query, Limit: req.Limit})
	if err != nil {
		return nil, err
	}
	return append(required, recommended...), nil
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
	receipt := ResolverReadReceipt{
		TaskID:   req.TaskID,
		DocID:    req.DocID,
		Version:  req.Version,
		ChunkIDs: uniqueTrimmedStrings(req.ChunkIDs),
		Reader:   req.Reader,
		ReadAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := s.resolverDB().ExecContext(ctx, `
	INSERT INTO resolver_read_receipts(task_id, doc_id, version, chunk_ids_json, reader, read_at)
	VALUES(?, ?, ?, ?, ?, ?);
	`, receipt.TaskID, receipt.DocID, receipt.Version, mustJSON(receipt.ChunkIDs), receipt.Reader, receipt.ReadAt); err != nil {
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
	SELECT rr.task_id, rr.doc_id, rr.version
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
		var taskID, docID, version string
		if err := rows.Scan(&taskID, &docID, &version); err != nil {
			return nil, err
		}
		doc, err := s.getResolverDocument(ctx, docID)
		if err != nil {
			if err == ErrNotFound {
				stale = append(stale, ResolverStaleReason{TaskID: taskID, DocID: docID, PreviousVersion: version, Reason: "document_missing", DetectedAt: time.Now().UTC().Format(time.RFC3339)})
				continue
			}
			return nil, err
		}
		if doc.Version != version {
			stale = append(stale, ResolverStaleReason{TaskID: taskID, DocID: docID, PreviousVersion: version, CurrentVersion: doc.Version, Reason: "version_mismatch", DetectedAt: time.Now().UTC().Format(time.RFC3339)})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stale, nil
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