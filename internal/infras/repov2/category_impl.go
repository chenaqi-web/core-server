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
	_, err := r.DB(ctx).Category.Create().
		SetParentID(value.ParentID).
		SetName(value.Name).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *CategoryRepo) DeleteCate(ctx context.Context, id uint64) error {
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

func (r *CategoryRepo) DeleteType(ctx context.Context, id uint64) error {
	err := r.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := r.DB(ctx).Category.UpdateOneID(id).
			Where(category.DeletedAtIsNil()).
			SetDeletedAt(time.Now()).
			Save(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return sqlrepo.ErrNotFound
			}
			return err
		}

		err = r.DB(ctx).Category.Update().
			Where(
				category.ParentID(id),
				category.DeletedAtIsNil(),
			).
			SetDeletedAt(time.Now()).
			Exec(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return err
}

func (r *CategoryRepo) ListType(ctx context.Context) ([]*entity.Category, error) {
	nodes, err := r.DB(ctx).Category.Query().
		Where(category.ParentIDEQ(0), category.DeletedAtIsNil()).
		Order(ent.Asc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCategories(nodes), nil
}

func (r *CategoryRepo) ListCate(ctx context.Context, parentID uint64) ([]*entity.Category, error) {
	nodes, err := r.DB(ctx).Category.Query().
		Where(category.ParentIDEQ(parentID), category.DeletedAtIsNil()).
		Order(ent.Asc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCategories(nodes), nil
}
