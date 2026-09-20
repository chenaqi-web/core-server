package repov2

import (
	"context"

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

func (r *CategoryRepo) CreateType(ctx context.Context, value *entity.Category) error {
	_, err := r.DB(ctx).Category.Create().
		SetParentID(entity.RootCategoryParentID).
		SetName(value.Name).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *CategoryRepo) CreateCate(ctx context.Context, value *entity.Category) error {
	if _, err := r.DB(ctx).Category.Query().
		Where(category.IDEQ(value.ParentID), category.ParentIDEQ(entity.RootCategoryParentID)).
		Only(ctx); err != nil {
		return err
	}

	_, err := r.DB(ctx).Category.Create().
		SetParentID(value.ParentID).
		SetName(value.Name).
		Save(ctx)
	return err
}

func (r *CategoryRepo) DeleteCate(ctx context.Context, id uint64) error {
	err := r.DB(ctx).Category.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return sqlrepo.ErrNotFound
	}
	return err
}

func (r *CategoryRepo) DeleteType(ctx context.Context, id uint64) error {
	err := r.WithTransaction(ctx, func(ctx context.Context) error {
		deleteBuilder := r.DB(ctx).Category.Delete().Where(category.IDEQ(id))
		deleted, err := deleteBuilder.Exec(ctx)
		if err != nil {
			return err
		}
		if deleted == 0 {
			return sqlrepo.ErrNotFound
		}

		_, err = r.DB(ctx).Category.Delete().
			Where(category.ParentIDEQ(id)).
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
		Where(category.ParentIDEQ(0)).
		Order(ent.Asc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCategories(nodes), nil
}

func (r *CategoryRepo) ListCate(ctx context.Context, parentID uint64) ([]*entity.Category, error) {
	nodes, err := r.DB(ctx).Category.Query().
		Where(category.ParentIDEQ(parentID)).
		Order(ent.Asc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCategories(nodes), nil
}
