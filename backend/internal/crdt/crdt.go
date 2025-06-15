package crdt

import (
	"sync"

	"github.com/automerge/automerge-go"
)

var store = struct {
	sync.RWMutex
	docs map[string]*automerge.Doc
}{docs: make(map[string]*automerge.Doc)}

func LoadDoc(id string) (*automerge.Doc, error) {
	store.RLock()
	doc, ok := store.docs[id]
	store.RUnlock()
	if ok {
		return doc, nil
	}
	newDoc := automerge.New()
	store.Lock()
	store.docs[id] = newDoc
	store.Unlock()
	return newDoc, nil
}

func ApplyChanges(id string, changes []byte) (*automerge.Doc, error) {
	doc, err := LoadDoc(id)
	if err != nil {
		return nil, err
	}
	chgs, err := automerge.LoadChanges(changes)
	if err != nil {
		return nil, err
	}
	if err := doc.Apply(chgs...); err != nil {
		return nil, err
	}
	return doc, nil
}

func SaveSnapshot(id string) ([]byte, error) {
	doc, err := LoadDoc(id)
	if err != nil {
		return nil, err
	}
	return doc.Save(), nil
}
