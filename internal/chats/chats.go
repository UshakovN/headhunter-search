package chats

import (
	"context"
	"main/pkg/tree"
	"sync"
)

type Link string

type State struct {
	MessageID int64
	Event     func(input *EventInput) (messageID int64, err error)
}

type EventInput struct {
	Ctx     context.Context
	ChatID  int64
	UserID  int64
	Text    string
	Command string
}

type Trees struct {
	mtx sync.Mutex
	m   map[int64]tree.Tree[Link, *State]
	t   func() tree.Tree[Link, *State]
}

func NewChatsTrees(t func() tree.Tree[Link, *State]) *Trees {
	return &Trees{
		m: map[int64]tree.Tree[Link, *State]{},
		t: t,
	}
}

func (trs *Trees) NewTree(chatID int64) tree.Tree[Link, *State] {
	trs.mtx.Lock()
	defer trs.mtx.Unlock()

	trs.m[chatID] = trs.t()
	return trs.m[chatID]
}

func (trs *Trees) Tree(chatID int64) tree.Tree[Link, *State] {
	trs.mtx.Lock()
	defer trs.mtx.Unlock()

	if t, ok := trs.m[chatID]; ok {
		return t
	}
	trs.m[chatID] = trs.t()

	return trs.m[chatID]
}

func (trs *Trees) SetTree(chatID int64, t tree.Tree[Link, *State]) {
	trs.m[chatID] = t
}
