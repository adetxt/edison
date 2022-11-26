package edison

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

type (
	RestContext struct {
		EchoContext echo.Context
	}

	RestHandler        func(ctx context.Context, clientCtx RestContext) error
	GRPCGatewayHandler func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
)

type Edison struct {
	ec           *echo.Echo
	grpcEnabled  bool
	grpcServer   *grpc.Server
	serveMux     *runtime.ServeMux
	grpcGateways []GRPCGatewayHandler
	option       option
}

type EdisonJSONSerializer struct {
	echo.DefaultJSONSerializer
}

func (d EdisonJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	code := c.Response().Status
	isOK := code < 400

	res := map[string]interface{}{}

	if !isOK {
		res["status"] = "error"
		res["code"] = code
		res["error"] = strings.ToUpper(http.StatusText(code))
		res["message"] = i
	} else {
		res["status"] = "success"
		res["message"] = strings.ToUpper(http.StatusText(code))
		res["data"] = i
	}

	return d.DefaultJSONSerializer.Serialize(c, res, indent)
}

func New() *Edison {
	ec := echo.New()
	ec.HideBanner = true
	ec.Debug = true
	ec.JSONSerializer = EdisonJSONSerializer{}

	ec.Use(
		middleware.Logger(),
	)

	ec.GET("/__health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "failed")
	})

	return &Edison{
		ec: ec,
	}
}

func (ed *Edison) RestRouter(method, path string, h RestHandler) {
	ed.ec.Router().Add(method, path, func(c echo.Context) error {
		return h(context.Background(), RestContext{
			EchoContext: c,
		})
	})
}

func (c *RestContext) Bind(i interface{}) error {
	return c.EchoContext.Bind(i)
}
