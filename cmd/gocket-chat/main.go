package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	gochat "github.com/pandatix/gocket-chat"
	"github.com/pandatix/gocket-chat/api/chat"
	"github.com/urfave/cli/v2"
)

const version = "v1.0-alpha"

const header = `
  ________               __           __    _________ .__            __
 /  _____/  ____   ____ |  | __ _____/  |_  \_   ___ \|  |__ _____ _/  |_
/   \  ___ /  _ \_/ ___\|  |/ // __ \   __\ /    \  \/|  |  \\__  \\   __\
\    \_\  (  <_> )  \___|    <\  ___/|  |   \     \___|   Y  \/ __ \|  |
 \______  /\____/ \___  >__|_ \\___  >__|    \______  /___|  (____  /__|
        \/            \/     \/    \/               \/     \/     \/
		                                                                ` + version + `
`

const defaultMessage = `
{{ if eq .Drone.Build.Status "success" }}✅ {{ else }}❌ {{ end -}}
Pipeline {{ .Drone.Build.Number }} of repo ` + "`{{ .Drone.Repo.Fullname }}` on `{{ .Drone.Repo.Branch }}`" + ` branch {{ if eq .Drone.Build.Status "success" }}**passed magically**{{ else }}**failed miserably**
**Failed steps:** {{ .Drone.Failed.Steps }}
{{ end }}
`

