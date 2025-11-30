package main

import (
	"context"
	"fmt"
	"log"
	"slices"

	"google.golang.org/api/sheets/v4"
)

type sheet[T any] struct {
	name        string
	columnNames []string
	values      []func(x *T) string
}

type sheetUpdateState struct {
	header  []string
	values  [][]string
	rows    [][]any
	success bool
}

func updateSheet[T any](
	client *sheets.SpreadsheetsService,
	ctx context.Context,
	spreadsheet string,
	state *sheetUpdateState,
	sheet *sheet[T],
	items []T,
) (err error) {
	// if there is an update or if the previous update was not successfully
	// submitted, then attempt another update
	updated := computeUpdate(state, sheet, items)
	if updated || !state.success {
		log.Printf("DEBUG an update was computed; will update sheet")
		if _, err = client.Values.Clear(
			spreadsheet,
			sheet.name,
			&sheets.ClearValuesRequest{},
		).Context(ctx).Do(); err != nil {
			state.success = false
			return fmt.Errorf("clearing sheet `%s`: %w", sheet.name, err)
		}

		if _, err = client.Values.Update(
			spreadsheet,
			sheet.name,
			&sheets.ValueRange{Values: state.rows},
		).ValueInputOption("USER_ENTERED").Context(ctx).Do(); err != nil {
			state.success = false
			return fmt.Errorf("updating sheet `%s`: %w", sheet.name, err)
		}
	} else if !updated {
		log.Printf("DEBUG no update needed")
	}

	state.success = true
	return
}

func computeUpdate[T any](
	state *sheetUpdateState,
	sheet *sheet[T],
	items []T,
) (update bool) {
	// if there are not exactly the same number of elements in `items` and
	// `updater.values` then increase or decrease the size of `updater.values`
	// so they have exactly the same number of elements.
	if len(items) != len(state.values) {
		log.Printf(
			"DEBUG number of rows has changed from %d to %d",
			len(state.values),
			len(items),
		)
	}
	update = len(items) != len(state.values) ||
		len(items)+1 != len(state.rows)
	state.values = trunc(state.values, len(items))
	state.rows = trunc(state.rows, len(items)+1)

	// ensure the header row is correct
	if len(state.header) != len(sheet.columnNames) {
		log.Printf(
			"DEBUG number of headers has changed from %d to %d",
			len(state.header),
			len(sheet.columnNames),
		)
	}
	update = update ||
		len(state.header) != len(sheet.columnNames) ||
		len(state.rows[0]) != len(sheet.columnNames)
	state.header = trunc(state.header, len(sheet.columnNames))
	state.rows[0] = trunc(state.rows[0], len(sheet.columnNames))
	for i, s := range sheet.columnNames {
		if state.header[i] != s {
			log.Printf(
				"DEBUG header %d has changed from `%s` to `%s`",
				i,
				state.header[i],
				s,
			)
			state.header[i] = s
			state.rows[0][i] = s
			update = true
		}
	}

	// update rows for each item
	for i := range items {
		// ensure that each row is properly sized
		if len(state.rows[i+1]) != len(sheet.values) {
			log.Printf(
				"DEBUG size of row %d has changed from %d to %d",
				i+1,
				len(state.rows[i+1]),
				len(sheet.values),
			)
		}
		update = update ||
			len(state.rows[i+1]) != len(sheet.values) ||
			len(state.values[i]) != len(state.values[i])
		state.rows[i+1] = trunc(state.rows[i+1], len(sheet.values))
		state.values[i] = trunc(state.values[i], len(sheet.values))

		// ensure that each row has the proper values
		for j, vf := range sheet.values {
			s := vf(&items[i])
			if s != state.values[i][j] {
				log.Printf(
					"DEBUG value changed at (%d, %d) from `%s` to `%s`",
					i,
					j,
					state.values[i][j],
					s,
				)
				state.values[i][j] = s
				state.rows[i+1][j] = s
				update = true
			}
		}
	}

	return
}

func trunc[T any](buf []T, n int) []T {
	// 1. `buf[:0]`: truncate `buf` to zero
	// 2. `slices.Grow(..., n)`: increase the capacity to `n`
	// 3. `...[:n]`: increase the length to `n`
	return slices.Grow(buf[:0], n)[:n]
}
