package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/interfaces/export"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/service/logistics_service"
	"github.com/gin-gonic/gin"
)

func NewRouter(service logistics_service.Queries) *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "Logistics Network API"})
	})
	router.GET("/api/regions", func(c *gin.Context) {
		regions, err := service.Regions(c.Request.Context())

		if err != nil {
			respondError(c, err)

			return
		}

		c.JSON(http.StatusOK, regions)
	})
	router.POST("/api/graphs", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
		decoder := json.NewDecoder(c.Request.Body)
		decoder.DisallowUnknownFields()
		var request struct {
			RegionID *string `json:"region_id"`
			Bounds   *struct {
				West  *float64 `json:"west"`
				South *float64 `json:"south"`
				East  *float64 `json:"east"`
				North *float64 `json:"north"`
			} `json:"bounds"`
		}

		if err := decoder.Decode(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ожидается JSON с region_id или bounds."})

			return
		}

		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) || (request.RegionID == nil) == (request.Bounds == nil) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите только region_id или bounds с четырьмя координатами."})

			return
		}

		var result logistics_service.BuildResult
		var err error

		if request.Bounds != nil {
			b := request.Bounds

			if b.West == nil || b.South == nil || b.East == nil || b.North == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите west, south, east и north."})

				return
			}

			bounds := logistics.Bounds{West: *b.West, South: *b.South, East: *b.East, North: *b.North}

			if err := bounds.Validate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные границы: долгота от −180 до 180, широта от −90 до 90; west < east, south < north."})

				return
			}

			result, err = service.RequestAreaBuild(c.Request.Context(), bounds)
		} else {
			id := *request.RegionID

			if id == "" || len(id) > 128 || strings.TrimSpace(id) != id {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный region_id."})

				return
			}

			result, err = service.RequestBuild(c.Request.Context(), id)
		}

		if err != nil {
			respondError(c, err)

			return
		}

		if result.GraphID != "" {
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusAccepted, result)
		}
	})
	router.GET("/api/jobs/:id", func(c *gin.Context) {
		job, err := service.GetJob(c.Request.Context(), c.Param("id"))

		if err != nil {
			respondError(c, err)

			return
		}

		c.JSON(http.StatusOK, job)
	})
	router.GET("/api/graphs/:id", func(c *gin.Context) {
		graph, err := service.GetGraph(c.Request.Context(), c.Param("id"))

		if err != nil {
			respondError(c, err)

			return
		}

		mode := logistics.TransportMode(c.Query("mode"))

		if mode != "" {
			if !mode.Valid() {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неизвестный вид транспорта."})

				return
			}

			nodes := []logistics.LogisticsNode{}
			edges := []logistics.LogisticsEdge{}

			for _, n := range graph.Nodes {
				if slices.Contains(n.Modes, mode) {
					nodes = append(nodes, n)
				}
			}

			for _, e := range graph.Edges {
				if e.Mode == mode {
					edges = append(edges, e)
				}
			}

			graph.Nodes = nodes
			graph.Edges = edges
		}

		c.JSON(http.StatusOK, graph)
	})
	router.GET("/api/graphs/:id/coverage", func(c *gin.Context) {
		graph, err := service.GetGraph(c.Request.Context(), c.Param("id"))

		if err != nil {
			respondError(c, err)

			return
		}

		c.JSON(http.StatusOK, graph.Coverage)
	})
	router.GET("/api/graphs/:id/export/:format", func(c *gin.Context) {
		format := c.Param("format")

		if !slices.Contains(export.Formats, format) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неизвестный формат экспорта.", "formats": export.Formats})

			return
		}

		graph, err := service.GetGraph(c.Request.Context(), c.Param("id"))

		if err != nil {
			respondError(c, err)

			return
		}

		if (format == "adjacency.csv" && len(graph.Nodes) > 5000) || (format == "incidence.csv" && len(graph.Nodes) > 0 && len(graph.Edges) > 25000000/len(graph.Nodes)) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Матрица слишком велика; используйте edges.csv или GraphML."})

			return
		}

		c.Header("Content-Type", export.ContentType(format))
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=graph-%s", format))

		if err := export.Write(c.Writer, graph, format); err != nil {
			slog.Error("export graph", "error", err)
			c.Abort()
		}
	})

	return router
}

func respondError(c *gin.Context, err error) {

	if errors.Is(err, logistics_service.ErrAreaUnavailable) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Для этой области нет источника данных. Настройте покрывающую выгрузку OSM или Overpass."})

		return
	}

	if errors.Is(err, logistics_service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор."})

		return
	}

	if errors.Is(err, logistics_service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Регион, задача или граф не найдены."})

		return
	}

	slog.Error("API repository operation", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хранилища. Повторите запрос позже."})
}
