package repov2

import (
	"context"
	"core-server/internal/infras/repov2/ent"
	"core-server/internal/infras/repov2/ent/user"
	"core-server/internal/model/entity"
	"database/sql"
	"errors"
	"math"
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
		Where(user.IDEQ(int64(id)), user.DeletedAtIsNil()).
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
		return nil, err
	}
	return toEntityUser(node), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	node, err := r.db.User.Query().
		Where(user.EmailEQ(email), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityUser(node), nil
}

func (r *UserRepo) Create(ctx context.Context, value *entity.User) error {
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
	value.ID = uint64(node.ID)
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

	entIDs := make([]int64, len(ids))
	for i, id := range ids {
		if id > math.MaxInt64 {
			return nil, errors.New("user id exceeds int64 range")
		}
		entIDs[i] = int64(id)
	}
	nodes, err := r.db.User.Query().
		Where(user.IDIn(entIDs...), user.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityUsers(nodes), nil
}

func (r *UserRepo) GetLikeCount(ctx context.Context, userID uint64) (int64, error) {
	node, err := r.db.User.Query().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return 0, err
	}
	return int64(node.LikeCount), nil
}

func (r *UserRepo) GetReceiveLikeCount(ctx context.Context, userID uint64) (int64, error) {
	node, err := r.db.User.Query().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return 0, err
	}
	return int64(node.ReceiveLikeCount), nil
}

func (r *UserRepo) IncrementLikeCount(ctx context.Context, userID uint64) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil()).
		AddLikeCount(1).
		Save(ctx)
	return err
}

func (r *UserRepo) DecrementLikeCount(ctx context.Context, userID uint64) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil(), user.LikeCountGT(0)).
		AddLikeCount(-1).
		Save(ctx)
	return err
}

func (r *UserRepo) SetReceiveLikeCount(ctx context.Context, userID uint64, count int64) error {
	if count < 0 {
		return errors.New("receive like count cannot be negative")
	}
	_, err := r.db.User.Update().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil()).
		SetReceiveLikeCount(uint64(count)).
		Save(ctx)
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, value *entity.User) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(int64(value.ID)), user.DeletedAtIsNil()).
		SetName(value.Name).
		SetPhone(value.Phone).
		SetSex(value.Sex).
		SetAge(value.Age).
		Save(ctx)
	return err
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, userID uint64, avatar string) error {
	_, err := r.db.User.Update().
		Where(user.IDEQ(int64(userID)), user.DeletedAtIsNil()).
		SetAvatar(avatar).
		Save(ctx)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uint64, password string) error {
	err := r.db.User.UpdateOneID(int64(userID)).SetPassword(password).Exec(ctx)
	if ent.IsNotFound(err) {
		return errors.New("user not found")
	}
	return err
}

func (r *UserRepo) UpdateStatus(ctx context.Context, userID uint64, status string) error {
	err := r.db.User.UpdateOneID(int64(userID)).
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
		ID:               uint64(node.ID),
		CreatedAt:        node.CreatedAt,
		UpdatedAt:        node.UpdatedAt,
		Name:             node.Name,
		Password:         node.Password,
		Phone:            node.Phone,
		Avatar:           node.Avatar,
		Email:            node.Email,
		Role:             node.Role,
		Sex:              node.Sex,
		Age:              node.Age,
		LikeCount:        node.LikeCount,
		ReceiveLikeCount: node.ReceiveLikeCount,
		Status:           node.Status,
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
