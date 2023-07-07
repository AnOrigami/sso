package util

import (
	"context"
	"errors"
)

// Ticket的获取和储存操作
var (
	ErrTicketNotExists = errors.New("ticket not exists")
)

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type Store interface {
	GetTicket(ctx context.Context, ticket string) (UserInfo, error)
	SetTicket(ctx context.Context, ticket string, info UserInfo) error
}

type memStore struct {
	ticketMap map[string]UserInfo
}

func NewMemStore() Store {
	return &memStore{
		ticketMap: map[string]UserInfo{},
	}
}

func (store *memStore) SetTicket(_ context.Context, ticket string, info UserInfo) error {
	store.ticketMap[ticket] = info
	return nil
}

func (store *memStore) GetTicket(_ context.Context, ticket string) (UserInfo, error) {
	info, exists := store.ticketMap[ticket]
	if exists {
		//获取ticket之后，删除ticket，防止内存中ticket过多
		delete(store.ticketMap, ticket)
		return info, nil
	}
	return UserInfo{}, ErrTicketNotExists
}
