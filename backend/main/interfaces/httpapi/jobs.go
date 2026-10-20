package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
	"github.com/gin-gonic/gin"
)

func jobStatus(service logistics_service.Queries) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		wait := 0
		var after time.Time
		var err error

		if raw, present := c.GetQuery("wait"); present {
			wait, err = strconv.Atoi(raw)

			if err != nil || wait < 0 || wait > int(logistics_service.MaxJobWait/time.Second) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "wait должен быть целым числом от 0 до 25 секунд."})

				return
			}
		}

		if raw, present := c.GetQuery("after"); present {
			after, err = time.Parse(time.RFC3339Nano, raw)

			if err != nil || after.IsZero() {
				c.JSON(http.StatusBadRequest, gin.H{"error": "after должен содержать updated_at предыдущего ответа в формате RFC3339."})

				return
			}
		}

		if wait > 0 && after.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для ожидания передайте after с updated_at предыдущего ответа."})

			return
		}

		var job logistics_service.Job

		if wait == 0 {
			job, err = service.GetJob(c.Request.Context(), c.Param("id"))
		} else {
			job, err = service.WaitJob(c.Request.Context(), c.Param("id"), after, time.Duration(wait)*time.Second)
		}

		if err != nil {
			if c.Request.Context().Err() != nil {
				c.Abort()
			} else {
				respondError(c, err)
			}

			return
		}

		c.JSON(http.StatusOK, job)
	}
}
