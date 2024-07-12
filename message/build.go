package message

import (
	"encoding/json"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

type ElementBuilder interface {
	BuildElement() []*message.Elem
}

func (e *TextElement) BuildElement() []*message.Elem {
	return []*message.Elem{{Text: &message.Text{Str: proto.Some(e.Content)}}}
}

func (e *AtElement) BuildElement() []*message.Elem {
	var atAll int32 = 2
	if e.TargetUin == 0 {
		atAll = 1
	}
	reserve := message.MentionExtra{
		Type:   proto.Some(atAll),
		Uin:    proto.Some(uint32(0)),
		Field5: proto.Some(int32(0)),
		Uid:    proto.Some(e.TargetUid),
	}
	reserveData, _ := proto.Marshal(&reserve)
	return []*message.Elem{{Text: &message.Text{
		Str:       proto.Some(e.Display),
		PbReserve: reserveData,
	}}}
}

func (e *FaceElement) BuildElement() []*message.Elem {
	faceId := int32(e.FaceID)
	if e.isLargeFace {
		qFace := message.QFaceExtra{
			Field1:  proto.Some("1"),
			Field2:  proto.Some("8"),
			FaceId:  proto.Some(faceId),
			Field4:  proto.Some(int32(1)),
			Field5:  proto.Some(int32(1)),
			Field6:  proto.Some(""),
			Preview: proto.Some(""),
			Field9:  proto.Some(int32(1)),
		}
		qFaceData, _ := proto.Marshal(&qFace)
		return []*message.Elem{{
			CommonElem: &message.CommonElem{
				ServiceType:  37,
				PbElem:       qFaceData,
				BusinessType: 1,
			},
		}}
	} else {
		return []*message.Elem{{
			Face: &message.Face{Index: proto.Some(faceId)},
		}}
	}
}

func (e *ImageElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		return nil
	}
	msg := []*message.Elem{{}, {
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 10,
		},
	}}
	if e.CompatFace != nil {
		msg[0].CustomFace = e.CompatFace
	}
	if e.CompatImage != nil {
		msg[0].NotOnlineImage = e.CompatImage
	}
	return msg
}

func (e *ReplyElement) BuildElement() []*message.Elem {
	forwardReserve := message.Preserve{
		MessageId:   uint64(e.ReplySeq),
		ReceiverUid: e.SenderUid,
	}
	forwardReserveData, err := proto.Marshal(&forwardReserve)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		SrcMsg: &message.SrcMsg{
			OrigSeqs:  []uint32{e.ReplySeq},
			SenderUin: uint64(e.SenderUin),
			Time:      proto.Some(int32(e.Time)),
			Elems:     PackElements(e.Elements),
			PbReserve: forwardReserveData,
			ToUin:     proto.Some(uint64(0)),
		},
	}}
}

func (e *VoiceElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 22,
		},
	}}
}

func (e *ShortVideoElement) BuildElement() []*message.Elem {
	return nil
}

func (e *LightAppElement) BuildElement() []*message.Elem {
	return []*message.Elem{{
		LightAppElem: &message.LightAppElem{
			Data: append([]byte{0x01}, binary.ZlibCompress([]byte(e.Content))...),
		},
	}}
}

func (e *ForwardMessage) BuildElement() []*message.Elem {
	fileId := utils.NewUUID()
	extra := MultiMsgLightAppExtra{
		FileName: fileId,
		Sum:      len(e.Nodes),
	}
	extraData, err := json.Marshal(&extra)
	if err != nil {
		return nil
	}

	var news []News
	if len(e.Nodes) != 0 {
		news = make([]News, len(e.Nodes))
		for i, node := range e.Nodes {
			news[i] = News{Text: fmt.Sprintf("%s: %s", node.SenderName, ToReadableString(node.Message))}
		}
	} else {
		news = []News{{Text: "转发消息"}}
	}

	content := MultiMsgLightApp{
		App: "com.tencent.multimsg",
		Config: Config{
			Autosize: 1,
			Forward:  1,
			Round:    1,
			Type:     "normal",
			Width:    300,
		},
		Desc:  "[聊天记录]",
		Extra: utils.B2S(extraData),
		Meta: Meta{
			Detail: Detail{
				News:    news,
				Resid:   e.ResID,
				Source:  "聊天记录",
				Summary: fmt.Sprintf("查看%d条转发消息", len(e.Nodes)),
				UniSeq:  fileId,
			},
		},
		Prompt: "[聊天记录]",
		Ver:    "0.0.0.5",
		View:   "contact",
	}

	contentData, err := json.Marshal(&content)
	return NewLightApp(utils.B2S(contentData)).BuildElement()
}
