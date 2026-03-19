package remote

import (
	"encoding/json"
	"log/slog"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Serializer interface {
	Serialize(msg any) ([]byte, error)
	TypeName(any) string
}

type Deserializer interface {
	Deserialize([]byte, string) (any, error)
}

type VTMarshaler interface {
	proto.Message
	MarshalVT() ([]byte, error)
}

type VTUnmarshaler interface {
	proto.Message
	UnmarshalVT([]byte) error
}

// Todo: delete this or state why it isn't deleted.
// type DefaultSerializer struct{}

// func (DefaultSerializer) Serialize(msg any) ([]byte, error) {
// 	switch msg.(type) {
// 	case VTMarshaler:
// 		return VTProtoSerializer{}.Serialize(msg)
// 	case proto.Message:
// 		return ProtoSerializer{}.Serialize(msg)
// 	default:
// 		return nil, fmt.Errorf("unsupported message type (%v) for serialization", reflect.TypeOf(msg))
// 	}
// }

// func (DefaultSerializer) Deserialize(data []byte, mtype string) (any, error) {
// 	switch msg.(type) {
// 	case VTMarshaler:
// 		return VTProtoSerializer{}.Serialize(msg)
// 	case proto.Message:
// 		return ProtoSerializer{}.Serialize(msg)
// 	default:
// 		return nil, fmt.Errorf("unsupported message type (%v) for serialization", reflect.TypeOf(msg))
// 	}
// }

type ProtoSerializer struct{}

func (ProtoSerializer) Serialize(msg any) ([]byte, error) {
	var b []byte
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback to JSON if proto.Marshal panics
				b, err = json.Marshal(msg)
			}
		}()
		b, err = proto.Marshal(msg.(proto.Message))
		if err != nil {
			// Fallback to JSON if proto.Marshal errors
			b, err = json.Marshal(msg)
		}
	}()
	return b, err
}

func (ProtoSerializer) Deserialize(data []byte, tname string) (any, error) {
	// 1. Try manual registry first (handles manual/JSON types)
	v, err := registryGetType(tname)
	if err == nil {
		pm := reflect.New(reflect.TypeOf(v).Elem()).Interface()
		err = json.Unmarshal(data, pm)
		return pm, err
	}
	slog.Info("registryGetType failed, falling back to proto", "tname", tname)

	// 2. Fallback to standard Protobuf registry
	pname := protoreflect.FullName(tname)
	n, err := protoregistry.GlobalTypes.FindMessageByName(pname)
	if err != nil {
		return nil, err
	}
	pm := n.New().Interface()
	err = proto.Unmarshal(data, pm)
	return pm, err
}

func (ProtoSerializer) TypeName(msg any) string {
	var tname string
	func() {
		defer func() {
			if r := recover(); r != nil {
				tname = ""
			}
		}()
		tname = string(proto.MessageName(msg.(proto.Message)))
	}()

	if tname == "" {
		typ := reflect.TypeOf(msg)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		tname = typ.Name()
	}
	return tname
}

type VTProtoSerializer struct{}

func (VTProtoSerializer) TypeName(msg any) string {
	var tname string
	func() {
		defer func() {
			if r := recover(); r != nil {
				tname = ""
			}
		}()
		tname = string(proto.MessageName(msg.(proto.Message)))
	}()

	if tname == "" {
		typ := reflect.TypeOf(msg)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		tname = typ.Name()
	}
	return tname
}

func (VTProtoSerializer) Serialize(msg any) ([]byte, error) {
	var b []byte
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback to standard proto or JSON
				if m, ok := msg.(proto.Message); ok {
					b, err = proto.Marshal(m)
					if err != nil {
						b, err = json.Marshal(msg)
					}
				} else {
					b, err = json.Marshal(msg)
				}
			}
		}()
		if vm, ok := msg.(VTMarshaler); ok {
			b, err = vm.MarshalVT()
		} else if pm, ok := msg.(proto.Message); ok {
			b, err = proto.Marshal(pm)
		} else {
			b, err = json.Marshal(msg)
		}
	}()
	return b, err
}

func (VTProtoSerializer) Deserialize(data []byte, mtype string) (any, error) {
	v, err := registryGetType(mtype)
	if err != nil {
		return nil, err
	}
	err = v.UnmarshalVT(data)
	return v, err
}
