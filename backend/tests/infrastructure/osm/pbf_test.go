package osmdata_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/repository/osm"
	"github.com/paulmach/osm"
	"google.golang.org/protobuf/encoding/protowire"
)

func tinyPBF() []byte {

	stringField := func(msg []byte, field protowire.Number, value string) []byte {
		msg = protowire.AppendTag(msg, field, protowire.BytesType)

		return protowire.AppendString(msg, value)
	}
	bytesField := func(msg []byte, field protowire.Number, value []byte) []byte {
		msg = protowire.AppendTag(msg, field, protowire.BytesType)

		return protowire.AppendBytes(msg, value)
	}
	varintField := func(msg []byte, field protowire.Number, value uint64) []byte {
		msg = protowire.AppendTag(msg, field, protowire.VarintType)

		return protowire.AppendVarint(msg, value)
	}
	block := func(kind string, payload []byte) []byte {
		blob := bytesField(nil, 1, payload)
		header := stringField(nil, 1, kind)
		header = varintField(header, 3, uint64(len(blob)))
		prefix := make([]byte, 4)
		binary.BigEndian.PutUint32(prefix, uint32(len(header)))

		return append(append(prefix, header...), blob...)
	}
	header := stringField(nil, 4, "OsmSchema-V0.6")
	header = stringField(header, 4, "DenseNodes")
	node := bytesField(nil, 1, protowire.AppendVarint(nil, protowire.EncodeZigZag(123)))
	node = bytesField(node, 8, protowire.AppendVarint(nil, protowire.EncodeZigZag(599000000)))
	node = bytesField(node, 9, protowire.AppendVarint(nil, protowire.EncodeZigZag(302500000)))
	node = bytesField(node, 10, protowire.AppendVarint(nil, 0))
	stringTable := stringField(nil, 1, "")
	group := bytesField(nil, 2, node)
	primitive := bytesField(nil, 1, stringTable)
	primitive = bytesField(primitive, 2, group)

	return append(block("OSMHeader", header), block("OSMData", primitive)...)
}

func TestPBFScannerReadsNode(t *testing.T) {

	scanner, err := osmdata.NewPBFScanner(context.Background(), bytes.NewReader(tinyPBF()), 1)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := scanner.Close(); err != nil {
			t.Error(err)
		}
	}()

	if !scanner.Scan() {
		t.Fatalf("expected node, got %v", scanner.Err())
	}

	node, ok := scanner.Object().(*osm.Node)

	if !ok || node.ID != 123 || math.Abs(node.Lat-59.9) > 1e-9 || math.Abs(node.Lon-30.25) > 1e-9 {
		t.Fatalf("unexpected object: %#v", scanner.Object())
	}

	if scanner.Scan() {
		t.Fatal("expected only one node")
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestPBFScannerRejectsTruncatedInput(t *testing.T) {

	scanner, err := osmdata.NewPBFScanner(context.Background(), bytes.NewReader([]byte{0, 0}), 1)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = scanner.Close()
	}()

	if scanner.Scan() || scanner.Err() == nil {
		t.Fatal("truncated PBF must report an error")
	}
}

func TestPBFScannerStopsOnCancellation(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	scanner, err := osmdata.NewPBFScanner(ctx, bytes.NewReader(tinyPBF()), 1)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = scanner.Close()
	}()

	if !scanner.Scan() {
		t.Fatalf("expected node before cancellation: %v", scanner.Err())
	}

	cancel()

	if scanner.Scan() || !errors.Is(scanner.Err(), context.Canceled) {
		t.Fatalf("expected cancellation, got %v", scanner.Err())
	}
}

func TestPBFScannerValidatesArguments(t *testing.T) {

	cases := []struct {
		name    string
		ctx     context.Context
		reader  io.Reader
		workers int
	}{
		{"nil context", nil, bytes.NewReader(nil), 1},
		{"nil reader", context.Background(), nil, 1},
		{"no decoders", context.Background(), bytes.NewReader(nil), 0},
		{"negative decoders", context.Background(), bytes.NewReader(nil), -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := osmdata.NewPBFScanner(tc.ctx, tc.reader, tc.workers); err == nil {
				t.Fatal("expected argument error")
			}
		})
	}
}
