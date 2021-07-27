package scenario

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/isucon/isucon11-final/benchmarker/fails"

	"github.com/isucon/isucon11-final/benchmarker/generate"

	"github.com/isucon/isucandar/failure"

	api "github.com/isucon/isucon11-final/benchmarker/api"

	"github.com/isucon/isucon11-final/benchmarker/model"

	"github.com/isucon/isucandar/agent"
)

// action.go
// apiパッケージでリクエストを行う関数群
// param: modelオブジェクト
// POSTとか return: http.Response, error
// GETとか return: apiパッケージのオブジェクト, http.Response, error

const (
	RequestDuration = 100
)

func InitializeAction(ctx context.Context, agent *agent.Agent) (*http.Response, error) {
	return api.Initialize(ctx, agent)
}

func LoginAction(ctx context.Context, agent *agent.Agent, useraccount *model.UserAccount) (*http.Response, error) {
	req := api.LoginRequest{
		Code:     useraccount.Code,
		Password: useraccount.RawPassword,
	}
	hres, err := api.Login(ctx, agent, req)
	if err != nil {
		return nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK, http.StatusBadRequest, http.StatusForbidden})
	if err != nil {
		return nil, err
	}

	return hres, nil
}

func GetGradeAction(ctx context.Context, agent *agent.Agent) (*http.Response, api.GetGradeResponse, error) {
	res := api.GetGradeResponse{}
	hres, err := api.GetGrades(ctx, agent)
	if err != nil {
		return nil, res, err
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return nil, res, err
	}

	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, res, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, res, nil
}

func SearchCourseAction(ctx context.Context, agent *agent.Agent) (*http.Response, []*api.GetCourseDetailResponse, error) {
	// FIXME: param
	// MEMO: model.course.DayOfWeekは処理のしやすさを優先してintで持っている.apiリクエストに詰む際はよしなに変換して(hattori)
	hres, err := api.SearchCourse(ctx, agent, &api.SearchCourseRequest{})
	if err != nil {
		return nil, nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return hres, nil, err
	}

	res := make([]*api.GetCourseDetailResponse, 0)
	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, res, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, res, nil
}

func TakeCoursesAction(ctx context.Context, agent *agent.Agent, courses []*model.Course) (*http.Response, error) {
	req := make([]api.RegisterCourseRequestContent, 0, len(courses))
	for _, c := range courses {
		req = append(req, api.RegisterCourseRequestContent{ID: c.ID})
	}

	hres, err := api.RegisterCourses(ctx, agent, req)
	if err != nil {
		return nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return hres, err
	}

	res := make([]*api.GetCourseDetailResponse, 0)
	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, nil
}

func GetAnnouncementListAction(ctx context.Context, agent *agent.Agent, next string) (*http.Response, api.AnnouncementList, error) {
	res := api.AnnouncementList{}
	if next == "" {
		next = "/api/announcements"
	}
	hres, err := api.GetAnnouncementList(ctx, agent, next, nil)
	if err != nil {
		return nil, res, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return hres, nil, err
	}

	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, res, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, res, nil
}

func GetAnnouncementDetailAction(ctx context.Context, agent *agent.Agent, id string) (*http.Response, api.Announcement, error) {
	res := api.Announcement{}
	hres, err := api.GetAnnouncementDetail(ctx, agent, id)
	if err != nil {
		return nil, res, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return hres, res, err
	}

	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, res, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, res, nil
}

func AddClassAction(ctx context.Context, agent *agent.Agent, course *model.Course, part int) (*http.Response, *model.Class, *model.Announcement, error) {
	class := generate.Class(part)
	req := api.AddClassRequest{
		Part:        uint8(part),
		Title:       class.Title,
		Description: class.Desc,
		CreatedAt:   class.CreatedAt,
	}
	hres, err := api.AddClass(ctx, agent, course.ID, req)
	if err != nil {
		return nil, nil, nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusCreated})
	if err != nil {
		return hres, nil, nil, err
	}

	res := api.AddClassResponse{}
	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, nil, nil, failure.NewError(fails.ErrHTTP, err)
	}

	class.ID = res.ID

	// TODO
	announcement := model.NewAnnouncement("", course.ID, course.Name, "test title")
	return hres, class, announcement, nil
}

func AddCourseAction(ctx context.Context, faculty *model.Faculty, course *model.Course) (*http.Response, api.AddCourseResponse, error) {
	req := api.AddCourseRequest{
		Code:        course.Code,
		Type:        api.CourseType(course.Type),
		Name:        course.Name,
		Description: course.Description,
		Credit:      course.Credit,
		Period:      course.Period,
		DayOfWeek:   api.DayOfWeekTable[course.DayOfWeek],
		Keywords:    course.Keywords,
	}
	res := api.AddCourseResponse{}
	hres, err := api.AddCourse(ctx, faculty.Agent, req)
	if err != nil {
		return nil, res, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusCreated})
	if err != nil {
		return nil, res, err
	}

	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		return hres, res, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, res, nil
}

func SubmitAssignmentAction(ctx context.Context, agent *agent.Agent, courseID, classID string, submission *model.Submission) (*http.Response, error) {
	hres, err := api.SubmitAssignment(ctx, agent, courseID, classID, submission.Title, submission.Data)
	if err != nil {
		return nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}

	return hres, nil
}

func DownloadSubmissionsAction(ctx context.Context, agent *agent.Agent, courseID, classID string) (*http.Response, []byte, error) {
	hres, err := api.DownloadSubmittedAssignments(ctx, agent, courseID, classID)
	if err != nil {
		return nil, nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(hres.Body)
	if err != nil {
		return nil, nil, failure.NewError(fails.ErrHTTP, err)
	}

	return hres, data, nil
}

func PostGradeAction(ctx context.Context, agent *agent.Agent, courseID, classID string, scores []StudentScore) (*http.Response, error) {
	req := make([]api.RegisterScoreRequestContent, 0, len(scores))
	for _, v := range scores {
		req = append(req, api.RegisterScoreRequestContent{
			UserCode: v.code,
			Score:    v.score,
		})
	}
	hres, err := api.RegisterScores(ctx, agent, courseID, classID, req)
	if err != nil {
		return nil, failure.NewError(fails.ErrHTTP, err)
	}
	defer hres.Body.Close()

	err = verifyStatusCode(hres, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}

	return hres, nil
}
