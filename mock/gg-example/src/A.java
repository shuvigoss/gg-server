项目名称：{{ .project }}
作者名称：{{ .author }}
爱好：
{{ range $index, $element := .hobby }} 
    {{$element}} 
{{ end }}