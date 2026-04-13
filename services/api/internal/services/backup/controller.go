package backup

import (
	"compress/gzip"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"io"
	"net/http"
	"strings"
)

type Controller struct {
	request         request.CustomRequestInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	backupRepo      repository.BackupRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		ocservUserRepo:  repository.NewtOcservUserRepository(),
		ocservGroupRepo: repository.NewOcservGroupRepository(),
		backupRepo:      repository.NewBackupRepository(),
	}
}

// OcservGroupBackup
// @Summary      Backup ocserv groups
// @Description  Download gzip compressed JSON backup of all ocserv groups including default group configuration
// @Tags         System(Backup)
// @Produce      application/json
// @Produce      application/gzip
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200 {file} file "ocserv_groups_backup.json.gz"
// @Router       /backup/ocserv_groups [get]
func (ctl *Controller) OcservGroupBackup(c echo.Context) error {
	defaultGroup, err := ctl.ocservGroupRepo.DefaultGroup()
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_groups_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err = ctl.backupRepo.OcservGroupBackup(
		c.Request().Context(),
		gz,
		defaultGroup,
	); err != nil {
		gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err = gz.Close(); err != nil {
		return err
	}

	return nil
}

// OcservGroupRestore
// @Summary      Restore ocserv groups
// @Description  Upload JSON or gzip-compressed (.json.gz) backup of ocserv groups
// @Tags         System(Restore)
// @Produce      application/json
// @Accept       multipart/form-data
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        file formData file true "JSON or JSON.GZ file"
// @Success      200 {object} RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_groups [post]
func (ctl *Controller) OcservGroupRestore(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	src, err := file.Open()
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	defer src.Close()

	var reader io.Reader = src
	if strings.HasSuffix(file.Filename, ".gz") {
		gzReader, err := gzip.NewReader(src)
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}
		defer gzReader.Close()

		reader = gzReader
	}

	type groupFile struct {
		DefaultGroup *models.OcservGroupConfig `json:"default_group"`
		Groups       *[]models.OcservGroup     `json:"groups"`
	}

	var groupData groupFile
	if err = json.NewDecoder(reader).Decode(&groupData); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if err = ctl.ocservGroupRepo.UpdateDefaultGroup(groupData.DefaultGroup); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var inserted, existing *[]string

	if len(*groupData.Groups) > 0 {
		inserted, existing, err = ctl.backupRepo.OcservGroupRestore(c.Request().Context(), groupData.Groups)
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}

// OcservUserBackup
// @Summary      Backup ocserv users
// @Description  Download gzip compressed JSON backup of all ocserv users
// @Tags         System(Backup)
// @Produce      application/json
// @Produce      application/gzip
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200 {file} file "ocserv_users_backup.json.gz"
// @Router       /backup/ocserv_users [get]
func (ctl *Controller) OcservUserBackup(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_users_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err := ctl.backupRepo.OcservUserBackup(
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

// OcservUserRestore
// @Summary      Restore ocserv users
// @Description  Upload JSON or gzip-compressed (.json.gz) backup of ocserv users
// @Tags         System(Restore)
// @Produce      application/json
// @Accept       multipart/form-data
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        file formData file true "JSON or JSON.GZ file"
// @Success      200 {object} RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_users [post]
func (ctl *Controller) OcservUserRestore(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	src, err := file.Open()
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	defer src.Close()

	var reader io.Reader = src
	if strings.HasSuffix(file.Filename, ".gz") {
		gzReader, err := gzip.NewReader(src)
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}
		defer gzReader.Close()

		reader = gzReader
	}

	var users []models.OcservUser
	if err = json.NewDecoder(reader).Decode(&users); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if len(users) == 0 {
		return c.JSON(http.StatusOK, RestoreResponse{
			Inserted: nil,
			Existing: nil,
		})
	}

	inserted, existing, err := ctl.backupRepo.OcservUserRestore(c.Request().Context(), &users)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}
