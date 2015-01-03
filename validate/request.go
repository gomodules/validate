package validate

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/casualjim/go-swagger"
	"github.com/casualjim/go-swagger/httputils"
	"github.com/casualjim/go-swagger/spec"
)

var textUnmarshalType = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()

type formats map[string]map[string]reflect.Type

type operationBinder struct {
	Parameters map[string]spec.Parameter
	Consumers  map[string]swagger.Consumer
	Formats    formats
}

func (o *operationBinder) Bind(request *http.Request, routeParams swagger.RouteParams, data interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(data))
	isMap := val.Kind() == reflect.Map
	for fieldName, param := range o.Parameters {
		binder := new(paramBinder)
		binder.name = fieldName
		binder.parameter = &param
		binder.consumers = o.Consumers
		binder.formats = o.Formats
		binder.request = request
		binder.routeParams = routeParams

		if !isMap {
			binder.target = val.FieldByName(fieldName)
		}

		if isMap {
			binder.name = param.Name
			tpe := binder.Type()
			if tpe == nil {
				continue
			}
			binder.target = reflect.Indirect(reflect.New(tpe))
		}

		if !binder.target.IsValid() {
			return fmt.Errorf("parameter name %q is an unknown field", binder.name)
		}

		if err := binder.Bind(); err != nil {
			return err
		}

		if isMap {
			val.SetMapIndex(reflect.ValueOf(param.Name), binder.target)
		}
	}

	return nil
}

const defaultMaxMemory = 32 << 20

func contentType(req *http.Request) (string, error) {
	mt, _, err := httputils.ContentType(req.Header)
	if err != nil {
		return "", err
	}
	return mt, nil
}

func readSingle(from getValue, name string) string {
	return from.Get(name)
}

var evaluatesAsTrue = []string{"true", "1", "yes", "ok", "y", "on", "selected", "checked", "t", "enabled"}

func split(data, format string) []string {
	if data == "" {
		return nil
	}
	var sep string
	switch format {
	case "ssv":
		sep = " "
	case "tsv":
		sep = "\t"
	case "pipes":
		sep = "|"
	case "multi":
		return nil
	default:
		sep = ","
	}
	var result []string
	for _, s := range strings.Split(data, sep) {
		if ts := strings.TrimSpace(s); ts != "" {
			result = append(result, ts)
		}
	}
	return result
}
