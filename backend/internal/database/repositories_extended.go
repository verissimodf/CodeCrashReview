package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChallengeTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewChallengeTemplateRepository(pool *pgxpool.Pool) *ChallengeTemplateRepository {
	return &ChallengeTemplateRepository{pool: pool}
}

func (r *ChallengeTemplateRepository) Create(ctx context.Context, template *ChallengeTemplate) error {
	acceptanceCriteriaJSON, _ := json.Marshal(template.AcceptanceCriteriaTemplate)
	scoringWeightsJSON, _ := json.Marshal(template.ScoringWeightsTemplate)
	knownIssuesJSON, _ := json.Marshal(template.KnownIssuesTemplate)

	return r.pool.QueryRow(ctx, `
		INSERT INTO challenge_templates (name, description, category, default_severity, default_time_limit_seconds, default_max_turns, acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`,
		template.Name, template.Description, template.Category, template.DefaultSeverity,
		template.DefaultTimeLimitSeconds, template.DefaultMaxTurns,
		acceptanceCriteriaJSON, scoringWeightsJSON, knownIssuesJSON, template.ExampleRepository,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
}

func (r *ChallengeTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*ChallengeTemplate, error) {
	template := &ChallengeTemplate{}
	var acceptanceCriteria, scoringWeights, knownIssues []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, category, default_severity, default_time_limit_seconds, default_max_turns,
		       acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository,
		       created_at, updated_at
		FROM challenge_templates WHERE id = $1
	`, id).Scan(
		&template.ID, &template.Name, &template.Description, &template.Category,
		&template.DefaultSeverity, &template.DefaultTimeLimitSeconds, &template.DefaultMaxTurns,
		&acceptanceCriteria, &scoringWeights, &knownIssues, &template.ExampleRepository,
		&template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(acceptanceCriteria, &template.AcceptanceCriteriaTemplate)
	json.Unmarshal(scoringWeights, &template.ScoringWeightsTemplate)
	json.Unmarshal(knownIssues, &template.KnownIssuesTemplate)
	return template, nil
}

func (r *ChallengeTemplateRepository) List(ctx context.Context, category string) ([]*ChallengeTemplate, error) {
	query := `SELECT id, name, description, category, default_severity, default_time_limit_seconds, default_max_turns,
	           acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository,
	           created_at, updated_at
	        FROM challenge_templates`
	args := []interface{}{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*ChallengeTemplate
	for rows.Next() {
		t := &ChallengeTemplate{}
		var acceptanceCriteria, scoringWeights, knownIssues []byte
		var exampleRepo *string

		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.Category,
			&t.DefaultSeverity, &t.DefaultTimeLimitSeconds, &t.DefaultMaxTurns,
			&acceptanceCriteria, &scoringWeights, &knownIssues, &exampleRepo,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.ExampleRepository = exampleRepo
		json.Unmarshal(acceptanceCriteria, &t.AcceptanceCriteriaTemplate)
		json.Unmarshal(scoringWeights, &t.ScoringWeightsTemplate)
		json.Unmarshal(knownIssues, &t.KnownIssuesTemplate)
		templates = append(templates, t)
	}
	return templates, nil
}

type SnippetCollectionRepository struct {
	pool *pgxpool.Pool
}

func NewSnippetCollectionRepository(pool *pgxpool.Pool) *SnippetCollectionRepository {
	return &SnippetCollectionRepository{pool: pool}
}

func (r *SnippetCollectionRepository) Create(ctx context.Context, collection *SnippetCollection) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO snippet_collections (name, description, snippet_ids, tags, is_curated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`,
		collection.Name, collection.Description, collection.SnippetIDs, collection.Tags, collection.IsCurated,
	).Scan(&collection.ID, &collection.CreatedAt)
}

func (r *SnippetCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*SnippetCollection, error) {
	collection := &SnippetCollection{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, snippet_ids, tags, is_curated, created_at
		FROM snippet_collections WHERE id = $1
	`, id).Scan(
		&collection.ID, &collection.Name, &collection.Description,
		&collection.SnippetIDs, &collection.Tags, &collection.IsCurated, &collection.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (r *SnippetCollectionRepository) List(ctx context.Context, tags []string) ([]*SnippetCollection, error) {
	query := `SELECT id, name, description, snippet_ids, tags, is_curated, created_at FROM snippet_collections`
	args := []interface{}{}
	if len(tags) > 0 {
		query += " WHERE $1 && tags"
		args = append(args, tags)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []*SnippetCollection
	for rows.Next() {
		c := &SnippetCollection{}
		var snippetIDs, tags pgtype.TextArray

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description,
			&snippetIDs, &tags, &c.IsCurated, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.SnippetIDs = snippetIDs.Elements
		c.Tags = tags.Elements
		collections = append(collections, c)
	}
	return collections, nil
}

func (r *SnippetCollectionRepository) AddSnippet(ctx context.Context, collectionID uuid.UUID, snippetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE snippet_collections SET snippet_ids = array_append(snippet_ids, $2) WHERE id = $1
	`, collectionID, snippetID)
	return err
}

func (r *SnippetCollectionRepository) RemoveSnippet(ctx context.Context, collectionID uuid.UUID, snippetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE snippet_collections SET snippet_ids = array_remove(snippet_ids, $2) WHERE id = $1
	`, collectionID, snippetID)
	return err
}