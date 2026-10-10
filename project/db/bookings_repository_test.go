package db_test

import (
	"context"
	"sync"
	"testing"
	ticketsDb "tickets/db"
	"tickets/entities"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingsRepository_Add_seats_limit(t *testing.T) {
	ctx := context.Background()

	db := getDb()

	err := ticketsDb.InitializeDBSchema(db)
	require.NoError(t, err)

	bookingsRepo := ticketsDb.NewBookingsRepository(db)
	showsRepo := ticketsDb.NewShowsRepository(db)

	t.Run("overbooking", func(t *testing.T) {
		showID := uuid.New()
		availableSeats := 2

		err := showsRepo.Add(ctx, entities.Show{
			ShowID:          showID.String(),
			DeadNationID:    uuid.New().String(),
			NumberOfTickets: availableSeats,
			StartTime:       time.Now().Add(time.Hour),
			Title:           "Example title",
			Venue:           "Example venue",
		})
		require.NoError(t, err)

		err = bookingsRepo.Add(ctx, entities.Booking{
			BookingID:       uuid.New().String(),
			ShowID:          showID.String(),
			NumberOfTickets: availableSeats,
			CustomerEmail:   "foo@bar.com",
		})
		require.NoError(t, err)

		err = bookingsRepo.Add(ctx, entities.Booking{
			BookingID:       uuid.New().String(),
			ShowID:          showID.String(),
			NumberOfTickets: availableSeats,
			CustomerEmail:   "foo@bar.com",
		})
		require.ErrorIs(t, err, ticketsDb.ErrOverBooking)
	})

	t.Run("parallel_overbooking", func(t *testing.T) {
		showID := uuid.New()
		availableSeats := 2

		err := showsRepo.Add(ctx, entities.Show{
			ShowID:          showID.String(),
			DeadNationID:    uuid.New().String(),
			NumberOfTickets: availableSeats,
			StartTime:       time.Now().Add(time.Hour),
			Title:           "Example title",
			Venue:           "Example venue",
		})
		require.NoError(t, err)

		workersCount := 50
		workersErrs := make(chan error, workersCount)
		unlock := make(chan struct{})

		wg := sync.WaitGroup{}
		wg.Add(workersCount)

		for range workersCount {
			go func() {
				defer wg.Done()

				<-unlock

				err = bookingsRepo.Add(ctx, entities.Booking{
					BookingID:       uuid.New().String(),
					ShowID:          showID.String(),
					NumberOfTickets: availableSeats,
					CustomerEmail:   "foo@bar.com",
				})

				workersErrs <- err
			}()
		}

		close(unlock)

		wg.Wait()
		close(workersErrs)

		failedWorkers := 0
		succeededWorkers := 0
		errors := []error{}

		for err := range workersErrs {
			if err != nil {
				failedWorkers++
				errors = append(errors, err)
			} else {
				succeededWorkers++
			}
		}

		assert.Equal(t, 1, succeededWorkers)
		assert.Equal(t, workersCount-1, failedWorkers)
	})
}
