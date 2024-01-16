package restserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"mymicro/micro/server/restserver/pprof"
	"mymicro/micro/server/restserver/validation"
	"mymicro/pkg/log"
	"net/http"
	"time"

	mws "mymicro/micro/server/restserver/middlewares"
)

type JwtInfo struct {
	Realm      string        `json:"realm"`
	Key        string        `json:"key"`
	Timeout    time.Duration `json:"timeout"`
	MaxRefresh time.Duration `json:"max_refresh"`
}

type Server struct {
	*gin.Engine
	port int
	// 开发模式
	mode string
	// 是否开启健康检查
	enableHealth bool
	// 是否开启pprof接口
	enableProfiling bool
	middlewares     []string
	jwt             *JwtInfo

	// 翻译器
	transName string
	trans     ut.Translator

	server      *http.Server
	serviceName string
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		port:            8080,
		mode:            "debug",
		enableHealth:    true,
		enableProfiling: true,
		jwt: &JwtInfo{
			"JWT",
			"jsGUSNksdfhsdixakjklwox5lf9jqzmw",
			7 * 24 * time.Hour,
			7 * 24 * time.Hour,
		},
		Engine:      gin.Default(),
		transName:   "zh",
		serviceName: "micro",
	}
	for _, o := range opts {
		o(srv)
	}

	srv.Use(mws.TracingHandler(srv.serviceName))

	for _, m := range srv.middlewares {
		mw, ok := mws.Middlewares[m]
		if !ok {
			log.Warnf("Can not find middleware: %s", m)
			continue
		}
		log.Infof("Install middleware: %s", m)
		srv.Use(mw)
	}
	return srv
}

func (s *Server) Start(ctx context.Context) error {
	if s.mode != gin.DebugMode && s.mode != gin.ReleaseMode && s.mode != gin.TestMode {
		return errors.New("mode must be one of debug/release/test")
	}
	gin.SetMode(s.mode)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Infof("%-6s %-s --> %s(%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// 初始化翻译器
	err := s.initTrans(s.transName)
	if err != nil {
		log.Errorf("InitTrans error %s", err.Error())
		return err
	}
	// 注册mobile验证码
	validation.RegisterMobile(s.trans)

	// 根据配置初始化pprof路由
	if s.enableProfiling {
		pprof.Register(s.Engine)
	}

	log.Infof("Rest server is running on port: %d", s.port)
	address := fmt.Sprintf(":%d", s.port)
	s.server = &http.Server{
		Addr:    address,
		Handler: s.Engine,
	}
	_ = s.SetTrustedProxies(nil)
	if err = s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Infof("Rest server is stopping")
	if err := s.server.Shutdown(ctx); err != nil {
		log.Errorf("Rest server shutdown error: %s", err.Error())
		return err
	}
	log.Info("Rest server stopped")
	return nil
}
