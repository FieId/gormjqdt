package gormjqdt

type RequestString string

type ParsedRequest struct {
	Draw                   int
	Start                  int
	Length                 int
	GlobalSearch           string
	GlobalSearchRegex      bool
	Columns                map[string]interface{}
	Orders                 map[string]interface{}
	SpesificParams         map[int]map[string]interface{}
	SpesificParamKeySlices map[string]int
}

type Response struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int64       `json:"recordsTotal"`
	RecordsFiltered int64       `json:"recordsFiltered"`
	Data            interface{} `json:"data,nilasempty"`
}
