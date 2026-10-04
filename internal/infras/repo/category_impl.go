package repo

import (
	"context"
	"core-server/internal/infras/repo/ent"
	"core-server/internal/infras/repo/ent/category"

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
		if ent.IsNotFound(err) {
			return ErrCategoryNotFound
		}
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
		return ErrCategoryNotFound
	}
	return err
}

func (r *CategoryRepo) DeleteType(ctx context.Context, id uint64) error {
	err := r.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := r.DB(ctx).Category.Query().
			Where(category.IDEQ(id), category.ParentIDEQ(entity.RootCategoryParentID)).
			Only(ctx); err != nil {
			if ent.IsNotFound(err) {
				return ErrCategoryNotFound
			}
			return err
		}

		// 删除二级分类
		_, err := r.DB(ctx).Category.Delete().
			Where(category.ParentIDEQ(id)).
			Exec(ctx)
		if err != nil {
			return err
		}

		// 删除该父类评论
		_, err = r.DB(ctx).Category.Delete().
			Where(category.IDEQ(id), category.ParentIDEQ(entity.RootCategoryParentID)).
			Exec(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *CategoryRepo) ListType(ctx context.Context) ([]*entity.Category, error) {
	nodes, err := r.DB(ctx).Category.Query().
		Where(category.ParentIDEQ(entity.RootCategoryParentID)).
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
