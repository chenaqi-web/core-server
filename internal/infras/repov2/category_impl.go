package repov2

import (
	"context"
	"time"

	sqlrepo "core-server/internal/infras/repo"
	"core-server/internal/infras/repov2/ent"
	"core-server/internal/infras/repov2/ent/category"
	"core-server/internal/model/entity"
)

type CategoryRepo struct {
	*EntClient
}

func NewCategoryRepo(client *EntClient) *CategoryRepo {
	return &CategoryRepo{
		EntClient: client,
	}
}

func (r *CategoryRepo) Create(ctx context.Context, value *entity.Category) error {
	node, err := r.DB(ctx).Category.Create().
		SetParentID(value.ParentID).
		SetName(value.Name).
		Save(ctx)
	if err != nil {
		return err
	}
	value.ID = node.ID
	value.CreatedAt = node.CreatedAt
	value.UpdatedAt = node.UpdatedAt
	return nil
}

func (r *CategoryRepo) DeleteByID(ctx context.Context, id uint64) error {
	_, err := r.DB(ctx).Category.UpdateOneID(id).
		Where(category.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if ent.IsNotFound(err) {
		return sqlrepo.ErrNotFound
	}
	if err != nil {
		return err
	}
	return err
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uint64) (*entity.Category, error) {
	node, err := r.DB(ctx).Category.Query().
		Where(category.IDEQ(id), category.DeletedAtIsNil()).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toEntityCategory(node), nil
}

func (r *CategoryRepo) ListByParentID(ctx context.Context, parentID uint64) ([]*entity.Category, error) {
	nodes, err := r.DB(ctx).Category.Query().
		Where(category.ParentIDEQ(parentID), category.DeletedAtIsNil()).
		Order(ent.Asc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCategories(nodes), nil
}

func (r *CategoryRepo) DeleteByParentID(ctx context.Context, parentID uint64) error {
	_, err := r.DB(ctx).Category.Update().
		Where(category.ParentIDEQ(parentID), category.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return err
}
