package datasources

import "fmt"

type Datasource interface {
	Id() uint64
	Name() string
	Description() string
	Value() (uint64, error)
}

func GetAllDatasources() []Datasource {
	var datasources []Datasource
	datasources = append(datasources, &UsdBtc{})
	return datasources
}

func GetDatasource(id uint64) (Datasource, error) {
	switch id {
	case 1:
		return &UsdBtc{}, nil
	default:
		return nil, fmt.Errorf("Data source with ID %d not known", id)
	}
}

func HasDatasource(id uint64) bool {
	return (id <= 1)
}
