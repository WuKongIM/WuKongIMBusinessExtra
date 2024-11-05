package message

import (
	"sort"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/gocraft/dbr/v2"
)

type messageExtraDB struct {
	ctx     *config.Context
	session *dbr.Session
}

func newMessageExtraDB(ctx *config.Context) *messageExtraDB {
	return &messageExtraDB{
		ctx:     ctx,
		session: ctx.DB(),
	}
}

func (m *messageExtraDB) insert(md *messageExtraModel) error {
	_, err := m.session.InsertInto("message_extra").Columns(util.AttrToUnderscore(md)...).Record(md).Exec()
	return err
}

func (m *messageExtraDB) update(md *messageExtraModel) error {
	_, err := m.session.Update("message_extra").SetMap(map[string]interface{}{
		"readed_count": md.ReadedCount,
		"version":      md.Version,
		"revoke":       md.Revoke,
		"revoker":      md.Revoker,
	}).Where("message_id=?", md.MessageID).Exec()
	return err
}

func (m *messageExtraDB) queryWithMessageIDsAndUID(messageIDs []string, loginUID string) ([]*messageExtraDetailModel, error) {
	if len(messageIDs) <= 0 {
		return nil, nil
	}
	var models []*messageExtraDetailModel
	_, err := m.session.Select("message_extra.*,(select count(*) from member_readed where member_readed.message_id=message_extra.message_id and member_readed.uid='"+loginUID+"') readed,(select created_at from member_readed where member_readed.message_id=message_extra.message_id and member_readed.uid='"+loginUID+"') readed_at").From("message_extra").Where("message_id in ?", messageIDs).Load(&models)
	return models, err
}

func (m *messageExtraDB) queryWithMessageID(messageID string) (*messageExtraModel, error) {
	var model *messageExtraModel
	_, err := m.session.Select("*").From("message_extra").Where("message_id=?", messageID).Load(&model)
	return model, err
}

func (m *messageExtraDB) sync(version int64, channelID string, channelType uint8, limit uint64, loginUID string) ([]*messageExtraDetailModel, error) {
	var models []*messageExtraDetailModel
	selectSql := "message_extra.*,(select count(*) from member_readed where member_readed.message_id=message_extra.message_id and member_readed.uid='" + loginUID + "') readed,(select created_at from member_readed where member_readed.message_id=message_extra.message_id and member_readed.uid='" + loginUID + "') readed_at"
	builder := m.session.Select(selectSql).From("message_extra")
	var err error
	if version == 0 {
		builder = builder.Where("channel_id=? and channel_type=?", channelID, channelType).OrderDesc("version").Limit(limit)
		_, err = builder.Load(&models)
		newModels := messageExtraDetailModelSlice(models)
		sort.Sort(newModels)
		models = newModels
	} else {
		builder = builder.Where("channel_id=? and channel_type=? and version>?", channelID, channelType, version).OrderAsc("version").Limit(limit)
		_, err = builder.Load(&models)
	}

	return models, err
}

type messageExtraDetailModelSlice []*messageExtraDetailModel

func (m messageExtraDetailModelSlice) Len() int {
	return len(m)
}
func (m messageExtraDetailModelSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m messageExtraDetailModelSlice) Less(i, j int) bool { return m[i].Version < m[j].Version }

type messageExtraDetailModel struct {
	messageExtraModel
	Readed   int          // 是否已读（针对于自己）
	ReadedAt dbr.NullTime // 已读时间

}

type messageExtraModel struct {
	MessageID       string
	MessageSeq      uint32
	FromUID         string
	ChannelID       string
	ChannelType     uint8
	Revoke          int
	Revoker         string // 消息撤回者的uid
	CloneNo         string
	ReadedCount     int            // 已读数量
	ContentEdit     dbr.NullString // 编辑后的正文
	ContentEditHash string
	EditedAt        int // 编辑时间 时间戳（秒）
	IsDeleted       int
	Version         int64 // 数据版本
	IsPinned        int   // 是否置顶
	db.BaseModel
}