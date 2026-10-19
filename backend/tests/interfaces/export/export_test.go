package export_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/interfaces/export"
)

func fixture() logistics.LogisticsGraph {
	return logistics.LogisticsGraph{ID: "g", RegionID: "r", Nodes: []logistics.LogisticsNode{{ID: "a", Name: "=SUM(1,2)"}, {ID: "b", Name: "Порт & терминал"}, {ID: "isolated"}}, Edges: []logistics.LogisticsEdge{{ID: "e1", FromNodeID: "a", ToNodeID: "b", Mode: logistics.TransportRoad, DistanceKM: 3, Directed: true, Geometry: []logistics.Coordinates{{Longitude: 30, Latitude: 60}, {Longitude: 31, Latitude: 61}}}, {ID: "e2", FromNodeID: "a", ToNodeID: "b", Mode: logistics.TransportRail, DistanceKM: 2, Directed: true}}}
}
func TestExports(t *testing.T) {
	for _, format := range export.Formats {
		t.Run(format, func(t *testing.T) {
			var b bytes.Buffer
			if err := export.Write(&b, fixture(), format); err != nil {
				t.Fatal(err)
			}
			switch format {
			case "geojson":
				var collection struct {
					Features []struct {
						Geometry struct {
							Type        string
							Coordinates json.RawMessage
						}
					}
				}
				if err := json.Unmarshal(b.Bytes(), &collection); err != nil {
					t.Fatal(err)
				}
				if len(collection.Features) != 5 || collection.Features[0].Geometry.Type != "Point" {
					t.Fatal(b.String())
				}
			case "graphml":
				var document struct {
					Graph struct {
						Nodes []struct {
							ID string `xml:"id,attr"`
						} `xml:"node"`
						Edges []struct {
							Directed bool `xml:"directed,attr"`
						} `xml:"edge"`
					} `xml:"graph"`
				}
				if err := xml.Unmarshal(b.Bytes(), &document); err != nil {
					t.Fatal(err)
				}
				if len(document.Graph.Nodes) != 3 || !document.Graph.Edges[0].Directed {
					t.Fatal(b.String())
				}
			default:
				rows, err := csv.NewReader(&b).ReadAll()
				if err != nil {
					t.Fatal(err)
				}
				if format == "adjacency.csv" && (rows[1][2] != "2" || rows[2][1] != "" || rows[3][1] != "") {
					t.Fatalf("adjacency=%v", rows)
				}
				if format == "incidence.csv" && (rows[1][1] != "-1" || rows[2][1] != "1" || rows[3][1] != "0") {
					t.Fatalf("incidence=%v", rows)
				}
				if format == "nodes.csv" && !strings.HasPrefix(rows[1][1], "'") {
					t.Fatal("spreadsheet formula not escaped")
				}
			}
		})
	}
}
