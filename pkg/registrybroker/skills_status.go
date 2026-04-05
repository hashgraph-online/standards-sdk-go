package registrybroker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *RegistryBrokerClient) GetSkillStatus(
	ctx context.Context,
	request SkillStatusRequest,
) (SkillStatusResponse, error) {
	if err := ensureNonEmpty(request.Name, "name"); err != nil {
		return SkillStatusResponse{}, err
	}
	query := url.Values{}
	addQueryString(query, "name", request.Name)
	addQueryString(query, "version", request.Version)
	var response SkillStatusResponse
	err := c.requestTypedJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/status", query),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) GetSkillStatusByRepo(
	ctx context.Context,
	request SkillPreviewByRepoRequest,
) (SkillStatusResponse, error) {
	query, err := skillPreviewRepoQuery(request)
	if err != nil {
		return SkillStatusResponse{}, err
	}
	var response SkillStatusResponse
	err = c.requestTypedJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/status/by-repo", query),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) QuoteSkillPublishPreview(
	ctx context.Context,
	request SkillQuotePreviewRequest,
) (SkillQuotePreviewResponse, error) {
	if request.FileCount <= 0 {
		return SkillQuotePreviewResponse{}, fmt.Errorf("fileCount must be greater than 0")
	}
	if request.TotalBytes <= 0 {
		return SkillQuotePreviewResponse{}, fmt.Errorf("totalBytes must be greater than 0")
	}
	var response SkillQuotePreviewResponse
	err := c.requestTypedJSON(
		ctx,
		http.MethodPost,
		"/skills/quote-preview",
		request,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) GetSkillConversionSignalsByRepo(
	ctx context.Context,
	request SkillPreviewByRepoRequest,
) (SkillConversionSignalsResponse, error) {
	query, err := skillPreviewRepoQuery(request)
	if err != nil {
		return SkillConversionSignalsResponse{}, err
	}
	var response SkillConversionSignalsResponse
	err = c.requestTypedJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/conversion-signals/by-repo", query),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) GetSkillPreview(
	ctx context.Context,
	request SkillPreviewLookupRequest,
) (SkillPreviewLookupResponse, error) {
	if err := ensureNonEmpty(request.Name, "name"); err != nil {
		return SkillPreviewLookupResponse{}, err
	}
	query := url.Values{}
	addQueryString(query, "name", request.Name)
	addQueryString(query, "version", request.Version)
	var response SkillPreviewLookupResponse
	err := c.requestTypedJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/preview", query),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) GetSkillPreviewByRepo(
	ctx context.Context,
	request SkillPreviewByRepoRequest,
) (SkillPreviewLookupResponse, error) {
	query, err := skillPreviewRepoQuery(request)
	if err != nil {
		return SkillPreviewLookupResponse{}, err
	}
	var response SkillPreviewLookupResponse
	err = c.requestTypedJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/preview/by-repo", query),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) GetSkillPreviewByID(
	ctx context.Context,
	previewID string,
) (SkillPreviewLookupResponse, error) {
	if err := ensureNonEmpty(previewID, "previewID"); err != nil {
		return SkillPreviewLookupResponse{}, err
	}
	var response SkillPreviewLookupResponse
	err := c.requestTypedJSON(
		ctx,
		http.MethodGet,
		"/skills/preview/"+percentPath(strings.TrimSpace(previewID)),
		nil,
		nil,
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) UploadSkillPreviewFromGithubOIDC(
	ctx context.Context,
	token string,
	report *SkillPreviewReport,
) (SkillPreviewRecord, error) {
	if err := ensureNonEmpty(token, "token"); err != nil {
		return SkillPreviewRecord{}, err
	}
	if report == nil {
		return SkillPreviewRecord{}, ensureNonEmpty("", "report")
	}
	var response SkillPreviewRecord
	err := c.requestTypedJSON(
		ctx,
		http.MethodPost,
		"/skills/preview/github-oidc",
		report,
		map[string]string{
			"authorization": "Bearer " + strings.TrimSpace(token),
		},
		&response,
	)
	return response, err
}

func (c *RegistryBrokerClient) UploadSkillPreviewFromGitHubOIDC(
	ctx context.Context,
	token string,
	report *SkillPreviewReport,
) (SkillPreviewRecord, error) {
	return c.UploadSkillPreviewFromGithubOIDC(ctx, token, report)
}

func skillPreviewRepoQuery(request SkillPreviewByRepoRequest) (url.Values, error) {
	if err := ensureNonEmpty(request.Repo, "repo"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(request.SkillDir, "skillDir"); err != nil {
		return nil, err
	}
	query := url.Values{}
	addQueryString(query, "repo", request.Repo)
	addQueryString(query, "skillDir", request.SkillDir)
	addQueryString(query, "ref", request.Ref)
	return query, nil
}
