package project

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	model "github.com/dooray-go/dooray-sdk/openapi/model/project"
)

// GetPostsOptions contains query parameters for the GetPosts API.
type GetPostsOptions struct {
	// Paging
	Page *int // 기본값 0
	Size *int // 기본값 20, 최댓값 100

	// Filter
	FromEmailAddress    string // From 이메일 주소로 업무 필터링
	FromMemberIds       string // 특정 멤버가 작성한 업무 목록 (comma-separated)
	ToMemberIds         string // 특정 멤버가 담당자인 업무 목록 (comma-separated)
	ToMemberSize        *int   // 업무 담당자 수 (0: 담당자 없는 업무, 1: toMemberIds[0]이 혼자 담당인 업무)
	CcMemberIds         string // 특정 멤버가 참조자인 업무 목록 (comma-separated)
	TagIds              string // 특정 태그가 붙은 업무 목록 (comma-separated)
	ParentPostId        string // 특정 업무의 하위 업무 목록
	PostNumber          string // 특정 업무의 번호
	PostWorkflowClasses string // backlog, registered, working, closed (comma-separated)
	PostWorkflowIds     string // 워크플로우 ID 필터 (comma-separated)
	MilestoneIds        string // 단계 ID 기준 필터 (comma-separated)
	Subjects            string // 업무 제목으로 필터

	// Date filters (DATE_PATTERN: today, thisweek, prev-{N}d, next-{N}d, or ISO8601 range)
	CreatedAt string // 생성시간 기준 필터
	UpdatedAt string // 업데이트 기준 필터
	DueAt     string // 만기시간 기준 필터

	// Sort
	Order string // postDueAt, postUpdatedAt, createdAt (역순: -createdAt)
}

func (c *Project) GetPosts(apikey string, projectId string, toMemberIds string, postWorkflowClasses string) (*model.GetPostsResponse, error) {
	opts := GetPostsOptions{
		ToMemberIds:         toMemberIds,
		PostWorkflowClasses: postWorkflowClasses,
	}
	return c.GetPostsWithOptionsCustomHTTPContext(context.Background(), apikey, c.httpClient, projectId, opts)
}

func (c *Project) GetPostsContext(ctx context.Context, apikey string, projectId string, toMemberIds string, postWorkflowClasses string) (*model.GetPostsResponse, error) {
	opts := GetPostsOptions{
		ToMemberIds:         toMemberIds,
		PostWorkflowClasses: postWorkflowClasses,
	}
	return c.GetPostsWithOptionsCustomHTTPContext(ctx, apikey, c.httpClient, projectId, opts)
}

func (c *Project) GetPostsCustomHTTP(apikey string, httpClient *http.Client, projectId string, toMemberIds string, postWorkflowClasses string) (*model.GetPostsResponse, error) {
	opts := GetPostsOptions{
		ToMemberIds:         toMemberIds,
		PostWorkflowClasses: postWorkflowClasses,
	}
	return c.GetPostsWithOptionsCustomHTTPContext(context.Background(), apikey, httpClient, projectId, opts)
}

func (c *Project) GetPostsCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, projectId string, toMemberIds string, postWorkflowClasses string) (*model.GetPostsResponse, error) {
	opts := GetPostsOptions{
		ToMemberIds:         toMemberIds,
		PostWorkflowClasses: postWorkflowClasses,
	}
	return c.GetPostsWithOptionsCustomHTTPContext(ctx, apikey, httpClient, projectId, opts)
}

// GetPostsWithOptions retrieves posts with full query parameter support.
func (c *Project) GetPostsWithOptions(apikey string, projectId string, opts GetPostsOptions) (*model.GetPostsResponse, error) {
	return c.GetPostsWithOptionsCustomHTTPContext(context.Background(), apikey, c.httpClient, projectId, opts)
}

func (c *Project) GetPostsWithOptionsContext(ctx context.Context, apikey string, projectId string, opts GetPostsOptions) (*model.GetPostsResponse, error) {
	return c.GetPostsWithOptionsCustomHTTPContext(ctx, apikey, c.httpClient, projectId, opts)
}

func (c *Project) GetPostsWithOptionsCustomHTTP(apikey string, httpClient *http.Client, projectId string, opts GetPostsOptions) (*model.GetPostsResponse, error) {
	return c.GetPostsWithOptionsCustomHTTPContext(context.Background(), apikey, httpClient, projectId, opts)
}

func (c *Project) GetPostsWithOptionsCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, projectId string, opts GetPostsOptions) (*model.GetPostsResponse, error) {
	url := c.endPoint + "/project/v1/projects/" + projectId + "/posts"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	// Paging
	if opts.Page != nil {
		query.Set("page", strconv.Itoa(*opts.Page))
	}
	if opts.Size != nil {
		query.Set("size", strconv.Itoa(*opts.Size))
	}

	// Filters
	if opts.FromEmailAddress != "" {
		query.Set("fromEmailAddress", opts.FromEmailAddress)
	}
	if opts.FromMemberIds != "" {
		query.Set("fromMemberIds", opts.FromMemberIds)
	}
	if opts.ToMemberIds != "" {
		query.Set("toMemberIds", opts.ToMemberIds)
	}
	if opts.ToMemberSize != nil {
		query.Set("toMemberSize", strconv.Itoa(*opts.ToMemberSize))
	}
	if opts.CcMemberIds != "" {
		query.Set("ccMemberIds", opts.CcMemberIds)
	}
	if opts.TagIds != "" {
		query.Set("tagIds", opts.TagIds)
	}
	if opts.ParentPostId != "" {
		query.Set("parentPostId", opts.ParentPostId)
	}
	if opts.PostNumber != "" {
		query.Set("postNumber", opts.PostNumber)
	}
	if opts.PostWorkflowClasses != "" {
		query.Set("postWorkflowClasses", opts.PostWorkflowClasses)
	}
	if opts.PostWorkflowIds != "" {
		query.Set("postWorkflowIds", opts.PostWorkflowIds)
	}
	if opts.MilestoneIds != "" {
		query.Set("milestoneIds", opts.MilestoneIds)
	}
	if opts.Subjects != "" {
		query.Set("subjects", opts.Subjects)
	}

	// Date filters
	if opts.CreatedAt != "" {
		query.Set("createdAt", opts.CreatedAt)
	}
	if opts.UpdatedAt != "" {
		query.Set("updatedAt", opts.UpdatedAt)
	}
	if opts.DueAt != "" {
		query.Set("dueAt", opts.DueAt)
	}

	// Sort
	if opts.Order != "" {
		query.Set("order", opts.Order)
	}

	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", "dooray-api "+apikey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		resp.Body.Close()
	}()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var getpostsResponse model.GetPostsResponse
	if err := json.Unmarshal(resBody, &getpostsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	getpostsResponse.RawJSON = string(resBody)

	return &getpostsResponse, nil
}
