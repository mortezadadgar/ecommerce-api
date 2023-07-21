package postgres_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/postgres"
)

func TestCategoryService_Create(t *testing.T) {
	db := newTestDB(t, "category_create")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	want := domain.Category{Name: "name"}
	err := postgres.NewCategoryStore(db).Create(ctx, &want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := postgres.NewCategoryStore(db).GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Name != want.Name {
		t.Errorf("expected name of %s, want: %s ", got.Name, want.Name)
	}

	err = postgres.NewCategoryStore(db).Create(ctx, &want)
	if err != domain.ErrDuplicatedCategory {
		t.Errorf("expected %q from Create, got: %q", domain.ErrDuplicatedCategory, err)
	}
}

func TestCategoryService_List(t *testing.T) {
	db := newTestDB(t, "category_list")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := postgres.NewCategoryStore(db).List(ctx, domain.CategoryFilter{})
	if err != domain.ErrNoCategoryFound {
		t.Errorf("expected %q from List, got: %q", domain.ErrNoCategoryFound, err)
	}

	const n = 3
	for i := 0; i < n; i++ {
		category := domain.Category{Name: "name" + strconv.Itoa(i)}
		err = postgres.NewCategoryStore(db).Create(ctx, &category)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	categories, err := postgres.NewCategoryStore(db).List(ctx, domain.CategoryFilter{})
	if err != nil {
		t.Errorf("List: %v", err)
	}

	if len(categories) != n {
		t.Errorf("expected length of %d, got: %d", n, len(categories))
	}

	if categories[0].Name == "name1" {
		t.Errorf("expected name of %s, got: %s", "name1", categories[0].Name)
	}
}

func TestCategoryService_Update(t *testing.T) {
	db := newTestDB(t, "category_update")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	want := domain.Category{Name: "name"}
	err := postgres.NewCategoryStore(db).Create(ctx, &want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	duplicatedName := "duplicated_name"
	got, err := postgres.NewCategoryStore(db).Update(ctx, 1, domain.CategoryUpdate{
		Name:    &duplicatedName,
		Version: 1,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	want, err = postgres.NewCategoryStore(db).GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if want.Name != got.Name {
		t.Errorf("expected name of %s, got %s", want.Name, got.Name)
	}

	err = postgres.NewCategoryStore(db).Create(ctx, &domain.Category{
		Name: "temporary_name",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = postgres.NewCategoryStore(db).Update(ctx, 2, domain.CategoryUpdate{
		Name:    &duplicatedName,
		Version: 1,
	})
	if err != domain.ErrDuplicatedCategory {
		t.Errorf("expected %q from Update, got %q", domain.ErrDuplicatedCategory, err)
	}

	invalidVersion := 10
	_, err = postgres.NewCategoryStore(db).Update(ctx, 1, domain.CategoryUpdate{
		Version: invalidVersion,
	})
	if err != domain.ErrCategoryConflict {
		t.Errorf("expected %q from Update, got %q", domain.ErrCategoryConflict, err)
	}
}

func TestCategoryService_Delete(t *testing.T) {
	db := newTestDB(t, "category_delete")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	category := domain.Category{Name: "name"}
	err := postgres.NewCategoryStore(db).Create(ctx, &category)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = postgres.NewCategoryStore(db).Delete(ctx, 1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = postgres.NewCategoryStore(db).List(ctx, domain.CategoryFilter{})
	if err != domain.ErrNoCategoryFound {
		t.Fatalf("expected %q from List, got %q", domain.ErrNoCategoryFound, err)
	}

	err = postgres.NewCategoryStore(db).Delete(ctx, 1)
	if err != domain.ErrNoCategoryFound {
		t.Errorf("expected %q from Delete, got %q", domain.ErrNoCategoryFound, err)
	}
}