func main() {
	app := &cli.App{
		Name:      "🤖 Gocket Chat 🚀",
		Usage:     "Drone plugin that sends messages to a RocketChat server",
		Copyright: "Copyright (c) " + strconv.Itoa(time.Now().Year()) + " PandatiX (Lucas TESSON)",
		Authors: []*cli.Author{
			{
				Name:  "PandatiX (Lucas TESSON)",
				Email: "lucasfloriantesson@gmail.com",
			},
		},
		Version: version,
		Action:  gocketChat,
		Flags: []cli.Flag{
			// CLI's specific flags
			cli.HelpFlag,
			cli.VersionFlag,
			// RocketChat's specific flags
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"u"},
				Usage:   "RocketChat server URL",
				EnvVars: []string{"URL"},
			},
			&cli.StringFlag{
				Name:    "user-id",
				Aliases: []string{"i"},
				Usage:   "RocketChat user's ID",
				EnvVars: []string{"USER_ID"},
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "RocketChat token",
				EnvVars: []string{"TOKEN"},
			},
			&cli.StringFlag{
				Name:    "channel",
				Aliases: []string{"c"},
				Usage:   "RocketChat channel (or user ID)",
				EnvVars: []string{"CHANNEL"},
			},
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "RocketChat message. Takes a Go template, see Drone variables accessible in other flags",
				EnvVars: []string{"MESSAGE"},
				Value:   defaultMessage,
			},
			&cli.StringFlag{
				Name:    "alias",
				Usage:   "RocketChat alias to post the message with.",
				EnvVars: []string{"ALIAS"},
				Value:   "🤖 Gocket Chat 🚀",
			},
			&cli.StringFlag{
				Name:    "avatar-url",
				Usage:   "RocketChat avatar URL to post the message with.",
				EnvVars: []string{"AVATAR_URL"},
			},
			// Drone's specific flags
			&cli.StringFlag{
				Name:    "ci",
				Usage:   "Identifies the current environment as a Continuous Integration environment.",
				EnvVars: []string{"CI"},
			},
			&cli.StringFlag{
				Name:    "drone",
				Usage:   "Identifies the current environment as the Drone Continuous Integration environment.",
				EnvVars: []string{"DRONE"},
			},
			&cli.StringFlag{
				Name:    "branch",
				Usage:   "Provides the target branch for the push or pull request. This value may be empty for tag events.",
				EnvVars: []string{"DRONE_BRANCH"},
			},
			&cli.StringFlag{
				Name:    "build.action",
				Usage:   "Provides the action that triggered the pipeline execution. Use this value to differentiate between the pull request being opened vs synchronized.",
				EnvVars: []string{"DRONE_BUILD_ACTION"},
			},
			&cli.StringFlag{
				Name:    "build.created",
				Usage:   "Provides the unix timestamp for when the build object was created by the system.",
				EnvVars: []string{"DRONE_BUILD_CREATED"},
			},
			&cli.StringFlag{
				Name:    "build.event",
				Usage:   "Provides the event that triggered the pipeline execution.",
				EnvVars: []string{"DRONE_BUILD_EVENT"},
			},
			&cli.StringFlag{
				Name:    "build.finished",
				Usage:   "Provides the unix timestamp for when the build is finished. A running build cannot have a finish timestamp, therefore, the system always sets this value to the current timestamp.",
				EnvVars: []string{"DRONE_BUILD_FINISHED"},
			},
			&cli.StringFlag{
				Name:    "build.number",
				Usage:   "Provides the build number for the current running build.",
				EnvVars: []string{"DRONE_BUILD_NUMBER"},
			},
			&cli.StringFlag{
				Name:    "build.parent",
				Usage:   "Provides the parent build number for the current running build. The parent build number is populated from an exiting build that is being promoted.",
				EnvVars: []string{"DRONE_BUILD_PARENT"},
			},
			&cli.StringFlag{
				Name:    "build.started",
				Usage:   "Provides the unix timestamp for when the build was started by the system.",
				EnvVars: []string{"DRONE_BUILD_STARTED"},
			},
			&cli.StringFlag{
				Name:    "build.status",
				Usage:   "Provides the status for the current running build. If build pipelines and build steps are passing, the build status defaults to success.",
				EnvVars: []string{"DRONE_BUILD_STATUS"},
			},
			&cli.StringFlag{
				Name:    "calver",
				Usage:   "If the git tag is a valid calendar version string, the system provides the tag as a calendar version string.",
				EnvVars: []string{"DRONE_CALVER"},
			},
			&cli.StringFlag{
				Name:    "commit.after",
				Usage:   "Provides the git commit sha after the patch is applied. This may be used in conjunction with the before commit sha to create a diff.",
				EnvVars: []string{"DRONE_COMMIT_AFTER"},
			},
			&cli.StringFlag{
				Name:    "commit.author",
				Usage:   "Provides the commit author username for the current running build. This is the username from source control management system (e.g. GitHub username).",
				EnvVars: []string{"DRONE_COMMIT_AUTHOR"},
			},
			&cli.StringFlag{
				Name:    "commit.author.avatar",
				Usage:   "Provides the commit author avatar for the current running build. This is the avatar from source control management system (e.g. GitHub).",
				EnvVars: []string{"DRONE_COMMIT_AUTHOR_AVATAR"},
			},
			&cli.StringFlag{
				Name:    "commit.author.email",
				Usage:   "Provides the commit email address for the current running build. Note this is a user-defined value and may be empty or inaccurate.",
				EnvVars: []string{"DRONE_COMMIT_AUTHOR_EMAIL"},
			},
			&cli.StringFlag{
				Name:    "commit.author.name",
				Usage:   "Provides the commit author name for the current running build. Note this is a user-defined value and may be empty or inaccurate.",
				EnvVars: []string{"DRONE_COMMIT_AUTHOR_NAME"},
			},
			&cli.StringFlag{
				Name:    "commit.before",
				Usage:   "Provides the git commit sha before the patch is applied. This may be used in conjunction with the after commit sha to create a diff.",
				EnvVars: []string{"DRONE_COMMIT_BEFORE"},
			},
			&cli.StringFlag{
				Name:    "commit.branch",
				Usage:   "Provides the target branch for the push or pull request. This value may be empty for tag events.",
				EnvVars: []string{"DRONE_COMMIT_BRANCH"},
			},
			&cli.StringFlag{
				Name:    "commit.link",
				Usage:   "Provides a link the git commit or object in the source control management system.",
				EnvVars: []string{"DRONE_COMMIT_LINK"},
			},
			&cli.StringFlag{
				Name:    "commit.message",
				Usage:   "Provides the commit message for the current running build.",
				EnvVars: []string{"DRONE_COMMIT_MESSAGE"},
			},
			&cli.StringFlag{
				Name:    "commit.ref",
				Usage:   "Provides the git reference for the current running build.",
				EnvVars: []string{"DRONE_COMMIT_REF"},
			},
			&cli.StringFlag{
				Name:    "commit.sha",
				Usage:   "Provides the git commit sha for the current running build.",
				EnvVars: []string{"DRONE_COMMIT_SHA"},
			},
			&cli.StringFlag{
				Name:    "deploy-to",
				Usage:   "Provides the target deployment environment for the running build. This value is only available to promotion and rollback pipelines.",
				EnvVars: []string{"DRONE_DEPLOY_TO"},
			},
			&cli.StringFlag{
				Name:    "failed.stages",
				Usage:   "Provides a comma-separate list of failed pipeline stages for the current running build.",
				EnvVars: []string{"DRONE_FAILED_STAGES"},
			},
			&cli.StringFlag{
				Name:    "failed.steps",
				Usage:   "Provides a comma-separate list of failed pipeline steps.",
				EnvVars: []string{"DRONE_FAILED_STEPS"},
			},
			&cli.StringFlag{
				Name:    "git.http-url",
				Usage:   "Provides the git+http url that should be used to clone the repository.",
				EnvVars: []string{"DRONE_GIT_HTTP_URL"},
			},
			&cli.StringFlag{
				Name:    "git.ssh-url",
				Usage:   "Provides the git+ssh url that should be used to clone the repository.",
				EnvVars: []string{"DRONE_GIT_SSH_URL"},
			},
			&cli.StringFlag{
				Name:    "pull-request",
				Usage:   "Provides the pull request number for the current running build. If the build is not a pull request the variable is empty.",
				EnvVars: []string{"DRONE_PULL_REQUEST"},
			},
			&cli.StringFlag{
				Name:    "remote-url",
				Usage:   "Provides the git+https url that should be used to clone the repository. This is a legacy value included for backward compatibility only.",
				EnvVars: []string{"DRONE_REMOTE_URL"},
			},
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "Provides the full repository name for the current running build.",
				EnvVars: []string{"DRONE_REPO"},
			},
			&cli.StringFlag{
				Name:    "repo.branch",
				Usage:   "Provides the default repository branch for the current running build.",
				EnvVars: []string{"DRONE_REPO_BRANCH"},
			},
			&cli.StringFlag{
				Name:    "repo.link",
				Usage:   "Provides the repository link for the current running build.",
				EnvVars: []string{"DRONE_REPO_LINK"},
			},
			&cli.StringFlag{
				Name:    "repo.name",
				Usage:   "Provides the repository name for the current running build.",
				EnvVars: []string{"DRONE_REPO_NAME"},
			},
			&cli.StringFlag{
				Name:    "repo.namespace",
				Usage:   "Provides the repository namespace for the current running build. The namespace is an alias for the source control management account that owns the repository.",
				EnvVars: []string{"DRONE_REPO_NAMESPACE"},
			},
			&cli.StringFlag{
				Name:    "repo.owner",
				Usage:   "Provides the repository namespace for the current running build. The namespace is an alias for the source control management account that owns the repository.",
				EnvVars: []string{"DRONE_REPO_OWNER"},
			},
			&cli.StringFlag{
				Name:    "repo.private",
				Usage:   "Provides a boolean flag that indicates whether or not the repository is private or public.",
				EnvVars: []string{"DRONE_REPO_PRIVATE"},
			},
			&cli.StringFlag{
				Name:    "repo.scm",
				Usage:   "Provides the repository type for the current running build.",
				EnvVars: []string{"DRONE_REPO_SCM"},
			},
			&cli.StringFlag{
				Name:    "repo.visibility",
				Usage:   "Provides the repository visibility level for the current running build.",
				EnvVars: []string{"DRONE_REPO_VISIBILITY"},
			},
			&cli.StringFlag{
				Name:    "semver",
				Usage:   "If the git tag is a valid semantic version string, the system provides the tag as a semver string.",
				EnvVars: []string{"DRONE_SEMVER"},
			},
			&cli.StringFlag{
				Name:    "semver.build",
				Usage:   "If the git tag is a valid semver string, this variable provides the build from the semver string.",
				EnvVars: []string{"DRONE_SEMVER_BUILD"},
			},
			&cli.StringFlag{
				Name:    "semver.error",
				Usage:   "If the git tag is not a valid semver string, this variable provides the semver parsing error.",
				EnvVars: []string{"DRONE_SEMVER_ERROR"},
			},
			&cli.StringFlag{
				Name:    "semver.major",
				Usage:   "If the git tag is a valid semver string, this variable provides the major version number from the semver string.",
				EnvVars: []string{"DRONE_SEMVER_MAJOR"},
			},
			&cli.StringFlag{
				Name:    "semver.minor",
				Usage:   "If the git tag is a valid semver string, this variable provides the minor version number from the semver string.",
				EnvVars: []string{"DRONE_SEMVER_MINOR"},
			},
			&cli.StringFlag{
				Name:    "semver.patch",
				Usage:   "If the git tag is a valid semver string, this variable provides the patch from the semver string.",
				EnvVars: []string{"DRONE_SEMVER_PATCH"},
			},
			&cli.StringFlag{
				Name:    "semver.prerelease",
				Usage:   "If the git tag is a valid semver string, this variable provides the prelease from the semver string.",
				EnvVars: []string{"DRONE_SEMVER_PRERELEASE"},
			},
			&cli.StringFlag{
				Name:    "semver.short",
				Usage:   "If the git tag is a valid semver string, this variable provides the short version of the semver string where labels and metadata are truncated.",
				EnvVars: []string{"DRONE_SEMVER_SHORT"},
			},
			&cli.StringFlag{
				Name:    "source-branch",
				Usage:   "Provides the source branch for the pull request. This value may be empty for certain source control management providers.",
				EnvVars: []string{"DRONE_SOURCE_BRANCH"},
			},
			&cli.StringFlag{
				Name:    "stage.arch",
				Usage:   "Provides the platform architecture for the current build stage.",
				EnvVars: []string{"DRONE_STAGE_ARCH"},
			},
			&cli.StringFlag{
				Name:    "stage.depends-on",
				Usage:   "Provides a comma-separated list of dependencies for the current pipeline stage.",
				EnvVars: []string{"DRONE_STAGE_DEPENDS_ON"},
			},
			&cli.StringFlag{
				Name:    "stage.finished",
				Usage:   "Provides the unix timestamp for when the pipeline is finished. A running pipeline cannot have a finish timestamp, therefore, the system always sets this value to the current timestamp.",
				EnvVars: []string{"DRONE_STAGE_FINISHED"},
			},
			&cli.StringFlag{
				Name:    "stage.kind",
				Usage:   "Provides the kind of resource being executed. This value is sourced from the kind attribute in the yaml configuration file.",
				EnvVars: []string{"DRONE_STAGE_KIND"},
			},
			&cli.StringFlag{
				Name:    "stage.machine",
				Usage:   "Provides the name of the host machine on which the pipeline is currently running.",
				EnvVars: []string{"DRONE_STAGE_MACHINE"},
			},
			&cli.StringFlag{
				Name:    "stage.name",
				Usage:   "Provides the stage name for the current running pipeline stage.",
				EnvVars: []string{"DRONE_STAGE_NAME"},
			},
			&cli.StringFlag{
				Name:    "stage.number",
				Usage:   "Provides the stage number for the current running pipeline stage.",
				EnvVars: []string{"DRONE_STAGE_NUMBER"},
			},
			&cli.StringFlag{
				Name:    "stage.os",
				Usage:   "Provides the target operating system for the current build stage.",
				EnvVars: []string{"DRONE_STAGE_OS"},
			},
			&cli.StringFlag{
				Name:    "stage.started",
				Usage:   "Provides the unix timestamp for when the pipeline was started by the runner.",
				EnvVars: []string{"DRONE_STAGE_STARTED"},
			},
			&cli.StringFlag{
				Name:    "stage.status",
				Usage:   "Provides the status for the current running pipeline. If all pipeline steps are passing, the pipeline status defaults to success.",
				EnvVars: []string{"DRONE_STAGE_STATUS"},
			},
			&cli.StringFlag{
				Name:    "stage.type",
				Usage:   "Provides the type of resource being executed. This value is sourced from the type attribute in the yaml configuration file.",
				EnvVars: []string{"DRONE_STAGE_TYPE"},
			},
			&cli.StringFlag{
				Name:    "stage.variant",
				Usage:   "Provides the target operating architecture variable for the current build stage. This variable is optional and is only available for arm architectures.",
				EnvVars: []string{"DRONE_STAGE_VARIANT"},
			},
			&cli.StringFlag{
				Name:    "step.name",
				Usage:   "Provides the step name for the current running pipeline step.",
				EnvVars: []string{"DRONE_STEP_NAME"},
			},
			&cli.StringFlag{
				Name:    "step.number",
				Usage:   "Provides the step number for the current running pipeline step.",
				EnvVars: []string{"DRONE_STEP_NUMBER"},
			},
			&cli.StringFlag{
				Name:    "system.host",
				Usage:   "Provides the hostname used by the Drone server. This can be combined with the protocol to construct to the server url.",
				EnvVars: []string{"DRONE_SYSTEM_HOST"},
			},
			&cli.StringFlag{
				Name:    "system.hostname",
				Usage:   "Provides the hostname used by the Drone server. This can be combined with the protocol to construct to the server url.",
				EnvVars: []string{"DRONE_SYSTEM_HOSTNAME"},
			},
			&cli.StringFlag{
				Name:    "system.proto",
				Usage:   "Provides the protocol used by the Drone server. This can be combined with the hostname to construct to the server url.",
				EnvVars: []string{"DRONE_SYSTEM_PROTO"},
			},
			&cli.StringFlag{
				Name:    "system.version",
				Usage:   "Provides the version of the Drone server.",
				EnvVars: []string{"DRONE_SYSTEM_VERSION"},
			},
			&cli.StringFlag{
				Name:    "tag",
				Usage:   "Provides the tag for the current running build. This value is only populated for tag events and promotion events that are derived from tags.",
				EnvVars: []string{"DRONE_TAG"},
			},
			&cli.StringFlag{
				Name:    "target-branch",
				Usage:   "Provides the target branch for the push or pull request. This value may be empty for tag events.",
				EnvVars: []string{"DRONE_TARGET_BRANCH"},
			},
		},
		UsageText:       "gocket-chat [flags]",
		HideHelpCommand: true,
	}
	cli.AppHelpTemplate = header + cli.AppHelpTemplate

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("❌ Process failed: %s\n", err)
	}
}

