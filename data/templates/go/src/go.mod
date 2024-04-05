module {{.app.name}}

go 1.19

{{if .app.options.use_libhlapi_module}}require hlapi v1.0.0{{end}}
{{if .app.options.use_libhlapi_module}}replace hlapi => ../libhlapi/hlapi_go{{end}}


