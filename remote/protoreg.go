package remote

import (
	"fmt"
	"log/slog"
	"reflect"

	"google.golang.org/protobuf/proto"
)

var registry = map[string]VTUnmarshaler{}

func RegisterType(v VTUnmarshaler) {
	defer func() {
		if r := recover(); r != nil {
			// If proto.MessageName panics (e.g. due to nil ProtoReflect), 
			// we rely on the reflection fallback below.
		}
	}()
	if tname == "" {
		typ := reflect.TypeOf(v)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		tname = typ.Name()
	}
	slog.Info("registering type", "name", tname)
	registry[tname] = v
}

func registryGetType(t string) (VTUnmarshaler, error) {
	if m, ok := registry[t]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("given type (%s) is not registered. Did you forget to register your type with remote.RegisterType(&instance{})?", t)
}
