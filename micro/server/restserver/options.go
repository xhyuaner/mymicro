package restserver

type ServerOption func(*Server)

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

func WithEnableProfiling(profiling bool) ServerOption {
	return func(s *Server) {
		s.enableProfiling = profiling
	}
}

func WithEnableHealth(enableHealth bool) ServerOption {
	return func(s *Server) {
		s.enableHealth = enableHealth
	}
}

func WithMode(mode string) ServerOption {
	return func(s *Server) {
		s.mode = mode
	}
}

func WithServiceName(srvName string) ServerOption {
	return func(s *Server) {
		s.serviceName = srvName
	}
}

func WithMiddlewares(middlewares []string) ServerOption {
	return func(s *Server) {
		s.middlewares = middlewares
	}
}

func WithJwt(jwt *JwtInfo) ServerOption {
	return func(s *Server) {
		s.jwt = jwt
	}
}

func WithTransName(transName string) ServerOption {
	return func(s *Server) {
		s.transName = transName
	}
}
