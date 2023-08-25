package dialog

import (
	"main/internal/storage"
	"main/pkg/telegram"
	"main/pkg/tree"
	"sync"
)

type (
	Link   string
	Action func() error
)

const (
	LinkStart         Link = "start"
	LinkSub           Link = "sub"
	LinkUnsub         Link = "unsub"
	LinkContacts      Link = "contacts"
	LinkMan           Link = "man"
	LinkSubArea       Link = "area"
	LinkSubExperience Link = "experience"
	LinkSubKeywords   Link = "keywords"
	LinkUnsubSub      Link = "sub"
	LinkBack          Link = "back"
)

type Tree struct {
	storage storage.Storage
	tree    tree.Tree[Link, Action]
}

type ChatsTrees struct {
	mtx sync.Mutex
	m   map[int64]*Tree
}

func NewChatsTrees() *ChatsTrees {
	return &ChatsTrees{
		m: map[int64]*Tree{},
	}
}

func (d *ChatsTrees) Dialog(chatID int64) *Tree {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if dialog, ok := d.m[chatID]; ok {
		return dialog
	}
	d.m[chatID] = NewDialog(chatID)

	return d.m[chatID]
}

func NewDialog(chatID int64) *Tree {
	head := tree.NewTree[Link, Action]()

	head.Push(LinkStart, func() error {

	})

	start := head.Next(LinkStart)

	start.Push(LinkSub, newSubMessage(chatID))
	start.Push(LinkUnsub, newUnsubMessage(chatID, nil))
	start.Push(LinkContacts, newContactsMessage(chatID))
	start.Push(LinkMan, newManMessage(chatID))

	sub := start.Next(LinkSub)
	sub.Push(LinkSubArea, newSubAreaMessage(chatID))
	sub.Push(LinkSubExperience, newSubExperienceMessage(chatID))
	sub.Push(LinkSubKeywords, newSubKeywordsMessage(chatID))

	unsub := start.Next(LinkUnsub)
	unsub.Push(LinkUnsubSub, newUnsubSubMessage(chatID, ""))

	return &Tree{
		tree: head,
	}
}

func (d *Tree) Process(chatID int64, link Link) *telegram.SendMessage {
	if link == LinkBack {
		if entity := d.tree.Prev().Entity(); entity == nil {
			return entity
		}
	}
	d.tree = d.tree.Next(link)

	if entity := d.tree.Entity(); entity != nil {
		return entity
	}
	d.tree.Prev()

	return newUndefinedMessage(chatID)
}
