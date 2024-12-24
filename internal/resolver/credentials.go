package resolver

import (
	"bytes"
	"fmt"
	"github.com/go-git/go-git/v5/config"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"net/url"
	"os/exec"
	"strings"
)

const credentialSection = "credential"

type CredentialConfigEntry struct {
	Helper      []string
	Username    string
	UseHttpPath bool
}

type Credentials map[string]*CredentialConfigEntry

type GitCredential struct {
	Protocol string
	Host     string
	Username string
	Password string
	URL      string
}

func (c GitCredential) String() string {
	var b strings.Builder

	if c.Protocol != "" {
		fmt.Fprintf(&b, "protocol=%s\n", c.Protocol)
	}

	// If full URL is provided, use that
	if c.URL != "" {
		fmt.Fprintf(&b, "url=%s\n", c.URL)
	}

	// Otherwise fall back to protocol/host
	if c.Host != "" {
		fmt.Fprintf(&b, "host=%s\n", c.Host)
	}

	if c.Username != "" {
		fmt.Fprintf(&b, "username=%s\n", c.Username)
	}
	if c.Password != "" {
		fmt.Fprintf(&b, "password=%s\n", c.Password)
	}
	return b.String()
}

func newCredential(opts format.Options) *CredentialConfigEntry {
	return &CredentialConfigEntry{
		Helper:      opts.GetAll("helper"),
		Username:    opts.Get("username"),
		UseHttpPath: opts.Get("usehttppath") == "true",
	}
}

func buildCredentialCommand(helperName, action string) (*exec.Cmd, error) {
	if strings.HasPrefix(helperName, "!") {
		// For helpers starting with !, execute the command directly
		cmdStr := strings.TrimPrefix(helperName, "!")
		cmdParts := strings.Fields(cmdStr)
		if len(cmdParts) == 0 {
			return nil, fmt.Errorf("invalid credential helper command: %s", helperName)
		}

		// Append the action to the command parts
		cmdParts = append(cmdParts, action)
		return exec.Command(cmdParts[0], cmdParts[1:]...), nil
	}

	// Regular git credential helper
	return exec.Command("git", "credential-"+helperName, action), nil
}

func invokeCredentialHelper(helperName, action string, cred GitCredential) (GitCredential, error) {
	cmd, err := buildCredentialCommand(helperName, action)
	if err != nil {
		return GitCredential{}, err
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(cred.String())

	if err := cmd.Run(); err != nil {
		return GitCredential{}, fmt.Errorf("credential helper failed: %v, stderr: %s", err, stderr.String())
	}

	// Parse output
	output := stdout.String()
	result := GitCredential{}

	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "protocol":
			result.Protocol = parts[1]
		case "host":
			result.Host = parts[1]
		case "username":
			result.Username = parts[1]
		case "password":
			result.Password = parts[1]
		}
	}

	return result, nil
}

var ErrNoCredentialHelperFound = fmt.Errorf("no credential helper found")

func (c *CredentialConfigEntry) Evaluate(repoURL string) (*GitCredential, error) {
	parsedUrl, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}

	testedUrl := repoURL
	if !c.UseHttpPath {
		testedUrl = parsedUrl.Scheme + "://" + parsedUrl.Host
	}

	cred := GitCredential{
		Protocol: parsedUrl.Scheme,
		Host:     parsedUrl.Host,
		URL:      testedUrl,
		Username: c.Username,
	}

	for _, helperName := range c.Helper {
		cred, err = invokeCredentialHelper(helperName, "get", cred)
		if err != nil {
			return nil, err
		}

		return &cred, nil
	}

	return nil, ErrNoCredentialHelperFound
}

func ParseGitCredentials() (Credentials, error) {
	c, err := config.LoadConfig(config.GlobalScope)
	if err != nil {
		return nil, err
	}

	sect := c.Raw.Section(credentialSection)

	result := make(map[string]*CredentialConfigEntry)
	result["default"] = newCredential(sect.Options)

	for _, sub := range sect.Subsections {
		result[sub.Name] = newCredential(sub.Options)
	}

	return result, nil
}

func (c Credentials) Has(section string) bool {
	_, ok := c[section]
	return ok
}

func (c Credentials) Get(host string) *CredentialConfigEntry {

	parsedURL, err := url.Parse(host)
	if err != nil {
		return c["default"]
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	checkTargets := []string{parsedURL.Scheme + "://" + parsedURL.Host, "default"}
	for _, target := range checkTargets {
		if c.Has(target) {
			return c[target]
		}
	}

	return nil
}