func gocketChat(ctx *cli.Context) error {
	// Get flags values
	url := ctx.String("url")
	user_id := ctx.String("user-id")
	token := ctx.String("token")
	channel := ctx.String("channel")
	message := ctx.String("message")
	alias := ctx.String("alias")
	avatar_url := ctx.String("avatar-url")

	// Extract CLI flags/env variables for message templating
	plugin := buildPlugin(ctx)

	// Parse template message
	fmt.Printf("🔬 Parsing message template\n")
	tmpl, err := template.New("msg-tmpl").Parse(message)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, plugin)
	if err != nil {
		return err
	}
	msg := buf.String()

	// Init client
	rc, _ := gochat.NewRocketClient(&http.Client{}, url, token, user_id)

	// Send the message
	pmp := chat.PostMessageParams{
		Channel: channel,
		Text:    &msg,
		Alias:   &alias,
		Avatar:  &avatar_url,
	}

	fmt.Printf("📩 Sending the message to %s as \"%s\"\n", channel, alias)
	_, err = chat.PostMessage(rc, pmp)
	if err != nil {
		return err
	}

	// Log response
	fmt.Printf("🚀 Everything worked fine. See ya' !\n")

	return nil
}

func buildPlugin(ctx *cli.Context) *Plugin {
	return &Plugin{
		CI: ctx.String("ci"),
		Drone: Drone{
			Status: ctx.String("drone"),
			Branch: ctx.String("branch"),
			Build: Build{
				Action:   ctx.String("build.action"),
				Created:  ctx.String("build.created"),
				Event:    ctx.String("build.event"),
				Finished: ctx.String("build.finished"),
				Number:   ctx.String("build.number"),
				Parent:   ctx.String("build.parent"),
				Started:  ctx.String("build.started"),
				Status:   ctx.String("build.status"),
			},
			Calver: ctx.String("calver"),
			Commit: Commit{
				After: ctx.String("commit.after"),
				Author: Author{
					Username: ctx.String("commit.author"),
					Avatar:   ctx.String("commit.author.avatar"),
					Email:    ctx.String("commit.author.email"),
					Name:     ctx.String("commit.author.name"),
				},
				Before:  ctx.String("commit.before"),
				Branch:  ctx.String("commit.branch"),
				Link:    ctx.String("commit.link"),
				Message: ctx.String("commit.message"),
				Ref:     ctx.String("commit.ref"),
				SHA:     ctx.String("commit.sha"),
			},
			DeployTo: ctx.String("deploy-to"),
			Failed: Failed{
				Stages: ctx.String("failed.stages"),
				Steps:  ctx.String("failed.steps"),
			},
			Git: Git{
				HTTPURL: ctx.String("git.http-url"),
				SSHURL:  ctx.String("git.ssh-url"),
			},
			PullRequest: ctx.String("pull-request"),
			RemoteURL:   ctx.String("remote-url"),
			Repo: Repo{
				Fullname:   ctx.String("repo"),
				Branch:     ctx.String("repo.branch"),
				Link:       ctx.String("repo.link"),
				Name:       ctx.String("repo.name"),
				Namespace:  ctx.String("repo.namespace"),
				Owner:      ctx.String("repo.owner"),
				Private:    ctx.String("repo.private"),
				SCM:        ctx.String("repo.scm"),
				Visibility: ctx.String("repo.visibility"),
			},
			Semver: Semver{
				Full:       ctx.String("semver"),
				Build:      ctx.String("semver.build"),
				Error:      ctx.String("semver.error"),
				Major:      ctx.String("semver.major"),
				Minor:      ctx.String("semver.minor"),
				Patch:      ctx.String("semver.patch"),
				Prerelease: ctx.String("semver.prerelease"),
				Short:      ctx.String("semver.short"),
			},
			SourceBranch: ctx.String("source-branch"),
			Stage: Stage{
				Arch:      ctx.String("stage.arch"),
				DependsOn: ctx.String("stage.depends-on"),
				Finished:  ctx.String("stage.finished"),
				Kind:      ctx.String("stage.kind"),
				Machine:   ctx.String("stage.machine"),
				Name:      ctx.String("stage.name"),
				Number:    ctx.String("stage.number"),
				OS:        ctx.String("stage.os"),
				Started:   ctx.String("stage.started"),
				Status:    ctx.String("stage.status"),
				Type:      ctx.String("stage.type"),
				Variant:   ctx.String("stage.variant"),
			},
			Step: Step{
				Name:   ctx.String("step.name"),
				Number: ctx.String("step.number"),
			},
			System: System{
				Host:     ctx.String("system.host"),
				Hostname: ctx.String("system.hostname"),
				Proto:    ctx.String("system.proto"),
				Version:  ctx.String("system.version"),
			},
			Tag:          ctx.String("tag"),
			TargetBranch: ctx.String("target-branch"),
		},
	}
}

