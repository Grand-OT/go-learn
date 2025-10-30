package todo

import "context"

type Repository interface {
	Create(ctx context.Context, t Todo) (Todo, error)
	Get(ctx context.Context, id int64) (Todo, error)
	Remove(ctx context.Context, id int64) error
	Ping(ctx context.Context) error
}
