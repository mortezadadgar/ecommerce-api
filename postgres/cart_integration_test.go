package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/postgres"
)

func newCartTestDB(t *testing.T, dbName string) *pgxpool.Pool {
	db := newTestDB(t, dbName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := &domain.User{Email: "name@gmail.com", Password: []byte("123")}
	err := postgres.NewUserStore(db).Create(ctx, user)
	if err != nil {
		t.Fatalf("user Create: %v", err)
	}

	category := domain.Category{Name: "category"}
	err = postgres.NewCategoryStore(db).Create(ctx, &category)
	if err != nil {
		t.Fatalf("category Create: %v", err)
	}

	product := domain.Product{Name: "product", CategoryID: 1}
	err = postgres.NewProductStore(db).Create(ctx, &product)
	if err != nil {
		t.Fatalf("product Create: %v", err)
	}

	return db
}

func TestCartService_Create(t *testing.T) {
	db := newCartTestDB(t, "carts_create")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	want := domain.Cart{ProductID: 1, UserID: 1, Quantity: 1}
	err := postgres.NewCartStore(db).Create(ctx, &want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := postgres.NewCartStore(db).GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("mismatch\n got: %#v\nwant: %#v", got, want)
	}

	err = postgres.NewCartStore(db).Create(ctx, &domain.Cart{UserID: 99})
	if err != domain.ErrCartInvalidUserID {
		t.Errorf("expected %q from Create, got: %q", domain.ErrCartInvalidUserID, err)
	}

	err = postgres.NewCartStore(db).Create(ctx, &domain.Cart{ProductID: 99, UserID: 1})
	if err != domain.ErrCartInvalidProductID {
		t.Errorf("expected %q from Create, got: %q", domain.ErrCartInvalidProductID, err)
	}
}

func TestCartService_List(t *testing.T) {
	db := newCartTestDB(t, "carts_list")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// No cart is inserted into db.
	_, err := postgres.NewCartStore(db).List(ctx, domain.CartFilter{})
	if err != domain.ErrNoCartsFound {
		t.Errorf("expected %q from List, got %q", domain.ErrNoCartsFound, err)
	}

	const n = 3
	for i := 0; i < n; i++ {
		cart := domain.Cart{ProductID: 1, UserID: 1, Quantity: 1}
		err = postgres.NewCartStore(db).Create(ctx, &cart)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	carts, err := postgres.NewCartStore(db).List(ctx, domain.CartFilter{})
	if err != nil {
		t.Errorf("List: %v", err)
	}

	if len(carts) != n {
		t.Errorf("expected length of %d, got: %v", n, len(carts))
	}

	if carts[0].Quantity != 1 {
		t.Errorf("expected quantity of %d, got: %d", 1, carts[0].Quantity)
	}
}

func TestCartService_Update(t *testing.T) {
	db := newCartTestDB(t, "carts_update")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	want := domain.Cart{ProductID: 1, UserID: 1, Quantity: 1}
	err := postgres.NewCartStore(db).Create(ctx, &want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	want.Quantity = 10
	err = postgres.NewCartStore(db).Update(ctx, &want)
	if err != nil {
		t.Errorf("Update: %v", err)
	}

	got, err := postgres.NewCartStore(db).GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Quantity != want.Quantity {
		t.Errorf("expected quantity of %d, want: %d", got.Quantity, want.Quantity)
	}

	err = postgres.NewCartStore(db).Update(ctx, &domain.Cart{ID: 1, ProductID: 99})
	if err != domain.ErrCartInvalidProductID {
		t.Errorf("expected %q from Update, got %q", domain.ErrCartInvalidProductID, err)
	}

	err = postgres.NewCartStore(db).Update(ctx, &domain.Cart{ID: 99, ProductID: 1})
	if err != domain.ErrNoCartsFound {
		t.Errorf("expected %q from Update, got %q", domain.ErrNoCartsFound, err)
	}
}

func TestCartService_Delete(t *testing.T) {
	db := newCartTestDB(t, "carts_delete")
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cart := domain.Cart{ProductID: 1, UserID: 1, Quantity: 1}
	err := postgres.NewCartStore(db).Create(ctx, &cart)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = postgres.NewCartStore(db).Delete(ctx, 1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = postgres.NewCartStore(db).GetByID(ctx, 1)
	if err != domain.ErrNoCartsFound {
		t.Fatalf("expected %q from List, got %q", domain.ErrNoCartsFound, err)
	}

	err = postgres.NewCartStore(db).Delete(ctx, 1)
	if err != domain.ErrNoCartsFound {
		t.Fatalf("expected %q from List, got %q", domain.ErrNoCartsFound, err)
	}
}
