package gitlab

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// GetRefPipeline ..
func (c *Client) GetRefPipeline(ctx context.Context, ref schemas.Ref, pipelineID int) (p schemas.Pipeline, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetRefPipeline")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", ref.Project.Name))
	span.SetAttributes(attribute.String("ref_name", ref.Name))
	span.SetAttributes(attribute.Int("pipeline_id", pipelineID))

	c.rateLimit(ctx)

	gp, resp, err := c.Pipelines.GetPipeline(ref.Project.Name, pipelineID, goGitlab.WithContext(ctx))
	if err != nil || gp == nil {
		return schemas.Pipeline{}, fmt.Errorf("could not read content of pipeline %s - %s | %s", ref.Project.Name, ref.Name, err.Error())
	}

	c.requestsRemaining(resp)

	return schemas.NewPipeline(ctx, *gp), nil
}

// GetProjectPipelines ..
func (c *Client) GetProjectPipelines(
	ctx context.Context,
	projectName string,
	options *goGitlab.ListProjectPipelinesOptions,
) (
	[]*goGitlab.PipelineInfo,
	*goGitlab.Response,
	error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectPipelines")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", projectName))

	fields := log.Fields{
		"project-name": projectName,
	}

	if options.Page == 0 {
		options.Page = 1
	}

	if options.PerPage == 0 {
		options.PerPage = 100
	}

	if options.Ref != nil {
		fields["ref"] = *options.Ref
	}

	if options.Scope != nil {
		fields["scope"] = *options.Scope
	}

	fields["page"] = options.Page
	log.WithFields(fields).Trace("listing project pipelines")

	c.rateLimit(ctx)

	pipelines, resp, err := c.Pipelines.ListProjectPipelines(projectName, options, goGitlab.WithContext(ctx))
	if err != nil {
		return nil, resp, fmt.Errorf("error listing project pipelines for project %s: %s", projectName, err.Error())
	}

	c.requestsRemaining(resp)

	return pipelines, resp, nil
}

// GetRefPipelineVariablesAsConcatenatedString ..
func (c *Client) GetRefPipelineVariablesAsConcatenatedString(ctx context.Context, ref schemas.Ref) (string, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetRefPipelineVariablesAsConcatenatedString")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", ref.Project.Name))
	span.SetAttributes(attribute.String("ref_name", ref.Name))

	if reflect.DeepEqual(ref.LatestPipeline, (schemas.Pipeline{})) {
		log.WithFields(
			log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting..")

		return "", nil
	}

	log.WithFields(
		log.Fields{
			"project-name": ref.Project.Name,
			"ref":          ref.Name,
			"pipeline-id":  ref.LatestPipeline.ID,
		},
	).Debug("fetching pipeline variables")

	variablesFilter, err := regexp.Compile(ref.Project.Pull.Pipeline.Variables.Regexp)
	if err != nil {
		return "", fmt.Errorf(
			"the provided filter regex for pipeline variables is invalid '(%s)': %v",
			ref.Project.Pull.Pipeline.Variables.Regexp,
			err,
		)
	}

	c.rateLimit(ctx)

	variables, resp, err := c.Pipelines.GetPipelineVariables(ref.Project.Name, ref.LatestPipeline.ID, goGitlab.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("could not fetch pipeline variables for %d: %s", ref.LatestPipeline.ID, err.Error())
	}

	c.requestsRemaining(resp)

	var keptVariables []string

	if len(variables) > 0 {
		for _, v := range variables {
			if variablesFilter.MatchString(v.Key) {
				keptVariables = append(keptVariables, strings.Join([]string{v.Key, v.Value}, ":"))
			}
		}
	}

	return strings.Join(keptVariables, ","), nil
}

