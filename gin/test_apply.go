package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gin/internal/repository/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gin/internal/app"
)

func main() {
	app.LoadConfig() // load .env natively or manually fallback
	db, err := sql.Open("pgx", "postgresql://ff789:vbW7h40mp6ZXV7dI@localhost:5432/ff789?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	repo := postgres.NewDepositRepository(db)

	ctx := context.Background()

	clientRef := "DEP-5616cfd3447ea7b954b58bf4a34efdea"
	
	_, err = repo.FindDepositIntentByClientRef(ctx, clientRef)
	if err != nil {
		fmt.Printf("Mocking intent since it wasn't found: %v\n", err)
		_, err = repo.CreateDepositIntent(ctx, postgres.CreateDepositIntentParams{
			UserID: 1, // Assume user 1 exists
			WalletID: 1, // Assume wallet 1 exists
			ClientRef: clientRef,
			Unit: 1,
			Type: 1,
			Amount: "50000",
			Status: 1,
			Provider: "sepay",
		})
		if err != nil {
			fmt.Printf("Could not create mock intent: %v\n", err)
		}
	}

	res, err := repo.ApplyDeposit(ctx, postgres.ApplyDepositParams{
		Provider:       "sepay",
		ProviderStatus: "finished",
		ClientRef:      clientRef,
		ProviderTxnID:  "FT26103422070458",
		Amount:         "50000",
		Currency:       "VND",
		PaidAt:         time.Now(),
		Raw:            map[string]any{"test": 1},
	})
	
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: %+v\n", res)
	}
}
