package remote

import (
	"fmt"
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
				err = fmt.Errorf("proto.Marshal panic: %v", r)
			}
		}()
		b, err = proto.Marshal(msg.(proto.Message))
	}()
	return b, err
}

func (ProtoSerializer) Deserialize(data []byte, tname string) (any, error) {
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
				// Fallback to standard proto if MarshalVT is missing or panics
				if m, ok := msg.(proto.Message); ok {
					b, err = proto.Marshal(m)
				} else {
					err = fmt.Errorf("VTMarshaler cast or MarshalVT panic: %v", r)
				}
			}
		}()
		b, err = msg.(VTMarshaler).MarshalVT()
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
