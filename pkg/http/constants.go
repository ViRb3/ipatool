package http

type ResponseFormat string

const (
	ResponseFormatJSON ResponseFormat = "json"
	ResponseFormatXML  ResponseFormat = "xml"
	DefaultUserAgent                  = "Configurator/2.17 (Macintosh; OS X 15.2; 24C5089c) AppleWebKit/0620.1.16.11.6"
)