type (
	Plugin struct {
		CI    string
		Drone Drone
	}

	Drone struct {
		Status       string
		Branch       string
		Build        Build
		Calver       string
		Commit       Commit
		DeployTo     string
		Failed       Failed
		Git          Git
		PullRequest  string
		RemoteURL    string
		Repo         Repo
		Semver       Semver
		SourceBranch string
		Stage        Stage
		Step         Step
		System       System
		Tag          string
		TargetBranch string
	}

	Build struct {
		Action   string
		Created  string
		Event    string
		Finished string
		Number   string
		Parent   string
		Started  string
		Status   string
	}

	Commit struct {
		After   string
		Author  Author
		Before  string
		Branch  string
		Link    string
		Message string
		Ref     string
		SHA     string
	}

	Author struct {
		Username string
		Avatar   string
		Email    string
		Name     string
	}

	Failed struct {
		Stages string
		Steps  string
	}

	Git struct {
		HTTPURL string
		SSHURL  string
	}

	Repo struct {
		Fullname   string
		Branch     string
		Link       string
		Name       string
		Namespace  string
		Owner      string
		Private    string
		SCM        string
		Visibility string
	}

	Semver struct {
		Full       string
		Build      string
		Error      string
		Major      string
		Minor      string
		Patch      string
		Prerelease string
		Short      string
	}

	Stage struct {
		Arch      string
		DependsOn string
		Finished  string
		Kind      string
		Machine   string
		Name      string
		Number    string
		OS        string
		Started   string
		Status    string
		Type      string
		Variant   string
	}

	Step struct {
		Name   string
		Number string
	}

	System struct {
		Host     string
		Hostname string
		Proto    string
		Version  string
	}
)
