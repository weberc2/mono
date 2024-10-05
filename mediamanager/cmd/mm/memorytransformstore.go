package main

import (
	"fmt"
	"sync"
)

type MemoryTransformStore struct {
	transforms []Transform
	lock       sync.RWMutex
}

var _ TransformStore = (*MemoryTransformStore)(nil)

func (store *MemoryTransformStore) CreateTransform(
	ctx Context,
	t *Transform,
) (transform Transform, err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.transforms {
		if store.transforms[i].ID == t.ID {
			err = fmt.Errorf("transform exists: %s", t.ID)
			return
		}
	}

	transform = *t
	if t.Spec.Type == TransformTypeFilm {
		transform.Files = make([]TransformFile, len(t.Spec.Film.Files))
		for i := range t.Spec.Film.Files {
			transform.Files[i].Path = t.Spec.Film.Files[i].Path
			transform.Files[i].Status = TransformFileStatusPending
		}
	}
	store.transforms = append(store.transforms, transform)
	return
}

func (store *MemoryTransformStore) ListTransforms(
	Context,
) (transforms []Transform, err error) {
	store.lock.RLock()
	transforms = store.transforms
	store.lock.RUnlock()
	return
}

func (store *MemoryTransformStore) UpdateTransformFile(
	ctx Context,
	id TransformID,
	tf *TransformFile,
) (file TransformFile, err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.transforms {
		if store.transforms[i].ID == id {
			for j := range store.transforms[i].Files {
				if store.transforms[i].Files[j].Path == tf.Path {
					store.transforms[i].Files[j].Status = tf.Status
					file = store.transforms[i].Files[j]
					return
				}
			}
			err = fmt.Errorf(
				"updating file in transform `%s`: file not found: %w",
				id,
				tf.Path,
			)
			return
		}
	}
	err = fmt.Errorf("transform not found: %s", id)
	return
}
