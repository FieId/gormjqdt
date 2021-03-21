# GORMJQDT
Golang jQuery Datatable Server Side Pagination for gorm

[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/sutejoramadhan/gormjqdt.svg)](https://pkg.go.dev/github.com/sutejoramadhan/gormjqdt#section-documentation)

## 📖 Requirements
- [Go >= 1.15](https://golang.org/)
- [GORM >= 1.20.x](https://gorm.io/)
- [jQuery DataTables v1.10.x](http://datatables.net/)

## 🚀 Installation
This module is using and for gorm (golang ORM module), before you using this module make sure that you already install gorm.

```bash
go get -u github.com/sutejoramadhan/gormjqdt
```

## 👀 Overview
- Work on Common functiononal jQuery Datatable like sorting single or multi column, limit, global search and individual search
- Support automatic spesific filter (params besides common jQuery Datatable params).
- Support array value on spesific filter params
- Support RegExr filtering
- Support Json Body
- Support relation (not yet tested on nested relations)
- Support spesific conditional filter like between on date range.
> For date range, please set the request parameter to this pattern, i.e dateRange~~start dateRange~~end (add `~~start|end`)

## 🎸 Examples
Here is common and simple example about how using this module
```go
// Define your struct slice of table model
// Why slice? because Datatable response for the `data` is array. i.e data: [{}]
var arrayDatas []User

// Initialize DataTable
// Must set `Model` (pointer struct of table model) & `Engine` (your *gorm.DB context)
datatable, err := gormjqdt.New(gormjqdt.Config{
    Model:  &User{},
    Engine: bootstrap.DB,
})
if err != nil {
    return err.Error()
}

// Define the string of request parameters
// Whatever http module that you using, convert request struct to string
// Example below, using gofiber and the http method is POST with form url encoded body
req := string(ctx.Body())

// Get DataTable Response
// The gormjqdt datatable is returning struct with field that same like jQuery Datatable supported response.
response, err := datatable.Simple(gormjqdt.RequestString(req), &arrayDatas)
if err != nil {
    return err.Error()
}
```

Example using relation (complex)
```go
// Define your struct slice of table model
// Why slice? because Datatable response for the `data` is array. i.e data: [{}]
var arrayDatas []User

// Initialize DataTable
// Must set `Model` (pointer struct of table model) & `Engine` (your *gorm.DB context)
datatable, err := gormjqdt.New(gormjqdt.Config{
    Model:  &User{},
    Engine: bootstrap.DB,
})
if err != nil {
    return err.Error()
}

// Define the string of request parameters
// Whatever http module that you using, convert request struct to string
// Example below, using gofiber and the http method is POST with form url encoded body
req := string(ctx.Body())

// Define the relation
relations := make(map[int]interface{})
relations[0] = &Profile{}

// Get DataTable Response
// The gormjqdt datatable is returning struct with field that same like jQuery Datatable supported response.
response, err := datatable.Complex(gormjqdt.RequestString(req), &arrayDatas, relations)
if err != nil {
    return err.Error()
}
```

## 👨‍💻 To-Do
- Improvement and bug fixing