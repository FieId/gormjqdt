# GORMJQDT
Golang jQuery Datatable Server Side Pagination for gorm

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

## 🎸 Examples
Here is common and simple example about how using this module
```go
// Define your struct slice of table mode
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

## 👨‍💻 To-Do
- Support Json Body (currently only support Form URL Encoded Body (POST) & URL Queries (GET))