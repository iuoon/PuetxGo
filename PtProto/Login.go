package PtProto

/**
 * Title:登陆信息
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import ("github.com/golang/protobuf/proto"
	"fmt"
	"math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// 登陆帐号请求.
type LoginServerReq struct {
	Tt               *TerminalType `protobuf:"varint,1,opt,name=tt,enum=GxProto.TerminalType" json:"tt,omitempty"`
	Pt               *PlatformType `protobuf:"varint,2,opt,name=pt,enum=GxProto.PlatformType" json:"pt,omitempty"`
	Gr               *GameRunType  `protobuf:"varint,3,opt,name=gr,enum=GxProto.GameRunType" json:"gr,omitempty"`
	Raw              *PlayerRaw    `protobuf:"bytes,4,opt,name=raw" json:"raw,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *LoginServerReq) Reset()         { *m = LoginServerReq{} }
func (m *LoginServerReq) String() string { return proto.CompactTextString(m) }
func (*LoginServerReq) ProtoMessage()    {}

func (m *LoginServerReq) GetTt() TerminalType {
	if m != nil && m.Tt != nil {
		return *m.Tt
	}
	return TerminalType_GC_TT_IPHONE
}

func (m *LoginServerReq) GetPt() PlatformType {
	if m != nil && m.Pt != nil {
		return *m.Pt
	}
	return PlatformType_GC_PT_RAW_GAS
}

func (m *LoginServerReq) GetGr() GameRunType {
	if m != nil && m.Gr != nil {
		return *m.Gr
	}
	return GameRunType_GC_GR_TEST
}

func (m *LoginServerReq) GetRaw() *PlayerRaw {
	if m != nil {
		return m.Raw
	}
	return nil
}

// 登陆帐号响应
type LoginServerRsp struct {
	Info             *LoginRspInfo `protobuf:"bytes,1,opt,name=info" json:"info,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *LoginServerRsp) Reset()         { *m = LoginServerRsp{} }
func (m *LoginServerRsp) String() string { return proto.CompactTextString(m) }
func (*LoginServerRsp) ProtoMessage()    {}

func (m *LoginServerRsp) GetInfo() *LoginRspInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

type GetGatesInfoRsp struct {
	Host             []string `protobuf:"bytes,1,rep,name=host" json:"host,omitempty"`
	Port             []uint32 `protobuf:"varint,2,rep,name=port" json:"port,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *GetGatesInfoRsp) Reset()         { *m = GetGatesInfoRsp{} }
func (m *GetGatesInfoRsp) String() string { return proto.CompactTextString(m) }
func (*GetGatesInfoRsp) ProtoMessage()    {}

func (m *GetGatesInfoRsp) GetHost() []string {
	if m != nil {
		return m.Host
	}
	return nil
}

func (m *GetGatesInfoRsp) GetPort() []uint32 {
	if m != nil {
		return m.Port
	}
	return nil
}