package v1

import (
	"bytes"
	"encoding/json"
	"io"

	pb "gateway/api/upload/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/valyala/fasthttp"
)

const (
	url          = "https://i.impiggy.cn/api/1/upload/"
	token        = "8b86e6f30a68d4bac5430a8d21a93f51"
	format       = "json"
	cheveretoUrl = url + "?key=" + token + "&format=" + format
)

type UploadService struct {
	pb.UnimplementedUploadServer

	log *log.Helper
}

func NewUploadService(logger log.Logger) *UploadService {
	return &UploadService{
		log: log.NewHelper(log.With(logger, "module", "gateway/service/upload/v1", "caller", log.DefaultCaller)),
	}
}

func (s *UploadService) ImgUpload(conn pb.Upload_ImgUploadServer) error {
	var buffer bytes.Buffer

	// TODO 传输太慢, 打断
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if req.Content != nil {
			buffer.Write(req.Content)
		}
	}

	args := &fasthttp.Args{}
	args.AddBytesV("source", buffer.Bytes())

	stat, resp, err := fasthttp.Post(nil, cheveretoUrl, args)

	if err != nil {
		s.log.Error(err)
		_ = conn.SendAndClose(&pb.ImgUploadResponse{
			Code:    pb.Code_UNAVAILABLE,
			Message: "服务错误",
		})
		return err
	}
	if stat != fasthttp.StatusOK {
		s.log.Error(stat, string(resp))
	}

	r := &CheveretoResp{}
	err = json.Unmarshal(resp, r)
	if err != nil {
		s.log.Error(err)
		_ = conn.SendAndClose(&pb.ImgUploadResponse{
			Code:    pb.Code_UNAVAILABLE,
			Message: "服务错误",
		})
		return err
	}

	s.log.Infof("success upload url: %s", r.Image.URL)

	return conn.SendAndClose(&pb.ImgUploadResponse{
		Code:    pb.Code_OK,
		Message: "",
		Url:     r.Image.URL,
	})
}

type CheveretoResp struct {
	StatusCode int `json:"status_code"`
	//Success    struct {
	//	Message string `json:"message"`
	//	Code    int    `json:"code"`
	//} `json:"success"`
	Image struct {
		Name              string      `json:"name"`
		Extension         string      `json:"extension"`
		Size              string      `json:"size"`
		Width             string      `json:"width"`
		Height            string      `json:"height"`
		Date              string      `json:"date"`
		DateGmt           string      `json:"date_gmt"`
		Title             string      `json:"title"`
		Description       interface{} `json:"description"`
		Nsfw              string      `json:"nsfw"`
		StorageMode       string      `json:"storage_mode"`
		Md5               string      `json:"md5"`
		SourceMd5         interface{} `json:"source_md5"`
		OriginalFilename  string      `json:"original_filename"`
		OriginalExifdata  interface{} `json:"original_exifdata"`
		Views             string      `json:"views"`
		CategoryID        interface{} `json:"category_id"`
		Chain             string      `json:"chain"`
		ThumbSize         string      `json:"thumb_size"`
		MediumSize        string      `json:"medium_size"`
		ExpirationDateGmt interface{} `json:"expiration_date_gmt"`
		Likes             string      `json:"likes"`
		IsAnimated        string      `json:"is_animated"`
		IsApproved        string      `json:"is_approved"`
		File              struct {
			Resource struct {
				Type string `json:"type"`
			} `json:"resource"`
		} `json:"file"`
		IDEncoded string `json:"id_encoded"`
		Filename  string `json:"filename"`
		Mime      string `json:"mime"`
		URL       string `json:"url"`
		URLViewer string `json:"url_viewer"`
		URLShort  string `json:"url_short"`
		Image     struct {
			Filename  string `json:"filename"`
			Name      string `json:"name"`
			Mime      string `json:"mime"`
			Extension string `json:"extension"`
			URL       string `json:"url"`
			Size      string `json:"size"`
		} `json:"image"`
		Thumb struct {
			Filename  string `json:"filename"`
			Name      string `json:"name"`
			Mime      string `json:"mime"`
			Extension string `json:"extension"`
			URL       string `json:"url"`
			Size      string `json:"size"`
		} `json:"thumb"`
		SizeFormatted string `json:"size_formatted"`
		DisplayURL    string `json:"display_url"`
		//DisplayWidth       string `json:"display_width"`
		//DisplayHeight      string `json:"display_height"`
		ViewsLabel         string `json:"views_label"`
		LikesLabel         string `json:"likes_label"`
		HowLongAgo         string `json:"how_long_ago"`
		DateFixedPeer      string `json:"date_fixed_peer"`
		TitleTruncated     string `json:"title_truncated"`
		TitleTruncatedHTML string `json:"title_truncated_html"`
		IsUseLoader        bool   `json:"is_use_loader"`
	} `json:"image"`
	StatusTxt string `json:"status_txt"`
}
