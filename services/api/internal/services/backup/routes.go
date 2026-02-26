package backup

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/backup", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	g.GET("/ocserv_users", ctl.OcservUserBackup)
	g.GET("/ocserv_groups", ctl.OcservUserBackup)
}
