package api

import (
	"pkg.mon.icu/monicu/internal/storage/_model"
)

type imageModel struct {
	URL    string `json:"url"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Size   uint64 `json:"size"`
}

type postModel struct {
	ID        _model.Ref    `json:"id"`
	ChannelID _model.Ref    `json:"channel"`
	UserID    _model.Ref    `json:"user"`
	Images    []*imageModel `json:"images"`
	Reactions uint32        `json:"reactions"`
}

func (a *API) getAllPosts(page uint32) ([]*postModel, error) {
	var pm []*postModel
	//if err := a.storage.Begin(a.ctx, func(tx pgx.Tx) error {
	//	var posts []*_model.Post
	//	var err error
	//	if posts, err = _model.FindPosts(a.ctx, tx, page * 100, 100); err != nil {
	//		return err
	//	}
	//
	//	pm = make([]*postModel, len(posts))
	//	for i, p := range posts {
	//		var images []*_model.Image
	//		if images, err = _model.FindImages(a.ctx, tx, p); err != nil {
	//			return err
	//		}
	//
	//		imm := make([]*imageModel, len(images))
	//		for j, im := range images {
	//			imm[j] = &imageModel{im.URL, im.Width, im.Height, im.Size}
	//		}
	//
	//		var rc uint32
	//		if rc, err = _model.CountUserReactions(a.ctx, tx, p); err != nil {
	//			return err
	//		}
	//
	//		pm[i] = &postModel{p.ID, p.ChannelID, p.UserID, imm, rc}
	//	}
	//
	//	return nil
	//}); err != nil {
	//	return nil, err
	//}

	return pm, nil
}