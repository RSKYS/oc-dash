package backup

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"net/http"
)

type Controller struct {
	request         request.CustomRequestInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		ocservUserRepo:  repository.NewtOcservUserRepository(),
		ocservGroupRepo: repository.NewOcservGroupRepository(),
	}
}

func (ctl *Controller) OcservUserBackup(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_users_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err := ctl.ocservUserRepo.Backup(
		c.Request().Context(),
		gz,
	); err != nil {
		gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err := gz.Close(); err != nil {
		return err
	}

	return nil
}
