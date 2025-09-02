type {{.Name}}HTTPServer struct {
	server {{.ServiceName}}
	router gin.IRouter
}

func Register{{.ServiceName}}HTTPServer(srv {{.ServiceName}}, r gin.IRouter) {
	s := {{.Name}}HTTPServer{
		server: srv,
		router: r,
	}
	s.RegisterService()
}
{{range .Methods}}
func (s *{{$.Name}}HTTPServer) {{.HandlerName}}(c *gin.Context) {
	var in {{.Request}}
{{if eq .Method "GET" "DELETE"}}
	if err := c.ShouldBindQuery(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
{{else if eq .Method "POST" "PUT" "PATCH"}}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
{{else}}
	if err := c.ShouldBind(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
{{end}}
{{if .HasPathParams}}{{range $param := .PathParams}}
	in.{{ProtoFieldName $param}} = c.Param("{{$param}}")
{{end}}{{end}}
	ctx := c.Request.Context()
	out, err := s.server.{{.Name}}(ctx, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
{{end}}
func (s *{{.Name}}HTTPServer) RegisterService() {
{{range .Methods}}	s.router.Handle("{{.Method}}", "{{.Path}}", s.{{.HandlerName}})
{{end}}
}