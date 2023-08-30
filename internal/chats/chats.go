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

type Trees interface {
	Tree(chatID int64) tree.Tree[Link, *State]
	SetTree(chatID int64, t tree.Tree[Link, *State])
	RebuildTree(chatID int64) tree.Tree[Link, *State]
}

type trees struct {
	mtx       sync.Mutex
	chatTrees map[int64]tree.Tree[Link, *State]
	buildTree func() tree.Tree[Link, *State]
}

func NewChatsTrees(buildTree func() tree.Tree[Link, *State]) Trees {
	return &trees{
		chatTrees: map[int64]tree.Tree[Link, *State]{},
		buildTree: buildTree,
	}
}

func (trs *trees) Tree(chatID int64) tree.Tree[Link, *State] {
	trs.mtx.Lock()
	defer trs.mtx.Unlock()

	if t, ok := trs.chatTrees[chatID]; ok {
		return t
	}
	trs.chatTrees[chatID] = trs.buildTree()

	return trs.chatTrees[chatID]
}

func (trs *trees) SetTree(chatID int64, t tree.Tree[Link, *State]) {
	trs.chatTrees[chatID] = t
}

func (trs *trees) RebuildTree(chatID int64) tree.Tree[Link, *State] {
	trs.mtx.Lock()
	defer trs.mtx.Unlock()

	trs.chatTrees[chatID] = trs.buildTree()
	return trs.chatTrees[chatID]
}
