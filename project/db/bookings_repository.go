package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tickets/entities"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/jmoiron/sqlx"
)

type BookingsRepository struct {
	db *sqlx.DB
}

func NewBookingsRepository(db *sqlx.DB) BookingsRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingsRepository{db: db}
}

var ErrOverBooking = errors.New("overbooking")

func (b BookingsRepository) Add(ctx context.Context, booking entities.Booking) (err error) {
	return updateInTx(
		ctx,
		b.db,
		sql.LevelSerializable, // make sure no over booking happens
		func(ctx context.Context, tx *sqlx.Tx) error {
			var availableSeats int
			err := tx.GetContext(ctx, &availableSeats,
				`SELECT 
					number_of_tickets 
				FROM 
					shows
				WHERE 
					show_id = $1`,
				booking.ShowID,
			)
			if err != nil {
				return fmt.Errorf("could not get available seats: %w", err)
			}

			var bookedSeats int
			err = tx.GetContext(ctx, &bookedSeats,
				`SELECT
					COALESCE(SUM(number_of_tickets), 0) 
				FROM 
					bookings
				WHERE 
					show_id = $1`,
				booking.ShowID,
			)
			if err != nil {
				return fmt.Errorf("could not get booked seats: %w", err)
			}

			if bookedSeats+booking.NumberOfTickets > availableSeats {
				return ErrOverBooking
			}

			_, err = tx.NamedExecContext(
				ctx,
				`INSERT INTO 
					bookings (booking_id, show_id, number_of_tickets, customer_email) 
				VALUES 
					(:booking_id, :show_id, :number_of_tickets, :customer_email)
				ON CONFLICT DO NOTHING`,
				booking,
			)
			if err != nil {
				return fmt.Errorf("could not save booking: %w", err)
			}

			outboxPublisher, err := outbox.NewPostgresPublisher(ctx, tx)
			if err != nil {
				return fmt.Errorf("could not create event bus: %w", err)
			}

			bus := event.NewBus(outboxPublisher)

			event := entities.BookingMade{
				Header:          entities.NewMessageHeader(),
				BookingID:       booking.BookingID,
				ShowID:          booking.ShowID,
				NumberOfTickets: booking.NumberOfTickets,
				CustomerEmail:   booking.CustomerEmail,
			}

			err = bus.Publish(ctx, event)
			if err != nil {
				return fmt.Errorf("could not publish event: %w", err)
			}

			return nil
		},
	)
}
