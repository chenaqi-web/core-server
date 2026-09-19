package repov2

import (
	"context"
	"core-server/internal/infras/repov2/ent"
	"core-server/internal/infras/repov2/ent/user"
	"core-server/internal/infras/repov2/ent/userstat"
	"core-server/internal/model/entity"
	"database/sql"
	"errors"
)

type UserRepo struct {
	*EntClient
}

func NewUserRepo(client *EntClient) *UserRepo {
	return &UserRepo{
		EntClient: client,
	}
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	node, err := r.db.User.Query().
		Where(user.IDEQ(id), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityUser(node), nil
}

func (r *UserRepo) GetByName(ctx context.Context, name string) (*entity.User, error) {
	node, err := r.db.User.Query().
		Where(user.NameEQ(name), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toEntityUser(node), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	node, err := r.db.User.Query().
		Where(user.EmailEQ(email), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toEntityUser(node), nil
}

func (r *UserRepo) CreateUser(ctx context.Context, value *entity.User) error {
	err := r.WithTransaction(ctx, func(ctx context.Context) error {
		node, err := r.db.User.Create().
			SetName(value.Name).
			SetPassword(value.Password).
			SetEmail(value.Email).
			SetRole(value.Role).
			SetStatus(value.Status).
			Save(ctx)
		if err != nil {
			return err
		}

		_, err = r.db.UserStat.Create().
			SetUserID(node.ID).
			SetArticleCount(0).
			SetFollowersCount(0).
			SetFollowingCount(0).
			SetLikeCount(0).
			SetReceiveLikeCount(0).
			SetFavorCount(0).
			SetReceiveFavorCount(0).
			Save(ctx)
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

func (r *UserRepo) List(ctx context.Context, limit, offset uint32) ([]*entity.User, uint64, error) {
	query := r.db.User.Query().Where(user.DeletedAtIsNil())
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	nodes, err := query.Order(ent.Desc(user.FieldID)).Limit(int(limit)).Offset(int(offset)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return toEntityUsers(nodes), uint64(total), nil
}

func (r *UserRepo) Search(ctx context.Context, keyword string, limit, offset uint32) ([]*entity.User, uint64, error) {
	query := r.db.User.Query().Where(
		user.DeletedAtIsNil(),
		user.Or(user.NameContains(keyword), user.EmailContains(keyword)),
	)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	nodes, err := query.Order(ent.Desc(user.FieldID)).Limit(int(limit)).Offset(int(offset)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return toEntityUsers(nodes), uint64(total), nil
}

func (r *UserRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	entIDs := ids
	nodes, err := r.db.User.Query().
		Where(user.IDIn(entIDs...), user.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityUsers(nodes), nil
}

func (r *UserRepo) GetStat(ctx context.Context, userID uint64) (*entity.UserStat, error) {
	stat, err := r.db.UserStat.Query().Where(userstat.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityUserStat(stat), nil
}

func (r *UserRepo) GetLikeCount(ctx context.Context, userID uint64) (int64, error) {
	return 0, nil
}

func (r *UserRepo) GetReceiveLikeCount(ctx context.Context, userID uint64) (int64, error) {
	return 0, nil
}

func (r *UserRepo) IncrementLikeCount(ctx context.Context, userID uint64) error {
	return nil
}

func (r *UserRepo) DecrementLikeCount(ctx context.Context, userID uint64) error {
	return nil
}

func (r *UserRepo) SetReceiveLikeCount(ctx context.Context, userID uint64, count int64) error {
	return nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, value *entity.User) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(value.ID), user.DeletedAtIsNil()).
		SetName(value.Name).
		SetPhone(value.Phone).
		SetSex(value.Sex).
		//SetAge(value.Age).
		Save(ctx)
	return err
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, userID uint64, avatar string) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(userID), user.DeletedAtIsNil()).
		SetAvatar(avatar).
		Save(ctx)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uint64, password string) error {
	err := r.db.User.UpdateOneID(userID).SetPassword(password).Exec(ctx)
	if ent.IsNotFound(err) {
		return errors.New("user not found")
	}
	return err
}

func (r *UserRepo) UpdateStatus(ctx context.Context, userID uint64, status string) error {
	err := r.db.User.UpdateOneID(userID).
		Where(user.DeletedAtIsNil()).
		SetStatus(status).
		Exec(ctx)
	return err
}

// =====================================================================================================================

func toEntityUser(node *ent.User) *entity.User {
	if node == nil {
		return nil
	}
	value := &entity.User{
		ID:        node.ID,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
		Name:      node.Name,
		Password:  node.Password,
		Phone:     node.Phone,
		Avatar:    node.Avatar,
		Email:     node.Email,
		Role:      node.Role,
		Sex:       node.Sex,
		//Age:       node.Age,
		Status: node.Status,
	}
	if node.DeletedAt != nil {
		value.DeletedAt = sql.NullTime{Time: *node.DeletedAt, Valid: true}
	}
	return value
}

func toEntityUsers(nodes []*ent.User) []*entity.User {
	users := make([]*entity.User, 0, len(nodes))
	for _, node := range nodes {
		users = append(users, toEntityUser(node))
	}
	return users
}

func toEntityUserStat(stat *ent.UserStat) *entity.UserStat {
	if stat == nil {
		return nil
	}
	return &entity.UserStat{
		UserID:            stat.UserID,
		ArticleCount:      stat.ArticleCount,
		FollowersCount:    stat.FollowersCount,
		FollowingCount:    stat.FollowingCount,
		LikeCount:         stat.LikeCount,
		ReceiveLikeCount:  stat.ReceiveLikeCount,
		FavorCount:        stat.FavorCount,
		ReceiveFavorCount: stat.ReceiveFavorCount,
	}
}