// GetRefsFromPipelines ..
func (c *Client) GetRefsFromPipelines(ctx context.Context, p schemas.Project, refKind schemas.RefKind) (refs schemas.Refs, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetRefsFromPipelines")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", p.Name))
	span.SetAttributes(attribute.String("ref_kind", string(refKind)))

	refs = make(schemas.Refs)

	options := &goGitlab.ListProjectPipelinesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		OrderBy: goGitlab.String("updated_at"),
	}

	var re *regexp.Regexp

	if re, err = schemas.GetRefRegexp(p.Pull.Refs, refKind); err != nil {
		return
	}

	var (
		mostRecent, maxAgeSeconds         uint
		limitToMostRecent, excludeDeleted bool
		existingRefs                      schemas.Refs
	)

	switch refKind {
	case schemas.RefKindMergeRequest:
		maxAgeSeconds = p.Pull.Refs.MergeRequests.MaxAgeSeconds
		mostRecent = p.Pull.Refs.MergeRequests.MostRecent
	case schemas.RefKindBranch:
		options.Scope = goGitlab.String("branches")
		maxAgeSeconds = p.Pull.Refs.Branches.MaxAgeSeconds
		mostRecent = p.Pull.Refs.Branches.MostRecent

		if p.Pull.Refs.Branches.ExcludeDeleted {
			excludeDeleted = true

			if existingRefs, err = c.GetProjectBranches(ctx, p); err != nil {
				return
			}
		}
	case schemas.RefKindTag:
		options.Scope = goGitlab.String("tags")
		maxAgeSeconds = p.Pull.Refs.Tags.MaxAgeSeconds
		mostRecent = p.Pull.Refs.Tags.MostRecent

		if p.Pull.Refs.Tags.ExcludeDeleted {
			excludeDeleted = true

			if existingRefs, err = c.GetProjectTags(ctx, p); err != nil {
				return
			}
		}
	default:
		return refs, fmt.Errorf("unsupported ref kind %v", refKind)
	}

	if mostRecent > 0 {
		limitToMostRecent = true
	}

	if maxAgeSeconds > 0 {
		t := time.Now().Add(-time.Second * time.Duration(maxAgeSeconds))
		options.UpdatedAfter = &t
	}

	for {
		var (
			pipelines []*goGitlab.PipelineInfo
			resp      *goGitlab.Response
		)

		pipelines, resp, err = c.GetProjectPipelines(ctx, p.Name, options)
		if err != nil {
			return
		}

		for _, pipeline := range pipelines {
			refName := pipeline.Ref
			if !re.MatchString(refName) {
				// It is quite verbose otherwise..
				if refKind != schemas.RefKindMergeRequest {
					log.WithField("ref", refName).Debug("discovered pipeline ref not matching regexp")
				}

				continue
			}

			if refKind == schemas.RefKindMergeRequest {
				if refName, err = schemas.GetMergeRequestIIDFromRefName(refName); err != nil {
					log.WithContext(ctx).
						WithField("ref", refName).
						WithError(err).
						Warn()

					continue
				}
			}

			ref := schemas.NewRef(
				p,
				refKind,
				refName,
			)

			if excludeDeleted {
				if _, refExists := existingRefs[ref.Key()]; !refExists {
					log.WithFields(log.Fields{
						"project-name": ref.Project.Name,
						"ref":          ref.Name,
						"ref-kind":     ref.Kind,
					}).Debug("found deleted ref, ignoring..")

					continue
				}
			}

			if _, ok := refs[ref.Key()]; !ok {
				log.WithFields(log.Fields{
					"project-name": ref.Project.Name,
					"ref":          ref.Name,
					"ref-kind":     ref.Kind,
				}).Trace("found ref")

				refs[ref.Key()] = ref

				if limitToMostRecent {
					mostRecent--
					if mostRecent <= 0 {
						return
					}
				}
			}
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}

		options.Page = resp.NextPage
	}

	return
}

// GetRefPipelineTestReport ..
func (c *Client) GetRefPipelineTestReport(ctx context.Context, ref schemas.Ref) (schemas.TestReport, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetRefPipelineTestReport")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", ref.Project.Name))
	span.SetAttributes(attribute.String("ref_name", ref.Name))

	if reflect.DeepEqual(ref.LatestPipeline, (schemas.Pipeline{})) {
		log.WithFields(
			log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting...")

		return schemas.TestReport{}, nil
	}

	log.WithFields(
		log.Fields{
			"project-name": ref.Project.Name,
			"ref":          ref.Name,
			"pipeline-id":  ref.LatestPipeline.ID,
		},
	).Debug("fetching pipeline test report")

	c.rateLimit(ctx)

	type pipelineDef struct {
		projectNameOrID string
		pipelineID      int
	}

	var currentPipeline pipelineDef

	baseTestReport := schemas.TestReport{
		TotalTime:    0,
		TotalCount:   0,
		SuccessCount: 0,
		FailedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites:   []schemas.TestSuite{},
	}
	childPipelines := []pipelineDef{{ref.Project.Name, ref.LatestPipeline.ID}}

	for {
		if len(childPipelines) == 0 {
			return baseTestReport, nil
		}

		currentPipeline, childPipelines = childPipelines[0], childPipelines[1:]

		testReport, resp, err := c.Pipelines.GetPipelineTestReport(currentPipeline.projectNameOrID, currentPipeline.pipelineID, goGitlab.WithContext(ctx))
		if err != nil {
			return schemas.TestReport{}, fmt.Errorf("could not fetch test report for %d: %s", ref.LatestPipeline.ID, err.Error())
		}

		c.requestsRemaining(resp)

		convertedTestReport := schemas.NewTestReport(*testReport)

		baseTestReport = schemas.TestReport{
			TotalTime:    baseTestReport.TotalTime + convertedTestReport.TotalTime,
			TotalCount:   baseTestReport.TotalCount + convertedTestReport.TotalCount,
			SuccessCount: baseTestReport.SuccessCount + convertedTestReport.SuccessCount,
			FailedCount:  baseTestReport.FailedCount + convertedTestReport.FailedCount,
			SkippedCount: baseTestReport.SkippedCount + convertedTestReport.SkippedCount,
			ErrorCount:   baseTestReport.ErrorCount + convertedTestReport.ErrorCount,
			TestSuites:   append(baseTestReport.TestSuites, convertedTestReport.TestSuites...),
		}

		if ref.Project.Pull.Pipeline.TestReports.FromChildPipelines.Enabled {
			foundBridges, err := c.ListPipelineBridges(ctx, currentPipeline.projectNameOrID, currentPipeline.pipelineID)
			if err != nil {
				return baseTestReport, err
			}

			for _, foundBridge := range foundBridges {
				if foundBridge.DownstreamPipeline == nil {
					continue
				}

				childPipelines = append(childPipelines, pipelineDef{strconv.Itoa(foundBridge.DownstreamPipeline.ProjectID), foundBridge.DownstreamPipeline.ID})
			}
		}
	}
}
