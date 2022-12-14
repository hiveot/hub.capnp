// Code generated by capnpc-go. DO NOT EDIT.

package thing

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	schemas "capnproto.org/go/capnp/v3/schemas"
)

type TD capnp.Struct

// TD_TypeID is the unique identifier for the type TD.
const TD_TypeID = 0x9a428fdf25da5bb7

func NewTD(s *capnp.Segment) (TD, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return TD(st), err
}

func NewRootTD(s *capnp.Segment) (TD, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return TD(st), err
}

func ReadRootTD(msg *capnp.Message) (TD, error) {
	root, err := msg.Root()
	return TD(root.Struct()), err
}

func (s TD) String() string {
	str, _ := text.Marshal(0x9a428fdf25da5bb7, capnp.Struct(s))
	return str
}

func (s TD) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (TD) DecodeFromPtr(p capnp.Ptr) TD {
	return TD(capnp.Struct{}.DecodeFromPtr(p))
}

func (s TD) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s TD) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s TD) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s TD) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s TD) Id() (string, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.Text(), err
}

func (s TD) HasId() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s TD) IdBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.TextBytes(), err
}

func (s TD) SetId(v string) error {
	return capnp.Struct(s).SetText(0, v)
}

func (s TD) TdJson() (string, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.Text(), err
}

func (s TD) HasTdJson() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s TD) TdJsonBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.TextBytes(), err
}

func (s TD) SetTdJson(v string) error {
	return capnp.Struct(s).SetText(1, v)
}

// TD_List is a list of TD.
type TD_List = capnp.StructList[TD]

// NewTD creates a new list of TD.
func NewTD_List(s *capnp.Segment, sz int32) (TD_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return capnp.StructList[TD](l), err
}

// TD_Future is a wrapper for a TD promised by a client call.
type TD_Future struct{ *capnp.Future }

func (p TD_Future) Struct() (TD, error) {
	s, err := p.Future.Struct()
	return TD(s), err
}

type TDList capnp.Struct

// TDList_TypeID is the unique identifier for the type TDList.
const TDList_TypeID = 0xaa7e714db5bf7f48

func NewTDList(s *capnp.Segment) (TDList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return TDList(st), err
}

func NewRootTDList(s *capnp.Segment) (TDList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return TDList(st), err
}

func ReadRootTDList(msg *capnp.Message) (TDList, error) {
	root, err := msg.Root()
	return TDList(root.Struct()), err
}

func (s TDList) String() string {
	str, _ := text.Marshal(0xaa7e714db5bf7f48, capnp.Struct(s))
	return str
}

func (s TDList) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (TDList) DecodeFromPtr(p capnp.Ptr) TDList {
	return TDList(capnp.Struct{}.DecodeFromPtr(p))
}

func (s TDList) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s TDList) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s TDList) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s TDList) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s TDList) Tds() (capnp.TextList, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return capnp.TextList(p.List()), err
}

func (s TDList) HasTds() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s TDList) SetTds(v capnp.TextList) error {
	return capnp.Struct(s).SetPtr(0, v.ToPtr())
}

// NewTds sets the tds field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s TDList) NewTds(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(capnp.Struct(s).Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = capnp.Struct(s).SetPtr(0, l.ToPtr())
	return l, err
}

// TDList_List is a list of TDList.
type TDList_List = capnp.StructList[TDList]

// NewTDList creates a new list of TDList.
func NewTDList_List(s *capnp.Segment, sz int32) (TDList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return capnp.StructList[TDList](l), err
}

// TDList_Future is a wrapper for a TDList promised by a client call.
type TDList_Future struct{ *capnp.Future }

func (p TDList_Future) Struct() (TDList, error) {
	s, err := p.Future.Struct()
	return TDList(s), err
}

const schema_c2181f7117220fb9 = "x\xda\x12hs`1\xe4\x15gb`\x0a\x94`e" +
	"\xfb\xbf=\xfa\x96\xea\xfd~\xa7Y\x0c\x82\xb2\x8c\xffw" +
	"\xf2+\x89\x17\xcaK\x1cb`ebg`0\xfc\xc8" +
	"\xc4(\xf8\x97\x9d\x81A\xf0g9\x03\xe3\x7f\x8f\xfa\xfd" +
	"[}\x0b\xebV\xa1\xa9ddg`0\x0eeTb" +
	"\x14N\x051\x85\x13\x19\xed\x19t\xff\x97dd\xe6\xa5" +
	"\xeb\x87d0g\xe6\xa5\xbb\xa4\x16'\x17e\x16\x94d" +
	"\xe6\xe7\xe9%'\x16\xe4\x15X\x85\xb80\x0402\x06" +
	"r0\xb300\xb0020\x08jJ10\x04\xaa" +
	"03\x06\x1a01\x0a22\x8a0\x82\x04u\xad\x18" +
	"\x18\x025\x98\x19\x03M\x98\x18\x993S\x18y\x18\x98" +
	"\x18y\x18\x18\xedKR\xbc\x8a\xf3\xf3`\\\x82V\xf9" +
	"\xb0g\x16\x97\x80\xacc\x81[\xc7\xab\xc4\xc0\x10\xc8\xc1" +
	"\xcc\x18\xa8\xc2\xc4\xc8^\x92R\xcc\xc8\xc7\xc0\x18\xc0\xcc" +
	"\x086\x92\x8f\x81\x11\x10\x00\x00\xff\xffV\x10?\x97"

func init() {
	schemas.Register(schema_c2181f7117220fb9,
		0x9a428fdf25da5bb7,
		0xaa7e714db5bf7f48)
}
