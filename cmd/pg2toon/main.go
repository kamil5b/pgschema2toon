package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/kamil5b/pgschema2toon/pkg/toon"
)

const extractQuery = `
SELECT jsonb_strip_nulls(jsonb_agg(table_info))
FROM (
    SELECT jsonb_strip_nulls(jsonb_build_object(
        'table', t.relname,
        'comment', NULLIF(obj_description(t.oid, 'pg_class'), ''),
        'columns', (
            SELECT jsonb_agg(jsonb_strip_nulls(jsonb_build_object(
                'name', a.attname,
                'type', format_type(a.atttypid, a.atttypmod),
                'comment', NULLIF(col_description(t.oid, a.attnum), ''),
                'nullable', CASE WHEN a.attnotnull THEN NULL ELSE true END,
                'is_pk', EXISTS (
                    SELECT 1 FROM pg_index i
                    WHERE i.indrelid = t.oid AND a.attnum = ANY(i.indkey) AND i.indisprimary
                )
            )))
            FROM pg_attribute a
            WHERE a.attrelid = t.oid AND a.attnum > 0 AND NOT a.attisdropped
        ),
        'constraints', (
            SELECT NULLIF(jsonb_agg(jsonb_build_object(
                'name', c.conname,
                'def', pg_get_constraintdef(c.oid)
            )), '[]'::jsonb)
            FROM pg_constraint c
            WHERE c.conrelid = t.oid AND c.contype = 'f'
        ),
        'indexes', (
            SELECT NULLIF(jsonb_agg(jsonb_build_object(
                'name', i.relname,
                'def', pg_get_indexdef(i.oid)
            )), '[]'::jsonb)
            FROM pg_index x
            JOIN pg_class i ON i.oid = x.indexrelid
            WHERE x.indrelid = t.oid AND NOT x.indisprimary
        )
    )) AS table_info
    FROM pg_class t
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE t.relkind = 'r' AND n.nspname = 'public'
    ORDER BY t.relname
) sub;`

func main() {
	dbURL := flag.String("db", "", "Postgres URL")
	output := flag.String("out", "", "Output File")
	flag.Parse()

	if *dbURL == "" {
		fmt.Println("Usage: pg2toon -db <url>")
		os.Exit(1)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	var jsonRaw []byte
	err = conn.QueryRow(ctx, extractQuery).Scan(&jsonRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SQL Query failed: %v\n", err)
		os.Exit(1)
	}

	result, err := toon.ToToon(jsonRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conversion failed: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		os.WriteFile(*output, []byte(result), 0644)
	} else {
		fmt.Print(result)
	}
}
