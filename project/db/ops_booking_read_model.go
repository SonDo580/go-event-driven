package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"tickets/entities"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
)

type OpsBookingReadModel struct {
	db *sqlx.DB
}

func NewOpsBookingReadModel(db *sqlx.DB) OpsBookingReadModel {
	if db == nil {
		panic("db is nil")
	}

	return OpsBookingReadModel{db: db}
}

func (r OpsBookingReadModel) OnBookingMade(ctx context.Context, event *entities.BookingMade) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return err
	}

	rm := entities.OpsBooking{
		BookingID:  bookingID,
		BookedAt:   event.Header.PublishedAt,
		LastUpdate: time.Now(),
	}

	err = r.createReadModel(ctx, rm)
	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) OnTicketBookingConfirmed(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	return r.updateReadModelByBookingID(
		ctx,
		event.BookingID,
		func(rm entities.OpsBooking) (entities.OpsBooking, error) {
			ticket, ok := rm.Tickets[event.TicketID]
			if !ok {
				// use zero-value of OpsTicket
			}

			ticket.PriceAmount = event.Price.Amount
			ticket.PriceCurrency = event.Price.Currency
			ticket.CustomerEmail = event.CustomerEmail
			ticket.Status = "confirmed"

			rm.Tickets[event.TicketID] = ticket

			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketReceiptIssued(ctx context.Context, event *entities.TicketReceiptIssued) error {
	return r.updateReadModelByTicketID(
		ctx,
		event.TicketID,
		func(rm entities.OpsTicket) (entities.OpsTicket, error) {
			rm.ReceiptNumber = event.ReceiptNumber
			rm.ReceiptIssuedAt = event.IssuedAt
			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketPrinted(ctx context.Context, event *entities.TicketPrinted) error {
	return r.updateReadModelByTicketID(
		ctx,
		event.TicketID,
		func(rm entities.OpsTicket) (entities.OpsTicket, error) {
			rm.PrintedFileName = event.FileName
			rm.PrintedAt = event.Header.PublishedAt
			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketRefunded(ctx context.Context, event *entities.TicketRefunded) error {
	return r.updateReadModelByTicketID(
		ctx,
		event.TicketID,
		func(rm entities.OpsTicket) (entities.OpsTicket, error) {
			rm.Status = "refunded"
			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) AllBookings(ctx context.Context, receiptIssueDate string) ([]entities.OpsBooking, error) {
	var rows *sql.Rows
	var err error

	if receiptIssueDate != "" {
		rows, err = r.db.QueryContext(
			ctx,
			`
				SELECT payload FROM read_model_ops_bookings
				WHERE booking_id IN (
					SELECT booking_id FROM (
						SELECT 
							booking_id, 
							DATE(jsonb_path_query(payload, '$.tickets.*.receipt_issued_at')::text) AS receipt_issued_at 
						FROM read_model_ops_bookings
					) 
					WHERE receipt_issued_at = $1
				)
			`,
			receiptIssueDate,
		)
	} else {
		rows, err = r.db.QueryContext(ctx, `SELECT payload FROM read_model_ops_bookings`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.OpsBooking
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		var reservation entities.OpsBooking
		if err := json.Unmarshal(payload, &reservation); err != nil {
			return nil, err
		}

		result = append(result, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r OpsBookingReadModel) ReservationReadModel(ctx context.Context, bookingID string) (entities.OpsBooking, error) {
	return r.findReadModelByBookingID(ctx, bookingID, r.db)
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r OpsBookingReadModel) createReadModel(
	ctx context.Context,
	booking entities.OpsBooking,
) (err error) {
	payload, err := json.Marshal(booking)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`
		INSERT INTO
			read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO NOTHING
		`,
		payload, booking.BookingID,
	)

	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) updateReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	updateFunc func(entities.OpsBooking) (entities.OpsBooking, error),
) (err error) {
	return updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead, // avoid lost updates (abort if a previously committed Tx updated the same row)
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByBookingID(ctx, bookingID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order -> nack to redeliver
				return fmt.Errorf("read model for booking %s no exists yet", bookingID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			updatedRm, err := updateFunc(rm)
			if err != nil {
				return err
			}

			return r.updateReadModel(ctx, tx, updatedRm)
		},
	)
}

func (r OpsBookingReadModel) updateReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	updateFunc func(entities.OpsTicket) (entities.OpsTicket, error),
) (err error) {
	return updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead, // avoid lost updates (abort if a previously committed Tx updated the same row)
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByTicketID(ctx, ticketID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order -> nack to redeliver
				return fmt.Errorf("read model for ticket %s no exists yet", ticketID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			ticket, _ := rm.Tickets[ticketID]

			updatedRm, err := updateFunc(ticket)
			if err != nil {
				return err
			}

			rm.Tickets[ticketID] = updatedRm

			return r.updateReadModel(ctx, tx, rm)
		},
	)
}

func (r OpsBookingReadModel) updateReadModel(
	ctx context.Context,
	tx *sqlx.Tx,
	rm entities.OpsBooking,
) error {
	rm.LastUpdate = time.Now()

	payload, err := json.Marshal(rm)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO
			read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO UPDATE SET payload = EXCLUDED.payload
		`,
		payload, rm.BookingID,
	)

	if err != nil {
		return fmt.Errorf("could not update read model: %w", err)
	}
	return nil
}

func (r OpsBookingReadModel) findReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	db dbExecutor,
) (entities.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		`
		SELECT payload FROM read_model_ops_bookings
		WHERE payload::jsonb -> 'tickets' ? $1
		`,
		ticketID,
	).Scan(&payload)

	if err != nil {
		return entities.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) findReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	db dbExecutor,
) (entities.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		`
		SELECT payload FROM read_model_ops_bookings
		WHERE booking_id = $1
		`,
		bookingID,
	).Scan(&payload)

	if err != nil {
		return entities.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) unmarshalReadModelFromDB(payload []byte) (entities.OpsBooking, error) {
	var dbReadModel entities.OpsBooking
	if err := json.Unmarshal(payload, &dbReadModel); err != nil {
		return entities.OpsBooking{}, err
	}

	if dbReadModel.Tickets == nil {
		dbReadModel.Tickets = map[string]entities.OpsTicket{}
	}

	return dbReadModel, nil
}
