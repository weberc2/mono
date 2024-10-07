package mm

import (
	"context"
	"fmt"
	"slices"
	"sync"
)

type MemoryImportStore struct {
	lock    sync.RWMutex
	imports []Import
}

var _ ImportStore = (*MemoryImportStore)(nil)

func (store *MemoryImportStore) ListImports(
	ctx context.Context,
) (imports []Import, err error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	imports = make([]Import, len(store.imports))
	copy(imports, store.imports)

	for i := range imports {
		copy(imports[i].Files, store.imports[i].Files)
	}

	return
}

func (store *MemoryImportStore) CreateImport(
	ctx context.Context,
	imp *Import,
) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.imports {
		if store.imports[i].ID == imp.ID {
			return fmt.Errorf(
				"creating import: %w",
				&ImportExistsErr{Import: imp.ID},
			)
		}
	}

	store.imports = append(store.imports, *imp)
	return nil
}

func (store *MemoryImportStore) UpdateImport(
	ctx context.Context,
	imp *Import,
) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.imports {
		if store.imports[i].ID == imp.ID {
			store.imports[i] = *imp
			return nil
		}
	}

	return &ImportNotFoundErr{Import: imp.ID}
}

func (store *MemoryImportStore) DeleteImport(
	ctx context.Context,
	imp ImportID,
) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.imports {
		if store.imports[i].ID == imp {
			slices.Delete(store.imports, i, i)
			return nil
		}
	}

	return &ImportNotFoundErr{Import: imp}
}
