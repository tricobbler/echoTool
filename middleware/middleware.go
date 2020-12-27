package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/maybgit/glog"
	"github.com/spf13/cast"
	r "github.com/tricobbler/echoTool/httpError"
	"runtime"
)

//校验渠道id和来源，并写入context
func Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			//排除swagger文档
			//if strings.Contains(c.Request().URL.Path, "/swagger/") {
			//	return next(c)
			//}
			//
			//p := struct {
			//	ChannelId string `query:"channel_id" validate:"required" label:"渠道id"`
			//	UserAgent string `query:"user_agent" validate:"required" label:"来源id"`
			//}{
			//	c.Request().Header.Get("channel_id"),
			//	c.Request().Header.Get("user_agent"),
			//}
			//if err := c.Validate(p); err != nil {
			//	err := validate.Translate(err.(validator.ValidationErrors))
			//	return r.NewHTTPError(400, err.One())
			//}

			c.Set("channel_id", cast.ToInt32(c.Request().Header.Get("channel_id")))
			c.Set("user_agent", cast.ToInt32(c.Request().Header.Get("user_agent")))
			return next(c)
		}
	}
}

func MyRecover(config middleware.RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = middleware.DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if rErr := recover(); rErr != nil {
					err, ok := rErr.(error)
					if !ok {
						err = fmt.Errorf("%v", rErr)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						glog.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}
					r.NewDefaultHttpError()
				}
			}()
			return next(c)
		}
	}
}

//返回错误信息处理，屏蔽内部服务调用错误
func MyErrorHandle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			if err != nil {
				glog.Errorf("[内部错误]，%v，%v", c.Path(), err)
				return r.NewDefaultHttpError()
			}

			return err
		}
	}
}
