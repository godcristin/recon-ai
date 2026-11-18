package service

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"time"

	appV1 "recon-ai/api/app/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/bff/internal/biz"
	"recon-ai/app/bff/internal/conf"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAppService)

// AppService 对外 HTTP 接口：协议转换 + 聚合
type AppService struct {
	appV1.UnimplementedAppServer

	uc        *biz.ReconUseCase
	log       *log.Helper
	maxUpload int64
}

func NewAppService(uc *biz.ReconUseCase, c *conf.Bootstrap, logger log.Logger) *AppService {
	maxUpload := c.Client.MaxUploadBytes
	if maxUpload <= 0 {
		maxUpload = 20 << 20 // 20MB
	}
	return &AppService{
		uc:        uc,
		log:       log.NewHelper(log.With(logger, "module", "service/app")),
		maxUpload: maxUpload,
	}
}

func (s *AppService) GetTask(ctx context.Context, req *reconV1.GetTaskReq) (*reconV1.Task, error) {
	return s.uc.GetTask(ctx, req.Id)
}

func (s *AppService) ListDiffs(ctx context.Context, req *reconV1.ListDiffsReq) (*reconV1.ListDiffsRsp, error) {
	return s.uc.ListDiffs(ctx, req)
}

func (s *AppService) GetDiff(ctx context.Context, req *reconV1.GetDiffReq) (*reconV1.DiffDetail, error) {
	return s.uc.GetDiff(ctx, req.Id)
}

func (s *AppService) ConfirmDiff(ctx context.Context, req *reconV1.ConfirmDiffReq) (*reconV1.ConfirmDiffRsp, error) {
	return s.uc.ConfirmDiff(ctx, req)
}

func (s *AppService) GetTaskSummary(ctx context.Context, req *reconV1.GetTaskSummaryReq) (*reconV1.TaskSummaryRsp, error) {
	return s.uc.Summary(ctx, req.TaskId)
}

// UploadTask 自定义路由：multipart 上传两份 CSV（google.api.http 不适合表达文件上传）
func (s *AppService) UploadTask(ctx khttp.Context) error {
	r := ctx.Request()
	r.Body = stdhttp.MaxBytesReader(ctx.Response(), r.Body, s.maxUpload)

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		return kerrors.BadRequest("UPLOAD_INVALID", "解析上传内容失败: "+err.Error())
	}

	name := r.FormValue("name")
	if name == "" {
		name = "task-" + time.Now().Format("20060102-150405")
	}

	channel, err := readPart(r, "channel")
	if err != nil {
		return kerrors.BadRequest("UPLOAD_INVALID", "渠道账单: "+err.Error())
	}
	system, err := readPart(r, "system")
	if err != nil {
		return kerrors.BadRequest("UPLOAD_INVALID", "系统订单: "+err.Error())
	}

	rsp, err := s.uc.CreateTask(r.Context(), name, channel, system)
	if err != nil {
		s.log.WithContext(r.Context()).Errorw("msg", "CreateTask failed", "err", err, "name", name)
		return kerrors.InternalServer("CREATE_TASK_FAILED", err.Error())
	}

	// 用生成的 HTTP 层同一套编码（protojson），保证所有接口字段风格一致
	return ctx.Result(stdhttp.StatusOK, rsp)
}

func readPart(r *stdhttp.Request, field string) ([]byte, error) {
	f, _, err := r.FormFile(field)
	if err != nil {
		return nil, fmt.Errorf("缺少文件字段 %q", field)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	data, err := io.ReadAll(io.LimitReader(f, 20<<20))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("文件为空")
	}
	return data, nil
}
