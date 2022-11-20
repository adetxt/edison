package edison

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
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
	gwmux        *runtime.ServeMux
	grpcGateways []GRPCGatewayHandler
}

func NewEdison() *Edison {
	return &Edison{
		ec: echo.New(),
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

func (c *RestContext) JSON(code int, i interface{}, message string) error {
	isOK := code < 400

	res := map[string]interface{}{}

	if !isOK {
		res["status"] = "error"
		res["error"] = strings.ToUpper(http.StatusText(code))
		res["message"] = message
	} else {
		res["status"] = "success"
		res["message"] = strings.ToUpper(http.StatusText(code))
		res["data"] = i
	}

	return c.EchoContext.JSON(code, res)
}

func (c *RestContext) Success(i interface{}) error {
	return c.JSON(200, i, "")
}

func (c *RestContext) Error(e error, code int) error {
	return c.JSON(code, nil, e.Error())
}

func (c *RestContext) ErrorWithCustomMessage(code int, message string) error {
	return c.JSON(code, nil, message)
}
